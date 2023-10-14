package mastermgmt

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	orga_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/organization/repositories"
	orga_service "github.com/manabie-com/backend/internal/mastermgmt/modules/organization/services"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	v1 "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

var (
	organizationID   string
	tenantID         string
	organizationName string
	domainName       string
	logoURL          string
	countryCode      string
)

func init() {
	bootstrap.RegisterJob("mastermgmt_create_organization", runJobCreateOrganization).
		Desc("Create Organization").
		StringVar(&organizationID, "organizationId", "", "Create Organization with ID").
		StringVar(&tenantID, "tenantId", "", "Create Organization with tenant ID").
		StringVar(&organizationName, "organizationName", "", "Create Organization with name").
		StringVar(&domainName, "domainName", "", "Create Organization with domain name").
		StringVar(&logoURL, "logoUrl", "", "Create Organization with logo url").
		StringVar(&countryCode, "countryCode", "", "Create Organization with country code")
}

func runJobCreateOrganization(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	// init logger
	zLogger := rsc.Logger()
	defer zLogger.Sugar().Sync() //nolint:errcheck
	zLogger.Sugar().Info("-----START: Job create organization -----")

	// init db
	dbTrace := rsc.DBWith("bob")

	// init nats-jetstream
	jsm := rsc.NATS()

	// init service
	organService := &orga_service.OrganizationService{
		DB:               dbTrace,
		JSM:              jsm,
		OrganizationRepo: &orga_repo.OrganizationRepo{},
	}

	if countryCode == "" {
		return fmt.Errorf("RunJobCreateOrganization: [err] countryCode cannot be empty")
	}

	countryVal, err := strconv.Atoi(countryCode)
	if err != nil {
		if value, found := v1.Country_value[countryCode]; found {
			countryVal = int(value)
		} else {
			return fmt.Errorf("RunJobCreateOrganization: [err] invalid country string value %q", countryCode)
		}
	}

	req := &pb.CreateOrganizationRequest{
		Organization: &pb.Organization{
			OrganizationId:   organizationID,
			TenantId:         tenantID,
			OrganizationName: organizationName,
			DomainName:       domainName,
			LogoUrl:          logoURL,
			CountryCode:      v1.Country(countryVal),
		},
	}

	// add resource path to ctx
	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	// Run create organization
	_, err = organService.CreateOrganization(ctx, req)
	if err != nil {
		return fmt.Errorf("RunJobCreateOrganization: [err] when CreateOrganization: %s", err)
	}

	zLogger.Sugar().Info("-----DONE: Job create organization -----")
	return nil
}
