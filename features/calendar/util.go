package calendar

import (
	"context"
	crypto_rand "crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

type userOption func(u *entity.LegacyUser)

const (
	NilValue   string = "nil"
	TimeLayout string = "2006-01-02"

	studentType         = "student"
	teacherType         = "teacher"
	parentType          = "parent"
	schoolAdminType     = "school admin"
	organizationType    = "organization manager"
	unauthenticatedType = "unauthenticated"
)

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

func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

func NotMatchError(field, expected, actual interface{}) error {
	return fmt.Errorf(fmt.Sprintf("not match %s: expected %s, actual %s", field, expected, actual))
}

func LoadLocalLocation() *time.Location {
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	return loc
}

func buildAccessPath(rootLocation, rand string, locationPrefixes []string) string {
	rs := rootLocation
	for _, str := range locationPrefixes {
		rs += "/" + str + rand
	}
	return rs
}

func (s *suite) enterASchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aSignedInAsSchoolAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.SignedAsAccountV2(ctx, "school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.CurrentSchoolID = constants.ManabieSchool

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) SignedAsAccountV2(ctx context.Context, account string) (context.Context, error) {
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

func (s *suite) returnRootContext(ctx context.Context) context.Context {
	return common.ValidContext(ctx, constants.ManabieSchool, s.RootAccount[constants.ManabieSchool].UserID, s.RootAccount[constants.ManabieSchool].Token)
}

func (s *suite) signedCtx(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)
	return helper.GRPCContext(ctx, "token", stepState.AuthToken)
}

func (s *suite) aValidUser(ctx context.Context, opts ...userOption) error {
	stepState := StepStateFromContext(ctx)

	schoolID := int64(stepState.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = constants.ManabieSchool
	}
	ctx = auth.InjectFakeJwtToken(ctx, fmt.Sprint(schoolID))

	user, err := newUserEntity()
	if err != nil {
		return errors.Wrap(err, "newUserEntity")
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
		}

		return nil
	})
	if err != nil {
		return err
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
		return err
	}

	userGroupRepo := &repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.BobDBTrace, uGroup)
	if err != nil {
		return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, user.Group.String)
	}

	return nil
}

