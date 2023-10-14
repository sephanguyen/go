package multitenant

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/internal/bob/entities"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	gandalfconf "github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	common.RegisterTest("eibanam.multitenant", &common.SuiteBuilder[gandalfconf.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	bobConn                  *grpc.ClientConn
	tomConn                  *grpc.ClientConn
	yasuoConn                *grpc.ClientConn
	eurekaConn               *grpc.ClientConn
	fatimaConn               *grpc.ClientConn
	shamirConn               *grpc.ClientConn
	usermgmtConn             *grpc.ClientConn
	entryExitMgmtConn        *grpc.ClientConn
	bobDB                    *pgxpool.Pool
	tomDB                    *pgxpool.Pool
	eurekaDB                 *pgxpool.Pool
	fatimaDB                 *pgxpool.Pool
	zeusDB                   *pgxpool.Pool
	bobPostgresDB            *pgxpool.Pool
	enigmaSrvURL             string
	jprepKey                 string
	jprepSignature           string
	requestAt                *timestamppb.Timestamp
	hasClassJPREPInActive    bool
	firebaseAddr             string
	applicantID              string
	bobDBTrace               *database.DBTrace
	bobHasuraAdminUrl        string
	googleIdentityToolkitUrl string
	zapLogger                *zap.Logger
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func setup(c *gandalf.Config, appID, fakeFirebaseAddr string) {
	firebaseAddr = fakeFirebaseAddr
	applicantID = appID
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	err := c.ConnectGRPCInsecure(ctx, &bobConn, &tomConn, &yasuoConn, &eurekaConn, &fatimaConn, &shamirConn, &usermgmtConn, &entryExitMgmtConn)
	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to run BDD setup: %s", err))
	}

	c.ConnectDB(ctx, &bobDB, &tomDB, &eurekaDB, &fatimaDB, &zeusDB)
	c.ConnectSpecificDB(ctx, &bobPostgresDB)

	enigmaSrvURL = "http://" + c.EnigmaSrvAddr
	jprepKey = c.JPREPSignatureSecret
	db, dbcancel, err := database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["bob"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zapLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()
	bobDBTrace = &database.DBTrace{
		DB: db,
	}
	bob.SetFirebaseAddr(fakeFirebaseAddr)
	bobHasuraAdminUrl = c.BobHasuraAdminURL
	googleIdentityToolkitUrl = "https://identitytoolkit.googleapis.com"
}

// TestSuiteInitializer ...
func TestSuiteInitializer(c *gandalfconf.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	oldConf := &gandalf.Config{Config: *c}
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(oldConf, f.ApplicantID, f.FirebaseAddr)
		})

		ctx.AfterSuite(func() {
			bobDB.Close()
			tomDB.Close()
			eurekaDB.Close()
			fatimaDB.Close()
			bobConn.Close()
			tomConn.Close()
			yasuoConn.Close()
			eurekaConn.Close()
			fatimaConn.Close()
			shamirConn.Close()
			entryExitMgmtConn.Close()
			bobPostgresDB.Close()
		})
	}
}

type connections struct {
	bobConn               *grpc.ClientConn
	tomConn               *grpc.ClientConn
	yasuoConn             *grpc.ClientConn
	eurekaConn            *grpc.ClientConn
	fatimaConn            *grpc.ClientConn
	bobDB                 *pgxpool.Pool
	tomDB                 *pgxpool.Pool
	eurekaDB              *pgxpool.Pool
	fatimaDB              *pgxpool.Pool
	enigmaSrvURL          string
	JPREPKey              string
	JPREPSignature        string
	RequestAt             *timestamppb.Timestamp
	HasClassJPREPInActive bool
	bobDBTrace            *database.DBTrace
	bobPostgresDB         *pgxpool.Pool
}

type requestStack struct {
	lock     sync.Mutex
	Requests []interface{}
}

func (s *requestStack) Push(v interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Requests = append(s.Requests, v)
}

func (s *requestStack) Pop() (interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.Requests)
	if l == 0 {
		return 0, errors.New("empty stack")
	}

	res := s.Requests[l-1]
	s.Requests = s.Requests[:l-1]
	return res, nil
}

func (s *requestStack) Peek() (interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.Requests)
	if l == 0 {
		return 0, errors.New("empty stack")
	}

	return s.Requests[l-1], nil
}

func (s *requestStack) PeekMulti(nItems int) ([]interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.Requests)
	if l < nItems {
		return nil, errors.New("not enough items in stack")
	}

	return s.Requests[l-nItems:], nil
}

