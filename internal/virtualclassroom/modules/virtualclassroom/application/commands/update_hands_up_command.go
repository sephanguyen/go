package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"go.uber.org/multierr"
)

type UpdateHandsUpCommand struct {
	*VirtualClassroomCommand
	UserID string // user who will be changed hands up state
	State  *domain.UserHandsUp
}

type UpdateHandsUpCommandHandler struct {
	command          *UpdateHandsUpCommand
	db               database.Ext
	lessonMemberRepo infrastructure.LessonMemberRepo
	ctx              context.Context
}

func (h *UpdateHandsUpCommandHandler) Execute() error {
	state := &repo.LessonMemberStateDTO{}
	database.AllNullEntity(state)

	now := time.Now()
	if err := multierr.Combine(
		state.LessonID.Set(h.command.LessonID),
		state.UserID.Set(h.command.UserID),
		state.StateType.Set(domain.LearnerStateTypeHandsUp),
		state.CreatedAt.Set(now),
		state.UpdatedAt.Set(now),
		state.BoolValue.Set(h.command.State.Value),
	); err != nil {
		return err
	}

	if err := h.lessonMemberRepo.UpsertLessonMemberState(
		h.ctx,
		h.db,
		state,
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertLessonMemberState: %w", err)
	}

	return nil
}
