package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/repositories"
	service "github.com/manabie-com/backend/internal/payment/services/internal_service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
)

type sysBillingStatusConfig struct {
	Common       configs.CommonConfig
	PostgresV2   configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS       configs.NatsJetStreamConfig
	KafkaCluster configs.KafkaClusterConfig `yaml:"kafka_cluster"`
}

func init() {
	bootstrap.RegisterJob("payment_update_billing_status", runSysBillingStatus).
		Desc("Start a cronjob update billing status pending to billed").
		DescLong("update billing status from pending to billed when due")
}

var MapOrgIDUserID = map[string]string{
	"-2147483622": "01GTZYX224982Z1X4MHZQW6BO2",
	"-2147483623": "01GTZYX224982Z1X4MHZQW6BO3",
	"-2147483624": "01GTZYX224982Z1X4MHZQW6BO4",
	"-2147483625": "01GTZYX224982Z1X4MHZQW6BO5",
	"-2147483626": "01GTZYX224982Z1X4MHZQW6BO6",
	"-2147483627": "01GTZYX224982Z1X4MHZQW6DO1",
	"-2147483628": "01GTZYX224982Z1X4MHZQW6DO2",
	"-2147483629": "01GTZYX224982Z1X4MHZQW6DP1",
	"-2147483630": "01GTZYX224982Z1X4MHZQW6DP2",
	"-2147483631": "01GTZYX224982Z1X4MHZQW6DP3",
	"-2147483632": "01GTZYX224982Z1X4MHZQW6DP4",
	"-2147483633": "01GTZYX224982Z1X4MHZQW6DP5",
	"-2147483634": "01GTZYX224982Z1X4MHZQW6DP6",
	"-2147483635": "01GTZYX224982Z1X4MHZQW6DP7",
	"-2147483637": "01GTZYX224982Z1X4MHZQW6DP8",
	"-2147483638": "01GTZYX224982Z1X4MHZQW6DP9",
	"-2147483639": "01GTZYX224982Z1X4MHZQW6DQ1",
	"-2147483640": "01GTZYX224982Z1X4MHZQW6DQ2",
	"-2147483641": "01GTZYX224982Z1X4MHZQW6DQ3",
	"-2147483642": "01GTZYX224982Z1X4MHZQW6DQ4",
	"-2147483643": "01GTZYX224982Z1X4MHZQW6DQ5",
	"-2147483644": "01GTZYX224982Z1X4MHZQW6DQ6",
	"-2147483645": "01GTZYX224982Z1X4MHZQW6DQ7",
	"-2147483646": "01GTZYX224982Z1X4MHZQW6DQ8",
	"-2147483647": "01GTZYX224982Z1X4MHZQW6DQ9",
	"-2147483648": "01GTZYX224982Z1X4MHZQW6DR1",
}

func runSysBillingStatus(ctx context.Context, config sysBillingStatusConfig, rsc *bootstrap.Resources) error {
	zlogger := rsc.Logger().Sugar()
	zlogger.Info("start process UpdateBillingStatus")

	paymentDB := rsc.DBWith("fatima")

	orgQuery := "SELECT organization_id, name FROM organizations" //nolint:goconst
	organizations, err := paymentDB.Query(ctx, orgQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer organizations.Close()

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
			err = updateBillingStatusFromPendingToBilledWhenDue(ctx, orderService)
			if err != nil {
				return fmt.Errorf("failed to update billing status: %s", err)
			}
		}
	}

	zlogger.Info("End scheduled process UpdateBillingStatus: %v", time.Now())
	zlogger.Info("Process done")
	return nil
}

func updateBillingStatusFromPendingToBilledWhenDue(ctx context.Context, s *service.InternalService) error {
	r := &repositories.BillItemRepo{}

	billingItems, err := r.GetBillingItemsThatNeedToBeBilled(ctx, s.DB)
	if len(billingItems) < 1 {
		return err
	}

	items := make([]*pb.UpdateBillItemStatusRequest_UpdateBillItem, 0, len(billingItems))

	for _, billItem := range billingItems {
		items = append(items, &pb.UpdateBillItemStatusRequest_UpdateBillItem{
			BillItemSequenceNumber: billItem.BillItemSequenceNumber.Int,
			BillingStatusTo:        pb.BillingStatus_BILLING_STATUS_BILLED,
		})
	}

	itemsUpdate := &pb.UpdateBillItemStatusRequest{
		UpdateBillItems: items,
	}
	_, err = s.UpdateBillItemStatus(ctx, itemsUpdate)
	if err != nil {
		return fmt.Errorf("s.UpdateBillItemStatus: %s", err)
	}
	return nil
}
