package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type FoldHandAllCommand struct {
	*ModifyLiveRoomCommand
}

type FoldHandAllCommandHandler struct {
	command      *FoldHandAllCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (f *FoldHandAllCommandHandler) pExecute(db database.Ext) error {
	if err := f.LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState(
		f.ctx,
		db,
		f.command.ChannelID,
		vc_domain.LearnerStateTypeHandsUp,
		&vc_domain.StateValue{
			BoolValue:        false,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState, channel %s, state %s: %w",
			f.command.ChannelID,
			vc_domain.LearnerStateTypeHandsUp,
			err,
		)
	}

	return nil
}

func (f *FoldHandAllCommandHandler) Execute() error {
	switch f.lessonmgmtDB.(type) {
	case pgx.Tx:
		return f.pExecute(f.lessonmgmtDB)
	default:
		return database.ExecInTx(f.ctx, f.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return f.pExecute(tx)
		})
	}
}
