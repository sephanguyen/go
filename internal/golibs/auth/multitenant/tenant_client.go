package multitenant

import (
	"context"
	"encoding/base64"
	"net/url"

	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/gcp"

	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/auth/hash"
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"google.golang.org/api/iterator"
)

type TenantClient interface {
	TenantID() string
	GetHashConfig() *gcp.HashConfig
	UserPager(ctx context.Context, nextPageToken string, pageSize int) *Pager
	GetUser(ctx context.Context, uid string) (internal_user.User, error)
	LegacyUpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (*auth.UserRecord, error)
	ImportUsers(ctx context.Context, users internal_user.Users, importHash internal_auth.ScryptHash) (*internal_user.ImportUsersResult, error)
	IterateAllUsers(ctx context.Context, pageSize int, iteratedUsersCallback func(users internal_user.Users) error) error
	CustomToken(ctx context.Context, uid string) (string, error)
	PasswordResetLink(ctx context.Context, email, langCode string) (string, error)

	//SetHashConfig sets a custom hash config instead of fetching from tenant config. ONLY USER FOR TESTING FOR NOW
	SetHashConfig(hashConfig *gcp.HashConfig)
}

type tenantClient struct {
	tenantID   string
	HashConfig *gcp.HashConfig

	gcpClient internal_auth.GCPTenantClient
	gcpUtils  internal_auth.GCPUtils
}

func defaultTenantClient(gcpTenantClient internal_auth.GCPTenantClient) *tenantClient {
	t := tenantClient{
		gcpClient: gcpTenantClient,
		gcpUtils:  internal_auth.NewGCPUtils(),
	}
	return &t
}

func NewFirebaseAuthClientFromGCP(ctx context.Context, gcpApp *gcp.App) (TenantClient, error) {
	authClient, err := gcpApp.Auth(ctx)
	if err != nil {
		return nil, err
	}

	tenantClient := defaultTenantClient(authClient)
	tenantClient.HashConfig = gcpApp.ProjectConfig.SignIn.HashConfig

	return tenantClient, nil
}

// TenantID returns an id of the tenant that current instance interacts with
func (tc *tenantClient) TenantID() string {
	return tc.tenantID
}

// GetHashConfig returns a scrypt hash config of the tenant that current instance interacts with
// Required permission
func (tc *tenantClient) GetHashConfig() *gcp.HashConfig {
	if tc.HashConfig == nil {
		return &gcp.HashConfig{}
	}
	return tc.HashConfig
}

// GetUser return a User has uid
// Return ErrUserNotFound if user not found
// Return ErrUserUIDEmpty if uid is empty
func (tc *tenantClient) GetUser(ctx context.Context, uid string) (internal_user.User, error) {
	if uid == "" {
		return nil, internal_user.ErrUserUIDEmpty
	}

	userRecord, err := tc.gcpClient.GetUser(ctx, uid)
	if err != nil {
		if tc.gcpUtils.IsUserNotFound(err) {
			err = internal_user.ErrUserNotFound
		}
		return nil, err
	}

	return internal_auth.NewUserFromGCPUserRecord(userRecord), nil
}

func (tc *tenantClient) CreateUser(ctx context.Context, user internal_user.User, userFieldsToCreate ...internal_user.UserField) error {
	if user == nil {
		return internal_user.ErrUserIsNil
	}

	if len(userFieldsToCreate) < 1 {
		userFieldsToCreate = internal_user.DefaultUserFieldsToCreate
	}

	userToCreate, err := internal_auth.ToGCPUsersToCreate(user, userFieldsToCreate...)
	if err != nil {
		return err
	}

	_, err = tc.gcpClient.CreateUser(ctx, userToCreate)
	if err != nil {
		return err
	}

	return nil
}

// LegacyImportUsers import auth users with firebase library interface
// Use for backward compatible
func (tc *tenantClient) LegacyImportUsers(ctx context.Context, users []*auth.UserToImport, opts ...auth.UserImportOption) (*auth.UserImportResult, error) {
	return tc.gcpClient.ImportUsers(ctx, users, opts...)
}