type responseStack struct {
	lock      sync.Mutex
	Responses []interface{}
}

func (s *responseStack) Push(v interface{}) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Responses = append(s.Responses, v)
}

func (s *responseStack) Pop() (interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.Responses)
	if l == 0 {
		return 0, errors.New("empty stack")
	}

	res := s.Responses[l-1]
	s.Responses = s.Responses[:l-1]
	return res, nil
}

func (s *responseStack) Peek() (interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.Responses)
	if l == 0 {
		return 0, errors.New("empty stack")
	}

	return s.Responses[l-1], nil
}

func (s *responseStack) PeekMulti(nItems int) ([]interface{}, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	l := len(s.Responses)
	if l < nItems {
		return nil, errors.New("not enough items in stack")
	}

	return s.Responses[l-nItems:], nil
}

type userCredential struct {
	UserID    string
	AuthToken string
	UserGroup string
}

type StepState struct {
	ID string

	Value interface{}

	User            *bob_entities.User
	Class           *bob_entities.Class
	School          *bob_entities.School
	CurrentSchoolID int32
	RequestStack    *requestStack
	ResponseStack   *responseStack
	// ResponseErr     error
	Payload interface{}

	UserGroupInContext   string
	Random               string
	currentTableName     string
	UserResourcePath     map[string]string
	UserToken            map[string]string
	UserGroupCredentials map[string]*userCredential
	ResourcePath         string
}

type suite struct {
	*gandalf.Config
	connections
	ZapLogger *zap.Logger
	*StepState
}

func newSuite(id string, c *gandalf.Config) *suite {
	return &suite{
		Config: c,
		connections: connections{
			bobConn,
			tomConn,
			yasuoConn,
			eurekaConn,
			fatimaConn,
			bobDB,
			tomDB,
			eurekaDB,
			fatimaDB,
			enigmaSrvURL,
			jprepKey,
			jprepSignature,
			requestAt,
			hasClassJPREPInActive,
			bobDBTrace,
			bobPostgresDB,
		},
		StepState: &StepState{
			ID:                   id,
			currentTableName:     "",
			UserGroupCredentials: make(map[string]*userCredential),
			RequestStack:         &requestStack{sync.Mutex{}, []interface{}{}},
			ResponseStack:        &responseStack{sync.Mutex{}, []interface{}{}},
		},
		ZapLogger: zapLogger,
	}
}

// ScenarioInitializer ...
func ScenarioInitializer(c *gandalfconf.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	oldConf := &gandalf.Config{Config: *c}
	return func(ctx *godog.ScenarioContext) {
		ctx.BeforeScenario(func(p *messages.Pickle) {
			s := newSuite(p.Id, oldConf)
			initSteps(ctx, s)
		})
	}
}

func (s *suite) dbConnForSchema(schema string) *pgxpool.Pool {
	switch schema {
	case "bob":
		return s.bobDB
	case "tom":
		return s.tomDB
	case "eureka":
		return s.eurekaDB
	case "fatima":
		return s.fatimaDB
	}
	return nil
}

func (s *suite) getSchoolId() int64 {
	schoolID := int64(s.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = 1
	}
	return schoolID
}

func (s *suite) aSignedInAdmin() error {
	id := idutil.ULIDNow()
	var err error
	schoolID := s.getSchoolId()
	err = s.aValidUser(withID(id), withRole(constant.UserGroupAdmin))
	if err != nil {
		return err
	}
	authToken, err := generateExchangeToken(id, constant.UserGroupAdmin, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[constant.UserGroupAdmin] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: constant.UserGroupAdmin,
	}
	return nil
}

func (s *suite) aSignedInStudent() error {
	id := idutil.ULIDNow()
	var err error
	schoolID := s.getSchoolId()
	authToken, err := generateExchangeToken(id, constant.UserGroupStudent, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[constant.UserGroupStudent] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: constant.UserGroupStudent,
	}

	return s.aValidStudentInDB(id)
}

func (s *suite) aValidStudentInDB(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentRepo := &bob_repo.StudentRepo{}
	now := time.Now()
	student := &bob_entities.Student{}
	database.AllNullEntity(student)
	err := multierr.Combine(
		student.ID.Set(id),
		student.CurrentGrade.Set(12),
		student.OnTrial.Set(true),
		student.TotalQuestionLimit.Set(10),
		student.SchoolID.Set(1),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
	)
	if err != nil {
		return err
	}
	err = studentRepo.CreateEn(ctx, s.bobPostgresDB, student)
	if err != nil {
		return err
	}
	s.aValidUser(withID(student.ID.String), withRole(constant.UserGroupStudent))
	return err
}

