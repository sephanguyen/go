package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/cucumber/godog"
)

func (s *suite) updateFlashcardStudyWithArguments(ctx context.Context, typ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := &epb.UpdateFlashCardStudyProgressRequest{}
	numSkippedQuestions := rand.Int31n(8)
	numRememberedQuestions := rand.Int31n(8)
	studyingIndex := numSkippedQuestions + numRememberedQuestions
	skippedQuestionIds := make([]string, 0, numSkippedQuestions)
	rememberedQuestionIds := make([]string, 0, numRememberedQuestions)
	for _, quiz := range stepState.Quizzes[:numSkippedQuestions] {
		skippedQuestionIds = append(skippedQuestionIds, quiz.ExternalID.String)
	}
	for _, quiz := range stepState.Quizzes[numSkippedQuestions : numRememberedQuestions+numSkippedQuestions] {
		rememberedQuestionIds = append(rememberedQuestionIds, quiz.ExternalID.String)
	}

	switch typ {
	case "empty":
		// no-op
	case "valid":
		req = &epb.UpdateFlashCardStudyProgressRequest{
			StudySetId:            stepState.StudySetID,
			StudentId:             stepState.StudentID,
			SkippedQuestionIds:    skippedQuestionIds,
			RememberedQuestionIds: rememberedQuestionIds,
			StudyingIndex:         studyingIndex,
		}
	case withoutStudySetID:
		req = &epb.UpdateFlashCardStudyProgressRequest{
			StudentId:             stepState.StudentID,
			SkippedQuestionIds:    skippedQuestionIds,
			RememberedQuestionIds: rememberedQuestionIds,
			StudyingIndex:         studyingIndex,
		}
	case withoutStudentID:
		req = &epb.UpdateFlashCardStudyProgressRequest{
			StudySetId:            stepState.StudySetID,
			SkippedQuestionIds:    skippedQuestionIds,
			RememberedQuestionIds: rememberedQuestionIds,
			StudyingIndex:         studyingIndex,
		}
	case "without studying_index":
		req = &epb.UpdateFlashCardStudyProgressRequest{
			StudySetId:            stepState.StudySetID,
			StudentId:             stepState.StudentID,
			SkippedQuestionIds:    skippedQuestionIds,
			RememberedQuestionIds: rememberedQuestionIds,
		}
	default:
		return StepStateToContext(ctx, stepState), godog.ErrPending
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.Conn).
		UpdateFlashCardStudyProgress(s.signedCtx(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) flashcardStudyProgressMustBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req := stepState.Request.(*epb.UpdateFlashCardStudyProgressRequest)
	resp := stepState.Response.(*epb.UpdateFlashCardStudyProgressResponse)

	if !resp.IsSuccess {
		return StepStateToContext(ctx, stepState), fmt.Errorf("flashcard wasn't updated")
	}

	flashcardProgression := new(entities.FlashcardProgression)
	fieldName, values := flashcardProgression.FieldMap()
	query := fmt.Sprintf("SELECT %s FROM %s WHERE study_set_id = $1", strings.Join(fieldName, ","), flashcardProgression.TableName())
	err := s.DB.QueryRow(ctx, query, &req.StudySetId).Scan(values...)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if req.StudyingIndex != flashcardProgression.StudyingIndex.Int {
		return StepStateToContext(ctx, stepState), fmt.Errorf("flashcard wasn't updated. Expected studying_index %v, got %v", req.StudyingIndex, flashcardProgression.StudyingIndex.Int)
	}
	if len(req.SkippedQuestionIds) != len(flashcardProgression.SkippedQuestionIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("flashcard wasn't updated. Expected length of skipped_question_ids %v, got %v", len(req.SkippedQuestionIds), len(flashcardProgression.SkippedQuestionIDs.Elements))
	}
	if len(req.RememberedQuestionIds) != len(flashcardProgression.RememberedQuestionIDs.Elements) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("flashcard wasn't updated. Expected length of remembered_question_ids %v, got %v", len(req.RememberedQuestionIds), len(flashcardProgression.RememberedQuestionIDs.Elements))
	}

	return StepStateToContext(ctx, stepState), nil
}
