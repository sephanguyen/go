package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"go.mongodb.org/mongo-driver/mongo"
)

type NewPageRepo interface {
	GetBySlug(ctx context.Context, db *mongo.Database, slug, applicationID, branchName string) (n *domain.NewPage, err error)
}

type LogRepo interface {
	SaveLog(ctx context.Context, db *mongo.Database, log domain.EventLog) (domain.EventLog, error)
}
