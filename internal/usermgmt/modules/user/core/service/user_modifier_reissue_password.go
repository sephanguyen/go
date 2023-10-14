package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserModifierService) ReissueUserPassword(ctx context.Context, req *pb.ReissueUserPasswordRequest) (*pb.ReissueUserPasswordResponse, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	if err := validateReissueUserPasswordReq(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	callerID := interceptors.UserIDFromContext(ctx)
	existedUser, err := s.UserRepo.Get(ctx, s.DB, database.Text(req.UserId))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get user: %s", err.Error()))
	}

	if callerID != existedUser.ID.String {
		callerRoles, err := s.UserRepo.GetUserRoles(ctx, s.DB, database.Text(callerID))
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get caller's roles: %s", err.Error()))
		}

		ableToReissuePassword := false
		for _, roleName := range callerRoles.ListRoleNames() {
			if golibs.InArrayString(roleName, []string{constant.RoleSchoolAdmin, constant.RoleHQStaff, constant.RoleCentreLead, constant.RoleCentreManager, constant.RoleCentreStaff}) {
				ableToReissuePassword = true
				break
			}
		}
		if !ableToReissuePassword {
			return nil, status.Error(codes.PermissionDenied, "user don't have permission to reissue password")
		}
	}

	switch existedUser.Group.String {
	case entity.UserGroupStudent, entity.UserGroupParent, entity.UserGroupTeacher:
		break
	default:
		return nil, status.Error(codes.PermissionDenied, "school staff and school admin don't have permission to reissue this user password")
	}

	// Import to identity platform
	tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, s.DB, organization.OrganizationID().String())
	if err != nil {
		zapLogger.Error(
			"cannot get tenant id",
			zap.Error(err),
			zap.String("organizationID", organization.OrganizationID().String()),
		)
		switch err {
		case pgx.ErrNoRows:
			return nil, status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: organization.OrganizationID().String()}.Error())
		default:
			return nil, status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
		}
	}

	err = s.overrideUserPasswordInIdentityPlatform(ctx, tenantID, req.UserId, req.NewPassword)
	if err != nil {
		zapLogger.Error(
			"cannot update user password on identity platform",
			zap.Error(err),
			zap.String("organizationID", organization.OrganizationID().String()),
			zap.String("tenantID", tenantID),
			zap.String("userID", req.UserId),
		)
		switch err {
		case user.ErrUserNotFound:
			return nil, status.Error(codes.NotFound, errcode.NewUserNotFoundErr(req.UserId).Error())
		default:
			return nil, status.Error(codes.Internal, errors.Wrap(err, "failed to update user password on identity platform").Error())
		}
	}

	return &pb.ReissueUserPasswordResponse{Successful: true}, nil
}

func validateReissueUserPasswordReq(req *pb.ReissueUserPasswordRequest) error {
	switch {
	case req.UserId == "", req.NewPassword == "":
		return fmt.Errorf("invalid params")
	case len(req.NewPassword) < firebase.MinimumPasswordLength:
		return fmt.Errorf("password length must be larger than 6")
	}

	return nil
}

func (s *UserModifierService) overrideUserPasswordInIdentityPlatform(ctx context.Context, tenantID string, userID string, password string) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	tenantClient, err := s.TenantManager.TenantClient(ctx, tenantID)
	if err != nil {
		zapLogger.Warnw(
			"cannot get tenant client",
			"tenantID", tenantID,
			"err", err.Error(),
		)
		return errors.Wrap(err, "TenantClient")
	}

	return overrideUserPassword(ctx, tenantClient, userID, password)
}
