package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aUserAccessPathRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user, err := s.insertUserToBob(ctx, bob_entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	location, err := s.insertLocationToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stmt := `INSERT INTO user_access_paths (
				user_id,
				location_id,
				created_at,
				updated_at)
			VALUES ($1, $2, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		user.ID,
		location.LocationID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "insert to bob user_access_paths")
	}

	stepState.CurrentUserID = user.ID.String
	stepState.LocationID = location.LocationID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserAccessPathRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			count(user_id)
		FROM
			user_access_paths
		WHERE
			user_id = $1
		AND
			location_id = $2
		`

	if err := try.Do(func(attempt int) (bool, error) {
		var count int
		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID, stepState.LocationID).Scan(&count)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if count == 1 {
			return false, nil
		}
		if count > 1 {
			return false, fmt.Errorf("unexpected number of user access path created on entryexitmgmt: %d", count)
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("user_access_path record not sync correctly on entryexitmgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
