package eureka

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/eureka/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
)

func (s *suite) aListOfLosCreated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aValidBookContent(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	req := s.generateLOsReq(ctx)
	for _, lo := range req.LearningObjectives {
		stepState.LoIDs = append(stepState.LoIDs, lo.Info.Id)
	}

	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.Conn).UpsertLOs(s.signedCtx(ctx), req); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) losHaveBeenDeletedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	e := &entities.LearningObjective{}
	stmt := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE lo_id = ANY($1) AND deleted_at IS NOT NULL", e.TableName())
	var count pgtype.Int8
	if err := s.DB.QueryRow(ctx, stmt, stepState.LoIDs).Scan(&count); err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count.Int != int64(len(stepState.LoIDs)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("amount of los deleted is wrong, expect %v but got %v", len(stepState.LoIDs), count.Int)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteLos(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewLearningObjectiveModifierServiceClient(s.Conn).DeleteLos(s.signedCtx(ctx), &epb.DeleteLosRequest{
		LoIds: stepState.LoIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteLosAgain(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewLearningObjectiveModifierServiceClient(s.Conn).DeleteLos(s.signedCtx(ctx), &epb.DeleteLosRequest{
		LoIds: stepState.LoIDs,
	})
	return StepStateToContext(ctx, stepState), nil
}
