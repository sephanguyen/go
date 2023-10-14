package repo

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// follow: https://github.com/victorneuret/mongo-go-driver-mock/blob/master/insert_test.go

func TestEventLogRepo_SaveLog(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mt.Run("success", func(mt *mtest.T) {
		repo := &LogRepo{}
		log := map[string]interface{}{
			"UserID": "user-id",
			"Event":  "event",
		}
		log["Context"] = map[string]interface{}{
			"IP": "user-ip",
		}

		mt.AddMockResponses(mtest.CreateSuccessResponse())
		res, err := repo.SaveLog(ctx, mt.DB, log)

		assert.Nil(t, err)

		assert.NotEmpty(t, log["ID"])

		assert.Equal(t, log["Context"], res["Context"])
		assert.Equal(t, log["Event"], res["Event"])
		assert.Equal(t, log["UserID"], res["UserID"])
	})

	mt.Run("custom error duplicate", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))
		repo := &LogRepo{}

		res, err := repo.SaveLog(ctx, mt.DB, domain.EventLog{})

		assert.Nil(t, res)
		assert.NotNil(t, err)
		assert.True(t, mongo.IsDuplicateKeyError(err))
	})

	mt.Run("simple error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 0}})
		repo := &LogRepo{}

		res, err := repo.SaveLog(ctx, mt.DB, domain.EventLog{})

		assert.Nil(t, res)
		assert.NotNil(t, err)
	})
}
