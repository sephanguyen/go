package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type ExternalConfiguration struct {
	ConfigurationIDAttr field.String
	ConfigKeyAttr       field.String
	ConfigValueAttr     field.String
	ConfigValueTypeAttr field.String
	LastEditorAttr      field.String
	OrganizationIDAttr  field.String

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

type DomainExternalConfigurationRepo struct{}

func (e *ExternalConfiguration) ConfigID() field.String {
	return e.ConfigurationIDAttr
}
func (e *ExternalConfiguration) ConfigKey() field.String {
	return e.ConfigKeyAttr
}
func (e *ExternalConfiguration) ConfigValue() field.String {
	return e.ConfigValueAttr
}
func (e *ExternalConfiguration) ConfigValueType() field.String {
	return e.ConfigValueTypeAttr
}
func (e *ExternalConfiguration) OrganizationID() field.String {
	return e.ConfigurationIDAttr
}

func (e *ExternalConfiguration) FieldMap() ([]string, []interface{}) {
	return []string{
			"configuration_id",
			"config_key",
			"config_value",
			"config_value_type",
			"last_editor",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.ConfigurationIDAttr,
			&e.ConfigKeyAttr,
			&e.ConfigValueAttr,
			&e.ConfigValueTypeAttr,
			&e.LastEditorAttr,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.OrganizationIDAttr,
		}
}

func (e *ExternalConfiguration) TableName() string {
	return "external_configuration_value"
}

func (*DomainExternalConfigurationRepo) GetConfigurationByKeys(ctx context.Context, db database.QueryExecer, keys []string) ([]entity.DomainConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "DomainExternalConfigurationRepo.GetConfigurationByKeys")
	defer span.End()

	stmt := `SELECT %s FROM %s WHERE config_key = ANY($1)`

	externalConfiguration := ExternalConfiguration{}
	fieldName, _ := externalConfiguration.FieldMap()

	stmt = fmt.Sprintf(stmt, strings.Join(fieldName, ","), externalConfiguration.TableName())

	rows, err := db.Query(ctx, stmt, database.TextArray(keys))
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	result := make([]entity.DomainConfiguration, 0, len(keys))
	for rows.Next() {
		externalConfiguration := ExternalConfiguration{}
		_, fieldValue := externalConfiguration.FieldMap()

		err := rows.Scan(fieldValue...)
		if err != nil {
			return nil, fmt.Errorf("row.Scan: %w", err)
		}

		result = append(result, &externalConfiguration)
	}
	return result, nil
}
