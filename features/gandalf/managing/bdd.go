package managing

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/eureka"
	"github.com/manabie-com/backend/features/fatima"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/features/tom"
	"github.com/manabie-com/backend/features/yasuo"
	"github.com/manabie-com/backend/internal/bob/entities"
	bobEnt "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	tomPb "github.com/manabie-com/backend/pkg/genproto/tom"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	natsgo "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	bobConn           *grpc.ClientConn
	tomConn           *grpc.ClientConn
	yasuoConn         *grpc.ClientConn
	eurekaConn        *grpc.ClientConn
	fatimaConn        *grpc.ClientConn
	shamirConn        *grpc.ClientConn
	userMgmtConn      *grpc.ClientConn
	entryExitMgmtConn *grpc.ClientConn
	bobDB             *pgxpool.Pool
	tomDB             *pgxpool.Pool
	eurekaDB          *pgxpool.Pool
	fatimaDB          *pgxpool.Pool
	zeusDB            *pgxpool.Pool
	firebaseAddr      string
	bobDBTrace        *database.DBTrace
	offStdErrScenario []string
	jetStreamAddress  string
	jsm               nats.JetStreamManagement
	firebaseClient    *auth.Client
	zapLogger         *zap.Logger
	applicantID       string
)

func init() {
	rand.Seed(time.Now().UnixNano())
	offStdErrScenario = append(offStdErrScenario, "publish a message with subject that he does not have permission (and have permission) to publish message")
	offStdErrScenario = append(offStdErrScenario, "subscribe a message with subject that he does not have permission (and have permission) to subscribe")
}

// TestSuiteInitializer ...
func TestSuiteInitializer(c *gandalf.Config, fakeFirebaseAddr string) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, fakeFirebaseAddr)
		})

		ctx.AfterSuite(func() {
			bobConn.Close()
			bobDB.Close()
			tomConn.Close()
			tomDB.Close()
			yasuoConn.Close()
			fatimaConn.Close()
			fatimaDB.Close()
			eurekaConn.Close()
			eurekaDB.Close()
			shamirConn.Close()
			entryExitMgmtConn.Close()
			zeusDB.Close()
			jsm.Close()
		})
	}
}

type stateKeyForGandalf struct{}

func GandalfStepStateFromContext(ctx context.Context) *StepState {
	state := ctx.Value(stateKeyForGandalf{})
	if state == nil {
		return &StepState{}
	}
	return state.(*StepState)
}

func GandalfStepStateToContext(ctx context.Context, state *StepState) context.Context {
	return context.WithValue(ctx, stateKeyForGandalf{}, state)
}

// ScenarioInitializer ...
func ScenarioInitializer(c *gandalf.Config) func(ctx *godog.ScenarioContext) {
	backUpStdErr := os.Stderr
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			ctx = GandalfStepStateToContext(ctx, &s.StepState)
			ctx, _ = context.WithTimeout(ctx, time.Second*40) //nolint:lostcancel
			if golibs.InArrayString(sc.Name, offStdErrScenario) {
				os.Stderr = nil
			}

			ctx, _ = yasuo.InitYasuoState(ctx)
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

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			if golibs.InArrayString(sc.Name, offStdErrScenario) {
				os.Stderr = backUpStdErr
			}
			return ctx, nil
		})
	}
}

func setup(c *gandalf.Config, fakeFirebaseAddr string) {
	firebaseAddr = fakeFirebaseAddr
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	err := c.ConnectGRPCInsecure(ctx, &bobConn, &tomConn, &yasuoConn, &eurekaConn, &fatimaConn, &shamirConn, &userMgmtConn, &entryExitMgmtConn)

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to run BDD setup: %s", err))
	}

	c.ConnectDB(ctx, &bobDB, &tomDB, &eurekaDB, &fatimaDB, &zeusDB)
	db, _, _ := database.NewPool(ctx, zap.NewNop(), c.PostgresV2.Databases["bob"])
	bobDBTrace = &database.DBTrace{
		DB: db,
	}
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot connect to firebase: %v", err))
	}
	firebaseClient, err = app.Auth(ctx)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot create firebase client: %v", err))
	}

	jetStreamAddress = c.NatsJS.Address
	jsm, err = nats.NewJetStreamManagement(c.NatsJS.Address, "Gandalf", "m@n@bi3",
		c.NatsJS.MaxReconnects, c.NatsJS.ReconnectWait, c.NatsJS.IsLocal, zapLogger)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create jetstream management: %v", err))
	}
	jsm.ConnectToJS()

	applicantID = c.JWTApplicant

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
	_, err = bobDBTrace.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

