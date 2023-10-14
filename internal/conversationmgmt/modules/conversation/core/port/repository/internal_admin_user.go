package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type InternalAdminUserRepo interface {
	GetOne(ctx context.Context, db database.QueryExecer) (*domain.InternalAdminUser, error)
}
