package main

import (
	"fmt"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type UserField string

const (
	UserFieldUserID         UserField = "user_id"
	UserFieldEmail          UserField = "user_email"
	UserFieldOrganizationID UserField = "user_org_id"
)

//go:generate ./hexagen ent-impl --type=User ../../../../modules/user/core/valueobj .
type User interface {
	UserID() field.String
	Email() field.String

	valueobj.HasOrganizationID
}

// Test values
const (
	validEmailAddress = "example@manabie.com"
)

func main() {
	// ----------------------------------------------------------------------------------------------------------
	// Generated code in end-to-end test must fill missing code, pass all requirements and build successfully
	// ----------------------------------------------------------------------------------------------------------

	// Check generated null entity
	nullUser := &NullUser{}
	switch {
	case nullUser.UserID().Status() != field.StatusNull:
		panic("expected status is null")
	case nullUser.UserID().RawValue() != "":
		panic("expected raw value is empty")
	}

	// Data to test
	user1 := &user{
		userID: field.NewString("1"),
		email:  field.NewString(validEmailAddress),
	}
	user2 := &user{
		userID: field.NewString("2"),
		email:  field.NewString(validEmailAddress),
	}

	// Check generated comparison function
	if err := CompareUserValues(user1, user2); err == nil {
		panic("expected return error because ids are different")
	}

	// Check generated methods for slice
	users := Users{user1, user2}

	if len(users) != len(users.UserIDs()) {
		panic("expected len of users and len of their user ids are equal")
	}

	for i, userID := range users.UserIDs() {
		if userID.Status() != users[i].UserID().Status() {
			panic("expected attribute in slice equal with root entity")
		}
		if userID.RawValue() != users[i].UserID().RawValue() {
			panic("expected attribute in slice equal with root entity")
		}
	}

	existingUser := &user{userID: field.NewString("1"), email: field.NewString(validEmailAddress)}
	newUser := NewUser(
		UserFields.From(existingUser),
	)
	// Check generated comparison function
	if err := CompareUserValues(existingUser, newUser); err != nil {
		panic(fmt.Sprintf("expected two user has the same field values: %v", err))
	}
}
