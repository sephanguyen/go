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

func (s *suite) aUserRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	userID := fmt.Sprintf("kafka-test-invoice-user-id-%v", id)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateUser(ctx, s.BobDBTrace, userID, bobEntities.UserGroupStudent),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserRecordMustBeRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT 
			user_id,
			name,
			user_group
		FROM
			users
		WHERE user_id = $1
		`

	// Get the user from bob DB
	bobUser := &entities.User{}
	bobRow := s.BobDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID)
	err := bobRow.Scan(
		&bobUser.UserID,
		&bobUser.Name,
		&bobUser.Group,
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "record not found in bob")
	}

	if err := try.Do(func(attempt int) (bool, error) {

		// Get the user from invoicemgmt DB
		userInvoiceMgmt := &entities.User{}
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID).Scan(&userInvoiceMgmt.UserID, &userInvoiceMgmt.Name, &userInvoiceMgmt.Group)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}

		if bobUser.Name == userInvoiceMgmt.Name && bobUser.UserID == userInvoiceMgmt.UserID && bobUser.Group == userInvoiceMgmt.Group {
			return false, nil
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("user record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil

}
