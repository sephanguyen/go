package commands

import "github.com/manabie-com/backend/internal/spike/modules/email/domain/dto"

type UpsertEmailEventPayload struct {
	EmailEvents   []dto.SGEmailEvent
	AllowedOrgIDs []string
}
