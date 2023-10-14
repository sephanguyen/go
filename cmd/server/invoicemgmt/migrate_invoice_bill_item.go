package invoicemgmt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/configurations"
	"github.com/manabie-com/backend/internal/invoicemgmt/repositories"
	services "github.com/manabie-com/backend/internal/invoicemgmt/services/data_migration"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"go.uber.org/zap"
)

// Use to get the organization ID where the job will run
var (
	organizationID string
	userID         string
)

func init() {
	bootstrap.RegisterJob("invoicemgmt_migrate_invoice_bill_item", RunMigrateInvoiceBillItem).
		StringVar(&organizationID, "organizationID", "", "organization ID to run the job").
		StringVar(&userID, "userID", "", "user ID to run the job")
}

// To run this job, use `invoicemgmt migrate-invoice-bill-item -- --organizationID="org-id-value"`
func RunMigrateInvoiceBillItem(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	db := rsc.DB()
	sugaredLogger := rsc.Logger().Sugar()

	return MigrateInvoiceBillItem(ctx, db, sugaredLogger, organizationID, userID)
}

func MigrateInvoiceBillItem(
	ctx context.Context,
	db *database.DBTrace,
	sugaredLogger *zap.SugaredLogger,
	orgID string,
	adminUserID string,
) error {
	err := validateOrgAndUserID(ctx, db, orgID, adminUserID)
	if err != nil {
		return err
	}

	tenantContext := setTenantContext(ctx, orgID, adminUserID)

	repos := initRepositories()
	migrationService := services.NewDataMigrationModifierService(*sugaredLogger, db, getDataMigrationServiceRepositories(repos))
	err = migrationService.InsertInvoiceBillItemDataMigration(tenantContext)
	if err != nil {
		return fmt.Errorf("%v orgID: %v", err, orgID)
	}

	return nil
}

func setTenantContext(ctx context.Context, orgID, adminUserID string) context.Context {
	// Set the orgID in the context
	tenantContext := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: orgID,
			UserGroup:    cpb.UserGroup_USER_GROUP_SCHOOL_ADMIN.String(),
		},
	})

	// Assign the user_id to Manabie claims
	claims := interceptors.JWTClaimsFromContext(tenantContext)
	if claims != nil {
		claims.Manabie.UserID = adminUserID
		tenantContext = interceptors.ContextWithJWTClaims(tenantContext, claims)
	}

	return tenantContext
}

func validateOrgAndUserID(ctx context.Context, db database.QueryExecer, orgID, adminUserID string) error {
	if strings.TrimSpace(orgID) == "" {
		return errors.New("organizationID cannot be empty")
	}

	if strings.TrimSpace(adminUserID) == "" {
		return errors.New("userID cannot be empty")
	}

	// Check if valid organization ID
	orgRepo := &repositories.OrganizationRepo{}
	_, err := orgRepo.FindByID(ctx, db, orgID)
	if err != nil {
		return fmt.Errorf("error querying organization err: %v", err)
	}

	return nil
}
