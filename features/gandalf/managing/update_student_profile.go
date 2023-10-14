package managing

import (
	"context"
	"errors"

	"github.com/manabie-com/backend/features/bob"
	"github.com/manabie-com/backend/features/gandalf"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func (s *suite) tomMustRecordNewUserDeviceTokenWithMessageTypeIsUpdateProfileRequest(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := "select count(t1.user_id) " +
			"from user_device_tokens t1 join user_device_tokens t2 on t1.user_id = t2.user_id " +
			"where t1.created_at = t2.updated_at and t1.user_id = $1 and t1.user_name= $2"
		request := bob.StepStateFromContext(ctx).Request.(*pb.UpdateProfileRequest)
		userName := request.Name
		userId := stepState.BobStepState.CurrentStudentId
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
	return ctx, gandalf.Execute(mainProcess, gandalf.DefaultOption...)
}
