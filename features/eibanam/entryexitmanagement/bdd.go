package entryexitmanagement

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/features/helper"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	gandalfconf "github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func init() {
	common.RegisterTest("eibanam.entryexitmanagement", &common.SuiteBuilder[gandalfconf.Config]{
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
	userMgmtConn             *grpc.ClientConn
	entryExitMgmtConn        *grpc.ClientConn
	bobDB                    *pgxpool.Pool
	tomDB                    *pgxpool.Pool
	eurekaDB                 *pgxpool.Pool
	fatimaDB                 *pgxpool.Pool
	zeusDB                   *pgxpool.Pool
	enigmaSrvURL             string
	jprepKey                 string
	jprepSignature           string
	requestAt                *timestamppb.Timestamp
	hasClassJPREPInActive    bool
	firebaseAddr             string
	applicantID              string
	bobDBTrace               *database.DBTrace
	bobHasuraAdminUrl        string
	eurekaHasuraAdminUrl     string
	googleIdentityToolkitUrl string
	zapLogger                *zap.Logger
	JSM                      nats.JetStreamManagement
)

func init() {
	rand.Seed(time.Now().UnixNano())
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
			zeusDB.Close()
			entryExitMgmtConn.Close()
			userMgmtConn.Close()
			JSM.Close()
		})
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

func setup(c *gandalf.Config, appID, fakeFirebaseAddr string) {
	firebaseAddr = fakeFirebaseAddr
	applicantID = appID
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)
	err := c.ConnectGRPCInsecure(ctx, &bobConn, &tomConn, &yasuoConn, &eurekaConn, &fatimaConn, &shamirConn, &userMgmtConn, &entryExitMgmtConn)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to run BDD setup: %s", err))
	}

	c.ConnectDB(ctx, &bobDB, &tomDB, &eurekaDB, &fatimaDB, &zeusDB)

	JSM, err = nats.NewJetStreamManagement(c.NatsJS.Address, c.NatsJS.User, c.NatsJS.Password, c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)
	if err != nil {
		zapLogger.Panic("failed to connect to nats jetstream", zap.Error(err))
	}
	JSM.ConnectToJS()
	defaultValues := (&repository.OrganizationRepo{}).DefaultOrganizationAuthValues(c.Common.Environment)
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
	eurekaHasuraAdminUrl = c.EurekaHasuraAdminURL
	googleIdentityToolkitUrl = "https://identitytoolkit.googleapis.com"

	// Init auth info
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
	_, err = bobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

type suite struct {
	*gandalf.Config
	connections
	ZapLogger *zap.Logger
	stepState
}

func (s *suite) getSchoolId() int64 {
	schoolID := int64(s.CurrentSchoolID)
	if schoolID == 0 {
		schoolID = 1
	}
	return schoolID
}

type connections struct {
	bobConn               *grpc.ClientConn
	tomConn               *grpc.ClientConn
	yasuoConn             *grpc.ClientConn
	eurekaConn            *grpc.ClientConn
	fatimaConn            *grpc.ClientConn
	entryExitMgmtConn     *grpc.ClientConn
	userMgmtConn          *grpc.ClientConn
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
	JSM                   nats.JetStreamManagement
}