// signedInAsAccount create a new user with user_group <=> role in db
// and generate token for that new user using firebase emulator with
// predefined template and shamir ExchangeToken api
func (s *suite) signedInAsAccount(role string) error {
	var authToken string
	if role == "unauthenticated" {
		return nil
	}

	if role == "admin" {
		return s.aSignedInAdmin()
	}

	if role == "student" {
		return s.aSignedInStudent()
	}

	id := idutil.ULIDNow()
	var (
		userGroup string
		err       error
	)

	if role == "teacher" {
		userGroup = constant.UserGroupTeacher
	}
	if role == "school admin" {
		userGroup = constant.UserGroupSchoolAdmin
	}
	if role == "parent" {
		userGroup = constant.UserGroupParent
	}

	err = s.aValidUser(withID(id), withRole(userGroup))
	if err != nil {
		return err
	}

	schoolID := s.getSchoolId()
	authToken, err = generateExchangeToken(id, userGroup, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[userGroup] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: userGroup,
	}

	return nil
}

func (s *suite) signedInAsAccountWithResourcePath(role, resourcePath string) error {
	var authToken string
	if role == "unauthenticated" {
		return nil
	}

	if role == "admin" {
		return s.aSignedInAdmin()
	}

	if role == "student" {
		return s.aSignedInStudent()
	}

	id := idutil.ULIDNow()
	var (
		userGroup string
		err       error
	)

	if role == "teacher" {
		userGroup = constant.UserGroupTeacher
	}
	if role == "school admin" {
		userGroup = constant.UserGroupSchoolAdmin
	}
	if role == "parent" {
		userGroup = constant.UserGroupParent
	}

	err = s.aValidUserInBob(withIDInBob(id), withRoleInBob(userGroup), withResourcePathInBob(resourcePath))
	if err != nil {
		return err
	}

	schoolID := s.getSchoolId()
	authToken, err = generateExchangeToken(id, userGroup, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[userGroup] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: userGroup,
	}

	return nil
}

func (s *suite) aValidUser(opts ...userOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	num := rand.Int()

	u := &bob_entities.User{}
	database.AllNullEntity(u)

	err := multierr.Combine(
		u.LastName.Set(fmt.Sprintf("valid-user-%d", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		u.Country.Set(bob_pb.COUNTRY_VN.String()),
		u.Group.Set(constant.UserGroupAdmin),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
	)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(u)
	}

	err = s.createUserInDB(ctx, u)
	if err != nil {
		return err
	}

	uGroup := &bob_entities.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(u.Group.String),
		uGroup.UserID.Set(u.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
	)
	if err != nil {
		return err
	}

	userGroupRepo := &bob_repo.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.bobDBTrace, uGroup)
	if err != nil {
		return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
	}

	return nil
}

func (s *suite) aValidUserInBob(opts ...userOptionInBob) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	num := rand.Int()
	u := &bob_entities.User{}
	database.AllNullEntity(u)
	now := time.Now()
	err := multierr.Combine(
		u.LastName.Set(fmt.Sprintf("valid-user-%d", num)),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		u.Country.Set(bob_pb.COUNTRY_VN.String()),
		u.Group.Set(constant.UserGroupAdmin),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
		u.CreatedAt.Set(now),
		u.UpdatedAt.Set(now),
	)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(u)
	}

	err = s.createUserInBob(ctx, u)
	if err != nil {
		return err
	}

	uGroup := &bob_entities.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(u.Group.String),
		uGroup.UserID.Set(u.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
	)
	if err != nil {
		return err
	}

	userGroupRepo := &bob_repo.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.bobDBTrace, uGroup)
	if err != nil {
		return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
	}

	return nil
}

