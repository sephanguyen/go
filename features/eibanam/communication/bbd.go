package communication

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/eibanam"
	commuHelper "github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/bob/entities"
	gandalfconf "github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/cucumber/godog"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func init() {
	common.RegisterTest("eibanam.communication", &common.SuiteBuilder[gandalfconf.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	bobConn             *grpc.ClientConn
	tomConn             *grpc.ClientConn
	yasuoConn           *grpc.ClientConn
	eurekaConn          *grpc.ClientConn
	fatimaConn          *grpc.ClientConn
	shamirConn          *grpc.ClientConn
	usermgmtConn        *grpc.ClientConn
	entryExitMgmtConn   *grpc.ClientConn
	bobDB               *pgxpool.Pool
	tomDB               *pgxpool.Pool
	bobPostgresDB       *pgxpool.Pool
	eurekaDB            *pgxpool.Pool
	fatimaDB            *pgxpool.Pool
	zeusDB              *pgxpool.Pool
	firebaseAddr        string
	firebaseKey         string
	communicationHelper *commuHelper.CommunicationHelper
	jsm                 nats.JetStreamManagement
	zapLogger           *zap.Logger
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TestSuiteInitializer ...
func TestSuiteInitializer(c *gandalfconf.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	// takes too much time to refactor old types, i created a wrapper for it instead
	oldConf := &gandalf.Config{Config: *c}
	h, err := eibanam.NewHelper(
		ctx,
		oldConf,
		f.ApplicantID,
		f.FirebaseAddr,
		c.BobHasuraAdminURL,
		"https://identitytoolkit.googleapis.com",
	)
	if err != nil {
		panic(err)
	}
	helperInstance = h
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(oldConf, f.FirebaseAddr)
		})

		ctx.AfterSuite(func() {
			bobConn.Close()
			tomConn.Close()
			yasuoConn.Close()
			usermgmtConn.Close()
			eurekaConn.Close()
			fatimaConn.Close()
			shamirConn.Close()
			entryExitMgmtConn.Close()
			bobDB.Close()
			tomDB.Close()
			eurekaDB.Close()
			fatimaDB.Close()
			jsm.Close()
		})
	}
}

// nolint
func ScenarioInitializer(c *gandalfconf.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(parentCtx *godog.ScenarioContext) {
		parentCtx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
					DefaultRole:  entities.UserGroupAdmin,
					UserGroup:    entities.UserGroupAdmin,
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
			return ctx, nil
		})

		parentCtx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			uriSplit := strings.Split(sc.Uri, ":")
			uri := uriSplit[0]

			switch uri {
			case "eibanam/communication/edit_scheduled_notification.feature":
				s := NewEditScheduledNotificationSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/send_and_receive_scheduled_notification.feature":
				s := NewSendAndReceiveScheduledNotificationSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/another_school_admin_edit_scheduled_notification.feature":
				s := NewAnotherSchoolAdminEditScheduledNotificationSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/parent_leave_conversation.feature":
				s := NewParentLeaveConversationSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/teacher_leave_conversation.feature":
				s := NewTeacherLeaveConversationSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/show_error_message_after_send_scheduled_notification.feature":
				s := NewEditScheduledNotificationAfterSentSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/show_error_message_after_discard_scheduled_notification.feature":
				s := NewEditScheduledNotificationAfterDiscardSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/create_scheduled_notification_fail.feature":
				s := NewCreateScheduledNotificationFailedSuite(communicationHelper)
				s.InitScenario(parentCtx)
			case "eibanam/communication/edit_scheduled_notification_fail.feature":
				s := NewUpdateScheduledNotificationFailedSuite(communicationHelper)
				s.InitScenario(parentCtx)
			default:
				s := newSuite()
				initSteps(parentCtx, s)

				state := &stepState{
					credentials: map[string]credential{},
					Cfg:         c,
				}
				// TODO: remove this line when complete moving to using state in context
				s.stepState = state
				state.chat.streams = map[string]*stream{}
				state.studentInfos = make(map[string]studentInfo)
				state.courseInfos = make(map[string]courseInfo)
				state.users = make(map[string]*profile)
				ctx = StepStateToContext(ctx, state)
			}
			ctx, _ = context.WithTimeout(ctx, time.Minute)
			return ctx, nil
		})
	}
}

