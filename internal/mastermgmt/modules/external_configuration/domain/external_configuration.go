package domain

import (
	"time"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

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

type LocationConfiguration struct {
	ID              string
	ConfigKey       string
	LocationID      string
	ConfigValue     string
	ConfigValueType string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       *time.Time
	ResourcePath    string
}

func (e *LocationConfiguration) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"location_config_id",
		"config_key",
		"location_id",
		"config_value",
		"config_value_type",
		"created_at",
		"updated_at",
		"deleted_at",
		"resource_path",
	}
	values = []interface{}{
		&e.ID,
		&e.ConfigKey,
		&e.LocationID,
		&e.ConfigValue,
		&e.ConfigValueType,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
		&e.ResourcePath,
	}
	return
}

func (e *LocationConfiguration) TableName() string {
	return "location_configuration_value"
}

func (e *LocationConfiguration) ToLocationConfigurationGRPCMessage() *mpb.LocationConfiguration {
	return &mpb.LocationConfiguration{
		Id:              e.ID,
		ConfigKey:       e.ConfigKey,
		LocationId:      e.LocationID,
		ConfigValue:     e.ConfigValue,
		ConfigValueType: e.ConfigValueType,
		CreatedAt:       timestamppb.New(e.CreatedAt),
		UpdatedAt:       timestamppb.New(e.UpdatedAt),
	}
}
