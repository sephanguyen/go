package eureka

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/golibs/database"
	epb "github.com/manabie-com/backend/pkg/manabuf/eureka/v1"
)

func (s *suite) userChangeOrderWithTimesInQuizSet(ctx context.Context, numStr string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	var pairs []*epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair
	num, _ := strconv.Atoi(numStr)
	for i := 0; i < num; i++ {
		pairs = append(pairs, &epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair{
			First:  stepState.QuizSet.QuizExternalIDs.Elements[rand.Intn(len(stepState.QuizSet.QuizExternalIDs.Elements))].String,
			Second: stepState.QuizSet.QuizExternalIDs.Elements[rand.Intn(len(stepState.QuizSet.QuizExternalIDs.Elements))].String,
		})
	}

	stepState.Request = &epb.UpdateDisplayOrderOfQuizSetRequest{
		LoId:  stepState.LoID,
		Pairs: pairs,
	}
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpdateDisplayOrderOfQuizSet(ctx, stepState.Request.(*epb.UpdateDisplayOrderOfQuizSetRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userMoveOneQuizInQuizSet(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx = contextWithToken(s, ctx)

	var pairs []*epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair
	switch action {
	case "up":
		quizIdx := rand.Intn(len(stepState.QuizSet.QuizExternalIDs.Elements)-1) + 1
		pairs = append(pairs, &epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair{
			First:  stepState.QuizSet.QuizExternalIDs.Elements[quizIdx].String,
			Second: stepState.QuizSet.QuizExternalIDs.Elements[quizIdx-1].String,
		})
	case "down":
		quizIdx := rand.Intn(len(stepState.QuizSet.QuizExternalIDs.Elements) - 1)
		pairs = append(pairs, &epb.UpdateDisplayOrderOfQuizSetRequest_QuizExternalIDPair{
			First:  stepState.QuizSet.QuizExternalIDs.Elements[quizIdx].String,
			Second: stepState.QuizSet.QuizExternalIDs.Elements[quizIdx+1].String,
		})
	}

	stepState.Request = &epb.UpdateDisplayOrderOfQuizSetRequest{
		LoId:  stepState.LoID,
		Pairs: pairs,
	}
	stepState.Response, stepState.ResponseErr = epb.NewQuizModifierServiceClient(s.Conn).UpdateDisplayOrderOfQuizSet(ctx, stepState.Request.(*epb.UpdateDisplayOrderOfQuizSetRequest))

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updateTheOrderQuizzesInQuizSetAsExpected(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.returnsStatusCode(ctx, "OK")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req, ok := stepState.Request.(*epb.UpdateDisplayOrderOfQuizSetRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expect epb.UpdateDisplayOrderOfQuizSetRequest but got %T", stepState.Request)
	}

	mapIdx := make(map[string]int)
	for i, quizExternalID := range stepState.QuizSet.QuizExternalIDs.Elements {
		mapIdx[quizExternalID.String] = i
	}

	for _, pair := range req.Pairs {
		i := mapIdx[pair.First]
		j := mapIdx[pair.Second]
		stepState.QuizSet.QuizExternalIDs.Elements[i], stepState.QuizSet.QuizExternalIDs.Elements[j] = stepState.QuizSet.QuizExternalIDs.Elements[j], stepState.QuizSet.QuizExternalIDs.Elements[i]
		mapIdx[pair.First], mapIdx[pair.Second] = mapIdx[pair.Second], mapIdx[pair.First]
	}

	quizSetRepo := repositories.QuizSetRepo{}
	updatedQuizSet, err := quizSetRepo.GetQuizSetByLoID(ctx, s.DB, database.Text(stepState.LoID))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for i, quizExternalID := range updatedQuizSet.QuizExternalIDs.Elements {
		if quizExternalID.String != stepState.QuizSet.QuizExternalIDs.Elements[i].String {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect quiz external ids %v but got %v", stepState.QuizSet.QuizExternalIDs, updatedQuizSet.QuizExternalIDs)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
