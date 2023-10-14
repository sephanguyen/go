package usermgmt

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	"github.com/manabie-com/backend/j4/serviceutil/usermgmt"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
	j4 "github.com/manabie-com/j4/pkg/runner"
	"google.golang.org/grpc/metadata"
)

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	runConfig, err := c.GetScenarioConfig("import_student")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)
	bobDB := dep.Connections.DBConnPools["bob"]
	tokenGen := serviceutil.NewTokenGenerator(c, dep.Connections)

	runCfg.TestFunc = func(ctx context.Context) error {
		conn, err := dep.PoolToGateWay.Get(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()
		tok, err := tokenGen.GetTokenFromShamir(ctx, c.AdminID, c.SchoolID)
		if err != nil {
			return err
		}

		userMgmtCl := upb.NewStudentServiceClient(conn)
		_, err = usermgmt.CreateStudents(contextWithToken(ctx, tok), 10, c.SchoolID, bobDB, userMgmtCl)
		return err
	}
	importStuScene, err := j4.NewScenario("import_student", *runCfg)
	if err != nil {
		return nil, err
	}
	hasuraScenarios, err := serviceutil.GenHasuraScenarios(ctx, c, dep, "bob", hasuraQueries)
	if err != nil {
		return nil, err
	}
	scenarios := append(hasuraScenarios, importStuScene)
	return scenarios, nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}
