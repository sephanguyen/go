package managing

import (
	"context"
	"errors"
	"time"

	"github.com/manabie-com/backend/features/bob"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) tomMustRecordNewUserDeviceTokenWithDeviceTokenInfo(ctx context.Context) (context.Context, error) {
	mainProcess := func() error {
		query := "select count(t1.user_id) " +
			"from user_device_tokens t1 join user_device_tokens t2 on t1.user_id = t2.user_id " +
			"where t1.created_at = t2.updated_at and t1.user_id = $1 and t1.token= $2"
		request := bob.StepStateFromContext(ctx).Request.(*pb.UpdateUserDeviceTokenRequest)
		userId := request.UserId
		token := request.DeviceToken
		rows, err := s.tomDB.Query(ctx, query, userId, token)

		if err != nil {
			return err
		}
		defer rows.Close()

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
	return ctx, s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
