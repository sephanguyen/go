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

func (s *suite) aStudentRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	student, err := s.insertStudentToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentID = student.ID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentRecordMustBeRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			student_id,
			current_grade,
			school_id,
			resource_path
		FROM
			students
		WHERE
		student_id = $1
		`

	bobStudent := &bob_entities.Student{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.StudentID)
	err := bobRow.Scan(
		&bobStudent.ID,
		&bobStudent.CurrentGrade,
		&bobStudent.SchoolID,
		&bobStudent.EnrollmentStatus,
		&bobStudent.ResourcePath,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		entryExitMgmtStudent := &entryexit_entities.Student{}

		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, stepState.StudentID).Scan(
			&entryExitMgmtStudent.ID,
			&entryExitMgmtStudent.CurrentGrade,
			&entryExitMgmtStudent.SchoolID,
			&entryExitMgmtStudent.ResourcePath,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		// Compare the existing columns
		if bobStudent.ID == entryExitMgmtStudent.ID &&
			bobStudent.CurrentGrade == entryExitMgmtStudent.CurrentGrade &&
			bobStudent.SchoolID == entryExitMgmtStudent.SchoolID &&
			bobStudent.ResourcePath == entryExitMgmtStudent.ResourcePath {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("students record not sync correctly on entryexitmgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
