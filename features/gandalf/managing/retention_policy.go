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

func (s *suite) publishSomeMessage(ctx context.Context, userName string, number int, subjectName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	for i := 0; i < number; i++ {
		payload := map[string]interface{}{}
		payloadJson, _ := json.Marshal(payload)
		msg := npb.ActivityLogEvtCreated{UserId: ksuid.New().String(), ActionType: "TestRetentionPolicy", ResourcePath: "gandalf", RequestAt: timestamppb.Now(), Payload: payloadJson}
		data, err := proto.Marshal(&msg)
		if err != nil {
			return GandalfStepStateToContext(ctx, stepState), err

		}
		pubAck, err := s.bobSuite.JSM.PublishContext(ctx, subjectName, data)
		if err != nil {
			return ctx, err
		}
		stepState.ZeusStepState.SequenceIDs = append(stepState.ZeusStepState.SequenceIDs, pubAck.Sequence)
		stepState.ZeusStepState.ActivityLogPublished = append(stepState.ZeusStepState.ActivityLogPublished, npb.ActivityLogEvtCreated{UserId: msg.UserId, ActionType: msg.ActionType, ResourcePath: msg.ResourcePath, RequestAt: msg.RequestAt, Payload: msg.Payload})
	}
	stepState.ZeusStepState.CurrentUserNameConnectedToJetStream = userName
	return GandalfStepStateToContext(ctx, stepState), nil
}

func (s *suite) theseActivityLogAreCreatedByZeus(ctx context.Context) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	mainProcess := func() error {
		var totalLogInserted int
		for i := range stepState.ZeusStepState.ActivityLogPublished {
			count, err := s.countActivityLog(ctx, &stepState.ZeusStepState.ActivityLogPublished[i])
			if err != nil {
				return err

			}
			if count == 0 {
				return fmt.Errorf("not found any activity_logs where user_id = %s, action_type = %s, resource_path = %s, request_at = %v", stepState.ZeusStepState.ActivityLogPublished[i].UserId, stepState.ZeusStepState.ActivityLogPublished[i].ActionType, stepState.ZeusStepState.ActivityLogPublished[i].ResourcePath, stepState.ZeusStepState.ActivityLogPublished[i].RequestAt)
			}
			totalLogInserted = totalLogInserted + count
		}
		totalActivityLogPublished := len(stepState.ZeusStepState.ActivityLogPublished)
		if totalLogInserted != totalActivityLogPublished {
			return fmt.Errorf("expected total log inserted is %d, but the fact is %d", totalActivityLogPublished, totalLogInserted)
		}
		return nil
	}
	return GandalfStepStateToContext(ctx, stepState), s.ExecuteWithRetry(mainProcess, 2*time.Second, 10)
}

func (s *suite) countActivityLog(ctx context.Context, input *npb.ActivityLogEvtCreated) (int, error) {
	query := `SELECT count(activity_log_id)
			FROM activity_logs
			WHERE user_id = $1
			AND action_type = $2
			AND resource_path = $3`
	rows, err := s.zeusDB.Query(ctx, query, input.UserId, input.ActionType, input.ResourcePath)
	defer rows.Close()
	if err != nil {
		return 0, err
	}
	var count int
	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}
func (s *suite) messageMustBeDeletedFromStream(ctx context.Context, streamName string) (context.Context, error) {
	stepState := GandalfStepStateFromContext(ctx)
	jsm := stepState.ZeusStepState.MapJSContext[stepState.ZeusStepState.CurrentUserNameConnectedToJetStream]
	defer s.closeAllJetStreamConnection(ctx)

	for _, v := range stepState.ZeusStepState.SequenceIDs {
		msg, err := jsm.GetMsg(streamName, v)
		if err == nil {
			return ctx, fmt.Errorf("expected message is deleted, but the fact we find a msg with sequence_id: %d", msg.Sequence)
		}
		if err.Error() != nats.ErrMsgNotFound.Error() {
			return ctx, err
		}
	}

	return ctx, nil
}
