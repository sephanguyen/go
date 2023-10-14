package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) staffUpdateUserStatusSuccessfully(ctx context.Context, activationStatus string) (context.Context, error) {
	domainStudent := service.DomainStudent{
		UserRepo: &repository.DomainUserRepo{},
		DB:       s.BobDBTrace,
	}

	deactivatedAt, err := activationStatusToDeactivatedAt(activationStatus)
	if err != nil {
		return ctx, err
	}

	if err := domainStudent.UpdateUserActivation(
		ctx,
		entity.Users{
			&repository.User{
				ID:                field.NewString(s.UserId),
				DeactivatedAtAttr: deactivatedAt,
			},
		},
	); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *suite) checkLoginStatus(ctx context.Context, _, loginStatus string) (context.Context, error) {
	if s.OldAuthToken != "" {
		if _, err := authenticateUserToken(ctx, s.BobDBTrace, s.ShamirConn, s.Cfg.JWTApplicant, s.UserId, s.OldAuthToken, loginStatus); err != nil {
			return ctx, errors.Wrap(err, "failed to check login status for old fakeIDToken")
		}
	}

	fakeIDToken, exchangedToken, err := authenticateUserAndGenerateTokens(ctx, s.BobDBTrace, s.ShamirConn, s.Cfg.JWTApplicant, s.UserId, s.CurrentUserGroup, loginStatus)
	if err != nil {
		return ctx, errors.Wrap(err, "failed to check login status for new fakeIDToken")
	}

	s.OldAuthToken = fakeIDToken
	s.AuthToken = exchangedToken
	return ctx, nil
}

func (s *suite) userGetSelfProfileByOldToken(ctx context.Context, role, abilityOperation string) (context.Context, error) {
	ctx = contextWithTokenV2(ctx, s.AuthToken)
	ctx = interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: golibs.ResourcePathFromCtx(ctx),
			UserID:       s.UserId,
			UserGroup:    GetLegacyUserGroupFromConstant(role),
		},
	})

	_, errResp := getBasicProfile(ctx, s.UserMgmtConn, &upb.GetBasicProfileRequest{})
	if err := checkAbilityOfUserAction(ctx, abilityOperation, errResp, s.BobDBTrace, s.UserId); err != nil {
		return nil, errors.Wrap(err, "failed to check ability to get self profile")
	}

	return ctx, nil
}

func (s *suite) staffTryUpdateUserStatus(ctx context.Context, activationStatus, userType string) (context.Context, error) {
	var userID string
	switch userType {
	case "none-existed":
		userID = "00000000-0000-0000-0000-000000000000"
	case "deleted":
		user, err := s.createStudentUser(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create student user")
		}

		userID = user.UserID().String()
		if _, err := s.BobDBTrace.Exec(
			ctx, `
			UPDATE users
			SET deleted_at = NOW()
			WHERE user_id = $1
			`, userID,
		); err != nil {
			return ctx, errors.Wrap(err, "failed to delete user")
		}
	default:
		return ctx, fmt.Errorf("invalid user type: %s", userType)
	}

	domainStudent := service.DomainStudent{
		UserRepo: &repository.DomainUserRepo{},
		DB:       s.BobDBTrace,
	}

	deactivatedAt, err := activationStatusToDeactivatedAt(activationStatus)
	if err != nil {
		return ctx, err
	}

	if err := domainStudent.UpdateUserActivation(
		ctx,
		entity.Users{
			&repository.User{
				ID:                field.NewString(userID),
				DeactivatedAtAttr: deactivatedAt,
			},
		},
	); err != nil {
		s.ResponseErr = errcodeToGRPCStatus(err)
	}
	return ctx, nil
}

func authenticateUserAndGenerateTokens(ctx context.Context, db database.QueryExecer, conn *grpc.ClientConn, jwtApplicant, userID, userGroup, loginStatus string) (string, string, error) {
	fakeIDToken, err := GenFakeIDToken(firebaseAddr, userID, "templates/"+userGroup+".template")
	if err != nil {
		return "", "", errors.Wrap(err, "failed to generate fake ID token")
	}

	exchangedToken, err := authenticateUserToken(ctx, db, conn, jwtApplicant, userID, fakeIDToken, loginStatus)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to authenticate user token")
	}

	return fakeIDToken, exchangedToken, nil
}

func authenticateUserToken(ctx context.Context, db database.QueryExecer, conn *grpc.ClientConn, jwtApplicant string, userID string, token string, abilityLogin string) (string, error) {
	// use exchange token to check if user can log in
	exchangedToken, errResp := ExchangeToken(ctx, conn, jwtApplicant, userID, token)
	if err := checkAbilityOfUserAction(ctx, abilityLogin, errResp, db, userID); err != nil {
		return "", errors.Wrap(err, "failed to check ability login")
	}
	return exchangedToken, nil
}

func activationStatusToDeactivatedAt(activationStatus string) (field.Time, error) {
	switch activationStatus {
	case "re-activates":
		return field.NewNullTime(), nil
	case "deactivates":
		return field.NewTime(time.Now()), nil
	default:
		return field.NewNullTime(), errors.New("invalid activation status")
	}
}

func errcodeToGRPCStatus(err error) error {
	statusCodeByErrCode := map[errcode.Code]codes.Code{
		errcode.DataExist:          codes.AlreadyExists,
		errcode.DuplicatedData:     codes.AlreadyExists,
		errcode.InternalError:      codes.Internal,
		errcode.BadRequest:         codes.InvalidArgument,
		errcode.InvalidData:        codes.InvalidArgument,
		errcode.InvalidMaximumRows: codes.InvalidArgument,
		errcode.InvalidPayloadSize: codes.InvalidArgument,
		errcode.MissingField:       codes.InvalidArgument,
		errcode.MissingMandatory:   codes.InvalidArgument,
		errcode.UpdateFieldFail:    codes.InvalidArgument,
	}

	e, _ := err.(errcode.Error)
	code, ok := statusCodeByErrCode[errcode.Code(e.Code)]
	if !ok {
		return status.Error(codes.Unknown, e.Error())
	}
	return status.Error(code, e.Error())
}

func checkAbilityOfUserAction(ctx context.Context, abilityOperation string, errResp error, db database.QueryExecer, userID string) error {
	switch abilityOperation {
	case "can":
		if errResp != nil {
			return errors.Wrap(errResp, "expect can but actually cannot")
		}
	case "cannot":
		if errResp == nil {
			return errors.Wrap(errResp, "expect cannot but actually can")
		}

		if err := validateDeactivatedResponse(ctx, db, userID, errResp); err != nil {
			return err
		}
	}
	return nil
}

func validateDeactivatedResponse(ctx context.Context, db database.QueryExecer, userID string, errResp error) error {
	isDeactivated, err := isDeactivatedUser(ctx, db, userID)
	if err != nil {
		return errors.Wrap(err, "failed to check if user is deactivated")
	}

	if isDeactivated {
		if !strings.Contains(errResp.Error(), errorx.ErrDeactivatedUser.Error()) {
			return errors.Wrap(errResp, "expect user cannot login because user is deactivated but got other error")
		}
	}
	return nil
}

func isDeactivatedUser(ctx context.Context, db database.QueryExecer, userID string) (bool, error) {
	isDeactivated := false
	err := db.QueryRow(ctx, `
		SELECT
			CASE WHEN
				deactivated_at IS NOT NULL
				THEN true
				ELSE false
			END AS is_deactivated
		FROM users
		WHERE user_id = $1
	`, userID).Scan(&isDeactivated)
	if err != nil {
		return false, err
	}
	return isDeactivated, nil
}
