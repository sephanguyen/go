package eureka

import (
	"context"

	"github.com/manabie-com/backend/internal/eureka/configurations"
	"github.com/manabie-com/backend/internal/eureka/repositories"
	"github.com/manabie-com/backend/internal/eureka/services"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/zap"
)

// only for Eishinkai
func RunRegenerateSpeechesAudioLink(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	logger := rsc.Logger().Sugar()
	db := rsc.DB()
	bobConn := rsc.GRPCDial("bob")
	bobMediaModifierClient := bpb.NewMediaModifierServiceClient(bobConn)
	quizService := &services.QuizService{
		SpeechesRepo:     &repositories.SpeechesRepository{},
		DB:               db,
		BobMediaModifier: bobMediaModifierClient,
	}

	organizationID := "-2147483631"
	ctx = auth.InjectFakeJwtToken(ctx, organizationID)

	if err := quizService.RegenerateSpeechesAudioLink(ctx); err != nil {
		logger.Error("RegenerateSpeechesAudioLink", zap.Error(err))
		return err
	}
	return nil
}
