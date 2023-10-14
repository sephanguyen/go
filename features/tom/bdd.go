package tom

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/features/yasuo"
	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	tomCfg "github.com/manabie-com/backend/internal/tom/configurations"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/cucumber/godog"
	"github.com/go-kafka/connect"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func init() {
	common.RegisterTest("tom", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

// new setup, to avoid conflict
var (
	connections *common.Connections
	tomConfig   tomCfg.Config
)

// old setup
var (
	firebaseAddr string
	applicantID  string

	conn              *grpc.ClientConn
	shamirConn        *grpc.ClientConn
	grpcWebConn       *grpcweb.ClientConn
	db                *pgxpool.Pool
	bobDBTrace        *database.DBTrace
	masterMgmtDBTrace *database.DBTrace
	// jsm         nats.JetStreamManagement

	zapLogger     *zap.Logger
	otelFlushFunc = func() {} // noop by default
	searchClient  *elastic.SearchFactoryImpl
)

// func initOtel(c *common.Config) trace.TracerProvider {
//         _, tp, flush := interceptors.InitTelemetry(&c.Common, "tom-gandalf", 1)
//         if flush != nil {
//                 otelFlushFunc = flush
//         }
//         return tp
// }

// nolint:gosec
func connectGrpcWeb(addr string) (*grpcweb.ClientConn, error) {
	return grpcweb.DialContext(addr, grpcweb.WithInsecure())
}

// nolint:unparam
func setup(c *common.Config, fakeFirebaseAddr, inputApplicantID string, otelEnabled bool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	firebaseAddr = fakeFirebaseAddr
	applicantID = inputApplicantID
	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)
	var err error
	opts := []grpc.DialOption{}

	// if otelEnabled {
	// trprovider := initOtel(c)
	// opts = append(opts, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor(otelgrpc.WithTracerProvider(trprovider))))
	// }

	connections = &common.Connections{}
	err = connections.ConnectGRPC(ctx,
		common.WithCredentials(grpc.WithTransportCredentials(insecure.NewCredentials())),
		common.WithDialOptions(opts...),
		common.WithBobSvcAddress(),
		common.WithTomSvcAddress(),
		common.WithShamirSvcAddress(),
		common.WithYasuoSvcAddress(),
		common.WithUserMgmtSvcAddress(),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connect GRPC Server: %v", err))
	}

	err = connections.ConnectDB(
		ctx,
		common.WithBobDBConfig(c.PostgresV2.Databases["bob"]),
		common.WithTomDBConfig(c.PostgresV2.Databases["tom"]),
		common.WithMastermgmtPostgresDBConfig(c.PostgresV2.Databases["mastermgmt"], c.PostgresMigrate.Database.Password),
		common.WithBobPostgresDBConfig(c.PostgresV2.Databases["bob"], c.PostgresMigrate.Database.Password),
		common.WithTomPostgresDBConfig(c.PostgresV2.Databases["tom"], c.PostgresMigrate.Database.Password),
	)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("failed to connectDB: %v", err))
	}

	connections.JSM, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	connections.JSM.ConnectToJS()

	conn = connections.TomConn
	shamirConn = connections.ShamirConn
	db = connections.TomDB
	bobDBTrace = connections.BobDBTrace
	masterMgmtDBTrace = connections.MasterMgmtPostgresDBTrace

	grpcweb, err := connectGrpcWeb("tom-grpc-web:5151")
	if err != nil {
		zapLogger.Fatal("can't connect grpc web", zap.Error(err))
	}
	grpcWebConn = grpcweb

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

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
	_, err = bobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}

	tomConfig = tomCfg.Config{
		Common:     c.Common,
		PostgresV2: c.PostgresV2,
	}

	searchClient, err = elastic.NewSearchFactory(zapLogger, c.ElasticSearch.Addresses, c.ElasticSearch.Username, c.ElasticSearch.Password, "", "")
	if err != nil {
		zapLogger.Fatal("unable to connect elasticsearch", zap.Error(err))
	}

	if err := setupRls(ctx, connections.TomPostgresDB); err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot setup rls %v", err))
	}
	_, err = connections.TomPostgresDB.Exec(ctx, insertDefaultLocations)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
	err = waitForKafkaConnect(c.KafkaConnectConfig.Addr)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("waiting for kafka connect: %v", err))
	}
}

