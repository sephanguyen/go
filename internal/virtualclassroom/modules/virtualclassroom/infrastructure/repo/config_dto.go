package repo

import (
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
)

type Config struct {
	Key       pgtype.Text
	Group     pgtype.Text
	Country   pgtype.Text
	Value     pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
}

func (c *Config) FieldMap() ([]string, []interface{}) {
	return []string{
			"config_key",
			"config_group",
			"country",
			"config_value",
			"updated_at",
			"created_at",
		}, []interface{}{
			&c.Key,
			&c.Group,
			&c.Country,
			&c.Value,
			&c.UpdatedAt,
			&c.CreatedAt,
		}
}

func (c *Config) TableName() string {
	return "configs"
}

func (c *Config) ToConfigDomain() *domain.Config {
	return &domain.Config{
		Key:       c.Key.String,
		Group:     c.Key.String,
		Value:     c.Value.String,
		Country:   domain.Country(c.Country.String),
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}
}
