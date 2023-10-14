package bob

import (
	"context"
)

func (s *suite) studentsClassIsRemovedFromCourse(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := "UPDATE courses_classes SET deleted_at = NOW(), status = 'COURSE_CLASS_STATUS_INACTIVE' WHERE class_id =$1"
	_, err := s.DB.Exec(ctx, query, stepState.CurrentClassID)
	return StepStateToContext(ctx, stepState), err
}
