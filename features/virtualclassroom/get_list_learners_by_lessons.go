package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/bxcodec/faker/v3/support/slice"
)

func (s *suite) userGetsListOfLearnersFromLessons(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonIDs := stepState.LessonIDs
	if len(lessonIDs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("no lesson IDs are found for query")
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetLearnersByLessonIDs(helper.GRPCContext(ctx, "token", stepState.AuthToken), &vpb.GetLearnersByLessonIDsRequest{
			LessonId: lessonIDs,
		})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAListOfStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.GetLearnersByLessonIDsResponse)

	lessonLearners := response.GetLessonLearners()
	if len(lessonLearners) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expecting lesson learners but got 0")
	}

	expectedStudentIDs := stepState.StudentIds
	expectedLessonIDs := stepState.LessonIDs
	for _, lessonLearner := range lessonLearners {
		if !slice.Contains(expectedLessonIDs, lessonLearner.LessonId) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s is not expected in the lesson list %v", lessonLearner.LessonId, expectedLessonIDs)
		}

		for _, learner := range lessonLearner.Learners {
			if !slice.Contains(expectedStudentIDs, learner.LearnerId) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s from lesson %s is not expected in the learner list %v", learner.LearnerId, lessonLearner.LessonId, expectedStudentIDs)
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
