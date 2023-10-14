package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	userEntities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aPrefectureRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreatePrefecture(ctx, s.BobDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisPrefectureRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT prefecture_id, prefecture_code, country, name 
		FROM prefecture 
		WHERE prefecture_id = $1
	`

	// Get the prefecture in bob DB
	bobPrefecture := &userEntities.Prefecture{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.PrefectureID)
	err := bobRow.Scan(&bobPrefecture.ID, &bobPrefecture.PrefectureCode, &bobPrefecture.Country, &bobPrefecture.Name)
	if err != nil {
		return ctx, err
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get the prefecture from invoicemgmt DB
		prefecture := &entities.Prefecture{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.PrefectureID).Scan(
			&prefecture.ID, &prefecture.PrefectureCode, &prefecture.Country, &prefecture.Name,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if prefecture.ID.String == bobPrefecture.ID.String &&
			prefecture.PrefectureCode.String == bobPrefecture.PrefectureCode.String &&
			prefecture.Country.String == bobPrefecture.Country.String &&
			prefecture.Name.String == bobPrefecture.Name.String {
			return false, nil
		}
		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("prefecture record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
