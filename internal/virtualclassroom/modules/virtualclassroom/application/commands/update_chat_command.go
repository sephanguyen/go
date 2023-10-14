package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"github.com/jackc/pgx/v4"
)

type UpdateChatCommand struct {
	*VirtualClassroomCommand
	UserIDs []string // list user whose chat permission will be changed
	State   *domain.UserChat
}

type UpdateChatCommandHandler struct {
	command          *UpdateChatCommand
	ctx              context.Context
	db               database.Ext
	lessonMemberRepo infrastructure.LessonMemberRepo
}

func (h *UpdateChatCommandHandler) pExecute(db database.Ext) error {
	err := h.lessonMemberRepo.UpsertMultiLessonMemberStateByState(
		h.ctx,
		db,
		h.command.LessonID,
		domain.LearnerStateTypeChat,
		h.command.UserIDs,
		&repo.StateValueDTO{
			BoolValue:        database.Bool(h.command.State.Value),
			StringArrayValue: database.TextArray([]string{}),
		},
	)
	return err
}

func (h *UpdateChatCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
