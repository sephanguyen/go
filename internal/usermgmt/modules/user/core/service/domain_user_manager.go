package service

import (
	"context"
	"fmt"

	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"

	"github.com/jackc/pgtype"
)

type AuthUserUpserter func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error)

// TODO: cut-off dependency with entity.Users
func NewAuthUserUpserter(userRepo UserRepo, organizationRepo OrganizationRepo, firebaseTenantClient multitenant.TenantClient, identityPlatformTenantManager multitenant.TenantManager) AuthUserUpserter {
	return func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, _ unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
		usersToUpsert := make(entity.Users, 0, len(usersToCreate)+len(usersToUpdate))
		usersToUpsert = append(usersToUpsert, usersToCreate...)
		usersToUpsert = append(usersToUpsert, usersToUpdate...)

		authUsers := make(entity.LegacyUsers, 0, len(usersToCreate)+len(usersToUpdate))
		if len(usersToUpsert) == 0 {
			return authUsers, nil
		}

		tenantID, err := organizationRepo.GetTenantIDByOrgID(ctx, db, organization.OrganizationID().String())
		switch err {
		case nil:
			if tenantID == "" {
				return nil, errcode.ErrTenantOfOrgNotFound{OrganizationID: organization.OrganizationID().String()}
			}
		default:
			return nil, errcode.ErrInternalFailedToImportAuthErr{Err: err}
		}

		existingUsers, err := userRepo.GetByIDs(ctx, db, usersToUpsert.UserIDs())
		if err != nil {
			return nil, err
		}

		// old logic
		/*for _, userToUpsert := range usersToUpsert {
			authUser := &entity.User{
				ID:    libdatabase.Text(userToUpsert.UserID().String()),
				Email: libdatabase.Text(userToUpsert.Email().String()),
				UserAdditionalInfo: entity.UserAdditionalInfo{
					Password: userToUpsert.Password().String(),
				},
			}
			authUsers = append(authUsers, authUser)
		}*/

		tenantClient, err := identityPlatformTenantManager.TenantClient(ctx, tenantID)
		if err != nil {
			switch err {
			case internal_auth_user.ErrTenantNotFound:
				return nil, errcode.ErrIdentityPlatformTenantNotFound{TenantID: tenantID}
			default:
				return nil, errcode.ErrInternalFailedToImportAuthErr{Err: err}
			}
		}
		identityPlatformUserProfiles, err := userToAuthProfile(tenantClient, existingUsers, usersToUpsert, int64(organization.SchoolID().Int32()))
		if err != nil {
			return nil, err
		}
		if err := upsertUserInAuthPlatform(ctx, tenantClient, identityPlatformUserProfiles); err != nil {
			return nil, err
		}

		return authUsers, nil
	}
}

