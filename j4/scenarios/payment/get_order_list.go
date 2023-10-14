package payment

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	"github.com/manabie-com/backend/j4/serviceutil/payment"
	paymentpb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	j4 "github.com/manabie-com/j4/pkg/runner"
	"google.golang.org/grpc/metadata"
)

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	runConfig, err := c.GetScenarioConfig("Payment_GetOrderList")
	if err != nil {
		return nil, err
	}
	runCfg := infras.MustOptionFromConfig(&runConfig)
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

		paymentClient := paymentpb.NewOrderServiceClient(conn)
		_, err = payment.GetOrderList(contextWithToken(ctx, tok), paymentClient)
		return err
	}
	scenario, err := j4.NewScenario("Payment_GetOrderList", *runCfg)
	if err != nil {
		return nil, err
	}
	return []*j4.Scenario{scenario}, nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}