func waitForKafkaConnect(addr string) error {
	tomConnectors := []string{
		"local_manabie_bob_source_connector",
		"local_manabie_bob_to_tom_locations_sink_connector",
	}
	cl := connect.NewClient(addr)
	checkConnectorStatusFunc := func(cl *connect.Client, connector string) (bool, error) {
		stat, res, err := cl.GetConnectorStatus(connector)
		if err != nil {
			if res.StatusCode == 404 {
				return true, fmt.Errorf("connector %s is not created", connector)
			}
			return false, err
		}
		defer func() {
			if err = res.Body.Close(); err != nil {
				err = fmt.Errorf("res.Body.Close() error: %w", err)
			}
		}()
		if len(stat.Tasks) == 0 {
			return true, fmt.Errorf("connector %s has no task", connector)
		}
		for _, task := range stat.Tasks {
			if task.State != "RUNNING" {
				return true, fmt.Errorf("task %d of connector %s is not running", task.ID, connector)
			}
		}
		return false, nil
	}
	return doRetry(func() (bool, error) {
		for _, connector := range tomConnectors {
			isReady, err := checkConnectorStatusFunc(cl, connector)
			if err != nil {
				return isReady, err
			}
		}
		return false, nil
	})
}

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() { setup(c, f.FirebaseAddr, f.ApplicantID, f.OtelEnabled) })
		ctx.AfterSuite(func() {
			if connections != nil {
				connections.CloseAllConnections()
			}
			otelFlushFunc()
		})
	}
}

func traceStepSpanKey(id string) string {
	return "x-trace-step-key" + id
}

type (
	traceScenarioSpanKey struct{}
	cancelFuncKey        struct{}
)

func getTimeOutForScenario(sc *godog.Scenario) time.Duration {
	return 45 * time.Second
	// TODO: Change back to 30 * time.Second after improve test performance
	// realURI := strings.Split(sc.Uri, ":")[0]
	// parts := strings.Split(realURI, "/")
	// suffix := parts[len(parts)-1]
	// switch suffix {
	// case "reproduce_chat_missing_message.feature":
	// 	return 45 * time.Second
	// default:
	// 	return 30 * time.Second
	// }
}

func setupTraceForStepFuncs(ctx *godog.ScenarioContext) {
	ctx.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		// Get parent scenario span, so that next span created is a children of this span
		// instead of the current span in the context, which should be its cousin
		sometype := ctx.Value(traceScenarioSpanKey{})
		if sometype != nil {
			parSpan, ok := sometype.(interceptors.TimedSpan)
			if ok {
				ctx = trace.ContextWithSpan(ctx, parSpan.Span())
			}
		}
		ctx, stepspan := interceptors.StartSpan(ctx, st.Text)
		ctx = context.WithValue(ctx, traceStepSpanKey(st.Id), stepspan) //nolint:revive,staticcheck
		return ctx, nil
	})

	ctx.StepContext().After(func(ctx context.Context, st *godog.Step, stat godog.StepResultStatus, err error) (context.Context, error) {
		someval := ctx.Value(traceStepSpanKey(st.Id))
		if someval != nil {
			stepspan, ok := someval.(interceptors.TimedSpan)
			if ok {
				if err != nil {
					stepspan.RecordError(err)
				}
				stepspan.End()
			}
		}

		return ctx, nil
	})
}

func newSuite(c *common.Config) *suite {
	s := &suite{Cfg: c, ZapLogger: zapLogger, CommonSuite: &common.Suite{}}
	s.Conn = conn
	s.DB = db
	s.bobDBTrace = bobDBTrace
	s.masterMgmtDBTrace = masterMgmtDBTrace
	s.ShamirConn = shamirConn
	s.GrpcWebConn = grpcWebConn

	s.JSM = connections.JSM
	s.CommonSuite.Connections = connections
	s.CommonSuite.StepState = &common.StepState{}
	s.CommonSuite.StepState.FirebaseAddress = firebaseAddr
	s.CommonSuite.StepState.ApplicantID = applicantID
	return s
}

func setupCustomTags(ctx *godog.ScenarioContext) {
	ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		for _, t := range sc.Tags {
			if t.Name == "@throttle" {
				time.Sleep(2 * time.Second)
			}
		}
		return ctx, nil
	})
}

