package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type ResetAllChatCommand struct {
	*ModifyLiveRoomCommand
}

type ResetAllChatCommandHandler struct {
	command      *ResetAllChatCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (r *ResetAllChatCommandHandler) pExecute(db database.Ext) error {
	if err := r.LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState(
		r.ctx,
		db,
		r.command.ChannelID,
		vc_domain.LearnerStateTypeChat,
		&vc_domain.StateValue{
			BoolValue:        true,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState, channel %s, state %s: %w",
			r.command.ChannelID,
			vc_domain.LearnerStateTypeChat,
			err,
		)
	}

	return nil
}

func (r *ResetAllChatCommandHandler) Execute() error {
	switch r.lessonmgmtDB.(type) {
	case pgx.Tx:
		return r.pExecute(r.lessonmgmtDB)
	default:
		return database.ExecInTx(r.ctx, r.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return r.pExecute(tx)
		})
	}
}
