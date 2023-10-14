package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aUserBasicInfoRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateUser(ctx, s.BobDBTrace, id, bobEntities.UserGroupSchoolAdmin),
		s.EntitiesCreator.CreateUserBasicInfo(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserBasicInfoRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

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
			grade_id
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
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get the user basic info from invoicemgmt DB
		invoiceMgmtUserBasicInfo := &entities.UserBasicInfo{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID).Scan(
			&invoiceMgmtUserBasicInfo.UserID,
			&invoiceMgmtUserBasicInfo.Name,
			&invoiceMgmtUserBasicInfo.FirstName,
			&invoiceMgmtUserBasicInfo.LastName,
			&invoiceMgmtUserBasicInfo.FullNamePhonetic,
			&invoiceMgmtUserBasicInfo.FirstNamePhonetic,
			&invoiceMgmtUserBasicInfo.LastNamePhonetic,
			&invoiceMgmtUserBasicInfo.CurrentGrade,
			&invoiceMgmtUserBasicInfo.GradeID,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if bobUserBasicInfo.UserID.String == invoiceMgmtUserBasicInfo.UserID.String &&
			bobUserBasicInfo.Name.String == invoiceMgmtUserBasicInfo.Name.String &&
			bobUserBasicInfo.FirstName.String == invoiceMgmtUserBasicInfo.FirstName.String &&
			bobUserBasicInfo.LastName.String == invoiceMgmtUserBasicInfo.LastName.String &&
			bobUserBasicInfo.FullNamePhonetic.String == invoiceMgmtUserBasicInfo.FullNamePhonetic.String &&
			bobUserBasicInfo.FirstNamePhonetic.String == invoiceMgmtUserBasicInfo.FirstNamePhonetic.String &&
			bobUserBasicInfo.LastNamePhonetic.String == invoiceMgmtUserBasicInfo.LastNamePhonetic.String &&
			bobUserBasicInfo.CurrentGrade.Int == invoiceMgmtUserBasicInfo.CurrentGrade.Int &&
			bobUserBasicInfo.GradeID.String == invoiceMgmtUserBasicInfo.GradeID.String {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("user basic info record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
