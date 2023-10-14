package usermgmt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	golibs_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"firebase.google.com/go/auth"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	stubEmail        = "valid-user-%d@email.com"
	stubPhoneNumber  = "+848%d"
	stubLocationName = "location_%s"
)

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

func (s *suite) aValidUserInFatima(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}
	ctx = golibs_auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	user, err := newUserEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newUserEntity")
	}

	for _, opt := range opts {
		opt(user)
	}

	err = database.ExecInTx(ctx, s.FatimaDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		_, err := database.InsertExceptOnConflictDoNothing(ctx, user, []string{"remarks"}, s.FatimaDBTrace.Exec)
		if err != nil {
			return fmt.Errorf("cannot create user Fatima: %w", err)
		}

		switch user.Group.String {
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

func (s *suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}
	user, err := newUserEntity()
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "newUserEntity")
	}

	for _, opt := range opts {
		opt(user)
	}
	e := user.ResourcePath.Set(database.Text(fmt.Sprint(schoolID)))
	if e != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot set resource path: %w", e)
	}
	err = database.ExecInTx(ctx, s.BobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := repository.UserRepo{}
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}
		if user.Group.String != constant.UserGroupSchoolAdmin {
			return fmt.Errorf("only allow to create school admin by this way")
		}

		schoolAdminRepo := repository.SchoolAdminRepo{}
		schoolAdminAccount := &entity.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		if err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(user.ID.String),
			schoolAdminAccount.SchoolID.Set(schoolID),
			schoolAdminAccount.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
		); err != nil {
			return fmt.Errorf("cannot create school admin: %w", err)
		}

		if err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entity.SchoolAdmin{schoolAdminAccount}); err != nil {
			return err
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
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aSignedInTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.signedAsAccount(ctx, teacherType)
	if err != nil {
		return ctx, err
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

func aValidParentInDB(ctx context.Context, db *database.DBTrace, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	parentRepo := repository.ParentRepo{}
	now := time.Now()
	num := rand.Int()
	parent := &entity.Parent{}
	database.AllNullEntity(parent)
	database.AllNullEntity(&parent.LegacyUser)
	firstName := fmt.Sprintf("valid-user-first-name-%d", num)
	lastName := fmt.Sprintf("valid-user-last-name-%d", num)
	err := multierr.Combine(
		parent.LegacyUser.ID.Set(id),
		parent.LegacyUser.Email.Set(fmt.Sprintf(stubEmail, num)),
		parent.LegacyUser.PhoneNumber.Set(fmt.Sprintf(stubPhoneNumber, num)),
		parent.LegacyUser.GivenName.Set(""),
		parent.LegacyUser.Avatar.Set(""),
		parent.LegacyUser.IsTester.Set(false),
		parent.LegacyUser.FacebookID.Set(id),
		parent.LegacyUser.PhoneVerified.Set(false),
		parent.LegacyUser.EmailVerified.Set(false),
		parent.LegacyUser.DeletedAt.Set(nil),
		parent.LegacyUser.LastLoginDate.Set(nil),
		parent.LegacyUser.FullName.Set(helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)),
		parent.LegacyUser.FirstName.Set(firstName),
		parent.LegacyUser.LastName.Set(lastName),
		parent.LegacyUser.Country.Set(cpb.Country_COUNTRY_VN.String()),
		parent.LegacyUser.Group.Set(entity.UserGroupStudent),
		parent.LegacyUser.Birthday.Set(now),
		parent.LegacyUser.Gender.Set(pb.Gender_MALE),
		parent.LegacyUser.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),

		parent.ID.Set(id),
		parent.SchoolID.Set(constants.ManabieSchool),
		parent.CreatedAt.Set(now),
		parent.UpdatedAt.Set(now),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	err = parentRepo.Create(ctx, db, parent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
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

const (
	IdentityToolkitURL     = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key="
	ManabieOrgLocationType = "01FR4M51XJY9E77GSN4QZ1Q9M1"
)

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

func (s *suite) importUserToFirebaseAndIdentityPlatform(ctx context.Context, usr *entity.LegacyUser, rawPassword []byte, passwordSalt []byte) error {
	// We don't auth with phone number fow now
	userToImport := *usr
	if err := userToImport.PhoneNumber.Set(nil); err != nil {
		return err
	}
	userToImport.PasswordSalt = passwordSalt

	hashedPwdForFirebase, err := golibs_auth.HashedPassword(s.FirebaseAuthClient.GetHashConfig(), rawPassword, passwordSalt)
	if err != nil {
		return errors.Wrap(err, "HashedPassword")
	}
	userToImport.PasswordHash = hashedPwdForFirebase

	_, err = tenantClientImportUsers(ctx, s.FirebaseAuthClient, []user.User{&userToImport}, s.FirebaseAuthClient.GetHashConfig())
	if err != nil {
		return errors.Wrap(err, "tenantClientImportUsers")
	}

	// Import user on identity platform
	tenantClient, err := s.TenantManager.TenantClient(ctx, golibs_auth.LocalTenants[constants.ManabieSchool])
	if err != nil {
		return errors.Wrap(err, "TenantClient")
	}

	hashedPwdForIdentityPlatform, err := golibs_auth.HashedPassword(tenantClient.GetHashConfig(), rawPassword, userToImport.PasswordSalt)
	if err != nil {
		return errors.Wrap(err, "HashedPassword")
	}
	userToImport.PasswordHash = hashedPwdForIdentityPlatform

	_, err = tenantClientImportUsers(ctx, tenantClient, []user.User{&userToImport}, tenantClient.GetHashConfig())
	if err != nil {
		return errors.Wrap(err, "tenantClientImportUsers")
	}

	return nil
}

func (s *suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string) (*pb.CreateUserGroupResponse, error) {
	req := &pb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", newID()),
	}
	currentSchoolID := StepStateFromContext(ctx).CurrentSchoolID

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
				LocationIds: []string{GetOrgLocation(int(currentSchoolID))},
			},
		)
	}

	resp, err := pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).CreateUserGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("createUserGroupWithRoleNames: %w", err)
	}

	return resp, nil
}

