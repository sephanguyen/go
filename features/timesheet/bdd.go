package timesheet

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	user_libs "github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/bob/entities"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/gcp"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	staffEntity "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pbb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("timesheet", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	connections                   *common.Connections
	zapLogger                     *zap.Logger
	firebaseAddr                  string
	applicantID                   string
	locationIDs                   = []string{"1", "2", "3", "4", "5", "6", "7", "8"}
	initStaffID                   = idutil.ULIDNow()
	initTimesheetDate             = time.Date(2022, 1, 1, 0, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))
	initLocationTimesheet         = locationIDs[3]
	initTimesheetConfigID1        = "1"
	initTimesheetConfigID2        = "2"
	invalidTimesheetConfigID3     = "3"
	initTimesheetConfigType1      = "OTHER_WORKING_HOUR"
	initTimesheetConfigValue1     = "Office" // for ManabieSchool
	initTimesheetConfigValue2     = "TA"     // for JPREPSchool
	timesheetRemark               = "Create timesheet with other working hours"
	remarksLimit                  = 500
	otherWorkingHoursRemarksLimit = 100
	listOtherWorkingHoursLimit    = 5
	transportExpenseRemark        = "Create timesheet with transportation expense"
	transportExpenseFromToLimit   = 100
	transportExpenseRemarksLimit  = 100
	listTransportExpensesLimit    = 10
	configStatusOn                = "on"
	mapOrgStaff                   map[int]common.MapRoleAndAuthInfo
	rootAccount                   map[int]common.AuthInfo
)

const (
	insertUsersStmtFormat                   = `INSERT INTO users (%s) VALUES (%s);`
	insertStaffStmtFormat                   = `INSERT INTO staff (%s) VALUES (%s);`
	insertTimesheetStmtFormat               = `INSERT INTO timesheet (%s) VALUES (%s);`
	insertTimesheetConfigStmtFormat         = `INSERT INTO timesheet_config (%s) VALUES (%s) ON CONFLICT DO NOTHING;`
	insertTimesheetOWHsStmtFormat           = `INSERT INTO other_working_hours (%s) VALUES (%s);`
	insertTimesheetLessonHourStmtFormat     = `INSERT INTO timesheet_lesson_hours (%s) VALUES (%s);`
	insertLessonStmtFormat                  = `INSERT INTO lessons (%s) VALUES (%s);`
	insertLessonTeacherStmtFormat           = `INSERT INTO lessons_teachers (%s) VALUES (%s);`
	insertTeacherStmtFormat                 = `INSERT INTO teachers (%s) VALUES (%s);`
	insertTransportExpenseStmtFormat        = `INSERT INTO transportation_expense (%s) VALUES (%s);`
	insertMastermgmtConfigurationStmtFormat = `INSERT INTO configuration (%s) VALUES (%s);`
	insertStaffTransportExpenseStmtFormat   = `INSERT INTO staff_transportation_expense (%s) VALUES (%s);`
)

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, f.FirebaseAddr)
		})

		ctx.AfterSuite(func() {
			connections.CloseAllConnections()
		})
	}
}

func StepStateFromContext(ctx context.Context) *common.StepState {
	state := ctx.Value(common.StepStateKey{})
	if state == nil {
		return &common.StepState{}
	}
	return state.(*common.StepState)
}

func StepStateToContext(ctx context.Context, state *common.StepState) context.Context {
	return context.WithValue(ctx, common.StepStateKey{}, state)
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx = StepStateToContext(ctx, s.CommonSuite.StepState)
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: strconv.Itoa(constants.ManabieSchool),
					DefaultRole:  constant.UserGroupAdmin,
					UserGroup:    constant.UserGroupAdmin,
					UserID:       GetAdminID(constants.ManabieSchool),
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

			return ctx, nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			stepState := StepStateFromContext(ctx)
			for _, v := range stepState.Subs {
				if v.IsValid() {
					err := v.Drain()
					if err != nil {
						return nil, err
					}
				}
			}
			return ctx, nil
		})
	}
}

func GetAdminID(orgID int) string {
	return mapOrgAndAdminID[orgID]
}

var mapOrgAndAdminID = map[int]string{
	constants.ManabieSchool: "bdd_admin+manabie",
	constants.JPREPSchool:   "bdd_admin+jprep",
	constants.TestingSchool: "bdd_admin+e2e",
}

