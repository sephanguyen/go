package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type UpdateHandsUpCommand struct {
	*ModifyLiveRoomCommand
	UserID string // user who will be changed hands up state
	State  *vc_domain.UserHandsUp
}

type UpdateHandsUpCommandHandler struct {
	command      *UpdateHandsUpCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (u *UpdateHandsUpCommandHandler) pExecute(db database.Ext) error {
	if len(u.command.UserID) == 0 {
		return fmt.Errorf("user ID found empty, cannot update hands up")
	}

	userIDs := []string{u.command.UserID}
	if err := u.LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState(
		u.ctx,
		db,
		u.command.ChannelID,
		userIDs,
		vc_domain.LearnerStateTypeHandsUp,
		&vc_domain.StateValue{
			BoolValue:        u.command.State.Value,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState, channel %s, users %v, state %s: %w",
			u.command.ChannelID,
			userIDs,
			vc_domain.LearnerStateTypeHandsUp,
			err,
		)
	}

	return nil
}

func (u *UpdateHandsUpCommandHandler) Execute() error {
	switch u.lessonmgmtDB.(type) {
	case pgx.Tx:
		return u.pExecute(u.lessonmgmtDB)
	default:
		return database.ExecInTx(u.ctx, u.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return u.pExecute(tx)
		})
	}
}
