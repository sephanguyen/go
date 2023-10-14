package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/infrastructure"

	"go.mongodb.org/mongo-driver/mongo"
)

type AppsmithQueryHandler struct {
	DB          *mongo.Database
	NewPageRepo infrastructure.NewPageRepo
}

func (a *AppsmithQueryHandler) GetPageInfoBySlug(ctx context.Context, slug, applicationID, branchName string) (*domain.NewPage, error) {
	return a.NewPageRepo.GetBySlug(ctx, a.DB, slug, applicationID, branchName)
}
