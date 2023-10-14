package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	noti_entities "github.com/manabie-com/backend/internal/notification/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

type Student struct {
	StudentID    string
	StudentToken string
}

func (s *suite) aListStudentsLogins(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	n := rand.Intn(3) + 2
	studentIDs := make([]string, 0, n)
	students := make([]*Student, 0, n)
	for i := 0; i <= n; i++ {
		if ctx, err := s.logins(ctx, studentRawText); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		studentIDs = append(studentIDs, stepState.StudentID)
		students = append(students, &Student{
			StudentID:    stepState.StudentID,
			StudentToken: stepState.StudentToken,
		})
	}

	stepState.StudentIDs = studentIDs
	stepState.Students = students
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addStudentsToCourse(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, studentID := range stepState.StudentIDs {
		stepState.StudentID = studentID
		if ctx, err := s.userAddCourseToStudent(ctx); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) gradeSubmissionsWithStatusInProgress(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.TeacherToken
	ctx = contextWithToken(s, ctx)
	for _, submissionID := range stepState.SubmissionIDs {
		if _, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).GradeStudentSubmission(ctx, &epb.GradeStudentSubmissionRequest{
			Grade: &epb.SubmissionGrade{
				SubmissionId: submissionID,
			},
			Status: epb.SubmissionStatus_SUBMISSION_STATUS_IN_PROGRESS,
		}); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) notificationsHasBeenStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &noti_entities.UserInfoNotification{}
	stmt := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM %s
	WHERE user_id = ANY($1)
	`, e.TableName())
	count := 0
	if err := try.Do(func(attempt int) (retry bool, err error) {
		if err := s.BobDB.QueryRow(ctx, stmt, stepState.StudentIDs).Scan(&count); err != nil {
			return false, err
		}
		if count == 0 {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("notifications is not sent")
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != len(stepState.StudentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("notifications is not sent, expect %d but got %d", len(stepState.StudentIDs), count)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentsDoAssignment(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	submissionIDs := make([]string, 0, len(stepState.Students))
	for _, student := range stepState.Students {
		stepState.StudentID = student.StudentID
		stepState.StudentToken = student.StudentToken
		if ctx, err := s.doAssignment(ctx, studentRawText); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		submissionIDs = append(submissionIDs, stepState.SubmissionID)
	}
	stepState.SubmissionIDs = submissionIDs
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStudentSubmissionsStatusToReturned(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.TeacherToken
	ctx = contextWithToken(s, ctx)
	if _, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).UpdateStudentSubmissionsStatus(ctx, &epb.UpdateStudentSubmissionsStatusRequest{
		SubmissionIds: stepState.SubmissionIDs,
		Status:        epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
