package commands

import "github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"

type CreateEmailPayload struct {
	Email *dto.Email
}
