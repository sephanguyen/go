package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"
)

func (s *suite) currentLessonHasStreamingLearner(ctx context.Context, capacity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var numberOfStream int

	switch capacity {
	case "max":
		numberOfStream = s.Cfg.Agora.MaximumLearnerStreamings
	case StatusNone:
		numberOfStream = 0
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported lesson streaming capacity")
	}

	query := `UPDATE lessons SET stream_learner_counter = $1 WHERE lesson_id = $2`
	_, err := s.LessonmgmtDB.Exec(ctx, query, numberOfStream, stepState.CurrentLessonID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to update stream learner counter for lesson %s: %w", stepState.CurrentLessonID, err)
	}

	stepState.NumberOfStream = numberOfStream
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userPreparesToPublish(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &vpb.PreparePublishRequest{
		LessonId:  stepState.CurrentLessonID,
		LearnerId: stepState.CurrentStudentID,
	}

	stepState.Response, stepState.ResponseErr = vpb.NewVirtualClassroomModifierServiceClient(s.VirtualClassroomConn).
		PreparePublish(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetsPublishStatus(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	response := stepState.Response.(*vpb.PreparePublishResponse)
	var expectedStatus vpb.PrepareToPublishStatus
	switch status {
	case StatusNone:
		expectedStatus = vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_NONE
	case "prepared before":
		expectedStatus = vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_PREPARED_BEFORE
	case "max limit":
		expectedStatus = vpb.PrepareToPublishStatus_PREPARE_TO_PUBLISH_STATUS_REACHED_MAX_UPSTREAM_LIMIT
	default:
		return StepStateToContext(ctx, stepState), fmt.Errorf("unsupported expected status")
	}

	if response.Status != expectedStatus {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected status %s does not match with actual status %s", expectedStatus.String(), response.Status.String())
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) currentLessonIncludeStreamingLearner(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	lessonID := stepState.CurrentLessonID
	learnerIDs, err := (&repo.VirtualLessonRepo{}).GetStreamingLearners(ctx, s.LessonmgmtDB, lessonID, false)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to get streaming learners for lesson %s: %w", lessonID, err)
	}

	learnerID := stepState.CurrentStudentID
	switch status {
	case "includes":
		if !sliceutils.Contains(learnerIDs, learnerID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s is not part of the streaming learners in lesson %s", learnerID, lessonID)
		}
	case "does not include":
		if sliceutils.Contains(learnerIDs, learnerID) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("learner %s is part of the streaming learners in lesson %s", learnerID, lessonID)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
