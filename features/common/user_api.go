package common

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// This is for document purpurse, you can always use concrete struct instead of this interface
// Ctx should have token of school admin already
type UserSuite interface {
	CreateTeacher(ctx context.Context, schoolID int64) (profile *cpb.BasicProfile, exchangedToken string, err error)
	CreateStudent(ctx context.Context, locIDs []string) (*upb.Student, error)
	CreateStudentWithParent(ctx context.Context, locIDs []string) (*upb.Student, *upb.Parent, error)
	CreateParentForStudent(ctx context.Context, studentID string) (*upb.Parent, error)
}

type CreateStudentWithParentOpt struct {
	StudentName string
	ParentName  string
}

func (s *suite) newParentInfo(name string) *upb.CreateParentsAndAssignToStudentRequest_ParentProfile {
	randomStr := idutil.ULIDNow()

	if name == "" {
		name = fmt.Sprintf("parent-%s %s", randomStr, randomStr)
	}

	return &upb.CreateParentsAndAssignToStudentRequest_ParentProfile{
		Email:        fmt.Sprintf("parent-%s@gmail.com", randomStr),
		Password:     randomStr,
		Name:         name,
		PhoneNumber:  fmt.Sprintf("+84%d", rand.Int()),
		CountryCode:  cpb.Country_COUNTRY_VN,
		Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
	}
}

func (s *suite) newstudentInfo(schoolID int32, locIDs []string, optP *CreateStudentOpt) *upb.CreateStudentRequest {
	opt := CreateStudentOpt{}
	if optP != nil {
		opt = *optP
	}

	randomID := idutil.ULIDNow()
	password := fmt.Sprintf("password-%v", randomID)
	email := fmt.Sprintf("%v@example.com", randomID)
	name := "student-" + randomID
	if opt.Name != "" {
		name = opt.Name
	}
	req := &upb.CreateStudentRequest{
		SchoolId: schoolID,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:            email,
			Password:         password,
			Name:             name,
			CountryCode:      cpb.Country_COUNTRY_VN,
			PhoneNumber:      fmt.Sprintf("phone-number-%v", randomID),
			Grade:            5,
			EnrollmentStatus: upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			LocationIds:      locIDs,
		},
	}
	return req
}

func (s *suite) CreateStudentByStudentInfo(ctx context.Context, req *upb.CreateStudentRequest) (*upb.CreateStudentResponse_StudentProfile, error) {
	res, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(s.SignedCtx(ctx), req)
	if err != nil {
		return nil, err
	}

	stu := res.StudentProfile
	return stu, nil
}

func (s *suite) CreateStudent(ctx context.Context, locIDs []string, opt *CreateStudentOpt) (*upb.Student, error) {
	intschool := int32ResourcePathFromCtx(ctx)
	req := s.newstudentInfo(intschool, locIDs, opt)

	res, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(s.SignedCtx(ctx), req)
	if err != nil {
		return nil, err
	}

	stu := res.StudentProfile.Student
	return stu, nil
}

type CreateStudentOpt struct {
	Name string
}

func (s *suite) CreateParentForStudent(ctx context.Context, studentID string) (*upb.Parent, error) {
	intschool := int32ResourcePathFromCtx(ctx)
	parent := s.newParentInfo("")
	req := &upb.CreateParentsAndAssignToStudentRequest{
		SchoolId:       intschool,
		StudentId:      studentID,
		ParentProfiles: []*upb.CreateParentsAndAssignToStudentRequest_ParentProfile{parent},
	}
	res, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(s.SignedCtx(ctx), req)
	if err != nil {
		return nil, err
	}
	return res.ParentProfiles[0].Parent, nil
}

// Note that this will remove current parent of student and replace with the new one
func (s *suite) UpdateStudentParent(ctx context.Context, studentID string, parentID, parentEmail string) error {
	intschool := int32ResourcePathFromCtx(ctx)

	req := &upb.UpdateParentsAndFamilyRelationshipRequest{
		SchoolId:  intschool,
		StudentId: studentID,
		ParentProfiles: []*upb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile{
			{
				Id:           parentID,
				Email:        parentEmail,
				Relationship: upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
			},
		},
	}

	_, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).UpdateParentsAndFamilyRelationship(s.SignedCtx(ctx), req)
	return err
}

