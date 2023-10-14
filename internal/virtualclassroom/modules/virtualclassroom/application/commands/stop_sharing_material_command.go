package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type StopSharingMaterialCommand struct {
	*VirtualClassroomCommand
}

type StopSharingMaterialCommandHandler struct {
	command    *StopSharingMaterialCommand
	ctx        context.Context
	db         database.Ext
	dispatcher Dispatcher
}

func (h *StopSharingMaterialCommandHandler) Execute() error {
	if err := h.dispatcher.Dispatch(&ShareMaterialCommand{
		VirtualClassroomCommand: &VirtualClassroomCommand{
			CommanderID: h.command.CommanderID,
			LessonID:    h.command.LessonID},
	}); err != nil {
		return err
	}

	return nil
}
