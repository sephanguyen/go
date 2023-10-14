package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewPageRepo_GetBySlug(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	mt.Run("success", func(mt *mtest.T) {
		repo := &NewPageRepo{}
		expectedPage := domain.NewPage{
			ID:            "id-1",
			ApplicationID: "app-1",
			Deleted:       false,
		}

		mt.AddMockResponses(mtest.CreateCursorResponse(1, "appsmith.newPage", mtest.FirstBatch, bson.D{
			primitive.E{Key: "publishedPage.slug", Value: "slug-1"},
			primitive.E{Key: "deleted", Value: false},
			primitive.E{Key: "defaultResources.applicationId", Value: "app-1"},
			primitive.E{Key: "defaultResources.branchName", Value: "branchName-1"},
		}))
		res, err := repo.GetBySlug(ctx, mt.DB, "slug-1", "app-1", "branchName-1")

		assert.Nil(t, err)
		assert.Equal(t, expectedPage.Deleted, res.Deleted)
	})
}
