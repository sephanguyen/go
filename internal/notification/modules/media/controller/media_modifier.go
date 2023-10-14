package controller

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/media/application/commands"
	"github.com/manabie-com/backend/internal/notification/modules/media/infrastructure"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

type MediaModifierService struct {
	npb.MediaModifierServiceServer

	UpsertMediaCommandHandler commands.UpsertMediaCommandHandler
}

func NewMediaModifierService(db database.Ext, mediaRepo infrastructure.MediaRepo) *MediaModifierService {
	return &MediaModifierService{
		UpsertMediaCommandHandler: commands.UpsertMediaCommandHandler{
			DB:        db,
			MediaRepo: mediaRepo,
		},
	}
}
