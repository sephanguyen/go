package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"github.com/jackc/pgx/v4"
)

type ResetAllChatCommand struct {
	*VirtualClassroomCommand
}

type ResetAllChatCommandHandler struct {
	command          *ResetAllChatCommand
	ctx              context.Context
	db               database.Ext
	lessonMemberRepo infrastructure.LessonMemberRepo
}

func (h *ResetAllChatCommandHandler) pExecute(db database.Ext) error {
	if err := h.lessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		h.ctx,
		db,
		h.command.LessonID,
		domain.LearnerStateTypeChat,
		&repo.StateValueDTO{
			BoolValue:        database.Bool(true), // default chat value true
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}

func (h *ResetAllChatCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
