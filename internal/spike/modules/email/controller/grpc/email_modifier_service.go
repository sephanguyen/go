package grpc

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/commands"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure/repositories"
	"github.com/manabie-com/backend/internal/spike/modules/email/metrics"
)

type EmailModifierService struct {
	KafkaMgmt    kafka.KafkaManagement
	Env          string
	DB           database.Ext
	EmailMetrics metrics.EmailMetrics

	EmailCommandHandler interface {
		CreateEmail(ctx context.Context, payload *commands.CreateEmailPayload) (*dto.Email, error)
	}

	EmailRepo infrastructure.EmailRepo
}

func NewEmailModifierService(db database.Ext, kafkaMgmt kafka.KafkaManagement, metrics metrics.EmailMetrics, env string) *EmailModifierService {
	return &EmailModifierService{
		KafkaMgmt:    kafkaMgmt,
		Env:          env,
		DB:           db,
		EmailMetrics: metrics,

		EmailRepo: &repositories.EmailRepo{},

		EmailCommandHandler: &commands.CreateEmailHandler{
			DB:                 db,
			EmailRepo:          &repositories.EmailRepo{},
			EmailRecipientRepo: &repositories.EmailRecipientRepo{},
		},
	}
}
