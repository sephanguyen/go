package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type UpsertSessionTimeCommand struct {
	*ModifyLiveRoomCommand
}

type UpsertSessionTimeCommandHandler struct {
	command      *UpsertSessionTimeCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo infrastructure.LiveRoomStateRepo
}

func (u *UpsertSessionTimeCommandHandler) pExecute(db database.Ext) error {
	if err := u.LiveRoomStateRepo.UpsertLiveRoomSessionTime(
		u.ctx,
		db,
		u.command.ChannelID,
		time.Now(),
	); err != nil {
		return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomSessionTime, channel %s: %w",
			u.command.ChannelID,
			err,
		)
	}

	return nil
}

func (u *UpsertSessionTimeCommandHandler) Execute() error {
	switch u.lessonmgmtDB.(type) {
	case pgx.Tx:
		return u.pExecute(u.lessonmgmtDB)
	default:
		return database.ExecInTx(u.ctx, u.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return u.pExecute(tx)
		})
	}
}
