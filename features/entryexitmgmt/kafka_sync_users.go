package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	entryexit_entities "github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aUserRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user, err := s.insertUserToBob(ctx, bob_entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentUserID = user.ID.String

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserRecordMustBeRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			user_id,
			country,
			name,
			given_name,
			device_token,
			allow_notification,
			user_group,
			resource_path
		FROM
			users
		WHERE user_id = $1
		`

	bobUser := &bob_entities.User{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID)
	err := bobRow.Scan(
		&bobUser.ID,
		&bobUser.Country,
		&bobUser.LastName,
		&bobUser.GivenName,
		&bobUser.DeviceToken,
		&bobUser.AllowNotification,
		&bobUser.Group,
		&bobUser.ResourcePath,
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		entryExitUser := &entryexit_entities.User{}
		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID).Scan(
			&entryExitUser.ID,
			&entryExitUser.Country,
			&entryExitUser.FullName,
			&entryExitUser.GivenName,
			&entryExitUser.DeviceToken,
			&entryExitUser.AllowNotification,
			&entryExitUser.Group,
			&entryExitUser.ResourcePath,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		// Compare the existing columns
		if bobUser.ID == entryExitUser.ID &&
			bobUser.Country == entryExitUser.Country &&
			bobUser.LastName == entryExitUser.FullName &&
			bobUser.GivenName == entryExitUser.GivenName &&
			bobUser.DeviceToken == entryExitUser.DeviceToken &&
			bobUser.AllowNotification == entryExitUser.AllowNotification &&
			bobUser.Group == entryExitUser.Group &&
			bobUser.ResourcePath == entryExitUser.ResourcePath {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("users record not sync correctly on entryexitmgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil

}
