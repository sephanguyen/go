package lessonmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure/repo"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

func (s *Suite) lessonsSchedulingStatusAreUpdatedCorrectlyTo(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	lessonRepo := repo.LessonRepo{}
	expectedLessonCount := len(s.CommonSuite.StepState.LessonIDs)
	var expectedSchedulingStatus domain.LessonSchedulingStatus
	var actualLessonCount int
	lessons, err := lessonRepo.GetLessonByIDs(ctx, s.CommonSuite.BobDB, s.CommonSuite.StepState.LessonIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query lesson: %s", err)
	}
	actualLessonCount = len(lessons)
	if actualLessonCount != expectedLessonCount {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("query missing lesson: expected %d lesson, got %d", expectedLessonCount, actualLessonCount)
	}
	switch status {
	case "canceled":
		expectedSchedulingStatus = domain.LessonSchedulingStatusCanceled
	case "published":
		expectedSchedulingStatus = domain.LessonSchedulingStatusPublished
	case "completed":
		expectedSchedulingStatus = domain.LessonSchedulingStatusCompleted
	case "draft":
		expectedSchedulingStatus = domain.LessonSchedulingStatusDraft
	}

	for _, lesson := range lessons {
		if lesson.SchedulingStatus != expectedSchedulingStatus {
			return StepStateToContext(ctx, stepState), fmt.Errorf("incorrect bulk update lesson status of ID %s: expected status %s, got %s", lesson.LessonID, expectedSchedulingStatus, lesson.SchedulingStatus)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userBulkUpdatesStatusWithAction(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	updateReq := &lpb.BulkUpdateLessonSchedulingStatusRequest{
		LessonIds: s.CommonSuite.LessonIDs,
	}
	switch action {
	case "cancel":
		updateReq.Action = lpb.LessonBulkAction_LESSON_BULK_ACTION_CANCEL
	case "publish":
		updateReq.Action = lpb.LessonBulkAction_LESSON_BULK_ACTION_PUBLISH
	}
	stepState.Response, stepState.ResponseErr = lpb.NewLessonModifierServiceClient(s.LessonMgmtConn).BulkUpdateLessonSchedulingStatus(s.CommonSuite.SignedCtx(ctx), updateReq)
	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}
