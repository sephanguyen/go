package draft

import (
	"context"
	"math/rand"
	"regexp"
	"sync"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/cucumber/godog"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func init() {
	common.RegisterTest("draft", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    TestSuiteInitializer,
		ScenarioInitFunc: ScenarioInitializer,
	})
}

var (
	conn *grpc.ClientConn
	db   *pgxpool.Pool
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
	zapLogger          *zap.Logger
)

// type StepState struct {
// 	Request     interface{}
// 	Response    interface{}
// 	ResponseErr error
// }

type suite struct {
	Cfg       *common.Config
	DB        database.Ext
	Conn      *grpc.ClientConn
	ZapLogger *zap.Logger
	// *StepState
	mergeBlockSuite
}

func setup(c *common.Config) {
	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	var err error
	conn = rsc.GRPCDial("draft")

	zapLogger = logger.NewZapLogger(c.Common.Log.ApplicationLevel, true)

	var dbcancel func() error
	db, dbcancel, err = database.NewPool(context.Background(), zapLogger, c.PostgresV2.Databases["draft"])
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := dbcancel(); err != nil {
			zapLogger.Error("dbcancel() failed", zap.Error(err))
		}
	}()
}

func TestSuiteInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {
		ctx.BeforeSuite(func() {
			setup(c)
		})

		ctx.AfterSuite(func() {
			conn.Close()
		})
	}
}

func ScenarioInitializer(c *common.Config, f common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	rand.Seed(time.Now().Unix())
	return func(ctx *godog.ScenarioContext) {
		s := &suite{
			Cfg: c,
		}
		s.Conn = conn
		s.DB = db
		s.ZapLogger = zapLogger

		initStep(ctx, s)
	}
}

func initStep(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^a "([^"]*)" flag$`:                                    s.aFlag,
		`^a git branch with no recorded coverage$`:              s.aGitBranchWithNoRecordedCoverage,
		`^a git branch with recorded coverage$`:                 s.aGitBranchWithRecordedCoverage,
		`^client calls CreateTargetCoverage$`:                   s.clientCallsCreateTargetCoverage,
		`^client calls UpdateTargetCoverage$`:                   s.clientCallsUpdateTargetCoverage,
		`^client calls TestCoverage with a "([^"]*)" coverage$`: s.clientCallsTestCoverageWithACoverage,
		`^coverage amount is recorded in the database$`:         s.coverageAmountIsRecordedInTheDatabase,
		`^coverage is updated in the database$`:                 s.coverageIsUpdatedInTheDatabase,
		`^created coverage$`:                                    s.createdCoverage,
		`^servers returns a "([^"]*)" result to the client$`:    s.serversReturnsAResultToTheClient,

		`^repo has merge status is "([^"]*)"$`:   s.repoHasMergeStatusIs,
		`^a repo "([^"]*)" of owner "([^"]*)"$`:  s.aRepoOfOwner,
		`^workflow "([^"]*)" is called to repo$`: s.workflowIsCalledToRepo,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMap(steps)
	})

	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}

func (s *suite) aFlag(arg1 string) error {
	return nil
}

func (s *suite) aGitBranchWithNoRecordedCoverage() error {
	return nil
}

func (s *suite) aGitBranchWithRecordedCoverage() error {
	return nil
}

func (s *suite) clientCallsCreateTargetCoverage() error {
	return s.createCoverageTest()
}

func (s *suite) clientCallsTestCoverageWithACoverage(arg1 string) error {
	return s.compareCoverageTest()
}

func (s *suite) clientCallsUpdateTargetCoverage() error {
	return s.updateCoverageTest()
}

func (s *suite) coverageAmountIsRecordedInTheDatabase() error {
	return nil
}

func (s *suite) coverageIsUpdatedInTheDatabase() error {
	return nil
}

func (s *suite) createdCoverage() error {
	return nil
}

func (s *suite) serversReturnsAResultToTheClient(arg1 string) error {
	return nil
}
