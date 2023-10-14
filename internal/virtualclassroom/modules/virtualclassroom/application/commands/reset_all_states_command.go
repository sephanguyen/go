package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type ResetAllStatesCommand struct {
	*VirtualClassroomCommand
}

type ResetAllStatesCommandHandler struct {
	command    *ResetAllStatesCommand
	ctx        context.Context
	db         database.Ext
	dispatcher Dispatcher
}

func (r *ResetAllStatesCommandHandler) pExecute(db database.Ext) error {
	if err := r.dispatcher.DispatchWithTransaction(db, &StopSharingMaterialCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
	}); err != nil {
		return fmt.Errorf("StopSharingMaterialCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &DisableAllAnnotationCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
	}); err != nil {
		return fmt.Errorf("DisableAllAnnotationCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &FoldHandAllCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
	}); err != nil {
		return fmt.Errorf("FoldHandAllCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &ResetPollingCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
	}); err != nil {
		return fmt.Errorf("ResetPollingCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &WhiteboardZoomStateCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
		WhiteboardZoomState: new(domain.WhiteboardZoomState).SetDefault(),
	}); err != nil {
		return fmt.Errorf("WhiteboardZoomStateCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &SpotlightCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
		IsEnable: false,
	}); err != nil {
		return fmt.Errorf("SpotlightCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &ResetAllChatCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
	}); err != nil {
		return fmt.Errorf("ResetChatCommand: %w", err)
	}

	if err := r.dispatcher.DispatchWithTransaction(db, &ClearRecordingCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: r.command.CommanderID,
			LessonID:    r.command.LessonID},
	}); err != nil {
		return fmt.Errorf("ClearRecordingCommand: %w", err)
	}

	return nil
}

func (r *ResetAllStatesCommandHandler) Execute() error {
	switch r.db.(type) {
	case pgx.Tx:
		return r.pExecute(r.db)
	default:
		return database.ExecInTx(r.ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
			return r.pExecute(tx)
		})
	}
}
