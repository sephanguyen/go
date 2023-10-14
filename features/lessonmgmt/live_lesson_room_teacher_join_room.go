package lessonmgmt

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/helper"
	repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) userJoinLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = bpb.NewClassModifierServiceClient(s.BobConn).
		JoinLesson(helper.GRPCContext(ctx, "token", stepState.AuthToken), &bpb.JoinLessonRequest{
			LessonId: stepState.CurrentLessonID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnsValidInformationForBroadcast(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rsp := stepState.Response.(*bpb.JoinLessonResponse)

	if rsp.VideoToken == "" {
		return StepStateToContext(ctx, stepState), fmt.Errorf("did not return valid video token")
	}

	return s.returnsValidInformationForStudentBroadcast(ctx)
}

func (s *Suite) haveAUncompletedVirtualClassRoomLog(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
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

	return s.haveAVirtualClassRoomLog(ctx, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection, false)
}

func (s *Suite) haveACompletedVirtualClassRoomLog(ctx context.Context, arg1, arg2, arg3, arg4 string) (context.Context, error) {
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

	ctx, err = s.haveAVirtualClassRoomLog(ctx, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection, true)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.CurrentVirtualClassroomLogID = ""
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) haveAVirtualClassRoomLog(ctx context.Context, expectedJoinedAttendees, expectedNumberOfTimesGettingRoomState, expectedNumberOfTimesUpdatingRoomState, expectedNumberOfTimesReconnection int, isCompleted bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	logRepo := repo.VirtualClassroomLogRepo{}
	actual, err := logRepo.GetLatestByLessonID(ctx, s.CommonSuite.BobDB, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("logRepo.GetLatestByLessonID: %w", err)
	}

	expectedLogID := stepState.CurrentVirtualClassroomLogID
	if len(expectedLogID) != 0 {
		if expectedLogID != actual.LogID.String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected log id %s but got %s", expectedLogID, actual.LogID.String)
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

	stepState.CurrentVirtualClassroomLogID = actual.LogID.String
	return StepStateToContext(ctx, stepState), nil
}
