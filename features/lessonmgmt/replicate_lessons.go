package lessonmgmt

import (
	"context"
	"fmt"
	"time"
)

func (s *Suite) lessonsSyncedToLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for sync process done
	time.Sleep(5 * time.Second)
	var count int
	query := "SELECT count(lesson_id) from lessons where lesson_id = any($1)"
	err := s.LessonmgmtDBTrace.QueryRow(ctx, query, stepState.LessonIDs).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(stepState.LessonIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found lesson ids: `%s`", stepState.LessonIDs)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonTeachersSyncedToLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for sync process done
	time.Sleep(5 * time.Second)
	var count int
	query := "SELECT count(teacher_id) from lessons_teachers where teacher_id = any($1)"
	err := s.LessonmgmtDBTrace.QueryRow(ctx, query, stepState.TeacherIDs).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != len(stepState.TeacherIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found teacher ids: `%s`", stepState.TeacherIDs)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) lessonMembersSyncedToLessonmgmt(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for sync process done
	time.Sleep(5 * time.Second)
	var userID string
	studentID := stepState.StudentIDWithCourseID[0]
	query := "SELECT user_id from lesson_members where lesson_id = $1 and user_id = $2"
	err := s.LessonmgmtDBTrace.QueryRow(ctx, query, stepState.CurrentLessonID, studentID).Scan(&userID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if userID != studentID {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found member `%s` in lesson `%s`", studentID, stepState.CurrentLessonID)
	}
	return StepStateToContext(ctx, stepState), nil
}
