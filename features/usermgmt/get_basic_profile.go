package usermgmt

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userCanNotGetBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return ctx, errors.New("expected response has err but actual is nil")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := genGetBasicProfileRequest([]string{s.CurrentUserID})
	stepState.Request = req

	resp, err := getBasicProfile(ctx, s.UserMgmtConn, req)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetBasicProfileWithInvalidRequest(ctx context.Context, invalidType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := genGetBasicProfileRequest([]string{s.CurrentUserID})

	switch invalidType {
	case "invalid user id":
		ctx = contextWithTokenV2(ctx, invalidToken)
		req.UserIds = []string{"invalid user_id"}
	case "missing token":
		ctx = contextWithTokenV2(ctx, "")
	}

	stepState.Request = req
	_, err := getBasicProfile(ctx, s.UserMgmtConn, req)
	if err != nil {
		stepState.ResponseErr = err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceiveBasicProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("expected response has no err but got :%w", stepState.ResponseErr)
	}

	req := stepState.Request.(*pb.GetBasicProfileRequest)
	userIDs := req.UserIds
	if len(userIDs) == 0 {
		// if userIDs in request empty, get user_id from ctx
		userIDs = []string{interceptors.UserIDFromContext(ctx)}
	}

	resp := stepState.Response.(*pb.GetBasicProfileResponse)
	if len(userIDs) != len(resp.Profiles) {
		return ctx, fmt.Errorf("expected return %d profiles, but got: %d", len(userIDs), len(resp.Profiles))
	}

	schoolID, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return ctx, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	school, err := (&repository.SchoolRepo{}).Find(ctx, s.BobDB, database.Int4(int32(schoolID)))
	if err != nil {
		return ctx, status.Error(codes.Internal, err.Error())
	}

	userRepo := repository.UserRepo{}
	users, err := userRepo.Retrieve(ctx, s.BobDB, database.TextArray(userIDs))
	if err != nil {
		return ctx, fmt.Errorf("userRepo.GetProfile: %w", err)
	}

	for _, user := range users {
		userGroupV2, err := getUserGroupV2(ctx, s.BobDB, user.ID.String)
		if err != nil {
			return ctx, status.Error(codes.Internal, err.Error())
		}

		userProfile := toUserBasicProfile(user, school, userGroupV2)
		for _, profile := range resp.Profiles {
			if user.ID.String == profile.UserId {
				if err := compareProfile(userProfile, profile); err != nil {
					return ctx, fmt.Errorf("profile not match: %w", err)
				}
				break
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func getBasicProfile(ctx context.Context, userClientConn *grpc.ClientConn, req *pb.GetBasicProfileRequest) (*pb.GetBasicProfileResponse, error) {
	return pb.NewUserReaderServiceClient(userClientConn).GetBasicProfile(ctx, req)
}

func genGetBasicProfileRequest(userIDs []string) *pb.GetBasicProfileRequest {
	return &pb.GetBasicProfileRequest{
		UserIds: userIDs,
	}
}

func toUserBasicProfile(user *entity.LegacyUser, school *entity.School, userGroupV2 []*pb.BasicProfile_UserGroup) *pb.BasicProfile {
	profile := &pb.BasicProfile{
		UserId:    user.ID.String,
		Name:      user.GetName(),
		Email:     user.Email.String,
		Avatar:    user.Avatar.String,
		UserGroup: user.Group.String,
		Country:   cpb.Country(cpb.Country_value[user.Country.String]),
		School: &pb.BasicProfile_School{
			SchoolId:   int64(school.ID.Int),
			SchoolName: school.Name.String,
		},
		UserGroupV2:   userGroupV2,
		CreatedAt:     timestamppb.New(user.CreatedAt.Time),
		LastLoginDate: timestamppb.New(user.LastLoginDate.Time),
		FirstName:     user.FirstName.String,
		LastName:      user.LastName.String,
	}

	return profile
}

func compareProfile(expected, actual *pb.BasicProfile) error {
	switch {
	case expected.UserId != actual.UserId:
		return fmt.Errorf(`expected user_id: %s but actual: %s`, expected.UserId, actual.UserId)
	case expected.Name != actual.Name:
		return fmt.Errorf(`expected name: %s but actual: %s`, expected.Name, actual.Name)
	case expected.Email != actual.Email:
		return fmt.Errorf(`expected email: %s but actual: %s`, expected.Email, actual.Email)
	case expected.Avatar != actual.Avatar:
		return fmt.Errorf(`expected avatar: %s but actual: %s`, expected.Avatar, actual.Avatar)
	case expected.UserGroup != actual.UserGroup:
		return fmt.Errorf(`expected user_group: %s but actual: %s`, expected.UserGroup, actual.UserGroup)
	case expected.Country.String() != actual.Country.String():
		return fmt.Errorf(`expected country: %s but actual: %s`, expected.Country.String(), actual.Country.String())
	case expected.School.SchoolId != actual.School.SchoolId:
		return fmt.Errorf(`expected school_id: %d but actual: %d`, expected.School.SchoolId, actual.School.SchoolId)
	case expected.CreatedAt.AsTime() != actual.CreatedAt.AsTime():
		return fmt.Errorf(`expected created_at: %s but actual: %s`, expected.CreatedAt, actual.CreatedAt)
	case expected.LastLoginDate.AsTime() != actual.LastLoginDate.AsTime():
		return fmt.Errorf(`expected last_login_date: %s but actual: %s`, expected.CreatedAt, actual.LastLoginDate)
	case expected.FirstName != actual.FirstName:
		return fmt.Errorf(`expected first name: %s but actual: %s`, expected.FirstName, actual.FirstName)
	case expected.LastName != actual.LastName:
		return fmt.Errorf(`expected last name: %s but actual: %s`, expected.LastName, actual.LastName)
	}
	for idx := range expected.UserGroupV2 {
		if expected.UserGroupV2[idx].UserGroup != actual.UserGroupV2[idx].UserGroup {
			return fmt.Errorf(`expected user_group_v2: %s but actual: %s`, expected.UserGroupV2[idx].UserGroup, actual.UserGroupV2[idx].UserGroup)
		}
		if len(expected.UserGroupV2[idx].Roles) != len(actual.UserGroupV2[idx].Roles) {
			return fmt.Errorf(`expected roles: %s but actual: %s`, expected.UserGroupV2[idx].Roles, actual.UserGroupV2[idx].Roles)
		}
		if expected.UserGroupV2[idx].UserGroupId != actual.UserGroupV2[idx].UserGroupId {
			return fmt.Errorf(`expected UserGroupId: %s but actual: %s`, expected.UserGroupV2[idx].UserGroupId, actual.UserGroupV2[idx].UserGroupId)
		}
		for roleIndex := range expected.UserGroupV2[idx].Roles {
			if expected.UserGroupV2[idx].Roles[roleIndex].RoleId != actual.UserGroupV2[idx].Roles[roleIndex].RoleId {
				return fmt.Errorf(`expected RoleId: %s but actual: %s`, expected.UserGroupV2[idx].Roles[roleIndex].RoleId, actual.UserGroupV2[idx].Roles[roleIndex].RoleId)
			}
			if expected.UserGroupV2[idx].Roles[roleIndex].Role != actual.UserGroupV2[idx].Roles[roleIndex].Role {
				return fmt.Errorf(`expected Role: %s but actual: %s`, expected.UserGroupV2[idx].Roles[roleIndex].Role, actual.UserGroupV2[idx].Roles[roleIndex].Role)
			}
		}
	}
	return nil
}

func getUserGroupV2(ctx context.Context, db database.Ext, userID string) ([]*pb.BasicProfile_UserGroup, error) {
	userReaderService := service.UserReaderService{
		DB: &database.DBTrace{
			DB: db,
		},
		UserGroupV2Repo: &repository.UserGroupV2Repo{},
	}
	userGroupV2, err := userReaderService.GetUserGroupV2(ctx, userID)
	if err != nil {
		return nil, err
	}
	return userGroupV2, nil
}
