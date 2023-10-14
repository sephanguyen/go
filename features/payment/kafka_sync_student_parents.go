package payment

import (
	"context"
	"fmt"
	"time"

	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	discountEntities "github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

func (s *suite) aRecordIsInsertedInStudentParentInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	studentID, err := s.insertUserToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// insert user parent on bob
	userParentID, err := s.insertParentToBob(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.insertStudentParentToBob(ctx, studentID, userParentID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.StudentID = studentID
	stepState.CurrentParentID = userParentID

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theStudentParentMustBeRecordedInPayment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(3 * time.Second)

	stmt := `
		SELECT
			student_id,
			parent_id,
			relationship
		FROM
			student_parents
		WHERE
			parent_id = $1 AND student_id = $2
		`
	// added try do for catching kafka sync delay
	if err := try.Do(func(attempt int) (bool, error) {
		paymentStudentParent := &discountEntities.StudentParent{}

		err := s.FatimaDBTrace.QueryRow(ctx, stmt, stepState.CurrentParentID, stepState.StudentID).Scan(
			&paymentStudentParent.StudentID,
			&paymentStudentParent.ParentID,
			&paymentStudentParent.Relationship,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		// Compare the existing columns
		if paymentStudentParent.StudentID.String == stepState.StudentID && paymentStudentParent.ParentID.String == stepState.CurrentParentID {
			return false, nil
		}

		time.Sleep(1 * time.Second)
		return attempt < 10, fmt.Errorf("student_parents record not sync correctly on fatima db")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertParentToBob(ctx context.Context) (string, error) {
	id := idutil.ULIDNow()
	user := &bobEntities.User{}
	err := multierr.Combine(
		user.ID.Set(id),
		user.Country.Set("COUNTRY_VN"),
		user.LastName.Set(fmt.Sprintf("kafka-test-name-%v", id)),
		user.GivenName.Set(fmt.Sprintf("kafka-test-given-name-%v", id)),
		user.DeviceToken.Set(fmt.Sprintf("kafka-test-device-token-%v", id)),
		user.AllowNotification.Set(true),
		user.Group.Set(constant.UserGroupParent),
	)
	if err != nil {
		return "", errors.WithMessage(err, "err multi combine student parent")
	}

	stmt := `INSERT INTO users (
				user_id,
				country,
				name,
				given_name,
				device_token,
				allow_notification,
				user_group,
				created_at,
				updated_at
			) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())`
	_, err = s.BobDBTrace.Exec(ctx, stmt,
		user.ID.String,
		user.Country.String,
		user.LastName.String,
		user.GivenName.String,
		user.DeviceToken.String,
		user.AllowNotification.Bool,
		user.Group.String,
	)
	if err != nil {
		return "", errors.WithMessage(err, "insert parent to bob")
	}

	return id, nil
}

func (s *suite) insertStudentParentToBob(ctx context.Context, studentID, parentID string) error {
	stmt := `INSERT INTO student_parents (
		student_id,
		parent_id,
		relationship,
		created_at,
		updated_at
	)
	VALUES ($1, $2, $3, now(), now())`
	_, err := s.BobDBTrace.Exec(ctx, stmt,
		studentID,
		parentID,
		upb.FamilyRelationship_FAMILY_RELATIONSHIP_FATHER,
	)
	if err != nil {
		return errors.Wrap(err, "insert to bob student parents")
	}

	return nil
}
