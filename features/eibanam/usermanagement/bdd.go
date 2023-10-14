package usermanagement

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
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	bob_repo "github.com/manabie-com/backend/internal/bob/repositories"
	gandalfconf "github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	bob_pb "github.com/manabie-com/backend/pkg/genproto/bob"

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
	common.RegisterTest("eibanam.usermanagement", &common.SuiteBuilder[gandalfconf.Config]{
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
			userMgmtConn.Close()
			eurekaConn.Close()
			fatimaConn.Close()
			shamirConn.Close()
			entryExitMgmtConn.Close()
			zeusDB.Close()
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
	userMgmtConn          *grpc.ClientConn
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
}

type stepState struct {
	ID string

	User            *bob_entities.User
	Class           *bob_entities.Class
	School          *bob_entities.School
	CurrentSchoolID int32
	RequestStack    *requestStack
	ResponseStack   *responseStack
	// ResponseErr     error
	Payload interface{}

	UserGroupInContext   string
	UserGroupCredentials map[string]*userCredential
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
			userMgmtConn,
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
	err = s.aValidStudentInDB(id)
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
	err = studentRepo.CreateEn(ctx, s.bobDBTrace, student)
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

	schoolID := s.getSchoolId()
	err = s.aValidUser(withID(id), withRole(userGroup))
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

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^"([^"]*)" logins Learner App$`: s.loginsLearnerApp,

		// Create student
		`^"([^"]*)" logins CMS$`:                                                                  s.loginsCMS,
		`^"([^"]*)" logins Teacher App$`:                                                          s.loginsTeacherApp,
		`^school admin creates a new student with student info$`:                                  s.schoolAdminCreatesANewStudentWithStudentInfo,
		`^school admin sees newly created student on CMS$`:                                        s.schoolAdminSeesNewlyCreatedStudentOnCMS,
		`^student logins Learner App successfully with credentials which school admin gives$`:     s.studentLoginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives,
		`^new parent logins Learner App successfully with credentials which school admin gives$`:  s.newParentLoginsLearnerAppSuccessfullyWithCredentialsWhichSchoolAdminGives,
		`^school admin creates a new student with parent info$`:                                   s.schoolAdminCreatesANewStudentWithParentInfo,
		`^existed parent logins Learner App successfully with his existed credentials$`:           s.existedParentLoginsLearnerAppSuccessfullyWithHisExistedCredentials,
		`^parent sees (\d+) student\'s stats on Learner App$`:                                     s.parentSeesStudentsStatsOnLearnerApp,
		`^school admin creates a new student with existed parent info$`:                           s.schoolAdminCreatesANewStudentWithExistedParentInfo,
		`^school admin has created a student with parent info$`:                                   s.schoolAdminHasCreatedAStudentWithParentInfo,
		`^school admin creates a new student with course which has "([^"]*)"$`:                    s.schoolAdminCreatesANewStudentWithCourseWhichHas,
		`^student "([^"]*)" the course on Learner App when "([^"]*)"$`:                            s.studentTheCourseOnLearnerAppWhen,
		`^teacher sees newly created student on Teacher App$`:                                     s.teacherSeesNewlyCreatedStudentOnTeacherApp,
		`^all parent sees student\'s stats on Learner App$`:                                       s.allParentSeesStudentsStatsOnLearnerApp,
		`^school admin creates a new student with new parent, existed parent and visible course$`: s.schoolAdminCreatesANewStudentWithNewParentExistedParentAndVisibleCourse,
		`^student sees the course on Learner App$`:                                                s.studentSeesTheCourseOnLearnerApp,

		// Create teacher
		`^school admin creates a teacher$`:                                s.schoolAdminCreatesATeacher,
		`^school admin sees newly created teacher on CMS$`:                s.schoolAdminSeesNewlyCreatedTeacherOnCMS,
		`^teacher logins Teacher App successfully after forgot password$`: s.teacherLoginsTeacherAppSuccessfullyAfterForgotPassword,

		// Edit teacher
		`^school admin has created a teacher$`:                                     s.schoolAdminHasCreatedATeacher,
		`^school admin has created a student with parent info and visible course$`: s.schoolAdminHasCreatedAStudentWithParentInfoAndVisibleCourse,
		`^school admin edits teacher name$`:                                        s.schoolAdminEditsTeacherName,
		`^school admin sees the edited teacher name on CMS$`:                       s.schoolAdminSeesTheEditedTeacherNameOnCMS,
		`^teacher sees the edited teacher name on Teacher App$`:                    s.teacherSeesTheEditedTeacherNameOnTeacherApp,
		`^student sees the edited teacher name on Learner App$`:                    s.studentSeesTheEditedTeacherNameOnLearnerApp,
		`^parent sees the edited teacher name on Learner App$`:                     s.parentSeesTheEditedTeacherNameOnLearnerApp,

		// Create course
		`^school admin is on the course page$`:                s.schoolAdminIsOnTheCoursePage,
		`^school admin creates a new course$`:                 s.schoolAdminCreatesANewCourse,
		`^school admin sees the new course on CMS$`:           s.schoolAdminSeesTheNewCourseOnCMS,
		`^teacher sees the new course on Teacher App$`:        s.teacherSeesTheNewCourseOnTeacherApp,
		`^student can not see the new course on Learner App$`: s.studentCanNotSeeTheNewCourseOnLearnerApp,

		// Sync course
		`^school admin sees course on CMS$`:               s.schoolAdminSeesCourseOnCMS,
		`^system syncs course which belong to "([^"]*)"$`: s.systemSyncsCourseWhichBelongTo,
		`^teacher "([^"]*)" the course on Teacher App$`:   s.teacherTheCourseOnTeacherApp,

		`^school admin sees edited course name on CMS$`:     s.schoolAdminSeesEditedCourseNameOnCMS,
		`^student sees edited course name on Learner App$`:  s.studentSeesEditedCourseNameOnLearnerApp,
		`^system has synced course and class from partner$`: s.systemHasSyncedCourseAndClassFromPartner,
		`^system syncs course with edited course name$`:     s.systemSyncsCourseWithEditedCourseName,
		`^teacher sees edited course name on Teacher App$`:  s.teacherSeesEditedCourseNameOnTeacherApp,

		// Sync student
		`^school admin creates study plan for course$`:                                   s.schoolAdminCreatesStudyPlanForCourse,
		`^school admin sees this student on student-study plan page$`:                    s.schoolAdminSeesThisStudentOnStudentstudyPlanPage,
		`^staff creates school admin account for partner manually$`:                      s.staffCreatesSchoolAdminAccountForPartnerManually,
		`^student logins Learner App successfully with student partner account$`:         s.studentLoginsLearnerAppSuccessfullyWithStudentPartnerAccount,
		`^student sees course which student joins on Learner App$`:                       s.studentSeesCourseWhichStudentJoinsOnLearnerApp,
		`^system has synced course and (\d+) classes from partner$`:                      s.systemHasSyncedCourseAndClassesFromPartner,
		`^system has synced teacher from partner$`:                                       s.systemHasSyncedTeacherFromPartner,
		`^system syncs student account which associate with all available course-class$`: s.systemSyncsStudentAccountWhichAssociateWithAllAvailableCourseclass,
		`^teacher sees this student info on Teacher App$`:                                s.teacherSeesThisStudentInfoOnTeacherApp,
		`^school admin sees the edited name on student-study plan page$`:                 s.schoolAdminSeesTheEditedNameOnStudentstudyPlanPage,
		`^student sees the edited name on Learner App$`:                                  s.studentSeesTheEditedNameOnLearnerApp,
		`^system has synced student account which associate with all available class$`:   s.systemHasSyncedStudentAccountWhichAssociateWithAllAvailableClass,
		`^system syncs existed student with new name$`:                                   s.systemSyncsExistedStudentWithNewName,
		`^teacher sees the edited name on Teacher App$`:                                  s.teacherSeesTheEditedNameOnTeacherApp,

		// Sync class
		`^system has synced course from partner$`:                              s.systemHasSyncedCourseFromPartner,
		`^system syncs class to course which "([^"]*)" current academic year$`: s.systemSyncsClassToCourseWhichCurrentAcademicYear,
		`^teacher sees class available in class filter on Teacher App$`:        s.teacherSeesClassAvailableInClassFilterOnTeacherApp,
	}

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}
