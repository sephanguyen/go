package entryexitmgmt

import (
	"context"
	"fmt"
	"time"

	entryexit_entities "github.com/manabie-com/backend/internal/entryexitmgmt/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aGradeRecordIsInsertedInMastermgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	gradeID := idutil.ULIDNow()
	insertStmt := `
		INSERT INTO grade(
			grade_id,
			name,
			is_archived,
			partner_internal_id,
			updated_at,
			created_at
		) VALUES ($1, $2, $3, $4, now(), now())
	`
	_, err := s.MasterMgmtDBTrace.Exec(ctx, insertStmt,
		gradeID,
		fmt.Sprintf("grade-%v", gradeID),
		false,
		fmt.Sprintf("internal-%v", gradeID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "insert to mastermgmt.grade")
	}

	stepState.GradeID = gradeID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisGradeRecordMustBeRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	stmt := "SELECT grade_id, name FROM grade WHERE grade_id = $1"

	mastermgmtGrade := &entryexit_entities.Grade{}
	mastermgmtRow := s.MasterMgmtDBTrace.QueryRow(ctx, stmt, stepState.GradeID)
	err := mastermgmtRow.Scan(
		&mastermgmtGrade.ID,
		&mastermgmtGrade.Name,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "grade record not found in mastermgmt")
	}

	if err := try.Do(func(attempt int) (bool, error) {
		entryExitGrade := &entryexit_entities.Grade{}
		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, stepState.GradeID).Scan(
			&entryExitGrade.ID,
			&entryExitGrade.Name,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if mastermgmtGrade.ID.String == entryExitGrade.ID.String &&
			mastermgmtGrade.Name.String == entryExitGrade.Name.String {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("grade record not sync correctly on entryexitmgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
