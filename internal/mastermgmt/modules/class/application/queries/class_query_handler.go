package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"
)

type ClassQueryHandler struct {
	DB        database.Ext
	ClassRepo infrastructure.ClassRepo
}

func (c *ClassQueryHandler) GetByIDs(ctx context.Context, payload GetByIds) ([]*domain.Class, error) {
	return c.ClassRepo.RetrieveByIDs(ctx, c.DB, payload.IDs)
}
