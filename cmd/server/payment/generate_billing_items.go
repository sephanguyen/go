package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	service "github.com/manabie-com/backend/internal/payment/services/internal_service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type generateBillingItemsConfig struct {
	Common       configs.CommonConfig
	PostgresV2   configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS       configs.NatsJetStreamConfig
	KafkaCluster configs.KafkaClusterConfig `yaml:"kafka_cluster"`
}

func init() {
	bootstrap.RegisterJob("payment_generate_billing_items", generateBillingItems).
		Desc("Generate billing items")
}

func generateBillingItems(ctx context.Context, config generateBillingItemsConfig, rsc *bootstrap.Resources) error {
	zlogger := rsc.Logger().Sugar()
	zlogger.Info("start process GenerateBillingItems")

	paymentDB := rsc.DBWith("fatima")

	orgQuery := "SELECT organization_id, name FROM organizations"
	organizations, err := paymentDB.Query(ctx, orgQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer organizations.Close()
	resp := &pb.GenerateBillingItemsResponse{}
	for organizations.Next() {
		var organizationID, name string

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zlogger.Error("failed to scan an orgs row: %s", err)
			continue
		}
		if _, ok := MapOrgIDUserID[organizationID]; ok {
			// for db RLS query
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
					ResourcePath: organizationID,
					UserID:       MapOrgIDUserID[organizationID],
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)
			paymentDBTrace := &database.DBTrace{
				DB: paymentDB,
			}
			orderService := service.NewInternalService(paymentDBTrace, rsc.NATS(), rsc.Kafka(), config.Common)
			req := &pb.GenerateBillingItemsRequest{}
			resp, err = orderService.GenerateBillingItems(ctx, req)
			if err != nil {
				return fmt.Errorf("orderService.GenerateBillingItems: %s", err)
			}
		}
	}

	zlogger.Info("End scheduled process GenerateBillingItems: ", time.Now())
	zlogger.Info("Successful case ", resp.Successed, "/ Failed case ", resp.Failed)
	zlogger.Info("Process done")
	return nil
}
