package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type SharePollingCommand struct {
	*ModifyLiveRoomCommand
	IsShared bool
}

type SharePollingCommandHandler struct {
	command      *SharePollingCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo infrastructure.LiveRoomStateRepo
}

func (s *SharePollingCommandHandler) pExecute(db database.Ext) error {
	channelID := s.command.ChannelID
	state, err := s.LiveRoomStateRepo.GetLiveRoomStateByChannelID(s.ctx, db, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		return fmt.Errorf("error in LiveRoomStateRepo.GetLiveRoomStateByChannelID, channel %s: %w", channelID, err)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("the polling does not exist in live room %s", channelID)
	}
	if state.CurrentPolling.Status != vc_domain.CurrentPollingStatusStopped {
		return fmt.Errorf("cannot share polling in live room %s as polling is not in stopped status", channelID)
	}

	state.CurrentPolling.IsShared = s.command.IsShared
	if err := s.LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState(s.ctx, db, channelID, state.CurrentPolling); err != nil {
		return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState, channel %s: %w", channelID, err)
	}
	return nil
}

func (s *SharePollingCommandHandler) Execute() error {
	switch s.lessonmgmtDB.(type) {
	case pgx.Tx:
		return s.pExecute(s.lessonmgmtDB)
	default:
		return database.ExecInTx(s.ctx, s.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
