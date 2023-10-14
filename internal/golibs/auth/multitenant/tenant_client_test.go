package multitenant

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	internal_auth "github.com/manabie-com/backend/internal/golibs/auth"
	internal_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	mocks "github.com/manabie-com/backend/mock/golibs/auth"

	"firebase.google.com/go/v4/auth"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func anUserWithValidInfo() internal_user.User {
	random := rand.Intn(12345678)
	user := internal_user.NewUser(
		internal_user.WithUID(fmt.Sprintf("uid-%v", random)),
		internal_user.WithEmail(fmt.Sprintf("email-%v@example.com", random)),
		internal_user.WithPhoneNumber(fmt.Sprintf("+81%v", 1000000000+random)),
		internal_user.WithPhotoURL(fmt.Sprintf("photoURL-%v", random)),
		internal_user.WithDisplayName(fmt.Sprintf("displayName-%v", random)),
		internal_user.WithCustomClaims(map[string]interface{}{
			"external-info": "example-info",
		}),
		internal_user.WithRawPassword(fmt.Sprintf("rawPassword-%v", random)),
	)
	return user
}

func aValidUserRecord() *auth.UserRecord {
	random := rand.Intn(12313)
	userRecord := &auth.UserRecord{
		UserInfo: &auth.UserInfo{
			DisplayName: fmt.Sprintf("displayName-%v", random),
			Email:       fmt.Sprintf("email-%v@example.com", random),
			PhoneNumber: fmt.Sprintf("phoneNumber-%v", random),
			PhotoURL:    fmt.Sprintf("photoURL-%v", random),
			UID:         fmt.Sprintf("uid-%v", random),
		},
		CustomClaims: nil,
	}
	return userRecord
}

func aMockTenantClient() *tenantClient {
	tc := defaultTenantClient(new(mocks.GCPTenantClient))
	tc.gcpUtils = new(mocks.GCPUtils)
	return tc
}

func TestNewTenantClient(t *testing.T) {
	t.Parallel()

	var tc *tenantClient

	testCases := []struct {
		name      string
		setupFunc func(ctx context.Context)
	}{
		{
			name: "init tenant client",
			setupFunc: func(ctx context.Context) {
				tc = defaultTenantClient(new(mocks.GCPTenantClient))
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			assert.NotNil(t, tc.gcpClient)
		})
	}
}

func TestTenantClient_GetUser(t *testing.T) {
	t.Parallel()

	validUserRecord := aValidUserRecord()

	var tc *tenantClient

	testCases := []struct {
		name         string
		inputUserID  string
		expectedUser internal_user.User
		expectedErr  error
		setupFunc    func(ctx context.Context)
	}{
		{
			name:         "get user with empty user id",
			inputUserID:  "",
			expectedUser: nil,
			expectedErr:  internal_user.ErrUserUIDEmpty,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
			},
		},
		{
			name:         "get user with non-existing user id",
			inputUserID:  "nonExistingUserID",
			expectedUser: nil,
			expectedErr:  internal_user.ErrUserNotFound,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()

				tc.gcpClient.(*mocks.GCPTenantClient).On("GetUser", ctx, mock.Anything).Return(nil, errors.New("a error occurs"))
				tc.gcpUtils.(*mocks.GCPUtils).On("IsUserNotFound", mock.Anything).Return(true)
			},
		},
		{
			name:         "get user with valid user id",
			inputUserID:  "validUserID",
			expectedUser: internal_auth.NewUserFromGCPUserRecord(validUserRecord),
			expectedErr:  nil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()

				tc.gcpClient.(*mocks.GCPTenantClient).On("GetUser", ctx, mock.Anything).Return(validUserRecord, nil)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			actualUser, actualErr := tc.GetUser(ctx, testCase.inputUserID)

			assert.Equal(t, testCase.expectedErr, actualErr)
			assert.Equal(t, testCase.expectedUser, actualUser)
		})
	}
}

func TestTenantClient_CreateUser(t *testing.T) {
	t.Parallel()

	var tc *tenantClient

	testCases := []struct {
		name        string
		inputUser   internal_user.User
		expectedErr error
		setupFunc   func(ctx context.Context)
	}{
		{
			name:        "create user with nil user",
			inputUser:   nil,
			expectedErr: internal_user.ErrUserIsNil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
			},
		},
		{
			name:        "create user with valid user",
			inputUser:   anUserWithValidInfo(),
			expectedErr: nil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()

				tc.gcpClient.(*mocks.GCPTenantClient).On("CreateUser", ctx, mock.Anything).Once().Return(nil, nil)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			actualErr := tc.CreateUser(ctx, testCase.inputUser)

			assert.Equal(t, testCase.expectedErr, actualErr)
		})
	}
}

