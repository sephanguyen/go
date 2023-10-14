package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/invoicemgmt/constant"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

func (s *suite) aStudyPlanItemIsLearningObjectiveBelongedToATopicWhichHasQuizsetWithQuizzes(ctx context.Context, topic, numQuizzes string) (context.Context, error) {
	ctx, err1 := s.aQuizsetWithQuizzesInLearningObjectiveBelongedToATopic(ctx, numQuizzes, topic)
	ctx, err2 := s.aStudyPlanItemId(ctx)
	return ctx, multierr.Combine(err1, err2)
}
func (s *suite) teacherRetrieveAllStudentsSubmissionHistoryInThatStudyPlanItem(ctx context.Context) (context.Context, error) {
	ctx, err1 := s.aSignedIn(ctx, "teacher")
	ctx, err2 := s.retrieveAllStudentsSubmissionHistory(ctx)
	return ctx, multierr.Combine(err1, err2)
}
func (s *suite) studentRetrieveAllStudentsSubmissionHistoryInThatStudyPlanItem(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.aSignedIn(ctx, "student")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	// we will skip this error func, and let the next step "returns "PermissionDenied" status code" which is defined in feature file handle this error
	s.retrieveAllStudentsSubmissionHistory(ctx)

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) getStudentsSubmissionHistory(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	nHistory, _ := strconv.Atoi(arg1)
	if stepState.NumShuffledQuizSetLogs != nHistory {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.getStudentsSubmissionHistory: number of submission history expect %v but got %v", arg1, stepState.NumShuffledQuizSetLogs)
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) retrieveAllStudentsSubmissionHistory(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.StudentDoingQuizExamLogs == nil {
		stepState.StudentDoingQuizExamLogs = make(map[string][]*cpb.AnswerLog)
	}
	time.Sleep(2 * time.Second)
	stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveQuizTests(s.signedCtx(ctx), &epb.RetrieveQuizTestsRequest{
		StudyPlanItemId: []string{stepState.StudyPlanItemID},
	})
	if stepState.ResponseErr != nil {
		return StepStateToContext(ctx, stepState), stepState.ResponseErr
	}

	resp, _ := stepState.Response.(*epb.RetrieveQuizTestsResponse)
	for k := range resp.Items {
		for _, item := range resp.Items[k].Items {
			stepState.SetIDs = append(stepState.SetIDs, item.SetId)
		}
	}

	stepState.Offset = 0
	for _, id := range stepState.SetIDs {
		logs := make([]*cpb.AnswerLog, 0)
		incompletedQuizzes := make([]string, 0)
		completedQuizzes := make([]string, 0)
		// first page
		stepState.NumShuffledQuizSetLogs++
		try.Do(func(attempt int) (bool, error) {
			stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveSubmissionHistory(s.signedCtx(ctx), &epb.RetrieveSubmissionHistoryRequest{
				SetId: id,
				Paging: &cpb.Paging{
					Limit: uint32(stepState.Limit),
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: int64(stepState.Offset),
					},
				},
			})
			if stepState.Response == nil {
				return false, nil
			}
			retry := attempt < 5
			if retry {
				time.Sleep(1 * time.Second)
				return true, fmt.Errorf("temporary error ListOfflineLearningRecords: %w", stepState.ResponseErr)
			}
			return false, stepState.ResponseErr
		})
		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}

		resp, _ := stepState.Response.(*epb.RetrieveSubmissionHistoryResponse)
		logs = append(logs, resp.Logs...)
		stepState.NextPage = resp.NextPage
		stepState.AnswerLogs = resp.Logs

		for _, log := range resp.Logs {
			if log.SubmittedAt == nil {
				incompletedQuizzes = append(incompletedQuizzes, log.QuizId)
				continue
			} else {
				completedQuizzes = append(completedQuizzes, log.QuizId)
			}
		}

		// next page
		for len(stepState.AnswerLogs) != 0 {
			stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveSubmissionHistory(s.signedCtx(ctx), &epb.RetrieveSubmissionHistoryRequest{
				SetId:  id,
				Paging: stepState.NextPage,
			})
			if stepState.ResponseErr != nil {
				return StepStateToContext(ctx, stepState), stepState.ResponseErr
			}

			resp, _ := stepState.Response.(*epb.RetrieveSubmissionHistoryResponse)
			logs = append(logs, resp.Logs...)
			for _, log := range resp.Logs {
				if log.SubmittedAt == nil {
					incompletedQuizzes = append(incompletedQuizzes, log.QuizId)
					continue
				} else {
					completedQuizzes = append(completedQuizzes, log.QuizId)
				}
			}

			stepState.NextPage = resp.NextPage
			stepState.AnswerLogs = resp.Logs
		}

		stepState.StudentDoingQuizExamLogs[id] = logs
		query := `
			SELECT COUNT(*) 
			FROM (
				SELECT UNNEST(quiz_external_ids) 
					FROM shuffled_quiz_sets 
					WHERE shuffled_quiz_set_id = $1 
				EXCEPT 
				SELECT value->>'quiz_id' 
					FROM shuffled_quiz_sets sqs 
						INNER JOIN JSONB_ARRAY_ELEMENTS(submission_history) 
						ON sqs.shuffled_quiz_set_id  = $1
			) AS FOO;
		`
		var count pgtype.Int8
		err := s.DB.QueryRow(ctx, query, id).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if int(count.Int) != len(incompletedQuizzes) {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect number of incompleted quizzes %v but got %v", count.Int, len(incompletedQuizzes))
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) studentsDidntFinishTheTest(ctx context.Context, numStudentsStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numStudents, _ := strconv.Atoi(numStudentsStr)

	stepState.Limit = rand.Intn(len(stepState.Quizzes)) + 1
	if stepState.QuizOptions == nil {
		stepState.QuizOptions = make(map[string]map[string][]*cpb.QuizOption)
	}

	for i := 0; i < numStudents; i++ {
		ctx, err1 := s.aSignedIn(ctx, constant.RoleStudent)
		ctx, err2 := s.aSignedIn(ctx, constant.RoleStudent)
		ctx, err3 := s.doQuizExam(ctx, stepState.Limit, true)
		if err := multierr.Combine(err1, err2, err3); err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) checkOrderedOfDoingQuizExamLogs(ctx context.Context, setID string, logs []*cpb.AnswerLog) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// check the ordered of returned quiz
	getQuizExternalIDsQuery := "SELECT ARRAY_AGG(quiz_id) FROM (SELECT quiz_id FROM shuffled_quiz_sets INNER JOIN UNNEST(quiz_external_ids) AS quiz_id ON shuffled_quiz_set_id = $1) AS quiz_ids;"
	var quizExternalIDs pgtype.TextArray

	err := s.DB.QueryRow(ctx, getQuizExternalIDsQuery, setID).Scan(&quizExternalIDs)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if len(quizExternalIDs.Elements) != len(logs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect returns %v logs but got %v", len(quizExternalIDs.Elements), len(logs))
	}
	for i := range quizExternalIDs.Elements {
		if quizExternalIDs.Elements[i].String != logs[i].QuizId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect returns quiz_external_id %v but got %v", quizExternalIDs.Elements[i].String, logs[i].QuizId)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) showCorrectLogsInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	for _, id := range stepState.SetIDs {
		for _, log := range stepState.StudentDoingQuizExamLogs[id] {
			if log.SubmittedAt == nil {
				// this is the log show that student didn't do this quiz with log.Core.ExternalId so we skip it
				continue
			}
			// expectedOpts is the options sequence when student do quiz before,
			// so the returned logs need to have the same options sequence
			expectedOpts, ok := stepState.QuizOptions[id][log.Core.ExternalId]
			if !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("not find quiz external id %v", log.Core.ExternalId)
			}

			for i, opt := range log.Core.Options {
				if opt.Content.Raw != expectedOpts[i].Content.Raw {
					return StepStateToContext(ctx, stepState), fmt.Errorf("expect option %v but got %v", expectedOpts[i].Content.Raw, opt.Content.Raw)
				}
			}

			selectedIdxAnswer := stepState.SelectedIndex[id][log.Core.ExternalId]
			if len(log.SelectedIndex) != len(selectedIdxAnswer) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expect len %v but got len %v student selected index: %v but got %v", len(selectedIdxAnswer), len(log.SelectedIndex), selectedIdxAnswer, log.SelectedIndex)
			}
			for i := range selectedIdxAnswer {
				if selectedIdxAnswer[i].GetSelectedIndex() != log.SelectedIndex[i] {
					return StepStateToContext(ctx, stepState), fmt.Errorf("expect student selected index: %v but got %v", selectedIdxAnswer, log.SelectedIndex)
				}
			}

			filledTextAnswer := stepState.FilledText[id][log.Core.ExternalId]
			if len(log.FilledText) != len(filledTextAnswer) {
				return StepStateToContext(ctx, stepState), fmt.Errorf("expect len %v but got %v student filled text: %v but got %v", len(filledTextAnswer), len(log.FilledText), filledTextAnswer, log.FilledText)
			}
			for i := range filledTextAnswer {
				if filledTextAnswer[i].GetFilledText() != log.FilledText[i] {
					return StepStateToContext(ctx, stepState), fmt.Errorf("expect student filled text: %v but got %v", filledTextAnswer, log.FilledText)
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) theOrderedOfLogsMustBeCorrect(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, id := range stepState.SetIDs {
		ctx, err := s.checkOrderedOfDoingQuizExamLogs(ctx, id, stepState.StudentDoingQuizExamLogs[id])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentTryToRetrieveSubmissionHistoryAfterTakeTheQuiz(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveSubmissionHistory(s.signedCtx(ctx), stepState.Request.(*epb.RetrieveSubmissionHistoryRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) eachItemHaveReturnedAdditionFieldsForFlashcard(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	for _, item := range stepState.AnswerLogs {
		if item.GetCore().Kind == cpb.QuizType_QUIZ_TYPE_POW {
			questionAtrrs := item.GetCore().Attribute
			if questionAtrrs.AudioLink == "" || questionAtrrs.ImgLink == "" || len(questionAtrrs.Configs) == 0 {
				return StepStateToContext(ctx, stepState), fmt.Errorf("response for flashcard is not correct for question")
			}
			for _, option := range item.Core.GetOptions() {
				if option.Attribute.AudioLink == "" || len(option.Attribute.Configs) == 0 {
					return StepStateToContext(ctx, stepState), fmt.Errorf("response for flashcard is not correct for option")
				}
			}
		}
	}

	return StepStateToContext(ctx, stepState), stepState.ResponseErr
}
