package entities

import (
	"github.com/jackc/pgtype"
)

// Config entity
type Config struct {
	Key       pgtype.Text
	Group     pgtype.Text
	Country   pgtype.Text
	Value     pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
}

// FieldMap return a map of field name and pointer to field
func (e *Config) FieldMap() ([]string, []interface{}) {
	return []string{
			"config_key", "config_group", "country", "config_value", "updated_at", "created_at",
		}, []interface{}{
			&e.Key, &e.Group, &e.Country, &e.Value, &e.UpdatedAt, &e.CreatedAt,
		}
}

// TableName returning "configs"
func (e *Config) TableName() string {
	return "configs"
}