func setup(c *common.Config, fakeFirebaseAddr string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	connections = &common.Connections{}

	firebaseAddr = fakeFirebaseAddr

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	err := connections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithInsecure()),
		common.WithBobSvcAddress(),
		common.WithFatimaSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithUserMgmtSvcAddress(),
		common.WithTimesheetSvcAddress(),
		common.WithLessonMgmtSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithTimesheetPostgresDBConfig(c.PostgresV2.Databases["timesheet"], c.PostgresMigrate.Database.Password),
		common.WithMastermgmtPostgresDBConfig(c.PostgresV2.Databases["mastermgmt"], c.PostgresMigrate.Database.Password),
		common.WithMastermgmtDBConfig(c.PostgresV2.Databases["mastermgmt"]),
		common.WithAuthPostgresDBConfig(c.PostgresV2.Databases["auth"], c.PostgresMigrate.Database.Password),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect DB: %v", err))
	}

	connections.JSM, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	connections.JSM.ConnectToJS()
	applicantID = c.JWTApplicant

	// Init auth info
	defaultValues := (&repository.OrganizationRepo{}).DefaultOrganizationAuthValues(c.Common.Environment)
	stmt := fmt.Sprintf(`
			INSERT INTO organization_auths
				(organization_id, auth_project_id, auth_tenant_id)
			SELECT
				school_id, 'fake_aud', ''
			FROM
				schools
			UNION 
			SELECT
				school_id, 'dev-manabie-online', ''
			FROM
				schools
			UNION %s
			ON CONFLICT 
				DO NOTHING
			;
			`, defaultValues)
	_, err = connections.BobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	connections.GCPApp, err = gcp.NewApp(ctx, "", c.Common.IdentityPlatformProject)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	connections.FirebaseAuthClient, err = internal_auth_tenant.NewFirebaseAuthClientFromGCP(ctx, connections.GCPApp)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}

	secondaryTenantConfigProvider := &repository.TenantConfigRepo{
		QueryExecer:      connections.BobPostgresDBTrace,
		ConfigAESKey:     c.IdentityPlatform.ConfigAESKey,
		ConfigAESIv:      c.IdentityPlatform.ConfigAESIv,
		OrganizationRepo: &repository.OrganizationRepo{},
	}
	connections.TenantManager, err = internal_auth_tenant.NewTenantManagerFromGCP(ctx, connections.GCPApp, internal_auth_tenant.WithSecondaryTenantConfigProvider(secondaryTenantConfigProvider))
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create tenant manager: %v", err))
	}
	// init root account
	rootAccount, err = user_libs.InitRootAccount(ctx, connections.ShamirConn, firebaseAddr, applicantID)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init root account: %v", err))
	}

	if err := initData(ctx); err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init data: %v", err))
	}
}

func initData(ctx context.Context) error {
	err := initLocation(ctx, locationIDs)
	if err != nil {
		return err
	}

	err = initUser(ctx, initStaffID)
	if err != nil {
		return err
	}

	err = initStaff(ctx, initStaffID, strconv.Itoa(constants.ManabieSchool) /*resource_path*/)
	if err != nil {
		return err
	}

	timesheetID, err := initTimesheet(
		ctx,
		initStaffID,
		initLocationTimesheet,
		pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
		initTimesheetDate,
		strconv.Itoa(constants.ManabieSchool), /*resource_path*/
	)
	if err != nil {
		return err
	}

	err = initTimesheetConfig(ctx, initTimesheetConfigID1, initTimesheetConfigType1, initTimesheetConfigValue1, strconv.Itoa(constants.ManabieSchool))
	if err != nil {
		return err
	}

	err = initTimesheetConfig(ctx, initTimesheetConfigID2, initTimesheetConfigType1, initTimesheetConfigValue2, strconv.Itoa(constants.JPREPSchool))
	if err != nil {
		return err
	}

	_, err = initOtherWorkingHours(ctx, timesheetID, initTimesheetConfigID1, initTimesheetDate, strconv.Itoa(constants.ManabieSchool))
	if err != nil {
		return err
	}
	//
	// lessonID, err := initLesson(ctx, initStaffID, initLocationTimesheet, cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED.String(), strconv.Itoa(constants.ManabieSchool))
	// if err != nil {
	//	return err
	// }

	err = initTeachers(ctx, initStaffID, constants.ManabieSchool)
	if err != nil {
		return err
	}
	//
	// err = initLessonTeachers(ctx, lessonID, initStaffID)
	// if err != nil {
	//	return err
	// }
	//
	// err = initTimesheetLessonHours(ctx, timesheetID, lessonID, strconv.Itoa(constants.ManabieSchool))
	// if err != nil {
	//	return err
	// }

	return err
}

