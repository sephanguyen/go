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

func (s *suite) aStudentParentsRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	student, err := s.insertStudentToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	userParent, err := s.insertUserToBob(ctx, bob_entities.UserGroupParent)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stmt := `INSERT INTO student_parents (
				student_id,
				parent_id,
				relationship,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		student.ID,
		userParent.ID,
		"FAMILY_RELATIONSHIP_FATHER",
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "insert to bob user_access_paths")
	}

	stepState.CurrentParentID = userParent.ID.String
	stepState.StudentID = student.ID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentParentsRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			student_id,
			parent_id,
			relationship
		FROM
			student_parents
		WHERE
			parent_id = $1 AND student_id = $2
		`

	bobStudentParent := &bob_entities.StudentParent{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.CurrentParentID, stepState.StudentID)
	err := bobRow.Scan(
		&bobStudentParent.StudentID,
		&bobStudentParent.ParentID,
		&bobStudentParent.Relationship,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		entryExitMgmtStudentParent := &entryexit_entities.StudentParent{}

		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, stepState.CurrentParentID, stepState.StudentID).Scan(
			&entryExitMgmtStudentParent.StudentID,
			&entryExitMgmtStudentParent.ParentID,
			&entryExitMgmtStudentParent.Relationship,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		// Compare the existing columns
		if bobStudentParent.StudentID == entryExitMgmtStudentParent.StudentID &&
			bobStudentParent.ParentID == entryExitMgmtStudentParent.ParentID &&
			bobStudentParent.Relationship == entryExitMgmtStudentParent.Relationship {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("student_parents record not sync correctly on entryexitmgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