func setup(c *gandalf.Config, fakeFirebaseAddr string) {
	firebaseKey = c.FirebaseAPIKey
	firebaseAddr = fakeFirebaseAddr

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)
	err := c.ConnectGRPCInsecure(ctx, &bobConn, &tomConn, &yasuoConn, &eurekaConn, &fatimaConn, &shamirConn, &usermgmtConn, &entryExitMgmtConn)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to run BDD setup: %s", err))
	}

	c.ConnectDB(ctx, &bobDB, &tomDB, &eurekaDB, &fatimaDB, &zeusDB)
	c.ConnectSpecificDB(ctx, &bobPostgresDB)

	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, c.NatsJS.User, c.NatsJS.Password, c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)
	if err != nil {
		zapLogger.Panic("failed to connect to nats jetstream", zap.Error(err))
	}
	jsm.ConnectToJS()

	// Init auth info
	stmt := `
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
		ON CONFLICT 
			DO NOTHING
		;
		`
	if err := setupRls(bobPostgresDB); err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot setup rls %v", err))
	}
	_, err = bobPostgresDB.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
	commonConnections := &common.Connections{}
	err = commonConnections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithTransportCredentials(insecure.NewCredentials())),
		common.WithBobSvcAddress(),
		common.WithTomSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithYasuoSvcAddress(),
		common.WithUserMgmtSvcAddress(),
	)
	if err != nil {
		zapLogger.Fatal("create common connection", zap.Error(err))
	}
	err = commonConnections.ConnectDB(ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
	)
	if err != nil {
		zapLogger.Fatal("create common connection", zap.Error(err))
	}

	communicationHelper = commuHelper.NewCommunicationHelper(
		bobDB,
		bobConn,
		tomConn,
		yasuoConn,
		usermgmtConn,
		firebaseAddr,
		firebaseKey,
		c.BobHasuraAdminURL,
		jsm,
		shamirConn,
		helperInstance.ApplicantID,
		commonConnections,
	)

	SetFirebaseAddr(fakeFirebaseAddr)
}

type suite struct {
	helper      *eibanam.Helper
	commuHelper *commuHelper.CommunicationHelper
	connections
	*stepState

	// filter               filter
	repliedConversations map[conversation]struct{}
	memo                 map[string]string
	studentChatState     studentchatState
	ZapLogger            *zap.Logger
}

// identifier for conversation, because conversation ID cannot be known
type conversation struct {
	conversationType tpb.ConversationType
	studentID        string
}

func studentConversation(studentID string) conversation {
	return conversation{conversationType: tpb.ConversationType_CONVERSATION_STUDENT, studentID: studentID}
}

func parentConversation(studentID string) conversation {
	return conversation{conversationType: tpb.ConversationType_CONVERSATION_PARENT, studentID: studentID}
}

type studentInfo struct {
	id        string
	name      string
	email     string
	parents   []parentInfo
	courseIDs []string
	grade     int32
}

type parentInfo struct {
	id    string
	email string
	name  string
}

type courseInfo struct {
	name string
	id   string
}

type connections struct {
	bobConn      *grpc.ClientConn
	yasuoConn    *grpc.ClientConn
	usermgmtConn *grpc.ClientConn
	eurekaConn   *grpc.ClientConn
	tomConn      *grpc.ClientConn
	shamirConn   *grpc.ClientConn
	bobDB        *pgxpool.Pool
	eurekaDB     *pgxpool.Pool
}

type stepState struct {
	Cfg         *gandalfconf.Config
	SchoolID    string
	Request     interface{}
	Response    interface{}
	ResponseErr error

	credentials map[string]credential
	profile     struct {
		schoolAdmin profile
		admin       profile

		defaultStudent      profile
		newlyCreatedStudent profile

		defaultTeacher profile

		defaultParent    profile
		multipleParents  []profile
		newParent        profile
		anExistingParent profile
	}
	chat struct {
		id            string
		streams       map[string]*stream // key by userid
		sentMessages  []*pb.SendMessageRequest
		sessionOffset int // offset based 0 of first message in session
	}
	lesson struct {
		id            string
		name          string
		usersInLesson int
	}

	notification          *cpb.Notification
	draftNotification     *cpb.Notification
	scheduledNotification *cpb.Notification
	courseIDs             []string
	schoolID              string
	attachedLinkInNoti    string
	notiRecipientType     string
	notiIndividuals       []*profile
	studentInfos          map[string]studentInfo // map key is student identifier, not student id
	users                 map[string]*profile
	courseInfos           map[string]courseInfo // map key is course identifier, not course id
	currentUserAccount    string
}

