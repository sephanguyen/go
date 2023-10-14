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

func (s *suite) aLocationRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	location, err := s.insertLocationToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.LocationID = location.LocationID.String

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisLocationRecordMustBeRecordedInEntryExitMgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			location_id,
			name,
			location_type,
			partner_internal_id
		FROM
			locations
		WHERE location_id = $1
		`

	bobLocation := &bob_entities.Location{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.LocationID)
	err := bobRow.Scan(
		&bobLocation.LocationID,
		&bobLocation.Name,
		&bobLocation.LocationType,
		&bobLocation.PartnerInternalID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		entryExitLocation := &entryexit_entities.Location{}
		err := s.EntryExitMgmtDBTrace.QueryRow(ctx, stmt, bobLocation.LocationID.String).Scan(
			&entryExitLocation.LocationID,
			&entryExitLocation.Name,
			&entryExitLocation.LocationType,
			&entryExitLocation.PartnerInternalID,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if bobLocation.LocationID.String == entryExitLocation.LocationID.String &&
			bobLocation.Name.String == entryExitLocation.Name.String &&
			bobLocation.LocationType.String == entryExitLocation.LocationType.String &&
			bobLocation.PartnerInternalParentID.String == entryExitLocation.PartnerInternalParentID.String {
			return false, nil
		}
		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("locations record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