func newUserEntity() (*entity.LegacyUser, error) {
	userID := idutil.ULIDNow()
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

func (s *suite) anExistingLocationInCalendarDB(ctx context.Context, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()

	loc := &dto.Location{
		LocationID: location,
		Name:       location + "-name",
	}

	sql := `INSERT INTO locations (location_id,name,created_at,updated_at) VALUES($1,$2,$3,$4) 
				ON CONFLICT ON CONSTRAINT location_pk
				DO UPDATE set updated_at = $4`

	if _, err := s.CalendarDBTrace.Exec(ctx, sql, loc.LocationID, loc.Name, now, now); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location with `id:%s`, %v", loc.LocationID, err)
	}

	stepState.LocationID = location

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anExistingDateTypeInDB(ctx context.Context, dateType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	sql := `INSERT INTO day_type (day_type_id) VALUES($1) 
				ON CONFLICT ON CONSTRAINT day_type_pk DO NOTHING`

	if _, err := s.CalendarDBTrace.Exec(ctx, sql, dateType); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert date type with `id:%s`, %v", dateType, err)
	}

	stepState.DateTypeID = dateType
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) someExistingLocations(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aListOfLocationTypesInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	ctx, err = s.aListOfLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

type CreateLocation struct {
	locationID        string
	partnerInternalID string
	name              string
	parentLocationID  string
	archived          bool
	expected          bool
	accessPath        string
	locationType      string
}

func (s *suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	nBig, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(27))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	addedRandom := "-" + strconv.Itoa(int(nBig.Int64()))

	listLocation := []CreateLocation{
		// satisfied
		{locationID: "1" + addedRandom, partnerInternalID: "partner-internal-id-1" + addedRandom, locationType: "locationtype-id-4", parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1"})},
		{locationID: "2" + addedRandom, partnerInternalID: "partner-internal-id-2" + addedRandom, locationType: "locationtype-id-5", parentLocationID: "1" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2"})},
		{locationID: "3" + addedRandom, partnerInternalID: "partner-internal-id-3" + addedRandom, locationType: "locationtype-id-6", parentLocationID: "2" + addedRandom, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"1", "2", "3"})},
		{locationID: "7" + addedRandom, partnerInternalID: "partner-internal-id-7" + addedRandom, locationType: "locationtype-id-7", parentLocationID: stepState.LocationID, archived: false, expected: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7"})},
		// unsatisfied
		{locationID: "4" + addedRandom, partnerInternalID: "partner-internal-id-4" + addedRandom, locationType: "locationtype-id-8", parentLocationID: stepState.LocationID, archived: true, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4"})},
		{locationID: "5" + addedRandom, partnerInternalID: "partner-internal-id-5" + addedRandom, locationType: "locationtype-id-9", parentLocationID: "4" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "6" + addedRandom, partnerInternalID: "partner-internal-id-6" + addedRandom, locationType: "locationtype-id-1", parentLocationID: "5" + addedRandom, archived: false, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"4", "5"})},
		{locationID: "8" + addedRandom, partnerInternalID: "partner-internal-id-8" + addedRandom, locationType: "locationtype-id-2", parentLocationID: "7" + addedRandom, archived: true, expected: false, accessPath: buildAccessPath(stepState.LocationID, addedRandom, []string{"7", "8"})},
	}

	if stepState.CreateStressTestLocation {
		listLocation = append(listLocation,
			CreateLocation{
				locationID:        "VCSTRESSTESTLOCATION",
				partnerInternalID: "partner-internal-id-99" + addedRandom,
				locationType:      "locationtype-id-4",
				parentLocationID:  stepState.LocationID,
				archived:          false,
				expected:          true,
				accessPath:        buildAccessPath(stepState.LocationID, addedRandom, []string{"VCSTRESSTESTLOCATION"}),
			},
		)
	}

	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,partner_internal_id,name,parent_location_id, is_archived, access_path, location_type) VALUES($1,$2,$3,$4,$5,$6,$7) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, l.locationID, l.partnerInternalID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived, l.accessPath,
			l.locationType)
		if err != nil {
			claims := interceptors.JWTClaimsFromContext(ctx)
			fmt.Println("claims: ", claims.Manabie.UserID, claims.Manabie.ResourcePath)
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
			stepState.CenterIDs = append(stepState.CenterIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aListOfLocationTypesInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	listLocationTypes := []struct {
		locationTypeID       string
		name                 string
		parentLocationTypeID string
		archived             bool
		expected             bool
	}{
		// satisfied
		{locationTypeID: "locationtype-id-1", name: "org test", expected: true},
		{locationTypeID: "locationtype-id-2", name: "brand test", parentLocationTypeID: "locationtype-id-1", expected: true},
		{locationTypeID: "locationtype-id-3", name: "area test", parentLocationTypeID: "locationtype-id-1", expected: true},
		{locationTypeID: "locationtype-id-4", name: "center test", parentLocationTypeID: "locationtype-id-2", expected: true},
		{locationTypeID: "locationtype-id-10", name: "center-10", parentLocationTypeID: "locationtype-id-2", expected: true},

		// unsatisfied
		{locationTypeID: "locationtype-id-5", name: "test-5", archived: true},
		{locationTypeID: "locationtype-id-6", name: "test-6", parentLocationTypeID: "locationtype-id-5"},
		{locationTypeID: "locationtype-id-7", name: "test-7", parentLocationTypeID: "locationtype-id-6"},
		{locationTypeID: "locationtype-id-8", name: "test-8", parentLocationTypeID: "locationtype-id-10", archived: true},
		{locationTypeID: "locationtype-id-9", name: "test-9", parentLocationTypeID: "locationtype-id-8"},
	}

	for _, lt := range listLocationTypes {
		stmt := `INSERT INTO location_types (location_type_id,name,parent_location_type_id, is_archived,updated_at,created_at) VALUES($1,$2,$3,$4,now(),now()) 
				ON CONFLICT DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, lt.locationTypeID,
			lt.name,
			NewNullString(lt.parentLocationTypeID),
			lt.archived)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert location types with `id:%s`, %v", lt.locationTypeID, err)
		}
		if lt.expected {
			stepState.LocationTypesID = append(stepState.LocationTypesID, lt.locationTypeID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
