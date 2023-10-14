package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/subject/domain"
)

type SubjectRepo interface {
	Import(ctx context.Context, db database.Ext, subjects []*domain.Subject) error
	GetByIDs(ctx context.Context, db database.QueryExecer, ids []string) ([]*domain.Subject, error)
	GetByNames(ctx context.Context, db database.QueryExecer, ids []string) ([]*domain.Subject, error)
	GetAll(ctx context.Context, db database.QueryExecer) ([]*domain.Subject, error)
}
