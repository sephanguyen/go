package domain

import "time"

type InternalConfiguration struct {
	ID              string
	ConfigKey       string
	ConfigValue     string
	ConfigValueType string
	LastEditor      *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	ResourcePath    string
}

func (i *InternalConfiguration) FieldMap() (fields []string, values []interface{}) {
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
		&i.ID,
		&i.ConfigKey,
		&i.ConfigValue,
		&i.ConfigValueType,
		&i.LastEditor,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.DeletedAt,
		&i.ResourcePath,
	}
	return
}

func (i *InternalConfiguration) TableName() string {
	return "internal_configuration_value"
}

type ExternalConfiguration struct {
	ID              string
	ConfigKey       string
	ConfigValue     string
	ConfigValueType string
	LastEditor      *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	ResourcePath    string
}

func (e *ExternalConfiguration) FieldMap() (fields []string, values []interface{}) {
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
		&e.ID,
		&e.ConfigKey,
		&e.ConfigValue,
		&e.ConfigValueType,
		&e.LastEditor,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
		&e.ResourcePath,
	}
	return
}

func (e *ExternalConfiguration) TableName() string {
	return "external_configuration_value"
}
