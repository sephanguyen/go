package commands

import "github.com/manabie-com/backend/internal/notification/modules/media/domain"

type UpsertMediaPayload struct {
	Medias domain.Medias
}
