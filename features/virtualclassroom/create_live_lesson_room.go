package virtualclassroom

import (
	"context"
	"fmt"
	"time"
)

func (s *suite) userCreatesAVirtualClassroomSession(ctx context.Context) (context.Context, error) {
	// sleep to make sure NATS sync data successfully from bob to lessonmgmt data
	time.Sleep(5 * time.Second)

	return s.CommonSuite.UserCreateALiveLessonWithMissingFields(ctx)
}

func (s *suite) lessonHasExistingRoomIDWithWait(ctx context.Context) (context.Context, error) {
	// wait for sync process done
	time.Sleep(10 * time.Second)

	return s.lessonHasExistingRoomID(ctx)
}

func (s *suite) lessonHasExistingRoomID(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonID := stepState.CurrentLessonID
	var roomID string

	query := `SELECT room_id 
		FROM lessons 
		WHERE lesson_id = $1 `

	if err := s.LessonmgmtDB.QueryRow(ctx, query, lessonID).Scan(&roomID); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get room ID of lesson %s db.QueryRow: %w", lessonID, err)
	}

	if len(roomID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting room ID for lesson %s but got empty ID %s", lessonID, roomID)
	}

	return StepStateToContext(ctx, stepState), nil
}
