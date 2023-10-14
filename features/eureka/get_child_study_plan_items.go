package eureka

import (
	"context"
	"fmt"
	"math/rand"

	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

// aCourseAndAssignedThisCourseToSomeStudents:
// - a course, a course_study_plan record, study_plan(1), SOME study_plan_item record(s) (2) will created (*)
// - with a root(master) study_plan above, our system will clone to n records with master_study_plan is (1).study_plan_id
// - with an arbitrary (2) study_plan_item, our system will created for accordinate each students with the copy_study_plan_item_id is (2).study_plan_item_id
// - p/s: (*) a study_plans can have one or more study_plan_items
func (s *suite) aCourseAndAssignedThisCourseToSomeStudents(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.someStudentsAreAssignedSomeValidStudyPlans(ctx)
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) teacherGetChildStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// random idx of study_plan_item root
	idx := rand.Intn(len(stepState.StudyPlanItemIDs) - 1)
	stepState.StudyPlanItemID = stepState.StudyPlanItemIDs[idx]
	req := &pb.GetChildStudyPlanItemsRequest{
		StudyPlanItemId: stepState.StudyPlanItemIDs[idx],
		UserIds:         stepState.StudentIDs,
	}
	stepState.Response, stepState.ResponseErr = pb.NewAssignmentReaderServiceClient(s.Conn).GetChildStudyPlanItems(contextWithToken(s, ctx), req)
	stepState.Request = req
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) ourSystemHaveToReturnChildStudyPlanItemsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}
	res := stepState.Response.(*pb.GetChildStudyPlanItemsResponse)
	if len(res.GetItems()) != len(stepState.StudentIDs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected number of user study plan item, expected: %d, actual: %d", len(stepState.StudentIDs), len(res.GetItems()))
	}
	req := stepState.Request.(*pb.GetChildStudyPlanItemsRequest)
	counter := 0
	for _, id := range req.UserIds {
		for _, item := range res.GetItems() {
			if item.UserId == id {
				counter++
			}
		}
	}
	if counter != len(res.GetItems()) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected response of study plan item")
	}
	return StepStateToContext(ctx, stepState), nil
}
