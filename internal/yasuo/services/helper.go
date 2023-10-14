// Package services
package services

import (
	"context"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	pb "github.com/manabie-com/backend/pkg/genproto/yasuo"

	"github.com/jackc/pgtype"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type getUserGroup func(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities.User, error)

func hasNotificationPermission(ctx context.Context, db database.QueryExecer, userGroup getUserGroup) bool {
	actorID := interceptors.UserIDFromContext(ctx)
	actorGroup, err := userGroup(ctx, db, database.Text(actorID))
	if err != nil {
		return false
	}

	return actorGroup.Group.String == pb.USER_GROUP_ADMIN.String()
}

func hasNotificationTargetPermission(ctx context.Context, db database.QueryExecer, userGroup getUserGroup) bool {
	actorID := interceptors.UserIDFromContext(ctx)
	actorGroup, err := userGroup(ctx, db, database.Text(actorID))
	if err != nil {
		return false
	}

	return actorGroup.Group.String == pb.USER_GROUP_ADMIN.String()
}

func inArrayInt(i int, arr []int) bool {
	for _, n := range arr {
		if i == n {
			return true
		}
	}
	return false
}

func ValidateAuth(ctx context.Context, db database.Ext, userProfile func(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (*entities.User, error), allowedGroups ...string) (string, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	uProfile, err := userProfile(ctx, db, database.Text(currentUserID))
	if err != nil {
		return "", status.Error(codes.Unknown, err.Error())
	}
	if len(allowedGroups) == 0 {
		allowedGroups = append(allowedGroups, constant.UserGroupStudent, constant.UserGroupAdmin)
	}
	for _, group := range allowedGroups {
		if uProfile.Group.String == group {
			return currentUserID, nil
		}
	}
	return "", status.Error(codes.PermissionDenied, "user group not allowed")
}