func mapFeaturesToStepFuncs(parctx *godog.ScenarioContext, conf *common.Config) {
	parctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		uriSplit := strings.Split(sc.Uri, ":")
		uri := uriSplit[0]

		switch uri {
		case "tom/tom_migration_scripts.feature":
			ctx = TomMigrationScriptStateToCtx(ctx, &StepState{
				locationPool: map[string]string{},
			})
			// some legacy code here, don't imitate me
			ctx2, err := yasuo.InitYasuoState(ctx)
			migrationSuite := newOldTomGandalfSuite(conf, connections)
			yasuo.SetFirebaseAddr(firebaseAddr)
			bob.SetFirebaseAddr(firebaseAddr)
			tomMigrationScriptSteps(parctx, migrationSuite)
			return ctx2, err

		default:
			s := newSuite(conf)
			s.StreamClients = make(map[string]pb.ChatService_StreamingEventClient)
			s.SubV2Clients = make(map[string]cancellableStream)
			s.ConversationMembers = make(map[string][]entities.ConversationMembers)
			s.parentChats = make(map[string]chatInfo)
			s.teacherTokens = make(map[string]string)
			s.LessonChatState = &LessonChatState{
				LessonConversationMap: make(map[string]string),
			}

			initStep(parctx, s)
			return ctx, nil
		}
	})
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	rand.Seed(time.Now().Unix())
	return func(ctx *godog.ScenarioContext) {
		// var cancel context.CancelFunc (this is not thread safe)
		setupTraceForStepFuncs(ctx)
		setupCustomTags(ctx)
		mapFeaturesToStepFuncs(ctx, c)
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx, span := interceptors.StartSpan(ctx, fmt.Sprintf("Starting: %s", sc.Name))
			// children steps need this parent span
			ctx = context.WithValue(ctx, traceScenarioSpanKey{}, span)
			timeout := getTimeOutForScenario(sc)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			ctx = context.WithValue(ctx, cancelFuncKey{}, cancel)
			return ctx, nil
		})
		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			cancel, ok := ctx.Value(cancelFuncKey{}).(context.CancelFunc)
			if ok {
				defer cancel()
			}
			sometype := ctx.Value(traceScenarioSpanKey{})
			traceID := "undefined"
			if sometype != nil {
				span, ok := sometype.(interceptors.TimedSpan)
				if ok {
					defer span.End()
					traceID = span.Span().SpanContext().TraceID().String()
					if err != nil {
						span.SetAttributes(attribute.KeyValue{
							Key:   "x-has-error",
							Value: attribute.StringValue("true"),
						})
						span.RecordError(err)
					}
				}
			}
			if err != nil {
				return ctx, fmt.Errorf("traceID %s: %s", traceID, err)
			}

			return ctx, err
		})
	}
}

type cancellableStream struct {
	pb.ChatService_SubscribeV2Client
	cancel context.CancelFunc
}
type suite struct {
	Cfg                                                           *common.Config
	DB                                                            database.Ext
	bobDBTrace                                                    *database.DBTrace
	masterMgmtDBTrace                                             *database.DBTrace
	Conn                                                          *grpc.ClientConn
	GrpcWebConn                                                   *grpcweb.ClientConn
	ShamirConn                                                    *grpc.ClientConn
	JSM                                                           nats.JetStreamManagement
	StreamClient                                                  pb.ChatService_SubscribeClient
	ZapLogger                                                     *zap.Logger
	ApplicantID                                                   string
	TeacherToken, studentToken, schoolAdminToken, parentToken     string
	teacherWhoLeftChat                                            string
	invalidLeavingChat                                            string
	Request                                                       interface{}
	Response                                                      interface{}
	ResponseErr                                                   error
	studentID, conversationID, teacherID, schoolAdminID, parentID string
	teacherIDs                                                    []string
	classID                                                       int32
	RequestAt                                                     time.Time
	AuthToken                                                     string
	schoolID                                                      string
	chatName                                                      string
	ConversationMembers                                           map[string][]entities.ConversationMembers
	StreamClients                                                 map[string]pb.ChatService_StreamingEventClient
	SubV2Clients                                                  map[string]cancellableStream
	StudentsInLesson                                              []string
	StudentRaisedHandInLesson                                     []string
	StudentPutHandDownInLesson                                    []string
	CurrentStudentQuestionID                                      string
	ConversationMembersLatestEvent                                []*tpb.RetrieveConversationMemberLatestEventResponse
	SchoolIds                                                     []string
	ConversationIDs                                               []string
	ParentConversationIDs                                         []string
	ParentIDs                                                     []string
	StudentIDs                                                    []string
	OldConversations                                              map[string]Message
	JoinedConversationIDs                                         []string
	LeftConversationIDs                                           []string
	singleParentID                                                string
	parentIDs                                                     []string
	childrenIDs                                                   []string
	parentChats                                                   map[string]chatInfo
	teachersInConversation                                        []string
	teacherWhoSentMessage                                         string
	parentWhoSentMessage                                          string
	teacherTokens                                                 map[string]string
	filterSuiteState                                              filterSuiteState
	commonState                                                   commonState
	sentMessages                                                  []*pb.SendMessageRequest
	additionalParentID                                            string
	messageID                                                     string
	studentMessageID                                              string
	senderToken                                                   string
	lastMessage                                                   string
	receiverToken                                                 string
	senderTokens                                                  []string
	privateConversationIDs                                        []string
	userGroupIDs                                                  []string
	TeacherProfileMap                                             map[string]*upb.CreateStaffResponse_StaffProfile
	StudentIDAndParentIDMap                                       map[string]string
	*LessonChatState

	CommonSuite *common.Suite
}
type LessonChatState struct {
	firstTeacher          string
	secondTeacher         string
	lessonID              string
	lessonName            string
	LessonConversationMap map[string]string
	studentsInLesson      []string
	TeachersInLesson      []string
}
type Suite struct{ suite }

