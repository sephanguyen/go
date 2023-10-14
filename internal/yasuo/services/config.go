package services

import (
	"context"
	"fmt"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/multierr"
)

type ConfigService struct {
	DB database.Ext

	ConfigRepo interface {
		Upsert(ctx context.Context, db database.Ext, configs []*entities_bob.Config) error
	}
}

type UpsertConfig struct {
	Key     string
	Group   string
	Country string
	Value   string
}

func (c *ConfigService) UpsertConfig(ctx context.Context, upsertReq *UpsertConfig) error {
	config := &entities_bob.Config{}
	resourcePath := golibs.ResourcePathFromCtx(ctx)
	database.AllNullEntity(config)
	if err := multierr.Combine(
		config.Key.Set(fmt.Sprintf("%s_%s", upsertReq.Key, resourcePath)),
		config.Group.Set(upsertReq.Group),
		config.Country.Set(upsertReq.Country),
		config.Value.Set(upsertReq.Value),
	); err != nil {
		return fmt.Errorf("ConfigService.Combine: %w", err)
	}

	if err := c.ConfigRepo.Upsert(ctx, c.DB, []*entities_bob.Config{config}); err != nil {
		return fmt.Errorf("ConfigRepo.Upsert: %w", err)
	}

	return nil
}
