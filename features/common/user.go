package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	bob_repository "github.com/manabie-com/backend/internal/bob/repositories"
	consta "github.com/manabie-com/backend/internal/entryexitmgmt/constant"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
	"go.uber.org/multierr"
)

func (s *suite) ASignedInStudent(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := s.newID()
	var err error
	ctx, err = s.aValidStudentInDB(ctx, id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.GenerateExchangeToken(id, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentStudentID = id

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ASignedInStudentInStudentList(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	studentID := stepState.StudentIds[0]
	stepState.AuthToken, err = s.GenerateExchangeToken(studentID, entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentUserID = studentID
	stepState.CurrentStudentID = studentID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	if ctx, err := s.createStudentWithName(ctx, id, name); err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	stepState.CurrentUserID = id
	stepState.StudentID = id
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidStudentInDB(ctx context.Context, id string) (context.Context, error) {
	return s.aValidStudentWithName(ctx, id, "")
}

func (s *suite) aValidStudentWithSchoolID(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	sql := "UPDATE students SET school_id = $1 WHERE student_id = $2"
	_, err := s.BobDB.Exec(ctx, sql, &schoolID, &id)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentWithName(ctx context.Context, id, name string) (context.Context, error) {
	num := s.newID()
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constant.ManabieSchool
	}

	if name == "" {
		name = fmt.Sprintf("valid-student-%s", num)
	}

	student := &entities.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)
	gradeIDs := stepState.GradeIDs
	if len(gradeIDs) > 0 {
		randomIndex := rand.Intn(len(gradeIDs))
		_ = student.GradeID.Set(gradeIDs[randomIndex])
	}
	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.Country.Set(pb.COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", num)),
		student.Email.Set(fmt.Sprintf("email+%s", num)),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(stepState.CurrentSchoolID),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	locationRepo := &location_repo.LocationRepo{}
	locationOrg, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(stepState.CurrentSchoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("locationRepo.GetLocationOrg: %v", err)
	}
	claims := interceptors.JWTClaimsFromContext(ctx)

	err = database.ExecInTx(ctx, s.BobDB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.AddUserAccessPath(ctx, id, locationOrg.LocationID, int64(stepState.CurrentSchoolID), tx); err != nil {
			return errors.Wrap(err, "s.AddUserAccessPath: "+locationOrg.LocationID+" "+fmt.Sprint(stepState.CurrentSchoolID)+" CurrentUserID "+claims.Manabie.UserID+" ResourcePath: "+claims.Manabie.ResourcePath)
		}
		if err := (&bob_repository.StudentRepo{}).Create(ctx, tx, student); err != nil {
			return errors.Wrap(err, "s.StudentRepo.CreateTx")
		}

		if student.AppleUser.ID.String != "" {
			if err := (&bob_repository.AppleUserRepo{}).Create(ctx, tx, &student.AppleUser); err != nil {
				return errors.Wrap(err, "s.AppleUserRepo.Create")
			}
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// since Bob entities hasn't been updated with the new first_name & last_name column
	// we will update them here
	updateNameQuery := "UPDATE users SET first_name = $1, last_name = $2 WHERE user_id = $3"
	_, err = s.BobDB.Exec(ctx, updateNameQuery, "bdd-test-create-first-name-"+student.ID.String,
		"bdd-test-create-last-name-"+student.ID.String, student.ID.String)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentNames = append(stepState.StudentNames, student.LastName.String)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createStudentWithNameInUserBasicInfoTable(ctx context.Context, id, name string) (context.Context, error) {
	num := s.newID()
	stepState := StepStateFromContext(ctx)

	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constant.ManabieSchool
	}

	if name == "" {
		name = fmt.Sprintf("valid-student-%s", num)
	}

	student := &entities.Student{}
	database.AllNullEntity(student)
	database.AllNullEntity(&student.User)
	database.AllNullEntity(&student.User.AppleUser)
	gradeIDs := stepState.GradeIDs
	if len(gradeIDs) > 0 {
		randomIndex := rand.Intn(len(gradeIDs))
		_ = student.GradeID.Set(gradeIDs[randomIndex])
	}
	err := multierr.Combine(
		student.ID.Set(id),
		student.LastName.Set(name),
		student.Country.Set(pb.COUNTRY_VN.String()),
		student.PhoneNumber.Set(fmt.Sprintf("phone-number+%s", num)),
		student.Email.Set(fmt.Sprintf("email+%s", num)),
		student.CurrentGrade.Set(12),
		student.TargetUniversity.Set("TG11DT"),
		student.TotalQuestionLimit.Set(5),
		student.SchoolID.Set(stepState.CurrentSchoolID),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = database.ExecInTx(ctx, s.BobDB, func(ctx context.Context, tx pgx.Tx) error {
		if err := (&bob_repository.StudentRepo{}).Create(ctx, tx, student); err != nil {
			return errors.Wrap(err, "s.StudentRepo.CreateTx")
		}

		if student.AppleUser.ID.String != "" {
			if err := (&bob_repository.AppleUserRepo{}).Create(ctx, tx, &student.AppleUser); err != nil {
				return errors.Wrap(err, "s.AppleUserRepo.Create")
			}
		}
		return nil
	})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// since Bob entities hasn't been updated with the new first_name & last_name column
	// we will update them here
	studentFirstName := "bdd-test-create-first-name-" + student.ID.String
	studentLastName := "bdd-test-create-last-name-" + student.ID.String
	now := time.Now()
	// Insert student info to user_basic_info table
	createStudentBasicInfoQuery := "INSERT INTO user_basic_info(user_id, name, first_name, last_name, created_at, updated_at)" +
		"VALUES($1,$2,$3,$4,$5,$6);"
	_, err = s.BobDB.Exec(ctx, createStudentBasicInfoQuery,
		student.ID.String, student.LastName.String, studentFirstName, studentLastName, now, now)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StudentNames = append(stepState.StudentNames, student.LastName.String)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GenerateExchangeTokenCtx(ctx context.Context, userID, userGroup string) (string, error) {
	firebaseToken, err := s.generateValidAuthenticationToken(userID)
	if err != nil {
		return "", err
	}
	rp := intResourcePathFromCtx(ctx)
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.StepState.ApplicantID, rp, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("failed to generate exchange token: %w", err)
	}
	return token, nil
}

func (s *suite) GenerateExchangeToken(userID, userGroup string) (string, error) {
	firebaseToken, err := s.generateValidAuthenticationToken(userID)
	if err != nil {
		return "", err
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.StepState.ApplicantID, 1, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("failed to generate exchange token: %w", err)
	}
	return token, nil
}

func (s *suite) GenerateExchangeTokenWithSchool(userID, userGroup string, school int64) (string, error) {
	firebaseToken, err := s.generateValidAuthenticationToken(userID)
	if err != nil {
		return "", err
	}
	// ALL test should have one resource_path
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, s.StepState.ApplicantID, school, s.ShamirConn)
	if err != nil {
		return "", fmt.Errorf("failed to generate exchange token with school: %w", err)
	}
	return token, nil
}

type genAuthTokenOption func(values url.Values)

func (s *suite) generateAuthenticationToken(sub, template string, opts ...genAuthTokenOption) (string, error) {
	v := url.Values{}
	v.Set("template", template)
	v.Set("UserID", sub)
	for _, opt := range opts {
		opt(v)
	}
	resp, err := http.Get("http://" + s.FirebaseAddress + "/token?" + v.Encode())
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(b), nil
}

func GenerateAuthenticationToken(firebaseAddress string, sub string, userGroup string, opts ...genAuthTokenOption) (string, error) {
	v := url.Values{}
	v.Set("template", "templates/"+userGroup+".template")
	v.Set("UserID", sub)
	for _, opt := range opts {
		opt(v)
	}
	resp, err := http.Get("http://" + firebaseAddress + "/token?" + v.Encode())
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken:cannot generate new user token, err: %v", err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("aValidAuthenticationToken: cannot read token from response, err: %v", err)
	}
	resp.Body.Close()
	return string(b), nil
}

func (s *suite) generateValidAuthenticationToken(sub string) (string, error) {
	return s.generateAuthenticationToken(sub, "templates/phone.template")
}

type userOption func(u *entities.User)

func withID(id string) userOption {
	return func(u *entities.User) {
		_ = u.ID.Set(id)
	}
}

func withRole(group string) userOption {
	return func(u *entities.User) {
		_ = u.Group.Set(group)
	}
}

func (s *suite) ASignedInAdmin(ctx context.Context) (context.Context, error) {
	id := s.newID()
	var err error
	stepState := StepStateFromContext(ctx)
	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(entities.UserGroupAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AuthToken, err = s.GenerateExchangeToken(id, constant.UserGroupAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GenerateExchangeToken:%w", err)
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupAdmin
	return StepStateToContext(ctx, stepState), nil
}

const (
	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"
)

// nolint
func (s *suite) aValidUser(ctx context.Context, opts ...userOption) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	adminContext := ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
	num := s.newID()

	userRepo := bob_repository.UserRepo{}
	u := &entities.User{}
	database.AllNullEntity(u)

	err := multierr.Combine(
		u.ID.Set(s.newID()),
		u.LastName.Set(fmt.Sprintf("valid-user-%s", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%s", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%s@email.com", num)),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%s", num)),
		u.Country.Set(pb.COUNTRY_VN.String()),
		u.Group.Set(entities.UserGroupStudent),
		u.DeviceToken.Set(nil),
		u.AllowNotification.Set(true),
		u.CreatedAt.Set(time.Now()),
		u.UpdatedAt.Set(time.Now()),
		u.IsTester.Set(nil),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for _, opt := range opts {
		opt(u)
	}

	err = userRepo.Create(adminContext, s.Connections.BobDB, u)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert user err: %w", err)
	}

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = 1
	}
	if u.Group.String == entities.UserGroupTeacher {
		teacher := &entities.Teacher{}
		database.AllNullEntity(teacher)

		err = multierr.Combine(
			teacher.ID.Set(u.ID.String),
			teacher.SchoolIDs.Set([]int64{schoolID}),
			teacher.UpdatedAt.Set(time.Now()),
			teacher.CreatedAt.Set(time.Now()),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(adminContext, teacher, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert teacher error: %w", err)
		}
	}

	if u.Group.String == constant.UserGroupSchoolAdmin {
		schoolAdminAccount := &entities.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(u.ID.String),
			schoolAdminAccount.SchoolID.Set(schoolID),
			schoolAdminAccount.UpdatedAt.Set(time.Now()),
			schoolAdminAccount.CreatedAt.Set(time.Now()),
			schoolAdminAccount.ResourcePath.Set("1"),
		)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		_, err = database.Insert(adminContext, schoolAdminAccount, s.BobDB.Exec)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("aValidUser insert school error: %w", err)
		}
	}

	ug := entities.UserGroup{}
	database.AllNullEntity(&ug)

	now := time.Now()
	ug.UserID.Set(u.ID.String)
	ug.GroupID.Set(u.Group.String)
	ug.UpdatedAt.Set(now)
	ug.CreatedAt.Set(now)
	ug.IsOrigin.Set(true)
	ug.Status.Set(entities.UserGroupStatusActive)

	userGroupRepo := bob_repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(adminContext, s.BobDB, &ug)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ASignedInTeacherWithSchoolID(ctx context.Context, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := ksuid.New().String()

	ctx, err := s.aValidTeacherProfileWithID(ctx, id, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentTeacherID = id
	stepState.CurrentUserID = id
	stepState.AuthToken, err = s.GenerateExchangeToken(id, entities.UserGroupTeacher)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) ASignedInTeacher(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// TODO: pls discuss about this matter: Why assign current school id to 1
	stepState.CurrentSchoolID = 1
	StepStateToContext(ctx, stepState)
	return s.ASignedInTeacherWithSchoolID(ctx, stepState.CurrentSchoolID)
}

func (s *suite) ASignedInTeacherWithOrdinalNumberInTeacherList(ctx context.Context, i string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var teacherID string
	switch i {
	case "first":
		teacherID = stepState.TeacherIDs[0]
	case "second":
		teacherID = stepState.TeacherIDs[1]
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not get account for teacher %s", i)
	}
	var err error
	stepState.AuthToken, err = s.GenerateExchangeToken(teacherID, entity.UserGroupTeacher)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentTeacherID = teacherID
	stepState.CurrentUserID = teacherID

	return StepStateToContext(ctx, stepState), nil
}

//nolint:errcheck
func (s *suite) aValidTeacherProfileWithID(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entities.Teacher{}
	database.AllNullEntity(&c.User)
	database.AllNullEntity(&c)
	c.ID.Set(id)
	var schoolIDs []int32
	if len(stepState.Schools) > 0 {
		schoolIDs = []int32{stepState.Schools[0].ID.Int}
	}
	if schoolID != 0 {
		schoolIDs = append(schoolIDs, schoolID)
	}
	c.SchoolIDs.Set(schoolIDs)
	now := time.Now()
	if err := c.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(now); err != nil {
		return nil, err
	}
	num := rand.Int()
	u := entities.User{}
	database.AllNullEntity(&u)
	u.ID = c.ID
	u.LastName.Set(fmt.Sprintf("valid-teacher-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-teacher-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-teacher-%d", num))
	u.Country.Set(pb.COUNTRY_VN.String())
	u.Group.Set(entities.UserGroupTeacher)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = c.CreatedAt
	u.UpdatedAt = c.UpdatedAt
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	uG := entities.UserGroup{UserID: c.ID, GroupID: database.Text(pb.USER_GROUP_TEACHER.String()), IsOrigin: database.Bool(true)}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt
	staff := entity.Staff{}
	staff.ID = c.ID
	staff.UpdatedAt = u.UpdatedAt
	staff.CreatedAt = u.CreatedAt
	staff.DeletedAt.Set(nil)
	staff.StartDate.Set(nil)
	staff.EndDate.Set(nil)
	staff.AutoCreateTimesheet.Set(false)
	staff.WorkingStatus.Set("AVAILABLE")
	_, err := database.InsertExcept(ctx, &u, []string{"resource_path"}, s.BobDB.Exec)
	if err != nil {
		return ctx, err
	}
	_, err = database.InsertExcept(ctx, &c, []string{"resource_path"}, s.BobDB.Exec)
	if err != nil {
		return ctx, err
	}
	_, err = database.InsertExcept(ctx, &staff, []string{"resource_path"}, s.BobDB.Exec)
	if err != nil {
		return ctx, err
	}
	cmdTag, err := database.InsertExcept(ctx, &uG, []string{"resource_path"}, s.BobDB.Exec)
	if err != nil {
		return ctx, err
	}
	if cmdTag.RowsAffected() == 0 {
		return ctx, errors.New("cannot insert teacher for testing")
	}
	location := stepState.CurrentCenterID
	if location == "" {
		locationRepo := &location_repo.LocationRepo{}
		locationOrg, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(schoolID))
		if err != nil {
			return ctx, err
		}
		location = locationOrg.LocationID
	}
	err = s.AddUserAccessPath(ctx, c.ID.String, location, int64(schoolID), nil)
	if err != nil {
		return ctx, fmt.Errorf("error when insert access path teacher %s %s %s", location, fmt.Sprint(schoolID), c.ID.String)
	}
	stepState.TeacherNames = append(stepState.TeacherNames, u.LastName.String)
	return ctx, nil
}

func (s *suite) HisOwnedStudentUUID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	t, err := jwt.ParseString(stepState.AuthToken)
	if err != nil {
		return ctx, err
	}
	stepState.CurrentStudentID = t.Subject()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aValidSchoolAdminProfileWithID(ctx context.Context, id, userGroup string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	c := entity.SchoolAdmin{}
	database.AllNullEntity(&c)
	if err := multierr.Combine(
		c.SchoolAdminID.Set(id),
		c.SchoolID.Set(schoolID),
		c.ResourcePath.Set(golibs.ResourcePathFromCtx(ctx)),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	if err := c.UpdatedAt.Set(now); err != nil {
		return nil, err
	}
	if err := c.CreatedAt.Set(now); err != nil {
		return nil, err
	}

	num := rand.Int()
	u := entity.LegacyUser{}
	database.AllNullEntity(&u)
	if userGroup == "" {
		userGroup = entity.UserGroupSchoolAdmin
	}

	if err := multierr.Combine(
		u.ID.Set(c.SchoolAdminID),
		u.FullName.Set(fmt.Sprintf("valid-school-admin-%d", num)),
		u.FirstName.Set(""),
		u.LastName.Set(""),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-school-admin-%d@email.com", num)),
		u.Avatar.Set(fmt.Sprintf("http://valid-school-admin-%d", num)),
		u.Country.Set(pb.COUNTRY_VN.String()),
		u.Group.Set(userGroup),
		u.DeviceToken.Set(nil),
		u.AllowNotification.Set(true),
		u.IsTester.Set(nil),
		u.FacebookID.Set(nil),
		u.CreatedAt.Set(c.CreatedAt),
		u.UpdatedAt.Set(c.UpdatedAt),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userRepo := repository.UserRepo{}

	err := userRepo.Create(ctx, s.BobDB, &u)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	schoolAdminRepo := repository.SchoolAdminRepo{}
	err = schoolAdminRepo.CreateMultiple(ctx, s.BobDB, []*entity.SchoolAdmin{&c})

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ug := entity.UserGroup{}
	database.AllNullEntity(&ug)

	if err := multierr.Combine(
		ug.UserID.Set(id),
		ug.GroupID.Set(userGroup),
		ug.UpdatedAt.Set(now),
		ug.CreatedAt.Set(now),
		ug.IsOrigin.Set(true),
		ug.Status.Set(entity.UserGroupStatusActive),
	); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userGroupRepo := repository.UserGroupRepo{}
	if hasResourcePath(ctx) {
		err := ug.ResourcePath.Set(resourcePathFromCtx(ctx))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	err = userGroupRepo.Upsert(ctx, s.BobDB, &ug)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	locationRepo := &location_repo.LocationRepo{}
	locationOrg, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("locationRepo.GetLocationOrg: %v", err)
	}

	userGroupID, err := s.createUserGroupWithRoleNames(ctx, []string{"School Admin"}, []string{locationOrg.LocationID}, int64(schoolID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.createUserGroupWithRoleNames: %v", err)
	}

	if err := addUserToUserGroups(ctx, s.BobPostgresDB, id, []string{userGroupID}, strconv.Itoa(int(schoolID))); err != nil {
		return ctx, err
	}

	if err := s.AddUserAccessPath(ctx, id, locationOrg.LocationID, int64(schoolID), nil); err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInSchoolAdminWithSchoolID(ctx context.Context, group string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := s.newID()
	ctx, err := s.aValidSchoolAdminProfileWithID(ctx, id, group, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if hasResourcePath(ctx) {
		stepState.AuthToken, err = s.GenerateExchangeTokenCtx(ctx, id, group)
	} else {
		stepState.AuthToken, err = s.GenerateExchangeToken(id, group)
	}
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx = ValidContext(ctx, int(schoolID), id, stepState.AuthToken)
	stepState.CurrentUserID = id
	return StepStateToContext(ctx, stepState), nil
}

func noRecord(err error) bool {
	if err == nil || (err != nil && strings.Contains(err.Error(), "no rows")) {
		return true
	}
	return false
}

func (s *suite) AddOrgLocationForSchool(ctx context.Context, schoolID int32) error {
	locationRepo := &location_repo.LocationRepo{}
	_, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(schoolID))
	if !noRecord(err) || schoolID == constants.ManabieSchool {
		return nil
	}
	return database.ExecInTx(ctx, s.BobPostgresDB, func(ctx context.Context, tx pgx.Tx) (err error) {
		locationID := fmt.Sprint(schoolID)
		locationTypeID := s.newID()

		stmtLocationType := `INSERT INTO location_types (location_type_id,"name",display_name,updated_at,created_at,resource_path,is_archived,"level") VALUES
	 ($1,$2,$3,now(),now(),$4,false,0) on conflict ON CONSTRAINT unique__location_type_name_resource_path do update set updated_at=now() RETURNING location_type_id`
		err = tx.QueryRow(ctx, stmtLocationType, locationTypeID, "org", locationTypeID, fmt.Sprint(schoolID)).Scan(&locationTypeID)
		if err != nil {
			return err
		}
		stmtLocation := `INSERT INTO locations (location_id,name, is_archived, resource_path, location_type, access_path) VALUES($1,$2,$3,$4,$5,$1)
				ON CONFLICT DO NOTHING`
		_, err = tx.Exec(ctx, stmtLocation, locationID, locationID, false, fmt.Sprint(schoolID), locationTypeID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *suite) AddLocationUnderManabieOrg(ctx context.Context, locationID string) error {
	stmtLocation := `INSERT INTO locations (location_id,name, is_archived, resource_path,access_path, parent_location_id) VALUES($1,$2,$3,$4,$5,$6)
				ON CONFLICT DO NOTHING`
	accessPath := constants.ManabieOrgLocation + "/" + locationID
	_, err := s.BobPostgresDB.Exec(ctx, stmtLocation, locationID, locationID, false, fmt.Sprint(constants.ManabieSchool), accessPath, constants.ManabieOrgLocation)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) AddLocationUnderShool(ctx context.Context, locationID string, schoolID int32) error {
	err := s.AddOrgLocationForSchool(ctx, schoolID)
	if err != nil {
		return err
	}

	locationRepo := &location_repo.LocationRepo{}
	orgLocation, err := locationRepo.GetLocationOrg(ctx, s.BobPostgresDB, fmt.Sprint(schoolID))
	if err != nil {
		return err
	}
	stmtLocation := `INSERT INTO locations (location_id,name, is_archived, resource_path,access_path, parent_location_id) VALUES($1,$2,$3,$4,$5,$6)
				ON CONFLICT DO NOTHING`
	accessPath := orgLocation.LocationID + "/" + locationID
	_, err = s.BobPostgresDB.Exec(ctx, stmtLocation, locationID, locationID, false, fmt.Sprint(schoolID), accessPath, orgLocation.LocationID)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) ASignedInWithSchool(ctx context.Context, role string, schoolID int32) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = schoolID
	err := s.AddOrgLocationForSchool(ctx, schoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error when add location for school %v", err)
	}
	switch role {
	case "school admin":
		return s.aSignedInSchoolAdminWithSchoolID(ctx,
			entity.UserGroupSchoolAdmin, schoolID)
	case "student":
		if ctx, err := s.ASignedInStudent(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		t, _ := jwt.ParseString(stepState.AuthToken)
		return s.aValidStudentWithSchoolID(ctx, t.Subject(), schoolID)

	case "teacher":
		return s.ASignedInTeacherWithSchoolID(ctx, schoolID)
	case "admin":
		return s.ASignedInAdmin(ctx)
	case "unauthenticated":
		stepState.AuthToken = "random-token"
		return StepStateToContext(ctx, stepState), nil
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateStudentAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudentIds = []string{s.newID(), s.newID()}

	for _, id := range stepState.StudentIds {
		if ctx, err := s.createStudentWithName(ctx, id, ""); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateSomeGrades(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = constant.ManabieSchool
	}
	schoolID := stepState.CurrentSchoolID
	// create grade
	stepState.GradeIDs = []string{s.newID()}
	for _, GradeID := range stepState.GradeIDs {
		gradeName := "name-" + GradeID
		sql := `insert into grade
	(grade_id, name , created_at, updated_at, resource_path, partner_internal_id)
	VALUES ($1, $2, $3, $4, $5, $6)  ON CONFLICT DO NOTHING;`
		_, err := s.BobDB.Exec(ctx, sql, database.Text(GradeID), gradeName, time.Now(), time.Now(), database.Text(fmt.Sprint(schoolID)), database.Text(fmt.Sprint(schoolID)))
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot init grade err: %s", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateATotalNumberOfStudentAccounts(ctx context.Context, name string, createdTotal int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < createdTotal; i++ {
		newStudentID := s.newID()

		if ctx, err := s.createStudentWithName(ctx, newStudentID, fmt.Sprintf("%s %s", name, newStudentID)); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds = append(stepState.StudentIds, newStudentID)
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) CreateATotalNumberOfStudentAccountsInUserBasicInfoTable(ctx context.Context, name string, createdTotal int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for i := 0; i < createdTotal; i++ {
		newStudentID := s.newID()

		if ctx, err := s.createStudentWithNameInUserBasicInfoTable(ctx, newStudentID, fmt.Sprintf("%s %s", name, newStudentID)); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		stepState.StudentIds = append(stepState.StudentIds, newStudentID)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) CreateTeacherAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{s.newID(), s.newID()}
	if stepState.CurrentSchoolID == 0 {
		stepState.CurrentSchoolID = 1
	}

	for _, id := range stepState.TeacherIDs {
		if ctx, err := s.aValidTeacherProfileWithID(ctx, id, stepState.CurrentSchoolID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GenerateValidAuthenticationToken(sub, userGroup string) (string, error) {
	return s.generateAuthenticationToken(sub, "templates/"+userGroup+".template")
}

func (s *suite) ASignedInAsSchoolAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.ASignedInWithSchool(ctx, "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) AValidTeacherProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
	return s.aValidTeacherProfileWithID(ctx, id, schoolID)
}
func (s *suite) GetTeacherProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.GetTeacherProfilesRequest{}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserServiceClient(s.BobConn).
		GetTeacherProfiles(contextWithToken(s, ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) GetStudentProfileV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveStudentProfileRequest{}

	stepState.Response, stepState.ResponseErr = bpb.NewStudentReaderServiceClient(s.BobConn).
		RetrieveStudentProfile(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ASignedInSchoolAdmin(ctx context.Context) (context.Context, error) {
	id := s.newID()
	var err error
	stepState := StepStateFromContext(ctx)
	ctx, err = s.aValidUser(StepStateToContext(ctx, stepState), withID(id), withRole(entities.UserGroupSchoolAdmin))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aValidUserInEureka(ctx, id, consta.RoleSchoolAdmin, entities.UserGroupSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create user in eureka: %w", err)
	}
	stepState.AuthToken, err = s.GenerateExchangeToken(id, consta.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("GenerateExchangeToken:%w", err)
	}
	stepState.CurrentUserID = id
	stepState.CurrentUserGroup = constant.UserGroupAdmin
	return StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *suite) aValidUserInEureka(ctx context.Context, id, newgroup, oldGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	num := rand.Int()
	var now pgtype.Timestamptz
	now.Set(time.Now())
	u := entities.User{}
	database.AllNullEntity(&u)
	u.ID = database.Text(id)
	u.LastName.Set(fmt.Sprintf("valid-user-%d", num))
	u.PhoneNumber.Set(fmt.Sprintf("+848%d", num))
	u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num))
	u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num))
	u.Country.Set(cpb.Country_COUNTRY_VN.String())
	u.Group.Set(oldGroup)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = now
	u.UpdatedAt = now
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	u.ResourcePath.Set("1")

	gr := &entities.Group{}
	database.AllNullEntity(gr)
	gr.ID.Set(oldGroup)
	gr.Name.Set(oldGroup)
	gr.UpdatedAt.Set(time.Now())
	gr.CreatedAt.Set(time.Now())
	fieldNames, _ := gr.FieldMap()
	placeHolders := database.GeneratePlaceholders(len(fieldNames))
	stmt := fmt.Sprintf("INSERT INTO groups (%s) VALUES(%s) ON CONFLICT DO NOTHING", strings.Join(fieldNames, ","), placeHolders)
	if _, err := s.EurekaDB.Exec(ctx, stmt, database.GetScanFields(gr, fieldNames)...); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert group error: %v", err)
	}
	cctx := s.setFakeClaimToContext(context.Background(), u.ResourcePath.String, oldGroup)

	ugroup := &entity.UserGroupV2{}
	database.AllNullEntity(ugroup)
	ugroup.UserGroupID.Set(idutil.ULIDNow())
	ugroup.UserGroupName.Set("name")
	ugroup.UpdatedAt.Set(time.Now())
	ugroup.CreatedAt.Set(time.Now())
	ugroup.ResourcePath.Set("1")

	ugMember := &entity.UserGroupMember{}
	database.AllNullEntity(ugMember)
	ugMember.UserID.Set(u.ID)
	ugMember.UserGroupID.Set(ugroup.UserGroupID.String)
	ugMember.CreatedAt.Set(time.Now())
	ugMember.UpdatedAt.Set(time.Now())
	ugMember.ResourcePath.Set("1")

	uG := entities.UserGroup{
		UserID:   u.ID,
		GroupID:  database.Text(oldGroup),
		IsOrigin: database.Bool(true),
	}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt

	role := &entity.Role{}
	database.AllNullEntity(role)
	role.RoleID.Set(idutil.ULIDNow())
	role.RoleName.Set(newgroup)
	role.CreatedAt.Set(time.Now())
	role.UpdatedAt.Set(time.Now())
	role.ResourcePath.Set("1")

	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	grantedRole.RoleID.Set(role.RoleID.String)
	grantedRole.UserGroupID.Set(ugroup.UserGroupID.String)
	grantedRole.GrantedRoleID.Set(idutil.ULIDNow())
	grantedRole.CreatedAt.Set(time.Now())
	grantedRole.UpdatedAt.Set(time.Now())
	grantedRole.ResourcePath.Set("1")

	if _, err := database.Insert(cctx, &u, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user error: %v", err)
	}

	if _, err := database.Insert(cctx, &uG, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("insert user group error: %v", err)
	}
	if _, err := database.Insert(cctx, ugroup, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(cctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.Insert(cctx, ugMember, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(cctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	if _, err := database.Insert(cctx, role, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(cctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}
	if _, err := database.Insert(cctx, grantedRole, s.EurekaDB.Exec); err != nil {
		return StepStateToContext(cctx, stepState), fmt.Errorf("insert user group member error: %v", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) setFakeClaimToContext(ctx context.Context, resourcePath string, userGroup string) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: resourcePath,
			UserGroup:    userGroup,
		},
	}
	return interceptors.ContextWithJWTClaims(ctx, claims)
}

type Permission struct {
	PermissionID   string
	PermissionName string
}

func (p *Permission) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"permission_id", "permission_name"}
	values = []interface{}{&p.PermissionID, &p.PermissionName}
	return
}

func (s *suite) getPermissionBySchoolID(ctx context.Context, resourcePath int64) ([]Permission, error) {
	stmt := "select permission_id, permission_name from permission where resource_path = $1"
	rows, err := s.BobPostgresDB.Query(ctx, stmt, fmt.Sprint(resourcePath))
	if err != nil {
		return []Permission{}, err
	}
	defer rows.Close()

	permissions := []Permission{}
	for rows.Next() {
		permission := Permission{}

		if err := rows.Scan(&permission.PermissionID, &permission.PermissionName); err != nil {
			return []Permission{}, err
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func (s *suite) addPermissionRole(ctx context.Context, permissionID, roleID string, resourcePath int64) error {
	stmt := "insert into permission_role (permission_id, role_id, created_at, updated_at, resource_path) values ($1,$2,now(),now(), $3) on conflict do nothing"
	_, err := s.BobPostgresDB.Exec(ctx, stmt, permissionID, roleID, fmt.Sprint(resourcePath))
	return err
}

func (s *suite) addPermission(ctx context.Context, permissionID, permissionName string, resourcePath int64) error {
	stmt := "insert into permission (permission_id, permission_name, created_at, updated_at, resource_path) values ($1,$2,now(),now(), $3) on conflict do nothing"
	_, err := s.BobPostgresDB.Exec(ctx, stmt, permissionID, permissionName, fmt.Sprint(resourcePath))
	return err
}

func (s *suite) addPermissionFromManabieOrg(ctx context.Context, resourcePath int64) ([]string, error) {
	permissionIDs := []string{}
	if rows, err := s.getPermissionBySchoolID(ctx, resourcePath); err == nil && len(rows) > 0 {
		for _, permission := range rows {
			permissionIDs = append(permissionIDs, permission.PermissionID)
		}
		return permissionIDs, nil
	}
	permissions, err := s.getPermissionBySchoolID(ctx, constants.ManabieSchool)
	if err != nil {
		return []string{}, err
	}
	for _, permission := range permissions {
		permissionID := s.newID()
		if err := s.addPermission(ctx, permissionID, permission.PermissionName, resourcePath); err != nil {
			return []string{}, err
		}
		permissionIDs = append(permissionIDs, permissionID)
	}
	return permissionIDs, nil
}

func (s *suite) addPermissionRoles(ctx context.Context, roleID string, resourcePath int64) error {
	permissionIDs, err := s.addPermissionFromManabieOrg(ctx, resourcePath)
	if err != nil {
		return err
	}
	for _, permission := range permissionIDs {
		err := s.addPermissionRole(ctx, permission, roleID, resourcePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *suite) addRole(ctx context.Context, roleName string, resourcePath int64) (string, error) {
	locationID := s.newID()
	stmt := "insert into role (role_id, role_name, resource_path, created_at, updated_at) values ($1, $2, $3, now(), now()) on conflict on Constraint role__pk do update set updated_at = now() returning role_id"
	rows, err := s.BobPostgresDB.Query(ctx, stmt, locationID, roleName, fmt.Sprint(resourcePath))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var roleID string
	for rows.Next() {
		if err := rows.Scan(&roleID); err != nil {
			return "", fmt.Errorf("rows.Err: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("rows.Err: %w", err)
	}

	if err != nil {
		return "", fmt.Errorf("s.addRole: %w", err)
	}
	return roleID, nil
}

func (s *suite) AddUserAccessPath(ctx context.Context, userID, locationID string, resourcePath int64, tx pgx.Tx) error {
	stmt := "insert into user_access_paths (user_id,location_id, access_path, created_at, updated_at, resource_path) values ($1,$2,$3,now(),now(), $4) on conflict do nothing"
	var err error
	if tx != nil {
		_, err = tx.Exec(ctx, stmt, userID, locationID, locationID, fmt.Sprint(resourcePath))
	} else {
		_, err = s.BobPostgresDB.Exec(ctx, stmt, userID, locationID, locationID, fmt.Sprint(resourcePath))
	}
	return err
}

func (s *suite) addMultiRoles(ctx context.Context, roleNames []string, resourcePath int64) ([]string, error) {
	roleIDs := []string{}
	for _, roleName := range roleNames {
		roleID, err := s.addRole(ctx, roleName, resourcePath)
		if err != nil {
			return []string{}, err
		}
		roleIDs = append(roleIDs, roleID)
	}
	return roleIDs, nil
}

func (s *suite) createUserGroupWithRoleNames(ctx context.Context, roleNames []string, grantedLocationIDs []string, resourcePath int64) (string, error) {
	req := &upb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
	}

	roleIDs, err := s.addMultiRoles(ctx, roleNames, resourcePath)
	if err != nil {
		return "", fmt.Errorf("s.addMultiRoles: %w", err)
	}

	for _, roleID := range roleIDs {
		err := s.addPermissionRoles(ctx, roleID, resourcePath)
		if err != nil {
			return "", fmt.Errorf("s.addPermissionRoles: %w", err)
		}
		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&upb.RoleWithLocations{
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
			return fmt.Errorf("userGroupV2Repo.Create: %w %d %s", err, resourcePath, orgLocation.LocationID)
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

func userGroupPayloadToUserGroupEnt(payload *upb.CreateUserGroupRequest, resourcePath string, orgLocation *domain.Location) (*entity.UserGroupV2, error) {
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

func roleWithLocationsPayloadToGrantedRole(payload *upb.RoleWithLocations, userGroupID string, resourcePath string) (*entity.GrantedRole, error) {
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

func addUserToUserGroup(ctx context.Context, dbBob database.QueryExecer, userID, userGroupID, resourcePath string) error {
	sql := "insert into user_group_member (user_id, user_group_id, created_at, updated_at, resource_path) values ($1, $2, now(), now(), $3) on conflict do nothing"
	_, err := dbBob.Exec(ctx, sql, userID, userGroupID, resourcePath)
	return err
}

func addUserToUserGroups(ctx context.Context, dbBob database.QueryExecer, userID string, userGroupIDs []string, resourcePath string) error {
	for _, userGroupID := range userGroupIDs {
		err := addUserToUserGroup(ctx, dbBob, userID, userGroupID, resourcePath)
		if err != nil {
			return err
		}
	}
	return nil
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
