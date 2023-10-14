package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type ImportMasterDataForTestService struct {
	DB          database.Ext
	ForTestRepo interface {
		Insert(ctx context.Context, db database.QueryExecer, e database.Entity, excludedFields []string) error
	}
}
