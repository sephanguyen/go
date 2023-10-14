package payment

import (
	"context"
	"fmt"
	"time"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) insertUserAccessPathsRecordToBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userID, err := s.insertUserToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	locationID := constants.ManabieOrgLocation

	bobUser, err := s.getUserFromBob(ctx, userID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	bobLocation, err := s.getLocationFromBob(ctx, locationID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stmt := `INSERT INTO user_access_paths (
				user_id,
				location_id,
				created_at,
				updated_at)
			VALUES ($1, $2, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		bobUser.ID,
		bobLocation.LocationID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "insert to bob user_access_paths")
	}

	stepState.UserData = bobUser
	stepState.LocationData = bobLocation

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAccessPathsRecordedInPayment(ctx context.Context) (context.Context, error) {
	time.Sleep(3 * time.Second)

	stepState := StepStateFromContext(ctx)

	userID := stepState.UserData.ID
	locationID := stepState.LocationData.LocationID

	stmt := `
		SELECT
			user_id,
			location_id
		FROM
			user_access_paths
		WHERE
			user_id = $1
		AND
			location_id = $2
		`

	bobUserAccessPaths := &entities.UserAccessPaths{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, userID, locationID)
	err := bobRow.Scan(
		&bobUserAccessPaths.UserID,
		&bobUserAccessPaths.LocationID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	paymentUserAccessPaths := &entities.UserAccessPaths{}
	paymentRow := s.FatimaDBTrace.QueryRow(ctx, stmt, userID, locationID)
	err = paymentRow.Scan(
		&paymentUserAccessPaths.UserID,
		&paymentUserAccessPaths.LocationID,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in payment")
	}

	if bobUserAccessPaths.UserID != paymentUserAccessPaths.UserID && bobUserAccessPaths.LocationID != paymentUserAccessPaths.LocationID {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "new user_access_paths mismatch values in bob and payment")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getUserFromBob(ctx context.Context, userID string) (*bobEntities.User, error) {
	stmt :=
		`
		SELECT 
			user_id
		FROM
			users
		WHERE
			user_id = $1
		`
	row := s.BobDBTrace.QueryRow(
		ctx,
		stmt,
		userID,
	)

	user := &bobEntities.User{}
	err := row.Scan(
		&user.ID,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "rows.Scan users from bob")
	}

	return user, nil
}

func (s *suite) getLocationFromBob(ctx context.Context, locationID string) (*bobEntities.Location, error) {
	stmt :=
		`
		SELECT 
			location_id
		FROM
			locations
		WHERE
			location_id = $1
		`
	row := s.BobDBTrace.QueryRow(
		ctx,
		stmt,
		locationID,
	)

	location := &bobEntities.Location{}
	err := row.Scan(
		&location.LocationID,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "rows.Scan locations from bob")
	}

	return location, nil
}

func (s *suite) insertUserToBob(ctx context.Context) (string, error) {
	id := idutil.ULIDNow()

	user := &bobEntities.User{}

	err := multierr.Combine(
		user.UserID.Set(fmt.Sprintf("kafka-sync-test-user-id-%v", id)),
		user.Country.Set("COUNTRY_VN"),
		user.LastName.Set(fmt.Sprintf("kafka-sync-test-user-name-%v", id)),
		user.PhoneNumber.Set(id),
		user.Email.Set(fmt.Sprintf("%v@manabie.com", id)),
		user.Group.Set(bobEntities.UserGroupStudent),
	)
	if err != nil {
		errors.WithMessage(err, "insert user to bob")
	}

	stmt := `INSERT INTO users (
				user_id,
				name,
				phone_number,
				email,
				country,
				user_group,
				created_at,
				updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		user.UserID.String,
		user.LastName.String,
		user.PhoneNumber.String,
		user.Email.String,
		user.Country.String,
		user.Group.String,
	)
	if err != nil {
		return "", errors.WithMessage(err, "insert user to bob")
	}

	return user.UserID.String, nil
}

func (s *suite) insertLocationToBob(ctx context.Context) (string, error) {
	id := idutil.ULIDNow()

	location := &bobEntities.Location{}
	err := multierr.Combine(
		location.LocationID.Set(fmt.Sprintf("kafka-sync-test-location-id-%v", id)),
		location.LocationType.Set("01FR4M51XJY9E77GSN4QZ1Q9M1"),
		location.Name.Set(fmt.Sprintf("kafka-sync-test-location-name-%v", id)),
		location.PartnerInternalID.Set("1"),
	)
	if err != nil {
		errors.WithMessage(err, "insert location to bob")
	}

	stmt := `INSERT INTO locations (
				location_id,
				name,
				location_type,
				partner_internal_id,
				created_at,
				updated_at)
			VALUES ($1, $2, $3, $4, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		location.LocationID.String,
		location.Name.String,
		location.LocationType.String,
		location.PartnerInternalID.String,
	)
	if err != nil {
		return "", errors.WithMessage(err, "insert location to bob")
	}

	return location.LocationID.String, nil
}
