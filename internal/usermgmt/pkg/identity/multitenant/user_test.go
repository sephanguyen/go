package multitenant

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func anUserWithValidInfo() *user {
	random := rand.Intn(12345678)
	user := &user{
		uid:         fmt.Sprintf("uid-%v", random),
		email:       fmt.Sprintf("email-%v@example.com", random),
		phoneNumber: fmt.Sprintf("+81%v", 1000000000+random),
		photoURL:    fmt.Sprintf("photoURL-%v", random),
		displayName: fmt.Sprintf("displayName-%v", random),
		customClaims: map[string]interface{}{
			"external-info": "example-info",
		},
		rawPassword:  fmt.Sprintf("rawPassword-%v", random),
		passwordHash: nil,
		passwordSalt: nil,
	}
	return user
}

func anUserWithEmptyFields(fields ...UserField) *user {
	user := anUserWithValidInfo()
	for _, field := range fields {
		switch field {
		case UserFieldUID:
			user.uid = ""
		case UserFieldEmail:
			user.email = ""
		case UserFieldDisplayName:
			user.displayName = ""
		case UserFieldPhoneNumber:
			user.phoneNumber = ""
		case UserFieldPhotoURL:
			user.photoURL = ""
		case UserFieldRawPassword:
			user.rawPassword = ""
		case UserFieldCustomClaims:
			user.customClaims = nil
		}
	}
	return user
}

func TestUser_UserImpl(t *testing.T) {
	t.Parallel()

	user := anUserWithValidInfo()

	assert.Equal(t, user.uid, user.UserID().String())
	assert.Equal(t, user.email, user.Email().String())
	assert.Equal(t, user.displayName, user.FullName().String())
	assert.Equal(t, user.phoneNumber, user.PhoneNumber().String())
	assert.Equal(t, user.photoURL, user.Avatar().String())
}

func TestUsersFailedToImport_SliceMethod(t *testing.T) {
	t.Parallel()

	users := []User{anUserWithValidInfo(), anUserWithValidInfo()}

	usersFailedToImport := UsersFailedToImport{
		{
			User: users[0],
		},
		{
			User: users[1],
		},
	}

	assert.Equal(t, usersFailedToImport.IDs(), []string{users[0].UserID().String(), users[1].UserID().String()})
	assert.Equal(t, usersFailedToImport.Emails(), []string{users[0].Email().String(), users[1].Email().String()})
}

func TestIsUserInfoValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		userToTest  User
		expectedErr error
		setupFunc   func(ctx context.Context)
	}{
		{
			name:        "user's uid is empty",
			userToTest:  anUserWithEmptyFields(UserFieldUID),
			expectedErr: ErrUserUIDEmpty,
			setupFunc:   func(ctx context.Context) {},
		},
		{
			name:        "user's email is empty",
			userToTest:  anUserWithEmptyFields(UserFieldEmail),
			expectedErr: ErrUserEmailEmpty,
			setupFunc:   func(ctx context.Context) {},
		},
		{
			name:        "user's raw password is empty",
			userToTest:  anUserWithEmptyFields(UserFieldRawPassword),
			expectedErr: ErrUserPasswordMinLength,
			setupFunc:   func(ctx context.Context) {},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setupFunc(ctx)
			actualErr := IsUserInfoValid(testCase.userToTest)
			assert.Equal(t, testCase.expectedErr, actualErr)
			assert.True(t, IsUserValidationErr(actualErr))
		})
	}
}