func userToAuthProfile(authClient multitenant.TenantClient, existingUsers entity.Users, usersToUpsert entity.Users, schoolID int64) (internal_auth_user.Users, error) {
	userIDToExistingUser := make(map[string]entity.User, len(existingUsers))
	for _, existingUser := range existingUsers {
		userIDToExistingUser[existingUser.UserID().String()] = existingUser
	}

	usersNeedToUpsert := make(entity.Users, 0, len(usersToUpsert))
	for _, userToUpsert := range usersToUpsert {
		existingUser, exist := userIDToExistingUser[userToUpsert.UserID().String()]

		// does not exist mean this is first time we create this user
		if !exist {
			usersNeedToUpsert = append(usersNeedToUpsert, userToUpsert)
			continue
		}

		// If user changed at least email or password, its profile need to be updated
		if userToUpsert.LoginEmail().String() != existingUser.LoginEmail().String() {
			usersNeedToUpsert = append(usersNeedToUpsert, userToUpsert)
			continue
		}

		encryptedUserIDByPassword, err := entity.EncryptedUserIDByPasswordFromUser(userToUpsert)
		if err != nil {
			return nil, err
		}
		// EncryptedUserIDByPasswordFromUser() return null string for user has empty password
		// if we import a user with empty password to firebase the user can not login
		// skip user with empty password
		if encryptedUserIDByPassword.String() == "" {
			continue
		}
		if encryptedUserIDByPassword.String() != existingUser.EncryptedUserIDByPassword().String() {
			usersNeedToUpsert = append(usersNeedToUpsert, userToUpsert)
			continue
		}
	}
	// re-assign to skip some users don't need to update
	usersToUpsert = usersNeedToUpsert

	authUsers := make(internal_auth_user.Users, 0, len(usersNeedToUpsert))
	for _, userToUpsert := range usersToUpsert {
		authUser := &entity.LegacyUser{
			ID:         libdatabase.Text(userToUpsert.UserID().String()),
			LoginEmail: libdatabase.Text(userToUpsert.LoginEmail().String()),
			UserAdditionalInfo: entity.UserAdditionalInfo{
				Password: userToUpsert.Password().String(),
			},
		}

		authUser.CustomClaims = utils.CustomUserClaims(authUser.Group.String, authUser.ID.String, schoolID)

		passwordSalt := []byte(idutil.ULIDNow())

		var hashedPwd []byte
		if authUser.Password != "" {
			var err error
			hashedPwd, err = internal_auth.HashedPassword(authClient.GetHashConfig(), []byte(authUser.Password), passwordSalt)
			if err != nil {
				return nil, errcode.ErrScryptIsInvalidErr{Err: err}
			}
		}

		authUser.PhoneNumber.Status = pgtype.Null
		authUser.PhoneNumber = libdatabase.Text("")
		authUser.PasswordSalt = passwordSalt
		authUser.PasswordHash = hashedPwd

		authUsers = append(authUsers, authUser)
	}

	return authUsers, nil
}

func upsertUserInAuthPlatform(ctx context.Context, authClient multitenant.TenantClient, authUsers internal_auth_user.Users) error {
	/*var authUsers internal_auth_user.Users
	for i := range users {
		users[i].CustomClaims = utils.CustomUserClaims(users[i].Group.String, users[i].ID.String, schoolID)

		passwordSalt := []byte(idutil.ULIDNow())

		var hashedPwd []byte
		if users[i].Password != "" {
			var err error
			hashedPwd, err = internal_auth.HashedPassword(authClient.GetHashConfig(), []byte(users[i].Password), passwordSalt)
			if err != nil {
				return errcode.ErrScryptIsInvalidErr{Err: err}
			}
		}

		users[i].PhoneNumber.Status = pgtype.Null
		users[i].PhoneNumber = libdatabase.Text("")
		users[i].PasswordSalt = passwordSalt
		users[i].PasswordHash = hashedPwd

		authUsers = append(authUsers, users[i])
	}*/

	if len(authUsers) < 1 {
		return nil
	}

	result, err := authClient.ImportUsers(ctx, authUsers, authClient.GetHashConfig())
	if err != nil {
		return errcode.ErrFailedToImportAuthUsersToTenantErr{
			Err:      err,
			TenantID: authClient.TenantID(),
		}
	}

	if len(result.UsersFailedToImport) > 0 {
		var errs []string
		for _, userFailedToImport := range result.UsersFailedToImport {
			errs = append(errs, fmt.Sprintf("{'%s' - '%s' : %s}", userFailedToImport.User.GetUID(), userFailedToImport.User.GetEmail(), userFailedToImport.Err))
		}
		return errcode.ErrAuthProfilesHaveIssueWhenImport{
			ErrMessages: errs,
			TenantID:    authClient.TenantID(),
		}
	}
	return nil
}

/*func UpsertUsersInIdentityPlatform(ctx context.Context, tenantManager multitenant.TenantManager, tenantID string, users []*entity.User, schoolID int64) error {
	tenantClient, err := tenantManager.TenantClient(ctx, tenantID)

	if err != nil {
		switch err {
		case internal_auth_user.ErrTenantNotFound:
			return errcode.ErrIdentityPlatformTenantNotFound{TenantID: tenantID}
		default:
			return errcode.ErrInternalFailedToImportAuthErr{Err: err}
		}
	}

	return upsertUserInAuthPlatform(ctx, tenantClient, users, schoolID)
}
*/
