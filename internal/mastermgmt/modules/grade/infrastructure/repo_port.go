package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/grade/domain"
)

type GradeRepo interface {
	GetByPartnerInternalIDs(ctx context.Context, db database.QueryExecer, pIDs []string) (g []*domain.Grade, err error)
	Import(ctx context.Context, db database.Ext, grades []*domain.Grade) error
	GetAll(ctx context.Context, db database.QueryExecer) (g []*domain.Grade, err error)
}