type suite struct {
	connections
	StepState
	bobSuite    *bob.Suite
	yasuoSuite  *yasuo.Suite
	fatimaSuite *fatima.Suite
	eurekaSuite *eureka.Suite
	tomSuite    *tom.Suite
	Cfg         *gandalf.Config
	ZapLogger   *zap.Logger
	ApplicantID string
}

type connections struct {
	bobConn    *grpc.ClientConn
	tomConn    *grpc.ClientConn
	yasuoConn  *grpc.ClientConn
	eurekaConn *grpc.ClientConn
	fatimaConn *grpc.ClientConn
	shamirConn *grpc.ClientConn
	bobDB      *pgxpool.Pool
	tomDB      *pgxpool.Pool
	eurekaDB   *pgxpool.Pool
	fatimaDB   *pgxpool.Pool
	bobDBTrace *database.DBTrace
	zeusDB     *pgxpool.Pool
	jsm        nats.JetStreamManagement
}

type StepState struct {
	GandalfStateAuthToken        string
	GandalfStateCurrentUserID    string
	GandalfStateCurrentClassID   int32
	GandalfStateConversationID   string
	GandalfStateSchool           *bobEnt.School
	GandalfStateClass            *bobEnt.Class
	GandalfStateTeacherIDsMap    map[string]string // map every teacher id with teacher token
	GandalfStateSubV2Clients     map[string]tomPb.ChatService_SubscribeV2Client
	GandalfStateRequest          interface{}
	GandalfStateResponse         interface{}
	GandalfStateResponseErr      error
	GandalfStateRequestSentAt    time.Time
	GandalfStateJetStreamAddress string
	GandalfStateUserIDs          []string

	TomStepState
	fatimaStepState
	EurekaStepState EurekaStepState
	ZeusStepState   ZeusStepState
	BobStepState    BobStepState
	YasuoStepState  YasuoStepState
}

func newSuite(c *gandalf.Config) *suite {
	s := &suite{
		connections: connections{
			bobConn,
			tomConn,
			yasuoConn,
			eurekaConn,
			fatimaConn,
			shamirConn,
			bobDB,
			tomDB,
			eurekaDB,
			fatimaDB,
			bobDBTrace,
			zeusDB,
			jsm,
		},
		StepState: StepState{
			GandalfStateJetStreamAddress: jetStreamAddress,
			ZeusStepState: ZeusStepState{
				MapJSContext:       make(map[string]natsgo.JetStreamContext),
				MapPublishStatus:   make(map[string]error),
				MapSubscribeStatus: make(map[string]error),
			},
		},
		Cfg:         c,
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
	}

	s.newBobSuite()
	s.newYasuoSuite()
	s.newTomSuite()
	bob.SetFirebaseAddr(firebaseAddr)
	yasuo.SetFirebaseAddr(firebaseAddr)
	tom.SetFirebaseAddr(firebaseAddr)

	s.newFatimaSuite(firebaseAddr)
	s.newEurekaSuite(firebaseAddr)

	return s
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	bobSteps := initStepForBobServiceFeature(s)
	tomSteps := initStepForTomServiceFeature(s)
	yasuoSteps := initStepForYasuoServiceFeature(s)
	fatimaSteps := initStepForFatimaServiceFeature(s)
	zeusSteps := initStepForZeus(s)
	eurekaSteps := initStepForEurekaServiceFeature(s)

	steps := make(map[string]interface{})
	appendSteps(bobSteps, steps)
	appendSteps(tomSteps, steps)
	appendSteps(yasuoSteps, steps)
	appendSteps(fatimaSteps, steps)
	appendSteps(zeusSteps, steps)
	appendSteps(eurekaSteps, steps)

	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}

func appendSteps(src, dest map[string]interface{}) {
	for k, v := range src {
		dest[k] = v
	}
}
