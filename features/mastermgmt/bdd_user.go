package mastermgmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/usermgmt"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
)

const (
	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"
)

func (s *suite) signedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		userGroup string
		err       error
	)
	switch group {
	case unauthenticatedType:
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case "staff granted role school admin":
		return s.aSignedInStaff(ctx, []string{constant.RoleSchoolAdmin})
	case "staff granted role teacher":
		return s.aSignedInStaff(ctx, []string{constant.RoleTeacher})
	// case "staff granted role school admin and teacher":
	// 	return s.aSignedInStaff(ctx, []string{constant.RoleSchoolAdmin, constant.RoleTeacher})
	case "staff granted role hq staff":
		return s.aSignedInStaff(ctx, []string{constant.RoleHQStaff})
	case "staff granted role centre lead":
		return s.aSignedInStaff(ctx, []string{constant.RoleCentreLead})
	case "staff granted role centre manager":
		return s.aSignedInStaff(ctx, []string{constant.RoleCentreManager})
	case "staff granted role centre staff":
		return s.aSignedInStaff(ctx, []string{constant.RoleCentreStaff})
	case studentType:
		userGroup = constant.UserGroupStudent
	case teacherType:
		userGroup = constant.UserGroupTeacher
	case schoolAdminType:
		userGroup = constant.UserGroupSchoolAdmin
	case parentType:
		userGroup = constant.UserGroupParent
	case organizationType:
		userGroup = constant.UserGroupOrganizationManager
	}

	id := newID()
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = userGroup

	if ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(userGroup)); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	token, err := s.generateExchangeToken(id, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateExchangeToken: %v", err)
	}
	stepState.AuthToken = token

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signedAsAccountV2(ctx context.Context, account string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	roleWithLocation := usermgmt.RoleWithLocation{}

	switch account {
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
	case "student":
		roleWithLocation.RoleName = constant.RoleStudent
	case schoolAdminType:
		roleWithLocation.RoleName = constant.RoleSchoolAdmin
	case teacherType:
		roleWithLocation.RoleName = constant.RoleTeacher
	}
	roleWithLocation.LocationIDs = []string{constants.ManabieOrgLocation}

	authInfo, err := usermgmt.SignIn(ctx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.FirebaseAddress, s.UserMgmtConn, roleWithLocation, []string{constants.ManabieOrgLocation})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = constants.ManabieOrgLocation
	return StepStateToContext(ctx, stepState), nil
}

func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot generate new user token, err: %v", err)
	}

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()

	return string(bodyResp), nil
}