type stepState struct {
	ID string

	User                 *bob_entities.User
	Class                *bob_entities.Class
	School               *bob_entities.School
	CurrentSchoolID      int32
	CurrentParentID      string
	CurrentUserGroup     string
	Request              interface{}
	Response             interface{}
	ResponseErr          error
	RequestStack         *requestStack
	ResponseStack        *responseStack
	Payload              interface{}
	UserGroupInContext   string
	UserGroupCredentials map[string]*userCredential
	ScannerResourcePath  string
	ResourcePath         string
	UserName             string
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

func newSuite(id string, c *gandalf.Config) *suite {
	return &suite{
		Config: c,
		connections: connections{
			bobConn,
			tomConn,
			yasuoConn,
			eurekaConn,
			fatimaConn,
			entryExitMgmtConn,
			userMgmtConn,
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
			JSM,
		},
		stepState: stepState{
			ID:                   id,
			UserGroupCredentials: make(map[string]*userCredential),
			RequestStack:         &requestStack{sync.Mutex{}, []interface{}{}},
			ResponseStack:        &responseStack{sync.Mutex{}, []interface{}{}},
		},
		ZapLogger: zapLogger,
	}
}

func (s *suite) aSignedInAdmin() error {
	id := idutil.ULIDNow()
	var err error
	schoolID := s.getSchoolId()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = getContextJWTClaims(ctx, fmt.Sprintf("%d", schoolID))
	err = s.aValidUser(ctx, withID(id), withRole(constant.UserGroupAdmin))
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

func (s *suite) aSignedInAdminWithResourcePath(ctx context.Context, resourcePath string) error {
	id := idutil.ULIDNow()
	schoolID := s.getSchoolId()
	err := s.aValidUserWithResourcePath(ctx, withIDInUser(id), withRoleInUser(constant.UserGroupAdmin), withResourcePathInUser(resourcePath))
	if err != nil {
		return err
	}
	authToken, err := generateExchangeToken(id, constant.UserGroupAdmin, schoolID)
	if err != nil {
		return err
	}
	s.stepState.CurrentUserGroup = constant.UserGroupAdmin
	s.UserGroupCredentials[constant.UserGroupAdmin] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: constant.UserGroupAdmin,
	}
	return nil
}

func (s *suite) aSignedInStudentWithResourcePath(ctx context.Context, resourcePath string) error {
	id := idutil.ULIDNow()
	schoolID := s.getSchoolId()
	err := s.aValidStudentInDBWithResourcePath(ctx, id, resourcePath)
	if err != nil {
		return err
	}
	authToken, err := generateExchangeToken(id, constant.UserGroupStudent, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[constant.UserGroupStudent] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: constant.UserGroupStudent,
	}

	return nil
}

func (s *suite) aSignedInStudent() error {
	id := idutil.ULIDNow()
	var err error
	schoolID := s.getSchoolId()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ctx = getContextJWTClaims(ctx, fmt.Sprintf("%d", schoolID))

	err = s.aValidStudentInDB(ctx, id)
	if err != nil {
		return err
	}
	authToken, err := generateExchangeToken(id, constant.UserGroupStudent, schoolID)
	if err != nil {
		return err
	}
	s.UserGroupCredentials[constant.UserGroupStudent] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: constant.UserGroupStudent,
	}

	return nil
}

func (s *suite) aValidStudentInDB(ctx context.Context, id string) error {
	schoolID := s.getSchoolId()
	studentRepo := &bob_repo.StudentRepo{}
	now := time.Now()
	student := &bob_entities.Student{}
	database.AllNullEntity(student)
	err := multierr.Combine(
		student.ID.Set(id),
		student.CurrentGrade.Set(12),
		student.OnTrial.Set(true),
		student.TotalQuestionLimit.Set(10),
		student.SchoolID.Set(schoolID),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
	)
	if err != nil {
		return err
	}
	err = studentRepo.CreateEn(ctx, s.bobDBTrace, student)
	if err != nil {
		return err
	}
	s.aValidUser(ctx, withID(student.ID.String), withRole(constant.UserGroupStudent))
	return err
}

