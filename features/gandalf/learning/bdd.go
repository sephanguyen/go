package learning

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/gandalf"
	"github.com/manabie-com/backend/internal/bob/entities"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	tomEntities "github.com/manabie-com/backend/internal/tom/domain/core"
	tomPb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
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
	zapLogger         *zap.Logger
	jsm               nats.JetStreamManagement
	applicantID       string
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// TestSuiteInitializer ...
func TestSuiteInitializer(c *gandalf.Config, fakeFirebaseAddr string) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c, fakeFirebaseAddr)
		})

		ctx.AfterSuite(func() {
			bobDB.Close()
			tomDB.Close()
			zeusDB.Close()
			bobConn.Close()
			tomConn.Close()
			yasuoConn.Close()
			entryExitMgmtConn.Close()
			jsm.Close()
		})
	}
}

// ScenarioInitializer ...
func ScenarioInitializer(c *gandalf.Config) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite()
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			bobCtx := bob.InitBobState(ctx)
			claim := interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					ResourcePath: "1",
					DefaultRole:  entities.UserGroupAdmin,
					UserGroup:    entities.UserGroupAdmin,
				},
			}
			bobCtx = interceptors.ContextWithJWTClaims(bobCtx, &claim)
			return bobCtx, nil
		})
	}
}

func setup(c *gandalf.Config, fakeFirebaseAddr string) {
	firebaseAddr = fakeFirebaseAddr

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)
	err := c.ConnectGRPCInsecure(ctx, &bobConn, &tomConn, &yasuoConn, &eurekaConn, &fatimaConn, &shamirConn, &userMgmtConn, &entryExitMgmtConn)
	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to run BDD setup: %s", err))
	}

	c.ConnectDB(ctx, &bobDB, &tomDB, &eurekaDB, &fatimaDB, &zeusDB)

	if err != nil {
		zapLogger.Panic(fmt.Sprintf("failed to create bus factory: %s", err))
	}

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
	_, err = bobDB.Exec(ctx, stmt)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
}

type suite struct {
	connections
	stepState
	bobSuite    *bob.Suite
	ZapLogger   *zap.Logger
	ApplicantID string
}

type connections struct {
	bobConn    *grpc.ClientConn
	tomConn    *grpc.ClientConn
	yasuoConn  *grpc.ClientConn
	shamirConn *grpc.ClientConn
	bobDB      *pgxpool.Pool
	tomDB      *pgxpool.Pool
	jsm        nats.JetStreamManagement
}

type stepState struct {
	User                         *bobEntities.User
	AdminToken, MainTeacherToken string
	CurrentUserID, MainTeacherID string
	CurrentClassID               int32
	ConversationID               string
	SchoolIDs                    []int32
	Class                        *bobEntities.Class
	Conversations                []*tomPb.Conversation
	TeacherTokens                []string
	TeacherIds                   []string
	ConfigName                   *bobEntities.Config
	ConfigPeriod                 *bobEntities.Config
	Configs                      []*bobEntities.Config
	ConversationMembers          map[string][]tomEntities.ConversationMembers
	MessageResponse              *tomPb.SendMessageResponse
	PlanPeriod                   string
	SubV2Clients                 map[string]tomPb.ChatService_SubscribeV2Client
	Request                      interface{}
	Response                     interface{}
	ResponseErr                  error
	RequestSentAt                time.Time

	BobStepState
}

func newSuite() *suite {
	s := &suite{
		connections: connections{
			bobConn,
			tomConn,
			yasuoConn,
			shamirConn,
			bobDB,
			tomDB,
			jsm,
		},
		stepState:   stepState{},
		ZapLogger:   zapLogger,
		ApplicantID: applicantID,
	}
	s.newBobSuite()
	bob.SetFirebaseAddr(firebaseAddr)

	return s
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	bobSteps := initStepForBobServiceFeature(s)
	tomSteps := initStepForTomServiceFeature(s)
	steps := map[string]interface{}{}

	appendSteps(bobSteps, steps)
	appendSteps(tomSteps, steps)
	for pattern, stepFunc := range steps {
		ctx.Step(pattern, stepFunc)
	}
}

func appendSteps(src, dest map[string]interface{}) {
	for k, v := range src {
		if _, ok := dest[k]; ok {
			panic(fmt.Sprintf(`key "%s" is duplicated in map`, k))
		}
		dest[k] = v
	}
}
