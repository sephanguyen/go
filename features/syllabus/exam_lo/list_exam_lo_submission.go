package exam_lo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/bob/constants"
	consta "github.com/manabie-com/backend/internal/eureka/constants"
	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/yasuo/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
	sspb "github.com/manabie-com/backend/pkg/manabuf/syllabus/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Suite) studentChooseOptionOfTheQuizForSubmitQuizAnswers(ctx context.Context, selectedIdxsStr, quizIdxStr string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
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

	QuizIndex, _ := strconv.Atoi(quizIdxStr)
	QuizIndex--
	stepState.SelectedIndex[stepState.SetID][stepState.QuizItems[QuizIndex].Core.ExternalId] = selectedIdxs
	stepState.SelectedQuiz = append(stepState.SelectedQuiz, QuizIndex)

	stepState.QuizAnswers = append(stepState.QuizAnswers, &epb.QuizAnswer{
		QuizId: stepState.QuizItems[QuizIndex].Core.ExternalId,
		Answer: selectedIdxs,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	quizAnswers := make([]*epb.QuizAnswer, 0, len(stepState.QuizAnswers))
	for _, quizAnswer := range stepState.QuizAnswers {
		var answers []*epb.Answer
		for _, answer := range quizAnswer.Answer {
			if _, ok := answer.GetFormat().(*epb.Answer_FilledText); ok {
				answers = append(answers, &epb.Answer{
					Format: &epb.Answer_FilledText{
						FilledText: answer.GetFilledText(),
					},
				})
			}

			if _, ok := answer.Format.(*epb.Answer_SelectedIndex); ok {
				answers = append(answers, &epb.Answer{
					Format: &epb.Answer_SelectedIndex{
						SelectedIndex: answer.GetSelectedIndex(),
					},
				})
			}
		}

		quizAnswers = append(quizAnswers, &epb.QuizAnswer{
			QuizId: quizAnswer.QuizId,
			Answer: answers,
		})
	}

	stepState.Response, stepState.ResponseErr = epb.NewCourseModifierServiceClient(s.EurekaConn).SubmitQuizAnswers(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.SubmitQuizAnswersRequest{
		SetId:      stepState.SetID,
		QuizAnswer: quizAnswers,
	})

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) listExamLOSubmissionsWithValidLocations(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	teacherID := idutil.ULIDNow()
	if err := s.AuthHelper.AValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.AuthHelper.GenerateExchangeToken(teacherID, consta.RoleTeacher)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.Token = token
	resp, err := sspb.NewExamLOClient(s.EurekaConn).ListExamLOSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListExamLOSubmissionRequest{
		CourseId:    wrapperspb.String(stepState.CourseID),
		LocationIds: stepState.LocationIDs,
		Start:       timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
		End:         timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
		SubmittedDate: &sspb.ListExamLOSubmissionRequest_SubmittedDate{
			Start: timestamppb.New(time.Now().Add(-2 * 24 * time.Hour)),
			End:   timestamppb.New(time.Now()),
		},
		LastUpdatedDate: &sspb.ListExamLOSubmissionRequest_LastUpdatedDate{
			Start: timestamppb.New(time.Now().Add(-2 * 24 * time.Hour)),
			End:   timestamppb.New(time.Now()),
		},
		Paging: &cpb.Paging{
			Limit: 100,
		},
	})

	stepState.ResponseErr = err
	stepState.Response = resp
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) listExamLOSubmissionsWithInvalidLocations(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	teacherID := idutil.ULIDNow()
	if err := s.AuthHelper.AValidUser(ctx, teacherID, consta.RoleTeacher); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.AuthHelper.GenerateExchangeToken(teacherID, consta.RoleTeacher)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.Token = token
	resp, err := sspb.NewExamLOClient(s.EurekaConn).ListExamLOSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), &sspb.ListExamLOSubmissionRequest{
		CourseId: wrapperspb.String(stepState.CourseID),
		LocationIds: []string{
			idutil.ULIDNow(),
			idutil.ULIDNow(),
		},
		Start: timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
		End:   timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
		SubmittedDate: &sspb.ListExamLOSubmissionRequest_SubmittedDate{
			Start: timestamppb.New(time.Now().Add(-2 * 24 * time.Hour)),
			End:   timestamppb.New(time.Now()),
		},
		LastUpdatedDate: &sspb.ListExamLOSubmissionRequest_LastUpdatedDate{
			Start: timestamppb.New(time.Now().Add(-2 * 24 * time.Hour)),
			End:   timestamppb.New(time.Now()),
		},
		Paging: &cpb.Paging{
			Limit: 100,
		},
	})

	stepState.ResponseErr = err
	stepState.Response = resp
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsListExamLOSubmissionsCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp := stepState.Response.(*sspb.ListExamLOSubmissionResponse)

	if len(resp.Items) != len(stepState.LOIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected number of items %d, but got %d", len(resp.Items), len(stepState.LOIDs))
	}
	for _, loID := range stepState.LOIDs {
		var containsExamLOSubmission bool
		for _, item := range resp.Items {
			if item.StudyPlanItemIdentity.LearningMaterialId == loID {
				containsExamLOSubmission = true
				break
			}
		}
		if !containsExamLOSubmission {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("response doesn't contain exam lo submission for exam_lo (%s)", loID)
		}
	}

	for _, examLOSubmission := range resp.Items {
		if examLOSubmission.LastAction.String() == "" {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected last_action not empty, but got empty")
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) createQuizTestsAndAnswersForExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	{
		ctx, err1 := s.aQuizTestIncludeMultipleChoiceQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx, "5", "5")
		ctx = s.studentChooseOptionOfTheQuiz(ctx, "1, 2, 3", "1")
		ctx = s.studentChooseOptionOfTheQuiz(ctx, "1, 2", "2")
		ctx = s.studentChooseOptionOfTheQuiz(ctx, "1, 3", "3")
		ctx = s.studentChooseOptionOfTheQuiz(ctx, "2, 3", "4")
		ctx = s.studentChooseOptionOfTheQuiz(ctx, "1, 4", "5")
		ctx, err2 := s.studentSubmitQuizAnswers(ctx)

		if err := multierr.Combine(err1, err2); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	{
		ctx, err1 := s.aQuizTestFillInTheBlankQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx, "5", "5")
		ctx = s.studentFillTextOfTheQuiz(ctx, "hello, goodbye, meeting, fine, bye", "1")
		ctx = s.studentFillTextOfTheQuiz(ctx, "hello, goodbye, fine, bye", "2")
		ctx = s.studentFillTextOfTheQuiz(ctx, "hello, meeting, fine, bye", "3")
		ctx = s.studentFillTextOfTheQuiz(ctx, "goodbye, meeting, fine, bye", "4")
		ctx = s.studentFillTextOfTheQuiz(ctx, "hello, bye", "5")
		ctx, err7 := s.studentSubmitQuizAnswers(ctx)

		if err := multierr.Combine(err1, err7); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	{
		ctx, err1 := s.aQuizTestIncludePairOfWordQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx, "5", "5")
		ctx, err2 := s.studentAnswerPairOfWordQuizzesForSubmitQuizAnswers(ctx)
		ctx, err3 := s.studentSubmitQuizAnswers(ctx)

		if err := multierr.Combine(err1, err2, err3); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	{
		ctx, err1 := s.aQuizTestIncludeTermAndDefinitionQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx, "5", "5")
		ctx, err2 := s.studentAnswerTermAndDefinitionQuizzesForSubmiteQuizAnswers(ctx)
		ctx, err3 := s.studentSubmitQuizAnswers(ctx)

		if err := multierr.Combine(err1, err2, err3); err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentAnswerPairOfWordQuizzesForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	return s.studentAnswerFillInTheBlankForSubmitQuizAnswers(ctx)
}

