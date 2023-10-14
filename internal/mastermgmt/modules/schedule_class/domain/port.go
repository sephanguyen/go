package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type ReserveClassRepo interface {
	InsertOne(ctx context.Context, db database.QueryExecer, reserveClass *ReserveClass) error
}
