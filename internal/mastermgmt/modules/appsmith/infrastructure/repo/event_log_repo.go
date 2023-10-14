package repo

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/appsmith/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type LogRepo struct {
}

func (l *LogRepo) SaveLog(ctx context.Context, db *mongo.Database, log domain.EventLog) (domain.EventLog, error) {
	ctx, span := interceptors.StartSpan(ctx, "LogRepo.SaveLog")
	defer span.End()

	e := domain.EventLog{}
	coll := db.Collection(e.CollectionName())
	res, err := coll.InsertOne(ctx, log)

	if err != nil {
		return nil, err
	}
	log["ID"] = res.InsertedID.(primitive.ObjectID).String()

	return log, nil
}