func (s *suite) SetBobDBTrace(d *database.DBTrace) { s.bobDBTrace = d }

func SetFirebaseAddr(fireBaseAddr string) { firebaseAddr = fireBaseAddr }
func generateValidAuthenticationToken(sub string, userGroup string) (string, error) {
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=templates/" + userGroup + ".template" + "&UserID=" + sub)
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

func GenerateFakeAuthenticationToken(firebaseAddr, sub string, userGroup string) (string, error) {
	template := "templates/" + userGroup + ".template"
	resp, err := http.Get("http://" + firebaseAddr + "/token?template=" + template + "&UserID=" + sub)
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

// nolint
func (s *suite) aInvalidToken(ctx context.Context, arg1 string) (context.Context, error) {
	if arg1 == "student" {
		s.studentToken = "invalid-token"
	} else if arg1 == "teacher" {
		s.TeacherToken = "invalid-token"
	}
	return ctx, nil
}

func newCommonSuite() *common.Suite {
	csuite := &common.Suite{}
	csuite.Connections = connections
	csuite.StepState = &common.StepState{}
	csuite.StepState.FirebaseAddress = firebaseAddr
	csuite.StepState.ApplicantID = applicantID
	return csuite
}

func (s *suite) generateExchangeTokens(userIDs []string, userGroup, applicantID string, schoolID int64, conn *grpc.ClientConn) ([]string, error) {
	toks := make([]string, 0, len(userIDs))
	for _, userID := range userIDs {
		firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
		if err != nil {
			return nil, err
		}
		token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, applicantID, schoolID, conn)
		if err != nil {
			return nil, err
		}
		toks = append(toks, token)
	}
	return toks, nil
}

func (s *suite) genTeacherToken(userID string) (string, error) {
	return s.generateExchangeToken(userID, cpb.UserGroup_USER_GROUP_TEACHER.String(), applicantID, s.getSchool(), s.ShamirConn)
}

func (s *suite) genStudentToken(userID string) (string, error) {
	return s.generateExchangeToken(userID, cpb.UserGroup_USER_GROUP_STUDENT.String(), applicantID, s.getSchool(), s.ShamirConn)
}

func (s *suite) genParentToken(userID string) (string, error) {
	return s.generateExchangeToken(userID, cpb.UserGroup_USER_GROUP_PARENT.String(), applicantID, s.getSchool(), s.ShamirConn)
}

func (s *suite) genTeacherTokens(userIDs []string) ([]string, error) {
	return s.generateExchangeTokens(userIDs, cpb.UserGroup_USER_GROUP_TEACHER.String(), applicantID, s.getSchool(), s.ShamirConn)
}

func (s *suite) genStudentTokens(userIDs []string) ([]string, error) {
	return s.generateExchangeTokens(userIDs, cpb.UserGroup_USER_GROUP_STUDENT.String(), applicantID, s.getSchool(), s.ShamirConn)
}

func (s *suite) generateExchangeToken(userID, userGroup, applicantID string, schoolID int64, conn *grpc.ClientConn) (string, error) {
	firebaseToken, err := generateValidAuthenticationToken(userID, userGroup)
	if err != nil {
		return "", err
	}
	token, err := helper.ExchangeToken(firebaseToken, userID, userGroup, applicantID, schoolID, conn)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *suite) aValidToken(ctx context.Context, arg1 string) (context.Context, error) {
	// sign in as admin
	if !s.CommonSuite.ContextHasToken(ctx) {
		ctx2, err := s.CommonSuite.ASignedInWithSchool(ctx, "school admin", int32ResourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}

	stepState := common.StepStateFromContext(ctx)
	switch arg1 {
	case "student":
		stu, err := s.CommonSuite.CreateStudent(ctx, []string{s.filterSuiteState.defaultLocationID}, nil)
		if err != nil {
			return ctx, err
		}
		s.studentID = stu.UserProfile.UserId
		s.studentToken, err = s.generateExchangeToken(s.studentID, cpb.UserGroup_USER_GROUP_STUDENT.String(), applicantID, constants.ManabieSchool, s.ShamirConn)
		if err != nil {
			return ctx, err
		}
	case "current teacher", "teacher":
		profile, tok, err := s.CommonSuite.CreateTeacher(ctx)
		if err != nil {
			return ctx, fmt.Errorf("CreateTeacher %w", err)
		}
		s.teacherID = profile.StaffId
		s.TeacherToken = tok
	case "teacher with user groups":
		profile, tok, err := s.CommonSuite.CreateTeacherWithUserGroups(ctx, s.userGroupIDs)
		if err != nil {
			return ctx, fmt.Errorf("CreateTeacher %w", err)
		}
		s.teacherID = profile.StaffId
		s.teacherIDs = append(s.teacherIDs, profile.StaffId)
		s.TeacherToken = tok
		if s.TeacherProfileMap == nil {
			s.TeacherProfileMap = make(map[string]*upb.CreateStaffResponse_StaffProfile)
		}

		s.TeacherProfileMap[profile.StaffId] = profile
	case "school admin":
		var err error
		ctx, err = s.CommonSuite.ASignedInWithSchool(ctx, arg1, int32ResourcePathFromCtx(ctx))
		if err != nil {
			return ctx, err
		}
		s.CommonSuite.DefaultLocationID = s.filterSuiteState.defaultLocationID
		s.schoolAdminToken = stepState.AuthToken
		s.schoolAdminID = stepState.CurrentUserID
	case "parent":
		locations := []string{constants.ManabieOrgLocation}
		_, par, err := s.createStudentParentByAPI(ctx, locations)
		if err != nil {
			return ctx, err
		}

		s.parentToken, err = s.CommonSuite.GenerateExchangeTokenCtx(ctx, par.UserProfile.UserId, entity.UserGroupParent)
		if err != nil {
			return ctx, err
		}
		s.parentID = par.UserProfile.UserId
	default:
		return ctx, fmt.Errorf(fmt.Sprintf("unknown token for target %s", arg1))
	}

	s.SchoolIds = append(s.SchoolIds, resourcePathFromCtx(ctx))
	return ctx, nil
}

func (s *suite) returnsStatusCode(ctx context.Context, arg1 string) (context.Context, error) {
	stt, ok := status.FromError(s.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", s.ResponseErr.Error())
	}
	if stt.Code().String() != arg1 {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", arg1, stt.Code().String(), stt.Message())
	}

	return ctx, nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.student_app", "version", "1.0.0")
}

func initStep(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^a invalid "([^"]*)" token$`:                                        s.aInvalidToken,
		`^a valid "([^"]*)" token$`:                                          s.aValidToken,
		`^returns "([^"]*)" status code$`:                                    s.returnsStatusCode,
		`^a invalid conversation_id$`:                                        s.aInvalidConversationID,
		`^a "([^"]*)" send a chat message to conversation$`:                  s.aSendAChatMessageToConversation,
		`^a SendMessageRequest$`:                                             s.aSendMessageRequest,
		`^a user go to chat$`:                                                s.aUserGoToChat,
		`^a valid conversationId$`:                                           s.aValidConversationID,
		`^Tom will close old connection when the same user Subscribe again$`: s.tomWillCloseStreamWhenAUserResubscribe,
		`^a list of messages with types "([^"]*)"$`:                          s.aListOfMessagesWithTypes,

		`^a valid user device token message$`:       s.aValidUserDeviceTokenMessage,
		`^bob send event upsert user device token$`: s.bobSendEventUpsertUserDeviceToken,
		`^tom must record device token message$`:    s.tomMustRecordDeviceTokenMessage,
		`^tom must update the user device token$`:   s.tomMustUpdateTheUserDeviceToken,

		`^Tom should push notification to the "([^"]*)"$`:      s.tomShouldPushNotificationToThe,
		`^user has not seen the message in a duration$`:        s.userHasNotSeenTheMessageInADuration,
		`^a "([^"]*)" device token is existed in DB$`:          s.aDeviceTokenIsExistedInDB,
		`^a "([^"]*)" has not seen the message in a duration$`: s.aHasNotSeenTheMessageInADuration,
		`^Tom should not push notification to the "([^"]*)"$`:  s.tomShouldNotPushNotificationToThe,

		`^a EvtClassRoom with message "([^"]*)"$`:                      s.aEvtClassRoomWithMessage,
		`^bob send event EvtClassRoom$`:                                s.bobSendEventEvtClassRoom,
		`^tom "([^"]*)" store message "([^"]*)" in this conversation$`: s.tomStoreMessageInThisConversation,

		`^a valid "([^"]*)" id in JoinClass$`:                      s.aValidIDInJoinClass,
		`^tom must add above user to this conversation$`:           s.tomMustAddAboveUserToThisConversation,
		`^tom "([^"]*)" remove above user from this conversation$`: s.tomRemoveAboveUserFromThisConversation,

		`^a GetConversationRequest$`:                                                    s.aGetConversationRequest,
		`^a teacher makes GetConversationRequest with "([^"]*)"$`:                       s.aTeacherMakesGetConversationRequestWith,
		`^a user makes GetConversationRequest with an invalid token$`:                   s.aUserMakesGetConversationRequestWithAnInvalidToken,
		`^tom must return conversation with type "([^"]*)" in GetConversationResponse$`: s.tomMustReturnConversationWithTypeInGetConversationResponse,

		`^a GetConversationV2Request$`:                                                    s.aGetConversationV2Request,
		`^a teacher makes GetConversationV2Request with "([^"]*)"$`:                       s.aTeacherMakesGetConversationV2RequestWith,
		`^a user makes GetConversationV2Request with an invalid token$`:                   s.aUserMakesGetConversationV2RequestWithAnInvalidToken,
		`^tom must return conversation with type "([^"]*)" in GetConversationV2Response$`: s.tomMustReturnConversationWithTypeInGetConversationV2Response,

		`^tom must add above teachers to this conversation$`: s.tomMustAddAboveTeachersToThisConversation,

		`^a valid user token$`: s.aValidUserToken,
		`^tom should "([^"]*)" this connection more than (\d+) seconds$`: s.tomShouldThisConnectionMoreThanSeconds,
		`^user send ping event to stream every (\d+) seconds$`:           s.userSendPingEventToStreamEverySeconds,
		`^user subscribe to endpoint streaming event$`:                   s.userSubscribeToEndpointStreamingEvent,

		`^student "([^"]*)" call ConversationList$`: s.studentListConversation,
		`^tom must not return lesson conversation$`: s.tomMustNotReturnLessonConversation,
		`^teacher list conversation$`:               s.teacherListConversation,

		`^user send ping subscribeV2 to stream via ping endpoint every (\d+) seconds$`:  s.userSendPingSubscribeV2ToStreamViaPingEndpointEverySeconds,
		`^user subscribe to endpoint subscribeV2`:                                       s.userSubscribeToEndpointSubscribeV2,
		`^tom should "([^"]*)" this connection of subscribeV2 more than (\d+) seconds$`: s.tomShouldThisConnectionOfSubscribeV2MoreThanSeconds,

		`^tom must send end live lesson message and remove all members from conversation$`: s.tomMustSendEndLiveLessonMessageAndRemoveAllMembersFromConversation,
		`^a EvtUser with message "([^"]*)"$`:                                               s.aEvtUserWithMessage,
		`^student must be in conversation$`:                                                s.studentMustBeInConversation,

		`^yasuo send event EvtUser$`:                                            s.yasuoSendEventEvtUser,
		`^a student conversation with (\d+) teacher$`:                           s.aStudentConversationWithTeacher,
		`^tom must create conversation for parent$`:                             s.tomMustCreateConversationForParent,
		`^all teacher in student conversation must be in parent conversation$`:  s.allTeacherInStudentConversationMustBeInParentConversation,
		`^"([^"]*)" join conversations$`:                                        s.userJoinConversations,
		`^"([^"]*)" must be member of conversations$`:                           s.userMustBeMemberOfConversations,
		`^bob send event upsert user device token with new token and new name$`: s.bobSendEventUpsertUserDeviceTokenWithNewTokenAndNewName,
		`^tom must update conversation correctly$`:                              s.tomMustUpdateConversationCorrectly,
		`^teacher joins all conversations$`:                                     s.teacherJoinAllConversations,
		`^"([^"]*)" joins all conversations$`:                                   s.userJoinAllConversations,
		`^teacher must be member of all conversations with specific schools$`:   s.teacherMustBeMemberOfAllConversationsWithSpecifySchools,
		`^"([^"]*)" must be member of all conversations with specific schools$`: s.userMustBeMemberOfAllConversationsWithSpecifySchools,
		`^system must send "([^"]*)" conversation message$`:                     s.systemMustSendConversationMessage,
		`^student conversation is created$`:                                     s.createStudentConversation,
		`^return ConversationList must have "([^"]*)"$`:                         s.returnConversationListMustHave,
		`^student send (\d+) message to teacher$`:                               s.studentSendMessageToTeacher,
		`^teacher joined some conversation in school$`:                          s.teacherJoinedSomeConversationInSchool,
		`^Tom must returns (\d+) total unread message$`:                         s.tomMustReturnsTotalUnreadMessage,
		`^teacher read all messages$`:                                           s.teacherReadAllMessages,
		`^random new conversations created$`:                                    s.randomNewConversationsCreated,

		`^current parents receive message "([^"]*)" with content "([^"]*)"$`: s.currentParentsReceiveMessageWithContent,
		`^teachers receive message "([^"]*)" with content "([^"]*)"$`:        s.teachersReceiveMessageWithContent,

		`^system must send only "([^"]*)" message which unjoined conversations before$$`: s.systemMustOnlySendConversationMessageWhichUnjoinedBefore,
		`^the teacher joins some conversations$`:                                         s.theTeacherJoinSomeConversations,
		`^a signed as a teacher$`:                                                        s.aSignedAsATeacher,
		`^student seen conversation$`:                                                    s.studentSeenConversation,
		`^teacher send message to conversation$`:                                         s.teacherSendMessageToConversation,
		`^tom must mark messages in conversation as read for student$`:                   s.tomMustMarkMessagesInConversationAsReadForStudent,
		`^"([^"]*)" get total unread message$`:                                           s.getTotalUnreadMessage,
		`^all member subscribe this conversation to chat and do not miss any message$`:   s.allMemberSubscribeToThisConversationAndChat,
		`^create a valid student conversation in db with a teacher and a student$`:       s.createAValidStudentConversationInDBWithATeacherAndAStudent,
		`^teacher does not receives notification$`:                                       s.teacherDoesNotReceivesNotification,
		`^teachers device tokens is existed in DB$`:                                      s.teachersDeviceTokensIsExistedInDB,

		`^a teacher who joined all conversations$`:                                   s.aTeacherWhoJoinedAllConversations,
		`^teacher leaves some conversations$`:                                        s.teacherLeaveSomeConversations,
		`^teacher must not be member of conversations recently left$`:                s.teacherMustNotBeMemberOfConversationsRecentlyLeft,
		`^teacher leaves student chat$`:                                              s.teacherLeavesStudentChat,
		`^teacher number (\d+) leaves student chat$`:                                 s.teacherNumberLeavesStudentChat,
		`^teacher rejoins student chat$`:                                             s.teacherRejoinsStudentChat,
		`^teacher who left chat does not receive sent message$`:                      s.teacherWhoLeftChatDoesNotReceiveSentMessage,
		`^teacher who left chat sends a message$`:                                    s.teacherWhoLeftChatSendsAMessage,
		`^teacher who left chat cannot send message$`:                                s.teacherWhoLeftChatCannotSendMessage,
		`^other teachers receive leave conversation system message$`:                 s.otherTeachersReceiveLeaveConversationSystemMessage,
		`^student receive leave conversation system message$`:                        s.studentReceiveLeaveConversationSystemMessage,
		`^teacher who left conversation receives leave conversation system message$`: s.teacherWhoLeftConversationReceivesLeaveConversationSystemMessage,
		`^the invalid chat does not record teacher membership$`:                      s.theInvalidChatDoesNotRecordTeacherMembership,
		`^teacher leaves student chat and "([^"]*)" chat he does not join$`:          s.teacherLeavesStudentChatAndChatHeDoesNotJoin,
		`^the conversation that teacher left is "([^"]*)" in conversation list$`:     s.theConversationThatTeacherLeftIsInConversationList,

		`^GetConversationResponse has "([^"]*)" user with role "([^"]*)" status "([^"]*)"$`: s.getConversationResponseHasUserWithRoleStatus,
		`^GetConversationResponse has latestMessage with content "([^"]*)"$`:                s.getConversationResponseHasLatestMessageWithContent,

		`^GetConversationV2Response has "([^"]*)" user with role "([^"]*)" status "([^"]*)"$`: s.getConversationV2ResponseHasUserWithRoleStatus,
		`^GetConversationV2Response has latestMessage with content "([^"]*)"$`:                s.getConversationV2ResponseHasLatestMessageWithContent,

		`^client calling ConversationDetail$`:                                                            s.clientCallingConversationDetail,
		`^response does not include system message$`:                                                     s.responseDoesNotIncludeSystemMessage,
		`^all connections receive "([^"]*)" msg in order$`:                                               s.allConnectionsReceiveMsgInOrder,
		`^spamming "([^"]*)" into conversation$`:                                                         s.spammingIntoConversation,
		`^a user subscribes stream using "([^"]*)" grpc connections and "([^"]*)" grpc web connections$`: s.aUserMakeSubscribesStreamUsingGrpcConnectionsAndGrpcWebConnections,
		`^all connections are routed to one node$`:                                                       s.allConnectionsAreRoutedToOneNode,
		`^grpc metadata with key "([^"]*)" in context$`:                                                  s.grpcMetadataWithKeyInContext,
		`^all connections are routed to multiple nodes$`:                                                 s.allConnectionsAreRoutedToMultipleNodes,
		// `^students parents chats are created$`:                                                           s.studentsParentsChatsAreCreated,
		`^resource path of school "([^"]*)" is applied$`: s.resourcePathOfSchoolIsApplied,
		// `^tom must update conversation location correctly$`:                                              s.tomMustUpdateConversationLocationCorrectly,
		`^tom must update conversation location correctly for event "([^"]*)"$`:             s.tomMustUpdateConversationLocationCorrectlyForEvent,
		`^usermgmt send event "([^"]*)" with new token and "([^"]*)" location in db$`:       s.usermgmtSendEventWithNewTokenAndLocationInDB,
		`^chats "([^"]*)" each has new message from student or parent$`:                     s.chatsEachHasNewMessageFromStudentOrParent,
		`^Tom must returns "([^"]*)" total unread message in locations "([^"]*)"$`:          s.tomMustReturnsTotalUnreadMessageInLocations,
		`^"([^"]*)" delete "([^"]*)" message$`:                                              s.aDeleteMessage,
		`^"([^"]*)" see deleted message in conversation$`:                                   s.aSeeDeletedMessageInConversation,
		`^a student sends "([^"]*)" item with content "([^"]*)"$`:                           s.aStudentSendsItemWithContent,
		`^migrate conversation locations$`:                                                  s.migrateConversationLocations,
		`^create student with no locations$`:                                                s.createStudentConversationWithNoLocation,
		`^insert org location access path to student$`:                                      s.insertStudentAccessPaths,
		`^conversation location is inserted`:                                                s.conversationLocationInserted,
		`^"([^"]*)" create new live lesson private conversation with "([^"]*)"`:             s.createLiveLessonPrivateConversation,
		`^"([^"]*)" see the live lesson private conversation`:                               s.seeLiveLessonConversation,
		`^"([^"]*)" get the lesson private conversation detail`:                             s.getTheLiveLessonPrivateConversationDetail,
		`^"([^"]*)" sees the lesson private conversation returned with the correct data`:    s.seeLiveLessonConversationDetail,
		`^a user group with "([^"]*)" role and "([^"]*)" location type$`:                    s.createUserGroupWithRoleNamesAndLocations,
		`^a signed in as teacher with user groups$`:                                         s.aSignedAsATeacherWithUserGroups,
		`^a chat between a student and "([^"]*)" teachers with user groups$`:                s.aChatBetweenAStudentAndTeachersWithUserGroups,
		`^a chat between a student with locations and "([^"]*)" teachers with user groups$`: s.aChatBetweenAStudentWithLocationsAndTeachersWithUserGroups,
		`^teachers are deactivated in conversation members$`:                                s.checkTeachersDeactivated,
		`^update "([^"]*)" teacher user groups$`:                                            s.updateTeacherUserGroups,
		`^update user group with "([^"]*)" role and "([^"]*)" location type$`:               s.updateUserGroupWithRoleNamesAndLocations,
	}

	initStudentParentTeacherSuite(steps, ctx, s)
	initFilterChatGroupSuite(steps, ctx, s)
	initLiveLessonChatSuite(steps, ctx, s)
	initJoinAllStepSuite(steps, ctx, s)

	// nolint
	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMap(steps)
	})

	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

