package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type DomainUser struct {
	DB       libdatabase.Ext
	UserRepo interface {
		GetByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, externalUserIDs []string) (entity.Users, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.Users, error)
	}
}

/*
	type RepoForDomainUser struct {
		DomainUserRepo interface {
			UpdateEmail(ctx context.Context, db database.QueryExecer, usersToUpdate entity.DomainUser) error
		}
		DomainUsrEmailRepo DomainUsrEmailRepo
		UserRepo           interface {
			GetByEmail(ctx context.Context, db database.QueryExecer, emails pgtype.TextArray) ([]*entity.User, error)
			UpdateEmail(ctx context.Context, db database.QueryExecer, u *entity.User) error
		}
		OrganizationRepo service.OrganizationRepo
	}

	type DomainUsrEmailRepo interface {
		UpdateEmail(ctx context.Context, db database.QueryExecer, user entity.DomainUser) error
	}

	func UpdateEmail(ctx context.Context, db database.Ext, tenantManager multitenant.TenantManager, repo RepoForDomainUser, orgID string, userToUpdate entity.DomainUser) error {
		zapLogger := ctxzap.Extract(ctx)

		err := updateUserEmailInDatabase(ctx, db, repo, userToUpdate)
		if err != nil {
			return err
		}

		if orgID == "" {
			orgID = userToUpdate.OrganizationID().String()
		}

		tenantID, err := repo.OrganizationRepo.GetTenantIDByOrgID(ctx, db, orgID)
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.String("organizationID", orgID),
			)
			switch err {
			case pgx.ErrNoRows:
				return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: orgID}.Error())
			default:
				return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
			}
		}

		err = updateUserEmailOnAuthPlatform(ctx, tenantManager, tenantID, userToUpdate)
		if err != nil {
			return err
		}

		return nil
	}

	func updateUserEmailInDatabase(ctx context.Context, db database.Ext, repo RepoForDomainUser, userToUpdate entity.DomainUser) error {
		// Check if edited email already exists
		users, err := repo.UserRepo.GetByEmail(ctx, db, database.TextArray([]string{userToUpdate.Email().String()}))
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmail: %w", err).Error())
		}
		if len(users) > 0 {
			return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot edit student with email existing in system: %s", userToUpdate.Email().String()))
		}

		err = repo.DomainUserRepo.UpdateEmail(ctx, db, userToUpdate)
		if err != nil {
			return err
		}

		err = repo.DomainUsrEmailRepo.UpdateEmail(ctx, db, userToUpdate)
		if err != nil {
			return err
		}
		return nil
	}

	func updateUserEmailOnAuthPlatform(ctx context.Context, tenantManager multitenant.TenantManager, tenantID string, userToUpdate entity.DomainUser) error {
		zapLogger := ctxzap.Extract(ctx)

		tenantClient, err := tenantManager.TenantClient(ctx, tenantID)
		if err != nil {
			zapLogger.Sugar().Warnw(
				"cannot get tenant client",
				"tenantID", tenantID,
				"err", err.Error(),
			)
			return errors.Wrap(err, "TenantClient")
		}

		_, err = tenantClient.GetUser(ctx, userToUpdate.UserID().String())
		if err != nil {
			return err
		}

		//should remove this dependency ("firebase.google.com/go/v4/auth" package) after refactor internal package for auth
		userToUpdateOnAuthPlatform := (&auth.UserToUpdate{}).Email(userToUpdate.Email().String())

		_, err = tenantClient.LegacyUpdateUser(ctx, userToUpdate.UserID().String(), userToUpdateOnAuthPlatform)
		if err != nil {
			return err
		}

		return nil
	}
*/

func publishDomainUserEvent(ctx context.Context, jsm nats.JetStreamManagement, eventType string, users ...*pb.EvtUser) error {
	for _, event := range users {
		data, err := proto.Marshal(event)
		if err != nil {
			return fmt.Errorf("marshal event %s error, %w", eventType, err)
		}
		_, err = jsm.TracedPublish(ctx, "publishDomainUserEvent", eventType, data)
		if err != nil {
			return fmt.Errorf("publishDomainUserEvent with %s: s.JSM.Publish failed: %w", eventType, err)
		}
	}

	return nil
}

func (service *DomainUser) ValidateExternalUserIDExistedInSystem(ctx context.Context, users entity.Users) error {
	zapLogger := ctxzap.Extract(ctx)
	externalUserIDs := users.ExternalUserIDs()
	existingUsers, err := service.UserRepo.GetByExternalUserIDs(ctx, service.DB, externalUserIDs)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "UserRepo.GetByExternalUserIDs"),
			zap.Strings("externalUserIDs", externalUserIDs),
		)
		return entity.InternalError{
			RawErr: errors.Wrap(err, "UserRepo.GetByExternalUserIDs"),
		}
	}
	for _, user := range existingUsers {
		if user.ExternalUserID().IsEmpty() {
			continue
		}
		idx := utils.IndexOf(externalUserIDs, user.ExternalUserID().String())
		if user.UserID().String() == users[idx].UserID().String() {
			continue
		}
		zapLogger.Error(
			"existed external_user_id",
			zap.String("Function", "validateExternalUserIDExistedInSystem"),
			zap.Strings("existingUsers", users.UserIDs()),
			zap.Strings("externalUserIDs", externalUserIDs),
		)
		return entity.ExistingDataError{
			FieldName:  fmt.Sprintf("users[%d].external_user_id", idx),
			EntityName: entity.StaffEntity,
			Index:      idx,
		}
	}

	return nil
}

func (service *DomainUser) ValidateExternalUserIDIsExists(ctx context.Context, users entity.Users) error {
	zapLogger := ctxzap.Extract(ctx)
	userIDsToUpdate := users.UserIDs()
	existedUsers, err := service.UserRepo.GetByIDs(ctx, service.DB, userIDsToUpdate)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "UserRepo.GetByIDs"),
		)
		return entity.InternalError{
			RawErr: errors.Wrap(err, "UserRepo.GetByIDs"),
		}
	}
	if len(existedUsers) != len(userIDsToUpdate) {
		existedUserIDs := existedUsers.UserIDs()
		for _, userID := range userIDsToUpdate {
			if !golibs.InArrayString(userID, existedUserIDs) {
				zapLogger.Error(
					"len(existedUsers) != len(userIDsToUpdate)",
					zap.Error(err),
				)
				return entity.InternalError{
					RawErr: errors.Wrap(err, "UserRepo.GetByIDs"),
				}
			}
		}
	}

	/*
		if the external_user_id was updated, it would be failed:
		- create user with `external_user_id`, so `external_user_id` have value
			-> update user with another `external_user_id` will be failed.
		- create user without `external_user_id`, so `external_user_id` is null
			-> update again user with another `external_user_id` will be success.
	*/
	for _, user := range users {
		for _, existedUser := range existedUsers {
			if !user.UserID().Equal(existedUser.UserID()) {
				continue
			}

			if existedUser.ExternalUserID().IsEmpty() {
				continue
			}

			if user.ExternalUserID() != existedUser.ExternalUserID() {
				return entity.InternalError{
					RawErr: errors.Wrap(err, "user has external_user_id before can't update"),
				}
			}
		}
	}

	return nil
}
