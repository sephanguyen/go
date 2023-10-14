package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	user_repo "github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"
)

func (s *suite) returnRootContext(ctx context.Context) context.Context {
	return common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
}

func (s *suite) signedAsAccountV2(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	roleWithLocation := usermgmt.RoleWithLocation{}
	adminCtx := s.returnRootContext(ctx)
	switch account {
	case unauthenticatedType:
		stepState.AuthToken = "random-token"
		stepState.CurrentUserID = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case "staff granted role school admin":
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case "staff granted role hq staff":
		roleWithLocation.RoleName = constant.RoleHQStaff
	case "staff granted role centre lead":
		roleWithLocation.RoleName = constant.RoleCentreLead
	case "staff granted role centre manager":
		roleWithLocation.RoleName = constant.RoleCentreManager
	case "staff granted role centre staff":
		roleWithLocation.RoleName = constant.RoleCentreStaff
	case "staff granted role teacher":
		roleWithLocation.RoleName = constant.RoleTeacher
	case "staff granted role teacher lead":
		roleWithLocation.RoleName = constant.RoleTeacherLead
	case studentType:
		roleWithLocation.RoleName = constant.RoleStudent
	case schoolAdminType:
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case teacherType:
		roleWithLocation.RoleName = constant.RoleTeacher
	case parentType:
		roleWithLocation.RoleName = constant.RoleParent
	}

	roleWithLocation.LocationIDs = []string{constants.ManabieOrgLocation}

	authInfo, err := usermgmt.SignIn(adminCtx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.CommonSuite.StepState.FirebaseAddress, s.Connections.UserMgmtConn, roleWithLocation, []string{constants.ManabieOrgLocation})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = constants.ManabieOrgLocation

	if account == studentType {
		stepState.CurrentStudentID = authInfo.UserID
	} else if account == teacherType {
		stepState.CurrentTeacherID = authInfo.UserID
	}

	ctx = common.ValidContext(ctx, constants.ManabieSchool, authInfo.UserID, authInfo.Token)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInAsSchoolAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccountV2(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInTeacherV2(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccountV2(ctx, "staff granted role teacher")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when create teacher %v", err)
	}
	stepState.CurrentTeacherID = stepState.CurrentUserID
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserWithRole(ctx context.Context, role string) (*common.AuthInfo, error) {
	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)

	roleWithLocation := usermgmt.RoleWithLocation{}
	roleWithLocation.RoleName = role
	roleWithLocation.LocationIDs = []string{constants.ManabieOrgLocation}

	authInfo, err := usermgmt.SignIn(ctx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.CommonSuite.StepState.FirebaseAddress, s.Connections.UserMgmtConn, roleWithLocation, []string{constants.ManabieOrgLocation})

	if err != nil {
		return nil, err
	}

	return &authInfo, nil
}

func (s *suite) CreateStudentAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentOne, err := s.createUserWithRole(ctx, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentIds = append(stepState.StudentIds, studentOne.UserID)

	studentTwo, err := s.createUserWithRole(ctx, constant.RoleStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentIds = append(stepState.StudentIds, studentTwo.UserID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateStudentNumberAccounts(ctx context.Context, studentCount string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	count, err := strconv.Atoi(studentCount)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert student count to a number: %w", err)
	}

	for i := 0; i < count; i++ {
		student, err := s.createUserWithRole(ctx, constant.RoleStudent)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds = append(stepState.StudentIds, student.UserID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateStudentNumberAccountsWithName(ctx context.Context, studentCount, firstName, lastName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	count, err := strconv.Atoi(studentCount)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert student count to a number: %w", err)
	}

	stepState.StudentIds = []string{}
	for i := 0; i < count; i++ {
		studentID, err := s.CreateStudentWithName(ctx, firstName, lastName)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds = append(stepState.StudentIds, studentID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateStudentWithName(ctx context.Context, firstName, lastName string) (string, error) {
	req := s.CreateStudentRequestWithName([]string{constants.ManabieOrgLocation}, firstName, lastName)
	resp, err := usermgmt.CreateStudent(ctx, s.Connections.UserMgmtConn, req, []string{constants.ManabieOrgLocation})
	if err != nil {
		return "", fmt.Errorf("failed to create student: %w", err)
	}

	return resp.StudentProfile.Student.UserProfile.UserId, nil
}

func (s *suite) CreateStudentRequestWithName(locationIDs []string, firstName, lastName string) *upb.CreateStudentRequest {
	randomID := s.NewID()
	req := &upb.CreateStudentRequest{
		SchoolId: constants.ManabieSchool,
		StudentProfile: &upb.CreateStudentRequest_StudentProfile{
			Email:             fmt.Sprintf("%v@example.com", randomID),
			Password:          fmt.Sprintf("password-%v", randomID),
			Name:              fmt.Sprintf("%s %s", firstName, lastName),
			FirstName:         firstName,
			LastName:          lastName,
			CountryCode:       cpb.Country_COUNTRY_VN,
			EnrollmentStatus:  upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED,
			PhoneNumber:       fmt.Sprintf("phone-number-%v", randomID),
			StudentExternalId: fmt.Sprintf("student-external-id-%v", randomID),
			StudentNote:       fmt.Sprintf("some random student note %v", randomID),
			Grade:             5,
			Birthday:          timestamppb.New(time.Now().Add(-87600 * time.Hour)),
			Gender:            upb.Gender_MALE,
			LocationIds:       locationIDs,
		},
	}

	return req
}

func (s *suite) CreateTeacherAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{}
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	teacherOne, err := s.createUserWithRole(ctx, constant.RoleTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacherOne.UserID)

	teacherTwo, err := s.createUserWithRole(ctx, constant.RoleTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacherTwo.UserID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateTeacherAccountsWithName(ctx context.Context, firstName, lastName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{}
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constants.ManabieSchool
	}

	teacherOneID, err := s.CreateStaffWithName(ctx, constant.RoleTeacher, firstName, lastName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacherOneID)

	teacherTwoID, err := s.CreateStaffWithName(ctx, constant.RoleTeacher, firstName, lastName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.TeacherIDs = append(stepState.TeacherIDs, teacherTwoID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateStaffWithName(ctx context.Context, role, firstName, lastName string) (string, error) {
	roleWithLocation := usermgmt.RoleWithLocation{}
	roleWithLocation.RoleName = role
	roleWithLocation.LocationIDs = []string{constants.ManabieOrgLocation}

	createUserGroupReq, err := s.CreateUserGroupRequest(ctx, s.BobDBTrace, []usermgmt.RoleWithLocation{roleWithLocation})
	if err != nil {
		return "", fmt.Errorf("failed to create user group request: %w", err)
	}

	createUserGroupResp, err := s.CreateUserGroup(ctx, createUserGroupReq)
	if err != nil {
		return "", fmt.Errorf("failed to create user group: %w", err)
	}

	req := s.CreateStaffRequestWithName([]string{createUserGroupResp.UserGroupId}, roleWithLocation.LocationIDs, firstName, lastName)

	resp, err := usermgmt.CreateStaff(ctx, s.BobDBTrace, s.Connections.UserMgmtConn, req, []usermgmt.RoleWithLocation{roleWithLocation}, roleWithLocation.LocationIDs)
	if err != nil {
		return "", fmt.Errorf("failed to create student: %w", err)
	}

	return resp.Staff.StaffId, nil
}

func (s *suite) CreateStaffRequestWithName(userGroupIDs []string, locationIDs []string, firstName, lastName string) *upb.CreateStaffRequest {
	randomULID := s.NewID()
	staff := &upb.CreateStaffRequest{
		Staff: &upb.CreateStaffRequest_StaffProfile{
			Name:        fmt.Sprintf("%s %s", firstName, lastName),
			Email:       fmt.Sprintf("staff+%s@gmail.com", randomULID),
			Country:     cpb.Country_COUNTRY_VN,
			PhoneNumber: "",
			UserGroup:   upb.UserGroup_USER_GROUP_TEACHER,
			StaffPhoneNumber: []*upb.StaffPhoneNumber{
				{PhoneNumber: "123456789", PhoneNumberType: upb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER},
			},
			Gender:         upb.Gender_MALE,
			Birthday:       timestamppb.New(time.Now().Add(-87600 * 10 * time.Hour)),
			WorkingStatus:  upb.StaffWorkingStatus_AVAILABLE,
			StartDate:      timestamppb.New(time.Now()),
			EndDate:        timestamppb.New(time.Now().Add(87600 * time.Hour)),
			Remarks:        "Hello remarks",
			UserGroupIds:   userGroupIDs,
			LocationIds:    locationIDs,
			ExternalUserId: fmt.Sprintf("external_user_id+%s", randomULID),
			Username:       fmt.Sprintf("username%s", randomULID),
		},
	}

	return staff
}

func (s *suite) CreateUserGroupRequest(ctx context.Context, db database.QueryExecer, roleWithLocations []usermgmt.RoleWithLocation) (*upb.CreateUserGroupRequest, error) {
	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user_group+%s", s.NewID()),
	}

	for _, roleWithLocation := range roleWithLocations {
		if len(roleWithLocation.LocationIDs) == 0 {
			return nil, fmt.Errorf("granted location empty")
		}
		role, err := (&user_repo.RoleRepo{}).GetByName(ctx, db, database.Text(roleWithLocation.RoleName))
		if err != nil {
			return nil, err
		}
		req.RoleWithLocations = append(req.RoleWithLocations, &upb.RoleWithLocations{
			RoleId:      role.RoleID.String,
			LocationIds: roleWithLocation.LocationIDs,
		})
	}

	return req, nil
}

func (s *suite) CreateUserGroup(ctx context.Context, req *upb.CreateUserGroupRequest) (*upb.CreateUserGroupResponse, error) {
	return upb.NewUserGroupMgmtServiceClient(s.Connections.UserMgmtConn).CreateUserGroup(ctx, req)
}
