package fatima

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/fatima/repositories"
	"github.com/manabie-com/backend/internal/fatima/services"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/configurations"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

var MapOrgIDUserID = map[string]string{
	"-2147483623": "01GTZYX224982Z1X4MHZQW6BO3",
	"-2147483642": "01GTZYX224982Z1X4MHZQW6DQ4",
	"-2147483635": "01GTZYX224982Z1X4MHZQW6DP7",
	// priority for KEC first
	"-2147483622": "01GTZYX224982Z1X4MHZQW6BO2",
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
	"-2147483637": "01GTZYX224982Z1X4MHZQW6DP8",
	"-2147483638": "01GTZYX224982Z1X4MHZQW6DP9",
	"-2147483639": "01GTZYX224982Z1X4MHZQW6DQ1",
	"-2147483640": "01GTZYX224982Z1X4MHZQW6DQ2",
	"-2147483641": "01GTZYX224982Z1X4MHZQW6DQ3",
	"-2147483643": "01GTZYX224982Z1X4MHZQW6DQ5",
	"-2147483644": "01GTZYX224982Z1X4MHZQW6DQ6",
	"-2147483645": "01GTZYX224982Z1X4MHZQW6DQ7",
	"-2147483646": "01GTZYX224982Z1X4MHZQW6DQ8",
	"-2147483647": "01GTZYX224982Z1X4MHZQW6DQ9",
	"-2147483648": "01GTZYX224982Z1X4MHZQW6DR1",
}

func init() {
	bootstrap.RegisterJob("job_schedule_class", jobScheduleClass).
		Desc("trigger schedule reserve classes by effective date")
}

func jobScheduleClass(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	zLogger := zapLogger.Sugar()
	dateTime := time.Now().Add(9 * time.Hour) // compatible with JP time
	dateStr := dateTime.Format("2006/01/02")
	zLogger.Infof("Trigger schedule reserve classes which effective at %s", dateStr)

	unleashClient := rsc.Unleash()

	fatimaDB := rsc.DBWith("fatima")
	mastermgmtConn := rsc.GRPCDial("mastermgmt")

	internalMastermgmtClient := mpb.NewMasterInternalServiceClient(mastermgmtConn)

	subscriptionModifySvc := &services.SubscriptionModifyService{
		DB:                           fatimaDB,
		PackageRepo:                  &repositories.PackageRepo{},
		StudentPackageRepo:           &repositories.StudentPackageRepo{},
		StudentPackageAccessPathRepo: &repositories.StudentPackageAccessPathRepo{},
		StudentPackageClassRepo:      &repositories.StudentPackageClassRepo{},
		JSM:                          rsc.NATS(),
	}

	orgQuery := "SELECT organization_id FROM organizations" //nolint:goconst
	organizations, err := fatimaDB.Query(ctx, orgQuery)
	if err != nil {
		return fmt.Errorf("failed to get orgs: %w", err)
	}
	defer organizations.Close()

	zLogger.Info("Get organizations to schedule class organizations")
	for organizations.Next() {
		var organizationID string
		if err := organizations.Scan(&organizationID); err != nil {
			zLogger.Error("failed to scan an orgs row: %w", err)
			continue
		}
		isEnabled, err := unleashClient.IsFeatureEnabledOnOrganization("Payment_StudentCourse_BackOffice_ClassManagementByArch", c.Common.Environment, organizationID)
		if err != nil {
			zLogger.Error("failed to get feature flag status of organization %s: %w", organizationID, err)
			continue
		}

		if !isEnabled {
			zLogger.Info(fmt.Sprintf("Feature schedule class is disabled on organization %s", organizationID))
			continue
		}

		if _, ok := MapOrgIDUserID[organizationID]; ok {
			claim := &interceptors.CustomClaims{
				Manabie: &interceptors.ManabieClaims{
					UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
					ResourcePath: organizationID,
					UserID:       MapOrgIDUserID[organizationID],
				},
			}
			ctx = interceptors.ContextWithJWTClaims(ctx, claim)

			req := &mpb.GetReserveClassesByEffectiveDateRequest{
				OrganizationId: organizationID,
				EffectiveDate:  timestamppb.New(dateTime),
			}

			resp, err := internalMastermgmtClient.GetReserveClassesByEffectiveDate(ctx, req)

			if err != nil {
				zLogger.Error("failed to GetReserveClasses at %s: %w", dateStr, err)
				continue
			}

			rcs := resp.GetReserveClasses()

			if len(rcs) == 0 {
				zLogger.Info(fmt.Sprintf("not found reserve class on organization %s at %s", organizationID, dateStr))
				continue
			}
			zLogger.Info(fmt.Sprintf("Get reserve classes on organization %s at %s done", organizationID, dateStr))

			rcInfo := sliceutils.Map(rcs, func(rc *mpb.GetReserveClassesByEffectiveDateResponse_ReserveClass) *fpb.WrapperRegisterStudentClassRequest_ReserveClassInformation {
				return &fpb.WrapperRegisterStudentClassRequest_ReserveClassInformation{
					StudentPackageId: rc.StudentPackageId,
					StudentId:        rc.StudentId,
					CourseId:         rc.CourseId,
					ClassId:          rc.ClassId,
				}
			})

			reqRegister := &fpb.WrapperRegisterStudentClassRequest{
				ReserveClassesInformation: rcInfo,
			}

			err = subscriptionModifySvc.WrapperRegisterStudentClass(ctx, reqRegister)

			if err != nil {
				zLogger.Error("failed to schedule reserve classes at %s: %w", dateStr, err)
				continue
			}
			zLogger.Info(fmt.Sprintf("Register all reserve classes on organization %s at %s done", organizationID, dateStr))

			deleteReserveClassesReq := &mpb.DeleteReserveClassByEffectiveDateRequest{
				OrganizationId: organizationID,
				EffectiveDate:  timestamppb.New(dateTime),
			}

			_, err = internalMastermgmtClient.DeleteReserveClassesByEffectiveDate(ctx, deleteReserveClassesReq)

			zLogger.Info(fmt.Sprintf("Delete all reserve classes on organization %s at %s done", organizationID, dateStr))
			if err != nil {
				zLogger.Error("failed to delete reserve classes at %s: %w", dateStr, err)
				continue
			}
		}
	}
	zLogger.Info(fmt.Sprintf("Schedule class done for all organizations at %s", dateStr))

	return nil
}
