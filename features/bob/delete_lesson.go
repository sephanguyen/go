package bob

import (
	"context"
	"fmt"

	repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/jackc/pgx/v4"
)

func (s *suite) userDeleteALesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := bpb.DeleteLessonRequest{
		LessonId: stepState.lessonID,
	}
	_, stepState.ResponseErr = bpb.NewLessonManagementServiceClient(s.Conn).DeleteLesson(s.signedCtx(ctx), &req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userNoLongerSeesTheLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonRepo{}
	_, err := repo.FindByID(ctx, s.DB, database.Text(stepState.lessonID))
	if err == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s be deleted, but not", stepState.lessonID)
	} else if err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userNoLongerSeesTheLessonReport(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonReportRepo{}
	_, err := repo.FindByLessonID(ctx, s.DB, database.Text(stepState.lessonID))
	if err == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected all lesson report of lesson %s be deleted, but not", stepState.lessonID)
	} else if err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
