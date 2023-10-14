package payment

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/payment/entities"
)

func (s *suite) adminInsertsAUserRecordToBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	uniqueID := idutil.ULIDNow()

	stepState.NameOfData = fmt.Sprintf("User-name-kafka %s", uniqueID)
	stmt := `INSERT INTO users 
	(user_id, name, user_group, country, phone_number, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5, now(), now())`
	_, err := s.BobDBTrace.Exec(ctx, stmt,
		fmt.Sprintf("User-id-kafka %s", uniqueID),
		stepState.NameOfData,
		"USER_GROUP_STUDENT",
		"COUNTRY_VN",
		uniqueID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) paymentUserTableWillBeUpdated(ctx context.Context) (context.Context, error) {
	time.Sleep(1250 * time.Millisecond)
	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT 
			user_id,
			name,
			user_group
		FROM
			users
		WHERE name = $1
		`
	userPayment := &entities.User{}
	paymentRow := s.FatimaDBTrace.QueryRow(ctx, stmt, stepState.NameOfData)
	err := paymentRow.Scan(
		&userPayment.UserID,
		&userPayment.Name,
		&userPayment.Group,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
