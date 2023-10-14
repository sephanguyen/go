package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	db "github.com/manabie-com/backend/internal/usermgmt/pkg/database"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	mock_multitenant "github.com/manabie-com/backend/mock/golibs/auth/multitenant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_repositories "github.com/manabie-com/backend/mock/usermgmt/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type ManabieOrg struct{}

func (org ManabieOrg) OrganizationID() field.String {
	return field.NewString(strconv.Itoa(constants.ManabieSchool))
}

func (org ManabieOrg) SchoolID() field.Int32 {
	return field.NewInt32(constants.ManabieSchool)
}

const validTenantID = "tenant-id"

var _ entity.User = (*ValidUser)(nil)

type ValidUser struct {
	entity.EmptyUser
	randomID string
	version  int
}

func (user *ValidUser) UserID() field.String {
	// Warning: ID must not be changed regardless version increasing/decreasing
	return field.NewString(user.randomID)
}
func (user *ValidUser) Email() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v@example.com", user.randomID, user.version))
}
func (user *ValidUser) Password() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v", user.randomID, user.version))
}
func (user *ValidUser) LoginEmail() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v@example.com", user.randomID, user.version))
}

type UserNeverChangeEmail struct {
	ValidUser
}

func (user *UserNeverChangeEmail) Email() field.String {
	return field.NewString(fmt.Sprintf("%v+v%v@example.com", user.randomID, 0))
}

