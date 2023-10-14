package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ClassCommandHandler struct {
	DB        database.Ext
	ClassRepo infrastructure.ClassRepo
}

func (c *ClassCommandHandler) Create(ctx context.Context, payload CreateClass) error {
	err := database.ExecInTx(ctx, c.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := c.ClassRepo.Insert(ctx, tx, payload.Classes)
		return err
	})
	if err != nil {
		return fmt.Errorf("ClassRepo.Insert: %w", err)
	}
	return nil
}

func (c *ClassCommandHandler) UpdateByID(ctx context.Context, payload UpdateClassById) error {
	return c.ClassRepo.UpdateClassNameByID(ctx, c.DB, payload.ID, payload.Name)
}

func (c *ClassCommandHandler) Delete(ctx context.Context, payload DeleteClassById) error {
	return c.ClassRepo.Delete(ctx, c.DB, payload.ID)
}
