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

type updateStudentPackageByStudentPackageOrderConfig struct {
	Common       configs.CommonConfig
	PostgresV2   configs.PostgresConfigV2 `yaml:"postgres_v2"`
	NatsJS       configs.NatsJetStreamConfig
	KafkaCluster configs.KafkaClusterConfig `yaml:"kafka_cluster"`
}

func init() {
	bootstrap.RegisterJob("payment_update_student_package", updateStudentPackageByStudentPackageOrder).
		Desc("Update student package by student package order when effective date arrives")
}

func updateStudentPackageByStudentPackageOrder(ctx context.Context, config updateStudentPackageByStudentPackageOrderConfig, rsc *bootstrap.Resources) error {
	zlogger := rsc.Logger().Sugar()
	zlogger.Info("---- Job updateStudentPackageByStudentPackageOrder starts ----")

	paymentDB := rsc.DBWith("fatima")
	orgQuery := "SELECT organization_id, name FROM organizations"
	organizations, err := paymentDB.Query(ctx, orgQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %s", err)
	}
	defer organizations.Close()
	resp := &pb.UpdateStudentPackageForCronjobResponse{}
	for organizations.Next() {
		var (
			organizationID, name, errDetail string
			now                             = time.Now()
		)

		if err := organizations.Scan(&organizationID, &name); err != nil {
			zlogger.Error("Cronjob payment_update_student_package: Failed to scan an organization_id row: %s", err.Error())
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
			req := &pb.UpdateStudentPackageForCronjobRequest{}
			resp, err = orderService.UpdateStudentPackage(ctx, req)
			if err != nil {
				zlogger.Error("Cronjob payment_update_student_package: Failed to update student packages by cronjob with org_id = %s and err = %s", organizationID, err.Error())
				continue
			}
			for _, cronjobError := range resp.Errors {
				errDetail += fmt.Sprintf(`
upcoming_student_package_id = %s and err = %s
`, cronjobError.UpcomingStudentPackageId, cronjobError.Error)
			}
			zlogger.Info(fmt.Sprintf(`
		Executed job UpdateStudentPackageByUpcomingStudentPackage at org_id = %s
		Success case: %d
		Failed case: %d
		Errors : %s
		Execute time: %f second
	`, organizationID, resp.Successed, resp.Failed, errDetail, time.Since(now).Seconds()))
		}
	}

	zlogger.Info(fmt.Sprintf("---- Job UpdateStudentPackageByUpcomingStudentPackage ends at %s ----", time.Now()))
	return nil
}
