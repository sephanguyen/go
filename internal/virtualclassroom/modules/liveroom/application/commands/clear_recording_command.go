package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ClearRecordingCommand struct {
	*ModifyLiveRoomCommand
}

type ClearRecordingCommandHandler struct {
	command      *ClearRecordingCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo infrastructure.LiveRoomStateRepo
}

func (c *ClearRecordingCommandHandler) pExecute(db database.Ext) error {
	channelID := c.command.ChannelID

	if err := c.LiveRoomStateRepo.UpsertRecordingState(c.ctx, db, channelID, nil); err != nil {
		return fmt.Errorf("LiveRoomStateRepo.UpsertRecordingState, channel %s: %w", channelID, err)
	}

	return nil
}

func (c *ClearRecordingCommandHandler) Execute() error {
	switch c.lessonmgmtDB.(type) {
	case pgx.Tx:
		return c.pExecute(c.lessonmgmtDB)
	default:
		return database.ExecInTx(c.ctx, c.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return c.pExecute(tx)
		})
	}
}
