package timesheet

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/bob/constants"
	golibsauth "github.com/manabie-com/backend/internal/golibs/auth"
	libConstants "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func (s *Suite) SignedAsAccount(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		userGroup string
		err       error
	)

	switch group {
	case "unauthenticated":
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	case "student":
		return s.aSignedInStudent(ctx)
	case "staff granted role school admin":
		return s.aSignedInStaff(ctx, []string{constant.RoleSchoolAdmin})
	case "staff granted role hq staff":
		return s.aSignedInStaff(ctx, []string{constant.RoleHQStaff})
	// case "staff granted role school admin and teacher":
	// 	return s.aSignedInStaff(ctx, []string{constant.RoleSchoolAdmin, constant.RoleTeacher})
	case "staff granted role centre lead":
		return s.aSignedInStaff(ctx, []string{constant.RoleCentreLead})
	case "staff granted role centre manager":
		return s.aSignedInStaff(ctx, []string{constant.RoleCentreManager})
	case "staff granted role centre staff":
		return s.aSignedInStaff(ctx, []string{constant.RoleCentreStaff})
	case "staff granted role teacher":
		return s.aSignedInStaff(ctx, []string{constant.RoleTeacher})
	case "teacher":
		userGroup = constant.UserGroupTeacher
	case "school admin":
		userGroup = constant.UserGroupSchoolAdmin
	case "parent":
		userGroup = constant.UserGroupParent
	case "organization manager":
		userGroup = constant.UserGroupOrganizationManager
	}

	userID := idutil.ULIDNow()

	stepState.CurrentUserID = userID
	stepState.CurrentUserGroup = userGroup

	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(userID), withRole(userGroup))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.generateExchangeToken(userID, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) signedAsAccountV2(ctx context.Context, account string) (context.Context, error) {
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
	}
	roleWithLocation.LocationIDs = []string{libConstants.ManabieOrgLocation}

	authInfo, err := usermgmt.SignIn(ctx, s.BobDBTrace, s.AuthPostgresDB, s.ShamirConn, s.Cfg.JWTApplicant, s.FirebaseAddress, s.UserMgmtConn, roleWithLocation, []string{libConstants.ManabieOrgLocation})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = authInfo.UserID
	stepState.AuthToken = authInfo.Token
	stepState.LocationID = libConstants.ManabieOrgLocation
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aSignedInStaff(ctx context.Context, roles []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := s.newID()

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
		return nil, err
	}

	if err := assignUserGroupToUser(ctx, s.BobPostgresDBTrace, id, []string{createUserGroupResp.UserGroupId}, stepState.CurrentSchoolIDString); err != nil {
		return nil, err
	}

	time.Sleep(time.Second) // wait for data sync

	stepState.CurrentUserID = id

	return ctx, nil
}

func (s *Suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string) (*pb.CreateUserGroupResponse, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", s.newID()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.BobPostgresDB.Query(ctx, stmt, roleNames, len(roleNames))

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
				LocationIds: []string{libConstants.ManabieOrgLocation},
			},
		)
	}

	resp, err := pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).CreateUserGroup(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	if err != nil {
		return nil, fmt.Errorf("createUserGroupWithRoleNames: %w", err)
	}

	return resp, nil
}

func (s *Suite) newID() string {
	return idutil.ULIDNow()
}

func generateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func generateAuthenticationToken(sub string, template string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot generate new user token, err: %v", err)
	}

	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(bodyResp), nil
}

