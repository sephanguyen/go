package bob

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	repo "github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *suite) userSignedAsSchoolAdmin(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	return s.aSignedInWithSchool(ctx, "school admin", int(stepState.CurrentSchoolID))
}
func (s *suite) userCreateLiveLessonWithStartTimeEndTimeAtFuture(ctx context.Context, name, brightcoveVideoURL string) (context.Context, error) {
	start := time.Now().Add(time.Hour).Format(time.RFC3339)
	end := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	return s.UserCreateLiveLesson(ctx, name, start, end, brightcoveVideoURL)
}
func (s *suite) userDeletesTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr

	}
	resp := stepState.Response.(*bpb.CreateLiveLessonResponse)
	stepState.CurrentLessonID = resp.Id

	req := &bpb.DeleteLiveLessonRequest{
		Id: resp.Id,
	}

	stepState.Response, stepState.ResponseErr = bpb.NewLessonModifierServiceClient(s.Conn).DeleteLiveLesson(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userNoLongerSeesTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonRepo{}
	_, err := repo.FindByID(ctx, s.DB, database.Text(stepState.CurrentLessonID))
	if err == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s be deleted, but not", stepState.CurrentLessonID)
	} else if err.Error() != fmt.Errorf("db.QueryRow: %w", pgx.ErrNoRows).Error() {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userCreateLiveLessonWithEndTimeAtFuture(ctx context.Context, name, start, brightcoveVideoURL string) (context.Context, error) {
	end := time.Now().Add(2 * time.Hour).Format(time.RFC3339)
	return s.UserCreateLiveLesson(ctx, name, start, end, brightcoveVideoURL)
}
func (s *suite) userCreateLiveLessonWithEndTimeAtPast(ctx context.Context, name, brightcoveVideoURL string) (context.Context, error) {
	start := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)
	end := time.Now().Add(-time.Hour).Format(time.RFC3339)
	return s.UserCreateLiveLesson(ctx, name, start, end, brightcoveVideoURL)
}
func (s *suite) userCanNotDeleteTheLiveLesson(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	repo := repo.LessonRepo{}
	res, err := repo.FindByID(ctx, s.DB, database.Text(stepState.CurrentLessonID))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s not be deleted, but it did: %w", stepState.CurrentLessonID, err)
	}

	if res.DeletedAt.Status == pgtype.Present {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected lesson %s not be deleted, but it did", stepState.CurrentLessonID)
	}

	return StepStateToContext(ctx, stepState), nil
}