func (s *Suite) studentAnswerTermAndDefinitionQuizzesForSubmiteQuizAnswers(ctx context.Context) (context.Context, error) {
	return s.studentAnswerFillInTheBlankForSubmitQuizAnswers(ctx)
}

func (s *Suite) getSampleFilledText(ctx context.Context) (context.Context, []string) {
	stepState := utils.StepStateFromContext[StepState](ctx)
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

func (s *Suite) studentAnswerFillInTheBlankForSubmitQuizAnswers(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for i := range stepState.QuizItems {
		ctx, sampleTexts := s.getSampleFilledText(ctx)
		quiz := &entities.Quiz{}
		_ = quiz.Options.Set(stepState.QuizItems[i].Core.Options)
		options, _ := quiz.GetOptionsWithAlternatives()
		answerText := make([]string, len(options))
		for j := range answerText {
			answerText[j] = sampleTexts[rand.Intn(len(sampleTexts))]
		}
		s.studentFillTextOfTheQuizForSubmitQuizAnswers(ctx, strings.Join(answerText, ", "), strconv.Itoa(i+1))
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) studentFillTextOfTheQuizForSubmitQuizAnswers(ctx context.Context, filledTextsStr, quizIdxStr string) context.Context {
	stepState := utils.StepStateFromContext[StepState](ctx)
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

	return utils.StepStateToContext(ctx, stepState)
}

func (s *Suite) studentChooseOptionOfTheQuiz(ctx context.Context, selectedIdxsStr, quizIdxStr string) context.Context {
	stepState := utils.StepStateFromContext[StepState](ctx)
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

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.CheckQuizCorrectnessRequest{
		SetId:  stepState.SetID,
		QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
		Answer: selectedIdxs,
	})

	resp := stepState.Response.(*epb.CheckQuizCorrectnessResponse)
	stepState.CheckQuizCorrectnessResponses = append(stepState.CheckQuizCorrectnessResponses, resp)

	return utils.StepStateToContext(ctx, stepState)
}

func (s *Suite) studentFillTextOfTheQuiz(ctx context.Context, filledTextsStr, quizIdxStr string) context.Context {
	stepState := utils.StepStateFromContext[StepState](ctx)
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

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.EurekaConn).CheckQuizCorrectness(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.CheckQuizCorrectnessRequest{
		SetId:  stepState.SetID,
		QuizId: stepState.QuizItems[quizIndex].Core.ExternalId,
		Answer: filledTexts,
	})

	resp := stepState.Response.(*epb.CheckQuizCorrectnessResponse)
	stepState.CheckQuizCorrectnessResponses = append(stepState.CheckQuizCorrectnessResponses, resp)

	return utils.StepStateToContext(ctx, stepState)
}

type block struct {
	Key               string      `json:"key"`
	Text              string      `json:"text"`
	Type              string      `json:"type"`
	Depth             string      `json:"depth"`
	InlineStyleRanges []string    `json:"inlineStyleRanges"`
	EntityRanges      []string    `json:"entityRanges"`
	Data              interface{} `json:"data"`
	EntityMap         interface{} `json:"entityMap"`
}

type raw struct {
	Blocks []block `json:"blocks"`
}

