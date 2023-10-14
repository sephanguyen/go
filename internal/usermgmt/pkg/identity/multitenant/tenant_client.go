package multitenant

import (
	"context"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/gcp"

	"encoding/base64"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/auth/hash"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
)

type TenantClient interface {
	TenantID() string
	GetHashConfig() *gcp.HashConfig
	UserPager(ctx context.Context, nextPageToken string, pageSize int) *Pager
	GetUser(ctx context.Context, uid string) (User, error)
	LegacyUpdateUser(ctx context.Context, uid string, user *auth.UserToUpdate) (*auth.UserRecord, error)
	ImportUsers(ctx context.Context, users Users, importHash ScryptHash) (*ImportUsersResult, error)
	IterateAllUsers(ctx context.Context, pageSize int, iteratedUsersCallback func(users Users) error) error
	CustomToken(ctx context.Context, uid string) (string, error)

	//SetHashConfig sets a custom hash config instead of fetching from tenant config. ONLY USER FOR TESTING FOR NOW
	SetHashConfig(hashConfig *gcp.HashConfig)
}

type tenantClient struct {
	tenantID   string
	HashConfig *gcp.HashConfig

	gcpClient GCPTenantClient
	gcpUtils  GCPUtils
}

func defaultTenantClient(gcpTenantClient GCPTenantClient) *tenantClient {
	t := tenantClient{
		gcpClient: gcpTenantClient,
		gcpUtils:  NewGCPUtils(),
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
func (tc *tenantClient) GetUser(ctx context.Context, uid string) (User, error) {
	if uid == "" {
		return nil, ErrUserUIDEmpty
	}

	userRecord, err := tc.gcpClient.GetUser(ctx, uid)
	if err != nil {
		if tc.gcpUtils.IsUserNotFound(err) {
			err = ErrUserNotFound
		}
		return nil, err
	}

	return NewUserFromGCPUserRecord(userRecord), nil
}

func (tc *tenantClient) CreateUser(ctx context.Context, user User, userFieldsToCreate ...UserField) error {
	if user == nil {
		return ErrUserIsNil
	}

	if len(userFieldsToCreate) < 1 {
		userFieldsToCreate = DefaultUserFieldsToCreate
	}

	userToCreate, err := ToGCPUsersToCreate(user, userFieldsToCreate...)
	if err != nil {
		return err
	}

	_, err = tc.gcpClient.CreateUser(ctx, userToCreate)
	if err != nil {
		return err
	}

	return nil
}

/*
// LegacyImportUsers import auth users with firebase library interface
// Use for backward compatible
func (tc *tenantClient) LegacyImportUsers(ctx context.Context, users []*auth.UserToImport, opts ...auth.UserImportOption) (*auth.UserImportResult, error) {
	return tc.gcpClient.ImportUsers(ctx, users, opts...)
}*/

// ImportUsers imports all users into auth system
// Beware: This import action works like an upsert action
func (tc *tenantClient) ImportUsers(ctx context.Context, users Users, importHash ScryptHash) (*ImportUsersResult, error) {
	if len(users) < 1 {
		return nil, ErrUserListEmpty
	}

	usersToImport, convertErr := ToGCPUsersToImport(users, importHash)
	if convertErr != nil {
		return nil, errors.Wrap(convertErr, "toGCPUsersToImport")
	}

	var usersImportResult *auth.UserImportResult
	var err error
	if importHash == nil {
		usersImportResult, err = tc.gcpClient.ImportUsers(ctx, usersToImport)
	} else {
		if err := IsScryptHashValid(importHash); err != nil {
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

	usersSuccessToImport := make([]User, 0, usersImportResult.SuccessCount)
	usersFailedToImport := make(UsersFailedToImport, 0, usersImportResult.FailureCount)
	for i, user := range users {
		err, hasErr := userIndexWithErr[i]
		if hasErr {
			userFailedToImport := &UserFailedToImport{
				User: user,
				Err:  err,
			}
			usersFailedToImport = append(usersFailedToImport, userFailedToImport)
		} else {
			usersSuccessToImport = append(usersSuccessToImport, user)
		}
	}

	result := &ImportUsersResult{
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

func (tc *tenantClient) CustomToken(ctx context.Context, uid string) (string, error) {
	return tc.gcpClient.CustomToken(ctx, uid)
}

type Pager struct {
	gcpPager GCPPager
	pageSize int
}

func (p *Pager) NextPage() ([]User, string, error) {
	exportedUsers := make([]*auth.ExportedUserRecord, 0, p.pageSize)

	nextPageToken, err := p.gcpPager.NextPage(&exportedUsers)
	if err != nil {
		return nil, "", err
	}

	users := make([]User, 0, p.pageSize)
	for _, exportedUser := range exportedUsers {
		passwordHash, err := base64.URLEncoding.DecodeString(exportedUser.PasswordHash)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to decode base 64 password hash")
		}
		passwordSalt, err := base64.URLEncoding.DecodeString(exportedUser.PasswordSalt)
		if err != nil {
			return nil, "", errors.Wrap(err, "failed to decode base 64 password salt")
		}

		user := &user{
			uid:          exportedUser.UID,
			email:        exportedUser.Email,
			phoneNumber:  exportedUser.PhoneNumber,
			photoURL:     exportedUser.PhotoURL,
			displayName:  exportedUser.DisplayName,
			customClaims: exportedUser.CustomClaims,
			passwordHash: passwordHash,
			passwordSalt: passwordSalt,
		}
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

func (tc *tenantClient) IterateAllUsers(ctx context.Context, pageSize int, iteratedUsersCallback func(users Users) error) error {
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
	srcScryptHash    ScryptHash
	srcNextPageToken string
	srcPageSize      int
	destTenantClient TenantClient
}

/*
//ImportUsersBetweenTenants used TESTING ONLY
func ImportUsersBetweenTenants(ctx context.Context, params *ImportUsersParams, usersImportedCallback func(result *ImportUsersResult)) error {
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
}*/

//SetHashConfig sets a custom hash config instead of fetching tenant config from gcloud API. ONLY USER FOR TESTING FOR NOW
func (tc *tenantClient) SetHashConfig(hashConfig *gcp.HashConfig) {
	tc.HashConfig = hashConfig
}
