package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	"github.com/pkg/errors"
)

type OrganizationRepo struct{}

func (r *OrganizationRepo) GetOrganizations(ctx context.Context, db database.QueryExecer) ([]*entities.Organization, error) {
	ctx, span := interceptors.StartSpan(ctx, "OrganizationRepo.GetOrganizations")
	defer span.End()

	e := &entities.Organization{}
	fields, _ := e.FieldMap()
	stmt := fmt.Sprintf("SELECT %s FROM %s", strings.Join(fields, ","), e.TableName())

	rows, err := db.Query(ctx, stmt)
	if err != nil {
		return nil, err
	}

	organizations := []*entities.Organization{}
	defer rows.Close()
	for rows.Next() {
		organization := new(entities.Organization)
		_, fieldValues := organization.FieldMap()
		err := rows.Scan(fieldValues...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}
		organizations = append(organizations, organization)
	}

	return organizations, nil
}

func (r *OrganizationRepo) FindByID(ctx context.Context, db database.QueryExecer, organizationID string) (*entities.Organization, error) {
	ctx, span := interceptors.StartSpan(ctx, "OrganizationRepo.FindByID")
	defer span.End()

	organization := &entities.Organization{}
	fields, _ := organization.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE organization_id = $1", strings.Join(fields, ","), organization.TableName())

	err := database.Select(ctx, db, query, organizationID).ScanOne(organization)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Errorf("db.QueryRowEx %v", organizationID).Error())
	}
	return organization, nil
}