func (s *Suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}

	ctx = golibsauth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	num := rand.Int()

	user := &entity.LegacyUser{}
	database.AllNullEntity(user)

	firstName := fmt.Sprintf("valid-user-first-name-%d", num)
	lastName := fmt.Sprintf("valid-user-last-name-%d", num)

	err := multierr.Combine(
		user.FullName.Set(helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		user.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		user.Country.Set(cpb.Country_COUNTRY_VN.String()),
		user.Group.Set(constant.UserGroupAdmin),
		user.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
		user.ResourcePath.Set(database.Text(fmt.Sprint(schoolID))),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
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
			t.ResourcePath = database.Text(strconv.FormatInt(schoolID, 10))
			err := t.SchoolIDs.Set([]int64{schoolID})

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

			locationRepo := &location_repo.LocationRepo{}

			locationOrg, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(schoolID))

			if err != nil {
				return fmt.Errorf("locationRepo.GetLocationOrg: %v", err)
			}

			userGroupID, err := s.createUserGroupWithRoleNamesUsingDB(ctx, []string{"School Admin"}, []string{locationOrg.LocationID}, schoolID)

			if err != nil {
				return fmt.Errorf("s.createUserGroupWithRoleNames: %v", err)
			}

			if err := assignUserGroupToUser(ctx, tx, user.ID.String, []string{userGroupID}, strconv.Itoa(int(schoolID))); err != nil {

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

	err = initStaff(ctx, user.ID.String, strconv.FormatInt(schoolID, 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createUserGroupWithRoleNamesUsingDB(ctx context.Context, roleNames []string, grantedLocationIDs []string, resourcePath int64) (string, error) {
	req := &pb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.BobPostgresDB.Query(ctx, stmt, roleNames, len(roleNames))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return "", fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("rows.Err: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("s.getOrgLocationID: %w", err)
	}

	for _, roleID := range roleIDs {
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&pb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: grantedLocationIDs,
			},
		)
	}

	locationRepo := &location_repo.LocationRepo{}
	userGroupV2Repo := repository.UserGroupV2Repo{}
	grantedRoleRepo := repository.GrantedRoleRepo{}

	var userGroup *entity.UserGroupV2
	if err = database.ExecInTx(ctx, s.BobPostgresDB, func(ctx context.Context, tx pgx.Tx) error {
		orgLocation, err := locationRepo.GetLocationOrg(ctx, tx, fmt.Sprint(resourcePath))
		if err != nil {
			return fmt.Errorf("locationRepo.GetLocationOrg: %w", err)
		}

		// convert payload to entity
		if userGroup, err = userGroupPayloadToUserGroupEnt(req, fmt.Sprint(resourcePath), orgLocation); err != nil {
			return fmt.Errorf("s.UserGroupPayloadToUserGroupEnts: %w", err)
		}

		// create user group first
		if err = userGroupV2Repo.Create(ctx, tx, userGroup); err != nil {
			return fmt.Errorf("userGroupV2Repo.Create: %w", err)
		}

		var grantedRole *entity.GrantedRole
		for _, roleWithLocations := range req.RoleWithLocations {
			// convert payload to entity
			if grantedRole, err = roleWithLocationsPayloadToGrantedRole(roleWithLocations, userGroup.UserGroupID.String, fmt.Sprint(resourcePath)); err != nil {
				return fmt.Errorf("s.RoleWithLocationsPayloadToGrantedRole: %w", err)
			}
			// create granted_role
			if err = grantedRoleRepo.Create(ctx, tx, grantedRole); err != nil {
				return fmt.Errorf("grantedRoleRepo.Create: %w", err)
			}

			// link granted_role to access path(by location ids)
			if err = grantedRoleRepo.LinkGrantedRoleToAccessPath(ctx, tx, grantedRole, roleWithLocations.LocationIds); err != nil {
				return fmt.Errorf("grantedRoleRepo.LinkGrantedRoleToAccessPath: %w", err)
			}
		}
		return nil
	}); err != nil {
		return "", fmt.Errorf("database.ExecInTx: %w", err)
	}

	return userGroup.UserGroupID.String, nil
}
func userGroupPayloadToUserGroupEnt(payload *pb.CreateUserGroupRequest, resourcePath string, orgLocation *domain.Location) (*entity.UserGroupV2, error) {
	userGroup := &entity.UserGroupV2{}
	database.AllNullEntity(userGroup)
	if err := multierr.Combine(
		userGroup.UserGroupID.Set(idutil.ULIDNow()),
		userGroup.UserGroupName.Set(payload.UserGroupName),
		userGroup.ResourcePath.Set(resourcePath),
		userGroup.OrgLocationID.Set(orgLocation.LocationID),
		userGroup.IsSystem.Set(false),
	); err != nil {
		return nil, fmt.Errorf("set user group fail: %w", err)
	}

	return userGroup, nil
}
func roleWithLocationsPayloadToGrantedRole(payload *pb.RoleWithLocations, userGroupID string, resourcePath string) (*entity.GrantedRole, error) {
	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	if err := multierr.Combine(
		grantedRole.GrantedRoleID.Set(idutil.ULIDNow()),
		grantedRole.UserGroupID.Set(userGroupID),
		grantedRole.RoleID.Set(payload.RoleId),
		grantedRole.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, fmt.Errorf("set granted role fail: %w", err)
	}

	return grantedRole, nil
}

func assignUserGroupToUser(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string, resourcePath string) error {
	userGroupMembers := make([]*entity.UserGroupMember, 0)

	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(fmt.Sprint(resourcePath)),
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
func (s *Suite) aSignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	var err error
	stepState.AuthToken, err = generateValidAuthenticationToken(id, "phone")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupStudent

	return s.aValidStudentInDB(StepStateToContext(ctx, stepState), id)
}

func (s *Suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentRepo := repository.StudentRepo{}
	now := time.Now()
	num := rand.Int()
	student := &entity.LegacyStudent{}
	database.AllNullEntity(student)
	err := multierr.Combine(
		student.LegacyUser.ID.Set(id),
		student.LegacyUser.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		student.LegacyUser.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		student.LegacyUser.GivenName.Set(""),
		student.LegacyUser.Avatar.Set(""),
		student.LegacyUser.IsTester.Set(false),
		student.LegacyUser.FacebookID.Set(id),
		student.LegacyUser.PhoneVerified.Set(false),
		student.LegacyUser.EmailVerified.Set(false),
		student.LegacyUser.DeletedAt.Set(nil),
		student.LegacyUser.LastLoginDate.Set(nil),
		student.LegacyUser.LastName.Set(fmt.Sprintf("valid-user-%d", num)),
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
		return StepStateToContext(ctx, stepState), err
	}
	err = studentRepo.Create(ctx, s.BobDBTrace, student)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func contextWithToken(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", stepState.AuthToken)
}

func (s *Suite) ReturnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}
	return ctx, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Gen random string with fixed length
func randStringBytes(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (s *Suite) createLessonRecords(ctx context.Context, lessonStatusStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	splitLessonStatus := strings.Split(lessonStatusStr, "-")
	if stepState.CurrentUserGroup != constant.UserGroupTeacher {
		err := initTeachers(ctx, stepState.CurrentUserID, constants.ManabieSchool)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	// create teacher record
	for _, lessonStatusFormat := range splitLessonStatus {
		var lessonStatus string
		switch lessonStatusFormat {
		case "PUBLISHED":
			lessonStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String()
		case "COMPLETED":
			lessonStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_COMPLETED.String()
		case "CANCELLED":
			lessonStatus = cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_CANCELED.String()
		default:
			return StepStateToContext(ctx, stepState), fmt.Errorf("invalid lesson status")
		}
		for _, currentTimesheetID := range stepState.CurrentTimesheetIDs {
			// create lesson record
			lessonID, err := initLesson(ctx, stepState.CurrentUserID, locationIDs[0], lessonStatus, strconv.Itoa(constants.ManabieSchool))
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			// create timesheet lesson teacher record
			err = initLessonTeachers(ctx, lessonID, stepState.CurrentUserID)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}

			stepState.TimesheetLessonIDs = append(stepState.TimesheetLessonIDs, lessonID)
			// create timesheet lesson hours record
			err = initTimesheetLessonHours(ctx, currentTimesheetID, lessonID, strconv.Itoa(constants.ManabieSchool))
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func convertTimesheetStatusFormat(timesheetStatusFormat string) string {
	var timesheetStatus string
	switch timesheetStatusFormat {
	case "DRAFT":
		timesheetStatus = tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String()
	case "SUBMITTED":
		timesheetStatus = tpb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String()
	case "APPROVED":
		timesheetStatus = tpb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String()
	case "CONFIRMED":
		timesheetStatus = tpb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String()
	default:
		timesheetStatus = tpb.TimesheetStatus_TIMESHEET_STATUS_NONE.String()
	}
	return timesheetStatus
}

func genDateTimeFormat(typeOfDate string) time.Time {
	var dateFormatted time.Time
	switch typeOfDate {
	case "END_DATE":
		dateFormatted = time.Date(2022, 07, 24, 8, 30, 00, 00, timeutil.Timezone(pbc.COUNTRY_JP))
	default:
		dateFormatted = time.Now().In(timeutil.Timezone(pbc.COUNTRY_JP))
	}
	return dateFormatted
}