func (s *suite) aValidStudentInDBWithResourcePath(ctx context.Context, id, resourcePath string) error {
	schoolID := s.getSchoolId()
	studentRepo := &bob_repo.StudentRepo{}
	now := time.Now()

	ctx = getContextJWTClaims(ctx, resourcePath)

	student := &bob_entities.Student{}
	database.AllNullEntity(student)
	err := multierr.Combine(
		student.ID.Set(id),
		student.CurrentGrade.Set(12),
		student.OnTrial.Set(true),
		student.TotalQuestionLimit.Set(10),
		student.SchoolID.Set(schoolID),
		student.CreatedAt.Set(now),
		student.UpdatedAt.Set(now),
		student.BillingDate.Set(now),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
	)
	if err != nil {
		return err
	}
	err = studentRepo.CreateEn(ctx, s.bobDBTrace, student)
	if err != nil {
		return err
	}
	err = s.aValidUserWithResourcePath(ctx, withIDInUser(id), withRoleInUser(constant.UserGroupStudent), withResourcePathInUser(resourcePath))
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

	schoolID := s.getSchoolId()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctx = getContextJWTClaims(ctx, fmt.Sprintf("%d", schoolID))
	err = s.aValidUser(ctx, withID(id), withRole(userGroup))
	if err != nil {
		return err
	}
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

func (s *suite) aValidUser(ctx context.Context, opts ...userOption) error {
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

func (s *suite) createUserInDB(ctx context.Context, user *bob_entities.User) error {
	err := database.ExecInTx(ctx, s.bobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := bob_repo.UserRepo{}
		schoolID := s.getSchoolId()
		claims := &interceptors.CustomClaims{
			Manabie: &interceptors.ManabieClaims{
				ResourcePath: fmt.Sprint(schoolID),
				UserGroup:    user.Group.String,
			},
		}
		ctx = interceptors.ContextWithJWTClaims(ctx, claims)
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return err
		}

		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := bob_repo.TeacherRepo{}
			t := &bob_entities.Teacher{}
			database.AllNullEntity(t)
			t.ID = user.ID
			t.SchoolIDs.Set([]int64{schoolID})
			err := teacherRepo.CreateMultiple(ctx, tx, []*bob_entities.Teacher{t})
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
			err = schoolAdminRepo.CreateMultiple(ctx, tx, []*bob_entities.SchoolAdmin{schoolAdminAccount})
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

func (s *suite) signedInAsAccountWithResourcePath(ctx context.Context, role, resourcePath string) error {
	var authToken string
	if role == "unauthenticated" {
		return nil
	}

	ctx = getContextJWTClaims(ctx, resourcePath)
	if role == "admin" {
		return s.aSignedInAdminWithResourcePath(ctx, resourcePath)
	}

	if role == "student" {
		return s.aSignedInStudentWithResourcePath(ctx, resourcePath)
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
		s.stepState.CurrentParentID = id
		userGroup = constant.UserGroupParent
	}

	err = s.aValidUserWithResourcePath(ctx, withIDInUser(id), withRoleInUser(userGroup), withResourcePathInUser(resourcePath))
	if err != nil {
		return err
	}

	schoolID := s.getSchoolId()
	authToken, err = generateExchangeToken(id, userGroup, schoolID)
	if err != nil {
		return err
	}
	s.stepState.CurrentUserGroup = userGroup
	s.UserGroupCredentials[userGroup] = &userCredential{
		UserID:    id,
		AuthToken: authToken,
		UserGroup: userGroup,
	}

	return nil
}

func (s *suite) aValidUserWithResourcePath(ctx context.Context, opts ...userOptionInUser) error {
	num := rand.Int()
	u := &entity.LegacyUser{}
	database.AllNullEntity(u)
	firstName := fmt.Sprintf("valid-user-first-name-%d", num)
	lastName := fmt.Sprintf("valid-user-last-name-%d", num)
	err := multierr.Combine(
		u.FullName.Set(helper.CombineFirstNameAndLastNameToFullName(firstName, lastName)),
		u.FirstName.Set(firstName),
		u.LastName.Set(lastName),
		u.PhoneNumber.Set(fmt.Sprintf("+848%d", num)),
		u.Email.Set(fmt.Sprintf("valid-user-%d@email.com", num)),
		u.Country.Set(cpb.Country_COUNTRY_VN.String()),
		u.Group.Set(constant.UserGroupAdmin),
		u.Avatar.Set(fmt.Sprintf("http://valid-user-%d", num)),
	)
	if err != nil {
		return err
	}

	for _, opt := range opts {
		opt(u)
	}
	err = s.createUserInUserRepo(ctx, u)
	if err != nil {
		return err
	}

	uGroup := &entity.UserGroup{}
	database.AllNullEntity(uGroup)

	err = multierr.Combine(
		uGroup.GroupID.Set(u.Group.String),
		uGroup.UserID.Set(u.ID.String),
		uGroup.IsOrigin.Set(true),
		uGroup.Status.Set("USER_GROUP_STATUS_ACTIVE"),
		uGroup.ResourcePath.Set(u.ResourcePath.String),
	)
	if err != nil {
		return err
	}

	userGroupRepo := &repository.UserGroupRepo{}
	err = userGroupRepo.Upsert(ctx, s.bobDBTrace, uGroup)
	if err != nil {
		return fmt.Errorf("userGroupRepo.Upsert: %w %s", err, u.Group.String)
	}
	s.stepState.UserName = u.FullName.String
	return nil
}

func (s *suite) createUserInUserRepo(ctx context.Context, user *entity.LegacyUser) error {
	err := database.ExecInTx(ctx, s.bobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		userRepo := repository.UserRepo{}
		schoolID := s.getSchoolId()
		err := userRepo.Create(ctx, tx, user)
		if err != nil {
			return fmt.Errorf("cannot create user: %w", err)
		}

		switch user.Group.String {
		case constant.UserGroupTeacher:
			teacherRepo := bob_repo.TeacherRepo{}
			t := &bob_entities.Teacher{}
			database.AllNullEntity(t)
			t.ID = user.ID
			t.SchoolIDs.Set([]int64{schoolID})
			err := teacherRepo.CreateMultiple(ctx, tx, []*bob_entities.Teacher{t})
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
				schoolAdminAccount.ResourcePath.Set(user.ResourcePath.String),
			)
			if err != nil {
				return err
			}
			err = schoolAdminRepo.CreateMultiple(ctx, tx, []*bob_entities.SchoolAdmin{schoolAdminAccount})
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
				parentEnt.ResourcePath.Set(user.ResourcePath),
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
	return nil
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		// Multitenant backoffice view Entry Exit Records
		`^"([^"]*)" logins CMS App with resource path from "([^"]*)"$`:            s.loginsWithResourcePathFrom,
		`^this school admin "([^"]*)" see the new student entry and exit record$`: s.thisSchoolAdminSeeTheNewStudentEntryAndExitRecord,
		`^this school admin creates a new student entry and exit record$`:         s.thisSchoolAdminCreatesANewStudentEntryAndExitRecord,
		`^another "([^"]*)" logins CMS App with resource path from "([^"]*)"$`:    s.loginsWithResourcePathFrom,
		// Multitenant parent view Entry Exit Records
		`^"([^"]*)" logins learner App with a resource path from "([^"]*)"$`:     s.loginsLearnerAppWithAResourcePathFrom,
		`^this parent has existing student with entry and exit record$`:          s.thisParentHasExistingStudentWithEntryAndExitRecord,
		`^"([^"]*)" visits its student\'s entry and exit record on learner App$`: s.visitsItsStudentsEntryAndExitRecordOnLearnerApp,
		`^"([^"]*)" only sees records from "([^"]*)"$`:                           s.onlySeesRecordsFrom,

		// Multitenant backoffice scan qr code
		`^scanner is setup on "([^"]*)"$`:                            s.scannerIsSetupOn,
		`^there is an existing student with qr code from "([^"]*)"$`: s.thereIsAnExistingStudentWithQrCodeFrom,
		`^this student scans qr code$`:                               s.thisStudentScansQrCode,
		`^scanner should return "([^"]*)"$`:                          s.scannerShouldReturn,
		// Multitenant student view qr code
		`^"([^"]*)" logins on Learner App$`:             s.loginsLearnerApp,
		`^"([^"]*)" with resource path from "([^"]*)"$`: s.withResourcePathFrom,
		`^"([^"]*)" has existing qr code$`:              s.hasExistingQrCode,
		`^"([^"]*)" "([^"]*)" see "([^"]*)" qr code$`:   s.seeQrCode,

		// Multitenant student view Entry Exit Records
		`^this student has existing entry and exit record$`: s.thisStudentHasExistingEntryAndExitRecord,
	}
	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}
