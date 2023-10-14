package repo

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type NewPageRepo struct{}

func (np *NewPageRepo) GetBySlug(ctx context.Context, db *mongo.Database, slug, applicationID, branchName string) (page *domain.NewPage, err error) {
	ctx, span := interceptors.StartSpan(ctx, "newPageRepo.GetBySlug")
	defer span.End()

	newPageDTO := &NewPage{}
	filter := bson.D{
		primitive.E{Key: "publishedPage.slug", Value: slug},
		primitive.E{Key: "deleted", Value: false},
		primitive.E{Key: "defaultResources.applicationId", Value: applicationID},
		primitive.E{Key: "defaultResources.branchName", Value: branchName},
	}
	collection := db.Collection(newPageDTO.ColletionName())

	err = collection.FindOne(ctx, filter).Decode(newPageDTO)
	if err != nil {
		return nil, err
	}
	return newPageDTO.ToNewPageEntity(), nil
}
