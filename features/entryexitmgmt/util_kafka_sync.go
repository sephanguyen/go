package entryexitmgmt

import (
	"context"
	"fmt"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) insertUserToBob(ctx context.Context, userGroup string) (*bob_entities.User, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	user := &bob_entities.User{}
	err := multierr.Combine(
		user.ID.Set(id),
		user.Country.Set("COUNTRY_VN"),
		user.LastName.Set(fmt.Sprintf("kafka-test-name-%v", id)),
		user.GivenName.Set(fmt.Sprintf("kafka-test-given-name-%v", id)),
		user.DeviceToken.Set(fmt.Sprintf("kafka-test-device-token-%v", id)),
		user.AllowNotification.Set(true),
		user.Group.Set(userGroup),
		user.ResourcePath.Set(stepState.ResourcePath),
	)
	if err != nil {
		errors.WithMessage(err, "insert user to bob")
	}

	stmt := `INSERT INTO users (
				user_id,
				country,
				name,
				given_name,
				device_token,
				allow_notification,
				user_group,
				resource_path,
				created_at,
				updated_at
			) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		user.ID.String,
		user.Country.String,
		user.LastName.String,
		user.GivenName.String,
		user.DeviceToken.String,
		user.AllowNotification.Bool,
		user.Group.String,
		user.ResourcePath.String,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "insert user to bob")
	}

	return user, nil
}

func (s *suite) insertLocationToBob(ctx context.Context) (*bob_entities.Location, error) {
	id := idutil.ULIDNow()

	location := &bob_entities.Location{}
	err := multierr.Combine(
		location.LocationID.Set(id),
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
				updated_at
			)
			VALUES ($1, $2, $3, $4, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		location.LocationID.String,
		location.Name.String,
		location.LocationType.String,
		location.PartnerInternalID.String,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "insert location to bob")
	}

	return location, nil
}

func (s *suite) insertStudentToBob(ctx context.Context) (*bob_entities.Student, error) {
	stepState := StepStateFromContext(ctx)

	user, err := s.insertUserToBob(ctx, bob_entities.UserGroupStudent)
	if err != nil {
		return nil, err
	}

	student := &bob_entities.Student{}
	err = multierr.Combine(
		student.ID.Set(user.ID),
		student.CurrentGrade.Set(5),
		student.SchoolID.Set(stepState.CurrentSchoolID),
		student.EnrollmentStatus.Set("STUDENT_ENROLLMENT_STATUS_ENROLLED"),
		student.ResourcePath.Set(user.ResourcePath),
	)
	if err != nil {
		errors.WithMessage(err, "insert user to bob")
	}

	stmt := `INSERT INTO students (
				student_id,
				current_grade,
				school_id,
				enrollment_status,
				resource_path,
				billing_date,
				created_at,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, now(), now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		&student.ID,
		&student.CurrentGrade,
		&student.SchoolID,
		&student.EnrollmentStatus,
		&student.ResourcePath,
	)
	if err != nil {
		return nil, errors.Wrap(err, "insert to bob students")
	}

	return student, nil
}
