package eureka

import (
	"context"
	"fmt"
	"time"

	bob_entities "github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"
	noti_entities "github.com/manabie-com/backend/internal/notification/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

func (s *suite) addStudentToCourse(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if ctx, err := s.userAddCourseToStudent(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createAStudyPlanWithBookHaveAnAssignment(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)
	if ctx, err := s.createBook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.schoolAdminCreateAtopicAndAChapter(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	assignments := s.prepareAssignment(ctx, stepState.TopicID, 1)
	if _, err := epb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(ctx, &epb.UpsertAssignmentsRequest{
		Assignments: assignments,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.AssignmentID = assignments[0].AssignmentId
	if ctx, err := s.createACourse(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if ctx, err := s.createStudyPlanFromTheBook(ctx); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) doAssignment(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.StudentToken
	ctx = contextWithToken(s, ctx)
	spie := &entities.StudyPlanItem{}
	sspe := &entities.StudentStudyPlan{}
	stmt := fmt.Sprintf(`
		SELECT study_plan_item_id
		FROM %s
		JOIN %s as ssp
		USING(study_plan_id)
		WHERE content_structure ->> 'assignment_id' = $1 
		AND copy_study_plan_item_id IS NOT NULL
		AND ssp.student_id = $2
	`, spie.TableName(), sspe.TableName())
	var studyPlanItemID string
	if err := try.Do(func(attempt int) (retry bool, err error) {
		if err := s.DB.QueryRow(ctx, stmt, stepState.AssignmentID, stepState.StudentID).Scan(&studyPlanItemID); err != nil {
			if err == pgx.ErrNoRows {
				time.Sleep(2 * time.Second)
				return attempt < 5, err
			}
			return false, fmt.Errorf("unable to get study plan item: %w", err)
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	resp, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).SubmitAssignment(ctx, &epb.SubmitAssignmentRequest{
		Submission: &epb.StudentSubmission{
			AssignmentId:    stepState.AssignmentID,
			StudyPlanItemId: studyPlanItemID,
			StudentId:       stepState.StudentID,
		},
	})

	stepState.SubmissionID = resp.SubmissionId
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to submit assignment: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) gradeSubmissionWithStatusReturned(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.TeacherToken
	ctx = contextWithToken(s, ctx)

	if _, err := epb.NewStudentAssignmentWriteServiceClient(s.Conn).GradeStudentSubmission(ctx, &epb.GradeStudentSubmissionRequest{
		Grade: &epb.SubmissionGrade{
			SubmissionId: stepState.SubmissionID,
		},
		Status: epb.SubmissionStatus_SUBMISSION_STATUS_RETURNED,
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) notificationHasBeenStoredCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// inject resource path for nats js stream
	ctxInjectResourcePath := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
		},
	})
	e := &noti_entities.UserInfoNotification{}
	stmt := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM %s
	WHERE user_id = $1
	`, e.TableName())
	count := 0
	if err := try.Do(func(attempt int) (retry bool, err error) {
		if err := s.BobDB.QueryRow(ctxInjectResourcePath, stmt, stepState.StudentID).Scan(&count); err != nil {
			return false, err
		}
		if count == 0 {
			time.Sleep(2 * time.Second)
			return attempt < 10, fmt.Errorf("notification is not sent")
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count != 1 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("notification is not sent, expect 1 but got %d", count)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) notificationHasNotBeenStored(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	time.Sleep(5 * time.Second)
	ctxInjectResourcePath := interceptors.ContextWithJWTClaims(ctx, &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: "1",
		},
	})
	e := &noti_entities.UserInfoNotification{}
	stmt := fmt.Sprintf(`
	SELECT COUNT(*)
	FROM %s
	WHERE user_id = $1
	`, e.TableName())

	var count pgtype.Int8
	if err := s.BobDB.QueryRow(ctxInjectResourcePath, stmt, stepState.StudentID).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if count.Int != 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("notification is still sent")
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateStudentSchoolToNull(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := bob_entities.Student{}
	stmt := fmt.Sprintf("UPDATE %s SET school_id = NULL WHERE student_id = $1", e.TableName())

	if _, err := s.BobDB.Exec(ctx, stmt, &stepState.StudentID); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
