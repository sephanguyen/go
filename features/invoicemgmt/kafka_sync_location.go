package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aLocationRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateLocation(ctx, s.BobDBTrace),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisLocationRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT location_id, name, location_type, partner_internal_parent_id 
		FROM locations 
		WHERE location_id = $1
	`

	// Get the location in bob DB
	bobLocation := &bobEntities.Location{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.LocationID)
	err := bobRow.Scan(&bobLocation.LocationID, &bobLocation.Name, &bobLocation.LocationType, &bobLocation.PartnerInternalParentID)
	if err != nil {
		return ctx, err
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get the location in invoicemgmt DB
		location := &entities.Location{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, bobLocation.LocationID.String).Scan(
			&location.LocationID, &location.Name, &location.LocationType, &location.PartnerInternalParentID,
		)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if location.LocationID.String == bobLocation.LocationID.String &&
			location.Name.String == bobLocation.Name.String &&
			location.LocationType.String == bobLocation.LocationType.String &&
			location.PartnerInternalParentID.String == bobLocation.PartnerInternalParentID.String {
			return false, nil
		}
		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("location record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
