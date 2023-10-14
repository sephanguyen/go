package entity

import (
	"time"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type TimesheetConfigs []*TimesheetConfig

type TimesheetConfig struct {
	ID          pgtype.Text
	ConfigType  pgtype.Text
	ConfigValue pgtype.Text
	IsArchived  pgtype.Bool
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (c *TimesheetConfig) FieldMap() ([]string, []interface{}) {
	return []string{
			"timesheet_config_id",
			"config_type",
			"config_value",
			"is_archived",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&c.ID,
			&c.ConfigType,
			&c.ConfigValue,
			&c.IsArchived,
			&c.CreatedAt,
			&c.UpdatedAt,
			&c.DeletedAt,
		}
}

func (*TimesheetConfig) TableName() string {
	return "timesheet_config"
}

func (*TimesheetConfig) PrimaryField() string {
	return "timesheet_config_id"
}

func (c *TimesheetConfig) PreInsert() error {
	now := time.Now()
	return multierr.Combine(
		c.CreatedAt.Set(now),
		c.UpdatedAt.Set(now),
		c.DeletedAt.Set(nil),
	)
}

func (c *TimesheetConfig) PreUpdate() error {
	now := time.Now()
	return multierr.Combine(
		c.UpdatedAt.Set(now),
		c.DeletedAt.Set(nil),
	)
}
