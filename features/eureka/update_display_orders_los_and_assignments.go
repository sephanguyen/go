package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) userCreateLosAndAssignments(ctx context.Context) (context.Context, error) {
	return s.hasCreatedAContentBook(ctx, "school admin")
}

func (s *suite) userUpdateDisplayOrdersForLOsAndAssignments(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx = contextWithToken(s, ctx)

	if len(stepState.LoIDs) > 0 && len(stepState.AssignmentIDs) > 0 {
		stepState.LoID = stepState.LoIDs[0]
		stepState.AssignmentID = stepState.AssignmentIDs[0]
	} else {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not have lo or assignment")
	}
	req := &pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest{}

	req.LearningObjectives = append(req.LearningObjectives, &pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest_LearningObjective{
		LoId:         stepState.LoID,
		DisplayOrder: rand.Int31n(200),
		TopicId:      stepState.TopicID,
	})

	req.Assignments = append(req.Assignments, &pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest_Assignment{
		AssignmentId: stepState.AssignmentID,
		DisplayOrder: rand.Int31n(200),
		TopicId:      stepState.TopicID,
	})

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewCourseModifierServiceClient(s.Conn).UpdateDisplayOrdersOfLOsAndAssignments(ctx, req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) displayOrderOfLOsAndAssignmentsMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.UpdateDisplayOrdersOfLOsAndAssignmentsRequest)

	mapLOAndDisplayOrder := make(map[string]int32)
	for _, lo := range req.LearningObjectives {
		mapLOAndDisplayOrder[lo.LoId] = lo.DisplayOrder
	}

	mapAssignmentAndDisplayOrder := make(map[string]int32)
	for _, assignment := range req.Assignments {
		mapAssignmentAndDisplayOrder[assignment.AssignmentId] = assignment.DisplayOrder
	}
	mainProcess := func() error {
		queryTopicsLearningObjectives := `SELECT lo_id, display_order
			FROM topics_learning_objectives
		    WHERE topic_id = $1 AND lo_id= $2 AND deleted_at IS NULL`

		row := s.DB.QueryRow(ctx, queryTopicsLearningObjectives, stepState.TopicID, stepState.LoID)
		var (
			loID         pgtype.Text
			assignmentID pgtype.Text
			displayOrder int32
		)
		err := row.Scan(&loID, &displayOrder)
		if err != nil {
			return fmt.Errorf("unable scan from topics_learning_objectives: %w", err)
		}

		if loID.Status == pgtype.Present && displayOrder != mapLOAndDisplayOrder[loID.String] {
			return fmt.Errorf("lo_id %v expected: display order = %v but got: %v", loID, displayOrder, mapLOAndDisplayOrder[loID.String])
		}

		queryLearningObjectives := `SELECT lo_id, display_order
			FROM learning_objectives
		    WHERE lo_id = $1 AND deleted_at IS NULL`

		row = s.DB.QueryRow(ctx, queryLearningObjectives, stepState.LoID)
		err = row.Scan(&loID, &displayOrder)
		if err != nil {
			return fmt.Errorf("unable get from learning_objectives: %w", err)
		}

		if displayOrder != mapLOAndDisplayOrder[loID.String] {
			return fmt.Errorf("lo_id %v expected: display order = %v but got: %v", loID, displayOrder, mapLOAndDisplayOrder[loID.String])
		}

		queryTopicsAssignments := `SELECT assignment_id, display_order
			FROM topics_assignments
		    WHERE topic_id = $1 AND assignment_id = $2 AND deleted_at IS NULL`

		row = s.DB.QueryRow(ctx, queryTopicsAssignments, stepState.TopicID, stepState.AssignmentID)

		err = row.Scan(&assignmentID, &displayOrder)
		if err != nil {
			return fmt.Errorf("unable get from topics_assignments: %w", err)
		}

		if displayOrder != mapAssignmentAndDisplayOrder[assignmentID.String] {
			return fmt.Errorf("assignment_id %v expected: display order = %v but got: %v", assignmentID, displayOrder, mapLOAndDisplayOrder[loID.String])
		}

		queryAssignments := `SELECT assignment_id, display_order
			FROM assignments
		    WHERE assignment_id = $1 AND deleted_at IS NULL`
		row = s.DB.QueryRow(ctx, queryAssignments, database.Text(stepState.AssignmentID))
		err = row.Scan(&assignmentID, &displayOrder)
		if err != nil {
			return fmt.Errorf("unable scan from assignments: %w", err)
		}

		if displayOrder != mapAssignmentAndDisplayOrder[assignmentID.String] {
			return fmt.Errorf("assignment_id %v expected: display order = %v but got: %v", assignmentID.String, displayOrder, mapLOAndDisplayOrder[loID.String])
		}
		return nil
	}
	return s.ExecuteWithRetry(ctx, mainProcess, 2*time.Second, 5)
}
