package managing

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/nats-io/nats.go"
	"github.com/segmentio/ksuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) publishSomeMessageWithSameNatsMsgID(ctx context.Context, userName string, number int, subjectName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	stepState.GandalfStateCurrentUserID = ksuid.New().String()
	stepState.ZeusStepState.CurrentActionType = "TestMessageDeduplication"
	stepState.ZeusStepState.CurrentResourcePath = "gandalf"
	for i := 0; i < number; i++ {
		payload := map[string]interface{}{}
		payloadJson, _ := json.Marshal(payload)
		msg := npb.ActivityLogEvtCreated{UserId: stepState.GandalfStateCurrentUserID, ActionType: stepState.ZeusStepState.CurrentActionType, ResourcePath: stepState.ZeusStepState.CurrentResourcePath, RequestAt: timestamppb.Now(), Payload: payloadJson}
		data, err := proto.Marshal(&msg)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
		_, err = s.bobSuite.JSM.PublishContext(ctx, subjectName, data, nats.MsgId(stepState.GandalfStateCurrentUserID))
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err
		}
	}

	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) totalRecordIsInsertedMustBeOne(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		query := `SELECT count(activity_log_id)
				FROM activity_logs
				WHERE user_id = $1
				AND action_type = $2
				AND resource_path = $3`
		rows, err := s.zeusDB.Query(ctx, query,
			stepState.GandalfStateCurrentUserID,
			stepState.ZeusStepState.CurrentActionType,
			stepState.ZeusStepState.CurrentResourcePath)
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

		if count != 10 {
			return fmt.Errorf("expected total record with user_id = %s, action_type = %s, resource_path = %s is %d, but the fact is %d",
				stepState.GandalfStateCurrentUserID,
				stepState.ZeusStepState.CurrentActionType,
				stepState.ZeusStepState.CurrentResourcePath,
				1,
				count)
		}

		return nil
	}

	return GandalfStepStateToContext(ctx, stepState), s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}
