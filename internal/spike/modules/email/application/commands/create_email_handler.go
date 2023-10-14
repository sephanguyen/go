package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure"
	"github.com/manabie-com/backend/internal/spike/modules/email/util/mapper"
	spb "github.com/manabie-com/backend/pkg/manabuf/spike/v1"

	"github.com/jackc/pgx/v4"
)

type CreateEmailHandler struct {
	DB database.Ext

	EmailRepo          infrastructure.EmailRepo
	EmailRecipientRepo infrastructure.EmailRecipientRepo
}

func (cmd *CreateEmailHandler) CreateEmail(ctx context.Context, payload *CreateEmailPayload) (*dto.Email, error) {
	payload.Email.EmailID = idutil.ULIDNow()
	payload.Email.EmailRecipients = golibs.GetUniqueElementStringArray(payload.Email.EmailRecipients)
	payload.Email.Status = spb.EmailStatus_EMAIL_STATUS_QUEUED.String()

	emailEnt, err := mapper.ToEmailEntity(payload.Email)
	if err != nil {
		return nil, fmt.Errorf("mapper.ToEmailEntity: %v", err)
	}

	emailRecipientEnts, err := mapper.ToEmailRecipientEntities(payload.Email)
	if err != nil {
		return nil, fmt.Errorf("mapper.ToEmailRecipientEntities: %v", err)
	}

	err = database.ExecInTx(ctx, cmd.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := cmd.EmailRepo.UpsertEmail(ctx, tx, emailEnt)
		if err != nil {
			return fmt.Errorf("cmd.EmailRepo.UpsertEmail: %v", err)
		}

		err = cmd.EmailRecipientRepo.BulkUpsertEmailRecipients(ctx, tx, emailRecipientEnts)
		if err != nil {
			return fmt.Errorf("cmd.EmailRecipientRepo.BulkUpsertEmailRecipients: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return payload.Email, nil
}
