package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type EnableAllAnnotationCommand struct {
	*ModifyLiveRoomCommand
}

type EnableAllAnnotationCommandHandler struct {
	command      *EnableAllAnnotationCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (e *EnableAllAnnotationCommandHandler) pExecute(db database.Ext) error {
	if err := e.LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState(
		e.ctx,
		db,
		e.command.ChannelID,
		vc_domain.LearnerStateTypeAnnotation,
		&vc_domain.StateValue{
			BoolValue:        true,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.UpdateAllLiveRoomMembersState, channel %s, state %s: %w",
			e.command.ChannelID,
			vc_domain.LearnerStateTypeAnnotation,
			err,
		)
	}

	return nil
}

func (e *EnableAllAnnotationCommandHandler) Execute() error {
	switch e.lessonmgmtDB.(type) {
	case pgx.Tx:
		return e.pExecute(e.lessonmgmtDB)
	default:
		return database.ExecInTx(e.ctx, e.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return e.pExecute(tx)
		})
	}
}
