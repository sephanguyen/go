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

	"google.golang.org/protobuf/types/known/timestamppb"
)

type updateStudentProductStatusConfig struct {
	Common       configs.CommonConfig
	PostgresV2   configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS       configs.NatsJetStreamConfig
	KafkaCluster configs.KafkaClusterConfig `yaml:"kafka_cluster"`
}

func init() {
	bootstrap.RegisterJob("payment_update_student_product_status", updateStudentProductStatus).
		Desc("Update student product status")
}

func updateStudentProductStatus(ctx context.Context, config updateStudentProductStatusConfig, rsc *bootstrap.Resources) error {
	zlogger := rsc.Logger().Sugar()
	zlogger.Info("start scheduled process UpdateStudentProductStatus: %v", time.Now())

	paymentDB := rsc.DBWith("fatima")
	orgQuery := "SELECT organization_id, name FROM organizations"
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
			req := &pb.UpdateStudentProductStatusRequest{
				StudentProductLabel: []string{
					pb.StudentProductLabel_WITHDRAWAL_SCHEDULED.String(),
					pb.StudentProductLabel_GRADUATION_SCHEDULED.String(),
					pb.StudentProductLabel_PAUSE_SCHEDULED.String(),
				},
				EffectiveDate: &timestamppb.Timestamp{Seconds: time.Now().Unix()},
			}
			_, err := orderService.UpdateStudentProductStatus(ctx, req)
			if err != nil {
				return fmt.Errorf("orderService.UpdateStudentProductStatus: %s", err)
			}
		}
	}

	zlogger.Info("End scheduled process UpdateStudentProductStatus: %v", time.Now())
	zlogger.Info("Process done")
	return nil
}
