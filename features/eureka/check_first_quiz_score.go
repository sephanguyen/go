package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) doQuizExam(ctx context.Context, limit int, isDropped bool) (context.Context, error) {
	ctx, err := s.userCreateQuizTestWithValidRequestAndLimitTheFirstTime(ctx, strconv.Itoa(limit))
	if err != nil {
		return ctx, err
	}
	ctx, err = s.returnListOfQuizItems(ctx, strconv.Itoa(limit))
	if err != nil {
		return ctx, err
	}

	stepState := StepStateFromContext(ctx)
	if stepState.QuizOptions == nil {
		stepState.QuizOptions = make(map[string]map[string][]*cpb.QuizOption)
	}
	curPage := 0

	quizOptionsContent := make(map[string][]*cpb.QuizOption)
	for len(stepState.QuizItems) != 0 {
		stepState.CheckQuizCorrectnessResponses = make([]*epb.CheckQuizCorrectnessResponse, 0)
		for i := 0; i < len(stepState.QuizItems); i++ {
			quizID := stepState.QuizItems[i].Core.ExternalId
			quizOptionsContent[quizID] = stepState.QuizItems[i].Core.Options

			var err error
			switch stepState.QuizItems[i].Core.Kind {
			case cpb.QuizType_QUIZ_TYPE_MCQ:
				ctx, err = s.doMultiChoiceQuiz(ctx, strconv.Itoa(i+1))
			case cpb.QuizType_QUIZ_TYPE_FIB:
				ctx, err = s.doFillInTheBlankQuiz(ctx, strconv.Itoa(i+1))
			case cpb.QuizType_QUIZ_TYPE_POW:
				ctx, err = s.doPairOfWordQuiz(ctx, strconv.Itoa(i+1))
			case cpb.QuizType_QUIZ_TYPE_TAD:
				ctx, err = s.doTermAndDefinitionQuiz(ctx, strconv.Itoa(i+1))
			case cpb.QuizType_QUIZ_TYPE_MIQ:
				ctx, err = s.doManualInputQuiz(ctx, strconv.Itoa(i+1))
			case cpb.QuizType_QUIZ_TYPE_MAQ:
				ctx, err = s.doMultiChoiceQuiz(ctx, strconv.Itoa(i+1))
			case cpb.QuizType_QUIZ_TYPE_ORD:
				ctx, err = s.doOrderingQuestion(ctx, stepState.QuizItems[i].Core.ExternalId)
			case cpb.QuizType_QUIZ_TYPE_ESQ:
				ctx = s.doEssayQuiz(ctx)
			}

			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}

		ctx = s.userGetNextPageOfQuizTest(ctx)
		ctx, err := s.returnQuizItems(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		curPage++
		if isDropped {
			break
		}
	}
	stepState.QuizOptions[stepState.SetID] = quizOptionsContent
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) doOrderingQuestion(ctx context.Context, questionID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.initMap()

	for _, quiz := range stepState.Quizzes {
		if quiz.ExternalID.String == questionID {
			isCorrect := true
			if rand.Intn(2) == 1 { //nolint:gosec
				isCorrect = false
			}
			if _, err := s.studentAnswerOrderingQuestion(ctx, quiz, isCorrect); err != nil {
				return ctx, fmt.Errorf("studentAnswerOrderingQuestion: %w", err)
			}
			if _, err := s.returnsStatusCode(ctx, "OK"); err != nil {
				return ctx, err
			}
			if _, err := s.checkCheckQuizCorrectnessResponseForOrderingQuestion(ctx, questionID); err != nil {
				return ctx, fmt.Errorf("checkCheckQuizCorrectnessResponseForOrderingQuestion: %w", err)
			}
			return StepStateToContext(ctx, stepState), nil
		}
	}

	return StepStateToContext(ctx, stepState), fmt.Errorf("could not found question ID %s in stepState.Quizzes", questionID)
}

func (s *suite) doMultiChoiceQuiz(ctx context.Context, quizIdx string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	idx, _ := strconv.Atoi(quizIdx)
	n := rand.Intn(len(stepState.QuizItems[idx-1].Core.Options)) + 1
	selectIndex := make([]string, n)
	for j := 0; j < n; j++ {
		selectIndex[j] = strconv.Itoa(j + 1)
	}

	ctx, err1 := s.studentChooseOptionOfTheQuiz(ctx, strings.Join(selectIndex, ", "), quizIdx)
	ctx, err2 := s.returnsStatusCode(ctx, "OK")
	ctx, err3 := s.checkResultMultipleChoiceType(ctx, idx-1)
	err := multierr.Combine(err1, err2, err3)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) doEssayQuiz(ctx context.Context) context.Context {
	stepState := StepStateFromContext(ctx)

	stepState.CheckQuizCorrectnessResponses = append(stepState.CheckQuizCorrectnessResponses, &epb.CheckQuizCorrectnessResponse{})

	return StepStateToContext(ctx, stepState)
}

func (s *suite) createStudentEventLogsAfterDoQuiz(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	now := time.Now()

	logs := []*epb.StudentEventLog{
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: &timestamppb.Timestamp{Seconds: now.Add(-5 * 360 * time.Second).Unix()},
			Payload: &epb.StudentEventLogPayload{
				LoId:            stepState.LoID,
				Event:           "started",
				SessionId:       stepState.SessionID,
				StudyPlanItemId: stepState.StudyPlanID,
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: &timestamppb.Timestamp{Seconds: now.Add(-5 * 300 * time.Second).Unix()},
			Payload: &epb.StudentEventLogPayload{
				LoId:            stepState.LoID,
				Event:           "exited",
				SessionId:       stepState.SessionID,
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: &timestamppb.Timestamp{Seconds: now.Add(-5 * 150 * time.Second).Unix()},
			Payload: &epb.StudentEventLogPayload{
				LoId:            stepState.LoID,
				Event:           "started",
				SessionId:       stepState.SessionID,
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
		{
			EventId:   strconv.Itoa(rand.Int()),
			EventType: "learning_objective",
			CreatedAt: &timestamppb.Timestamp{Seconds: now.Add(-5 * 100 * time.Second).Unix()},
			Payload: &epb.StudentEventLogPayload{
				LoId:            stepState.LoID,
				Event:           "completed",
				SessionId:       stepState.SessionID,
				StudyPlanItemId: stepState.StudyPlanItemID,
			},
		},
	}
	_, err := epb.NewStudentEventLogModifierServiceClient(s.Conn).CreateStudentEventLogs(s.signedCtx(ctx), &epb.CreateStudentEventLogsRequest{
		StudentEventLogs: logs,
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) doManualInputQuiz(ctx context.Context, quizIdx string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	idx, _ := strconv.Atoi(quizIdx)
	selectIndex := make([]string, 1)
	selectIndex[0] = strconv.Itoa(rand.Intn(2) + 1)

	ctx, err1 := s.studentChooseOptionOfTheQuiz(ctx, strings.Join(selectIndex, ", "), quizIdx)
	ctx, err2 := s.returnsStatusCode(ctx, "OK")
	ctx, err3 := s.checkResultManualInputType(ctx, idx-1)
	err := multierr.Combine(err1, err2, err3)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) doFillInTheBlankQuiz(ctx context.Context, quizIdxStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, texts := s.getSampleFilledText(ctx)

	quizIdx, _ := strconv.Atoi(quizIdxStr)
	filledTextsStr := make([]string, len(stepState.QuizItems[quizIdx-1].Core.Options))

	for i := range filledTextsStr {
		filledTextsStr[i] = texts[rand.Intn(len(texts))]
	}

	ctx, err1 := s.studentFillTextOfTheQuiz(ctx, strings.Join(filledTextsStr, ", "), quizIdxStr)
	ctx, err2 := s.returnsStatusCode(ctx, "OK")
	ctx, err3 := s.checkResultFillInTheBlank(ctx, quizIdx-1)
	err := multierr.Combine(err1, err2, err3)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) doPairOfWordQuiz(ctx context.Context, quizIdxStr string) (context.Context, error) {
	return s.doFillInTheBlankQuiz(ctx, quizIdxStr)
}
func (s *suite) doTermAndDefinitionQuiz(ctx context.Context, quizIdxStr string) (context.Context, error) {
	return s.doFillInTheBlankQuiz(ctx, quizIdxStr)
}

func (s *suite) getSampleFilledText(ctx context.Context) (context.Context, []string) {
	stepState := StepStateFromContext(ctx)
	texts := make([]string, 0, len(stepState.QuizItems))
	textMap := make(map[string]bool)
	for _, item := range stepState.QuizItems {
		for _, opt := range item.Core.Options {
			optEnt := &entities.QuizOption{
				Content: entities.RichText{
					Raw:         opt.Content.Raw,
					RenderedURL: opt.Content.Rendered,
				},
				Correctness: opt.Correctness,
				Label:       opt.Label,
			}
			texts = append(texts, optEnt.GetText())
			textMap[optEnt.GetText()] = true
		}
	}
	for t := range textMap {
		texts = append(texts, t)
	}
	texts = append(texts, "")
	return ctx, texts
}
func (s *suite) getSampleFilledEntityMapData(ctx context.Context) (context.Context, []string) {
	stepState := StepStateFromContext(ctx)
	data := make([]string, 0, len(stepState.QuizItems))
	dataMap := make(map[string]bool)
	for _, item := range stepState.QuizItems {
		for _, opt := range item.Core.Options {
			optEnt := &entities.QuizOption{
				Content: entities.RichText{
					Raw:         opt.Content.Raw,
					RenderedURL: opt.Content.Rendered,
				},
				Correctness: opt.Correctness,
				Label:       opt.Label,
			}
			data = append(data, optEnt.GetEntityMapData())
			dataMap[optEnt.GetEntityMapData()] = true
		}
	}
	for t := range dataMap {
		data = append(data, t)
	}
	data = append(data, "")
	return ctx, data
}
