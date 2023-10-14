package usermgmt

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth"
	internal_auth_tenant "github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/identity/multitenant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func anUserWithValidInfo(schoolID int, hasPassword bool, scryptHash auth.ScryptHash) (*entity.LegacyUser, error) {
	random := rand.Intn(12345678)

	user, err := newUserEntity()
	if err != nil {
		return nil, fmt.Errorf("newUserEntity: %w", err)
	}

	user.Group = database.Text(constant.UserGroupStudent)
	user.ResourcePath = database.Text(strconv.Itoa(schoolID))

	if hasPassword {
		pwd := fmt.Sprintf("password%v", random)
		pwdSalt := "12345"

		hashedPwd, err := auth.HashedPassword(scryptHash, []byte(pwd), []byte(pwdSalt))
		if err != nil {
			return nil, fmt.Errorf("auth.HashedPassword: %v", err)
		}

		user.UserAdditionalInfo.Password = pwd
		user.UserAdditionalInfo.PasswordSalt = []byte(pwdSalt)
		user.UserAdditionalInfo.PasswordHash = hashedPwd
	}
	return user, nil
}

func tenantClientImportUsers(ctx context.Context, tenantClient internal_auth_tenant.TenantClient, users internal_auth_user.Users, hash auth.ScryptHash) (*internal_auth_user.ImportUsersResult, error) {
	result, err := tenantClient.ImportUsers(ctx, users, hash)
	if err != nil {
		return nil, errors.Wrap(err, "ImportUsers")
	}

	if len(result.UsersFailedToImport) > 0 {
		return nil, errors.New("failed to import users")
	}

	userMap := users.IDAndUserMap()

	for _, importedUser := range result.UsersSuccessImport {
		user, exists := userMap[importedUser.GetUID()]
		if !exists {
			return nil, fmt.Errorf(`can't find src user with id: "%v"`, importedUser.GetUID())
		}
		if !internal_auth_user.IsUserValueEqual(user, importedUser) {
			return nil, errors.New("users are not equal")
		}
	}

	return result, nil
}

