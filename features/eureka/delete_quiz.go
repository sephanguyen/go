package eureka

import (
	"context"
	"strconv"

	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/golibs/constants"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) thereIsNoQuizsetThatContainsDeletedQuiz(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	quizSetRepo := &repositories.QuizSetRepo{}
	_, err := quizSetRepo.GetQuizSetsContainQuiz(ctx, s.DB, stepState.DeletedQuiz.ExternalID)
	if err != pgx.ErrNoRows {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), errors.New("expect no rows but find quiz set which contains deleted quiz")
}

func (s *suite) userDeleteAQuiz(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	idx, _ := strconv.Atoi(arg1)
	stepState.DeletedQuiz = stepState.Quizzes[idx]

	quizID := stepState.DeletedQuiz.ExternalID.String
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).DeleteQuiz(s.signedCtx(ctx), &epb.DeleteQuizRequest{
		QuizId:   quizID,
		SchoolId: constants.ManabieSchool,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteAQuizWithoutQuizId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).DeleteQuiz(s.signedCtx(ctx), &epb.DeleteQuizRequest{})

	return StepStateToContext(ctx, stepState), nil
}
