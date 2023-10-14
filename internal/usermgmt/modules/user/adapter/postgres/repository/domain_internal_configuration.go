package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainInternalConfigurationRepo struct{}

type InternalConfiguration struct {
	InternalConfigurationAttribute

	CreatedAt field.Time
	UpdatedAt field.Time
	DeletedAt field.Time
}

type InternalConfigurationAttribute struct {
	ConfigID        field.String
	ConfigKey       field.String
	ConfigValue     field.String
	ConfigValueType field.String
	LastEditor      field.String
	OrganizationID  field.String
}

func (config *InternalConfiguration) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"configuration_id",
		"config_key",
		"config_value",
		"config_value_type",
		"last_editor",
		"created_at",
		"updated_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&config.InternalConfigurationAttribute.ConfigID,
		&config.InternalConfigurationAttribute.ConfigKey,
		&config.InternalConfigurationAttribute.ConfigValue,
		&config.InternalConfigurationAttribute.ConfigValueType,
		&config.InternalConfigurationAttribute.LastEditor,
		&config.CreatedAt,
		&config.UpdatedAt,
		&config.DeletedAt,
		&config.InternalConfigurationAttribute.OrganizationID,
	}
	return
}

func NewInternalConfiguration(config entity.DomainConfiguration) *InternalConfiguration {
	now := field.NewTime(time.Now())
	return &InternalConfiguration{
		InternalConfigurationAttribute: InternalConfigurationAttribute{
			ConfigID:        config.ConfigID(),
			ConfigKey:       config.ConfigKey(),
			ConfigValue:     config.ConfigValue(),
			ConfigValueType: config.ConfigValueType(),
			OrganizationID:  config.OrganizationID(),
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: field.NewNullTime(),
	}
}

func (config *InternalConfiguration) TableName() string {
	return "internal_configuration_value"
}

func (config *InternalConfiguration) ConfigID() field.String {
	return config.InternalConfigurationAttribute.ConfigID
}

func (config *InternalConfiguration) ConfigKey() field.String {
	return config.InternalConfigurationAttribute.ConfigKey
}

func (config *InternalConfiguration) ConfigValue() field.String {
	return config.InternalConfigurationAttribute.ConfigValue
}

func (config *InternalConfiguration) ConfigValueType() field.String {
	return config.InternalConfigurationAttribute.ConfigValueType
}

func (config *InternalConfiguration) OrganizationID() field.String {
	return config.InternalConfigurationAttribute.OrganizationID
}

func (repo *DomainInternalConfigurationRepo) GetByKey(ctx context.Context, db database.QueryExecer, configKey string) (entity.DomainConfiguration, error) {
	ctx, span := interceptors.StartSpan(ctx, "ConfigRepo.GetByKey")
	defer span.End()

	config := NewInternalConfiguration(entity.NullDomainConfiguration{})
	fields, values := config.FieldMap()
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		WHERE config_key = $1
		AND deleted_at IS NULL`,
		strings.Join(fields, ","),
		config.TableName(),
	)

	if err := db.QueryRow(ctx, query, configKey).Scan(values...); err != nil {
		return nil, InternalError{
			RawError: err,
		}
	}

	return config, nil
}
