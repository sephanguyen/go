package service

import (
	"context"

	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
)

// contains all logics to deal with domain school admin
type DomainLocation struct {
	DB libdatabase.Ext
}

type DomainLocationRepo interface {
	Create(ctx context.Context, db libdatabase.QueryExecer, userToCreate aggregate.DomainStudent) error
}
