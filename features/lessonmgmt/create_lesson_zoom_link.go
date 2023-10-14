package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/zoom/domain"

	"github.com/google/uuid"
)

func (s *Suite) GenerateZoomAccount() *domain.ZoomAccount {
	id := uuid.New().String()
	return &domain.ZoomAccount{
		ID:       id,
		Email:    fmt.Sprintf("email-%s@gmail.com", id),
		UserName: id,
	}
}

func (s *Suite) UpsertValidZoomAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	zoomAccount := s.GenerateZoomAccount()

	_, err := s.BobDB.Exec(ctx, `INSERT INTO public."zoom_account"
					(zoom_id, email, user_name, created_at, updated_at)
					VALUES ($1, $2, $3, now(), now())
					ON CONFLICT ON CONSTRAINT pk__zoom_account DO UPDATE SET email = $2, user_name = $3`, database.Text(zoomAccount.ID),
		database.Text(zoomAccount.Email), database.Text(zoomAccount.UserName))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.ZoomAccount = zoomAccount

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) UserCreateALessonZoomWithAllRequiredFields(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ZoomLink = "https://us04web.zoom.us/s/79852036766"
	return s.CommonSuite.UserCreateALessonZoomWithMissingFieldsInLessonmgmt(ctx)
}
