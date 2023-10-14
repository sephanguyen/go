package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"github.com/jackc/pgx/v4"
)

type UpdateAnnotationCommand struct {
	*VirtualClassroomCommand
	UserIDs []string // list user who will be changed annotation state
	State   *domain.UserAnnotation
}

type UpdateAnnotationCommandHandler struct {
	command          *UpdateAnnotationCommand
	ctx              context.Context
	db               database.Ext
	lessonMemberRepo infrastructure.LessonMemberRepo
}

func (h *UpdateAnnotationCommandHandler) pExecute(db database.Ext) error {
	err := h.lessonMemberRepo.UpsertMultiLessonMemberStateByState(
		h.ctx,
		db,
		h.command.LessonID,
		domain.LearnerStateTypeAnnotation,
		h.command.UserIDs,
		&repo.StateValueDTO{
			BoolValue:        database.Bool(h.command.State.Value),
			StringArrayValue: database.TextArray([]string{}),
		},
	)
	return err
}

func (h *UpdateAnnotationCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