func initUser(ctx context.Context, userId string) error {
	email := fmt.Sprintf("valid-email+%s@gmail.com", idutil.ULIDNow())
	u := &bob_entities.User{
		ID:        database.Text(userId),
		Group:     database.Text(constant.UserGroupAdmin),
		Email:     database.Text(email),
		Country:   database.Text(cpb.Country_name[int32(cpb.Country_COUNTRY_VN)]),
		LastName:  database.Text(userId),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
	}
	fields := []string{"user_id", "email", "user_group", "country", "name", "updated_at", "created_at"}
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertUserStatement := fmt.Sprintf(insertUsersStmtFormat, strings.Join(fields, ","), placeHolder)

	_, err := connections.BobDBTrace.Exec(ctx, insertUserStatement, database.GetScanFields(u, fields)...)

	return err
}

func initStaff(ctx context.Context, userId string, resourcePath string) error {
	staff := &entity.Staff{
		StaffID:   database.Text(userId),
		CreatedAt: database.Timestamptz(time.Now()),
		UpdatedAt: database.Timestamptz(time.Now()),
		DeletedAt: pgtype.Timestamptz{Status: pgtype.Null},
	}

	fields, _ := staff.FieldMap()
	fields = append(fields, "resource_path")
	insertStaffStatement := fmt.Sprintf(insertStaffStmtFormat, strings.Join(fields, ","), database.GeneratePlaceholders(len(fields)))

	values := database.GetScanFields(staff, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertStaffStatement, values...)
	if err != nil {
		return fmt.Errorf("error inserting staff, userID: %v,error: %v", userId, err)
	}
	return nil
}

func initTimesheet(ctx context.Context, userId string, locationID string, status string, timesheetDate time.Time, resourcePath string) (string, error) {
	timesheet := entity.NewTimesheet()
	timesheet.StaffID = database.Text(userId)
	timesheet.TimesheetDate = database.Timestamptz(timesheetDate)
	timesheet.UpdatedAt = database.Timestamptz(time.Now())
	timesheet.CreatedAt = database.Timestamptz(time.Now())
	timesheet.LocationID = database.Text(locationID)
	timesheet.TimesheetStatus = database.Text(status)

	fields, _ := timesheet.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertStaffStatement := fmt.Sprintf(insertTimesheetStmtFormat, strings.Join(fields, ","), placeHolder)

	values := database.GetScanFields(timesheet, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertStaffStatement, values...)
	if err != nil {
		return "", fmt.Errorf("error inserting timesheet, userID: %v,error: %v", userId, err)
	}

	return timesheet.TimesheetID.String, err
}