func (s *suite) aSignedInStaff(ctx context.Context, roles []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := newID()
	ctx, err := s.aValidUser(ctx, withID(id), withRole(constant.UserGroupSchoolAdmin))
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
		return nil, err
	}

	if err := assignUserGroupToUser(ctx, s.BobDBTrace, id, []string{createUserGroupResp.UserGroupId}); err != nil {
		return nil, err
	}

	stepState.CurrentUserID = id
	return ctx, nil
}

func removeUserInFireBase(authClient *auth.Client, userIDs []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, id := range userIDs {
		err := authClient.DeleteUser(ctx, id)
		if err != nil {
			return fmt.Errorf("FirebaseClient.DeleteUser with UserID: %s: %w", id, err)
		}
	}
	return nil
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

func studentEntityWithFullNameOnly(id string) (*entity.LegacyStudent, error) {
	now := time.Now()
	num := rand.Int()
	student := &entity.LegacyStudent{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.LegacyUser)
	fullName := randomNameWithSpaceInside()

	err := multierr.Combine(
		student.LegacyUser.ID.Set(id),
		student.LegacyUser.Email.Set(fmt.Sprintf(stubEmail, num)),
		student.LegacyUser.PhoneNumber.Set(fmt.Sprintf(stubPhoneNumber, num)),
		student.LegacyUser.GivenName.Set(""),
		student.LegacyUser.Avatar.Set(""),
		student.LegacyUser.IsTester.Set(false),
		student.LegacyUser.FacebookID.Set(id),
		student.LegacyUser.PhoneVerified.Set(false),
		student.LegacyUser.EmailVerified.Set(false),
		student.LegacyUser.DeletedAt.Set(nil),
		student.LegacyUser.LastLoginDate.Set(nil),
		student.LegacyUser.FullName.Set(fullName),
		student.LegacyUser.FirstName.Set(""),
		student.LegacyUser.LastName.Set(""),
		student.LegacyUser.FirstNamePhonetic.Set(""),
		student.LegacyUser.LastNamePhonetic.Set(""),
		student.LegacyUser.FullNamePhonetic.Set(""),
		student.LegacyUser.Country.Set(cpb.Country_COUNTRY_VN.String()),
		student.LegacyUser.Group.Set(entity.UserGroupStudent),
		student.LegacyUser.Birthday.Set(now),
		student.LegacyUser.Gender.Set(pb.Gender_MALE),
		student.LegacyUser.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),

		student.ID.Set(id),
		student.CurrentGrade.Set(12),
		student.OnTrial.Set(true),
		student.TotalQuestionLimit.Set(10),
		student.SchoolID.Set(constants.ManabieSchool),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
		student.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
	)
	if err != nil {
		return nil, err
	}
	return student, nil
}

