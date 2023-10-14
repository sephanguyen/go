package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (s *suite) currentLiveRoomHasMaxStreamingLearner(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numberOfStream := s.Cfg.Agora.MaximumLearnerStreamings

	query := `INSERT INTO live_room_state (live_room_state_id, channel_id, stream_learner_counter) VALUES ($1, $2, $3)
		ON CONFLICT ON CONSTRAINT unique__channel_id 
		DO UPDATE SET stream_learner_counter = $3 `
	_, err := s.LessonmgmtDB.Exec(ctx, query, idutil.ULIDNow(), stepState.CurrentChannelID, numberOfStream)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to update stream learner counter for channel %s: %w", stepState.CurrentLessonID, err)
	}

	stepState.NumberOfStream = numberOfStream
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userPreparesToPublishInTheLiveRoom(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.PreparePublishLiveRoomRequest{
		ChannelId: stepState.CurrentChannelID,
		LearnerId: stepState.CurrentStudentID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewLiveRoomModifierServiceClient(s.VirtualClassroomConn).
		PreparePublishLiveRoom(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsPublishStatusInTheLiveRoom(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.PreparePublishLiveRoomResponse)
	var expectedStatus vpb.PrepareToPublishStatus
	switch status {
	case StatusNone:
		expectedStatus = vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE
	case "prepared before":
		expectedStatus = vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE
	case "max limit":
		expectedStatus = vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported expected status")
	}

	if response.Status != expectedStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected status %s does not match with actual status %s", expectedStatus.String(), response.Status.String())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) currentLiveRoomHasStreamingLearner(ctx context.Context, status, count string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	channelID := stepState.CurrentChannelID
	actualCount, learnerIDs, err := s.getLiveRoomStreamingCountAndStreamingLearners(ctx, channelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get streaming count and learners for channel %s: %w", channelID, err)
	}

	learnerID := stepState.CurrentStudentID
	switch status {
	case "includes":
		if !sliceutils.Contains(learnerIDs, learnerID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s is not found in the streaming learners live room %s but is expected", learnerID, channelID)
		}
	case "does not include":
		if sliceutils.Contains(learnerIDs, learnerID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s is found in the streaming learners of live room %s but expected not", learnerID, channelID)
		}
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported expected status")
	}

	expectedCount, err := strconv.Atoi(count)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to convert from string to int, count string %s: %w", count, err)
	}
	if actualCount != expectedCount {
		return StepStateToContext(ctx, stepState), fmt.Errorf("stream learner count is expected to be %d but got %d for channel %s", expectedCount, actualCount, channelID)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getLiveRoomStreamingCountAndStreamingLearners(ctx context.Context, channelID string) (int, []string, error) {
	query := `SELECT stream_learner_counter, streaming_learners FROM live_room_state
			  WHERE channel_id = $1 
			  AND deleted_at IS NULL `

	var streamLearnerCounter pgtype.Int4
	var streamingLearners pgtype.TextArray

	err := s.LessonmgmtDB.QueryRow(ctx, query, &channelID).Scan(&streamLearnerCounter, &streamingLearners)
	if err != nil && err != pgx.ErrNoRows {
		return 0, nil, fmt.Errorf("db.QueryRow: %w", err)
	}
	if err == pgx.ErrNoRows {
		return 0, []string{}, nil
	}

	return int(streamLearnerCounter.Int), database.FromTextArray(streamingLearners), nil
}
