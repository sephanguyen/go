package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) checkResultManualInputType(ctx context.Context, quizIdx int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.CheckQuizCorrectnessResponses[quizIdx]

	quizRepo := repositories.QuizRepo{}
	quizID := stepState.QuizItems[quizIdx].Core.ExternalId
	loID := stepState.QuizItems[quizIdx].LoId
	options, err := quizRepo.GetOptions(ctx, s.DB, database.Text(quizID), database.Text(loID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	expected := make([]bool, 0, len(options))
	for _, answer := range stepState.SelectedIndex[stepState.SetID][quizID] {
		idx := answer.GetSelectedIndex()
		if idx > uint32(len(options)) {
			continue
		}
		expected = append(expected, options[idx-1].Correctness)
	}

	if len(expected) != len(resp.Correctness) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect there are %v answer from response but got %v", len(expected), len(resp.Correctness))
	}
	for i, res := range resp.Correctness {
		if res != expected[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("manual input quiz type in the answer %v expected %v but got %v", i, expected[i], res)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResultFillInTheBlank(ctx context.Context, quizIdx int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.CheckQuizCorrectnessResponses[quizIdx]

	return s.checkResultFillInTheBlankWithArgs(ctx, quizIdx, resp.Correctness)
}

func (s *suite) checkResultFillInTheBlankWithArgs(ctx context.Context, quizIdx int, correctness []bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	quizRepo := repositories.QuizRepo{}

	quizID := stepState.QuizItems[quizIdx].Core.ExternalId
	loID := stepState.QuizItems[quizIdx].LoId
	options, err := quizRepo.GetOptions(ctx, s.DB, database.Text(quizID), database.Text(loID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	quiz := &entities.Quiz{}
	_ = quiz.Options.Set(options)
	optionsWithAlternatives, _ := quiz.GetOptionsWithAlternatives()
	expected := make([]bool, 0, len(optionsWithAlternatives))
	for i, answer := range stepState.FilledText[stepState.SetID][quizID] {
		text := answer.GetFilledText()
		if i >= len(optionsWithAlternatives) {
			continue
		}
		expected = append(expected, optionsWithAlternatives[i].IsCorrect(text))
	}

	if len(expected) != len(correctness) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect there are %v answer from response but got %v", len(expected), len(correctness))
	}
	for i, res := range correctness {
		if res != expected[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("in the answer %v expected %v but got %v", i, expected[i], res)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) checkResultMultipleChoiceType(ctx context.Context, quizIdx int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.CheckQuizCorrectnessResponses[quizIdx]

	return s.checkResultMultipleChoiceTypeWithArgs(ctx, quizIdx, resp.Correctness)
}

func (s *suite) checkResultMultipleChoiceTypeWithArgs(ctx context.Context, quizIdx int, correctness []bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	quizRepo := repositories.QuizRepo{}
	quizSetRepo := repositories.ShuffledQuizSetRepo{}

	quizID := stepState.QuizItems[quizIdx].Core.ExternalId
	loID := stepState.QuizItems[quizIdx].LoId
	options, err := quizRepo.GetOptions(ctx, s.DB, database.Text(quizID), database.Text(loID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	seedStr, err := quizSetRepo.GetSeed(ctx, s.DB, database.Text(stepState.SetID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	idx, err := quizSetRepo.GetQuizIdx(ctx, s.DB, database.Text(stepState.SetID), database.Text(quizID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	seed, _ := strconv.ParseInt(seedStr.String, 10, 64)
	r := rand.New(rand.NewSource(seed + int64(idx.Int)))
	r.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	expected := make([]bool, 0, len(options))
	for _, answer := range stepState.SelectedIndex[stepState.SetID][quizID] {
		idx := answer.GetSelectedIndex()
		if idx > uint32(len(options)) {
			continue
		}
		expected = append(expected, options[idx-1].Correctness)
	}
	if len(expected) != len(correctness) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect there are %v answer from response but got %v", len(expected), len(correctness))
	}
	for i, res := range correctness {
		if res != expected[i] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("in the answer %v expected %v but got %v", i, expected[i], res)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aQuizTestOfLearningObjectiveBelongToTopicIncludeQuizzesWithQuizzesEveryPage(ctx context.Context, arg1, arg2, arg3 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := arg1
	numOfQuizzes := arg2
	limit := arg3
	ctx, err := s.aQuizTest(ctx, topic, numOfQuizzes, limit, "mix")
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aQuizTestFillInTheBlankQuizzesWithQuizzesPerPageAndDoQuizTest(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTest(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)])
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aQuizTestIncludeMultipleChoiceQuizzesWithQuizzesPerPageAndDoQuizTest(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTest(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aQuizTestIncludePairOfWordQuizzesWithQuizzesPerPageAndDoQuizTest(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTest(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_POW)])
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) aQuizTestIncludeTermAndDefinitionQuizzesWithQuizzesPerPageAndDoQuizTest(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTest(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_TAD)])
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aQuizTestQuizzesWithQuizzesPerPageAndDoQuizTest(ctx context.Context, arg1, arg2, arg3 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg3

	var quizType cpb.QuizType
	switch arg2 {
	case "ordering":
		quizType = cpb.QuizType_QUIZ_TYPE_ORD
	case "essay":
		quizType = cpb.QuizType_QUIZ_TYPE_ESQ
	}

	ctx, err := s.aQuizTest(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(quizType)])
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) aQuizTest(ctx context.Context, topic, numOfQuizzes, limit, quizType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch quizType {
	case "mix":
		ctx, err = s.aQuizsetWithQuizzesInLearningObjectiveBelongedToATopic(ctx, numOfQuizzes, topic)
	default:
		ctx, err = s.quizSetWithAll(ctx, numOfQuizzes, topic, cpb.QuizType(cpb.QuizType_value[quizType]))
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err1 := s.aSignedIn(ctx, "student")
	ctx, err2 := s.userCreateQuizTestWithValidRequestAndLimitTheFirstTime(ctx, limit)
	ctx, err3 := s.returnListOfQuizItems(ctx, limit)
	return StepStateToContext(ctx, stepState), multierr.Combine(err1, err2, err3)
}

func (s *suite) studentChooseOptionOfTheQuiz(ctx context.Context, selectedIdxsStr, quizIdxStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.SelectedIndex == nil {
		stepState.SelectedIndex = make(map[string]map[string][]*epb.Answer)
	}

	if stepState.SelectedIndex[stepState.SetID] == nil {
		stepState.SelectedIndex[stepState.SetID] = make(map[string][]*epb.Answer)
	}
	selectedIdxs := make([]*epb.Answer, 0)

	for _, idx := range strings.Split(selectedIdxsStr, ",") {
		i, _ := strconv.Atoi(strings.TrimSpace(idx))
		selectedIdxs = append(selectedIdxs, &epb.Answer{Format: &epb.Answer_SelectedIndex{SelectedIndex: uint32(i)}})
	}

	quizIndex, _ := strconv.Atoi(quizIdxStr)
	quizIndex--
	stepState.SelectedIndex[stepState.SetID][stepState.QuizItems[quizIndex].Core.ExternalId] = selectedIdxs
	stepState.SelectedQuiz = append(stepState.SelectedQuiz, quizIndex)

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(s.signedCtx(ctx), &epb.CheckQuizCorrectnessRequest{
		SetId:  stepState.SetID,
		QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
		Answer: selectedIdxs,
	})

	resp := stepState.Response.(*epb.CheckQuizCorrectnessResponse)
	stepState.CheckQuizCorrectnessResponses = append(stepState.CheckQuizCorrectnessResponses, resp)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentMissingQuizIdInRequest(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, ok := stepState.Response.(*epb.CreateQuizTestResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("got: %T, expected: *epb.CreateQuizTestResponse", stepState.Response)
	}
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(s.signedCtx(ctx), &epb.CheckQuizCorrectnessRequest{
		SetId:  resp.QuizzesId,
		QuizId: "",
		Answer: []*epb.Answer{
			{
				Format: &epb.Answer_SelectedIndex{SelectedIndex: uint32(1)},
			},
		},
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentFillTextOfTheQuiz(ctx context.Context, filledTextsStr, quizIdxStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.FilledText == nil {
		stepState.FilledText = make(map[string]map[string][]*epb.Answer)
	}
	if stepState.FilledText[stepState.SetID] == nil {
		stepState.FilledText[stepState.SetID] = make(map[string][]*epb.Answer)
	}
	filledTexts := make([]*epb.Answer, 0)

	for _, text := range strings.Split(filledTextsStr, ",") {
		text = strings.TrimSpace(text)
		filledTexts = append(filledTexts, &epb.Answer{Format: &epb.Answer_FilledText{FilledText: text}})
	}

	quizIndex, _ := strconv.Atoi(quizIdxStr)
	quizIndex--

	stepState.FilledText[stepState.SetID][stepState.QuizItems[quizIndex].Core.ExternalId] = filledTexts
	stepState.SelectedQuiz = append(stepState.SelectedQuiz, quizIndex)

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(s.signedCtx(ctx), &epb.CheckQuizCorrectnessRequest{
		SetId:  stepState.SetID,
		QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
		Answer: filledTexts,
	})

	resp := stepState.Response.(*epb.CheckQuizCorrectnessResponse)
	stepState.CheckQuizCorrectnessResponses = append(stepState.CheckQuizCorrectnessResponses, resp)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentFillTextOfTheQuizForSubmitQuizAnswers(ctx context.Context, filledTextsStr, quizIdxStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.FilledText == nil {
		stepState.FilledText = make(map[string]map[string][]*epb.Answer)
	}
	if stepState.FilledText[stepState.SetID] == nil {
		stepState.FilledText[stepState.SetID] = make(map[string][]*epb.Answer)
	}
	filledTexts := make([]*epb.Answer, 0)

	for _, text := range strings.Split(filledTextsStr, ",") {
		text = strings.TrimSpace(text)
		filledTexts = append(filledTexts, &epb.Answer{Format: &epb.Answer_FilledText{FilledText: text}})
	}

	quizIndex, _ := strconv.Atoi(quizIdxStr)
	quizIndex--

	stepState.FilledText[stepState.SetID][stepState.QuizItems[quizIndex].Core.ExternalId] = filledTexts
	stepState.SelectedQuiz = append(stepState.SelectedQuiz, quizIndex)

	stepState.QuizAnswers = append(stepState.QuizAnswers, &epb.QuizAnswer{
		QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
		Answer: filledTexts,
	})

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAnswerCorrectOrderingOptionOfTheQuizForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, quiz := range stepState.Quizzes {
		stepState.QuizAnswers = append(stepState.QuizAnswers, &epb.QuizAnswer{
			QuizId: quiz.ExternalID.String,
			Answer: []*epb.Answer{
				{
					Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-A"},
				},
				{
					Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-B"},
				},
				{
					Format: &epb.Answer_SubmittedKey{SubmittedKey: "key-C"},
				},
			},
		})
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentFinishEssay(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, quiz := range stepState.Quizzes {
		stepState.QuizAnswers = append(stepState.QuizAnswers, &epb.QuizAnswer{
			QuizId: quiz.ExternalID.String,
			Answer: []*epb.Answer{
				{
					Format: &epb.Answer_FilledText{FilledText: "sample-text"},
				},
			},
		})
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAnswerQuiz(ctx context.Context, quizType cpb.QuizType) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch quizType {
	case cpb.QuizType_QUIZ_TYPE_MCQ:
		ctx, err = s.studentAnswerMultipleChoices(ctx)
	case cpb.QuizType_QUIZ_TYPE_FIB:
		ctx, err = s.studentAnswerFillInTheBlank(ctx)
	case cpb.QuizType_QUIZ_TYPE_TAD:
		ctx, err = s.studentAnswerTermAndDefinitionQuizzes(ctx)
	case cpb.QuizType_QUIZ_TYPE_POW:
		ctx, err = s.studentAnswerPairOfWordQuizzes(ctx)
	case cpb.QuizType_QUIZ_TYPE_MIQ:
		ctx, err = s.studentAnswerManualInput(ctx)
	case cpb.QuizType_QUIZ_TYPE_MAQ:
		ctx, err = s.studentAnswerMultipleChoices(ctx)
	case cpb.QuizType_QUIZ_TYPE_ORD:
		ctx, err = s.studentAnswerOrderingQuestions(ctx)
	}
	return StepStateToContext(ctx, stepState), err
}

// studentAnswerOrderingQuestions will submit answers for all ordering quiz items
// in step state (stepState.QuizItems) via CheckQuizCorrectness API
// and always guarantee have at least 1 correct answer
func (s *suite) studentAnswerOrderingQuestions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.initMap()

	// because stepState.QuizItems is response from BE, so its option was shuffled,
	// we will use stepState.Quizzes instead of
	quizzes := make(entities.Quizzes, 0, len(stepState.QuizItems))
	quizzesMap := make(map[string]*entities.Quiz)
	for _, quiz := range stepState.Quizzes {
		quizzesMap[quiz.ExternalID.String] = quiz
	}
	for _, quiz := range stepState.QuizItems {
		if v, ok := quizzesMap[quiz.Core.ExternalId]; ok {
			quizzes = append(quizzes, v)
		} else {
			return ctx, fmt.Errorf("could not found quiz %s in response", quiz.Core.ExternalId)
		}
	}

	for i, quiz := range quizzes {
		isCorrect := true
		if i == len(stepState.QuizItems)-1 {
			isCorrect = false
		} else if i > 0 {
			if rand.Intn(2) == 1 { //nolint:gosec
				isCorrect = false
			}
		}
		if _, err := s.studentAnswerOrderingQuestion(ctx, quiz, isCorrect); err != nil {
			return ctx, err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

//nolint:unparam
func (s *suite) studentAnswerOrderingQuestion(ctx context.Context, quiz *entities.Quiz, isCorrect bool) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.SubmittedKeys[stepState.SetID] == nil {
		stepState.SubmittedKeys[stepState.SetID] = make(map[string][]*epb.Answer)
	}
	option, err := quiz.GetOptions()
	if err != nil {
		return ctx, fmt.Errorf("quiz.GetOptions: %w", err)
	}
	quizID := quiz.ExternalID.String
	answer := make([]*epb.Answer, 0, len(option))
	stepState.expectedCorrectnessByQuizID[quizID] = make([]bool, 0, len(option))
	stepState.expectedCorrectKeysByQuizID[quizID] = make([]string, 0, len(option))
	for _, opt := range option {
		answer = append(answer, &epb.Answer{Format: &epb.Answer_SubmittedKey{SubmittedKey: opt.Key}})
		stepState.expectedCorrectnessByQuizID[quizID] = append(
			stepState.expectedCorrectnessByQuizID[quizID],
			true,
		)
		stepState.expectedCorrectKeysByQuizID[quizID] = append(
			stepState.expectedCorrectKeysByQuizID[quizID],
			opt.Key,
		)
	}
	stepState.expectedIsCorrectAllByQuizID[quizID] = true
	// case incorrect answer
	wrongAns := &epb.Answer{Format: &epb.Answer_SubmittedKey{SubmittedKey: "wrong-key"}}
	if !isCorrect {
		stepState.expectedIsCorrectAllByQuizID[quizID] = false
		answer[len(answer)-1] = wrongAns
		stepState.expectedCorrectnessByQuizID[quizID][len(stepState.expectedCorrectnessByQuizID[quizID])-1] = false
	}
	resp, err := epb.NewQuizModifierServiceClient(s.Conn).CheckQuizCorrectness(s.signedCtx(ctx), &epb.CheckQuizCorrectnessRequest{
		SetId:  stepState.SetID,
		QuizId: quizID,
		Answer: answer,
	})
	stepState.Response, stepState.ResponseErr = resp, err

	// persist input and output to use for checking result
	stepState.SubmittedKeys[stepState.SetID][quizID] = answer
	stepState.CheckQuizCorrectnessResponsesByQuizID[quizID] = resp
	// legacy state
	stepState.CheckQuizCorrectnessResponses = append(stepState.CheckQuizCorrectnessResponses, resp)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAnswerMultipleChoices(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := range stepState.QuizItems {
		idx := strconv.Itoa(i + 1)
		selectedIndex := make([]string, 0, len(stepState.QuizItems[i].Core.Options))
		for j := 0; j < rand.Intn(len(stepState.QuizItems[i].Core.Options))+1; j++ {
			selectedIndex = append(selectedIndex, strconv.Itoa(j+1))
		}
		s.studentChooseOptionOfTheQuiz(ctx, strings.Join(selectedIndex, ", "), idx)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAnswerManualInput(ctx context.Context) (context.Context, error) {
	return s.studentAnswerMultipleChoices(ctx)
}

func (s *suite) studentAnswerFillInTheBlank(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := range stepState.QuizItems {
		ctx, sampleTexts := s.getSampleFilledText(ctx)
		quiz := &entities.Quiz{}
		_ = quiz.Options.Set(stepState.QuizItems[i].Core.Options)
		options, _ := quiz.GetOptionsWithAlternatives()
		answerText := make([]string, len(options))
		for j := range answerText {
			answerText[j] = sampleTexts[rand.Intn(len(sampleTexts))]
		}
		ctx, err := s.studentFillTextOfTheQuiz(ctx, strings.Join(answerText, ", "), strconv.Itoa(i+1))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAnswerFillInTheBlankForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := range stepState.QuizItems {
		ctx, sampleTexts := s.getSampleFilledText(ctx)
		quiz := &entities.Quiz{}
		_ = quiz.Options.Set(stepState.QuizItems[i].Core.Options)
		options, _ := quiz.GetOptionsWithAlternatives()
		answerText := make([]string, len(options))
		for j := range answerText {
			answerText[j] = sampleTexts[rand.Intn(len(sampleTexts))]
		}
		ctx, err := s.studentFillTextOfTheQuizForSubmitQuizAnswers(ctx, strings.Join(answerText, ", "), strconv.Itoa(i+1))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentAnswerPairOfWordQuizzes(ctx context.Context) (context.Context, error) {
	return s.studentAnswerFillInTheBlank(ctx)
}

func (s *suite) studentAnswerPairOfWordQuizzesForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	return s.studentAnswerFillInTheBlankForSubmitQuizAnswers(ctx)
}

func (s *suite) studentAnswerTermAndDefinitionQuizzes(ctx context.Context) (context.Context, error) {
	return s.studentAnswerFillInTheBlank(ctx)
}

func (s *suite) studentAnswerTermAndDefinitionQuizzesForSubmiteQuizAnswers(ctx context.Context) (context.Context, error) {
	return s.studentAnswerFillInTheBlankForSubmitQuizAnswers(ctx)
}

func (s *suite) returnsExpectedResultQuiz(ctx context.Context, quizType cpb.QuizType) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch quizType {
	case cpb.QuizType_QUIZ_TYPE_MCQ:
		ctx, err = s.returnsExpectedResultMultipleChoiceType(ctx)
	case cpb.QuizType_QUIZ_TYPE_FIB:
		ctx, err = s.returnsExpectedResultFillInTheBlankType(ctx)
	case cpb.QuizType_QUIZ_TYPE_TAD:
		ctx, err = s.returnsExpectedResultTermAndDefinitionQuizzes(ctx)
	case cpb.QuizType_QUIZ_TYPE_POW:
		ctx, err = s.returnsExpectedResultPairOfWordQuizzes(ctx)
	case cpb.QuizType_QUIZ_TYPE_MIQ:
		ctx, err = s.returnsExpectedResultManualInputType(ctx)
	case cpb.QuizType_QUIZ_TYPE_MAQ:
		ctx, err = s.returnsExpectedResultMultipleChoiceType(ctx)
	case cpb.QuizType_QUIZ_TYPE_ORD:
		ctx, err = s.checkCheckQuizCorrectnessResponseForOrderingQuestions(ctx)
	}
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) returnsExpectedResultMultipleChoiceType(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, selectedQuiz := range stepState.SelectedQuiz {
		ctx, err := s.checkResultMultipleChoiceType(ctx, selectedQuiz)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsExpectedResultManualInputType(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, selectedQuiz := range stepState.SelectedQuiz {
		ctx, err := s.checkResultManualInputType(ctx, selectedQuiz)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	for _, selectedQuiz := range stepState.SelectedQuiz {
		var quiz entities.Quiz
		fields := database.GetFieldNames(&quiz)
		query := fmt.Sprintf(`SELECT %v FROM quizzes WHERE external_id = $1`, strings.Join(fields, ", "))
		err := s.DB.QueryRow(ctx, query, stepState.QuizItems[selectedQuiz].Core.ExternalId).Scan(database.GetScanFields(&quiz, fields)...)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		// manual input has 2 options,
		// first option always false
		// second option always true
		option, err := quiz.GetOptions()
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if len(option) != 2 {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect 2 options for manual input quiz type but got %v", len(option))
		}
		if option[0].Correctness {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect first option of manual quiz type is false")
		}

		if !option[1].Correctness {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect second option of manual quiz type is true")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedResultFillInTheBlankType(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, selectedQuiz := range stepState.SelectedQuiz {
		ctx, err := s.checkResultFillInTheBlank(ctx, selectedQuiz)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedResultFillInTheBlankTypeForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.SubmitQuizAnswersResponse)
	for _, log := range resp.Logs {
		idx := -1
		for i, quiz := range stepState.QuizItems {
			if quiz.Core.ExternalId == log.QuizId {
				idx = i
				break
			}
		}

		ctx, err := s.checkResultFillInTheBlankWithArgs(ctx, idx, log.Correctness)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedResultPairOfWordQuizzes(ctx context.Context) (context.Context, error) {
	return s.returnsExpectedResultFillInTheBlankType(ctx)
}

func (s *suite) returnsExpectedResultPairOfWordQuizzesForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	return s.returnsExpectedResultFillInTheBlankTypeForSubmitQuizAnswers(ctx)
}

func (s *suite) returnsExpectedResultTermAndDefinitionQuizzes(ctx context.Context) (context.Context, error) {
	return s.returnsExpectedResultFillInTheBlankType(ctx)
}

func (s *suite) returnsExpectedResultTermAndDefinitionQuizzesForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	return s.returnsExpectedResultFillInTheBlankTypeForSubmitQuizAnswers(ctx)
}

func (s *suite) checkCheckQuizCorrectnessResponseForOrderingQuestions(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, quiz := range stepState.QuizItems {
		if _, err := s.checkCheckQuizCorrectnessResponseForOrderingQuestion(ctx, quiz.Core.ExternalId); err != nil {
			return ctx, fmt.Errorf("checkCheckQuizCorrectnessResponseForOrderingQuestion: %w", err)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

//nolint:unparam
func (s *suite) checkCheckQuizCorrectnessResponseForOrderingQuestion(ctx context.Context, questionID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	actual := stepState.CheckQuizCorrectnessResponsesByQuizID[questionID]
	if !slices.Equal(actual.Correctness, stepState.expectedCorrectnessByQuizID[questionID]) {
		return ctx, fmt.Errorf("expected correctness list of quiz %s in response %v but got %v", questionID, stepState.expectedCorrectnessByQuizID[questionID], actual.Correctness)
	}
	if actual.IsCorrectAll != stepState.expectedIsCorrectAllByQuizID[questionID] {
		return ctx, fmt.Errorf("expected IsCorrectAll of quiz %s %v but got %v", questionID, stepState.expectedIsCorrectAllByQuizID[questionID], actual.IsCorrectAll)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsResultAllCorrectInSubmitQuizAnswersForOrderingQuestion(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var totalPoint int32
	for _, quiz := range stepState.Quizzes {
		totalPoint += quiz.Point.Int
	}

	expected := &epb.SubmitQuizAnswersResponse{
		Logs: []*cpb.AnswerLog{
			{
				QuizId:        "",
				QuizType:      cpb.QuizType_QUIZ_TYPE_ORD,
				SelectedIndex: nil,
				CorrectIndex:  nil,
				FilledText:    nil,
				CorrectText:   nil,
				Result: &cpb.AnswerLog_OrderingResult{OrderingResult: &cpb.OrderingResult{
					SubmittedKeys: []string{"key-A", "key-B", "key-C"},
					CorrectKeys:   []string{"key-A", "key-B", "key-C"},
				}},
				Correctness: []bool{true, true, true},
				IsAccepted:  true,
				Core:        nil,
				SubmittedAt: nil,
			},
		},
		TotalGradedPoint:   wrapperspb.UInt32(uint32(totalPoint)),
		TotalPoint:         wrapperspb.UInt32(uint32(totalPoint)),
		TotalCorrectAnswer: int32(len(stepState.Quizzes)),
		TotalQuestion:      int32(len(stepState.Quizzes)),
		SubmissionResult:   epb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED,
	}
	resp := stepState.Response.(*epb.SubmitQuizAnswersResponse)
	if resp.TotalGradedPoint.Value != expected.TotalGradedPoint.Value {
		return ctx, fmt.Errorf("expected total graded point %d but got %d", expected.TotalGradedPoint.Value, resp.TotalGradedPoint.Value)
	}
	if resp.TotalPoint.Value != expected.TotalPoint.Value {
		return ctx, fmt.Errorf("expected total point %d but got %d", expected.TotalPoint.Value, resp.TotalPoint.Value)
	}
	if resp.TotalCorrectAnswer != expected.TotalCorrectAnswer {
		return ctx, fmt.Errorf("expected total correct answer %d but got %d", expected.TotalCorrectAnswer, resp.TotalCorrectAnswer)
	}
	if resp.TotalQuestion != expected.TotalQuestion {
		return ctx, fmt.Errorf("expected total questions %d but got %d", expected.TotalQuestion, resp.TotalQuestion)
	}
	if resp.SubmissionResult != expected.SubmissionResult {
		return ctx, fmt.Errorf("expected submission result %s but got %ds", expected.SubmissionResult.String(), resp.SubmissionResult.String())
	}
	if len(resp.Logs) != len(stepState.Quizzes) {
		return ctx, fmt.Errorf("expected submission logs %d item but got %d item", len(stepState.Quizzes), len(resp.Logs))
	}

	for i, log := range resp.Logs {
		if log.QuizId != stepState.Quizzes[i].ExternalID.String {
			return ctx, fmt.Errorf("expected id of quesion %s but got %s", stepState.Quizzes[i].ExternalID.String, log.QuizId)
		}
		if log.QuizType != expected.Logs[0].QuizType {
			return ctx, fmt.Errorf("expected type of quesion %s %v but got %v", log.QuizId, expected.Logs[0].QuizType, log.QuizType)
		}
		if !slices.Equal(log.Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.SubmittedKeys, expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.SubmittedKeys) {
			return ctx, fmt.Errorf("expected submitted keys of quesion %s %v but got %v", log.QuizId, expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.SubmittedKeys, log.Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.SubmittedKeys)
		}
		if !slices.Equal(log.Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.CorrectKeys, expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.CorrectKeys) {
			return ctx, fmt.Errorf("expected correct keys of quesion %s %v but got %v", log.QuizId, expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.CorrectKeys, log.Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.CorrectKeys)
		}
		if !slices.Equal(log.Correctness, expected.Logs[0].Correctness) {
			return ctx, fmt.Errorf("expected correctness of quesion %s %v but got %v", log.QuizId, expected.Logs[0].Correctness, log.Correctness)
		}
		if log.IsAccepted != expected.Logs[0].IsAccepted {
			return ctx, fmt.Errorf("expected is_accepted of quesion %s %v but got %v", log.QuizId, expected.Logs[0].IsAccepted, log.IsAccepted)
		}
	}

	e := &entities.ExamLOSubmissionAnswer{}
	es := entities.ExamLOSubmissionAnswers{}
	listStmt := fmt.Sprintf(`
SELECT %s
FROM %s
WHERE 
	deleted_at IS NULL
	AND shuffled_quiz_set_id = $1`,
		strings.Join(database.GetFieldNames(e), ","), e.TableName())

	if err := database.Select(ctx, db, listStmt, stepState.ShuffledQuizSetID).ScanAll(&es); err != nil {
		return ctx, fmt.Errorf("failed to get exam lo submission: %w", err)
	}
	for _, item := range es {
		if !slices.Equal(database.FromTextArray(item.SubmittedKeysAnswer), expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.SubmittedKeys) {
			return ctx, fmt.Errorf("expected submitted keys of quesion %s %v but got %v", item.QuizID.String, expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.SubmittedKeys, database.FromTextArray(item.SubmittedKeysAnswer))
		}
		if !slices.Equal(database.FromTextArray(item.CorrectKeysAnswer), expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.CorrectKeys) {
			return ctx, fmt.Errorf("expected correct keys of quesion %s %v but got %v", item.QuizID.String, expected.Logs[0].Result.(*cpb.AnswerLog_OrderingResult).OrderingResult.CorrectKeys, database.FromTextArray(item.CorrectKeysAnswer))
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsEssaySubmitQuizAnswer(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	expected := &epb.SubmitQuizAnswersResponse{
		Logs: []*cpb.AnswerLog{
			{
				QuizId:        "",
				QuizType:      cpb.QuizType_QUIZ_TYPE_ESQ,
				SelectedIndex: nil,
				CorrectIndex:  nil,
				FilledText:    []string{"sample-text"},
			},
		},
		TotalQuestion:    int32(len(stepState.Quizzes)),
		SubmissionResult: epb.ExamLOSubmissionResult_EXAM_LO_SUBMISSION_COMPLETED,
	}
	resp := stepState.Response.(*epb.SubmitQuizAnswersResponse)

	if resp.TotalQuestion != expected.TotalQuestion {
		return ctx, fmt.Errorf("expected total questions %d but got %d", expected.TotalQuestion, resp.TotalQuestion)
	}

	for i, log := range resp.Logs {
		if log.QuizId != stepState.Quizzes[i].ExternalID.String {
			return ctx, fmt.Errorf("expected id of quesion %s but got %s", stepState.Quizzes[i].ExternalID.String, log.QuizId)
		}
		if log.QuizType != expected.Logs[0].QuizType {
			return ctx, fmt.Errorf("expected type of quesion %s %v but got %v", log.QuizId, expected.Logs[0].QuizType, log.QuizType)
		}
		if !slices.Equal(log.FilledText, expected.Logs[0].FilledText) {
			return ctx, fmt.Errorf("expected filled text of quesion %s %v but got %v", log.QuizId, expected.Logs[0].FilledText, log.FilledText)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) aQuizTestIncludeWithQuizzesWithQuizzesPerPageAndDoQuizTest(ctx context.Context, totalQuizzesStr, quizType, limitStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	topic := "TYPE_TOPIC_EXAM"
	var err error
	switch quizType {
	case "multiple choice":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	case "fill in the blank new":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)])
	case "fill in the blank old":
		// use new or old version of fill in the blank
		// the difference:
		// 		the old version does not have field Key
		// 		the new version have field Key
		stepState.UserFillInTheBlankOld = true
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)])

	case "term and definition":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_TAD)])
	case "pair of word":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_POW)])
	case "manual input":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MIQ)])
	case "multiple answer":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MAQ)])
	case "ordering":
		ctx, err = s.aQuizTest(ctx, topic, totalQuizzesStr, limitStr, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_ORD)])
	}

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) studentAnswerQuizzes(ctx context.Context, quizType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch quizType {
	case "multiple choice":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_MCQ)
	case "fill in the blank":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_FIB)
	case "term and definition":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_TAD)
	case "pair of word":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_POW)
	case "manual input":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_MIQ)
	case "multiple answer":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_MAQ)
	case "ordering":
		ctx, err = s.studentAnswerQuiz(ctx, cpb.QuizType_QUIZ_TYPE_ORD)
	}
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) returnsExpectedResultQuizzes(ctx context.Context, quizType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch quizType {
	case "multiple choice":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_MCQ)
	case "fill in the blank":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_FIB)
	case "term and definition":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_TAD)
	case "pair of word":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_POW)
	case "manual input":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_MIQ)
	case "multiple answer":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_MAQ)
	case "ordering":
		ctx, err = s.returnsExpectedResultQuiz(ctx, cpb.QuizType_QUIZ_TYPE_ORD)
	}
	return StepStateToContext(ctx, stepState), err
}

func (s *suite) studentAnswerFillInTheBlankQuizWithOcr(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for i := range stepState.QuizItems {
		ctx, sampleTexts := s.getSampleFilledText(ctx)
		ctx, sampleData := s.getSampleFilledEntityMapData(ctx)
		quiz := &entities.Quiz{}
		_ = quiz.Options.Set(stepState.QuizItems[i].Core.Options)
		options, _ := quiz.GetOptionsWithAlternatives()
		answerTexts := make([]string, len(options))
		for j := range answerTexts {
			answerTexts[j] = sampleTexts[rand.Intn(len(sampleTexts))]
			if answerTexts[j] == "" {
				answerTexts[j] = sampleData[rand.Intn(len(sampleData))]
			}
		}
		ctx, err := s.studentFillTextOfTheQuiz(ctx, strings.Join(answerTexts, ", "), strconv.Itoa(i+1))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsExpectedResultFillInTheBlankQuiz(ctx context.Context) (context.Context, error) {
	return s.returnsExpectedResultFillInTheBlankType(ctx)
}
