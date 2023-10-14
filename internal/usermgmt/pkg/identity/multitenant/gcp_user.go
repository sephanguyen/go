package multitenant

import (
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"firebase.google.com/go/v4/auth"
)

func NewUserFromGCPUserRecord(userRecord *auth.UserRecord) User {
	usr := &user{
		uid:          userRecord.UserInfo.UID,
		email:        userRecord.UserInfo.Email,
		phoneNumber:  userRecord.UserInfo.PhoneNumber,
		photoURL:     userRecord.UserInfo.PhotoURL,
		displayName:  userRecord.UserInfo.DisplayName,
		customClaims: userRecord.CustomClaims,
	}
	return usr
}

func ToGCPUsersToCreate(user User, userFieldsToCreate ...UserField) (*auth.UserToCreate, error) {
	if len(userFieldsToCreate) < 1 {
		userFieldsToCreate = DefaultUserFieldsToImport
	}

	if err := IsUserInfoValid(user, userFieldsToCreate...); err != nil {
		return nil, err
	}

	userToCreate := new(auth.UserToCreate).UID(user.UserID().String())

	for _, userFieldToCreate := range userFieldsToCreate {
		switch userFieldToCreate {
		case UserFieldEmail:
			userToCreate = userToCreate.Email(user.Email().String())
		case UserFieldDisplayName:
			userToCreate = userToCreate.DisplayName(user.FullName().String())
		case UserFieldPhoneNumber:
			userToCreate = userToCreate.PhoneNumber(user.PhoneNumber().String())
		case UserFieldPhotoURL:
			userToCreate = userToCreate.PhotoURL(user.Avatar().String())
		case UserFieldRawPassword:
			userToCreate = userToCreate.Password(user.Password().String())
		}
	}

	return userToCreate, nil
}

func ToGCPUsersToImport(users Users, hashConfig ScryptHash, userFieldsToImport ...UserField) ([]*auth.UserToImport, error) {
	if len(userFieldsToImport) < 1 {
		userFieldsToImport = DefaultUserFieldsToImport
	}

	usersToImport := make([]*auth.UserToImport, 0, len(users))

	for _, user := range users {
		err := IsUserInfoValid(user, userFieldsToImport...)
		if err != nil {
			return nil, err
		}

		userToImport := new(auth.UserToImport).UID(user.UserID().String())

		var passwordSalt, passwordHash []byte
		if hashConfig != nil {
			passwordSalt = []byte(idutil.ULIDNow())
			passwordHash, err = HashedPassword(hashConfig, []byte(user.Password().String()), passwordSalt)
			if err != nil {
				return nil, err
			}
		}

		for _, userFieldToImport := range userFieldsToImport {
			switch userFieldToImport {
			case UserFieldEmail:
				userToImport = userToImport.Email(user.Email().String())
			case UserFieldDisplayName:
				if user.FullName().String() == "" {
					continue
				}
				userToImport = userToImport.DisplayName(user.FullName().String())
			case UserFieldPhoneNumber:
				// Identity platform don't allow to input an empty string
				// Ignore import phone number if phone number is empty
				if user.PhoneNumber().String() == "" {
					continue
				}
				userToImport = userToImport.PhoneNumber(user.PhoneNumber().String())
			case UserFieldPhotoURL:
				if user.Avatar().String() == "" {
					continue
				}
				userToImport = userToImport.PhotoURL(user.Avatar().String())
				/*case UserFieldCustomClaims:
				userToImport = userToImport.CustomClaims(user.GetCustomClaims())*/
			case UserFieldPasswordHash:
				userToImport = userToImport.PasswordHash(passwordHash)
			case UserFieldPasswordSalt:
				userToImport = userToImport.PasswordSalt(passwordSalt)
			}
		}

		usersToImport = append(usersToImport, userToImport)
	}
	return usersToImport, nil
}