func (s *suite) CreateStudentWithParent(ctx context.Context, locIDs []string, optP *CreateStudentWithParentOpt) (*upb.Student, *upb.Parent, error) {
	opt := CreateStudentWithParentOpt{}
	if optP != nil {
		opt = *optP
	}

	intschool := int32ResourcePathFromCtx(ctx)
	req := s.newstudentInfo(intschool, locIDs, &CreateStudentOpt{
		Name: opt.StudentName,
	})

	res, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).CreateStudent(s.SignedCtx(ctx), req)
	if err != nil {
		return nil, nil, err
	}

	parent := s.newParentInfo(opt.ParentName)
	req2 := &upb.CreateParentsAndAssignToStudentRequest{
		SchoolId:       intschool,
		StudentId:      res.StudentProfile.Student.UserProfile.UserId,
		ParentProfiles: []*upb.CreateParentsAndAssignToStudentRequest_ParentProfile{parent},
	}
	res2, err := upb.NewUserModifierServiceClient(s.UserMgmtConn).CreateParentsAndAssignToStudent(s.SignedCtx(ctx), req2)
	if err != nil {
		return nil, nil, err
	}
	stu := res.StudentProfile.Student
	par := res2.ParentProfiles[0].Parent
	return stu, par, nil
}

func (s *suite) CreateTeacher(ctx context.Context) (*upb.CreateStaffResponse_StaffProfile, string, error) {
	schoolInt := resourcePathFromCtx(ctx)
	random := idutil.ULIDNow()
	teacherUserGroup, err := (&repository.UserGroupV2Repo{}).FindUserGroupByRoleName(ctx, s.BobDBTrace, constant.RoleTeacher)
	if err != nil {
		return nil, "", fmt.Errorf("FindUserGroupByRoleName %w", err)
	}
	req := &upb.CreateStaffRequest{
		Staff: &upb.CreateStaffRequest_StaffProfile{
			Name:           fmt.Sprintf("new-teacher-%s", random),
			OrganizationId: schoolInt,
			UserGroup:      upb.UserGroup_USER_GROUP_TEACHER,
			Country:        cpb.Country_COUNTRY_VN,
			Email:          fmt.Sprintf("create_staff+%s@gmail.com", random),
			LocationIds:    []string{s.DefaultLocationID},
			UserGroupIds:   []string{teacherUserGroup.UserGroupID.String},
		},
	}

	resp, err := s.CreateTeacherByTeacherInfo(ctx, req)

	if err != nil {
		return nil, "", fmt.Errorf("CreateTeacherByTeacherInfo %w", err)
	}

	tok, err := s.GenerateExchangeTokenCtx(ctx, resp.Staff.StaffId, cpb.UserGroup_USER_GROUP_TEACHER.String())
	if err != nil {
		return nil, "", fmt.Errorf("GenerateExchangeTokenCtx %w", err)
	}
	return resp.Staff, tok, nil
}

func (s *suite) CreateTeacherWithUserGroups(ctx context.Context, userGroupIDs []string) (*upb.CreateStaffResponse_StaffProfile, string, error) {
	schoolInt := resourcePathFromCtx(ctx)
	random := idutil.ULIDNow()
	req := &upb.CreateStaffRequest{
		Staff: &upb.CreateStaffRequest_StaffProfile{
			Name:           fmt.Sprintf("new-teacher-%s", random),
			OrganizationId: schoolInt,
			UserGroup:      upb.UserGroup_USER_GROUP_TEACHER,
			Country:        cpb.Country_COUNTRY_VN,
			Email:          fmt.Sprintf("create_staff+%s@gmail.com", random),
			LocationIds:    []string{s.DefaultLocationID},
			UserGroupIds:   userGroupIDs,
		},
	}

	resp, err := s.CreateTeacherByTeacherInfo(ctx, req)

	if err != nil {
		return nil, "", fmt.Errorf("CreateTeacherByTeacherInfo %w", err)
	}

	tok, err := s.GenerateExchangeTokenCtx(ctx, resp.Staff.StaffId, cpb.UserGroup_USER_GROUP_TEACHER.String())
	if err != nil {
		return nil, "", fmt.Errorf("GenerateExchangeTokenCtx %w", err)
	}
	return resp.Staff, tok, nil
}

func (s *suite) CreateTeacherByTeacherInfo(ctx context.Context, req *upb.CreateStaffRequest) (*upb.CreateStaffResponse, error) {
	// Create new teacher using CreateStaff API
	resp, err := upb.NewStaffServiceClient(s.UserMgmtConn).CreateStaff(contextWithToken(s, ctx), req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *suite) UpsertStudentCoursePackages(ctx context.Context, courseID, locationID, studentID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()
	profiles := []*upb.UpsertStudentCoursePackageRequest_StudentPackageProfile{
		{
			Id: &upb.UpsertStudentCoursePackageRequest_StudentPackageProfile_CourseId{
				CourseId: courseID,
			},
			StartTime:   timestamppb.New(now.Add(-24 * time.Hour)),
			EndTime:     timestamppb.New(now.Add(-24 * time.Hour).Add(24 * 7 * time.Hour)),
			LocationIds: []string{locationID},
		},
	}
	req := &upb.UpsertStudentCoursePackageRequest{
		StudentId:              studentID,
		StudentPackageProfiles: profiles,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = upb.NewUserModifierServiceClient(s.UserMgmtConn).
		UpsertStudentCoursePackage(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}
