package usermgmt

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) usersInOurSystemHaveBeenImportedToFirebaseAuth(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	firebaseAuthAlternative, err := s.TenantManager.TenantClient(ctx, usermgmt.LocalTestMigrationTenant)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)
	user, err := anUserWithValidInfo(constants.ManabieSchool, true, firebaseAuthAlternative.GetHashConfig())
	if err != nil {
		return nil, err
	}

	userAccessPaths := entity.DomainUserAccessPaths{}
	for _, locationID := range getChildrenLocation(OrgIDFromCtx(ctx)) {
		userAccessPath := &UserAccessPath{
			locationID: field.NewString(locationID),
			userID:     field.NewString(user.ID.String),
		}
		userAccessPaths = append(userAccessPaths, entity.UserAccessPathWillBeDelegated{
			HasLocationID:     userAccessPath,
			HasUserID:         userAccessPath,
			HasOrganizationID: &Organization{organizationID: golibs.ResourcePathFromCtx(ctx)},
		})
	}

	if err := libdatabase.ExecInTx(ctx, s.BobDBTrace, func(ctx context.Context, tx pgx.Tx) error {
		err = (&repository.DomainUserAccessPathRepo{}).UpsertMultiple(ctx, tx, userAccessPaths...)
		if err != nil {
			return errors.Wrap(err, "repo.UserAccessPathRepo.upsertMultiple")
		}

		err = (&repository.UserRepo{}).Create(ctx, tx, user)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	_, err = tenantClientImportUsers(ctx, firebaseAuthAlternative, []internal_auth_user.User{user}, firebaseAuthAlternative.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.Users = append(s.Users, user)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateUsersFromFirebaseAuth(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	orgID := golibs.ResourcePathFromCtx(ctx)
	err := usermgmt.MigrateUsersFromFirebase(ctx, &configurations.Config{
		Common:           s.Cfg.Common,
		PostgresV2:       s.Cfg.PostgresV2,
		FirebaseAPIKey:   s.Cfg.FirebaseAPIKey,
		JWTApplicant:     s.Cfg.JWTApplicant,
		IdentityPlatform: *s.Cfg.IdentityPlatform,
	}, orgID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) infoOfUsersInFirebaseAuthIsStillValid(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, userInDB := range s.Users {
		err := s.loginIdentityPlatform(ctx, usermgmt.LocalTestMigrationTenant, userInDB.GetEmail(), userInDB.GetRawPassword())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) infoOfUsersInTenantOfIdentityPlatformHasCorrespondingInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, userInDB := range s.Users {
		userSchoolIDStr := strings.Split(userInDB.ResourcePath.String, ":")[0]
		userSchoolID, err := strconv.Atoi(userSchoolIDStr)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		tenantID := usermgmt.LocalSchoolAndTenantIDMap[userSchoolID]

		tenantClient, err := s.TenantManager.TenantClient(ctx, tenantID)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		identityPlatformUser, err := tenantClient.GetUser(ctx, userInDB.ID.String)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if !internal_auth_user.IsUserValueEqual(identityPlatformUser, userInDB) {
			return StepStateToContext(ctx, stepState), errors.New("users are not equal")
		}

		for schoolID, tenantID := range usermgmt.LocalSchoolAndTenantIDMap {
			err := s.loginIdentityPlatform(ctx, tenantID, userInDB.GetEmail(), userInDB.GetRawPassword())

			if schoolID == userSchoolID && err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf(`expected user can login to this tenant id: "%v" but actual can't login'`, tenantID)
			}

			if schoolID != userSchoolID && err == nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf(`expected user can't login to this tenant id: "%v" but actual can login`, tenantID)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
