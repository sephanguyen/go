package discount

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/discount/configurations"
	service "github.com/manabie-com/backend/internal/discount/services"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	dpb "github.com/manabie-com/backend/pkg/manabuf/discount/v1"
)

func init() {
	bootstrap.RegisterJob("discount_srvc_highest_discount_automation", runSysHighestDiscountAutomation).
		Desc("Start a cronjob highest discount automation").
		DescLong("updating student product discount to eligible highest discount")
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

func runSysHighestDiscountAutomation(ctx context.Context, _ configurations.Config, rsc *bootstrap.Resources) (err error) {
	zlogger := rsc.Logger().Sugar()
	zlogger.Info("start process HighestDiscountAutomation")

	paymentDB := rsc.DBWith("fatima")
	jsm := rsc.NATS()
	kafka := rsc.Kafka()
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
			fatimaDBTrace := &database.DBTrace{
				DB: paymentDB,
			}

			discountInternalService := service.NewInternalService(fatimaDBTrace, jsm, rsc.Logger(), kafka)
			req := &dpb.AutoSelectHighestDiscountRequest{
				OrganizationId: organizationID,
			}

			resp, err := discountInternalService.AutoSelectHighestDiscount(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to run automation for highest eligible discount: %s", err)
			}

			if len(resp.Errors) == 0 {
				zlogger.Info(fmt.Sprintf("Discount automation for organization %v SUCCESS: %v, %v total updated products", organizationID, time.Now(), resp.TotalUpdatedProducts))
			} else {
				zlogger.Info(fmt.Sprintf("Discount automation for organization %v Finished with some errors, %v total updated products", organizationID, resp.TotalUpdatedProducts))
				for _, res := range resp.Errors {
					zlogger.Error(fmt.Sprintf("Discount automation for organization %v with error: %v", organizationID, res.Error))
				}
			}
		}
	}

	zlogger.Info(fmt.Sprintf("End scheduled process HighestDiscountAutomation: %v", time.Now()))
	zlogger.Info("Process done")
	return nil
}
