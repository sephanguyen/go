package assignment

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/features/repository/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/hasura/go-graphql-client"
)

type assignments_order_by struct {
	name         graphql.String
	assignmentId graphql.String `graphql:"assignment_id"`
}

type _text graphql.String

func (d *_text) GetGraphQLType() string {
	return ""
}

func (s *Suite) userGetAssignmentsByCallAssignmentOne(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	variables := map[string]interface{}{
		"assignment_id": graphql.String(stepState.AssignmentIDs[0]),
	}
	if err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.AssignmentOneQuery, variables); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query hasura assignment one, err: " + err.Error())
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentOneCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignments := stepState.AssignmentOneQuery.Assignments

	if len(assignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found assignment with this ID %s", stepState.AssignmentIDs[0])
	}

	if assignments[0].AssignmentID != stepState.AssignmentIDs[0] ||
		assignments[0].Name != stepState.Assignments[0].Name.String {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unexpected assignment id: want %s, actual: %s", stepState.AssignmentIDs[0], assignments[0].AssignmentID)
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAssignmentsByCallAssignmentsByTopicIds(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	orderBy := assignments_order_by{
		name:         "asc",
		assignmentId: "asc",
	}

	topicIdVariables := _text("{" + stepState.TopicIDs[0] + "}")
	variables := map[string]interface{}{
		"topic_id": topicIdVariables,
		"order_by": orderBy,
	}

	if err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.AssignmentsByTopicIdQuery, variables); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query hasura assignments by topic ids, err: " + err.Error())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentsByTopicIdsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	hasuraAssignments := stepState.AssignmentsByTopicIdQuery.Assignments
	if len(hasuraAssignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return result incorrect, not found any assignment")
	}

	// get assignments from database and check with hasura
	var assignmentRepo repositories.AssignmentRepo
	assignments, err := assignmentRepo.RetrieveAssignmentsByTopicIDs(ctx, s.DB, database.TextArray(stepState.TopicIDs))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query assignment, err: " + err.Error())
	}
	if len(assignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found any assignment with this topic id")
	}
	for i := range assignments {
		topicIdDatabase := GetTopicIdFromAssignment(*assignments[i])
		if stepState.TopicIDs[0] != topicIdDatabase {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("wrong assignment topic id, expected: %s, get: %s", stepState.TopicIDs[0], topicIdDatabase)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func GetTopicIdFromAssignment(assignment entities.Assignment) string {
	assignmentContent := assignment.Content.Get()
	contentStrArr := strings.Split(fmt.Sprintf("%v", assignmentContent), ":")           // convert interface to string and split it
	topicIdDatabase := strings.Replace(contentStrArr[len(contentStrArr)-1], "]", "", 1) // replace the last element in arr that has "]" with ""
	return topicIdDatabase
}

func (s *Suite) userGetAssignmentsByCallAssignmentsMany(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	assignmentIdVariables := make([]graphql.String, 0, len(stepState.AssignmentIDs))
	for i := range stepState.AssignmentIDs {
		assignmentIdVariables = append(assignmentIdVariables, graphql.String(stepState.AssignmentIDs[i]))
	}

	variables := map[string]interface{}{
		"assignment_id": assignmentIdVariables,
	}
	if err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.AssignmentsManyQuery, variables); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query hasura assignments many, err: " + err.Error())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentsManyCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	hasuraAssignments := stepState.AssignmentsManyQuery.Assignments

	if len(hasuraAssignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return result incorrect, not found any assignment")
	}
	for i := range hasuraAssignments {
		if i > 0 {
			if hasuraAssignments[i].AssignmentID < hasuraAssignments[i-1].AssignmentID || hasuraAssignments[i].Name < hasuraAssignments[i-1].Name {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return assignment wrong order")
			}
		}
	}

	for i := range hasuraAssignments {
		index := -1
		// find the hasura element in stepState array
		for j := range stepState.AssignmentIDs {
			if hasuraAssignments[i].AssignmentID == stepState.AssignmentIDs[j] {
				index = j
				break
			}
		}

		if index == -1 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return assignment incorrect, not found")
		} else if hasuraAssignments[i].Name != stepState.Assignments[index].Name.String ||
			hasuraAssignments[i].DisplayOrder != stepState.Assignments[index].DisplayOrder.Int ||
			hasuraAssignments[i].MaxGrade != stepState.Assignments[index].MaxGrade.Int {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return assignment incorrect field")
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userGetAssignmentsByCallAssignmentDisplayOrder(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	topicIdVariables := _text("{" + stepState.TopicIDs[0] + "}")
	variables := map[string]interface{}{
		"topic_id": topicIdVariables,
	}

	if err := utils.QueryHasura(ctx, s.HasuraAdminURL, s.HasuraPassword, &stepState.AssignmentDisplayOrderQuery, variables); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query hasura assignments display order, err: " + err.Error())
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnAssignmentDisplayOrderCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var assignmentRepo repositories.AssignmentRepo

	hasuraAssignments := stepState.AssignmentDisplayOrderQuery.Assignments
	if len(hasuraAssignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return result incorrect, not found any assignment")
	}

	assignments, err := assignmentRepo.RetrieveAssignmentsByTopicIDs(ctx, s.DB, database.TextArray(stepState.TopicIDs))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to query assignment, err: " + err.Error())
	}
	if len(assignments) == 0 {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not found any assignments with this topic id")
	}

	isDisplayOrderExist := false
	for i := range assignments {
		if int32(assignments[i].DisplayOrder.Int) == hasuraAssignments[0].DisplayOrder {
			isDisplayOrderExist = true
			break
		}
	}
	if !isDisplayOrderExist {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("hasura return wrong display order")
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
