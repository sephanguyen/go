package eureka

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/cucumber/godog"
)

func (s *suite) retrieveLastFlashcardStudyProgressWithArguments(ctx context.Context, typ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.RetrieveLastFlashcardStudyProgressRequest{}

	switch typ {
	case "empty":
		// no-op
	case "valid":
		req = &pb.RetrieveLastFlashcardStudyProgressRequest{
			StudyPlanItemId: stepState.StudyPlanItemID,
			LoId:            stepState.LoID,
			StudentId:       stepState.CurrentStudentID,
			IsCompleted:     true,
		}
	case "without study_plan_item_id":
		req = &pb.RetrieveLastFlashcardStudyProgressRequest{
			LoId:        stepState.LoID,
			StudentId:   stepState.CurrentStudentID,
			IsCompleted: true,
		}
	case withoutStudentID:
		req = &pb.RetrieveLastFlashcardStudyProgressRequest{
			StudyPlanItemId: stepState.StudyPlanItemID,
			LoId:            stepState.LoID,
			IsCompleted:     true,
		}
	case "without lo_id":
		req = &pb.RetrieveLastFlashcardStudyProgressRequest{
			StudyPlanItemId: stepState.StudyPlanItemID,
			StudentId:       stepState.CurrentStudentID,
			IsCompleted:     true,
		}
	case "without is_completed":
		req = &pb.RetrieveLastFlashcardStudyProgressRequest{
			StudyPlanItemId: stepState.StudyPlanItemID,
			LoId:            stepState.LoID,
			StudentId:       stepState.CurrentStudentID,
		}
	default:
		return StepStateToContext(ctx, stepState), godog.ErrPending
	}

	stepState.Request = req
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(1 * time.Second)
		stepState.Response, stepState.ResponseErr = pb.NewFlashCardReaderServiceClient(s.Conn).RetrieveLastFlashcardStudyProgress(s.signedCtx(ctx), req)
		resp := stepState.Response.(*pb.RetrieveLastFlashcardStudyProgressResponse)
		if resp != nil && typ == "valid" && stepState.StudySetID != resp.StudySetId && !req.IsCompleted {
			return attempt < 10, fmt.Errorf("sync data bob to eureka failed")
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.BobResponse = stepState.Response
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsStudySetID(ctx context.Context, typ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*pb.RetrieveLastFlashcardStudyProgressResponse)
	switch typ {
	case "empty":
		if resp.StudySetId != "" {
			return StepStateToContext(ctx, stepState), fmt.Errorf("study_set_id must be empty, but got study_set_id = %v", resp.StudySetId)
		}
	case "valid":
		if resp.StudySetId != stepState.StudySetID {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect study_set_id %v, but got %v", stepState.StudySetID, resp.StudySetId)
		}
	default:
		return StepStateToContext(ctx, stepState), godog.ErrPending
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) lastFlashcardStudyProgressResponseMatch(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := stepState.Request.(*pb.RetrieveLastFlashcardStudyProgressRequest)
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(1 * time.Second)
		stepState.Response, stepState.ResponseErr = pb.NewFlashCardReaderServiceClient(s.Conn).RetrieveLastFlashcardStudyProgress(
			s.signedCtx(ctx),
			&pb.RetrieveLastFlashcardStudyProgressRequest{
				StudyPlanItemId: req.StudyPlanItemId,
				LoId:            req.LoId,
				StudentId:       req.StudentId,
				IsCompleted:     req.IsCompleted,
			})
		resp := stepState.Response.(*pb.RetrieveLastFlashcardStudyProgressResponse)
		if resp.StudySetId == "" {
			return attempt < 10, fmt.Errorf("sync data bob to eureka failed")
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	bobResp := stepState.BobResponse.(*pb.RetrieveLastFlashcardStudyProgressResponse)
	eurekaResp := stepState.Response.(*pb.RetrieveLastFlashcardStudyProgressResponse)
	if bobResp.StudySetId != eurekaResp.StudySetId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not match study set id")
	}

	return StepStateToContext(ctx, stepState), nil
}
