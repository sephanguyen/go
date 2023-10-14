package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type UpdateAnnotationCommand struct {
	*ModifyLiveRoomCommand
	UserIDs []string
	State   *vc_domain.UserAnnotation
}

type UpdateAnnotationCommandHandler struct {
	command      *UpdateAnnotationCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (u *UpdateAnnotationCommandHandler) pExecute(db database.Ext) error {
	if len(u.command.UserIDs) == 0 {
		return fmt.Errorf("learners are found empty, cannot update annotation")
	}

	if err := u.LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState(
		u.ctx,
		db,
		u.command.ChannelID,
		u.command.UserIDs,
		vc_domain.LearnerStateTypeAnnotation,
		&vc_domain.StateValue{
			BoolValue:        u.command.State.Value,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState, channel %s, users %v, state %s: %w",
			u.command.ChannelID,
			u.command.UserIDs,
			vc_domain.LearnerStateTypeChat,
			err,
		)
	}

	return nil
}

func (u *UpdateAnnotationCommandHandler) Execute() error {
	switch u.lessonmgmtDB.(type) {
	case pgx.Tx:
		return u.pExecute(u.lessonmgmtDB)
	default:
		return database.ExecInTx(u.ctx, u.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return u.pExecute(tx)
		})
	}
}