type stream struct {
	cancel context.CancelFunc
	stream pb.ChatService_SubscribeV2Client
}

var helperInstance *eibanam.Helper

func newSuite() *suite {
	s := &suite{
		helper:      helperInstance,
		commuHelper: communicationHelper,
		connections: connections{
			bobConn:      bobConn,
			yasuoConn:    yasuoConn,
			usermgmtConn: usermgmtConn,
			eurekaConn:   eurekaConn,
			tomConn:      tomConn,
			bobDB:        bobDB,
			eurekaDB:     eurekaDB,
			shamirConn:   shamirConn,
		},
		studentChatState: studentchatState{
			newMessageBuffers: map[string][]message{},
		},
		memo:                 make(map[string]string),
		repliedConversations: make(map[conversation]struct{}),
		ZapLogger:            zapLogger,
	}
	return s
}

func checkDuplicateSteps(m map[string]interface{}, m2 map[string]interface{}) {
	for regexstring := range m2 {
		if _, exist := m[regexstring]; exist {
			panic(fmt.Sprintf("register duplicate step: %s", regexstring))
		}
		m[regexstring] = m2[regexstring]
	}
}

//nolint:unparam
func initSteps(ctx *godog.ScenarioContext, s *suite) {
	maps := []map[string]interface{}{
		s.createChatGroupSteps(),
		s.readReplyMessageSteps(),
		s.lessonChatSteps(),
		s.initStepsForNotification(),
	}

	buildRegexpMapOnce.Do(func() {
		// check duplicate, regex syntax once
		steps := map[string]interface{}{}
		for _, smallMap := range maps {
			checkDuplicateSteps(steps, smallMap)
		}
		helper.BuildRegexpMap(steps)
	})

	for _, smallMap := range maps {
		for pattern, stepFunc := range smallMap {
			ctx.Step(pattern, stepFunc)
		}
	}
}

var buildRegexpMapOnce sync.Once

func SetFirebaseAddr(fireBaseAddr string) {
	firebaseAddr = fireBaseAddr
}

