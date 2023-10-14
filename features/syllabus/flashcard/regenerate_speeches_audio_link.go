package flashcard

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/syllabus/utils"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/golibs/database"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
)

func (s *Suite) optionsAndSpeechesUpdatedCorrectly(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	repo := &repositories.SpeechesRepository{}
	speeches, err := repo.RetrieveAllSpeaches(ctx, s.EurekaDB, database.Int8(100), database.Int8(0))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	speechMap := make(map[string]string)
	for _, speech := range stepState.OldSpeeches {
		speechMap[speech.SpeechID.String] = speech.Link.String
	}

	for _, speech := range speeches {
		if link, ok := speechMap[speech.SpeechID.String]; !ok {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("speech id: %v not found", speech.SpeechID.String)
		} else if link == speech.Link.String {
			return utils.StepStateToContext(ctx, stepState), fmt.Errorf("link of speech id: %v not updated", speech.SpeechID.String)
		}
	}
	return utils.StepStateToContext(ctx, stepState), nil
}

func (s *Suite) regenerateSpeechesAudioLink(ctx context.Context) (context.Context, error) {
	stepState := utils.StepStateFromContext[StepState](ctx)
	speechesRepo := &repositories.SpeechesRepository{}
	speeches, err := speechesRepo.RetrieveAllSpeaches(ctx, s.EurekaDB, database.Int8(100), database.Int8(0))
	if err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}
	stepState.OldSpeeches = speeches
	bobMediaModifier := bpb.NewMediaModifierServiceClient(s.BobConn)
	service := &services.QuizService{
		DB:               s.EurekaDB,
		SpeechesRepo:     speechesRepo,
		BobMediaModifier: bobMediaModifier,
	}

	if err := service.RegenerateSpeechesAudioLink(ctx); err != nil {
		return utils.StepStateToContext(ctx, stepState), err
	}

	return utils.StepStateToContext(ctx, stepState), nil
}
