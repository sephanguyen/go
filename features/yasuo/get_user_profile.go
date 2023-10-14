package yasuo

import (
	"context"
	"fmt"
	"reflect"

	"github.com/manabie-com/backend/features/usermgmt"
	repositories_bob "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	userConstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	"github.com/manabie-com/backend/internal/yasuo/repositories"
	bobpb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) userGetProfile(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	stepState.Request = &pb.GetBasicProfileRequest{}
	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.Conn).GetBasicProfile(contextWithToken(s, ctx), stepState.Request.(*pb.GetBasicProfileRequest))
	return nil
}

func (s *suite) yasuoMustReturnUserProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}

	resp := stepState.Response.(*pb.GetBasicProfileResponse)
	respUser := resp.User

	schoolRepo := repositories.SchoolRepo{}
	userRepo := repositories_bob.UserRepo{}
	user, err := userRepo.Get(ctx, s.DBTrace, database.Text(respUser.Id))
	if err != nil {
		return ctx, fmt.Errorf("userRepo.GetProfile: %w", err)
	}

	userGroup := user.Group.String

	schoolIDs := []int64{}
	schools := []*pb.UserProfile_SchoolInfo{}
	switch userGroup {
	case constant.UserGroupTeacher:
		teacherRepo := repositories_bob.TeacherRepo{}
		teacher, err := teacherRepo.FindByID(ctx, s.DBTrace, database.Text(respUser.Id))
		if err != nil {
			return ctx, fmt.Errorf("teacherRepo.Get: %w", err)
		}
		sIDs := []int32{}
		for _, v := range teacher.SchoolIDs.Elements {
			sIDs = append(sIDs, v.Int)
		}

		enSchools, err := schoolRepo.Get(ctx, s.DBTrace, sIDs)
		if err != nil {
			return ctx, fmt.Errorf("s.SchoolRepo.Get: %w", err)
		}
		if len(enSchools) == 0 {
			return ctx, status.Error(codes.NotFound, "cannot find schools")
		}
		for _, v := range enSchools {
			schools = append(schools, &pb.UserProfile_SchoolInfo{
				SchoolId:   int64(v.ID.Int),
				SchoolName: v.Name.String,
			})
			schoolIDs = append(schoolIDs, int64(v.ID.Int))
		}

	case constant.UserGroupSchoolAdmin:
		schoolAdminRepo := repositories_bob.SchoolAdminRepo{}
		schoolAdmin, err := schoolAdminRepo.Get(ctx, s.DBTrace, database.Text(respUser.Id))
		if err != nil {
			return ctx, fmt.Errorf("schoolAdminRepo.Get: %w", err)
		}

		enSchools, err := schoolRepo.Get(ctx, s.DBTrace, []int32{schoolAdmin.SchoolID.Int})
		if err != nil {
			return ctx, fmt.Errorf("s.SchoolRepo.Get: %w", err)
		}
		enSchool, ok := enSchools[schoolAdmin.SchoolID.Int]
		if !ok {
			return ctx, status.Error(codes.NotFound, "cannot find school")
		}
		schools = []*pb.UserProfile_SchoolInfo{{
			SchoolId:   int64(enSchool.ID.Int),
			SchoolName: enSchool.Name.String,
		}}
		schoolIDs = []int64{int64(enSchool.ID.Int)}

	default:
		break
	}

	respUser.CreatedAt = nil
	respUser.UpdatedAt = nil

	profile := &pb.UserProfile{
		Id:          user.ID.String,
		Name:        user.GetName(),
		Country:     bobpb.Country(bobpb.Country_value[user.Country.String]),
		PhoneNumber: user.PhoneNumber.String,
		DeviceToken: user.DeviceToken.String,
		UserGroup:   user.Group.String,
		Email:       user.Email.String,
		Avatar:      user.Avatar.String,
		UpdatedAt:   nil,
		CreatedAt:   nil,
		Schools:     nil,
		SchoolIds:   nil,
		UserGroupV2: respUser.UserGroupV2,
	}

	if len(schools) > 0 {
		if !reflect.DeepEqual(schools, respUser.Schools) {
			return ctx, errors.New("schools not match")
		}
		if !reflect.DeepEqual(schoolIDs, respUser.SchoolIds) {
			return ctx, errors.New("school ids not match")
		}
	}
	respUser.Schools = nil
	profile.Schools = nil
	respUser.SchoolIds = nil
	profile.SchoolIds = nil

	if !reflect.DeepEqual(respUser, profile) {
		return ctx, errors.New("profile not match")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signedAsAccountHaveUserGroup(ctx context.Context, signedIn string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := s.createStaff(ctx, upb.UserGroup_USER_GROUP_TEACHER)
	if err != nil {
		return ctx, err
	}

	token, err := s.generateExchangeToken(resp.Staff.StaffId, upb.UserGroup_USER_GROUP_TEACHER.String())
	if err != nil {
		return ctx, err
	}

	stepState.AuthToken = token
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStaff(ctx context.Context, userGroup upb.UserGroup) (*upb.CreateStaffResponse, error) {
	ctx, err := s.aSignedIn(ctx, "school admin")
	if err != nil {
		return nil, err
	}
	userGroupResp, err := usermgmt.SeedUserGroup(s.signedCtx(ctx), s.DBTrace, s.userManagementConn, []string{userConstant.RoleTeacher})
	if err != nil {
		return nil, fmt.Errorf("createStaff: %w", err)
	}

	num := idutil.ULIDNow()
	staff := &upb.CreateStaffRequest_StaffProfile{
		Name:           fmt.Sprintf("create_staff+%s", num),
		Email:          fmt.Sprintf("create_staff+%s@gmail.com", num),
		Country:        cpb.Country_COUNTRY_VN,
		PhoneNumber:    "",
		UserGroup:      userGroup,
		UserGroupIds:   []string{userGroupResp.UserGroupId},
		OrganizationId: fmt.Sprint(constant.ManabieSchool),
	}

	req := &upb.CreateStaffRequest{
		Staff: staff,
	}

	return upb.NewStaffServiceClient(s.userManagementConn).CreateStaff(s.signedCtx(ctx), req)
}
