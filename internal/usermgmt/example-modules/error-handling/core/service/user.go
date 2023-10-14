package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/example-modules/error-handling/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type User struct {
	db database.QueryExecer

	UserRepo interface {
		UpsertUsers(ctx context.Context, db database.QueryExecer, users entity.Users) error
		GetUsers(ctx context.Context, db database.QueryExecer, userIDs field.Strings) (entity.Users, error)
	}
}

func (service User) UpsertUsers(ctx context.Context, users entity.Users) error {
	zapLogger := ctxzap.Extract(ctx)

	if err := entity.ValidateUsers(users); err != nil {
		//
		zapLogger.Error(
			"error occurs when validate users",
			zap.Error(err),
		)
		return err
	}

	err := service.UserRepo.UpsertUsers(ctx, service.db, users)
	if err != nil {
		return err
	}

	return nil
}

func (service User) GetUsers(ctx context.Context, userIDs field.Strings) (entity.Users, error) {
	zapLogger := ctxzap.Extract(ctx)

	existingUsers, err := service.UserRepo.GetUsers(ctx, service.db, userIDs)
	if err != nil {
		zapLogger.Error(
			"error occurs when get users",
			zap.Error(err),
		)
		return nil, err
	}

	if len(existingUsers) != len(userIDs) {
		userIDToUserMap := existingUsers.UserIDs()

		for i, userIDToCheck := range userIDs {
			if _, found := userIDToUserMap[userIDToCheck.String()]; found {
				continue
			}

			return nil, entity.NotFoundError{
				EntityName:         "user",
				Index:              i,
				SearchedFieldName:  string(entity.UserFieldUserID),
				SearchedFieldValue: userIDToCheck.String(),
			}
		}
	}

	return existingUsers, nil
}