func (s *suite) usersInFirebaseAuth(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Import user
	srcUser, err := anUserWithValidInfo(constants.ManabieSchool, true, s.FirebaseAuthClient.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "anUserWithValidInfo")
	}

	result, err := tenantClientImportUsers(ctx, s.FirebaseAuthClient, []internal_auth_user.User{srcUser}, s.FirebaseAuthClient.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.SrcUser = result.UsersSuccessImport[0]

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usersInTenant(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create tenant
	createdSrcTenant, err := s.TenantManager.CreateTenant(ctx, newTenantToCreate())
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "CreateTenant")
	}
	s.SrcTenant = createdSrcTenant

	srcTenantClient, err := s.TenantManager.TenantClient(ctx, s.SrcTenant.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "TenantClient")
	}

	// Import user
	srcUser, err := anUserWithValidInfo(constants.ManabieSchool, true, srcTenantClient.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "anUserWithValidInfo")
	}

	result, err := tenantClientImportUsers(ctx, srcTenantClient, []internal_auth_user.User{srcUser}, srcTenantClient.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.SrcUser = result.UsersSuccessImport[0]

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminImportUsersFromTenantToTenant(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create destination tenant
	createdDestTenant, err := s.TenantManager.CreateTenant(ctx, newTenantToCreate())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.DestTenant = createdDestTenant

	destTenantClient, err := s.TenantManager.TenantClient(ctx, createdDestTenant.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "TenantClient")
	}

	// Get src tenant
	srcTenantClient, err := s.TenantManager.TenantClient(ctx, s.SrcTenant.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "TenantClient")
	}

	err = srcTenantClient.IterateAllUsers(ctx, 1000, func(users internal_auth_user.Users) error {
		result, err := destTenantClient.ImportUsers(ctx, users, srcTenantClient.GetHashConfig())
		if err != nil {
			return errors.Wrap(err, "destTenantClient.ImportUsers")
		}
		if len(result.UsersFailedToImport) > 0 {
			return fmt.Errorf("failed to import users: %v", result.UsersFailedToImport.IDs())
		}
		s.DestUser = result.UsersSuccessImport[0]
		return nil
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usersInTenantStillHaveValidInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.loginIdentityPlatform(ctx, s.SrcTenant.GetID(), s.SrcUser.GetEmail(), s.SrcUser.GetRawPassword())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usersInTenantHasCorrespondingInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Login with password to import
	err := s.loginIdentityPlatform(ctx, s.DestTenant.GetID(), s.DestUser.GetEmail(), s.SrcUser.GetRawPassword())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) adminImportUsersFromFirebaseAuthToTenantInIdentityPlatform(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	// Create tenant
	createdDestTenant, err := s.TenantManager.CreateTenant(ctx, newTenantToCreate())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.DestTenant = createdDestTenant

	destTenantClient, err := s.TenantManager.TenantClient(ctx, s.DestTenant.GetID())
	if err != nil {
		return nil, errors.Wrap(err, "TenantClient")
	}

	// Import user
	result, err := tenantClientImportUsers(ctx, destTenantClient, []internal_auth_user.User{s.SrcUser}, s.FirebaseAuthClient.GetHashConfig())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	s.DestUser = result.UsersSuccessImport[0]

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usersInFirebaseAuthStillStillHaveValidInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if err := s.loginIdentityPlatform(ctx, auth.LocalTenants[constants.ManabieSchool], s.SrcUser.GetEmail(), s.SrcUser.GetRawPassword()); err != nil {
		return ctx, errors.Wrap(err, "loginIdentityPlatform")
	}

	return StepStateToContext(ctx, stepState), nil
}

var _ entity.User = (*ValidUser)(nil)

type ValidUser struct {
	entity.EmptyUser
	randomID      string
	attributeSeed string
	version       int
}

func (user *ValidUser) UserID() field.String {
	// Warning: ID must not be changed regardless version increasing/decreasing
	return field.NewString(user.randomID)
}
func (user *ValidUser) LoginEmail() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v@example.com", user.attributeSeed, user.version))
}
func (user *ValidUser) Password() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v", user.attributeSeed, user.version))
}
func (user *ValidUser) Group() field.String {
	return field.NewString(constant.UserGroupStudent)
}
func (user *ValidUser) FullName() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) FirstName() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) LastName() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) GivenName() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) FullNamePhonetic() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) FirstNamePhonetic() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) LastNamePhonetic() field.String {
	return field.NewString(user.attributeSeed)
}
func (user *ValidUser) Country() field.String {
	return field.NewString(pb.COUNTRY_VN.String())
}
func (user *ValidUser) PhoneNumber() field.String {
	return field.NewNullString()
}
func (user *ValidUser) Gender() field.String {
	return field.NewString("MALE")
}
func (user *ValidUser) OrganizationID() field.String {
	return ManabieOrg{}.OrganizationID()
}

var _ entity.User = (*UserHasEmptyID)(nil)

type UserHasEmptyID struct {
	*ValidUser
}

func (user *UserHasEmptyID) UserID() field.String {
	return field.NewString("")
}

var _ entity.User = (*UserHasInvalidEmailFormat)(nil)

type UserHasInvalidEmailFormat struct {
	*ValidUser
}

func (user *UserHasInvalidEmailFormat) LoginEmail() field.String {
	return field.NewString("123-invalid@.email.format@example.com")
}

type ManabieOrg struct {
}

func (org ManabieOrg) OrganizationID() field.String {
	return field.NewString(strconv.Itoa(constants.ManabieSchool))
}

func (org ManabieOrg) SchoolID() field.Int32 {
	return field.NewInt32(constants.ManabieSchool)
}

