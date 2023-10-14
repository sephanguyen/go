package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type WhiteboardZoomStateCommand struct {
	*ModifyLiveRoomCommand
	WhiteboardZoomState *vc_domain.WhiteboardZoomState
}

type WhiteboardZoomStateCommandHandler struct {
	command      *WhiteboardZoomStateCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo infrastructure.LiveRoomStateRepo
}

func (w *WhiteboardZoomStateCommandHandler) pExecute(db database.Ext) error {
	channelID := w.command.ChannelID

	if err := w.LiveRoomStateRepo.UpsertLiveRoomWhiteboardZoomState(w.ctx, db, channelID, w.command.WhiteboardZoomState); err != nil {
		return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomWhiteboardZoomState, channel %s: %w", channelID, err)
	}

	return nil
}

func (w *WhiteboardZoomStateCommandHandler) Execute() error {
	switch w.lessonmgmtDB.(type) {
	case pgx.Tx:
		return w.pExecute(w.lessonmgmtDB)
	default:
		return database.ExecInTx(w.ctx, w.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return w.pExecute(tx)
		})
	}
}
