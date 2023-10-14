package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aUserBasicInfoRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	user, err := s.insertUserToBob(ctx, bob_entities.UserGroupStudent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stmt := `INSERT INTO user_basic_info (
		user_id,
		name,
		first_name,
		last_name,
		full_name_phonetic,
		first_name_phonetic,
		last_name_phonetic,
		current_grade,
		grade_id,
		email,
		updated_at,
		created_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
	`
	userID := user.ID.String
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		userID,
		fmt.Sprintf("name-%v", userID),
		fmt.Sprintf("first-name-%v", userID),
		fmt.Sprintf("last-name-%v", userID),
		fmt.Sprintf("full-name-phonetic-%v", userID),
		fmt.Sprintf("first-name-phonetic-%v", userID),
		fmt.Sprintf("last-name-phonetic-%v", userID),
		1,
		fmt.Sprintf("grade-id-%v", userID),
		fmt.Sprintf("email-%v", userID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "insert to bob user_basic_info")
	}

	stepState.CurrentUserID = user.ID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserBasicInfoMustRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			user_id,
			name,
			first_name,
			last_name,
			full_name_phonetic,
			first_name_phonetic,
			last_name_phonetic,
			current_grade,
			grade_id,
			email
		FROM
			user_basic_info
		WHERE
			user_id = $1
	`
	// Get the user basic info from bob DB
	bobUserBasicInfo := &entities.UserBasicInfo{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID)
	err := bobRow.Scan(
		&bobUserBasicInfo.UserID,
		&bobUserBasicInfo.Name,
		&bobUserBasicInfo.FirstName,
		&bobUserBasicInfo.LastName,
		&bobUserBasicInfo.FullNamePhonetic,
		&bobUserBasicInfo.FirstNamePhonetic,
		&bobUserBasicInfo.LastNamePhonetic,
		&bobUserBasicInfo.CurrentGrade,
		&bobUserBasicInfo.GradeID,
		&bobUserBasicInfo.Email,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {
		// Get the user basic info from entryexitmgmt DB
		entryexitUserBasicInfo := &entities.UserBasicInfo{}
		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID).Scan(
			&entryexitUserBasicInfo.UserID,
			&entryexitUserBasicInfo.Name,
			&entryexitUserBasicInfo.FirstName,
			&entryexitUserBasicInfo.LastName,
			&entryexitUserBasicInfo.FullNamePhonetic,
			&entryexitUserBasicInfo.FirstNamePhonetic,
			&entryexitUserBasicInfo.LastNamePhonetic,
			&entryexitUserBasicInfo.CurrentGrade,
			&entryexitUserBasicInfo.GradeID,
			&entryexitUserBasicInfo.Email,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if bobUserBasicInfo.UserID.String == entryexitUserBasicInfo.UserID.String &&
			bobUserBasicInfo.Name.String == entryexitUserBasicInfo.Name.String &&
			bobUserBasicInfo.FirstName.String == entryexitUserBasicInfo.FirstName.String &&
			bobUserBasicInfo.LastName.String == entryexitUserBasicInfo.LastName.String &&
			bobUserBasicInfo.FullNamePhonetic.String == entryexitUserBasicInfo.FullNamePhonetic.String &&
			bobUserBasicInfo.FirstNamePhonetic.String == entryexitUserBasicInfo.FirstNamePhonetic.String &&
			bobUserBasicInfo.LastNamePhonetic.String == entryexitUserBasicInfo.LastNamePhonetic.String &&
			bobUserBasicInfo.CurrentGrade.Int == entryexitUserBasicInfo.CurrentGrade.Int &&
			bobUserBasicInfo.GradeID.String == entryexitUserBasicInfo.GradeID.String &&
			bobUserBasicInfo.Email.String == entryexitUserBasicInfo.Email.String {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("user basic info record not sync correctly on entryexitmgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
