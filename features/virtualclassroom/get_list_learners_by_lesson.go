package virtualclassroom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/bxcodec/faker/v3/support/slice"
)

func (s *suite) studentsHaveEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	studentIDs := stepState.StudentIds
	locationID := stepState.CenterIDs[len(stepState.CenterIDs)-1]
	now := time.Now()
	startDate := now.Add(-1 * 24 * time.Hour)
	endDate := now.Add(2 * 24 * time.Hour)

	for _, studentID := range studentIDs {
		query := `INSERT INTO student_enrollment_status_history(student_id, location_id, enrollment_status, start_date, end_date)
			VALUES ($1, $2, 'STUDENT_ENROLLMENT_STATUS_ENROLLED', $3, $4)`

		_, err := s.BobDB.Exec(ctx, query, studentID, locationID, startDate, endDate)
		if err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("create student enrollment status history %s db.Exec: %w", studentID, err)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsListOfLearnersInLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetLearnersByLessonID(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.GetLearnersByLessonIDRequest{
			LessonId: stepState.CurrentLessonID,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfStudentsWithEnrollmentStatus(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.GetLearnersByLessonIDResponse)
	learners := response.GetLearners()
	if len(learners) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting learners but got 0")
	}

	expectedStudentIDs := stepState.StudentIds
	expectedLocationID := stepState.CenterIDs[len(stepState.CenterIDs)-1]
	for _, learner := range learners {
		if !slice.Contains(expectedStudentIDs, learner.LearnerId) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s is not expected in the list %v", learner.LearnerId, expectedStudentIDs)
		}

		learnerEnrollmentInfos := learner.EnrollmentStatusInfo
		if len(learnerEnrollmentInfos) == 0 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expecting learner %s has enrollment info but got none", learner.LearnerId)
		}

		for _, learnerEnrollmentInfo := range learnerEnrollmentInfos {
			if learnerEnrollmentInfo.LocationId != expectedLocationID {
				return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s enrollment info location %s does not match with lesson location %s",
					learner.LearnerId,
					learnerEnrollmentInfo.LocationId,
					expectedLocationID,
				)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
