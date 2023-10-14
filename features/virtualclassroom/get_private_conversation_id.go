package virtualclassroom

import (
	"context"
	"fmt"
	"reflect"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userAlreadyHasExistingPrivConvWithOneOfTheStudentAcc(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if len(stepState.StudentIds) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no student IDs are found")
	}

	participantIDs := []string{stepState.StudentIds[0], stepState.CurrentUserID}

	tempCtx, err := s.userGetsPrivateConversationIDs(ctx, stepState.CurrentLessonID, participantIDs)
	tempStepState := StepStateFromContext(tempCtx)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed in creating private conversation: %w", err)
	}
	if tempStepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed in creating private conversation: %w", tempStepState.ResponseErr)
	}
	response := tempStepState.Response.(*vpb.GetPrivateConversationIDsResponse)
	privConvMap := response.GetParticipantConversationMap()
	if len(privConvMap) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected a conversation ID but did not receive any: %s", response.FailedPrivConv.GetErrorMsg())
	}
	if _, ok := privConvMap[stepState.StudentIds[0]]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("conversation ID is not found with user %s, current user %s", stepState.StudentIds[0], stepState.CurrentUserID)
	}

	stepState.ExpectedPrivConversationID = privConvMap[stepState.StudentIds[0]]

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsPrivateConversationIDs(ctx context.Context, lessonID string, participantIDs []string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetPrivateConversationIDsRequest{
		LessonId:       lessonID,
		ParticipantIds: participantIDs,
	}
	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomChatServiceClient(s.VirtualClassroomConn).
		GetPrivateConversationIDs(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsPrivateConversationIDsStep(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	participants := stepState.StudentIds

	return s.userGetsPrivateConversationIDs(ctx, stepState.CurrentLessonID, participants)
}

func (s *suite) userGetsPrivateConversationIDsAgainStep(ctx context.Context) (context.Context, error) {
	return s.userGetsPrivateConversationIDsStep(ctx)
}

func (s *suite) userGetsNonEmptyPrivateConversationIDs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetPrivateConversationIDsResponse)

	actualPrivConvMap := response.GetParticipantConversationMap()
	if len(actualPrivConvMap) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting private conversation map but got empty")
	}
	stepState.ExpectedPrivConversationMap = actualPrivConvMap

	request := stepState.Request.(*vpb.GetPrivateConversationIDsRequest)
	requestParticipantIDs := request.GetParticipantIds()
	if len(actualPrivConvMap) != len(requestParticipantIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the number of created conversation IDs does not match the number of participant IDs: err %s", response.GetFailedPrivConv().GetErrorMsg())
	}

	expectedConvID := stepState.ExpectedPrivConversationID
	if len(expectedConvID) > 0 {
		convIDFound := false
		for _, convID := range actualPrivConvMap {
			if convID == expectedConvID {
				convIDFound = true
			}
		}
		if !convIDFound {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected conversation ID %s was not in the returned conversation IDs map", expectedConvID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsExpectedPrivateConversationIDs(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetPrivateConversationIDsResponse)

	actualPrivConvMap := response.GetParticipantConversationMap()
	if len(actualPrivConvMap) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting private conversation map but got empty")
	}

	request := stepState.Request.(*vpb.GetPrivateConversationIDsRequest)
	requestParticipantIDs := request.GetParticipantIds()
	if len(actualPrivConvMap) != len(requestParticipantIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the number of created conversation IDs does not match the number of participant IDs: err %s", response.GetFailedPrivConv().GetErrorMsg())
	}

	expectedPrivConvMap := stepState.ExpectedPrivConversationMap
	if !reflect.DeepEqual(actualPrivConvMap, expectedPrivConvMap) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the expected conversation map [%s] does not match the actual conversation map [%s]", expectedPrivConvMap, actualPrivConvMap)
	}

	expectedConvID := stepState.ExpectedPrivConversationID
	if len(expectedConvID) > 0 {
		convIDFound := false
		for _, convID := range actualPrivConvMap {
			if convID == expectedConvID {
				convIDFound = true
			}
		}
		if !convIDFound {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expected conversation ID %s was not in the returned conversation IDs map", expectedConvID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsTheExpectedOnvePrivateConversationID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetConversationIDResponse)

	actualConvID := response.GetConversationId()
	if len(actualConvID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting private conversation ID but got empty")
	}

	expectedConvID := stepState.ExpectedPrivConversationID
	if actualConvID != expectedConvID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected conversation ID %s does not match with the actual conversation ID %s", expectedConvID, actualConvID)
	}

	return StepStateToContext(ctx, stepState), nil
}
