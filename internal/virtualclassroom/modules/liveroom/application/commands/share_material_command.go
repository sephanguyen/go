package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type ShareMaterialCommand struct {
	*ModifyLiveRoomCommand
	State *vc_domain.CurrentMaterial
}

type ShareMaterialCommandHandler struct {
	command      *ShareMaterialCommand
	ctx          context.Context
	lessonmgmtDB database.Ext
	dispatcher   Dispatcher

	LiveRoomState infrastructure.LiveRoomStateRepo
}

func (s *ShareMaterialCommandHandler) pExecute(db database.Ext) error {
	channelID := s.command.ChannelID
	newCurrentMaterial := s.command.State
	if newCurrentMaterial != nil {
		newCurrentMaterial.UpdatedAt = time.Now()

		if err := newCurrentMaterial.IsValid(); err != nil {
			return fmt.Errorf("invalid current material state: %w", err)
		}
	}
	if err := s.LiveRoomState.UpsertLiveRoomCurrentMaterialState(s.ctx, db, channelID, newCurrentMaterial); err != nil {
		return fmt.Errorf("error in LiveRoomState.UpsertLiveRoomCurrentMaterialState, channel %s: %w", channelID, err)
	}

	return nil
}

func (s *ShareMaterialCommandHandler) Execute() error {
	switch s.lessonmgmtDB.(type) {
	case pgx.Tx:
		return s.pExecute(s.lessonmgmtDB)
	default:
		return database.ExecInTx(s.ctx, s.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
