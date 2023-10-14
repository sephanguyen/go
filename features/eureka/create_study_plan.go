package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) removeAllItemsFromBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	if _, err := pb.NewAssignmentModifierServiceClient(s.Conn).DeleteAssignments(ctx, &pb.DeleteAssignmentsRequest{
		AssignmentIds: stepState.AssignmentIDs,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete assignments: %w", err)
	}

	if _, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).DeleteLos(ctx, &pb.DeleteLosRequest{
		LoIds: stepState.LoIDs,
	}); err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to delete los: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertAssignmentsIntoBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	stepState.AssignmentID = idutil.ULIDNow()
	stepState.AssignmentIDs = []string{stepState.AssignmentID}
	ctx, assignment := s.generateAssignment(ctx, stepState.AssignmentID, false, false, true)

	_, err := pb.NewAssignmentModifierServiceClient(s.Conn).UpsertAssignments(ctx, &pb.UpsertAssignmentsRequest{
		Assignments: []*pb.Assignment{
			assignment,
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert assignment: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) upsertLearningObjectiveIntoBook(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.AuthToken = stepState.SchoolAdminToken
	ctx = contextWithToken(s, ctx)

	learningObjective := s.generateLearningObjective1(ctx)
	stepState.LoID = learningObjective.Info.Id
	stepState.LoIDs = []string{stepState.LoID}

	_, err := pb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(ctx, &pb.UpsertLOsRequest{
		LearningObjectives: []*cpb.LearningObjective{
			learningObjective,
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to upsert learning objective: %w", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlanItemsBelongsToAssignmentsWereSuccesfullyCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if err := try.Do(func(attempt int) (bool, error) {
		query := `SELECT count(*) FROM study_plan_items Where content_structure ->> 'assignment_id' = ANY($1) and deleted_at IS NULL`
		var count int
		if err := s.DB.QueryRow(ctx, query, stepState.AssignmentIDs).Scan(&count); err != nil {
			return true, err
		}
		if count != 0 {
			return false, nil
		}

		time.Sleep(2 * time.Second)
		return attempt < 10, fmt.Errorf("study plan items not inserted")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studyPlanItemsBelongsToLOsWereSuccesfullyCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if err := try.Do(func(attempt int) (bool, error) {
		query := `SELECT count(*) FROM study_plan_items Where content_structure ->> 'lo_id' = ANY($1) and deleted_at IS NULL`
		var count int
		if err := s.DB.QueryRow(ctx, query, stepState.LoIDs).Scan(&count); err != nil {
			return true, err
		}
		if count != 0 {
			return false, nil
		}

		time.Sleep(2 * time.Second)
		return attempt < 10, fmt.Errorf("study plan items not inserted")
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
