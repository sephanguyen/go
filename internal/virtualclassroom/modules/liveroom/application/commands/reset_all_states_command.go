package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type ResetAllStatesCommand struct {
	*ModifyLiveRoomCommand
}

type ResetAllStatesCommandHandler struct {
	command      *ResetAllStatesCommand
	ctx          context.Context
	lessonmgmtDB database.Ext
	dispatcher   Dispatcher
}

func (r *ResetAllStatesCommandHandler) pExecute(db database.Ext) error {
	if err := r.dispatcher.DispatchWithTransaction(db, &StopSharingMaterialCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
	}); err != nil {
		return fmt.Errorf("StopSharingMaterialCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &EnableAllAnnotationCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
	}); err != nil {
		return fmt.Errorf("EnableAllAnnotationCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &FoldHandAllCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
	}); err != nil {
		return fmt.Errorf("FoldHandAllCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &ResetPollingCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
	}); err != nil {
		return fmt.Errorf("ResetPollingCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &WhiteboardZoomStateCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
		WhiteboardZoomState: new(domain.WhiteboardZoomState).SetDefault(),
	}); err != nil {
		return fmt.Errorf("WhiteboardZoomStateCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &SpotlightCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
		IsEnable: false,
	}); err != nil {
		return fmt.Errorf("SpotlightCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &ResetAllChatCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
	}); err != nil {
		return fmt.Errorf("ResetChatCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &ClearRecordingCommand{
		ModifyLiveRoomCommand: &ModifyLiveRoomCommand{
			CommanderID: r.command.CommanderID,
			ChannelID:   r.command.ChannelID},
	}); err != nil {
		return fmt.Errorf("ClearRecordingCommand: %w", err)
	}

	return nil
}

func (r *ResetAllStatesCommandHandler) Execute() error {
	switch r.lessonmgmtDB.(type) {
	case pgx.Tx:
		return r.pExecute(r.lessonmgmtDB)
	default:
		return database.ExecInTx(r.ctx, r.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return r.pExecute(tx)
		})
	}
}