func staffEntity(id string) (*entity.Staff, error) {
	now := time.Now()
	num := rand.Int()
	staff := &entity.Staff{}
	database.AllNullEntity(staff)
	database.AllNullEntity(&staff.LegacyUser)
	fullName := randomNameWithSpaceInside()
	min := 1
	max := 2
	userGroup := entity.UserGroupSchoolAdmin
	randomGroup := min + rand.Intn(max-min+1)

	if randomGroup == min {
		userGroup = entity.UserGroupTeacher
	}

	err := multierr.Combine(
		staff.LegacyUser.ID.Set(id),
		staff.LegacyUser.Email.Set(fmt.Sprintf(stubEmail, num)),
		staff.LegacyUser.PhoneNumber.Set(fmt.Sprintf(stubPhoneNumber, num)),
		staff.LegacyUser.GivenName.Set(""),
		staff.LegacyUser.Avatar.Set(""),
		staff.LegacyUser.IsTester.Set(false),
		staff.LegacyUser.FacebookID.Set(id),
		staff.LegacyUser.PhoneVerified.Set(false),
		staff.LegacyUser.EmailVerified.Set(false),
		staff.LegacyUser.DeletedAt.Set(nil),
		staff.LegacyUser.LastLoginDate.Set(nil),
		staff.LegacyUser.FullName.Set(fullName),
		staff.LegacyUser.FirstName.Set(""),
		staff.LegacyUser.LastName.Set(""),
		staff.LegacyUser.FirstNamePhonetic.Set(""),
		staff.LegacyUser.LastNamePhonetic.Set(""),
		staff.LegacyUser.FullNamePhonetic.Set(""),
		staff.LegacyUser.Country.Set(cpb.Country_COUNTRY_VN.String()),
		staff.LegacyUser.Group.Set(userGroup),
		staff.LegacyUser.Birthday.Set(now),
		staff.LegacyUser.Gender.Set(pb.Gender_MALE),
		staff.LegacyUser.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),

		staff.ID.Set(id),
		staff.WorkingStatus.Set(""),
		staff.CreatedAt.Set(now),
		staff.UpdatedAt.Set(now),
		staff.ResourcePath.Set(fmt.Sprint(constants.ManabieSchool)),
	)
	if err != nil {
		return nil, err
	}
	return staff, nil
}

func randomBool() bool {
	if rand.Intn(2) == 0 {
		return false
	}
	return true
}

const numericCharset = "0123456789"
const wordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func stringWithCharset(length int, charset string) string {
	seededRand := rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func randomTextString(length int) string {
	return stringWithCharset(length, wordCharset)
}

func randomNameWithSpaceInside() string {
	min := 0
	max := 3
	// set seed
	rand.Seed(time.Now().UnixNano())
	// generate random number and print on console
	randomSpace := rand.Intn(max-min) + min
	if randomSpace == 0 {
		return randomTextString(5)
	}
	result := ""
	for i := 0; i < randomSpace; i++ {
		result += randomTextString(4)
		result += " "
	}
	return strings.Trim(result, " ")
}

func randomNumericString(length int) string {
	return stringWithCharset(length, numericCharset)
}

func checkMatchDate(expect *timestamppb.Timestamp, actual time.Time) bool {
	if expect == nil {
		return actual.IsZero()
	}
	timeExpectFormatted := expect.AsTime().Format("2006-01-02")
	timeActualFormatted := actual.Format("2006-01-02")
	return timeExpectFormatted == timeActualFormatted
}

func getKeyByValueMap(m map[string]string, value string) (string, bool) {
	for key, val := range m {
		if val == value {
			return key, true
		}
	}
	return "", false
}