func (s *suite) createUserInDB(ctx context.Context, user *bob_entities.User) error {
	err := database.ExecInTx(ctx, s.bobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := bob_repo.UserRepo{}
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return err
		}
		schoolID := s.getSchoolId()
		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := bob_repo.TeacherRepo{}
			t := &bob_entities.Teacher{}
			database.AllNullEntity(t)
			t.ID = user.ID
			t.SchoolIDs.Set([]int64{schoolID})
			err := teacherRepo.Create(ctx, tx, t)
			if err != nil {
				return err
			}
		case constant.UserGroupSchoolAdmin:
			schoolAdminRepo := bob_repo.SchoolAdminRepo{}
			schoolAdminAccount := &bob_entities.SchoolAdmin{}
			database.AllNullEntity(schoolAdminAccount)
			err := multierr.Combine(
				schoolAdminAccount.SchoolAdminID.Set(user.ID.String),
				schoolAdminAccount.SchoolID.Set(schoolID),
			)
			if err != nil {
				return err
			}
			err = schoolAdminRepo.CreateMultiple(ctx, tx, []*entities.SchoolAdmin{schoolAdminAccount})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) createUserInBob(ctx context.Context, user *bob_entities.User) error {
	_, err := database.Insert(ctx, user, s.bobDB.Exec)
	if err != nil {
		return err
	}

	schoolID := s.getSchoolId()
	switch user.Group.String {
	case constant.UserGroupTeacher:
		t := &bob_entities.Teacher{}
		database.AllNullEntity(t)
		t.ID = user.ID
		t.SchoolIDs.Set([]int64{schoolID})
		now := time.Now()
		t.UpdatedAt.Set(now)
		t.CreatedAt.Set(now)
		_, err := database.Insert(ctx, t, s.bobDB.Exec)
		if err != nil {
			return err
		}
	case constant.UserGroupSchoolAdmin:
		schoolAdminAccount := &bob_entities.SchoolAdmin{}
		database.AllNullEntity(schoolAdminAccount)
		now := time.Now()
		schoolAdminAccount.UpdatedAt.Set(now)
		schoolAdminAccount.CreatedAt.Set(now)
		err := multierr.Combine(
			schoolAdminAccount.SchoolAdminID.Set(user.ID.String),
			schoolAdminAccount.SchoolID.Set(schoolID),
		)
		if err != nil {
			return err
		}
		_, err = database.Insert(ctx, schoolAdminAccount, s.bobDB.Exec)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (s *suite) saveCredential(userID string, userGroup string, schoolID int64) error {
	authToken, err := generateExchangeToken(userID, userGroup, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[userGroup] = &userCredential{
		UserID:    userID,
		AuthToken: authToken,
		UserGroup: userGroup,
	}
	return nil
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		// Scan RLS
		`^scanner scans on all tables$`:               s.scannerScansOnAllTables,
		`^those tables must has rls enabled$`:         s.thoseTablesMustHasRlsEnabled,
		`^those tables must has rls forced$`:          s.thoseTablesMustHasRlsForced,
		`^a random table in db$`:                      s.aRandomTableInDB,
		`^(\d+) record with different resource path$`: s.recordWithDifferentResourcePath,
		`^rls is enable for table$`:                   s.rlsIsEnableForTable,
		`^user can only fetch their data$`:            s.userCanOnlyFetchTheirData,

		// RLS access permission
		`^"([^"]*)" logins on CMS$`:                                     s.loginsOnCMS,
		`^"([^"]*)" logs out on CMS$`:                                   s.logsOutOnCMS,
		`^"([^"]*)" only interacts with content from "([^"]*)" on CMS$`: s.onlyInteractsWithContentFromOnCMS,
		`^super admin logins on CMS$`:                                   s.superAdminLoginsOnCMS,
		`^super admin sees all data of all organization on CMS$`:        s.superAdminSeesAllDataOfAllOrganizationOnCMS,

		// RLS select course
		`^school admin logins CMS App with resource path is "([^"]*)"$`: s.loginsCMSApp,
		`^teacher logins Teacher App with resource path is "([^"]*)"$`:  s.loginsTeacherApp,
		`^enable RLS on "([^"]*)" table$`:                               s.enableRLSCourseTable,
		`^disable RLS on "([^"]*)" table$`:                              s.disableRLSCourseTable,
		`^school admin create a new course$`:                            s.schoolAdminCreatesANewCourse,
		`^teacher "([^"]*)" the new course on Teacher App$`:             s.teacherSeeNewCourseWithResourcePath,

		// RLS on create table
		`create some table with random name`:                s.createSomeTableWithRandomName,
		`those tables must have column resource_path`:       s.thoseTablesMustHaveColumnResourcePath,
		`those tables must have rls enabled and rls forced`: s.thoseTablesMustHaveRlsEnabledAndRlsForced,
	}
	s.UserToken = make(map[string]string)

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}
