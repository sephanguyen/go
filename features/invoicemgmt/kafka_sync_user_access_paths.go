package invoicemgmt

import (
	"context"
	"fmt"
	"time"

	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	bobEntities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"

	pgx "github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) aUserAccessPathRecordIsInsertedInBob(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	id := idutil.ULIDNow()
	userID := fmt.Sprintf("kafka-test-invoice-user-id-%v", id)

	err := InsertEntities(
		stepState,
		s.EntitiesCreator.CreateUser(ctx, s.BobDBTrace, userID, bobEntities.UserGroupStudent),
		s.EntitiesCreator.CreateLocation(ctx, s.BobDBTrace),
		s.EntitiesCreator.CreateUserAccessPath(ctx, s.BobDBTrace),
	)

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) thisUserAccessPathRecordedInInvoicemgmt(ctx context.Context) (context.Context, error) {
	time.Sleep(invoiceConst.KafkaSyncSleepDuration)

	stepState := StepStateFromContext(ctx)

	stmt := `
		SELECT
			count(user_id)
		FROM
			user_access_paths
		WHERE
			user_id = $1
		AND
			location_id = $2
		`

	if err := try.Do(func(attempt int) (bool, error) {
		var count int
		err := s.InvoiceMgmtPostgresDBTrace.QueryRow(ctx, stmt, stepState.CurrentUserID, stepState.LocationID).Scan(&count)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return false, err
		}
		if count == 1 {
			return false, nil
		}
		if count > 1 {
			return false, fmt.Errorf("unexpected %d user access path created on invoicemgmt", count)
		}

		time.Sleep(invoiceConst.ReselectSleepDuration)
		return attempt < 10, fmt.Errorf("user_access_path record not sync correctly on invoicemgmt")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
