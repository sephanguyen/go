package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/cucumber/godog"
)

const (
	withoutStudentID  = "without student_id"
	withoutStudySetID = "without study_set_id"
)

func (s *suite) retrieveFlashcardStudyProgressWithArguments(ctx context.Context, typ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &pb.RetrieveFlashCardStudyProgressRequest{}

	switch typ {
	case "empty":
		// no-op
	case "valid":
		req = &pb.RetrieveFlashCardStudyProgressRequest{
			StudentId:  stepState.CurrentStudentID,
			StudySetId: stepState.StudySetID,
			Paging: &cpb.Paging{
				Limit: uint32(20),
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: int64(1),
				},
			},
		}
	case withoutStudySetID:
		req = &pb.RetrieveFlashCardStudyProgressRequest{
			StudentId: stepState.StudentID,
			Paging: &cpb.Paging{
				Limit: uint32(20),
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: int64(1),
				},
			},
		}
	case withoutStudentID:
		req = &pb.RetrieveFlashCardStudyProgressRequest{
			StudySetId: stepState.StudySetID,
			Paging: &cpb.Paging{
				Limit: uint32(20),
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: int64(1),
				},
			},
		}
	case "without paging":
		req = &pb.RetrieveFlashCardStudyProgressRequest{
			StudySetId: stepState.StudySetID,
			StudentId:  stepState.StudentID,
		}
	case "random":
		stepState.Limit = rand.Intn(21)
		req = &pb.RetrieveFlashCardStudyProgressRequest{
			StudySetId: idutil.ULIDNow(),
			StudentId:  idutil.ULIDNow(),
			Paging: &cpb.Paging{
				Limit: uint32(stepState.Limit),
				Offset: &cpb.Paging_OffsetInteger{
					OffsetInteger: int64(1),
				},
			},
		}
	default:
		return StepStateToContext(ctx, stepState), godog.ErrPending
	}

	stepState.Request = req
	if err := try.Do(func(attempt int) (retry bool, err error) {
		time.Sleep(2 * time.Second)
		stepState.Response, stepState.ResponseErr = pb.NewFlashCardReaderServiceClient(s.Conn).
			RetrieveFlashCardStudyProgress(s.signedCtx(ctx), req)
		resp := stepState.Response.(*pb.RetrieveFlashCardStudyProgressResponse)
		if resp != nil && len(resp.Items) != int(req.Paging.Limit) {
			return attempt < 10, fmt.Errorf("expected: number of quizzes equal %v, got %v", req.Paging.Limit, len(resp.Items))
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedFlashcardStudyProgress(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*pb.RetrieveFlashCardStudyProgressRequest)

	resp := stepState.Response.(*pb.RetrieveFlashCardStudyProgressResponse)
	if resp.StudySetId != req.StudySetId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: study_set_id equal %v, got %v", req.StudySetId, resp.StudySetId)
	}
	if len(resp.Items) != int(req.Paging.Limit) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: number of quizzes equal %v, got %v", req.Paging.Limit, len(resp.Items))
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) flashcardStudyProgressResponseMatch(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &bpb.RetrieveFlashCardStudyProgressRequest{
		StudentId:  stepState.CurrentStudentID,
		StudySetId: stepState.StudySetID,
		Paging: &cpb.Paging{
			Limit: uint32(20),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: int64(1),
			},
		},
	}
	stepState.Response, stepState.ResponseErr = bpb.NewCourseReaderServiceClient(s.Conn).
		RetrieveFlashCardStudyProgress(s.signedCtx(ctx), req)
	resp := stepState.Response.(*bpb.RetrieveFlashCardStudyProgressResponse)
	if resp != nil && resp.StudySetId != req.StudySetId {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: study_set_id equal %v, got %v", req.StudySetId, resp.StudySetId)
	}
	if resp != nil && len(resp.Items) != int(req.Paging.Limit) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected: number of quizzes equal %v, got %v", req.Paging.Limit, len(resp.Items))
	}

	return StepStateToContext(ctx, stepState), nil
}
