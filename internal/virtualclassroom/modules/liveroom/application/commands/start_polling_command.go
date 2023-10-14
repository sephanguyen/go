package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type StartPollingCommand struct {
	*ModifyLiveRoomCommand
	Options  vc_domain.CurrentPollingOptions
	Question string
}

type StartPollingCommandHandler struct {
	command      *StartPollingCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo infrastructure.LiveRoomStateRepo
}

func (s *StartPollingCommandHandler) pExecute(db database.Ext) error {
	pollOptions := s.command.Options
	if err := pollOptions.ValidatePollingOptions([]string{}); err != nil {
		return err
	}

	channelID := s.command.ChannelID
	state, err := s.LiveRoomStateRepo.GetLiveRoomStateByChannelID(s.ctx, db, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		return fmt.Errorf("error in LiveRoomStateRepo.GetLiveRoomStateByChannelID, channel %s: %w", channelID, err)
	}
	if state.CurrentPolling != nil {
		return fmt.Errorf("the polling already exists for live room %s", channelID)
	}

	now := time.Now()
	state.CurrentPolling = &vc_domain.CurrentPolling{
		Options:   s.command.Options,
		Status:    vc_domain.CurrentPollingStatusStarted,
		CreatedAt: now,
		UpdatedAt: now,
		IsShared:  false,
		Question:  s.command.Question,
	}
	if err := s.LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState(s.ctx, db, channelID, state.CurrentPolling); err != nil {
		return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState, channel %s: %w", channelID, err)
	}

	return nil
}

func (s *StartPollingCommandHandler) Execute() error {
	switch s.lessonmgmtDB.(type) {
	case pgx.Tx:
		return s.pExecute(s.lessonmgmtDB)
	default:
		return database.ExecInTx(s.ctx, s.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