func (s *suite) initStepsForNotification() map[string]interface{} {
	steps := map[string]interface{}{
		`^"([^"]*)" redirects to web browser$`:                                                                       s.redirectsToWebBrowser,
		`^"([^"]*)" has added created course for student$`:                                                           s.hasAddedCreatedCourseForStudent,
		`^"([^"]*)" has created a student with grade and parent info$`:                                               s.hasCreatedAStudentWithGradeAndParentInfo,
		`^"([^"]*)" has created (\d+) course$`:                                                                       s.hasCreatedCourse,
		`^"([^"]*)" logins Learner app$`:                                                                             s.loginsLearnerApp,
		`^"([^"]*)" receives the notification in their device$`:                                                      s.receivesTheNotificationInTheirDevice,
		`^school admin has saved a draft notification with required fields$`:                                         s.schoolAdminHasSavedADraftNotificationWithRequiredFields,
		`^school admin sends notification successfully$`:                                                             s.storeNotificationSuccessfully,
		`^school admin sends that draft notification for student and parent$`:                                        s.schoolAdminSendsThatDraftNotificationForStudentAndParent,
		`^"([^"]*)" sends notification with required fields to student and parent$`:                                  s.sendsNotificationWithRequiredFieldsToStudentAndParent,
		`^school admin sends notification$`:                                                                          s.schoolAdminSendsNotification,
		`^"([^"]*)" has added course "([^"]*)" for student "([^"]*)"$`:                                               s.hasAddedCourseForStudent,
		`^"([^"]*)" has created (\d+) courses "([^"]*)" and "([^"]*)"$`:                                              s.hasCreatedCoursesAnd,
		`^"([^"]*)" has created student "([^"]*)" with grade and parent "([^"]*)" info$`:                             s.hasCreatedStudentWithGradeAndParentInfo,
		`^"([^"]*)" has created student "([^"]*)" with grade and parent "([^"]*)", parent "([^"]*)" info$`:           s.hasCreatedStudentWithGradeAndParentParentInfo,
		`^"([^"]*)" is at Notification page$`:                                                                        s.isAtNotificationPage,
		`^"([^"]*)" login Learner App$`:                                                                              s.loginLearnerApp,
		`^school admin has created notification$`:                                                                    s.schoolAdminHasCreatedNotification,
		`^school admin sends a notification to the "([^"]*)" list in "([^"]*)", "([^"]*)", "([^"]*)"$`:               s.schoolAdminSendsANotificationToTheListIn,
		`^"([^"]*)" who relates to "([^"]*)" receive the notification$`:                                              s.whoRelatesToReceiveTheNotification,
		`^"([^"]*)" does not receive any notification$`:                                                              s.doesNotReceiveAnyNotification,
		`^"([^"]*)" has not read notification$`:                                                                      s.hasNotReadNotification,
		`^"([^"]*)" has read the notification$`:                                                                      s.hasReadTheNotification,
		`^"([^"]*)" receives notification$`:                                                                          s.receivesNotification,
		`^school admin has created a student "([^"]*)" with "([^"]*)", "([^"]*)"$`:                                   s.schoolAdminHasCreatedAStudentWith,
		`^school admin has created notification and sent for created student and parent$`:                            s.schoolAdminHasCreatedNotificationAndSentForCreatedStudentAndParent,
		`^school admin re-sends notification for unread recipients$`:                                                 s.schoolAdminResendsNotificationForUnreadRecipients,
		`^school admin sees "([^"]*)" people display in "([^"]*)" notification list on CMS$`:                         s.schoolAdminSeesPeopleDisplayInNotificationListOnCMS,
		`^school admin sees the status of "([^"]*)" is changed to "([^"]*)"$`:                                        s.schoolAdminSeesTheStatusOfIsChangedTo,
		`^matching recipients receive the notification$`:                                                             s.matchingRecipientsReceiveTheNotification,
		`^school admin has created notification that content includes hyperlink$`:                                    s.schoolAdminHasCreatedNotificationThatContentIncludesHyperlink,
		`^school admin has sent notification to student and parent$`:                                                 s.schoolAdminHasSentNotificationToStudentAndParent,
		`^"([^"]*)" interacts the hyperlink in the content on Learner App$`:                                          s.interactsTheHyperlinkInTheContentOnLearnerApp,
		`^"([^"]*)" clicks "([^"]*)" button$`:                                                                        s.clicksButton,
		`^"([^"]*)" fills scheduled notification information$`:                                                       s.fillsScheduledNotificationInformation,
		`^"([^"]*)" has created a student with grade, course and parent info$`:                                       s.hasCreatedAStudentWithGradeCourseAndParentInfo,
		`^"([^"]*)" has opened compose new notification full-screen dialog$`:                                         s.hasOpenedComposeNewNotificationFullscreenDialog,
		`^"([^"]*)" sees new scheduled notification on CMS$`:                                                         s.seesNewScheduledNotificationOnCMS,
		`^"([^"]*)" has created a draft notification$`:                                                               s.hasCreatedADraftNotification,
		`^"([^"]*)" opens editor full-screen dialog of draft notification$`:                                          s.opensEditorFullscreenDialogOfDraftNotification,
		`^"([^"]*)" sees draft notification has been saved to scheduled notification$`:                               s.seesDraftNotificationHasBeenSavedToScheduledNotification,
		`^"([^"]*)" selects date, time of schedule notification$`:                                                    s.selectsDateTimeOfScheduleNotification,
		`^"([^"]*)" selects notification status "([^"]*)"$`:                                                          s.selectsNotificationStatus,
		`^"([^"]*)" has created a scheduled notification$`:                                                           s.hasCreatedAScheduledNotification,
		`^"([^"]*)" has opened editor full-screen dialog of scheduled notification$`:                                 s.hasOpenedEditorFullscreenDialogOfScheduledNotification,
		`^"([^"]*)" is at "([^"]*)" page on CMS$`:                                                                    s.isAtPageOnCMS,
		`^"([^"]*)" sees scheduled notification has been saved to draft notification$`:                               s.seesScheduledNotificationHasBeenSavedToDraftNotification,
		`^"([^"]*)" selects status "([^"]*)"$`:                                                                       s.selectsStatus,
		`^status of scheduled notification is updated to "Sent$`:                                                     s.statusOfScheduledNotificationIsUpdatedToSent,
		`^"([^"]*)" waits for scheduled notification to be sent on time$`:                                            s.waitsForScheduledNotificationToBeSentOnTime,
		`^"([^"]*)" of "([^"]*)" logins Learner App$`:                                                                s.ofLoginsLearnerApp,
		`^scheduled notification has sent to "([^"]*)"$`:                                                             s.scheduledNotificationHasSentTo,
		`^"([^"]*)" with "([^"]*)" the scheduled notification$`:                                                      s.withTheScheduledNotification,
		`^"([^"]*)" receives notification with badge number of notification bell displays "([^"]*)" on Learner App$`: s.receivesNotificationWithBadgeNumberOfNotificationBellDisplaysOnLearnerApp,
		`^"([^"]*)" cancels to discard$`:                                                                             s.cancelsToDiscard,
		`^"([^"]*)" confirms to discard$`:                                                                            s.confirmsToDiscard,
		`^"([^"]*)" has created (\d+) "([^"]*)" notifications$`:                                                      s.hasCreatedNotifications,
		`^"([^"]*)" has opened editor full-screen dialog of "([^"]*)" notification$`:                                 s.hasOpenedEditorFullscreenDialogOfNotification,
		`^"([^"]*)" sees "([^"]*)" notification has been deleted on CMS$`:                                            s.seesNotificationHasBeenDeletedOnCMS,
		`^"([^"]*)" still sees "([^"]*)" notification on CMS$`:                                                       s.stillSeesNotificationOnCMS,
	}

	return steps
}