func initOtherWorkingHours(ctx context.Context, timesheetID string, workingTypeID string, now time.Time, resourcePath string) (*entity.OtherWorkingHours, error) {
	owhs := entity.NewOtherWorkingHours()

	owhs.TimesheetID = database.Text(timesheetID)
	owhs.TimesheetConfigID = database.Text(workingTypeID)
	owhs.Remarks = database.Text(randStringBytes(10))
	owhs.StartTime = database.Timestamptz(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
	owhs.EndTime = database.Timestamptz(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
	owhs.TotalHour = database.Int2(30)
	owhs.UpdatedAt = database.Timestamptz(time.Now())
	owhs.CreatedAt = database.Timestamptz(time.Now())

	fields, _ := owhs.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertOWHsStatement := fmt.Sprintf(insertTimesheetOWHsStmtFormat, strings.Join(fields, ","), placeHolder)

	values := database.GetScanFields(owhs, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertOWHsStatement, values...)
	if err != nil {
		return nil, fmt.Errorf("error inserting other working hours, error: %v", err)
	}

	return owhs, err
}

func initLocation(ctx context.Context, locationIDs []string) error {
	for _, locationID := range locationIDs {
		stmt := `INSERT INTO locations (location_id,name, is_archived) VALUES($1,$2,$3)
				ON CONFLICT DO NOTHING`
		_, err := connections.TimesheetDB.Exec(ctx, stmt, locationID, locationID,
			false)
		if err != nil {
			return err
		}
	}
	return nil
}

func initLesson(ctx context.Context, teacherID, locationID, scheduledStatus, resourcePath string) (string, error) {
	now := time.Now()
	lesson := new(bob_entities.Lesson)
	database.AllNullEntity(lesson)
	lesson.LessonID.Set(database.Text(idutil.ULIDNow()))
	lesson.SchedulingStatus.Set(database.Text(scheduledStatus))
	lesson.StreamLearnerCounter.Set(database.Int4(1))
	lesson.LearnerIds.Set([]string{idutil.ULIDNow()})
	lesson.TeacherID.Set(teacherID)
	lesson.StartTime.Set(database.Timestamptz(now))
	lesson.EndTime.Set(database.Timestamptz(now.Add(time.Hour)))
	lesson.UpdatedAt.Set(database.Timestamptz(now))
	lesson.CreatedAt.Set(database.Timestamptz(now))
	lesson.IsLocked.Set(database.Bool(false))

	if err := lesson.Normalize(); err != nil {
		return "", fmt.Errorf("lesson.Normalize err: %s", err)
	}

	fields, _ := lesson.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertLessonStatement := fmt.Sprintf(insertLessonStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(lesson, fields)

	_, err := connections.BobDBTrace.Exec(ctx, insertLessonStatement, values...)
	if err != nil {
		return "", fmt.Errorf("error inserting lesson: %v", err)
	}

	return lesson.LessonID.String, nil
}

func initTimesheetLessonHours(ctx context.Context, timesheetID, lessonID, resourcePath string) error {
	now := time.Now()
	timesheetLessonHours := new(entity.TimesheetLessonHours)
	database.AllNullEntity(timesheetLessonHours)

	timesheetLessonHours.TimesheetID.Set(database.Text(timesheetID))
	timesheetLessonHours.LessonID.Set(database.Text(lessonID))
	timesheetLessonHours.UpdatedAt.Set(database.Timestamptz(now))
	timesheetLessonHours.CreatedAt.Set(database.Timestamptz(now))
	timesheetLessonHours.FlagOn.Set(database.Bool(true))
	fields, _ := timesheetLessonHours.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertTimesheetLessonHourFormat := fmt.Sprintf(insertTimesheetLessonHourStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(timesheetLessonHours, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertTimesheetLessonHourFormat, values...)
	if err != nil {
		return fmt.Errorf("error inserting timesheet lesson hours: %v", err)
	}

	return nil
}

func initTeachers(ctx context.Context, teacherID string, resourcePath int32) error {
	teacher := new(bob_entities.Teacher)
	database.AllNullEntity(teacher)

	teacher.SchoolIDs.Set([]int32{resourcePath})
	teacher.ID.Set(database.Text(teacherID))
	now := time.Now()
	if err := teacher.UpdatedAt.Set(now); err != nil {
		return err
	}
	if err := teacher.CreatedAt.Set(now); err != nil {
		return err
	}

	fields, _ := teacher.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertTeacherFormat := fmt.Sprintf(insertTeacherStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(teacher, fields)

	_, err := connections.BobDBTrace.Exec(ctx, insertTeacherFormat, values...)
	if err != nil {
		return fmt.Errorf("error inserting teacher: %v", err)
	}

	return nil
}

func initLessonTeachers(ctx context.Context, lessonID, teacherID string) error {
	lessonTeacher := new(bob_entities.LessonsTeachers)
	database.AllNullEntity(lessonTeacher)

	lessonTeacher.LessonID.Set(database.Text(lessonID))
	lessonTeacher.TeacherID.Set(database.Text(teacherID))

	fields, _ := lessonTeacher.FieldMap()
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertLessonTeacherFormat := fmt.Sprintf(insertLessonTeacherStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(lessonTeacher, fields)

	_, err := connections.BobDBTrace.Exec(ctx, insertLessonTeacherFormat, values...)
	if err != nil {
		return fmt.Errorf("error inserting lesson teacher: %v", err)
	}

	return nil
}

func initTimesheetConfig(ctx context.Context, configID, configType, configValue, resourcePath string) error {
	now := time.Now()
	config := &entity.TimesheetConfig{
		ID:          database.Text(configID),
		ConfigType:  database.Text(configType),
		ConfigValue: database.Text(configValue),
		IsArchived:  database.Bool(false),
		CreatedAt:   database.Timestamptz(now),
		UpdatedAt:   database.Timestamptz(now),
		DeletedAt:   pgtype.Timestamptz{Status: pgtype.Null},
	}

	fields, _ := config.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertConfigStatement := fmt.Sprintf(insertTimesheetConfigStmtFormat, strings.Join(fields, ","), placeHolder)
	values := database.GetScanFields(config, fields)
	values = append(values, database.Text(resourcePath))
	_, err := connections.TimesheetDB.Exec(ctx, insertConfigStatement, values...)
	if err != nil {
		return fmt.Errorf("error inserting timesheet config, error: %v", err)
	}

	return nil
}

type Suite struct {
	DB database.Ext
	*common.Connections
	*common.StepState
	ZapLogger   *zap.Logger
	Cfg         *common.Config
	CommonSuite *common.Suite
	ApplicantID string
	UserIDs     []string
}

func newSuite(c *common.Config) *Suite {
	s := &Suite{
		Connections: connections,
		Cfg:         c,
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
		CommonSuite: &common.Suite{},
	}

	s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{
		RootAccount:     rootAccount,
		FirebaseAddress: firebaseAddr,
		ApplicantID:     applicantID,
		MapOrgStaff:     mapOrgStaff,
	}

	// s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	// s.CommonSuite.StepState.ApplicantID = applicantID

	s.CommonSuite.CurrentSchoolID = constants.ManabieSchool
	// s.CommonSuite.StepState.MapOrgStaff = mapOrgStaff

	return s
}

func getStaffIDDifferenceCurrentUserID(ctx context.Context, userID string) (string, error) {
	staff := entity.Staff{}
	staffs := entity.Staffs{}

	values, _ := staff.FieldMap()
	stmt := fmt.Sprintf(`SELECT %s
	FROM %s
	WHERE staff_id != '%s' 
    AND deleted_at IS NULL
	limit 1;`, strings.Join(values, ", "), staff.TableName(), userID)

	if err := database.Select(ctx, connections.TimesheetDB, stmt).ScanAll(&staffs); err != nil {
		return "", err
	}
	for _, staff := range staffs {
		if userID != staff.StaffID.String {
			return staff.StaffID.String, nil
		}
	}

	return "", errors.New("Can not find other staff")
}

func getOneTimesheetIDInDB(ctx context.Context, userID string, resourcePath string, isForCurrentUserID bool) (string, error) {
	var (
		timesheet  = entity.Timesheet{}
		timesheets = entity.Timesheets{}
		stmt       string
	)

	values, _ := timesheet.FieldMap()

	if isForCurrentUserID {
		stmt = fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE staff_id = '%s'
		AND resource_path = '%s' 
		AND deleted_at IS NULL
		limit 1;`, strings.Join(values, ", "), timesheet.TableName(), userID, resourcePath)
	} else {
		stmt = fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE staff_id != '%s' 
		AND resource_path = '%s' 
		AND deleted_at IS NULL
		limit 1;`, strings.Join(values, ", "), timesheet.TableName(), userID, resourcePath)
	}
	if err := database.Select(ctx, connections.TimesheetDB, stmt).ScanAll(&timesheets); err != nil {
		return "", err
	}

	if len(timesheets) > 0 {
		return timesheets[0].TimesheetID.String, nil
	}

	return "", nil
}

func updateTimesheetStatusInDB(ctx context.Context, timesheetID, status string) error {
	var (
		timesheet = entity.Timesheet{}
		stmt      string
	)

	stmt = fmt.Sprintf("UPDATE %s SET timesheet_status = $2 WHERE timesheet_id = $1;", timesheet.TableName())

	_, err := connections.TimesheetDB.Exec(ctx, stmt, timesheetID, status)

	return err
}

func (s *Suite) generateExchangeToken(userID, userGroup string) (string, error) {
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

func generateRandomDate() time.Time {
	min := time.Date(2023, 1, 0, 0, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)).Unix()
	max := time.Date(2300, 1, 0, 0, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)).Unix()
	delta := max - min

	sec := rand.Int63n(delta) + min
	return time.Unix(sec, 0)
}

func (s *Suite) enterASchool(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentSchoolID = constants.ManabieSchool
	ctx, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", stepState.CurrentSchoolID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) someCenters(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aListOfLocationsInDB(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aListOfLocationsInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	listLocation := []struct {
		locationID       string
		name             string
		parentLocationID string
		archived         bool
		expected         bool
	}{ // satisfied
		{locationID: "111", archived: false, expected: true},
		{locationID: "112", parentLocationID: "111", archived: false, expected: true},
		{locationID: "113", parentLocationID: "112", archived: false, expected: true},
		{locationID: "117", archived: false, expected: true},
		// unsatisfied
		{locationID: "114", archived: true},
		{locationID: "115", parentLocationID: "114", archived: false, expected: false},
		{locationID: "116", parentLocationID: "115", archived: false, expected: false},
		{locationID: "118", parentLocationID: "117", archived: true, expected: false},
	}
	for _, l := range listLocation {
		stmt := `INSERT INTO locations (location_id,name,parent_location_id, is_archived) VALUES($1,$2,$3,$4) 
				ON CONFLICT ON CONSTRAINT locations_pkey DO NOTHING`
		_, err := s.BobDB.Exec(ctx, stmt, l.locationID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		_, err = s.TimesheetDB.Exec(ctx, stmt, l.locationID,
			l.name,
			NewNullString(l.parentLocationID),
			l.archived)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert locations with `id:%s`, %v", l.locationID, err)
		}
		if l.expected {
			stepState.LocationIDs = append(stepState.LocationIDs, l.locationID)
			stepState.CenterIDs = append(stepState.CenterIDs, l.locationID)
		}
	}
	return StepStateToContext(ctx, stepState), nil
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

func (s *Suite) clonedTeacherToTimesheetDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, v := range append(stepState.TeacherIDs, stepState.TeacherIDsUpdateLesson...) {
		err := initStaff(ctx, v, strconv.Itoa(int(stepState.CurrentSchoolID)))
		if err != nil {
			return nil, err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) CreateTeacherAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{s.newID()}
	stepState.CurrentSchoolID = constants.ManabieSchool

	for _, id := range stepState.TeacherIDs {
		if ctx, err := s.aValidTeacherProfileWithId(ctx, id, stepState.CurrentSchoolID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) Create2TeacherAccounts(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDs = []string{s.newID(), s.newID()}
	stepState.CurrentSchoolID = constants.ManabieSchool

	for _, id := range stepState.TeacherIDs {
		if ctx, err := s.aValidTeacherProfileWithId(ctx, id, stepState.CurrentSchoolID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) CreateTeacherAccountsForUpdateLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.TeacherIDsUpdateLesson = []string{s.newID(), s.newID()}
	stepState.CurrentSchoolID = constants.ManabieSchool

	for _, id := range stepState.TeacherIDsUpdateLesson {
		if ctx, err := s.aValidTeacherProfileWithId(ctx, id, stepState.CurrentSchoolID); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint:errcheck
func (s *Suite) aValidTeacherProfileWithId(ctx context.Context, id string, schoolID int32) (context.Context, error) {
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
	u.Country.Set(pbb.COUNTRY_VN.String())
	u.Group.Set(entities.UserGroupTeacher)
	u.DeviceToken.Set(nil)
	u.AllowNotification.Set(true)
	u.CreatedAt = c.CreatedAt
	u.UpdatedAt = c.UpdatedAt
	u.IsTester.Set(nil)
	u.FacebookID.Set(nil)
	uG := entities.UserGroup{UserID: c.ID, GroupID: database.Text(pbb.USER_GROUP_TEACHER.String()), IsOrigin: database.Bool(true)}
	uG.Status.Set("USER_GROUP_STATUS_ACTIVE")
	uG.CreatedAt = u.CreatedAt
	uG.UpdatedAt = u.UpdatedAt
	staff := staffEntity.Staff{}
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
	stepState.TeacherNames = append(stepState.TeacherNames, u.LastName.String)
	return ctx, nil
}

func getOneStaffIDInDB(ctx context.Context, userID string, resourcePath string) (string, error) {
	var (
		staffs entity.Staffs
		staff  entity.Staff
	)

	values, _ := staff.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE staff_id != '%s' 
		AND resource_path = '%s' 
		AND deleted_at IS NULL
		limit 1;`, strings.Join(values, ", "), staff.TableName(), userID, resourcePath)
	if err := database.Select(ctx, connections.TimesheetDB, stmt).ScanAll(&staffs); err != nil {
		return "", err
	}

	if len(staffs) > 0 {
		return staffs[0].StaffID.String, nil
	}

	return "", nil
}

func initTransportExpenses(ctx context.Context, timesheetID string, resourcePath string) (*entity.TransportationExpense, error) {
	transportExpense := entity.NewTransportExpenses()

	transportExpense.TimesheetID = database.Text(timesheetID)
	transportExpense.TransportationType = database.Text(pb.TransportationType_TYPE_BUS.String())
	transportExpense.TransportationFrom = database.Text(randStringBytes(10))
	transportExpense.TransportationTo = database.Text(randStringBytes(10))
	transportExpense.RoundTrip = database.Bool(true)
	transportExpense.CostAmount = database.Int4(10)
	transportExpense.Remarks = database.Text(randStringBytes(10))
	transportExpense.UpdatedAt = database.Timestamptz(time.Now())
	transportExpense.CreatedAt = database.Timestamptz(time.Now())

	fields, _ := transportExpense.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertTransportExpensesStatement := fmt.Sprintf(insertTransportExpenseStmtFormat, strings.Join(fields, ","), placeHolder)

	values := database.GetScanFields(transportExpense, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertTransportExpensesStatement, values...)
	if err != nil {
		return nil, fmt.Errorf("error inserting transport expenses data, error: %v", err)
	}

	return transportExpense, err
}

func LoadLocalLocation() *time.Location {
	return timeutil.Timezone(pbb.COUNTRY_JP)
}

func (s *Suite) initMasterMgmtConfiguration(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `INSERT INTO internal_configuration_value (configuration_id, config_key, config_value, config_value_type, resource_path, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) 
	ON CONFLICT ON CONSTRAINT internal_configuration_value_resource_unique DO UPDATE SET config_value = $3`
	cmd, err := connections.MasterMgmtDBTrace.Exec(ctx, stmt, idutil.ULIDNow(), "hcm.timesheet_management", "on", "string", strconv.Itoa(constants.ManabieSchool), time.Now(), time.Now())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if cmd.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no rows affected")
	}

	return StepStateToContext(ctx, stepState), nil
}

func initStaffTransportExpenses(ctx context.Context, staffId, locationId, resourcePath string) (*entity.StaffTransportationExpense, error) {
	staffTransportExpense := entity.NewStaffTransportationExpense()

	staffTransportExpense.StaffID = database.Text(staffId)
	staffTransportExpense.LocationID = database.Text(locationId)
	staffTransportExpense.TransportationType = database.Text(pb.TransportationType_TYPE_BUS.String())
	staffTransportExpense.TransportationFrom = database.Text(randStringBytes(10))
	staffTransportExpense.TransportationTo = database.Text(randStringBytes(10))
	staffTransportExpense.RoundTrip = database.Bool(true)
	staffTransportExpense.CostAmount = database.Int4(10)
	staffTransportExpense.Remarks = database.Text(randStringBytes(10))
	staffTransportExpense.UpdatedAt = database.Timestamptz(time.Now())
	staffTransportExpense.CreatedAt = database.Timestamptz(time.Now())

	fields, _ := staffTransportExpense.FieldMap()
	fields = append(fields, "resource_path")
	placeHolder := database.GeneratePlaceholders(len(fields))
	insertTransportExpensesStatement := fmt.Sprintf(insertStaffTransportExpenseStmtFormat, strings.Join(fields, ","), placeHolder)

	values := database.GetScanFields(staffTransportExpense, fields)
	values = append(values, database.Text(resourcePath))

	_, err := connections.TimesheetDB.Exec(ctx, insertTransportExpensesStatement, values...)
	if err != nil {
		return nil, fmt.Errorf("error inserting staff transport expenses data, error: %v", err)
	}

	return staffTransportExpense, err
}

func getOneLocationIDInDB(ctx context.Context, resourcePath string) (string, error) {
	var (
		locations entity.Locations
		location  entity.Location
	)

	values, _ := location.FieldMap()

	stmt := fmt.Sprintf(`SELECT %s
		FROM %s
		WHERE resource_path = '%s' 
		AND deleted_at IS NULL
		limit 1;`, strings.Join(values, ", "), location.TableName(), resourcePath)
	if err := database.Select(ctx, connections.TimesheetDB, stmt).ScanAll(&locations); err != nil {
		return "", err
	}

	if len(locations) > 0 {
		return locations[0].LocationID.String, nil
	}

	return "", nil
}