// ImportUsers imports all users into auth system
// Beware: This import action works like an upsert action
func (tc *tenantClient) ImportUsers(ctx context.Context, users internal_user.Users, importHash internal_auth.ScryptHash) (*internal_user.ImportUsersResult, error) {
	if len(users) < 1 {
		return nil, internal_user.ErrUserListEmpty
	}

	usersToImport, convertErr := internal_auth.ToGCPUsersToImport(users)
	if convertErr != nil {
		return nil, errors.Wrap(convertErr, "toGCPUsersToImport")
	}

	var usersImportResult *auth.UserImportResult
	var err error
	if importHash == nil {
		usersImportResult, err = tc.gcpClient.ImportUsers(ctx, usersToImport)
	} else {
		if err := internal_auth.IsScryptHashValid(importHash); err != nil {
			return nil, err
		}

		scryptHash := hash.Scrypt{
			Key:           importHash.Key(),
			SaltSeparator: importHash.SaltSeparator(),
			Rounds:        importHash.Rounds(),
			MemoryCost:    importHash.MemoryCost(),
		}

		usersImportResult, err = tc.gcpClient.ImportUsers(ctx, usersToImport, auth.WithHash(scryptHash))
	}

	if err != nil {
		return nil, errors.Wrap(err, "ImportUsers")
	}

	userIndexWithErr := make(map[int]string, usersImportResult.FailureCount)
	for _, err := range usersImportResult.Errors {
		userIndexWithErr[err.Index] = err.Reason
	}

	usersSuccessToImport := make([]internal_user.User, 0, usersImportResult.SuccessCount)
	usersFailedToImport := make(internal_user.UsersFailedToImport, 0, usersImportResult.FailureCount)
	for i, user := range users {
		err, hasErr := userIndexWithErr[i]
		if hasErr {
			userFailedToImport := &internal_user.UserFailedToImport{
				User: user,
				Err:  err,
			}
			usersFailedToImport = append(usersFailedToImport, userFailedToImport)
		} else {
			usersSuccessToImport = append(usersSuccessToImport, user)
		}
	}

	result := &internal_user.ImportUsersResult{
		TenantID:            tc.tenantID,
		UsersSuccessImport:  usersSuccessToImport,
		UsersFailedToImport: usersFailedToImport,
	}
	return result, nil
}

// LegacyUpdateUser update an auth user with firebase library interface
// Use for backward compatible
func (tc *tenantClient) LegacyUpdateUser(ctx context.Context, uid string, userToUpdate *auth.UserToUpdate) (ur *auth.UserRecord, err error) {
	return tc.gcpClient.UpdateUser(ctx, uid, userToUpdate)
}

// UpdateUser update a user auth info
// If updateFields is empty, this func will update all user's fields
/*func (tc *tenantClient) UpdateUser(ctx context.Context, user internal_user.User, updateFields ...internal_user.UserField) (internal_user.User, error) {
	if user == nil {
		return nil, internal_user.ErrUserIsNil
	}
	if len(updateFields) < 1 {
		updateFields = internal_user.DefaultUserFieldsToUpdate
	}

	if err := internal_user.IsUserInfoValid(user, updateFields...); err != nil {
		return nil, err
	}

	updateInfo := &auth.UserToUpdate{}

	for _, updateField := range updateFields {
		switch updateField {
		case internal_user.UserFieldEmail:
			updateInfo = updateInfo.Email(user.GetEmail())
		case internal_user.UserFieldDisplayName:
			updateInfo = updateInfo.DisplayName(user.GetDisplayName())
		case internal_user.UserFieldPhoneNumber:
			updateInfo = updateInfo.PhoneNumber(user.GetPhoneNumber())
		case internal_user.UserFieldPhotoURL:
			updateInfo = updateInfo.PhotoURL(user.GetPhotoURL())
		case internal_user.UserFieldCustomClaims:
			updateInfo = updateInfo.CustomClaims(user.GetCustomClaims())
		case internal_user.UserFieldRawPassword:
			updateInfo = updateInfo.Password(user.GetRawPassword())
		}
	}

	result, err := tc.gcpClient.UpdateUser(ctx, user.GetUID(), updateInfo)
	if err != nil {
		return nil, err
	}

	return internal_auth.NewUserFromGCPUserRecord(result), nil
}*/

func (tc *tenantClient) CustomToken(ctx context.Context, uid string) (string, error) {
	return tc.gcpClient.CustomToken(ctx, uid)
}

type Pager struct {
	gcpPager internal_auth.GCPPager
	pageSize int
}

