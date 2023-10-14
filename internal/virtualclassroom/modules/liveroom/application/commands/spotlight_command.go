package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type SpotlightCommand struct {
	*ModifyLiveRoomCommand
	SpotlightedUser string
	IsEnable        bool
}

type SpotlightCommandHandler struct {
	command      *SpotlightCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo infrastructure.LiveRoomStateRepo
}

func (s *SpotlightCommandHandler) pExecute(db database.Ext) error {
	channelID := s.command.ChannelID

	if s.command.IsEnable {
		if len(s.command.SpotlightedUser) == 0 {
			return fmt.Errorf("spotlighted user cannot be empty when enabling spotlight state")
		}

		if err := s.LiveRoomStateRepo.UpsertLiveRoomSpotlightState(s.ctx, db, channelID, s.command.SpotlightedUser); err != nil {
			return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomSpotlightState, channel %s: %w", channelID, err)
		}
	} else {
		if err := s.LiveRoomStateRepo.UnSpotlight(s.ctx, db, channelID); err != nil {
			return fmt.Errorf("error in LiveRoomStateRepo.UnSpotlight, channel %s: %w", channelID, err)
		}
	}

	return nil
}

func (s *SpotlightCommandHandler) Execute() error {
	switch s.lessonmgmtDB.(type) {
	case pgx.Tx:
		return s.pExecute(s.lessonmgmtDB)
	default:
		return database.ExecInTx(s.ctx, s.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