func (s *Suite) someStudentsAddedToCourseInSomeValidLocations(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	m := rand.Int31n(3) + 1
	locationIDs := make([]string, 0, m)
	for i := int32(1); i <= m; i++ {
		id := idutil.ULIDNow()
		locationIDs = append(locationIDs, id)
	}

	stepState.CourseID = idutil.ULIDNow()

	// insert multi user to bob db
	studentIDs, err := utils.InsertMultiUserIntoBob(ctx, s.BobDB, 10)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("InsertMultiUserIntoBob: %w", err)
	}
	stepState.StudentIDs = studentIDs
	stepState.CurrentStudentID = studentIDs[0]
	courseStudents, err := utils.AValidCourseWithIDs(ctx, s.EurekaDB, stepState.StudentIDs, stepState.CourseID)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AValidCourseWithIDs: %w", err)
	}
	for _, courseStudent := range courseStudents {
		for _, locationID := range locationIDs {
			now := time.Now()
			e := &entities.CourseStudentsAccessPath{}
			database.AllNullEntity(e)
			if err := multierr.Combine(
				e.CourseStudentID.Set(courseStudent.ID.String),
				e.CourseID.Set(courseStudent.CourseID.String),
				e.StudentID.Set(courseStudent.StudentID.String),
				e.LocationID.Set(locationID),
				e.CreatedAt.Set(now),
				e.UpdatedAt.Set(now),
			); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("multierr.Combine: %w", err)
			}
			if _, err := database.Insert(ctx, e, s.EurekaDB.Exec); err != nil {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("database.Insert: %w", err)
			}
		}
	}
	stepState.CourseStudents = courseStudents
	stepState.LocationIDs = locationIDs
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aQuizTestIncludeMultipleChoiceQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTestForExamLO(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.aQuizTestForExamLO: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aQuizTestFillInTheBlankQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTestForExamLO(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.aQuizTestForExamLO: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aQuizTestIncludePairOfWordQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTestForExamLO(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_POW)])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.aQuizTestForExamLO: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aQuizTestIncludeTermAndDefinitionQuizzesWithQuizzesPerPageAndDoQuizTestForExamLO(ctx context.Context, arg1, arg2 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	topic := "TYPE_TOPIC_EXAM"
	numOfQuizzes := arg1
	limit := arg2
	ctx, err := s.aQuizTestForExamLO(ctx, topic, numOfQuizzes, limit, cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_TAD)])
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.aQuizTestForExamLO: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) aQuizTestForExamLO(ctx context.Context, topic, numOfQuizzes, limit, quizType string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	var err error
	switch quizType {
	// case "mix":
	// ctx, err = s.aQuizsetWithQuizzesInLearningObjectiveBelongedToATopic(ctx, numOfQuizzes, topic)

	case cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)]:
		ctx, err = s.quizSetWithAllForExamLO(ctx, numOfQuizzes, topic, cpb.QuizType_QUIZ_TYPE_MCQ)

	case cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)]:
		ctx, err = s.quizSetWithAllForExamLO(ctx, numOfQuizzes, topic, cpb.QuizType_QUIZ_TYPE_FIB)

	case cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_POW)]:
		ctx, err = s.quizSetWithAllForExamLO(ctx, numOfQuizzes, topic, cpb.QuizType_QUIZ_TYPE_POW)

	case cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_TAD)]:
		ctx, err = s.quizSetWithAllForExamLO(ctx, numOfQuizzes, topic, cpb.QuizType_QUIZ_TYPE_TAD)

	case cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MIQ)]:
		ctx, err = s.quizSetWithAllForExamLO(ctx, numOfQuizzes, topic, cpb.QuizType_QUIZ_TYPE_MIQ)

	case cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MAQ)]:
		ctx, err = s.quizSetWithAllForExamLO(ctx, numOfQuizzes, topic, cpb.QuizType_QUIZ_TYPE_MAQ)
	}

	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	if stepState.CurrentStudentID == "" {
		stepState.CurrentStudentID = idutil.ULIDNow()
	}
	if err := s.AuthHelper.AValidUser(ctx, stepState.CurrentStudentID, consta.RoleStudent); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create student: %w", err)
	}
	stepState.Token, err = s.AuthHelper.GenerateExchangeToken(stepState.CurrentStudentID, consta.RoleStudent)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("AuthHelper.GenerateExchangeToken: %w", err)
	}
	ctx = s.userCreateQuizTestWithValidRequestAndLimitTheFirstTimeForExamLO(ctx, limit)
	ctx, err = s.returnListOfQuizItems(ctx, limit)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.returnListOfQuizItems: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) returnListOfQuizItems(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	numOfQuizzes, _ := strconv.Atoi(arg1)

	resp, ok := stepState.Response.(*epb.CreateQuizTestResponse)
	if !ok {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("not receive create quiz test response")
	}
	stepState.ShuffledQuizSetID = resp.QuizzesId
	var limit int
	var offset int
	if stepState.NextPage == nil {
		limit = stepState.Limit
		offset = stepState.Offset
	} else {
		limit = int(stepState.NextPage.Limit)
		offset = int(stepState.NextPage.GetOffsetInteger())
	}
	stepState.NextPage = resp.NextPage
	stepState.SetID = resp.QuizzesId
	stepState.QuizItems = resp.Items
	if len(resp.Items) != numOfQuizzes {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expect number of quizzes are %v but got %v", numOfQuizzes, len(resp.Items))
	}

	quizExternalIDs := make([]string, 0)
	query := `SELECT quiz_external_id FROM shuffled_quiz_sets sqs INNER JOIN UNNEST(sqs.quiz_external_ids) AS quiz_external_id ON shuffled_quiz_set_id = $1 LIMIT $2 OFFSET $3;
	`

	rows, err := s.EurekaDB.Query(ctx, query, stepState.SetID, limit, offset-1)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()
	for rows.Next() {
		var id pgtype.Text
		err := rows.Scan(&id)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
		quizExternalIDs = append(quizExternalIDs, id.String)
	}

	if len(resp.Items) != len(quizExternalIDs) {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expect number of quizzes are %v but got %v", quizExternalIDs, len(resp.Items))
	}

	for i := range resp.Items {
		if resp.Items[i].Core.ExternalId != quizExternalIDs[i] {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz id %v but got %v", len(quizExternalIDs), resp.Items[i].Core.ExternalId)
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

// nolint
func (s *Suite) quizSetWithAllForExamLO(ctx context.Context, numberOfQuizzes, topicType string, quizType cpb.QuizType) (context.Context, error) {
	ctx, err1 := s.learningObjectiveBelongedToATopicForExamLO(ctx)
	ctx, err2 := s.ListOfAllQuiz(ctx, numberOfQuizzes, quizType)
	ctx, err3 := s.aQuizset(ctx)
	return ctx, multierr.Combine(err1, err2, err3)
}

func (s *Suite) aQuizset(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	quizExternalIDs := []string{}
	for _, quiz := range stepState.Quizzes {
		quizExternalIDs = append(quizExternalIDs, quiz.ExternalID.String)
	}

	questionHierarchy := make([]interface{}, 0)
	for _, extID := range quizExternalIDs {
		questionHierarchy = append(questionHierarchy, &entities.QuestionHierarchyObj{
			ID:   extID,
			Type: entities.QuestionHierarchyQuestion,
		})
	}

	quizSet := entities.QuizSet{}
	database.AllNullEntity(&quizSet)

	quizSet.ID = database.Text(idutil.ULIDNow())
	quizSet.LoID = database.Text(stepState.LoID)
	quizSet.QuizExternalIDs = database.TextArray(quizExternalIDs)
	quizSet.Status = database.Text("QUIZSET_STATUS_APPROVED")
	quizSet.QuestionHierarchy = database.JSONBArray(questionHierarchy)

	stepState.QuizSet = quizSet

	quizSetRepo := repositories.QuizSetRepo{}
	err := quizSetRepo.Create(ctx, s.EurekaDB, &stepState.QuizSet)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userCreateQuizTestWithValidRequestAndLimitTheFirstTimeForExamLO(ctx context.Context, arg1 string) context.Context {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Limit, _ = strconv.Atoi(arg1)
	stepState.Offset = 1
	stepState.NextPage = nil
	stepState.SetID = ""
	stepState.SessionID = strconv.Itoa(rand.Int())
	request := &epb.CreateQuizTestRequest{
		LoId:            stepState.LoID,
		StudentId:       stepState.CurrentStudentID,
		StudyPlanItemId: stepState.StudyPlanItemID,
		SessionId:       stepState.SessionID,
		Paging: &cpb.Paging{
			Limit: uint32(stepState.Limit),
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 1,
			},
		},
	}
	ctx = s.executeCreateQuizTestService(ctx, request)
	return utils.StepStateToContext(ctx, stepState)
}

func (s *Suite) executeCreateQuizTestService(ctx context.Context, request *epb.CreateQuizTestRequest) context.Context {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.EurekaConn).CreateQuizTest(s.AuthHelper.SignedCtx(ctx, stepState.Token), request)
	return utils.StepStateToContext(ctx, stepState)
}

func (s *Suite) aListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	bookResp, err := epb.NewBookModifierServiceClient(s.EurekaConn).UpsertBooks(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertBooksRequest{
		Books: utils.GenerateBooks(1, nil),
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create book: %w", err)
	}
	stepState.BookID = bookResp.BookIds[0]
	chapterResp, err := epb.NewChapterModifierServiceClient(s.EurekaConn).UpsertChapters(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertChaptersRequest{
		Chapters: utils.GenerateChapters(stepState.BookID, 1, nil),
		BookId:   stepState.BookID,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	stepState.ChapterID = chapterResp.ChapterIds[0]
	topics := utils.GenerateTopics(stepState.ChapterID, 3, nil)
	topics[0].Type = epb.TopicType_TOPIC_TYPE_LEARNING
	topics[1].Type = epb.TopicType_TOPIC_TYPE_EXAM
	topics[2].Type = epb.TopicType_TOPIC_TYPE_PRACTICAL
	stepState.Topics = topics
	stepState.Request = &epb.UpsertTopicsRequest{Topics: topics}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminInsertsAListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)

	resp, err := epb.NewTopicModifierServiceClient(s.EurekaConn).Upsert(
		s.AuthHelper.SignedCtx(ctx, stepState.Token), stepState.Request.(*epb.UpsertTopicsRequest))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.Response = resp

	stepState.TopicIDs = resp.GetTopicIds()
	stepState.TopicID = stepState.TopicIDs[0]
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) generateLearningObjective1(ctx context.Context) *cpb.LearningObjective {
	stepState := utils.StepStateFromContext[StepState](ctx)
	id := idutil.ULIDNow()

	return &cpb.LearningObjective{
		Info: &cpb.ContentBasicInfo{
			Id:        id,
			Name:      fmt.Sprintf("learning-%s", id),
			Country:   cpb.Country_COUNTRY_VN,
			Grade:     12,
			Subject:   cpb.Subject_SUBJECT_MATHS,
			MasterId:  "",
			SchoolId:  constants.ManabieSchool,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
		TopicId: stepState.TopicID,
		Prerequisites: []string{
			"AL-PH3.1", "AL-PH3.2",
		},
		StudyGuide: "https://guides/1/master",
		Video:      "https://videos/1/master",
	}
}

func (s *Suite) addBookToCourse(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	now := time.Now()
	cbe := &entities.CoursesBooks{}
	database.AllNullEntity(cbe)
	if err := multierr.Combine(
		cbe.BookID.Set(stepState.BookID),
		cbe.CourseID.Set(stepState.CourseID),
		cbe.CreatedAt.Set(now),
		cbe.UpdatedAt.Set(now),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to set value for course book: %w", err)
	}
	if _, err := database.Insert(ctx, cbe, s.EurekaDB.Exec); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to crete course book: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) aLearningObjectiveBelongedToATopicForExamLO(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	s.aSignedIn(ctx, "school admin")
	if len(stepState.TopicIDs) == 0 {
		return utils.StepStateToContext(ctx, stepState), errors.New("topic can't empty")
	}

	lo := s.generateLearningObjective1(ctx)
	lo.Type = cpb.LearningObjectiveType_LEARNING_OBJECTIVE_TYPE_EXAM_LO
	lo.ManualGrading = true

	if _, err := epb.NewLearningObjectiveModifierServiceClient(s.EurekaConn).UpsertLOs(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertLOsRequest{
		LearningObjectives: []*cpb.LearningObjective{
			lo,
		},
	}); err != nil {
		stepState.ResponseErr = err
		return utils.StepStateToContext(ctx, stepState), nil
	}

	stepState.LoID = lo.Info.Id
	stepState.LOIDs = append(stepState.LOIDs, stepState.LoID)
	stepState.Request = stepState.LoID

	ctx, err := s.addBookToCourse(ctx)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to add book to course: %w", err)
	}

	resp, err := epb.NewStudyPlanModifierServiceClient(s.EurekaConn).UpsertStudyPlan(s.AuthHelper.SignedCtx(ctx, stepState.Token), &epb.UpsertStudyPlanRequest{
		SchoolId: constants.ManabieSchool,
		Name:     idutil.ULIDNow(),
		CourseId: stepState.CourseID,
		BookId:   stepState.BookID,
		Status:   epb.StudyPlanStatus_STUDY_PLAN_STATUS_ACTIVE,
	})
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	var (
		spie entities.StudyPlanItem
		spe  entities.StudyPlan
	)
	stmt := fmt.Sprintf(`
	WITH TMP AS (
		SELECT study_plan_id
		FROM %s
		WHERE study_plan_id = $1
		OR master_study_plan_id = $1
	)
	UPDATE %s
	SET available_from = $2, available_to = $3, start_date = $4, updated_at = NOW()
	WHERE study_plan_id IN(SELECT * FROM TMP)
	`, spe.TableName(), spie.TableName())

	if _, err := s.EurekaDB.Exec(ctx,
		stmt,
		&resp.StudyPlanId,
		database.Timestamptz(time.Now().Add(-3*time.Hour)),
		database.Timestamptz(time.Now().Add(3*time.Hour)),
		database.Timestamptz(time.Now()),
	); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable update available date: %w", err)
	}

	if _, err := s.EurekaDB.Exec(ctx, `UPDATE study_plan_items SET completed_at = NOW() WHERE content_structure->>'lo_id' = $1`, database.Text(stepState.LoID)); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable update completed_at: %w", err)
	}

	query := "SELECT study_plan_item_id from study_plan_items where content_structure->>'lo_id' = $1 AND copy_study_plan_item_id IS NOT NULL"
	var studyPlanItemID string
	if err := s.EurekaDB.QueryRow(ctx, query, database.Text(stepState.LoID)).Scan(&studyPlanItemID); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	stepState.StudyPlanItemID = studyPlanItemID
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) learningObjectiveBelongedToATopicForExamLO(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.aSignedIn(ctx, "school admin")
	ctx, err2 := s.aListOfValidTopics(ctx)
	ctx, err3 := s.adminInsertsAListOfValidTopics(ctx)
	ctx, err4 := s.aLearningObjectiveBelongedToATopicForExamLO(ctx)
	return ctx, multierr.Combine(err1, err2, err3, err4)
}

func (s *Suite) aListOfValidChaptersInDB(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for i := 0; i < 10; i++ {
		c := new(entities.Chapter)
		database.AllNullEntity(c)
		now := time.Now()
		id := fmt.Sprintf("chapter_id_%d", i)
		name := fmt.Sprintf("chapter_name_%d", i)
		c.ID.Set(id)
		c.Name.Set(name)
		c.CreatedAt.Set(now)
		c.UpdatedAt.Set(now)
		c.Country.Set(epb.Country_COUNTRY_VN.String())
		c.Grade.Set(1)
		c.Subject.Set(epb.Subject_SUBJECT_CHEMISTRY.String())
		c.DisplayOrder.Set(1)
		c.SchoolID.Set(stepState.SchoolIDInt)
		c.DeletedAt.Set(nil)

		_, err := database.Insert(ctx, c, s.EurekaDB.Exec)
		if e, ok := err.(*pgconn.PgError); ok && e.Code != "23505" {
			return utils.StepStateToContext(ctx, stepState), err
		}
		stepState.CurrentChapterIDs = append(stepState.CurrentChapterIDs, id)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) AListOfValidTopics(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	s.aListOfValidChaptersInDB(ctx)

	s.BobDB.Exec(ctx, `
	(chapter_id, name, country, subject, grade, display_order, updated_at, created_at)
	VALUES
	('chapter_id_1', 'chapter_id_1', 'COUNTRY_VN', 'SUBJECT_BIOLOGY', 12, 1, now(), now()),
	('chapter_id_2', 'chapter_id_2', 'COUNTRY_VN', 'SUBJECT_BIOLOGY', 12, 1, now(), now()),
	('chapter_id_3', 'chapter_id_3', 'COUNTRY_VN', 'SUBJECT_BIOLOGY', 12, 1, now(), now())
	ON CONFLICT DO NOTHING;`)
	if stepState.ChapterID == "" {
		if ctx, err := s.insertAChapter(ctx); err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve ")
		}
	}
	t1 := s.generateValidTopic(stepState.ChapterID)
	t1.ChapterId = "chapter_id_1"
	t1.Type = epb.TopicType_TOPIC_TYPE_LEARNING
	stepState.Topics = append(stepState.Topics, &t1)

	t2 := s.generateValidTopic(stepState.ChapterID)
	t2.ChapterId = "chapter_id_2"
	t2.Type = epb.TopicType_TOPIC_TYPE_EXAM
	stepState.Topics = append(stepState.Topics, &t2)

	t3 := s.generateValidTopic(stepState.ChapterID)
	t3.ChapterId = "chapter_id_3"
	t3.Type = epb.TopicType_TOPIC_TYPE_PRACTICAL
	stepState.Topics = append(stepState.Topics, &t3)

	stepState.Request = &epb.UpsertTopicsRequest{Topics: []*epb.Topic{&t1, &t2, &t3}}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) insertAChapter(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	chapter := &entities.Chapter{}
	now := time.Now()
	database.AllNullEntity(chapter)
	stepState.ChapterID = idutil.ULIDNow()
	multierr.Combine(
		chapter.ID.Set(stepState.ChapterID),
		chapter.Country.Set(epb.Country_COUNTRY_VN),
		chapter.Name.Set(fmt.Sprintf("name-%s", stepState.ChapterID)),
		chapter.Grade.Set(12),
		chapter.SchoolID.Set(stepState.SchoolIDInt),
		chapter.CurrentTopicDisplayOrder.Set(0),
		chapter.CreatedAt.Set(now),
		chapter.UpdatedAt.Set(now),
	)
	_, err := database.Insert(ctx, chapter, s.EurekaDB.Exec)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create a chapter: %w", err)
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) generateValidTopic(chapterID string) epb.Topic {
	return epb.Topic{
		Id:           idutil.ULIDNow(),
		Name:         "topic 1",
		Country:      epb.Country_COUNTRY_VN,
		Grade:        "G12",
		Subject:      epb.Subject_SUBJECT_MATHS,
		Type:         epb.TopicType_TOPIC_TYPE_LEARNING,
		CreatedAt:    timestamppb.Now(),
		UpdatedAt:    timestamppb.Now(),
		Status:       epb.TopicStatus_TOPIC_STATUS_NONE,
		DisplayOrder: 1,
		PublishedAt:  timestamppb.Now(),
		SchoolId:     constant.ManabieSchool,
		IconUrl:      "topic-icon",
		ChapterId:    chapterID,
	}
}

func (s *Suite) ListOfAllQuiz(ctx context.Context, numStr string, quizType cpb.QuizType) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	numOfQuizzes, _ := strconv.Atoi(numStr)
	stepState.Quizzes = entities.Quizzes{}
	for i := 0; i < numOfQuizzes; i++ {
		quiz := s.genQuiz(ctx, quizType)
		stepState.Quizzes = append(stepState.Quizzes, quiz)
	}

	quizRepo := repositories.QuizRepo{}
	for _, quiz := range stepState.Quizzes {
		err := quizRepo.Create(ctx, s.EurekaDB, quiz)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), err
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) allStudentsAnswerAndSubmitQuizzesBelongToExam(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	for _, studentID := range stepState.StudentIDs {
		stepState.CurrentStudentID = studentID
		ctx, err := s.createQuizTestsAndAnswersForExamLO(ctx)
		if err != nil {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("s.createQuizTestsAndAnswersForExamLO: %w", err)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}
func (s *Suite) listExamLOSubmissionsWithFilterBy(ctx context.Context, filter string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	stepState.TeacherID = idutil.ULIDNow()
	if err := s.AuthHelper.AValidUser(ctx, stepState.TeacherID, consta.RoleTeacher); err != nil {
		return utils.StepStateToContext(ctx, stepState), fmt.Errorf("unable to create teacher: %w", err)
	}
	token, err := s.AuthHelper.GenerateExchangeToken(stepState.TeacherID, consta.RoleTeacher)
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.Token = token
	var (
		studentNameFilter *wrapperspb.StringValue
		examNameFilter    *wrapperspb.StringValue
		teacherId         *wrapperspb.StringValue
	)
	randomValidUserID := stepState.StudentIDs[rand.Intn(len(stepState.StudentIDs))]
	randomValidExamID := stepState.LOIDs[rand.Intn(len(stepState.LOIDs))]
	switch filter {
	case "student name":
		studentNameFilter = wrapperspb.String(randomValidUserID)
	case "random student name":
		studentNameFilter = wrapperspb.String("random student bla bla")
	case "exam name":
		examNameFilter = wrapperspb.String(randomValidExamID)
	case "random exam name":
		examNameFilter = wrapperspb.String("random exam bla bla")
	case "student name and exam name":
		studentNameFilter = wrapperspb.String(randomValidUserID)
		examNameFilter = wrapperspb.String(randomValidExamID)
	case "special character":
		studentNameFilter = wrapperspb.String("'")
		examNameFilter = wrapperspb.String("'")
	case "corrector":
		teacherId = &wrapperspb.StringValue{
			Value: stepState.CorrectorID,
		}
	}

	req := &sspb.ListExamLOSubmissionRequest{
		CourseId:    wrapperspb.String(stepState.CourseID),
		LocationIds: stepState.LocationIDs,
		Start:       timestamppb.New(time.Now().Add(-7 * 24 * time.Hour)),
		End:         timestamppb.New(time.Now().Add(7 * 24 * time.Hour)),
		Paging: &cpb.Paging{
			Limit: 100,
		},
		StudentName: studentNameFilter,
		ExamName:    examNameFilter,
		CorrectorId: teacherId,
		SubmittedDate: &sspb.ListExamLOSubmissionRequest_SubmittedDate{
			Start: timestamppb.New(time.Now().Add(-2 * 24 * time.Hour)),
			End:   timestamppb.New(time.Now()),
		},
		LastUpdatedDate: &sspb.ListExamLOSubmissionRequest_LastUpdatedDate{
			Start: timestamppb.New(time.Now().Add(-2 * 24 * time.Hour)),
			End:   timestamppb.New(time.Now()),
		},
	}
	stepState.Response, stepState.ResponseErr = sspb.NewExamLOClient(s.EurekaConn).ListExamLOSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token), req)
	stepState.Request = req
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) teacherGradesSubmission(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	response := stepState.Response.(*sspb.ListExamLOSubmissionResponse)
	stepState.CorrectorID = stepState.TeacherID
	for _, item := range response.GetItems() {
		_, err := sspb.NewExamLOClient(s.EurekaConn).GradeAManualGradingExamSubmission(s.AuthHelper.SignedCtx(ctx, stepState.Token),
			&sspb.GradeAManualGradingExamSubmissionRequest{
				SubmissionId:      item.SubmissionId,
				ShuffledQuizSetId: item.ShuffledQuizSetId,
				SubmissionStatus:  sspb.SubmissionStatus_SUBMISSION_STATUS_MARKED,
				TeacherFeedback:   "aaaaaaaaa",
				TeacherExamGrades: []*sspb.TeacherExamGrade{},
			})
		if err != nil {
			stepState.ResponseErr = err
			return utils.StepStateToContext(ctx, stepState), nil
		}
	}

	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ourSystemMustReturnsListExamLOSubmissionsWithFilterCorrectly(ctx context.Context, filter string) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	resp := stepState.Response.(*sspb.ListExamLOSubmissionResponse)
	req := stepState.Request.(*sspb.ListExamLOSubmissionRequest)
	if filter == "random student name" || filter == "random exam name" {
		if len(resp.Items) != 0 {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected length of exam lo submissions: %d, but got: %d", 0, len(resp.Items))
		}
	}
	for _, item := range resp.Items {
		examName := fmt.Sprintf("learning-%s", item.StudyPlanItemIdentity.LearningMaterialId)
		if req.ExamName != nil {
			if !strings.Contains(examName, req.ExamName.Value) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected exam lo submission name contains: %s, but got: %s", req.ExamName.Value, examName)
			}
		}
		studentName := fmt.Sprintf("valid-user-import-by-eureka%s", item.StudyPlanItemIdentity.StudentId.Value)
		if req.StudentName != nil {
			if !strings.Contains(studentName, req.StudentName.Value) {
				return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected student name contains: %s, but got: %s", req.StudentName.Value, studentName)
			}
		}
		if !item.SubmittedAt.AsTime().After(req.SubmittedDate.Start.AsTime()) || !item.SubmittedAt.AsTime().Before(req.SubmittedDate.End.AsTime()) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected submitted date from %v to %v, but got %v", req.SubmittedDate.Start.AsTime(), req.SubmittedDate.End.AsTime(), item.SubmittedAt.AsTime())
		}
		if !item.UpdatedAt.AsTime().After(req.LastUpdatedDate.Start.AsTime()) || !item.UpdatedAt.AsTime().Before(req.LastUpdatedDate.End.AsTime()) {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("expected last updated date from %v to %v, but got %v", req.LastUpdatedDate.Start.AsTime(), req.LastUpdatedDate.End.AsTime(), item.UpdatedAt.AsTime())
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) genQuiz(ctx context.Context, quizType cpb.QuizType) *entities.Quiz {
	stepState := utils.StepStateFromContext[StepState](ctx)

	switch quizType {
	case cpb.QuizType_QUIZ_TYPE_MCQ:
		return s.genMultipleChoicesQuiz(stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_FIB:
		return s.genFillInTheBlankQuiz(stepState.CurrentUserID, stepState.LoID, stepState.UserFillInTheBlankOld)
	case cpb.QuizType_QUIZ_TYPE_TAD:
		return s.genTermAndDefinitionQuiz(ctx)
	case cpb.QuizType_QUIZ_TYPE_POW:
		return s.genPairOfWordQuiz(stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_MIQ:
		return s.genManualInputQuiz(ctx, stepState.CurrentUserID, stepState.LoID)
	case cpb.QuizType_QUIZ_TYPE_MAQ:
		return s.genMultipleChoicesQuiz(stepState.CurrentUserID, stepState.LoID)
	}
	return nil
}

func (s *Suite) genMultipleChoicesQuiz(currentUserID, loID string) *entities.Quiz {
	quizRawObj := raw{
		Blocks: []block{
			{
				Key:               "eq20k",
				Text:              "3213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
			},
		},
	}
	quizRaw, _ := json.Marshal(quizRawObj)

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "B",
			Key:         "key-B",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "B",
			Key:         "key-B",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "C",
			Key:         "key-C",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "C",
			Key:         "key-C",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "D",
			Key:         "key-D",
			Attribute:   entities.QuizItemAttribute{},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(idutil.ULIDNow())
	quiz.ExternalID = database.Text(idutil.ULIDNow())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MCQ)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	return quiz
}

func (s *Suite) genManualInputQuiz(_ context.Context, currentUserID, loID string) *entities.Quiz {
	quizRawObj := raw{
		Blocks: []block{
			{
				Key:               "eq20k",
				Text:              "3213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
			},
		},
	}
	quizRaw, _ := json.Marshal(quizRawObj)

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(idutil.ULIDNow())
	quiz.ExternalID = database.Text(idutil.ULIDNow())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_MIQ)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	return quiz
}