func (s *suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}
	// ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	user, err := newUserEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newUserEntity")
	}

	for _, opt := range opts {
		opt(user)
	}

	err = database.ExecInTx(ctx, s.BobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := repository.UserRepo{}
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}

		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := repository.TeacherRepo{}
			t := &entity.Teacher{}
			database.AllNullEntity(t)
			t.ID = user.ID
			err := multierr.Combine(
				t.SchoolIDs.Set([]int64{schoolID}),
				t.ResourcePath.Set(fmt.Sprint(schoolID)),
			)
			if err != nil {
				return err
			}

			err = teacherRepo.CreateMultiple(ctx, tx, []*entity.Teacher{t})
			if err != nil {
				return fmt.Errorf("cannot create teacher: %w", err)
			}
		case constant.UserGroupSchoolAdmin:
			schoolAdminRepo := repository.SchoolAdminRepo{}
			schoolAdminAccount := &entity.SchoolAdmin{}
			database.AllNullEntity(schoolAdminAccount)
			err := multierr.Combine(
				schoolAdminAccount.SchoolAdminID.Set(user.ID.String),
				schoolAdminAccount.SchoolID.Set(schoolID),
				schoolAdminAccount.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
			)
			if err != nil {
				return fmt.Errorf("cannot create school admin: %w", err)
			}
			err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdminAccount})
			if err != nil {
				return err
			}
		case constant.UserGroupParent:
			parentRepo := repository.ParentRepo{}
			parentEnt := &entity.Parent{}
			database.AllNullEntity(parentEnt)
			err := multierr.Combine(
				parentEnt.ID.Set(user.ID.String),
				parentEnt.SchoolID.Set(schoolID),
				parentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
			)
			if err != nil {
				return err
			}
			err = parentRepo.CreateMultiple(ctx, tx, []*entity.Parent{parentEnt})
			if err != nil {
				return fmt.Errorf("cannot create parent: %w", err)
			}
		case constant.UserGroupStudent:
			studentRepo := repository.StudentRepo{}
			student, err := newStudentEntity()
			if err != nil {
				return err
			}
			err = multierr.Combine(
				student.ID.Set(user.ID.String),
				student.SchoolID.Set(schoolID),
				student.ResourcePath.Set(fmt.Sprint(schoolID)),
			)
			if err != nil {
				return err
			}
			err = studentRepo.CreateMultiple(ctx, tx, []*entity.LegacyStudent{student})
			if err != nil {
				return fmt.Errorf("cannot create student: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	uGroup := &entity.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(user.Group.String),
		uGroup.UserID.Set(user.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
		uGroup.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userGroupRepo := &repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDBTrace, uGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupRepo.Upsert: %w %s", err, user.Group.String)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, studentType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aSignedInTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, teacherType)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	studentRepo := repository.StudentRepo{}
	student, err := newStudentEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newStudentEntity")
	}
	if err := multierr.Combine(
		student.ID.Set(id),
		student.ResourcePath.Set(fmt.Sprint(schoolID)),
		student.LegacyUser.ID.Set(id),
		student.LegacyUser.ResourcePath.Set(fmt.Sprint(schoolID)),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = studentRepo.Create(golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID)), s.BobDBTrace, student)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStaffInDB(ctx context.Context, db *database.DBTrace, id string) error {
	now := time.Now()
	staffEntity := &entity.Staff{}
	database.AllNullEntity(staffEntity)
	ctx, err := s.aValidUser(ctx, withID(id), withRole(constant.UserGroupAdmin))
	if err != nil {
		return fmt.Errorf("aValidStaffInDB. s.aValidUser: %w", err)
	}
	if err := multierr.Combine(
		staffEntity.ID.Set(id),
		staffEntity.CreatedAt.Set(now),
		staffEntity.UpdatedAt.Set(now),
		staffEntity.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		staffEntity.AutoCreateTimesheet.Set(false),
		staffEntity.WorkingStatus.Set(pb.StaffWorkingStatus_AVAILABLE.String()),
	); err != nil {
		return fmt.Errorf("aValidStaffInDB multierr.Combine: %w", err)
	}
	cmdTag, err := database.Insert(ctx, staffEntity, db.Exec)
	if err != nil {
		return fmt.Errorf("err insert staff: %w", err)
	}
	if cmdTag.RowsAffected() != 1 {
		return fmt.Errorf("err insert staff: %d RowsAffected", cmdTag.RowsAffected())
	}

	return nil
}

func (s *suite) loginFirebaseAccount(ctx context.Context, email string, password string) error {
	return LoginFirebaseAccount(ctx, s.Cfg.FirebaseAPIKey, email, password)
}

func LoginFirebaseAccount(ctx context.Context, apiKey string, email string, password string) error {
	url := fmt.Sprintf("%s%s", IdentityToolkitURL, apiKey)

	loginInfo := struct {
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to login firebase and failed to decode error")
	}
	return errors.New("failed to login firebase" + string(data))
}

func (s *suite) loginIdentityPlatform(ctx context.Context, tenantID string, email string, password string) error {
	return LoginIdentityPlatform(ctx, s.Cfg.FirebaseAPIKey, tenantID, email, password)
}

const IdentityToolkitURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="

func LoginIdentityPlatform(ctx context.Context, apiKey string, tenantID string, email string, password string) error {
	url := fmt.Sprintf("%s%s", IdentityToolkitURL, apiKey)

	loginInfo := struct {
		TenantID          string `json:"tenantId"`
		Email             string `json:"email"`
		Password          string `json:"password"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		TenantID:          tenantID,
		Email:             email,
		Password:          password,
		ReturnSecureToken: true,
	}
	body, err := json.Marshal(&loginInfo)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to login identity platform: and failed to decode error")
	}
	return errors.New("failed to login identity platform:" + string(data))
}

func (s *suite) createSampleUserGroups(ctx context.Context, amountSampleUserGroupIDs int, roles []string) ([]string, error) {
	stepState := StepStateFromContext(ctx)
	var userGroupIDs []string
	// we have to sign in as admin to create staff, and have to store old auth token
	previousAuth := stepState.AuthToken

	ctx, err := s.signedAsAccount(ctx, schoolAdminType)
	if err != nil {
		return nil, fmt.Errorf("signedAsAccount: %w", err)
	}

	for i := 0; i < amountSampleUserGroupIDs; i++ {
		resp, err := s.createUserGroupWithRoleNames(ctx, roles)
		if err != nil {
			return nil, fmt.Errorf("signedInAndcreateUserGroupWithValidityPayload: %w", err)
		}

		userGroupIDs = append(userGroupIDs, resp.UserGroupId)
	}

	// restore previous auth token
	stepState.AuthToken = previousAuth
	return userGroupIDs, nil
}

func (s *suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string) (*pb.CreateUserGroupResponse, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", newID()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.BobDB.Query(ctx, stmt, roleNames, len(roleNames))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return nil, fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	for _, roleID := range roleIDs {
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&pb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: []string{constants.ManabieOrgLocation},
			},
		)
	}

	resp, err := pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).CreateUserGroup(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return nil, fmt.Errorf("createUserGroupWithRoleNames: %w", err)
	}

	return resp, nil
}

func (s *suite) aSignedInStaff(ctx context.Context, roles []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := newID()
	ctx, err := s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(constant.UserGroupSchoolAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser: %w", err)
	}

	token, err := s.generateExchangeToken(id, constant.UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken = token

	createUserGroupResp, err := s.createUserGroupWithRoleNames(ctx, roles)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if err := assignUserGroupToUser(ctx, s.BobDBTrace, id, []string{createUserGroupResp.UserGroupId}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupSchoolAdmin
	return StepStateToContext(ctx, stepState), nil
}

func SeedUserGroup(ctx context.Context, dbBob database.QueryExecer, usermgmtConn *grpc.ClientConn, roles []string) (*pb.CreateUserGroupResponse, error) {
	query := `SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2`
	rows, err := dbBob.Query(ctx, query, roles, len(roles))
	if err != nil {
		return nil, fmt.Errorf("BobDBTrace.Query: Find Role IDs: %w", err)
	}
	defer rows.Close()

	req := &pb.CreateUserGroupRequest{
		UserGroupName: "UserGroupName",
	}

	roleID := ""
	for rows.Next() {
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}

		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&pb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: []string{constants.ManabieOrgLocation},
			},
		)
	}

	resp, err := pb.NewUserGroupMgmtServiceClient(usermgmtConn).CreateUserGroup(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *suite) validateUsersHasUserGroupWithRole(ctx context.Context, userIDs []string, resourcePath, roleName string) error {
	// query count amount of user group members with role name and resourcePath
	expectedCountedUserGroupMembers := len(userIDs)
	fieldNames, _ := new(entity.UserGroupMember).FieldMap()
	for index := 0; index < len(fieldNames); index++ {
		fieldNames[index] = "ugm." + fieldNames[index]
	}
	stmt := `
			SELECT
				%s

			FROM
				user_group_member ugm

			INNER JOIN granted_role gr ON
				ugm.user_group_id = gr.user_group_id AND
				gr.deleted_at IS NULL

			INNER JOIN role r ON
				r.role_id = gr.role_id AND
				r.deleted_at IS NULL

			WHERE
				ugm.user_id = ANY($1) AND
				r.role_name = $2 AND
				ugm.deleted_at IS NULL
	`
	rows, err := s.BobDBTrace.Query(ctx, fmt.Sprintf(stmt, strings.Join(fieldNames, ", ")), database.TextArray(userIDs), roleName)
	if err != nil {
		return fmt.Errorf(`query error when finding user group member: %v`, err)
	}
	defer rows.Close()
	if err := rows.Err(); err != nil {
		return fmt.Errorf(`rows error when finding user group member: %v`, err)
	}
	userGroupMembers := []*entity.UserGroupMember{}
	for rows.Next() {
		userGroupMember := new(entity.UserGroupMember)
		_, fields := userGroupMember.FieldMap()
		if err := rows.Scan(fields...); err != nil {
			return fmt.Errorf(`scan error when finding user group member: %v`, err)
		}
		userGroupMembers = append(userGroupMembers, userGroupMember)
	}

	// assert count passed userids and amount of user group members existed in DB
	countedUserGroupMembers := len(userGroupMembers)
	if expectedCountedUserGroupMembers != countedUserGroupMembers {
		return fmt.Errorf(`expect %d user group members, but got %d`, expectedCountedUserGroupMembers, countedUserGroupMembers)
	}

	for _, userGroupMember := range userGroupMembers {
		if userGroupMember.ResourcePath.String != resourcePath {
			return fmt.Errorf(`ResourcePath of user %s: expect %s, but got %s`, userGroupMember.UserID.String, resourcePath, userGroupMember.ResourcePath.String)
		}
	}
	return nil
}

func newID() string {
	return idutil.ULIDNow()
}

type userOption func(u *entity.LegacyUser)

func withID(id string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entity.LegacyUser) {
		_ = u.Group.Set(group)
	}
}

func (s *suite) generateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", fmt.Errorf("error when create generateValidAuthenticationToken: %v", err)
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.ApplicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("error when create exchange token: %v", err)
	}
	return token, nil
}

func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func newUserEntity() (*entity.LegacyUser, error) {
	userID := newID()
	now := time.Now()
	user := new(entity.LegacyUser)
	firstName := fmt.Sprintf("user-first-name-%s", userID)
	lastName := fmt.Sprintf("user-last-name-%s", userID)
	fullName := helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)
	database.AllNullEntity(user)
	database.AllNullEntity(&user.AppleUser)
	if err := multierr.Combine(
		user.ID.Set(userID),
		user.Email.Set(fmt.Sprintf("valid-user-%s@email.com", userID)),
		user.Avatar.Set(fmt.Sprintf("http://valid-user-%s", userID)),
		user.IsTester.Set(false),
		user.FacebookID.Set(userID),
		user.PhoneVerified.Set(false),
		user.AllowNotification.Set(true),
		user.EmailVerified.Set(false),
		user.FullName.Set(fullName),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Country.Set(cpb.Country_COUNTRY_VN.String()),
		user.Group.Set(entity.UserGroupStudent),
		user.Birthday.Set(now),
		user.Gender.Set(pb.Gender_FEMALE.String()),
		user.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		user.CreatedAt.Set(now),
		user.UpdatedAt.Set(now),
		user.DeletedAt.Set(nil),
	); err != nil {
		return nil, errors.Wrap(err, "set value user")
	}

	user.UserAdditionalInfo = entity.UserAdditionalInfo{
		CustomClaims: map[string]interface{}{
			"external-info": "example-info",
		},
	}
	return user, nil
}

func newStudentEntity() (*entity.LegacyStudent, error) {
	now := time.Now()
	student := new(entity.LegacyStudent)
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	user, err := newUserEntity()
	if err != nil {
		return nil, errors.Wrap(err, "newUserEntity")
	}
	student.LegacyUser = *user
	schoolID, err := strconv.ParseInt(student.LegacyUser.ResourcePath.String, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseInt")
	}

	if err := multierr.Combine(
		student.ID.Set(student.LegacyUser.ID),
		student.SchoolID.Set(schoolID),
		student.EnrollmentStatus.Set(pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_ENROLLED),
		student.StudentExternalID.Set(student.LegacyUser.ID),
		student.StudentNote.Set("this is the note"),
		student.CurrentGrade.Set(1),
		student.TargetUniversity.Set("HUST"),
		student.TotalQuestionLimit.Set(32),
		student.OnTrial.Set(false),
		student.BillingDate.Set(now),
		student.CreatedAt.Set(student.LegacyUser.CreatedAt),
		student.UpdatedAt.Set(student.LegacyUser.UpdatedAt),
		student.DeletedAt.Set(student.LegacyUser.DeletedAt),
		student.PreviousGrade.Set(12),
	); err != nil {
		return nil, errors.Wrap(err, "set value student")
	}

	return student, nil
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)
	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
		); err != nil {
			return err
		}
		userGroupMembers = append(userGroupMembers, userGroupMem)
	}

	if err := (&repository.UserGroupsMemberRepo{}).UpsertBatch(ctx, dbBob, userGroupMembers); err != nil {
		return errors.Wrapf(err, "assignUserGroupToUser")
	}
	return nil
}
