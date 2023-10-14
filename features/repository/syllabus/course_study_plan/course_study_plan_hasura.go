package csp

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"

	"github.com/manabie-com/backend/features/repository/syllabus/entity"
	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/hasura/go-graphql-client"
)

func (s *Suite) someValidStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	studyPlans, err := utils.AUserInsertSomeStudyPlanToDatabase(ctx, s.DB)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	for _, studyPlan := range studyPlans {
		stepState.StudyPlanIDs = append(stepState.StudyPlanIDs, studyPlan.ID.String)
	}
	stepState.StudyPlanID = stepState.StudyPlanIDs[0]
	stepState.StudyPlans = studyPlans
	stepState.StudyPlan = studyPlans[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aUserInsertedSomeCourseStudyPlansToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	courseID := idutil.ULIDNow()
	courseStudyPlans, err := utils.AUserInsertSomeCourseStudyPlansToDatabase(ctx, s.DB, courseID, stepState.StudyPlanIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.CourseID = courseID
	stepState.CourseStudyPlans = courseStudyPlans
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallCourseStudyPlansList(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Limit = rand.Intn(5) + 3
	stepState.OffSet = rand.Intn(2)
	variables := map[string]interface{}{
		"course_id": graphql.String(stepState.CourseID),
		"limit":     graphql.Int(stepState.Limit),
		"offset":    graphql.Int(stepState.OffSet),
	}
	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.CourseStudyPlansListQuery, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallCourseStudyPlansByCourseID(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	variables := map[string]interface{}{
		"course_id":     graphql.String(stepState.CourseID),
		"study_plan_id": graphql.String(stepState.StudyPlanID),
	}
	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.CourseStudyPlansByCourseIDQuery, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnCourseStudyPlansCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.CourseStudyPlans) != stepState.CourseStudyPlansListQuery.CourseStudyPlanAggregate.Aggregate.Count {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("count return incorrect value, want: %v, actual: %v", len(stepState.CourseStudyPlans), stepState.CourseStudyPlansListQuery.CourseStudyPlanAggregate.Aggregate.Count)
	}
	// course study plans after order by desc
	utils.Reverse(stepState.CourseStudyPlans)
	// course study plans after offset
	stepState.CourseStudyPlans = stepState.CourseStudyPlans[stepState.OffSet:]
	// course study plans after limit
	stepState.CourseStudyPlans = stepState.CourseStudyPlans[:stepState.Limit]
	actualNumOfCourseStudyPlans := len(stepState.CourseStudyPlansListQuery.CourseStudyPlans)
	if actualNumOfCourseStudyPlans == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found any course study plans with CourseID %s", stepState.CourseID)
	}

	expectedNumOfCourseStudyPlans := len(stepState.CourseStudyPlans)
	if expectedNumOfCourseStudyPlans < actualNumOfCourseStudyPlans {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("missing %v course study plans when query from hasura ", expectedNumOfCourseStudyPlans-actualNumOfCourseStudyPlans)
	}

	for i := 0; i < actualNumOfCourseStudyPlans; i++ {
		actualCourseID := stepState.CourseStudyPlansListQuery.CourseStudyPlans[i].CourseID
		expectedCourseID := stepState.CourseStudyPlans[i].CourseID.String
		if actualCourseID != expectedCourseID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected courser id: want: %v, actual: %v", expectedCourseID, actualCourseID)
		}

		actualStudyPlanID := stepState.CourseStudyPlansListQuery.CourseStudyPlans[i].StudyPlanID
		expectedStudyPlanID := stepState.CourseStudyPlans[i].StudyPlanID.String
		if actualStudyPlanID != expectedStudyPlanID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan id: want: %v, actual: %v", expectedStudyPlanID, actualStudyPlanID)
		}

		actualStudyPlanName := stepState.CourseStudyPlansListQuery.CourseStudyPlans[i].StudyPlan.Name
		expectedStudyPlanName := utils.FormatName(stepState.CourseStudyPlans[i].StudyPlanID.String)
		if actualStudyPlanName != expectedStudyPlanName {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan name: want: %v, actual: %v", expectedStudyPlanName, actualStudyPlanName)
		}

		actualStudyPlanMasterID := stepState.CourseStudyPlansListQuery.CourseStudyPlans[i].StudyPlan.MasterStudyPlanID
		expectedStudyPlanMasterID := stepState.CourseStudyPlans[i].StudyPlanID.String
		if actualStudyPlanMasterID != expectedStudyPlanMasterID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan master id: want: %v, actual: %v", expectedStudyPlanMasterID, actualStudyPlanMasterID)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func withStudyPlanID(id string) utils.StudyPlanItemOption {
	return func(u *entities.StudyPlanItem) error {
		err := u.StudyPlanID.Set(id)
		return err
	}
}

