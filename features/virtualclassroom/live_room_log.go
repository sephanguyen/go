package virtualclassroom

import (
	"context"
	"fmt"
	"strconv"

	logger_repo "github.com/manabie-com/backend/internal/virtualclassroom/modules/logger/infrastructure/repo"
)

func (s *suite) haveAnUncompletedLiveRoomLog(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedJoinedAttendees, err := strconv.Atoi(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected joined attendees need a number")
	}
	expectedNumberOfTimesGettingRoomState, err := strconv.Atoi(arg2)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times getting room state need a number")
	}
	expectedNumberOfTimesUpdatingRoomState, err := strconv.Atoi(arg3)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times updating room state need a number")
	}
	expectedNumberOfTimesReconnection, err := strconv.Atoi(arg4)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times reconnection room state need a number")
	}

	return s.haveALiveRoomLog(ctx, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection, false)
}

func (s *suite) haveACompletedLiveRoomLog(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	expectedJoinedAttendees, err := strconv.Atoi(arg1)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected joined attendees need a number")
	}
	expectedNumberOfTimesGettingRoomState, err := strconv.Atoi(arg2)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times getting room state need a number")
	}
	expectedNumberOfTimesUpdatingRoomState, err := strconv.Atoi(arg3)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times updating room state need a number")
	}
	expectedNumberOfTimesReconnection, err := strconv.Atoi(arg4)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("number of times reconnection room state need a number")
	}

	return s.haveALiveRoomLog(ctx, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection, true)
}

func (s *suite) haveALiveRoomLog(ctx context.Context, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection int, isCompleted bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	logRepo := logger_repo.LiveRoomLogRepo{}
	actual, err := logRepo.GetLatestByChannelID(ctx, s.CommonSuite.LessonmgmtDBTrace, stepState.CurrentChannelID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("logRepo.GetLatestByChannelID: %w", err)
	}

	expectedLogID := stepState.CurrentLiveRoomLogID
	if len(expectedLogID) != 0 {
		if expectedLogID != actual.LiveRoomLogID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected log id %s but got %s", expectedLogID, actual.LiveRoomLogID.String)
		}
	}
	if actual.IsCompleted.Bool != isCompleted {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected this log have is_completed field: %v but got %v", isCompleted, actual.IsCompleted.Bool)
	}
	if expectedJoinedAttendees != len(actual.AttendeeIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of joined attendees is %d but got %d", expectedJoinedAttendees, len(actual.AttendeeIDs.Elements))
	}
	if int32(expectedNumberOfTimesGettingRoomState) != actual.TotalTimesGettingRoomState.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of times getting room state is %d but got %d", expectedNumberOfTimesGettingRoomState, actual.TotalTimesGettingRoomState.Int)
	}
	if int32(expectedNumberOfTimesUpdatingRoomState) != actual.TotalTimesUpdatingRoomState.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of times updating room state is %d but got %d", expectedNumberOfTimesUpdatingRoomState, actual.TotalTimesUpdatingRoomState.Int)
	}
	if int32(expectedNumberOfTimesReconnection) != actual.TotalTimesReconnection.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected number of times reconnection is %d but got %d", expectedNumberOfTimesReconnection, actual.TotalTimesReconnection.Int)
	}

	stepState.CurrentLiveRoomLogID = actual.LiveRoomLogID.String
	return StepStateToContext(ctx, stepState), nil
}