func CreateAuthUser(ctx context.Context, db database.Ext, firebaseClient internal_auth_tenant.TenantClient, tenantManager internal_auth_tenant.TenantManager, users ...entity.User) error {
	upserter := service.NewAuthUserUpserter(&repository.DomainUserRepo{}, (&repository.OrganizationRepo{}).WithDefaultValue("local"), firebaseClient, tenantManager)

	org := ManabieOrg{}

	err := database.ExecInTx(ctx, db, func(ctx context.Context, tx pgx.Tx) error {
		// Initialize
		for _, user := range users {
			// fmt.Println("user.UserID(): ", user.UserID())

			userAccessPaths := entity.DomainUserAccessPaths{}
			for _, locationID := range getChildrenLocation(OrgIDFromCtx(ctx)) {
				userAccessPath := &UserAccessPath{
					locationID: field.NewString(locationID),
					userID:     field.NewString(user.UserID().String()),
				}
				userAccessPaths = append(userAccessPaths, entity.UserAccessPathWillBeDelegated{
					HasLocationID:     userAccessPath,
					HasUserID:         userAccessPath,
					HasOrganizationID: &Organization{organizationID: golibs.ResourcePathFromCtx(ctx)},
				})
			}

			err := (&repository.DomainUserAccessPathRepo{}).UpsertMultiple(ctx, tx, userAccessPaths...)
			if err != nil {
				return errors.Wrap(err, "repo.UserAccessPathRepo.upsertMultiple")
			}

			err = (&repository.DomainUserRepo{}).UpsertMultiple(ctx, tx, true, user)
			if err != nil {
				return err
			}
		}

		if _, err := upserter(ctx, db, org, users, nil, unleash.DomainUserFeatureOption{}); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	/*existingUsers, err := (&repository.DomainUserRepo{}).GetByIDs(ctx, db, []string{users[0].UserID().String()})
	if err != nil {
		return err
	}
	fmt.Printf("test existingUsers: %+v\n", existingUsers)*/
	return err
}

func UpsertAuthProfiles(ctx context.Context, db database.Ext, firebaseClient internal_auth_tenant.TenantClient, tenantManager internal_auth_tenant.TenantManager, users ...entity.User) error {
	upserter := service.NewAuthUserUpserter(&repository.DomainUserRepo{}, (&repository.OrganizationRepo{}).WithDefaultValue("local"), firebaseClient, tenantManager)

	if _, err := upserter(ctx, db, ManabieOrg{}, users, nil, unleash.DomainUserFeatureOption{}); err != nil {
		return err
	}
	return nil
}

func UsersCanLoginIn(ctx context.Context, apiKey string, orgID int64, users ...entity.User) ([]*LoginInAuthPlatformResult, error) {
	identityPlatformLoginResults := make([]*LoginInAuthPlatformResult, 0, len(users))

	for _, user := range users {
		identityPlatformLoginResult, err := LoginInAuthPlatform(ctx, apiKey, multitenant.LocalTenants[orgID], user.LoginEmail().String(), user.Password().String())
		if err != nil {
			return nil, err
		}
		identityPlatformLoginResults = append(identityPlatformLoginResults, identityPlatformLoginResult)
	}
	return identityPlatformLoginResults, nil
}

func ValidUserToInvalidUser(currentValidUser *ValidUser, invalidCase string) entity.User {
	if currentValidUser == nil {
		currentValidUser = &ValidUser{
			randomID:      idutil.ULIDNow(),
			attributeSeed: idutil.ULIDNow(),
		}
	}
	switch invalidCase {
	case "but there is a profile has empty user id":
		return &UserHasEmptyID{
			ValidUser: currentValidUser,
		}
	case "but there is a profile has invalid email format":
		return &UserHasInvalidEmailFormat{
			ValidUser: currentValidUser,
		}
	default:
		// should fail to detect and fix immediately than continue running
		panic(errors.New("this invalid condition is not supported for user profile"))
	}
}

func (s *suite) systemCreateAuthProfilesWithValidInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := &ValidUser{
		randomID: idutil.ULIDNow(),
	}
	if err := UpsertAuthProfiles(s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin), s.BobDBTrace, s.FirebaseAuthClient, s.TenantManager, user); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.AuthUser = user
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) authProfilesAreCreatedSuccessfullyAndUsersCanUseThemToLoginInToSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if _, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, s.AuthUser); err != nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can not login in with auth profiles but actual they can")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemCreateAuthProfiles(ctx context.Context, invalidCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := ValidUserToInvalidUser(nil, invalidCase)
	// fmt.Println("invalid user:", user.UserID(), user.Email(), user.Password())

	s.ResponseErr = UpsertAuthProfiles(s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin), s.BobDBTrace, s.FirebaseAuthClient, s.TenantManager, user)
	s.AuthUser = user

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemFailedToCreateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	/*switch v := s.AuthUser.(type) {
	case *UserHasEmptyID:
		fmt.Printf("systemFailedToCreateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem *UserHasEmptyID '%v' '%v' '%v' \n", v.UserID(), v.Email(), v.Password())
	case *UserHasInvalidEmailFormat:
		fmt.Printf("systemFailedToCreateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem *UserHasInvalidEmailFormat '%v' '%v' '%v' \n", v.UserID(), v.Email(), v.Password())
	default:
		fmt.Printf("systemFailedToCreateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem s.AuthUser '%v' '%v' '%v' \n", v.UserID(), v.Email(), v.Password())
	}*/

	if s.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), errors.New("expected failed to create auth profiles, but created successfully")
	}
	if _, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, s.AuthUser); err == nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can not login in with auth profiles but actual they can")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) existingAuthProfilesInSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	attributeSeed := idutil.ULIDNow()

	/*createdUser, err := s.createStudentV2(s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin), attributeSeed)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}*/

	user := &ValidUser{
		randomID:      idutil.ULIDNow(),
		attributeSeed: attributeSeed,
	}
	// fmt.Println("1user: ", user.Email(), user.Password())

	if err := CreateAuthUser(s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin), s.BobDBTrace, s.FirebaseAuthClient, s.TenantManager, user); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	s.AuthUser = user

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemUpdateExistingAuthProfilesWithValidInfoArg(ctx context.Context, profileCondition string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	user := s.AuthUser.(*ValidUser)
	user.version++

	var userToUpdate entity.User
	switch profileCondition {
	case "but only email is changed":
		userToUpdate = &UserWithPasswordNeverChange{ValidUser: user}
	case "but only password is changed":
		userToUpdate = &UserWithEmailNeverChange{ValidUser: user}
	case "but email and password are not changed":
		userToUpdate = &UserWithEmailAndPasswordNeverChange{ValidUser: user}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf(`this "%s" is not supported`, profileCondition)
	}

	s.AuthUser = userToUpdate

	// fmt.Println("s.AuthUser.(*ValidUser).version: ", user.version)
	if err := UpsertAuthProfiles(s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin), s.BobDBTrace, s.FirebaseAuthClient, s.TenantManager, userToUpdate); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) authProfilesAreUpdatedSuccessfullyAndUsersCanUseThemToLoginInToSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := s.AuthUser.(*ValidUser)

	_, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can login in with new auth profiles but actual they can not")
	}
	/*fmt.Printf("%+v\n", firebaseIDTokens)
	fmt.Printf("%+v\n", identityPlatformIDTokens)*/

	// Profile before updated
	user.version--
	if _, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, user); err == nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can not login in with old auth profiles but actual they can")
	}
	return StepStateToContext(ctx, stepState), nil
}

