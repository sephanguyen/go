package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/organization/entities"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// OrganizationRepo provides method to work with organization entity
type OrganizationRepo struct{}

func (r *OrganizationRepo) Create(ctx context.Context, db database.QueryExecer, organization *entities.Organization) error {
	ctx, span := interceptors.StartSpan(ctx, "OrganizationRepo.Create")
	defer span.End()

	// database.AllNullEntity(organization)
	now := time.Now()

	err := multierr.Combine(
		organization.ID.Set(organization.ID.String),
		organization.TenantID.Set(organization.TenantID.String),
		organization.Name.Set(organization.Name.String),
		organization.ResourcePath.Set(organization.ResourcePath.String),
		organization.DomainName.Set(organization.DomainName.String),
		organization.LogoURL.Set(organization.LogoURL.String),
		organization.Country.Set(organization.Country.String),
		organization.UpdatedAt.Set(now),
		organization.CreatedAt.Set(now),
	)
	if err != nil {
		return fmt.Errorf("err set entity: %w", err)
	}

	orga, err := database.Insert(ctx, organization, db.Exec)
	if err != nil {
		return err
	}

	if orga.RowsAffected() != 1 {
		return errors.New("cannot insert new organization")
	}

	return nil
}
