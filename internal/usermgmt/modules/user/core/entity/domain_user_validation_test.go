package entity

import (
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/stretchr/testify/assert"
)

const (
	validEmail     = "user@example.com"
	validFirstName = "example"
)

type userHasEmptyEmail struct {
	EmptyUser
}

func (mockUser *userHasEmptyEmail) Email() field.String {
	return field.NewString("")
}

type userHasInvalidEmail struct {
	EmptyUser
}

func (mockUser *userHasInvalidEmail) Email() field.String {
	return field.NewString("user@.com")
}

type userHasInvalidGender struct {
	EmptyUser
}

func (mockUser *userHasInvalidGender) Gender() field.String {
	return field.NewString("invalid gender")
}

type userHasEmptyFirstName struct {
	EmptyUser
}

func (mockUser *userHasEmptyFirstName) FirstName() field.String {
	return field.NewNullString()
}

type userHasEmptyLastName struct {
	EmptyUser
}

func (mockUser *userHasEmptyLastName) LastName() field.String {
	return field.NewNullString()
}

type validUser struct {
	EmptyUser
}

func (mockUser *validUser) Email() field.String {
	return field.NewString(validEmail)
}

func (mockUser *validUser) Gender() field.String {
	return field.NewString(upb.Gender_FEMALE.String())
}

func (mockUser *validUser) FirstName() field.String {
	return field.NewString(validFirstName)
}

func (mockUser *validUser) LastName() field.String {
	return field.NewString("last_name")
}

type userHasEmptyUserName struct {
	EmptyUser
}

type userHasInvalidUserName struct {
	EmptyUser
}

func (mockUser *userHasInvalidUserName) UserName() field.String {
	return field.NewString("firstname.familyname+u001")
}

type userHasValidUserNameWithUserNameFormat struct {
	EmptyUser
}

func (mockUser *userHasValidUserNameWithUserNameFormat) UserName() field.String {
	return field.NewString("username")
}

type userHasValidUserNameWithEmailFormat struct {
	EmptyUser
}

func (mockUser *userHasValidUserNameWithEmailFormat) UserName() field.String {
	return field.NewString("firstname.familyname+u001@manabie.com")
}

func TestDomainUser_validUserEmail(t *testing.T) {
	t.Run("user email is empty", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserEmail(&userHasEmptyEmail{})
		assert.Equal(t, MissingMandatoryFieldError{
			FieldName:  string(UserFieldEmail),
			EntityName: UserEntity,
			Index:      -1,
		}, err)
	})

	t.Run("user email is invalid", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserEmail(&userHasInvalidEmail{})
		assert.Equal(t, InvalidFieldError{
			FieldName:  string(UserFieldEmail),
			EntityName: UserEntity,
			Index:      -1,
			Reason:     NotMatchingPattern,
		}, err)
	})
}

func TestDomainUser_validUserGender(t *testing.T) {
	t.Run("user gender is invalid", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserGender(&userHasInvalidGender{})
		assert.Equal(t, InvalidFieldError{
			FieldName:  string(UserFieldGender),
			EntityName: UserEntity,
			Index:      -1,
			Reason:     NotMatchingEnum,
		}, err)
	})
}

func TestDomainUser_validUserFirstName(t *testing.T) {
	t.Run("user first_name is empty", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserFirstName(&userHasEmptyFirstName{})
		assert.Equal(t, MissingMandatoryFieldError{
			FieldName:  string(UserFieldFirstName),
			EntityName: UserEntity,
			Index:      -1,
		}, err)
	})
}

func TestDomainUser_validUserLastName(t *testing.T) {
	t.Run("user last_name is empty", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserLastName(&userHasEmptyLastName{})
		assert.Equal(t, MissingMandatoryFieldError{
			FieldName:  string(UserFieldLastName),
			EntityName: UserEntity,
			Index:      -1,
		}, err)
	})
}

func TestValidUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         []User
		expectedError error
	}{
		{
			name:          "user is valid",
			input:         []User{&validUser{}},
			expectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedError, ValidUser(false, testCase.input...))
		})
	}
}

type HasIndexTest struct {
	indexAttr int
}

func (h HasIndexTest) Index() int {
	return h.indexAttr
}

type NoHasIndexTest struct {
	indexAttr int
}

func Test_GetIndex(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		input         interface{}
		expectedIndex int
	}{
		{
			name:          "happy case: there is an index with have an impl method Index()",
			input:         HasIndexTest{indexAttr: 1},
			expectedIndex: 1,
		},
		{
			name:          "happy case: there is an index = 2 with have an impl method Index()",
			input:         HasIndexTest{indexAttr: 2},
			expectedIndex: 2,
		},
		{
			name:          "happy case: there is no index with have an impl method Index()",
			input:         HasIndexTest{},
			expectedIndex: 0,
		},
		{
			name:          "happy case: there is no index without have an impl method Index()",
			input:         NoHasIndexTest{},
			expectedIndex: -1,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			index := GetIndex(testCase.input)
			assert.Equal(t, testCase.expectedIndex, index)
		})
	}
}

func TestValidateUserName(t *testing.T) {
	t.Run("user username is empty", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserName(&userHasEmptyUserName{})
		assert.Equal(t, MissingMandatoryFieldError{
			FieldName:  string(UserFieldUserName),
			EntityName: UserEntity,
			Index:      -1,
		}, err)
	})

	t.Run("user username is invalid", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserName(&userHasInvalidUserName{})
		assert.Equal(t, InvalidFieldError{
			FieldName:  string(UserFieldUserName),
			EntityName: UserEntity,
			Index:      -1,
			Reason:     NotMatchingPattern,
		}, err)
	})

	t.Run("user username is valid with username format", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserName(&userHasValidUserNameWithUserNameFormat{})
		assert.Equal(t, nil, err)
	})

	t.Run("user username is valid with email format", func(t *testing.T) {
		t.Parallel()

		err := ValidateUserName(&userHasValidUserNameWithEmailFormat{})
		assert.Equal(t, nil, err)
	})
}
