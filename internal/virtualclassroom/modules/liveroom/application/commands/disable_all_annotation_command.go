package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type DisableAllAnnotationCommand struct {
	*ModifyLiveRoomCommand
}

type DisableAllAnnotationCommandHandler struct {
	command      *DisableAllAnnotationCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (d *DisableAllAnnotationCommandHandler) pExecute(db database.Ext) error {
	if err := d.LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState(
		d.ctx,
		db,
		d.command.ChannelID,
		vc_domain.LearnerStateTypeAnnotation,
		&vc_domain.StateValue{
			BoolValue:        false,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState, channel %s, state %s: %w",
			d.command.ChannelID,
			vc_domain.LearnerStateTypeAnnotation,
			err,
		)
	}

	return nil
}

func (d *DisableAllAnnotationCommandHandler) Execute() error {
	switch d.lessonmgmtDB.(type) {
	case pgx.Tx:
		return d.pExecute(d.lessonmgmtDB)
	default:
		return database.ExecInTx(d.ctx, d.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return d.pExecute(tx)
		})
	}
}
