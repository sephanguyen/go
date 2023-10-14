package managing

import (
	"context"
	"errors"
	"time"

	"github.com/manabie-com/backend/features/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) tomMustInsertInfoToTableUserDeviceTokensWithUpdatedName(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := "SELECT COUNT(user_id) " +
			"FROM user_device_tokens " +
			"WHERE created_at = updated_at " +
			"AND user_id = $1 " +
			"AND user_name = $2"
		profile := bob.StepStateFromContext(ctx).Request.(*pb.UpdateUserProfileRequest).Profile
		userId := profile.Id
		userName := profile.Name
		rows, err := s.tomDB.Query(ctx, query, userId, userName)
		defer rows.Close()

		if err != nil {
			return err

		}

		var count int
		for rows.Next() {
			err = rows.Scan(&count)
			if err != nil {
				return err

			}
		}

		if count == 0 {
			return errors.New("tom not insert new info to table user_device_tokens")
		}

		return nil
	}
	return GandalfStepStateToContext(ctx, stepState), s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
