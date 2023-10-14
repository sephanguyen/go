package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aStudentRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	studentID := fmt.Sprintf("kafka-test-invoice-user-id-%v", id)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateStudent(ctx, s.BobDBTrace, studentID),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisStudentRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			student_id,
			current_grade
		FROM
			students
		WHERE
		student_id = $1
	`
	// Get the student from bob DB
	bobStudent := &entities.Student{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.StudentID)
	err := bobRow.Scan(
		&bobStudent.StudentID,
		&bobStudent.CurrentGrade,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get the student from invoicemgmt DB
		invoiceMgmtStudent := &entities.Student{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.StudentID).Scan(&invoiceMgmtStudent.StudentID, &invoiceMgmtStudent.CurrentGrade)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if bobStudent.StudentID == invoiceMgmtStudent.StudentID && bobStudent.CurrentGrade == invoiceMgmtStudent.CurrentGrade {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("student record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
