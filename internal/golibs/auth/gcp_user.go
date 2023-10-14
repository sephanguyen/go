package auth

import (
	internal_user "github.com/manabie-com/backend/internal/golibs/auth/user"

	"firebase.google.com/go/v4/auth"
)

func NewUserFromGCPUserRecord(userRecord *auth.UserRecord) internal_user.User {
	user := internal_user.NewUser(
		internal_user.WithUID(userRecord.UserInfo.UID),
		internal_user.WithEmail(userRecord.UserInfo.Email),
		internal_user.WithPhoneNumber(userRecord.UserInfo.PhoneNumber),
		internal_user.WithPhotoURL(userRecord.UserInfo.PhotoURL),
		internal_user.WithDisplayName(userRecord.UserInfo.DisplayName),
		internal_user.WithCustomClaims(userRecord.CustomClaims),
	)
	return user
}

func ToGCPUsersToCreate(user internal_user.User, userFieldsToCreate ...internal_user.UserField) (*auth.UserToCreate, error) {
	if len(userFieldsToCreate) < 1 {
		userFieldsToCreate = internal_user.DefaultUserFieldsToImport
	}

	if err := internal_user.IsUserInfoValid(user, userFieldsToCreate...); err != nil {
		return nil, err
	}

	userToCreate := new(auth.UserToCreate).UID(user.GetUID())

	for _, userFieldToCreate := range userFieldsToCreate {
		switch userFieldToCreate {
		case internal_user.UserFieldEmail:
			userToCreate = userToCreate.Email(user.GetEmail())
		case internal_user.UserFieldDisplayName:
			userToCreate = userToCreate.DisplayName(user.GetDisplayName())
		case internal_user.UserFieldPhoneNumber:
			userToCreate = userToCreate.PhoneNumber(user.GetPhoneNumber())
		case internal_user.UserFieldPhotoURL:
			userToCreate = userToCreate.PhotoURL(user.GetPhotoURL())
		case internal_user.UserFieldRawPassword:
			userToCreate = userToCreate.Password(user.GetRawPassword())
		}
	}

	return userToCreate, nil
}

func ToGCPUsersToImport(users internal_user.Users, userFieldsToImport ...internal_user.UserField) ([]*auth.UserToImport, error) {
	if len(userFieldsToImport) < 1 {
		userFieldsToImport = internal_user.DefaultUserFieldsToImport
	}

	usersToImport := make([]*auth.UserToImport, 0, len(users))

	for _, user := range users {
		if err := internal_user.IsUserInfoValid(user, userFieldsToImport...); err != nil {
			return nil, err
		}

		userToImport := new(auth.UserToImport).UID(user.GetUID())

		var passwordHash, passwordSalt []byte

		for _, userFieldToImport := range userFieldsToImport {
			switch userFieldToImport {
			case internal_user.UserFieldEmail:
				userToImport = userToImport.Email(user.GetEmail())
			case internal_user.UserFieldDisplayName:
				if user.GetDisplayName() == "" {
					continue
				}
				userToImport = userToImport.DisplayName(user.GetDisplayName())
			case internal_user.UserFieldPhoneNumber:
				// Identity platform don't allow to input an empty string
				// Ignore import phone number if phone number is empty
				if user.GetPhoneNumber() == "" {
					continue
				}
				userToImport = userToImport.PhoneNumber(user.GetPhoneNumber())
			case internal_user.UserFieldPhotoURL:
				if user.GetPhotoURL() == "" {
					continue
				}
				userToImport = userToImport.PhotoURL(user.GetPhotoURL())
			case internal_user.UserFieldCustomClaims:
				userToImport = userToImport.CustomClaims(user.GetCustomClaims())
			case internal_user.UserFieldPasswordHash:
				passwordHash = user.GetPasswordHash()
			case internal_user.UserFieldPasswordSalt:
				passwordSalt = user.GetPasswordSalt()
			}
		}

		if len(passwordHash) > 0 && len(passwordSalt) > 0 {
			userToImport = userToImport.PasswordHash(user.GetPasswordHash()).PasswordSalt(user.GetPasswordSalt())
		}

		usersToImport = append(usersToImport, userToImport)
	}
	return usersToImport, nil
}
