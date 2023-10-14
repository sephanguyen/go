package studentstudyplan

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/manabie-com/backend/features/repository/syllabus/entity"
	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/hasura/go-graphql-client"
)

func (s *Suite) someStudentsRegisterToTheCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.CourseIDs = append(stepState.CourseIDs, stepState.CourseID)
	courseStudents, err := utils.AUserInsertSomeCourseStudentToDatabase(ctx, s.DB, stepState.CourseIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.AUserInsertSomeCourseStudentToDatabase: %w", err)
	}
	for _, courseStudent := range courseStudents {
		stepState.StudentIDs = append(stepState.StudentIDs, courseStudent.StudentID.String)
	}
	stepState.StudentID = stepState.StudentIDs[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

func withCourseID(id string) utils.StudyPlanOption {
	return func(u *entities.StudyPlan) error {
		err := u.CourseID.Set(id)
		return err
	}
}

func (s *Suite) someValidStudyPlans(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.CourseID = idutil.ULIDNow()
	studyPlans, err := utils.AUserInsertSomeStudyPlanToDatabase(ctx, s.DB, withCourseID(stepState.CourseID))
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

func (s *Suite) aUserInsertedSomeStudentStudyPlansToDatabase(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := len(stepState.StudyPlanIDs)
	for i := 0; i < n; i++ {
		stepState.StudentIDs = append(stepState.StudentIDs, idutil.ULIDNow())
	}
	studentStudyPlans, err := utils.AUserInsertSomeStudentStudyPlanToDatabase(ctx, s.DB, stepState.StudentIDs, stepState.StudyPlanIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.StudentStudyPlans = studentStudyPlans
	stepState.NumberOfStudentStudyPlansAdded = n
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
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.AUserInsertSomeStudyPlanItemToDatabase: %w", err)
	}
	for _, studyPlanItem := range studyPlanItems {
		stepState.StudyPlanItemIDs = append(stepState.StudyPlanItemIDs, studyPlanItem.ID.String)
	}
	stepState.StudyPlanItems = studyPlanItems
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

func (s *Suite) thereAreLoStudyPlanItemsExistedInStudyPlan(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	loStudyPlanItems, err := utils.AUserInsertSomeLoStudyPlanItemsToDatabase(ctx, s.DB, stepState.StudyPlanItemIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.AUserInsertSomeLoStudyPlanItemsToDatabase: %w", err)
	}
	for _, loStudyPlanItem := range loStudyPlanItems {
		stepState.LoIDs = append(stepState.LoIDs, loStudyPlanItem.LoID.String)
	}
	stepState.LoStudyPlanItems = loStudyPlanItems

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnStudentStudyPlanCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if len(stepState.StudentStudyPlanQuery.StudentStudyPlans) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found student study plan")
	}
	actualStudentStudyPlan := stepState.StudentStudyPlanQuery.StudentStudyPlans[0]

	if actualStudentStudyPlan.StudentID != stepState.StudentID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study course ids: want: %v, actual: %v", stepState.StudentID, actualStudentStudyPlan.StudentID)
	}
	actualStudyPlan := entity.GetStudyPlan(&stepState.StudentStudyPlanQuery)
	expectedStudyPlan := stepState.StudyPlan
	err := utils.CompareStudyPlan(expectedStudyPlan, actualStudyPlan)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	// if actualStudyPlan
	expectedStudyPlanItems := stepState.StudyPlanItems
	actualStudyPlanItems := entity.GetStudyPlanItems(&stepState.StudentStudyPlanQuery)
	err = utils.CompareStudyPlanItem(expectedStudyPlanItems, actualStudyPlanItems)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	actualLOIDs := make([]string, 0, len(actualStudyPlanItems))
	actualAssignmentIDs := make([]string, 0, len(actualStudyPlanItems))
	for _, item := range stepState.StudentStudyPlanQuery.StudentStudyPlans[0].StudyPlan.StudyPlanItems {
		actualLOIDs = append(actualLOIDs, item.GetLoID())
		actualAssignmentIDs = append(actualAssignmentIDs, item.GetAssignmentID())
	}

	if !reflect.DeepEqual(actualLOIDs, stepState.LoIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected lo ids: want %s, got %s", stepState.LoIDs, actualLOIDs)
	}
	if !reflect.DeepEqual(actualAssignmentIDs, stepState.AssignmentIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected assignment ids: want %s, got %s", stepState.AssignmentIDs, actualAssignmentIDs)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemReturnStudentStudyPlansCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	if len(stepState.StudentStudyPlansManyV2Query.StudentStudyPlans) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found student study plan")
	}
	fmt.Println(len(stepState.StudentStudyPlansManyV2Query.StudentStudyPlans))
	for i := 0; i < len(stepState.StudentStudyPlansManyV2Query.StudentStudyPlans)-1; i++ {
		if stepState.StudentStudyPlansManyV2Query.StudentStudyPlans[i].StudyPlan.CreatedAt.Before(stepState.StudentStudyPlansManyV2Query.StudentStudyPlans[i+1].StudyPlan.CreatedAt) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("student study plans is not sorted by created at")
		}
	}
	utils.Reverse(stepState.StudentStudyPlansManyV2Query.StudentStudyPlans)
	for i, actualStudentStudyPlan := range stepState.StudentStudyPlansManyV2Query.StudentStudyPlans {
		if actualStudentStudyPlan.StudyPlanID != stepState.StudyPlanIDs[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected study plan id in study plan: want %s, got %s", stepState.StudyPlanIDs[i], actualStudentStudyPlan.StudyPlanID)
		}

		err := checkStudyPlanV2(stepState.StudyPlans[i], actualStudentStudyPlan.StudyPlan.StudyPlanAttrsV2)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallStudentStudyPlansByCourseId(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	variables := map[string]interface{}{
		"course_id":     graphql.String(stepState.CourseID),
		"study_plan_id": graphql.String(stepState.StudyPlanID),
	}

	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.StudentStudyPlanQuery, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.QueryHasura: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallStudentStudyPlansManyV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	studentIDs := []graphql.String{}
	for _, studentID := range stepState.StudentIDs {
		studentIDs = append(studentIDs, graphql.String(studentID))
	}
	variables := map[string]interface{}{
		"course_id":   graphql.String(stepState.CourseID),
		"student_ids": studentIDs,
		"status":      graphql.String("STUDY_PLAN_STATUS_ACTIVE"),
	}

	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.StudentStudyPlansManyV2Query, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("utils.QueryHasura: %w", err)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func checkStudyPlanV2(expectedStudyPlan *entities.StudyPlan, actualStudyPlan entity.StudyPlanAttrsV2) error {
	if actualStudyPlan.Name != expectedStudyPlan.Name.String {
		return fmt.Errorf("unexpected study plan name in study plan att: want %s, got %s", expectedStudyPlan.Name.String, actualStudyPlan.Name)
	}
	if actualStudyPlan.StudyPlanID != expectedStudyPlan.ID.String {
		return fmt.Errorf("unexpected study plan id in study plan att: want %s, got %s", expectedStudyPlan.ID.String, actualStudyPlan.StudyPlanID)
	}
	if actualStudyPlan.CreatedAt.Sub(expectedStudyPlan.CreatedAt.Time) > time.Second {
		return fmt.Errorf("unexpected created at: want %s, got %s", expectedStudyPlan.CreatedAt.Time, actualStudyPlan.CreatedAt)
	}
	if actualStudyPlan.MasterStudyPlanID != expectedStudyPlan.MasterStudyPlan.String {
		return fmt.Errorf("unexpected master study plan id: want %s, got %s", expectedStudyPlan.MasterStudyPlan.String, actualStudyPlan.MasterStudyPlanID)
	}
	if actualStudyPlan.BookID != expectedStudyPlan.BookID.String {
		return fmt.Errorf("unexpected book id: want %s, got %s", expectedStudyPlan.BookID.String, actualStudyPlan.BookID)
	}
	var expectedGrades []int64
	expectedStudyPlan.Grades.AssignTo(expectedGrades)
	if !reflect.DeepEqual(actualStudyPlan.Grades, expectedGrades) {
		return fmt.Errorf("unexpected grades: want %v, got %v", expectedGrades, actualStudyPlan.Grades)
	}
	return nil
}
