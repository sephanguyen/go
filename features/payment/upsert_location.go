package payment

import (
	"context"
	"fmt"
	"strings"
	"time"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"

	"go.uber.org/multierr"
)

func (s *suite) prepareLocationDataForInsert(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	id := idutil.ULIDNow()
	location := &bobEntities.Location{}
	err := multierr.Combine(
		location.LocationID.Set(id),
		location.LocationType.Set("01FR4M51XJY9E77GSN4QZ1Q9M1"),
		location.Name.Set(fmt.Sprintf("test-location-%v", id)),
		location.PartnerInternalID.Set("1"),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = location
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertLocationFromBobDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	location := stepState.Request.(*bobEntities.Location)

	stmt := `INSERT INTO locations (
                   location_id,
                   name,
                   location_type,
                   partner_internal_id,
                   updated_at,
                   created_at)
				VALUES ($1, $2, $3, $4, now(), now())`
	_, err := s.BobDBTrace.Exec(ctx, stmt,
		location.LocationID.String,
		location.Name.String,
		location.LocationType.String,
		location.PartnerInternalID.String,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkSyncLocationSuccess(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	bobLocation := stepState.Request.(*bobEntities.Location)
	location := &entities.Location{}
	locationFieldNames, locationFieldValues := location.FieldMap()
	stmt := `SELECT %s
		FROM 
			%s
		WHERE 
			location_id = $1`

	stmt = fmt.Sprintf(
		stmt,
		strings.Join(locationFieldNames, ","),
		location.TableName(),
	)
	row := s.FatimaDBTrace.QueryRow(ctx, stmt, bobLocation.LocationID.String)
	err := row.Scan(locationFieldValues...)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if !(location.LocationID.String == bobLocation.LocationID.String &&
		location.Name.String == bobLocation.Name.String &&
		location.LocationType.String == bobLocation.LocationType.String &&
		location.PartnerInternalParentID.String == bobLocation.PartnerInternalParentID.String) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("difference location between bob and payment")
	}
	return StepStateToContext(ctx, stepState), nil
}