func TestAuthUserUpserter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	type testCase struct {
		name                 string
		InitAuthUserUpserter func(db.Tx) (AuthUserUpserter, func(t *testing.T))
		inputTx              func() db.Tx
		inputOrg             entity.DomainOrganization
		inputUsers           func() entity.Users
		expectedErr          error
	}

	validUser := &ValidUser{
		randomID: idutil.ULIDNow(),
	}

	testCases := []testCase{
		{
			name: "happy case, user does not exist before",
			InitAuthUserUpserter: func(tx db.Tx) (AuthUserUpserter, func(t *testing.T)) {
				orgRepo := &mock_repositories.OrganizationRepo{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(validTenantID, nil)

				userRepo := &mock_repositories.MockDomainUserRepo{}
				userRepo.On("GetByIDs", ctx, tx, entity.Users{&ValidUser{}}.UserIDs()).Once().Return(entity.Users{}, nil)

				hashConfig := mockScryptHash()
				firebaseAuthClient := &mock_multitenant.TenantClient{}

				tenantClient := &mock_multitenant.TenantClient{}
				identityPlatformTenantManager := &mock_multitenant.TenantManager{}
				identityPlatformTenantManager.On("TenantClient", ctx, mock.Anything).Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)

				assertExpectationsForObjects := func(t *testing.T) {
					mock.AssertExpectationsForObjects(t, orgRepo, firebaseAuthClient, tenantClient, identityPlatformTenantManager)
				}

				return NewAuthUserUpserter(userRepo, orgRepo, firebaseAuthClient, identityPlatformTenantManager), assertExpectationsForObjects
			},
			inputTx: func() db.Tx {
				return &mock_database.Tx{}
			},
			inputOrg: ManabieOrg{},
			inputUsers: func() entity.Users {
				return entity.Users{&ValidUser{}}
			},
			expectedErr: nil,
		},
		{
			name: "happy case, user already exist then update with new profile that has password changed",
			InitAuthUserUpserter: func(tx db.Tx) (AuthUserUpserter, func(t *testing.T)) {
				userRepo := &mock_repositories.MockDomainUserRepo{}
				userRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{&UserNeverChangeEmail{*validUser}}, nil)

				orgRepo := &mock_repositories.OrganizationRepo{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(validTenantID, nil)

				hashConfig := mockScryptHash()
				firebaseAuthClient := &mock_multitenant.TenantClient{}

				tenantClient := &mock_multitenant.TenantClient{}
				identityPlatformTenantManager := &mock_multitenant.TenantManager{}
				identityPlatformTenantManager.On("TenantClient", ctx, mock.Anything).Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(&internal_auth_user.ImportUsersResult{}, nil)

				assertExpectationsForObjects := func(t *testing.T) {
					mock.AssertExpectationsForObjects(t, orgRepo, firebaseAuthClient, tenantClient, identityPlatformTenantManager)
				}

				return NewAuthUserUpserter(userRepo, orgRepo, firebaseAuthClient, identityPlatformTenantManager), assertExpectationsForObjects
			},
			inputTx: func() db.Tx {
				return &mock_database.Tx{}
			},
			inputOrg: ManabieOrg{},
			inputUsers: func() entity.Users {
				updateProfile := &UserNeverChangeEmail{
					ValidUser: *validUser,
				}
				updateProfile.version = 2
				return entity.Users{updateProfile}
			},
			expectedErr: nil,
		},
		{
			name: "cannot get tenant id by org id",
			InitAuthUserUpserter: func(tx db.Tx) (AuthUserUpserter, func(t *testing.T)) {
				userRepo := &mock_repositories.MockDomainUserRepo{}
				userRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{entity.EmptyUser{}}, nil)

				orgRepo := &mock_repositories.OrganizationRepo{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Return("", nil)

				assertExpectationsForObjects := func(t *testing.T) {
					mock.AssertExpectationsForObjects(t, orgRepo)
				}

				return NewAuthUserUpserter(userRepo, orgRepo, nil, nil), assertExpectationsForObjects
			},
			inputTx: func() db.Tx {
				return &mock_database.Tx{}
			},
			inputOrg: ManabieOrg{},
			inputUsers: func() entity.Users {
				return entity.Users{entity.EmptyUser{}}
			},
			expectedErr: errcode.ErrTenantOfOrgNotFound{OrganizationID: ManabieOrg{}.OrganizationID().String()},
		},
		{
			name: "cannot get tenant client to interact with tenant in identity platform",
			InitAuthUserUpserter: func(tx db.Tx) (AuthUserUpserter, func(t *testing.T)) {
				userRepo := &mock_repositories.MockDomainUserRepo{}
				userRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{&ValidUser{}}, nil)

				orgRepo := &mock_repositories.OrganizationRepo{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(validTenantID, nil)

				firebaseAuthClient := &mock_multitenant.TenantClient{}

				identityPlatformTenantManager := &mock_multitenant.TenantManager{}
				identityPlatformTenantManager.On("TenantClient", ctx, mock.Anything).Return(nil, internal_auth_user.ErrTenantNotFound)

				assertExpectationsForObjects := func(t *testing.T) {
					mock.AssertExpectationsForObjects(t, orgRepo, firebaseAuthClient, identityPlatformTenantManager)
				}

				return NewAuthUserUpserter(userRepo, orgRepo, firebaseAuthClient, identityPlatformTenantManager), assertExpectationsForObjects
			},
			inputTx: func() db.Tx {
				return &mock_database.Tx{}
			},
			inputOrg: ManabieOrg{},
			inputUsers: func() entity.Users {
				return entity.Users{entity.EmptyUser{}}
			},
			expectedErr: errcode.ErrIdentityPlatformTenantNotFound{TenantID: validTenantID},
		},
		{
			name: "cannot import to identity platform",
			InitAuthUserUpserter: func(tx db.Tx) (AuthUserUpserter, func(t *testing.T)) {
				userRepo := &mock_repositories.MockDomainUserRepo{}
				userRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{&ValidUser{}}, nil)

				orgRepo := &mock_repositories.OrganizationRepo{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(validTenantID, nil)

				hashConfig := mockScryptHash()
				firebaseAuthClient := &mock_multitenant.TenantClient{}

				tenantClient := &mock_multitenant.TenantClient{}
				identityPlatformTenantManager := &mock_multitenant.TenantManager{}
				identityPlatformTenantManager.On("TenantClient", ctx, mock.Anything).Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(nil, assert.AnError)
				tenantClient.On("TenantID").Return(validTenantID)

				assertExpectationsForObjects := func(t *testing.T) {
					mock.AssertExpectationsForObjects(t, orgRepo, firebaseAuthClient, tenantClient, identityPlatformTenantManager)
				}

				return NewAuthUserUpserter(userRepo, orgRepo, firebaseAuthClient, identityPlatformTenantManager), assertExpectationsForObjects
			},
			inputTx: func() db.Tx {
				return &mock_database.Tx{}
			},
			inputOrg: ManabieOrg{},
			inputUsers: func() entity.Users {
				return entity.Users{entity.EmptyUser{}}
			},
			expectedErr: errcode.ErrFailedToImportAuthUsersToTenantErr{
				Err:      assert.AnError,
				TenantID: validTenantID,
			},
		},
		{
			name: "cannot import to identity platform because one user profile is has issue",
			InitAuthUserUpserter: func(tx db.Tx) (AuthUserUpserter, func(t *testing.T)) {
				userRepo := &mock_repositories.MockDomainUserRepo{}
				userRepo.On("GetByIDs", ctx, tx, mock.Anything).Once().Return(entity.Users{&ValidUser{}}, nil)

				orgRepo := &mock_repositories.OrganizationRepo{}
				orgRepo.On("GetTenantIDByOrgID", ctx, tx, mock.Anything).Once().Return(validTenantID, nil)

				hashConfig := mockScryptHash()
				firebaseAuthClient := &mock_multitenant.TenantClient{}

				tenantClient := &mock_multitenant.TenantClient{}
				identityPlatformTenantManager := &mock_multitenant.TenantManager{}
				identityPlatformTenantManager.On("TenantClient", ctx, mock.Anything).Return(tenantClient, nil)
				tenantClient.On("GetHashConfig").Return(hashConfig)
				result := &internal_auth_user.ImportUsersResult{
					UsersFailedToImport: internal_auth_user.UsersFailedToImport{
						{
							User: internal_auth_user.NewUser(internal_auth_user.WithUID("example-uid"), internal_auth_user.WithEmail("example@manabie.com")),
							Err:  assert.AnError.Error(),
						},
					},
				}
				tenantClient.On("ImportUsers", ctx, mock.Anything, mock.Anything).Return(result, nil)
				tenantClient.On("TenantID").Return(validTenantID)

				assertExpectationsForObjects := func(t *testing.T) {
					mock.AssertExpectationsForObjects(t, orgRepo, firebaseAuthClient, tenantClient, identityPlatformTenantManager)
				}

				return NewAuthUserUpserter(userRepo, orgRepo, firebaseAuthClient, identityPlatformTenantManager), assertExpectationsForObjects
			},
			inputTx: func() db.Tx {
				return &mock_database.Tx{}
			},
			inputOrg: ManabieOrg{},
			inputUsers: func() entity.Users {
				return entity.Users{entity.EmptyUser{}}
			},
			expectedErr: errcode.ErrAuthProfilesHaveIssueWhenImport{
				ErrMessages: []string{fmt.Sprintf("{'%s' - '%s' : %s}", "example-uid", "example@manabie.com", assert.AnError.Error())},
				TenantID:    validTenantID,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			tx := testCase.inputTx()
			authUserUpserter, assertExpectationsForObjects := testCase.InitAuthUserUpserter(tx)
			_, err := authUserUpserter(ctx, tx, testCase.inputOrg, testCase.inputUsers(), nil, unleash.DomainUserFeatureOption{})
			assert.Equal(t, testCase.expectedErr, err)
			assertExpectationsForObjects(t)
		})
	}
}
