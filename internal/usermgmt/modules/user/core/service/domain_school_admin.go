package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/database"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DomainSchoolAdmin represents for a domain service that
// contains all logics to deal with domain school admin
type DomainSchoolAdmin struct {
	DB            libdatabase.Ext
	TenantManager multitenant.TenantManager

	DomainSchoolAdminRepo DomainSchoolAdminRepo
}

type DomainSchoolAdminRepo interface {
	Create(ctx context.Context, db libdatabase.QueryExecer, userToCreate aggregate.DomainSchoolAdmin) error
}

func (service *DomainSchoolAdmin) CreateSchoolAdmin(ctx context.Context, schoolAdminToCreate entity.DomainSchoolAdmin, isEnableUsername bool) error {
	zapLogger := ctxzap.Extract(ctx)

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "OrganizationFromContext")
	}

	schoolAdmin := aggregate.DomainSchoolAdmin{
		DomainSchoolAdmin: &entity.SchoolAdminToDelegate{
			DomainSchoolAdminProfile: schoolAdminToCreate,
			HasOrganizationID:        organization,
			HasSchoolID:              organization,
			HasUserID:                schoolAdminToCreate,
		},
		LegacyUserGroups: entity.LegacyUserGroups{entity.DelegateToLegacyUserGroup(&entity.SchoolAdminLegacyUserGroup{}, organization, schoolAdminToCreate)},
	}

	if err := aggregate.ValidSchoolAdmin(schoolAdmin, isEnableUsername); err != nil {
		return errors.Wrap(err, "ValidSchoolAdmin")
	}

	err = database.ExecInTx(ctx, service.DB, func(ctx context.Context, tx database.Tx) error {
		err := service.DomainSchoolAdminRepo.Create(ctx, tx, schoolAdmin)
		if err != nil {
			return errors.Wrap(err, "Create")
		}

		// Import to identity platform
		tenantID, err := new(repository.OrganizationRepo).GetTenantIDByOrgID(ctx, tx, organization.OrganizationID().String())
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.String("organizationID", organization.OrganizationID().String()),
			)
			switch err {
			case pgx.ErrNoRows:
				return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: organization.OrganizationID().String()}.Error())
			default:
				return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
			}
		}

		backwardCompatibleAuthUser := &entity.LegacyUser{
			ID:    libdatabase.Text(schoolAdmin.UserID().String()),
			Email: libdatabase.Text(schoolAdmin.Email().String()),
			UserAdditionalInfo: entity.UserAdditionalInfo{
				Password: schoolAdmin.Password().String(),
			},
		}

		err = CreateUsersInIdentityPlatform(ctx, service.TenantManager, tenantID, entity.LegacyUsers{backwardCompatibleAuthUser}, int64(organization.SchoolID().Int32()))
		if err != nil {
			zapLogger.Error(
				"cannot create users on identity platform",
				zap.Error(err),
				zap.String("organizationID", organization.OrganizationID().String()),
				zap.String("tenantID", tenantID),
				zap.String("email", backwardCompatibleAuthUser.Email.String),
			)
			switch err {
			case internal_auth_user.ErrUserNotFound:
				return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(backwardCompatibleAuthUser.Email.String).Error())
			default:
				return status.Error(codes.Internal, err.Error())
			}
		}

		return nil
	})
	if err != nil {
		return errors.Wrap(err, "ExecInTx")
	}

	return nil
}

func CreateUsersInIdentityPlatform(ctx context.Context, tenantManager multitenant.TenantManager, tenantID string, users []*entity.LegacyUser, resourcePath int64) error {
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

	err = createUserInAuthPlatform(ctx, tenantClient, users, resourcePath)
	if err != nil {
		zapLogger.Sugar().Warnw(
			"cannot create users on identity platform",
			"err", err.Error(),
		)
		return errors.Wrap(err, "createUserInAuthPlatform")
	}

	return nil
}