type UserWithEmptyPassword struct {
	*ValidUser
}

func (user *UserWithEmptyPassword) Password() field.String {
	return field.NewString("")
}

type UserWithEmailNeverChange struct {
	*ValidUser
}

func (user *UserWithEmailNeverChange) LoginEmail() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v@example.com", user.attributeSeed, 0))
}

type UserWithPasswordNeverChange struct {
	*ValidUser
}

func (user *UserWithPasswordNeverChange) Password() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v", user.randomID, 0))
}

type UserWithEmailAndPasswordNeverChange struct {
	*ValidUser
}

func (user *UserWithEmailAndPasswordNeverChange) LoginEmail() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v@example.com", user.attributeSeed, 0))
}

func (user *UserWithEmailAndPasswordNeverChange) Password() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v", user.attributeSeed, 0))
}

func (s *suite) userAlreadyLoggedInWithExistingAuthProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := s.AuthUser.(*ValidUser)

	identityPlatformIDTokens, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "expected users can login in with new auth profiles but actual they can not")
	}

	s.AuthUserIdentityPlatformTokenID = identityPlatformIDTokens[0].IDToken
	s.AuthUserIdentityPlatformRefreshToken = identityPlatformIDTokens[0].RefreshToken

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemUpdateExistingAuthProfilesWithValidInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	user := s.AuthUser.(*ValidUser)
	user.version++

	if err := UpsertAuthProfiles(ctx, s.BobDBTrace, s.FirebaseAuthClient, s.TenantManager, user); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) authProfilesAreUpdatedSuccessfullyButUserDoesntNeedToLoginInAgainAlsoTheyStillCanLoginInAgainWithOldProfile(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := s.AuthUser // .(*ValidUser)

	_, err := ExchangeIDTokenByRefreshToken(ctx, s.Cfg.FirebaseAPIKey, s.AuthUserIdentityPlatformRefreshToken)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "expected can exchange another id token by refresh token, but actual can not")
	}

	_, err = UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can login in with new auth profiles but actual they can not")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) authProfilesAreUpdatedSuccessfullyAndUserCanUseThemLoginInIfAlreadyLoggedInUserHaveToLoginInAgain(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := s.AuthUser // .(*ValidUser)

	_, err := ExchangeIDTokenByRefreshToken(ctx, s.Cfg.FirebaseAPIKey, s.AuthUserFirebaseRefreshToken)
	if err == nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "expected can not exchange another id token by refresh token, but actual can")
	}

	_, err = UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, user)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can login in with new auth profiles but actual they can not")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) authProfilesAreUpdatedSuccessfullyWithoutChangingPasswordAndUsersCanUseThemToLoginInToSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user := s.AuthUser.(*ValidUser)

	// Why using UserWithPasswordNeverChange?
	// the password doesn't change since initialization (v0)

	fmt.Println("&UserWithPasswordNeverChange{user}: ", (&UserWithPasswordNeverChange{user}).Password().String())

	_, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, &UserWithPasswordNeverChange{user})
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can login in with new auth profiles but actual they can not")
	}
	/*fmt.Printf("%+v\n", firebaseIDTokens)
	fmt.Printf("%+v\n", identityPlatformIDTokens)*/

	// Profile before updated
	user.version--

	fmt.Println("&UserWithPasswordNeverChange{user}: ", (&UserWithPasswordNeverChange{user}).Password().String())

	if _, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, &UserWithPasswordNeverChange{user}); err == nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can not login in with old auth profiles but actual they can")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemUpdateExistingAuthProfiles(ctx context.Context, invalidCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = s.signedIn(ctx, constants.ManabieSchool, StaffRoleSchoolAdmin)

	// Change existing data to new info
	s.AuthUser.(*ValidUser).version++

	invalidUser := ValidUserToInvalidUser(s.AuthUser.(*ValidUser), invalidCase)
	s.ResponseErr = UpsertAuthProfiles(ctx, s.BobDBTrace, s.FirebaseAuthClient, s.TenantManager, invalidUser)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemFailedToUpdateAuthProfilesAndUsersCanNotUseThemToLoginInToSystem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if s.ResponseErr == nil {
		return StepStateToContext(ctx, stepState), errors.New("expected failed to update auth profiles, but updated successfully")
	}

	_, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, s.AuthUser)
	if err == nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can not login in with new auth profiles but actual they can")
	}

	s.AuthUser.(*ValidUser).version--
	if _, err := UsersCanLoginIn(ctx, s.Cfg.FirebaseAPIKey, constants.ManabieSchool, s.AuthUser); err != nil {
		return StepStateToContext(ctx, stepState), errors.New("expected users can login in with old auth profiles but actual they can not")
	}

	return StepStateToContext(ctx, stepState), nil
}