func (p *Pager) NextPage() ([]internal_user.User, string, error) {
	exportedUsers := make([]*auth.ExportedUserRecord, 0, p.pageSize)

	nextPageToken, err := p.gcpPager.NextPage(&exportedUsers)
	if err != nil {
		return nil, "", err
	}

	users := make([]internal_user.User, 0, p.pageSize)
	for _, exportedUser := range exportedUsers {
		passwordHash, err := base64.URLEncoding.DecodeString(exportedUser.PasswordHash)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to decode base 64 password hash")
		}
		passwordSalt, err := base64.URLEncoding.DecodeString(exportedUser.PasswordSalt)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to decode base 64 password salt")
		}

		user := internal_user.NewUser(
			internal_user.WithUID(exportedUser.UID),
			internal_user.WithEmail(exportedUser.Email),
			internal_user.WithPhoneNumber(exportedUser.PhoneNumber),
			internal_user.WithPhotoURL(exportedUser.PhotoURL),
			internal_user.WithDisplayName(exportedUser.DisplayName),
			internal_user.WithCustomClaims(exportedUser.CustomClaims),
			internal_user.WithPasswordHash(passwordHash),
			internal_user.WithPasswordSalt(passwordSalt),
		)
		users = append(users, user)
	}

	return users, nextPageToken, nil
}

func (tc *tenantClient) UserPager(ctx context.Context, nextPageToken string, pageSize int) *Pager {
	gcpPager := iterator.NewPager(tc.gcpClient.Users(ctx, nextPageToken), pageSize, "")
	pager := &Pager{
		gcpPager: gcpPager,
	}
	return pager
}

func (tc *tenantClient) IterateAllUsers(ctx context.Context, pageSize int, iteratedUsersCallback func(users internal_user.Users) error) error {
	switch {
	case pageSize < 1:
		return errors.New("page size must be >= 1")
	}

	if pageSize > 1000 {
		pageSize = 1000
	}

	pager := tc.UserPager(ctx, "", pageSize)

	for {
		users, nextPageToken, err := pager.NextPage()
		if err != nil {
			return err
		}

		if err := iteratedUsersCallback(users); err != nil {
			return err
		}

		if nextPageToken == "" {
			break
		}
	}

	return nil
}

// ImportUsersParams contains params to run import process
// users will be imported from src tenant to destination tenant
type ImportUsersParams struct {
	srcTenantClient  TenantClient
	srcScryptHash    internal_auth.ScryptHash
	srcNextPageToken string
	srcPageSize      int
	destTenantClient TenantClient
}

func ImportUsersBetweenTenants(ctx context.Context, params *ImportUsersParams, usersImportedCallback func(result *internal_user.ImportUsersResult)) error {
	switch {
	case params == nil:
		return errors.New("params is nil")
	case params.srcPageSize < 1:
		return errors.New("page size must be >= 1")
	}

	srcPager := params.srcTenantClient.UserPager(ctx, params.srcNextPageToken, params.srcPageSize)

	for {
		users, nextPageToken, err := srcPager.NextPage()
		if err != nil {
			return err
		}

		if nextPageToken == "" {
			break
		}

		result, err := params.destTenantClient.ImportUsers(ctx, users, params.srcScryptHash)
		if err != nil {
			return err
		}

		if usersImportedCallback == nil {
			continue
		}

		usersImportedCallback(result)
	}
	return nil
}

// SetHashConfig sets a custom hash config instead of fetching tenant config from gcloud API. ONLY USER FOR TESTING FOR NOW
func (tc *tenantClient) SetHashConfig(hashConfig *gcp.HashConfig) {
	tc.HashConfig = hashConfig
}

const (
	langCodeURLKey = "lang"
)

// PasswordResetLink returns a password reset link for a user, currently sdk for golang does not support change language of reset password link
// so we have to manually update langCode of the link
func (tc *tenantClient) PasswordResetLink(ctx context.Context, email, langCode string) (string, error) {
	// check if langCode is valid
	if _, err := language.Parse(langCode); err != nil {
		return "", internal_user.ErrInvalidLangCode
	}

	// get reset password link from gcp sdk
	resetLink, err := tc.gcpClient.PasswordResetLink(ctx, email)
	if err != nil {
		return "", err
	}

	// change lang of the reset password link
	parsedURL, err := url.Parse(resetLink)
	if err != nil {
		return "", err
	}
	query := parsedURL.Query()
	query.Set(langCodeURLKey, langCode)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}