/*func TestTenantClient_UpdateUser(t *testing.T) {
	t.Parallel()

	userRecord := aValidUserRecord()

	var tc *tenantClient

	testCases := []struct {
		name           string
		inputUser      internal_user.User
		expectedResult internal_user.User
		expectedErr    error
		setupFunc      func(ctx context.Context)
	}{
		{
			name:           "update user with nil user",
			inputUser:      nil,
			expectedResult: nil,
			expectedErr:    internal_user.ErrUserIsNil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
			},
		},
		{
			name:           "update user with valid user",
			inputUser:      anUserWithValidInfo(),
			expectedResult: internal_auth.NewUserFromGCPUserRecord(userRecord),
			expectedErr:    nil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()

				tc.gcpClient.(*mocks.GCPTenantClient).On("UpdateUser", ctx, mock.Anything, mock.Anything).Once().Return(userRecord, nil)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			actualResult, actualErr := tc.UpdateUser(ctx, testCase.inputUser)

			assert.Equal(t, testCase.expectedErr, actualErr)
			assert.Equal(t, testCase.expectedResult, actualResult)
		})
	}
}*/

func TestTenantClient_ImportUsers(t *testing.T) {
	t.Parallel()

	userRecord := aValidUserRecord()

	var tc *tenantClient

	testCases := []struct {
		name           string
		inputUsers     internal_user.Users
		expectedResult internal_user.User
		expectedErr    error
		setupFunc      func(ctx context.Context)
	}{
		{
			name:           "import users with nil users",
			inputUsers:     nil,
			expectedResult: nil,
			expectedErr:    internal_user.ErrUserListEmpty,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
			},
		},
		{
			name:           "import users have valid info",
			inputUsers:     internal_user.Users{anUserWithValidInfo()},
			expectedResult: internal_auth.NewUserFromGCPUserRecord(userRecord),
			expectedErr:    nil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()

				result := &auth.UserImportResult{
					SuccessCount: 1,
					FailureCount: 0,
					Errors:       []*auth.ErrorInfo{},
				}

				tc.gcpClient.(*mocks.GCPTenantClient).On("ImportUsers", ctx, mock.Anything, mock.Anything).Once().Return(result, nil)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			actualResult, actualErr := tc.ImportUsers(ctx, testCase.inputUsers, nil)

			assert.Equal(t, testCase.expectedErr, actualErr)

			if testCase.expectedErr == nil {
				assert.Equal(t, len(testCase.inputUsers), len(actualResult.UsersSuccessImport))
			}
		})
	}
}

var (
	tmpSubDir    string
	credFilePath string
)

func init() {
	tmpDir := os.TempDir()
	tmpSubDir = filepath.Join(tmpDir, "manabie")
	credFilePath = filepath.Join(tmpSubDir, "service_credentials.json")
}