func StepStateFromContext(ctx context.Context) *stepState {
	state := ctx.Value(stateKeyForEibanamCommunication{})
	if state == nil {
		return &stepState{}
	}
	return state.(*stepState)
}

func StepStateToContext(ctx context.Context, state *stepState) context.Context {
	return context.WithValue(ctx, stateKeyForEibanamCommunication{}, state)
}

type stateKeyForEibanamCommunication struct{}

func setupRls(pgdb *pgxpool.Pool) error {
	st := `
		CREATE OR REPLACE function permission_check(resource_path TEXT, table_name TEXT)
		RETURNS BOOLEAN 
		AS $$
			select ($1 = current_setting('permission.resource_path') )::BOOLEAN
		$$  LANGUAGE SQL IMMUTABLE;
	`
	_, err := pgdb.Exec(context.Background(), st)
	if err != nil {
		return err
	}
	tables := []string{"info_notifications", "info_notification_msgs", "users_info_notifications"}
	creatingPolicies := map[string]string{}
	for _, item := range tables {
		policyname := fmt.Sprintf("rls_%s", item)
		creatingPolicies[policyname] = item
	}

	for policyname, table := range creatingPolicies {
		stmt := fmt.Sprintf(`CREATE POLICY %s ON "%s" using (permission_check(resource_path, '%s')) with check (permission_check(resource_path, '%s'))`,
			policyname, table, table, table)
		_, err = pgdb.Exec(context.Background(), stmt)
		if err != nil {
			if pgerr, ok := err.(*pgconn.PgError); ok {
				if pgerr.Code != pgerrcode.DuplicateObject {
					return err
				}
			} else {
				return err
			}
		}

		stmt = fmt.Sprintf(`ALTER TABLE %s ENABLE ROW LEVEL SECURITY;`, table)
		_, err = pgdb.Exec(context.Background(), stmt)
		if err != nil {
			return err
		}

		stmt = fmt.Sprintf(`ALTER TABLE %s FORCE ROW LEVEL SECURITY;`, table)
		_, err = pgdb.Exec(context.Background(), stmt)
		if err != nil {
			return err
		}
	}
	return nil
}
