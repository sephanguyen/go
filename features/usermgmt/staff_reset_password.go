package usermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

var (
	mapLanguageWithCode = map[string]string{
		"english":  constant.EnglishLanguageCode,
		"japanese": constant.JapanLanguageCode,
	}
)

func (s *suite) initUsernameAndDomainNameByConditions(ctx context.Context, usernameCondition, domainNameCondition string) (string, string, error) {
	var (
		username   string
		domainName string
	)

	switch usernameCondition {
	case "unavailable username":
		username = "unavailable username"

	case "available username":
		signedInCtx, err := s.signedAsAccount(ctx, StaffRoleSchoolAdmin)
		if err != nil {
			return "", "", err
		}
		// create staff
		if signedInCtx, err = s.createUserGroupWithRole(signedInCtx, constant.RoleSchoolAdmin); err != nil {
			return "", "", err
		}
		user := s.Users[0]

		if err = waitUserSyncFromBobToAuthDB(signedInCtx, s.AuthPostgresDB, user.ID.String); err != nil {
			return "", "", err
		}

		isUserNameEnable, err := s.isUserNameEnable(signedInCtx)
		if err != nil {
			return "", "", errors.Wrap(err, "failed to check username enable")
		}

		username = user.UserName.String
		if !isUserNameEnable {
			username = user.Email.String
		}

	case "username from another org":
		username = "username_jprep"

	default:
		return "", "", fmt.Errorf("invalid usernameCondition: %s", usernameCondition)
	}

	switch domainNameCondition {
	case "available domain name":
		domainName = "manabie"

	case "unavailable domain name":
		domainName = "unavailable domain name"

	default:
		return "", "", fmt.Errorf("invalid domainNameCondition: %s", domainNameCondition)
	}

	return username, domainName, nil
}

func (s *suite) userResetPasswordWithLoginEmailAndDomainName(ctx context.Context, usernameCondition, domainNameCondition, langCondition string) (context.Context, error) {
	username, domainName, err := s.initUsernameAndDomainNameByConditions(ctx, usernameCondition, domainNameCondition)
	if err != nil {
		return ctx, errors.Wrap(err, "failed to init username and domain name")
	}

	langCode := mapLanguageWithCode[langCondition]
	req := &pb.ResetPasswordRequest{
		Username:     username,
		DomainName:   domainName,
		LanguageCode: langCode,
	}

	// use empty context to make sure that the request is sent by the user who is not signed in
	emptyContext := context.TODO()
	resp, err := pb.NewAuthServiceClient(s.UserMgmtConn).ResetPassword(emptyContext, req)

	s.Request = req
	s.RequestSentAt = time.Now()
	s.Response = resp
	s.ResponseErr = err

	return ctx, nil
}

func (s *suite) userReceivedEmailWithContent(ctx context.Context, langCondition string) (context.Context, error) {
	langCode := mapLanguageWithCode[langCondition]
	userEmail := s.Users[0].Email.String

	err := TryUntilSuccess(ctx, defaultInterval, func(ctx context.Context) (bool, error) {
		subjects := make([]string, 0)
		err := s.NotificationMgmtPostgresDB.QueryRow(ctx, `
				SELECT ARRAY_AGG(subject)
				FROM emails
				WHERE
					    $1 = ANY(email_recipients)
					AND DATE_TRUNC('minute'::TEXT, created_at) = DATE_TRUNC('minute'::TEXT, $2::TIMESTAMPTZ)
			`, userEmail, s.RequestSentAt).Scan(&subjects)
		if err != nil {
			return false, errors.Wrap(err, "cannot query email subjects")
		}

		expectedEmail := service.GenerateResetPasswordEmail("mock_email", "mock_reset_link", langCode)
		isExisting := golibs.InArrayString(expectedEmail.Subject, subjects)
		if !isExisting {
			return true, fmt.Errorf("cannot find email subject: %s", expectedEmail.Subject)
		}

		return false, nil
	})
	if err != nil {
		return ctx, errors.Wrap(err, "failed to check email content")
	}

	return ctx, nil
}

func waitUserSyncFromBobToAuthDB(ctx context.Context, db database.QueryExecer, userID string) error {
	return TryUntilSuccess(ctx, defaultInterval, func(ctx context.Context) (bool, error) {
		var count int
		err := db.
			QueryRow(ctx, `select count(*) from users where user_id = $1`, userID).
			Scan(&count)
		if err != nil {
			return false, errors.Wrap(err, "failed to query user")
		}
		if count == 0 {
			return true, fmt.Errorf("user was not created")
		}
		return false, nil
	})
}
