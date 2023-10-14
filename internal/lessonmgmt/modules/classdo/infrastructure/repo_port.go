package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/classdo/infrastructure/repo"
)

type ClassDoAccountRepo interface {
	UpsertClassDoAccounts(ctx context.Context, db database.QueryExecer, classDoAccounts domain.ClassDoAccounts) error
	GetAllClassDoAccounts(ctx context.Context, db database.QueryExecer) ([]*repo.ClassDoAccount, error)
	GetClassDoAccountByID(ctx context.Context, db database.QueryExecer, ID string) (*repo.ClassDoAccount, error)
}