func setServiceCredential() error {
	// Note: This is secret info. Even though anyone with access to backend repo can get this info,
	// it should never be leaked outside.
	const bobServiceCredentialJSON = `{
  "type": "service_account",
  "project_id": "dev-manabie-online",
  "private_key_id": "b9429198d79997fda6d093bcebe81a7d7ab69938",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQC6EhNnLsCquIaL\nvTlBOU+O4q4r8ZHzLzXD7IqtPxq5uRm7I5H1eOxMOCQzzgWdTVIhvychuyvLC+Vq\nvOv3+RR4ZUejmLLJuFrDD2pWPr2Nvm1X1r57TTldIr1c3Qcf5ipSzQ2n6gX9Nt8D\nHEkg3svRIR4fC+NorFQYIRFU95+R54Ui5s5GhHYYZTT19oiJrm+QER57IfPZXYQr\n9QdBjBA6B0jr9wlyTLAjeFL4STEboxilfwMbSj80sQkar/KbR2hByinlf/LTgCwz\nskkBk+llb2c7nZKih2x5zpehAgekGzeoAB6QlLqM2mP6URSO+KdQP7pCCh0RQxol\nGMY4x1iTAgMBAAECggEAB4VdhWktXnkw7wsJ+mnvnk3pTltoU9UPrkisXk5TrTgf\nIyJP7wUhP/9w7ysfrPkIHdcVJNbk8UMc1dCnFRHbUvZ9C87LQz4RZRsFaFEG5mjR\nEKDceC1p6SrTTqKcfByYj1o8eBIMheym3QBSsGJxCJX3GrgnS/7TM1p60d1kdMg+\nDzw/j4b+mMuTzr7pkg7hbpdTXkroK0zf9t5SHZJi/WKP9oDwfHx6K62tfCNlu1Qi\nasqGOG6K2lse/RiTA4zqE5BgfaH8oyxaj1yD9SM7bTaWkUMfdLKiMGTfiPAWAzMa\n3pTfzrT9iaCWYh5yb5p4xqFsXC4aboSRGw+utS5E+QKBgQD+HWSzsMpcN0jwPRN2\nUIG0jGzYv6hHZlAD+pAacintjLJDZ5eMf/WFYgF1jRBWMTDN9JwlQcuOrLyKJ5vt\nh1P7c1CJko5VHZnYw94vDnosZWKMmry6Lc4Jn8uEi0SIPscs/K8QQv+Qbl5ynaGa\nVKTcggbGLoXwzKMg8qyJRg8wpwKBgQC7c3R/TkGUJkm6TVp86sGCVmOfvcSfMI9c\nEwWK7IUpXgikYyNg+fqFrdxrA7tU/7hhaJreVkfFOM8o0nS/AG+9jTL6GhK80C/M\nXH5BBpqBVIMpsdH9fIsbPQ/vP+l5R6X8ZZEv+W22MA+C7j332vPKsvNflBmFBTqi\nWnPMYB5KNQKBgBI5UWuBlkGexWBVQPwPMf4cxAGXXR4hvENMyODcpx0eJfqnhzrQ\nQm9aY/hmMXG8/V8H19rkKREGWk8eIBScy+0QjAoRtJtuEAZ3pYuCYkikzLiAsGA5\nwLj3+MR8qGGM/wO+618jLujQwX0+yMQkpd4ahRnZZEmso1ZNkQoXOCepAoGARt7t\n6rvhm2umcGOSlKwFIYwb+mc7EZzAduVSMSYfanZ8+fnphF6+0w/ayDMO/qH4SgvM\nkcc5N121JQ/8x8IYfSgHX/u/nddwWumVamxeugsD1B3A8P/HcDLz9VbKpOnr3bNg\n4yyAyGL/WldM4orLpZVm4noR8/L4Ki3cniaxDQkCgYA8RK6NjzOFlQ45Oc/R9w0L\njXfm/VdXxNgij87WTXKylS4Dmk/bhsViCCk2ZkE7Ruo7iq9uorBnvCtUaPKyhSSi\n6GIfmYo4GS2/E+usaKfVTru/5R6lQ3wwkHVpGNymVeu7XEqy5uI2hSIKj19V79hI\n63a8cwnIHu5JlRONCVYprg==\n-----END PRIVATE KEY-----\n",
  "client_email": "bootstrap@dev-manabie-online.iam.gserviceaccount.com",
  "client_id": "104103173896408079531",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/bootstrap%40dev-manabie-online.iam.gserviceaccount.com"
}`
	err := os.WriteFile(credFilePath, []byte(bobServiceCredentialJSON), 0666)
	if err != nil {
		return fmt.Errorf("failed to write to credential file: %s", err)
	}
	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credFilePath)
	if err != nil {
		return fmt.Errorf("failed to set GOOGLE_APPLICATION_CREDENTIALS env: %s", err)
	}
	return nil
}

func setConfigsAndSecrets() (err error, cleanup func()) {
	err = func() error {
		err := os.MkdirAll(tmpSubDir, 0777)
		if err != nil {
			return fmt.Errorf("os.MkdirAll: %s", err)
		}
		if err := setServiceCredential(); err != nil {
			return fmt.Errorf("setServiceCredential: %s", err)
		}
		return nil
	}()

	cleanup = func() {
		err := os.RemoveAll(tmpSubDir)
		if err != nil {
			log.Printf("could not clean up temp directory: %s", err)
		}
	}

	if err != nil {
		cleanup()
		return err, nil
	}
	return nil, cleanup
}

