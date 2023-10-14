package entity

import (
	"regexp"
	"strings"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type HasIndex interface {
	Index() int
}

func GetIndex(entity interface{}) int {
	switch e := entity.(type) {
	case HasIndex:
		return e.Index()
	default:
		return -1
	}
}

func ValidUser(isEnableUsername bool, users ...User) error {
	for _, user := range users {
		if isEnableUsername {
			if err := ValidateUserName(user); err != nil {
				return err
			}
		}
		err := errorx.ReturnFirstErr(
			ValidateUserEmail(user),
			ValidateUserGender(user),
			ValidateUserFirstName(user),
			ValidateUserLastName(user),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func ValidateUserName(user User) error {
	index := GetIndex(user)
	if user.UserName().TrimSpace().IsEmpty() {
		return MissingMandatoryFieldError{
			Index:      index,
			FieldName:  string(UserFieldUserName),
			EntityName: UserEntity,
		}
	}

	matchedWithUserNamePattern, err := regexp.MatchString(constant.UsernamePattern, user.UserName().String())
	if err != nil {
		return InternalError{RawErr: err}
	}

	matchedWithEmailPattern, err := regexp.MatchString(constant.EmailPattern, user.UserName().String())
	if err != nil {
		return InternalError{RawErr: err}
	}

	// username can be email or username pattern
	isValidUserName := matchedWithUserNamePattern || matchedWithEmailPattern
	if !isValidUserName {
		return InvalidFieldError{
			EntityName: UserEntity,
			FieldName:  string(UserFieldUserName),
			Index:      index,
			Reason:     NotMatchingPattern,
		}
	}
	return nil
}

func ValidateUserEmail(user User) error {
	index := GetIndex(user)
	if strings.TrimSpace(user.Email().String()) == "" {
		return MissingMandatoryFieldError{
			EntityName: UserEntity,
			FieldName:  string(UserFieldEmail),
			Index:      index,
		}
	}
	if user.Email().String() != "" {
		emailPattern := constant.EmailPattern
		matched, err := regexp.MatchString(emailPattern, user.Email().String())
		if err != nil || !matched {
			return InvalidFieldError{
				EntityName: UserEntity,
				FieldName:  string(UserFieldEmail),
				Index:      index,
				Reason:     NotMatchingPattern,
			}
		}
	}

	return nil
}

func ValidateUserGender(user User) error {
	index := GetIndex(user)
	if !field.IsPresent(user.Gender()) {
		return nil
	}

	switch user.Gender().String() {
	case constant.UserGenderFemale, constant.UserGenderMale:
		return nil
	default:
		return InvalidFieldError{
			FieldName:  string(UserFieldGender),
			EntityName: UserEntity,
			Reason:     NotMatchingEnum,
			Index:      index,
		}
	}
}

func ValidateUserFirstName(user User) error {
	index := GetIndex(user)
	if strings.TrimSpace(user.FirstName().String()) == "" {
		return MissingMandatoryFieldError{
			FieldName:  string(UserFieldFirstName),
			EntityName: UserEntity,
			Index:      index,
		}
	}
	return nil
}

func ValidateUserLastName(user User) error {
	index := GetIndex(user)
	if strings.TrimSpace(user.LastName().String()) == "" {
		return MissingMandatoryFieldError{
			FieldName:  string(UserFieldLastName),
			EntityName: UserEntity,
			Index:      index,
		}
	}
	return nil
}
