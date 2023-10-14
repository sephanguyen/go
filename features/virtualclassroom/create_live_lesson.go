package virtualclassroom

import (
	"context"
	"fmt"
	"time"
)

func (s *suite) anExistingVirtualClassroom(ctx context.Context) (context.Context, error) {
	ctx, err := s.CommonSuite.UserCreateALiveLessonWithMissingFields(ctx)
	stepState := StepStateFromContext(ctx)
	if err != nil || stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to create lesson, err: %w | response err: %w", err, stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) anExistingVirtualClassroomWithWait(ctx context.Context) (context.Context, error) {
	// sleep to make sure NATS sync data successfully from bob to lessonmgmt data
	time.Sleep(5 * time.Second)

	ctx, err := s.anExistingVirtualClassroom(ctx)

	// wait to sync from lessonmgmt to bob data
	time.Sleep(3 * time.Second)

	return ctx, err
}

func (s *suite) existingVirtualClassrooms(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numberOfLessons := 3

	// sleep to make sure NATS sync data successfully from bob to lessonmgmt data
	time.Sleep(5 * time.Second)

	for i := 0; i < numberOfLessons; i++ {
		ctx, err := s.anExistingVirtualClassroom(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	// wait to sync from lessonmgmt to bob data
	time.Sleep(3 * time.Second)

	return StepStateToContext(ctx, stepState), nil
}
