package studentstudyplan

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/repositories"
)

func (s *Suite) userCallFindStudentStudyPlanWithCourseIDs(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studentStudyPlanRepo := repositories.StudentStudyPlanRepo{}
	courseIDs := make([]string, 0, len(stepState.StudentIDs))
	for range stepState.StudentIDs {
		courseIDs = append(courseIDs, stepState.CourseID)
	}
	studyPlanIDs, err := studentStudyPlanRepo.FindStudentStudyPlanWithCourseIDs(ctx, s.DB, stepState.StudentIDs, courseIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("studentStudyPlanRepo.ListStudentAvailableContents: %w", err)
	}
	stepState.studyPlanIDsResponse = studyPlanIDs

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnAStudentStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	var expectedStudyPlanIDs []string
	for i := 0; i < stepState.NumberOfStudentStudyPlansAdded; i++ {
		expectedStudyPlanIDs = append(expectedStudyPlanIDs, stepState.StudyPlanIDs[i])
	}

	if len(expectedStudyPlanIDs) != len(stepState.studyPlanIDsResponse) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected StudyPlanIDs length: expected %d, got %d", len(stepState.StudyPlanIDs), len(stepState.studyPlanIDsResponse))
	}

	for i, studyPlanID := range expectedStudyPlanIDs {
		if studyPlanID != stepState.studyPlanIDsResponse[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected StudyPlanIDs: expected %v, got %v", stepState.StudyPlanIDs, stepState.studyPlanIDsResponse)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
