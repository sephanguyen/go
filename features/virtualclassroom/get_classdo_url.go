package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) anExistingLessonWithAClassDoAccount(ctx context.Context) (context.Context, error) {
	ctx, err := s.anExistingVirtualClassroom(ctx)
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	lessonID := stepState.CurrentLessonID
	classDoLink := "https://app.class.com/join?token=ABCDEFGHIJKLMNOP"

	cmdTag, err := s.LessonmgmtDB.Exec(ctx, `UPDATE lessons SET classdo_link = $2, classdo_owner_id = $3
			WHERE lesson_id = $1 and deleted_at IS NULL `,
		database.Text(lessonID),
		database.Text(classDoLink),
		database.Text(stepState.ClassDoAccount.ClassDoID),
	)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("error in update lesson %s: %w", lessonID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s was not updated", lessonID)
	}
	stepState.ClassDoLink = classDoLink

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsALessonWithAClassDoAccount(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.GetClassDoURLRequest{
		LessonId: stepState.CurrentLessonID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualLessonReaderServiceClient(s.VirtualClassroomConn).
		GetClassDoURL(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsTheExpectedClassDoLink(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	response := stepState.Response.(*vpb.GetClassDoURLResponse)

	lessonID := stepState.CurrentLessonID
	expectedLink := stepState.ClassDoLink
	actualLink := response.GetClassdoLink()
	if len(actualLink) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("lesson %s ClassDo link is empty", lessonID)
	}
	if actualLink != expectedLink {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the actual link %s is not the same as the expected link %s of lesson %s", actualLink, expectedLink, lessonID)
	}

	return StepStateToContext(ctx, stepState), nil
}
