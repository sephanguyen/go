package common

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/jackc/pgx/v4"
)

func (s *suite) SomeStudentSubscriptions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	courseID := stepState.CourseIDs[len(stepState.CourseIDs)-1]
	studentIDWithCourseID := make([]string, 0, len(stepState.StudentIds)*2)
	for _, studentID := range stepState.StudentIds {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	stepState.StartDate = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	stepState.EndDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	ids, err := s.insertStudentSubscription(ctx, stepState.StartDate, stepState.EndDate, studentIDWithCourseID...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}
	stepState.StudentIDWithCourseID = studentIDWithCourseID

	// create access path for above list student subscriptions
	for _, l := range stepState.LocationIDs {
		for _, id := range ids {
			stmt := `INSERT INTO lesson_student_subscription_access_path (student_subscription_id,location_id) VALUES($1,$2)`
			_, err := s.LessonmgmtDB.Exec(ctx, stmt, id, l)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_student_subscription_access_path with student_subscription_id:%s, location_id:%s, err:%v", id, l, err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) SomeStudentSubscriptionsWithParams(ctx context.Context, _startAt, _endAt string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	startAt, _ := time.Parse(time.RFC3339, _startAt)
	endAt, _ := time.Parse(time.RFC3339, _endAt)
	stepState.EndDate = endAt
	courseID := stepState.CourseIDs[len(stepState.CourseIDs)-1]
	studentIDWithCourseID := make([]string, 0, len(stepState.StudentIds)*2)

	for _, studentID := range stepState.StudentIds {
		studentIDWithCourseID = append(studentIDWithCourseID, studentID, courseID)
	}
	ids, err := s.insertStudentSubscription(ctx, startAt, endAt, studentIDWithCourseID...)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("could not insert student subscription: %w", err)
	}
	stepState.StudentIDWithCourseID = studentIDWithCourseID

	// create access path for above list student subscriptions
	for _, l := range stepState.LocationIDs {
		for _, id := range ids {
			stmt := `INSERT INTO lesson_student_subscription_access_path (student_subscription_id,location_id) VALUES($1,$2)`
			_, err := s.LessonmgmtDB.Exec(ctx, stmt, id, l)
			if err != nil {
				return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert lesson_student_subscription_access_path with student_subscription_id:%s, location_id:%s, err:%v", id, l, err)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) insertStudentSubscription(ctx context.Context, startAt, endAt time.Time, studentIDWithCourseID ...string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	queueFn := func(b *pgx.Batch, studentID, courseID string) string {
		id := idutil.ULIDNow()
		query := `INSERT INTO lesson_student_subscriptions (student_subscription_id, subscription_id, student_id, course_id, start_at, end_at) VALUES ($1, $2, $3, $4, $5, $6)`
		b.Queue(query, id, id, studentID, courseID, startAt, endAt)
		return id
	}

	b := &pgx.Batch{}
	ids := make([]string, 0, len(studentIDWithCourseID))
	for i := 0; i < len(studentIDWithCourseID); i += 2 {
		ids = append(ids, queueFn(b, studentIDWithCourseID[i], studentIDWithCourseID[i+1]))
	}
	result := s.LessonmgmtDB.SendBatch(ctx, b)
	defer result.Close()

	for i, iEnd := 0, b.Len(); i < iEnd; i++ {
		_, err := result.Exec()
		if err != nil {
			return nil, fmt.Errorf("result.Exec[%d]: %w", i, err)
		}
	}
	return ids, nil
}
