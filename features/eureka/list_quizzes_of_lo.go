package eureka

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"golang.org/x/exp/slices"
)

func (s *suite) returnsEmptyListOfQuizzes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if len(stepState.AnswerLogs) > 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect return empty response items but got %v", len(stepState.AnswerLogs))
	}
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) returnsExpectedListOfQuizzes(ctx context.Context, status string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error
	switch status {
	case "not existed":
		ctx, err = s.returnsEmptyListOfQuizzes(ctx)
	default:
		ctx, err = s.returnsListOfQuizzes(ctx)
	}
	return StepStateToContext(ctx, stepState), err
}
func (s *suite) returnsListOfQuizzes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	quiz := &entities.Quiz{}
	fields := database.GetFieldNames(quiz)
	for i := range fields {
		fields[i] = "q." + fields[i]
	}

	stmt := fmt.Sprintf(`SELECT %s 
		FROM quizzes q INNER JOIN learning_objectives lo ON lo.school_id = q.school_id AND lo_id = $1
		INNER JOIN (
			SELECT x.id, x.idx FROM quiz_sets qs, UNNEST(qs.quiz_external_ids) WITH ORDINALITY AS x(id, idx)
			WHERE qs.lo_id = $1 AND deleted_at IS NULL
		) AS t ON t.id = q.external_id
		AND q.deleted_at IS NULL
		ORDER BY t.idx ASC`, strings.Join(fields, ", "))
	rows, err := s.DB.Query(ctx, stmt, database.Text(stepState.LoID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	defer rows.Close()

	quizzesEnt := entities.Quizzes{}
	err = database.Select(ctx, s.DB, stmt, database.Text(stepState.LoID)).ScanAll(&quizzesEnt)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	quizzes := ([]*entities.Quiz)(quizzesEnt)

	if len(quizzes) != len(stepState.AnswerLogs) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect return %v items but got %v", len(quizzes), len(stepState.AnswerLogs))
	}

	for i := range quizzes {
		// return correct order of quizzes
		if quizzes[i].ExternalID.String != stepState.AnswerLogs[i].Core.ExternalId {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect return %v quiz external id but got %v", quizzes[i].ExternalID.String, stepState.AnswerLogs[i].Core.ExternalId)
		}
	}

	// validate correct order keys answer of ordering questions
	for _, answerLog := range stepState.AnswerLogs {
		if answerLog.IsAccepted {
			return ctx, fmt.Errorf("expected IsAccepted false but got true")
		}
		if answerLog.SubmittedAt != nil {
			return ctx, fmt.Errorf("expected SubmittedAt is zero but got %v", answerLog.SubmittedAt.AsTime())
		}

		switch answerLog.QuizType {
		case cpb.QuizType_QUIZ_TYPE_ORD:
			if len(answerLog.CorrectText) != 0 ||
				len(answerLog.CorrectIndex) != 0 ||
				len(answerLog.FilledText) != 0 ||
				len(answerLog.SelectedIndex) != 0 ||
				len(answerLog.Correctness) != 0 {
				return ctx, fmt.Errorf("all field CorrectText, CorrectIndex, FilledText, SelectedIndex, Correctness must empty but still have data %v", answerLog)
			}
			res := answerLog.Result.(*cpb.AnswerLog_OrderingResult)
			expectedResult := &cpb.OrderingResult{
				SubmittedKeys: []string{},
				CorrectKeys:   []string{"key-A", "key-B", "key-C"},
			}
			if !slices.Equal(res.OrderingResult.SubmittedKeys, expectedResult.SubmittedKeys) ||
				!slices.Equal(res.OrderingResult.CorrectKeys, expectedResult.CorrectKeys) {
				return ctx, fmt.Errorf(`expected result of ordering question is : %v but got %v`, expectedResult, res.OrderingResult)
			}
		default:
			fmt.Printf("plz implement for type %s soon \n", answerLog.QuizType.String())
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) teacherGetQuizzesOfLo(ctx context.Context, status string) (context.Context, error) {
	ctx, _ = s.aSignedIn(ctx, "teacher")

	var err error
	stepState := StepStateFromContext(ctx)
	switch status {
	case "not existed":
		ctx, err = s.userListQuizzesOfLO(ctx, "")
	default:
		ctx, err = s.userListQuizzesOfLO(ctx, stepState.LoID)
	}

	return StepStateToContext(ctx, stepState), err
}
func (s *suite) userListQuizzesOfLO(ctx context.Context, loID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var nextPage *cpb.Paging
	for {
		resp, err := epb.NewQuizReaderServiceClient(s.Conn).ListQuizzesOfLO(s.signedCtx(ctx), &epb.ListQuizzesOfLORequest{
			LoId:   loID,
			Paging: nextPage,
		})

		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		items := resp.Logs
		nextPage = resp.NextPage

		if len(items) == 0 {
			break
		}
		stepState.Response = resp
		stepState.ResponseErr = err
		stepState.AnswerLogs = append(stepState.AnswerLogs, items...)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedTotalQuestionWhenRetrieveLO(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	resp, err := epb.NewCourseReaderServiceClient(s.Conn).RetrieveLOs(s.signedCtx(ctx), &epb.RetrieveLOsRequest{
		StudentId: stepState.CurrentStudentID,
		LoIds:     []string{stepState.LoID},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	totalQuestions := resp.TotalQuestions[stepState.LoID]
	if totalQuestions != int32(len(stepState.Quizzes)) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect return %v quizzes but got %v", len(stepState.Quizzes), totalQuestions)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsExpectedTotalQuestionWhenRetrieveLOV1(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var totalQuestions int32
	if err := try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)
		resp, err := epb.NewCourseReaderServiceClient(s.Conn).RetrieveLOs(s.signedCtx(ctx), &epb.RetrieveLOsRequest{
			StudentId: stepState.CurrentStudentID,
			LoIds:     []string{stepState.LoID},
		})
		if err != nil {
			return true, err
		}
		totalQuestions = resp.TotalQuestions[stepState.LoID]
		if totalQuestions != int32(len(stepState.Quizzes)) {
			return attempt < 10, fmt.Errorf("expect return %v quizzes but got %v", len(stepState.Quizzes), totalQuestions)
		}
		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
