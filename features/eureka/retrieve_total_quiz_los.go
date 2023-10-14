package eureka

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"go.uber.org/multierr"
)

func (s *suite) aLearningObjectiveBelongedToATopicHasQuizsetWithQuizzes(ctx context.Context, topicType string, arg2 int) (context.Context, error) {
	numberOfQuizzes := strconv.Itoa(arg2)
	ctx, err1 := s.aSignedIn(ctx, "school admin")
	ctx, err2 := s.aListOfValidTopics(ctx)
	ctx, err3 := s.adminInsertsAListOfValidTopics(ctx)
	ctx, err4 := s.aLearningObjectiveBelongedToATopic(ctx, topicType)
	ctx, err5 := s.aListOfQuizzes(ctx, numberOfQuizzes)
	ctx, err6 := s.aQuizset(ctx)
	err := multierr.Combine(err1, err2, err3, err4, err5, err6)
	if err != nil {
		return ctx, err
	}
	stepState := StepStateFromContext(ctx)
	stepState.LOIDs = append(stepState.LOIDs, stepState.LoID)
	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) aLearningObjectiveBelongedToATopicHasNoQuizset(ctx context.Context, topicType string) (context.Context, error) {
	ctx, err1 := s.aSignedIn(ctx, "school admin")
	ctx, err2 := s.aListOfValidTopics(ctx)
	ctx, err3 := s.adminInsertsAListOfValidTopics(ctx)
	ctx, err4 := s.aLearningObjectiveBelongedToATopic(ctx, topicType)
	err := multierr.Combine(err1, err2, err3, err4)
	stepState := StepStateFromContext(ctx)
	stepState.LOIDs = append(stepState.LOIDs, stepState.LoID)
	return StepStateToContext(ctx, stepState), err
}

// nolint
func (s *suite) totalQuizSetIs(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	temp := strings.Split(arg1, ",")
	expect := make(map[string]int32)
	for i := 0; i < len(temp); i++ {
		temp[i] = strings.TrimSpace(temp[i])
		c, err := strconv.Atoi(temp[i])
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		expect[stepState.LOIDsInReq[i]] = int32(c)
	}

	resp := stepState.Response.(*epb.RetrieveTotalQuizLOsResponse)

	total := resp.LosTotalQuiz

	if len(total) != len(expect) {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v but got %v", len(expect), len(total))
	}

	for _, loTotalQuiz := range total {
		if loTotalQuiz.TotalQuiz != expect[loTotalQuiz.LoId] {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect %v but got %v", loTotalQuiz.TotalQuiz, expect[loTotalQuiz.LoId])
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userGetTotalQuizOfLoWithoutLoIds(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveTotalQuizLOs(s.signedCtx(ctx), &epb.RetrieveTotalQuizLOsRequest{})

	return StepStateToContext(ctx, stepState), nil
}
func (s *suite) userGetTotalQuizOfLo(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	temp := strings.Split(arg1, ",")
	idxs := make([]int, len(temp))
	for i := 0; i < len(temp); i++ {
		temp[i] = strings.TrimSpace(temp[i])
		idxs[i], _ = strconv.Atoi(temp[i])
	}

	stepState.LOIDsInReq = make([]string, len(idxs))
	for i := range stepState.LOIDsInReq {
		stepState.LOIDsInReq[i] = stepState.LOIDs[idxs[i]-1]
	}
	stepState.Response, stepState.ResponseErr = epb.NewQuizReaderServiceClient(s.Conn).RetrieveTotalQuizLOs(s.signedCtx(ctx), &epb.RetrieveTotalQuizLOsRequest{
		LoIds: stepState.LOIDsInReq,
	})

	query := "SELECT COUNT(*) FROM learning_objectives WHERE lo_id = ANY($1)  AND deleted_at IS NULL"
	count := 0
	if err := try.Do(func(attempt int) (bool, error) {
		err := s.DB.QueryRow(ctx, query, stepState.LOIDsInReq).Scan(&count)
		if err != nil {
			return false, err
		}

		if count == 0 && count != len(stepState.LOIDsInReq) {
			time.Sleep(2 * time.Second)
			return attempt < 10, fmt.Errorf("expected number of learning_objectives %v, but got %v", len(stepState.LOIDsInReq), count)
		}

		return false, nil
	}); err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
