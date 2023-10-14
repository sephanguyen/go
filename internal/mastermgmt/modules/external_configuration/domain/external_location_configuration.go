package domain

import (
	"time"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type LocationConfigurationV2 struct {
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

func (e *LocationConfigurationV2) FieldMap() (fields []string, values []interface{}) {
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

func (e *LocationConfigurationV2) TableName() string {
	return "location_configuration_value_v2"
}

func (e *LocationConfigurationV2) ToLocationConfigurationGRPCMessage() *mpb.LocationConfiguration {
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