func (s *Suite) genFillInTheBlankQuiz(currentUserID, loID string, userFillInTheBlankOld bool) *entities.Quiz {
	quizRawObj := raw{
		Blocks: []block{
			{
				Key:               "eq20k",
				Text:              "3213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
			},
		},
	}
	quizRaw, _ := json.Marshal(quizRawObj)

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/b9dbb04803e7cdde2e072edfd632809f.html",
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "A",
			Key:         "key-A",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "B",
			Key:         "key-B",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "B",
			Key:         "key-B",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "C",
			Key:         "key-C",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "C",
			Key:         "key-C",
			Attribute:   entities.QuizItemAttribute{},
		},
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/43fa9074c32bde05978702604947f6a1.html",
			},
			Correctness: false,
			Configs:     []string{},
			Label:       "D",
			Key:         "key-D",
			Attribute:   entities.QuizItemAttribute{},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(idutil.ULIDNow())
	quiz.ExternalID = database.Text(idutil.ULIDNow())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_FIB)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(string(explanationQuestion))
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	if userFillInTheBlankOld {
		options, _ := quiz.GetOptions()
		for _, opt := range options {
			opt.Key = ""
		}
		quiz.Options.Set(options)
	}
	return quiz
}

func (s *Suite) genPairOfWordQuiz(currentUserID, loID string) *entities.Quiz {
	quizRawObj := raw{
		Blocks: []block{
			{
				Key:               "1c0o5",
				Text:              "Banana",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
			},
		},
	}
	quizRaw, _ := json.Marshal(quizRawObj)

	quizQuestionObj := &entities.QuizQuestion{
		Raw:         string(quizRaw),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/d41d8cd98f00b204e9800998ecf8427e.html",
		Attribute: entities.QuizItemAttribute{
			ImgLink:   "https://storage.googleapis.com/stag-manabie-backend/user-upload/ce196d896889ce984b8c36f6c8ed64b001FF6SDP89K003NKW655N9CY1J.jpg",
			AudioLink: "https://storage.googleapis.com/stag-manabie-backend/user-upload/Banana01FFF0KPK7RBFXTFWFBFMPB81C.mp3",
			Configs:   []string{"FLASHCARD_LANGUAGE_CONFIG_ENG"},
		},
	}
	quizQuestion, _ := json.Marshal(quizQuestionObj)

	explanationObj := raw{
		Blocks: []block{
			{
				Key:               "4rpf3",
				Text:              "213213",
				Type:              "unstyled",
				Depth:             "0",
				InlineStyleRanges: []string{},
				EntityRanges:      []string{},
				Data:              nil,
				EntityMap:         nil,
			},
		},
	}
	explanation, _ := json.Marshal(explanationObj)

	explanationQuestionObj := &entities.QuizQuestion{
		Raw:         string(explanation),
		RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/ee8b6810089778c7021a70298399256c.html",
	}
	explanationQuestion, _ := json.Marshal(explanationQuestionObj)

	quizOptionObjs := []*entities.QuizOption{
		{
			Content: entities.RichText{
				Raw:         string(quizRaw),
				RenderedURL: "https://storage.googleapis.com/stag-manabie-backend/content/d41d8cd98f00b204e9800998ecf8427e.html",
			},
			Correctness: true,
			Configs:     []string{},
			Label:       "A",
			Key:         "01FFF0KP9ZBMV0X0RKJ9NMB33V",
			Attribute: entities.QuizItemAttribute{
				ImgLink:   "https://storage.googleapis.com/stag-manabie-backend/user-upload/ce196d896889ce984b8c36f6c8ed64b001FF6SDP89K003NKW655N9CY1J.jpg",
				AudioLink: "https://storage.googleapis.com/stag-manabie-backend/user-upload/Banana%20term01FFF0KPX73PNPF4MYBWA8ZX1D.mp3",
				Configs:   []string{cpb.QuizItemAttributeConfig_FLASHCARD_LANGUAGE_CONFIG_ENG.String()},
			},
		},
	}
	quizOptions, _ := json.Marshal(quizOptionObjs)

	quiz := &entities.Quiz{}
	database.AllNullEntity(quiz)
	quiz.ID = database.Text(idutil.ULIDNow())
	quiz.ExternalID = database.Text(idutil.ULIDNow())
	quiz.Country = database.Text("COUNTRY_VN")
	quiz.SchoolID = database.Int4(-2147483648)
	quiz.LoIDs = database.TextArray([]string{loID})
	quiz.Kind = database.Text(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_POW)])
	quiz.Question = database.JSONB(string(quizQuestion))
	quiz.Explanation = database.JSONB(explanationQuestion)
	quiz.Options = database.JSONB(string(quizOptions))
	quiz.TaggedLOs = database.TextArray([]string{"VN10-CH-01-L-001.1"})
	quiz.DifficultLevel = database.Int4(1)
	quiz.CreatedBy = database.Text(currentUserID)
	quiz.ApprovedBy = database.Text(currentUserID)
	quiz.Status = database.Text("QUIZ_STATUS_APPROVED")
	return quiz
}

func (s *Suite) genTermAndDefinitionQuiz(ctx context.Context) *entities.Quiz {
	stepState := utils.StepStateFromContext[StepState](ctx)
	quiz := s.genPairOfWordQuiz(stepState.CurrentUserID, stepState.LoID)
	_ = quiz.Kind.Set(cpb.QuizType_name[int32(cpb.QuizType_QUIZ_TYPE_TAD)])
	return quiz
}
