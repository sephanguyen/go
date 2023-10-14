package service

import (
	"context"
	"regexp"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	helper "github.com/manabie-com/backend/internal/usermgmt/pkg/utils"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type UserRepo interface {
	GetByEmails(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error)
	GetByEmailsInsensitiveCase(ctx context.Context, db database.QueryExecer, emails []string) (entity.Users, error)
	GetByUserNames(ctx context.Context, db database.QueryExecer, usernames []string) (entity.Users, error)
	GetByIDs(ctx context.Context, db database.QueryExecer, userIDs []string) (entity.Users, error)
	GetByExternalUserIDs(ctx context.Context, db database.QueryExecer, externalUserIDs []string) (entity.Users, error)
}

func ValidateUserEmailsExistedInSystem(ctx context.Context, userRepo UserRepo, db database.QueryExecer, users entity.Users) error {
	zapLogger := ctxzap.Extract(ctx)
	emails := users.Emails()
	existingUsers, err := userRepo.GetByEmailsInsensitiveCase(ctx, db, emails)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "userRepo.GetByEmailsInsensitiveCase"),
			zap.Strings("existingUserIDs", existingUsers.UserIDs()),
		)
		return err
	}
	for _, user := range existingUsers {
		idx := helper.IndexOf(emails, user.Email().String())
		if user.UserID().String() == users[idx].UserID().String() {
			continue
		}
		// Only compare external id if its value is not empty
		if user.ExternalUserID().String() != "" && users[idx].ExternalUserID().String() != "" {
			if user.ExternalUserID().String() == users[idx].ExternalUserID().String() {
				continue
			}
		}
		zapLogger.Error(
			"cannot upsert users with emails existing in system",
			zap.String("Function", "ValidateUserEmailsExistedInSystem"),
			zap.Strings("existingUserIDs", existingUsers.UserIDs()),
		)
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldEmail),
			EntityName: entity.UserEntity,
			Index:      idx,
		}
	}

	return nil
}

func ValidateUserNamesExistedInSystem(ctx context.Context, userRepo UserRepo, db database.QueryExecer, users entity.Users) error {
	zapLogger := ctxzap.Extract(ctx)
	usernames := users.LowerCasedUserNames()
	existingUsers, err := userRepo.GetByUserNames(ctx, db, usernames)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "userRepo.GetByUserNames"),
			zap.Strings("existingUserIDs", existingUsers.UserIDs()),
		)
		return err
	}
	for _, existingUser := range existingUsers {
		existingUserName := strings.ToLower(existingUser.UserName().String())
		idx := helper.IndexOf(usernames, existingUserName)
		if existingUser.UserID().String() == users[idx].UserID().String() {
			// if user id is the same, so we should skip this case
			continue
		}

		zapLogger.Error(
			"cannot upsert users with username existing in system",
			zap.String("Function", "ValidateUserNamesExistedInSystem"),
			zap.Strings("existingUserIDs", existingUsers.UserIDs()),
		)
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldUserName),
			EntityName: entity.UserEntity,
			Index:      idx,
		}
	}

	return nil
}

func getDuplicatedIndex(values []string) int {
	visitedFields := map[string]struct{}{}
	for i, value := range values {
		if value != "" {
			if _, ok := visitedFields[value]; ok {
				return i
			}
			visitedFields[value] = struct{}{}
		}
	}
	return -1
}

func ValidateUserDuplicatedFields(users entity.Users) error {
	userIDs := users.UserIDs()
	if index := getDuplicatedIndex(userIDs); index > -1 {
		e := entity.DuplicatedFieldError{
			EntityName:      entity.UserEntity,
			DuplicatedField: string(entity.UserFieldUserID),
			Index:           index,
		}
		return e
	}

	emails := users.LowerCaseEmails()
	if index := getDuplicatedIndex(emails); index > -1 {
		e := entity.DuplicatedFieldError{
			EntityName:      entity.UserEntity,
			DuplicatedField: string(entity.UserFieldEmail),
			Index:           index,
		}
		return e
	}

	usernames := users.LowerCasedUserNames()
	if index := getDuplicatedIndex(usernames); index > -1 {
		e := entity.DuplicatedFieldError{
			EntityName:      entity.UserEntity,
			DuplicatedField: string(entity.UserFieldUserName),
			Index:           index,
		}
		return e
	}

	externalUserIDs := users.ExternalUserIDs()
	if index := getDuplicatedIndex(externalUserIDs); index > -1 {
		e := entity.DuplicatedFieldError{
			EntityName:      entity.UserEntity,
			DuplicatedField: string(entity.UserFieldExternalUserID),
			Index:           index,
		}
		return e
	}

	return nil
}

func ValidateUserPhoneNumbers(userPhoneNumbers entity.DomainUserPhoneNumbers, idx int) error {
	phoneNumbers := userPhoneNumbers.PhoneNumbers()
	pattern := regexp.MustCompile(constant.PhoneNumberPattern)
	fieldName := "phone_number"

	for _, phoneNumber := range userPhoneNumbers {
		if phoneNumber.PhoneNumber().IsEmpty() {
			continue
		}

		switch phoneNumber.Type().String() {
		case constant.StudentPhoneNumber:
			fieldName = "student_phone_number"
		case constant.StudentHomePhoneNumber:
			fieldName = "home_phone_number"
		case constant.ParentPrimaryPhoneNumber:
			fieldName = "primary_phone_number"
		case constant.ParentSecondaryPhoneNumber:
			fieldName = "secondary_phone_number"
		}
		if ok := pattern.MatchString(phoneNumber.PhoneNumber().String()); !ok {
			return entity.InvalidFieldError{
				EntityName: entity.UserEntity,
				FieldName:  fieldName,
				Index:      idx,
			}
		}
	}

	if index := getDuplicatedIndex(phoneNumbers); index > -1 {
		if strings.Contains(userPhoneNumbers[0].Type().String(), "STUDENT") {
			fieldName = "home_phone_number"
		} else {
			fieldName = "secondary_phone_number"
		}
		return entity.DuplicatedFieldError{
			EntityName:      entity.UserEntity,
			DuplicatedField: fieldName,
			Index:           idx,
		}
	}

	return nil
}