func (s *Suite) thereAreStudyPlanItemsExistedInStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studyPlanItems, err := utils.AUserInsertSomeStudyPlanItemsToDatabaseWithStudyPlanID(ctx, s.DB, withStudyPlanID(stepState.StudyPlanID))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	for _, studyPlanItem := range studyPlanItems {
		stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, studyPlanItem.ID.String)
	}

	stepState.StudyPlanItemID = studyPlanItems[0].ID.String
	stepState.StudyPlanItems = studyPlanItems
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreLoStudyPlanItemsExistedInStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	loStudyPlanItems, err := utils.AUserInsertSomeLoStudyPlanItemsToDatabase(ctx, s.DB, stepState.StudyPlanItemIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	for _, loStudyPlanItem := range loStudyPlanItems {
		stepState.LoStudyPlanItemIDs = append(stepState.LoStudyPlanItemIDs, loStudyPlanItem.LoID.String)
	}
	stepState.LoStudyPlanItems = loStudyPlanItems
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) thereAreAssignmentStudyPlanItemsExistedInStudyPlanItems(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	if len(stepState.AssignmentIDs) == 0 {
		assignments, err := utils.AUserInsertSomeAssignmentsToDatabase(ctx, s.DB, len(stepState.StudyPlanItemIDs))
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.AUserInsertSomeAssignmentsToDatabase: %w", err)
		}
		for _, assignment := range assignments {
			stepState.AssignmentIDs = append(stepState.AssignmentIDs, assignment.ID.String)
		}
	}
	assignmentStudyPlanItems, err := utils.AUserInsertSomeAssignmentStudyPlanItemsToDatabase(ctx, s.DB, stepState.AssignmentIDs, stepState.StudyPlanItemIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.AssignmentStudyPlanItems = assignmentStudyPlanItems
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnCourseStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if len(stepState.CourseStudyPlansByCourseIDQuery.CourseStudyPlans) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found any course study plans with CourseID %s", stepState.CourseID)
	}
	actualCourseID := stepState.CourseStudyPlansByCourseIDQuery.CourseStudyPlans[0].CourseID
	if actualCourseID != stepState.CourseID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study course ids: want: %v, actual: %v", stepState.CourseID, actualCourseID)
	}

	actualStudyPlan := entity.GetStudyPlan(&stepState.CourseStudyPlansByCourseIDQuery)
	expectedStudyPlan := stepState.StudyPlan
	err := utils.CompareStudyPlan(expectedStudyPlan, actualStudyPlan)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	actualStudyPlanItems := entity.GetStudyPlanItems(&stepState.CourseStudyPlansByCourseIDQuery)
	expectedStudyPlanItems := stepState.StudyPlanItems
	err = utils.CompareStudyPlanItem(expectedStudyPlanItems, actualStudyPlanItems)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	actualLOIDs := make([]string, 0, len(actualStudyPlanItems))
	actualAssignmentIDs := make([]string, 0, len(actualStudyPlanItems))
	for _, item := range stepState.CourseStudyPlansByCourseIDQuery.CourseStudyPlans[0].StudyPlan.StudyPlanItems {
		actualLOIDs = append(actualLOIDs, item.GetLoID())
		actualAssignmentIDs = append(actualAssignmentIDs, item.GetAssignmentID())
	}

	if !reflect.DeepEqual(actualLOIDs, stepState.LoStudyPlanItemIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected lo ids: want %s, got %s", stepState.LoStudyPlanItemIDs, actualLOIDs)
	}
	if !reflect.DeepEqual(actualAssignmentIDs, stepState.AssignmentIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected assignment ids: want %s, got %s", stepState.AssignmentIDs, actualAssignmentIDs)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
