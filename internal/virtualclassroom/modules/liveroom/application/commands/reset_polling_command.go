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

type ResetPollingCommand struct {
	*ModifyLiveRoomCommand
}

type ResetPollingCommandHandler struct {
	command      *ResetPollingCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (r *ResetPollingCommandHandler) pExecute(db database.Ext) error {
	channelID := r.command.ChannelID
	state, err := r.LiveRoomStateRepo.GetLiveRoomStateByChannelID(r.ctx, db, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		return fmt.Errorf("error in LiveRoomStateRepo.GetLiveRoomStateByChannelID, channel %s: %w", channelID, err)
	}

	if state.CurrentPolling != nil {
		state.CurrentPolling = nil
		if err := r.LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState(r.ctx, db, channelID, state.CurrentPolling); err != nil {
			return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState, channel %s: %w", channelID, err)
		}

		if err := r.LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState(
			r.ctx,
			db,
			channelID,
			vc_domain.LearnerStateTypePollingAnswer,
			&vc_domain.StateValue{
				BoolValue:        false,
				StringArrayValue: []string{},
			},
		); err != nil {
			return fmt.Errorf("error in LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState, channel %s, state %s: %w",
				channelID,
				vc_domain.LearnerStateTypePollingAnswer,
				err,
			)
		}
	}

	return nil
}

func (r *ResetPollingCommandHandler) Execute() error {
	switch r.lessonmgmtDB.(type) {
	case pgx.Tx:
		return r.pExecute(r.lessonmgmtDB)
	default:
		return database.ExecInTx(r.ctx, r.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return r.pExecute(tx)
		})
	}
}