/*func TestIntegration_TenantClient_ImportUsers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Second)
	defer cancel()

	return

	err, cleanup := setConfigsAndSecrets()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	identityClient, err := NewTenantManagerFromCredentialFile(ctx, credFilePath)
	assert.NoError(t, err)

	srcScrypt, err := GetProjectScryptHashConfig(ctx, credFilePath)
	if err != nil {
		t.Fatal(err)
	}

	// Create src tenant
	srcTenant, err := identityClient.CreateTenant(ctx, &tenant{displayName: "src-test-import"})
	if err != nil {
		t.Fatal(err)
	}

	srcTenantClient, err := identityClient.TenantClient(srcTenant.GetID())
	if err != nil {
		t.Fatal(err)
	}

	// Create users in src tenant
	users := Users{
		anUserWithValidInfo(),
		anUserWithEmptyFields(UserFieldPhoneNumber),
		anUserWithEmptyFields(UserFieldDisplayName),
		anUserWithEmptyFields(UserFieldPhotoURL),
	}
	_, err = srcTenantClient.ImportUsers(ctx, users, srcScrypt)
	if err != nil {
		t.Fatal(err)
	}

	testUser := anUserWithValidInfo()
	testUser.uid = users[0].UID()
	testUser.customClaims = nil
	_, err = srcTenantClient.ImportUsers(ctx, Users{testUser}, srcScrypt)
	if err != nil {
		t.Fatal(err)
	}

	// Create dest tenant
	destTenant, err := identityClient.CreateTenant(ctx, &tenant{displayName: "dest-test-import"})
	if err != nil {
		t.Fatal(err)
	}
	destTenantClient, err := identityClient.TenantClient(destTenant.GetID())
	if err != nil {
		t.Fatal(err)
	}

	// Import users from src tenant to dest tenant
	srcUsers := make(Users, 0, 10)
	pager := srcTenantClient.UserPager(ctx, "", 2)
	for {
		users, nextPageToken, err := pager.NextPage()
		if err != nil {
			t.Fatal(err)
		}
		if nextPageToken == "" {
			break
		}

		srcUsers = append(srcUsers, users...)

		result, err := destTenantClient.ImportUsers(ctx, users, srcScrypt)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, 0, len(result.UsersFailedToImport))
	}

	// Assert users in src tenant and users in dest tenant
	srcIDAndUser := srcUsers.IDAndUserMap()
	destPager := destTenantClient.UserPager(ctx, "", 2)
	for {
		destUsers, nextPageToken, err := destPager.NextPage()
		if err != nil {
			t.Fatal(err)
		}
		if nextPageToken == "" {
			break
		}

		for _, destUser := range destUsers {
			srcUser, exist := srcIDAndUser[destUser.UID()]
			if destUser.UID() == testUser.UID() {
				fmt.Println(fmt.Printf("%+v", srcUser))
				fmt.Println(fmt.Printf("%+v", destUser))
			}

			assert.True(t, exist)
			assert.Equal(t, srcUser, destUser)
		}
	}

	// Cleanup test resources
	assert.NoError(t, identityClient.DeleteTenant(ctx, srcTenant.GetID()))
	assert.NoError(t, identityClient.DeleteTenant(ctx, destTenant.GetID()))
}*/

func Test_tenantClient_PasswordResetLink(t *testing.T) {
	t.Parallel()

	var tc *tenantClient

	testCases := []struct {
		name           string
		inputLangCode  string
		setupFunc      func(ctx context.Context)
		expectedResult string
		expectedErr    error
	}{
		{
			name:           "get password reset link with english lang code",
			inputLangCode:  "en",
			expectedResult: "https://firebase.net/resetpassword/?lang=en",
			expectedErr:    nil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
				tc.gcpClient.(*mocks.GCPTenantClient).
					On("PasswordResetLink", ctx, mock.Anything).
					Return("https://firebase.net/resetpassword/?lang=en", nil)
			},
		},
		{
			name:           "get password reset link with japan lang code",
			inputLangCode:  "ja",
			expectedResult: "https://firebase.net/resetpassword/?lang=ja",
			expectedErr:    nil,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
				tc.gcpClient.(*mocks.GCPTenantClient).
					On("PasswordResetLink", ctx, mock.Anything).
					Return("https://firebase.net/resetpassword/?lang=en", nil)
			},
		},
		{
			name:           "get password reset link with invalid lang code",
			inputLangCode:  "",
			expectedResult: "",
			expectedErr:    internal_user.ErrInvalidLangCode,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
			},
		},
		{
			name:           "get password reset link with error",
			inputLangCode:  "ja",
			expectedResult: "",
			expectedErr:    assert.AnError,
			setupFunc: func(ctx context.Context) {
				tc = aMockTenantClient()
				tc.gcpClient.(*mocks.GCPTenantClient).
					On("PasswordResetLink", ctx, mock.Anything).
					Return("", assert.AnError)
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)

			actualResult, actualErr := tc.PasswordResetLink(ctx, "email", testCase.inputLangCode)

			assert.Equal(t, testCase.expectedResult, actualResult)
			assert.Equal(t, testCase.expectedErr, actualErr)
		})
	}
}
