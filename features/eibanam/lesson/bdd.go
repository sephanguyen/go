package lesson

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/eibanam"
	"github.com/manabie-com/backend/features/gandalf"
	gandalfconf "github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/cucumber/godog"
	"github.com/cucumber/messages-go/v16"
	"go.uber.org/zap"
)

var suiteInstance *suite

func init() {
	rand.Seed(time.Now().UnixNano())
	common.RegisterTest("eibanam.lesson", &common.SuiteBuilder[gandalfconf.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

// TestSuiteInitializer ...

func TestSuiteInitializer(c *gandalfconf.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	oldConf := &gandalf.Config{Config: *c}
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			s, err := newSuite(oldConf, f.ApplicantID, f.FirebaseAddr)
			if err != nil {
				log.Panicf("failed to run BDD setup: %s", err)
			}
			suiteInstance = s
		})

		ctx.AfterSuite(func() {
			if err := suiteInstance.Destructor(); err != nil {
				log.Printf("failed to tear down BDD: %s", err)
			}
		})
	}
}

// ScenarioInitializer ...
func ScenarioInitializer(c *gandalfconf.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		ctx.BeforeScenario(func(p *messages.Pickle) {
			suiteInstance.initSteps(ctx, p.Id)
		})
	}
}

type suite struct {
	helper    *eibanam.Helper
	ZapLogger *zap.Logger
	stepState
}

type Session struct {
	Request  interface{}
	Response interface{}
	Error    error
}

type stepState struct {
	ID string

	CurrentSchoolID int32
	RequestStack    *golibs.Stack
	ResponseStack   *golibs.Stack
	SessionStack    *golibs.Stack
	TeacherIDs      []string
	StudentIDs      []string
	CourseIDs       []string
	CenterIDs       []string
	MediaIDs        []string

	UserGroupCredentials  map[string]*eibanam.UserCredential
	CredentialsByUserName map[string]*eibanam.UserCredential
}

func newSuite(c *gandalf.Config, appID, fakeFirebaseAddr string) (*suite, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	h, err := eibanam.NewHelper(
		ctx,
		c,
		appID,
		fakeFirebaseAddr,
		c.BobHasuraAdminURL,
		"https://identitytoolkit.googleapis.com",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create helper: %s", err)
	}

	zapLogger := logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	// load example metadata file
	file, err := eibanam.GetExampleHasuraMetadata("")
	if err != nil {
		return nil, fmt.Errorf("could not read example hasura metadata file:%v", err)
	}
	if err = eibanam.ReplaceHasuraMetadata(h.HasuraAdminUrl, string(file)); err != nil {
		return nil, err
	}

	return &suite{
		helper:    h,
		ZapLogger: zapLogger,
	}, nil
}

func (s *suite) Destructor() error {
	if err := s.helper.Destructor(); err != nil {
		return err
	}

	return nil
}

func (s *suite) initSteps(ctx *godog.ScenarioContext, id string) {
	s.stepState = stepState{
		ID:                    id,
		UserGroupCredentials:  make(map[string]*eibanam.UserCredential),
		CredentialsByUserName: make(map[string]*eibanam.UserCredential),
		RequestStack:          &golibs.Stack{Elements: []interface{}{}},
		ResponseStack:         &golibs.Stack{Elements: []interface{}{}},
		SessionStack:          &golibs.Stack{Elements: []interface{}{}},
	}

	steps := map[string]interface{}{
		`^"([^"]*)" logins CMS$`:         s.loginsCMS,
		`^"([^"]*)" logins Teacher App$`: s.loginsTeacherApp,
		`^"([^"]*)" logins Learner App$`: s.loginsLearnerApp,

		`^school admin creates a new lesson with all required fields$`:      s.schoolAdminCreatesLessonWithAllRequiredFields,
		`^school admin sees the new lesson on CMS$`:                         s.schoolAdminSeesNewLessonOnCMS,
		`^teacher sees the new lesson in respective course on Teacher App$`: s.teacherSeeNewLessonInRespectiveCourseOnTeacherApp,
		`^student sees the new lesson in lesson list on Learner App$`:       s.studentSeeNewLessonInLessonListOnLearnerApp,

		`^school admin has created a live lesson on CMS$`:                                          s.schoolAdminCreatesLessonWithAllRequiredFields,
		`^school admin creates a new lesson with exact information as that lesson created before$`: s.schoolAdminCreatesANewLessonWithExactInformationAsThatLessonCreatedBefore,
		`^"([^"]*)" sees the new lesson in lesson list on Learner App$`:                            s.seeNewLessonInLessonListOnLearnerApp,
		`^"([^"]*)" sees the new lesson in respective course on Teacher App$`:                      s.seeNewLessonInRespectiveCourseOnTeacherApp,

		`^school admin creates a new lesson with "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)", teachers, learners, center, media and "([^"]*)"$`: s.schoolAdminCreatesANewLessonWithTeachersLearnersCenterMedia,
		`^school admin sees the new lesson on Lesson management$`:                                                                              s.schoolAdminSeesTheNewLessonOnLessonManagement,
	}

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}

func (s *suite) AddUserCredential(uc *eibanam.UserCredential) {
	if s.UserGroupCredentials == nil {
		s.UserGroupCredentials = make(map[string]*eibanam.UserCredential)
	}
	s.UserGroupCredentials[uc.UserGroup] = uc
}

func (s *suite) GetUserCredentialByUserGroup(gr string) (*eibanam.UserCredential, error) {
	v, ok := s.UserGroupCredentials[gr]
	if !ok {
		return nil, fmt.Errorf("could not get credential of %s", gr)
	}

	return v, nil
}

func (s *suite) AddUserCredentialByName(uc *eibanam.UserCredential, name string) {
	if s.CredentialsByUserName == nil {
		s.CredentialsByUserName = make(map[string]*eibanam.UserCredential)
	}
	s.CredentialsByUserName[name] = uc
}

func (s *suite) GetUserCredentialByUserName(userName string) (*eibanam.UserCredential, error) {
	v, ok := s.CredentialsByUserName[userName]
	if !ok {
		return nil, fmt.Errorf("could not get credential of %s", userName)
	}

	return v, nil
}

func (s *suite) AddTeacherIDs(ids ...string) {
	s.TeacherIDs = append(s.TeacherIDs, ids...)
}

func (s *suite) AddStudentIDs(ids ...string) {
	s.StudentIDs = append(s.StudentIDs, ids...)
}

func (s *suite) AddCourseIDs(ids ...string) {
	s.CourseIDs = append(s.CourseIDs, ids...)
}

func (s *suite) AddCenterIDs(ids ...string) {
	s.CenterIDs = append(s.CenterIDs, ids...)
}

func (s *suite) AddMediaIDs(ids ...string) {
	s.MediaIDs = append(s.MediaIDs, ids...)
}
