package eureka

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *suite) ourSystemHaveToReturnTheRetryQuizzesCorrectly(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp := stepState.Response.(*epb.CreateRetryQuizTestResponse)
	stmt := `SELECT coalesce(array_agg(DISTINCT(value ->>'quiz_id')),ARRAY[]::TEXT[]) external_ids FROM shuffled_quiz_sets CROSS JOIN jsonb_array_elements(submission_history) WHERE shuffled_quiz_set_id=$1 AND value ->>'is_accepted'='true'`
	var externalIDsWithCorrectAns pgtype.TextArray
	err := database.Select(ctx, s.DB, stmt, stepState.ShuffledQuizSetID).ScanFields(&externalIDsWithCorrectAns)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve external ids with correct ans: %w", err)
	}
	mapCorrectQuizIDs := make(map[string]bool)
	for _, e := range externalIDsWithCorrectAns.Elements {
		mapCorrectQuizIDs[e.String] = true
	}
	for _, e := range resp.GetItems() {
		if _, ok := mapCorrectQuizIDs[e.Core.ExternalId]; ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("the question with correct answer have to not exist on retry quiz response")
		}
	}
	shuffleQuizSetRepo := &repositories.ShuffledQuizSetRepo{}
	retryshufflequizzes, err := shuffleQuizSetRepo.Retrieve(ctx, s.DB, database.TextArray([]string{resp.QuizzesId}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve shuffle quizset: %w", err)
	}
	if len(retryshufflequizzes) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found any retry shuffle quiz")
	}
	shuffleQuizs, err := shuffleQuizSetRepo.Retrieve(ctx, s.DB, database.TextArray([]string{stepState.ShuffledQuizSetID}))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to retrieve shuffle quizset: %w", err)
	}
	if len(shuffleQuizs) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found any shuffle quiz")
	}
	if retryshufflequizzes[0].OriginalShuffleQuizSetID.String != shuffleQuizs[0].ID.String {
		return StepStateToContext(ctx, stepState), fmt.Errorf("the endpoint work wrong, the original shuffle quizset ID not equal to origin shuffle quiz")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentDoesTheQuizSetAndWrongSomeQuizzes(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.StudyPlanItemID = s.newID()
	ctx, err := s.studentDoQuizTestTheFirstTime(ctx)
	if err != nil {
		return ctx, fmt.Errorf("unable to do quiz test: %w", err)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) theStudentChooseOptionRetryQuiz(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, err := epb.NewQuizModifierServiceClient(s.Conn).CreateRetryQuizTest(s.signedCtx(ctx), &epb.CreateRetryQuizTestRequest{
		StudyPlanItemId: stepState.StudyPlanItemID,
		LoId:            stepState.LoID,
		StudentId:       stepState.CurrentStudentID,
		SetId:           wrapperspb.String(stepState.ShuffledQuizSetID),
		SessionId:       idutil.ULIDNow(),
		Paging: &cpb.Paging{
			Limit:  10,
			Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 1},
		},
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unable to create retry quiz: %w", err)
	}
	stepState.Response = resp
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) studentDoQuizTestTheFirstTime(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var limit int
	if len(stepState.Quizzes) > 1 {
		limit = rand.Intn(len(stepState.Quizzes)-1) + 1
	} else {
		limit = 0
	}
	ctx, err := s.doQuizExam(ctx, limit, false)
	return ctx, err
}
