package eureka

import (
	"context"
	"strconv"

	"github.com/manabie-com/backend/internal/eureka/entities"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) lODoesNotContainDeletedQuiz(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	quizSetRepo := &repositories.QuizSetRepo{}
	quizSets, err := quizSetRepo.GetQuizSetsOfLOContainQuiz(ctx, s.DB, database.Text(stepState.LoID), stepState.DeletedQuiz.ExternalID)
	if err != pgx.ErrNoRows {
		for _, quizSet := range quizSets {
			questionHierarchy := entities.QuestionHierarchy{}
			questionHierarchy.UnmarshalJSONBArray(quizSet.QuestionHierarchy)

			for _, questionHierarchyObj := range questionHierarchy {
				if questionHierarchyObj.ID == stepState.DeletedQuiz.ExternalID.String {
					return ctx, errors.New("error still containing external id in question hierarchy")
				}

				if sliceutils.Contains(questionHierarchyObj.ChildrenIDs, stepState.DeletedQuiz.ExternalID.String) {
					return ctx, errors.New("error still containing external id in question hierarchy")
				}
			}
		}
		return ctx, err
	}
	return ctx, errors.New("expect no rows but find quiz set in LO which contains deleted quiz")
}

func (s *suite) userRemoveAQuizFromLo(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	idx, _ := strconv.Atoi(arg1)
	stepState.DeletedQuiz = stepState.Quizzes[idx]
	quizID := stepState.DeletedQuiz.ExternalID.String

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).RemoveQuizFromLO(s.signedCtx(ctx), &epb.RemoveQuizFromLORequest{
		LoId:   stepState.LoID,
		QuizId: quizID,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRemoveAQuizWithoutLoId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	idx := 0
	stepState.DeletedQuiz = stepState.Quizzes[idx]
	quizID := stepState.DeletedQuiz.ExternalID.String

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).RemoveQuizFromLO(s.signedCtx(ctx), &epb.RemoveQuizFromLORequest{
		QuizId: quizID,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRemoveAQuizWithoutQuizId(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).RemoveQuizFromLO(s.signedCtx(ctx), &epb.RemoveQuizFromLORequest{
		LoId: stepState.LoID,
	})
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userRemoveAQuizWithout(ctx context.Context, arg1 string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	idx := 0
	stepState.DeletedQuiz = stepState.Quizzes[idx]
	quizID := stepState.DeletedQuiz.ExternalID.String

	var req *epb.RemoveQuizFromLORequest
	switch arg1 {
	case "lo id":
		req = &epb.RemoveQuizFromLORequest{
			QuizId: quizID,
		}
	case "quiz id":
		req = &epb.RemoveQuizFromLORequest{
			LoId: stepState.LoID,
		}
	case "none":
		req = &epb.RemoveQuizFromLORequest{
			QuizId: quizID,
			LoId:   stepState.LoID,
		}
	}

	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).RemoveQuizFromLO(s.signedCtx(ctx), req)
	return StepStateToContext(ctx, stepState), nil
}