func ContextWithJWTClaims(ctx context.Context) context.Context {
	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
			DefaultRole:  bob_entities.UserGroupAdmin,
			UserGroup:    bob_entities.UserGroupAdmin,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)
	return ctx
}

// called by other features file to register bdd to main step maps, checking if duplicate steps are registered
func applyMergedSteps(ctx *godog.ScenarioContext, m map[string]interface{}, m2 map[string]interface{}) {
	for regexstring := range m2 {
		if _, exist := m[regexstring]; exist {
			panic(fmt.Sprintf("register duplicate step: %s", regexstring))
		}
	}

	for k, v := range m2 {
		ctx.Step(k, v)
	}
}

var (
	buildRegexpMapOnce     sync.Once
	regexpMap              map[string]*regexp.Regexp
	insertDefaultLocations = `
INSERT INTO public.locations
(location_id,access_path, name, location_type, partner_internal_id, partner_internal_parent_id, parent_location_id, resource_path, updated_at, created_at, is_archived)
VALUES	('01FR4M51XJY9E77GSN4QZ1Q9N1','01FR4M51XJY9E77GSN4QZ1Q9N1', 'Manabie','01FR4M51XJY9E77GSN4QZ1Q9M1','1', NULL, NULL, '-2147483648', now(), now(),false),
		('01FR4M51XJY9E77GSN4QZ1Q9N2','01FR4M51XJY9E77GSN4QZ1Q9N2', 'JPREP','01FR4M51XJY9E77GSN4QZ1Q9M2','1', NULL, NULL, '-2147483647', now(), now(),false) ON CONFLICT  ON CONSTRAINT locations_pkey DO NOTHING`
)
