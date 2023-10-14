package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
)

type FoldHandAllCommand struct {
	*VirtualClassroomCommand
}

type FoldHandAllCommandHandler struct {
	command          *FoldHandAllCommand
	ctx              context.Context
	db               database.Ext
	lessonMemberRepo infrastructure.LessonMemberRepo
}

func (h *FoldHandAllCommandHandler) Execute() error {
	if err := h.lessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		h.ctx,
		h.db,
		h.command.LessonID,
		domain.LearnerStateTypeHandsUp,
		&repo.StateValueDTO{
			BoolValue:        database.Bool(false),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}
