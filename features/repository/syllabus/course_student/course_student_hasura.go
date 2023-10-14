package course_student

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/hasura/go-graphql-client"
)

func (s *Suite) someStudentAssignedToCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 10
	stepState.CourseID = idutil.ULIDNow()
	stepState.CourseIDs = append(stepState.CourseIDs, stepState.CourseID)
	for i := 0; i < n; i++ {
		courseStudents, err := utils.AUserInsertSomeCourseStudentToDatabase(ctx, s.DB, stepState.CourseIDs)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		for _, courseStudent := range courseStudents {
			stepState.StudentIDs = append(stepState.StudentIDs, courseStudent.StudentID.String)
		}
	}
	stepState.StudentID = stepState.StudentIDs[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) someStudentAssignedToManyCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	n := rand.Intn(5) + 10
	for i := 0; i < n; i++ {
		stepState.CourseIDs = append(stepState.CourseIDs, idutil.ULIDNow())
	}
	courseStudents, err := utils.AUserInsertSomeCourseStudentToDatabase(ctx, s.DB, stepState.CourseIDs)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	for _, courseStudent := range courseStudents {
		stepState.StudentIDs = append(stepState.StudentIDs, courseStudent.StudentID.String)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallCourseStudentsByCourseIds(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var stepStateCourseIds = make([]graphql.String, 0, len(stepState.CourseIDs))
	for _, courseId := range stepState.CourseIDs {
		stepStateCourseIds = append(stepStateCourseIds, graphql.String(courseId))
	}
	variables := map[string]interface{}{
		"course_ids": stepStateCourseIds,
	}

	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.CourseStudentListByCourseIDQuery, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnCourseStudentsByCourseIDsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	actualCourseStudents := stepState.CourseStudentListByCourseIDQuery.CourseStudentsListByCourseIds
	actualCount := stepState.CourseStudentListByCourseIDQuery.CourseStudentsAggregate.Aggregate.Count

	if len(stepState.StudentIDs) != actualCount {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected count: want: %d, actual : %d", len(stepState.StudentIDs), actualCount)
	}
	if (stepState.StudentIDs[0]) != actualCourseStudents[len(actualCourseStudents)-1].StudentID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected studentId: want: %s, actual : %s", stepState.StudentIDs[0], actualCourseStudents[len(actualCourseStudents)-1].StudentID)
	}

	if stepState.StudentIDs[len(actualCourseStudents)-1] != actualCourseStudents[0].StudentID {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected studentId: want: %s, actual : %s", stepState.StudentIDs[len(actualCourseStudents)-1], actualCourseStudents[0].StudentID)
	}

	if len(stepState.StudentIDs) != len(actualCourseStudents) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected size: want: %d, actual : %d", len(stepState.StudentIDs), len(actualCourseStudents))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallCourseStudentList(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	variables := map[string]interface{}{
		"course_id": graphql.String(stepState.CourseID),
	}

	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.CourseStudentListQuery, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnCourseStudentsByCourseIDCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	// after order by
	utils.Reverse(stepState.StudentIDs)
	for i := 0; i < len(stepState.CourseStudentListQuery.CourseStudentsList); i++ {
		if stepState.StudentIDs[i] != stepState.CourseStudentListQuery.CourseStudentsList[i].StudentID {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected students id: want: %s, actual student: %s", stepState.StudentIDs[i], stepState.CourseStudentListQuery.CourseStudentsList[i].StudentID)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCallCourseStudentListV2(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	stepState.Limit = rand.Intn(5) + 3
	stepState.Offset = rand.Intn(3)
	variables := map[string]interface{}{
		"course_id": graphql.String(stepState.CourseID),
		"limit":     graphql.Int(stepState.Limit),
		"offset":    graphql.Int(stepState.Offset),
	}

	err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.CourseStudentListV2Query, variables)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnCourseStudentsByCourseIDWithLimitAndOffsetCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	actualCount := stepState.CourseStudentListV2Query.CourseStudentsAggregate.Aggregate.Count
	if len(stepState.StudentIDs) != actualCount {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected count: want: %d, actual: %d", len(stepState.StudentIDs), actualCount)
	}
	// after order by
	utils.Reverse(stepState.StudentIDs)
	// after offset
	stepState.StudentIDs = stepState.StudentIDs[stepState.Offset:]
	// after limit
	stepState.StudentIDs = stepState.StudentIDs[:stepState.Limit]

	actualCourseStudents := stepState.CourseStudentListV2Query.CourseStudentsListV2
	if len(stepState.StudentIDs) != len(stepState.CourseStudentListV2Query.CourseStudentsListV2) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected size: want: %d, actual: %d", len(stepState.StudentIDs), len(actualCourseStudents))
	}
	for i, courseStudent := range actualCourseStudents {
		if courseStudent.StudentID != stepState.StudentIDs[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected studentID: want: %s, actual: %s", stepState.StudentIDs[i], courseStudent.StudentID)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
