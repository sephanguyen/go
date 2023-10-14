package services

import (
	"context"

	entities_bob "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/caching"
	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgtype"
	"golang.org/x/sync/singleflight"
)

type InternalService struct {
	pb.UnimplementedInternalServer
	DB database.Ext

	StudentOrderRepo interface {
		ListOrderForProcessing(ctx context.Context, db database.QueryExecer, processingBefore pgtype.Timestamptz, status, gateway pgtype.Text) ([]*entities_bob.StudentOrder, error)
	}
}

func NewInternalServiceCacher(cacher caching.LocalCacher, svc pb.InternalServer) *InternalServiceCacher {
	return &InternalServiceCacher{
		Cacher:         cacher,
		InternalServer: svc,
		group:          singleflight.Group{},
	}
}

type InternalServiceCacher struct {
	Cacher caching.LocalCacher
	pb.InternalServer

	group singleflight.Group
}
