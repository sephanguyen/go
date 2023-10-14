package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/auth/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type OrganizationRepo struct{}

type OrganizationAttribute struct {
	OrganizationID     field.String
	OrganizationName   field.String
	TenantID           field.String
	SalesforceClientID field.String
}

type Organization struct {
	OrganizationAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

func NewOrganization(organization entity.DomainOrganization) *Organization {
	now := field.NewTime(time.Now())

	return &Organization{
		OrganizationAttribute: OrganizationAttribute{
			OrganizationID:     organization.OrganizationID(),
			OrganizationName:   organization.OrganizationName(),
			TenantID:           organization.TenantID(),
			SalesforceClientID: organization.SalesforceClientID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (repo *Organization) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"organization_id",
			"organization_name",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
			"tenant_id",
			"salesforce_client_id"},
		[]interface{}{
			&repo.OrganizationAttribute.OrganizationID,
			&repo.OrganizationAttribute.OrganizationName,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&repo.DeletedAt,
			&repo.OrganizationAttribute.OrganizationID,
			&repo.OrganizationAttribute.TenantID,
			&repo.OrganizationAttribute.SalesforceClientID}
}

func (repo *Organization) TableName() string {
	return "organizations"
}

func (repo *Organization) OrganizationID() field.String {
	return repo.OrganizationAttribute.OrganizationID
}

func (repo *Organization) OrganizationName() field.String {
	return repo.OrganizationAttribute.OrganizationName
}

func (repo *Organization) TenantID() field.String {
	return repo.OrganizationAttribute.TenantID
}

func (repo *Organization) SalesforceClientID() field.String {
	return repo.OrganizationAttribute.SalesforceClientID
}

func (repo *OrganizationRepo) GetSalesforceClientIDByOrganizationID(ctx context.Context, db database.QueryExecer, organizationID string) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, "OrganizationRepo.GetByOrganizationID")
	defer span.End()

	organization := NewOrganization(entity.NullOrganization{})
	stmt := fmt.Sprintf("SELECT salesforce_client_id FROM %s WHERE organization_id = $1", organization.TableName())

	salesforceClientID := ""
	err := db.QueryRow(ctx, stmt, database.Text(organizationID)).Scan(&salesforceClientID)
	if err != nil {
		return "", repository.InternalError{
			RawError: err,
		}
	}

	return salesforceClientID, nil
}
