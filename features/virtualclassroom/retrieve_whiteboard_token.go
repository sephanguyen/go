package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) userRetrievesWhiteboardToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.RetrieveWhiteboardTokenRequest{
		LessonId: stepState.CurrentLessonID,
	}
	stepState.Request = req

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomReaderServiceClient(s.VirtualClassroomConn).
		RetrieveWhiteboardToken(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userReceivesRoomIDAndWhiteboardToken(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.RetrieveWhiteboardTokenResponse)

	if len(response.RoomId) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing room id when user retrieves whiteboard token")
	}

	if len(response.WhiteboardToken) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard token when user retrieves whiteboard token")
	}

	if len(response.WhiteboardAppId) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("missing whiteboard app ID when user retrieves whiteboard token")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) lessonDoesNotHaveExistingRoomID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonID := stepState.CurrentLessonID
	query := `UPDATE lessons 
		SET room_id = NULL, updated_at = now()
		WHERE lesson_id = $1`

	_, err := s.LessonmgmtDB.Exec(ctx, query, &lessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to update room ID of lesson %s db.Exec: %w", lessonID, err)
	}

	return StepStateToContext(ctx, stepState), nil
}
