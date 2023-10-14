package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type WhiteboardZoomStateCommand struct {
	*VirtualClassroomCommand
	WhiteboardZoomState *domain.WhiteboardZoomState
}

type WhiteboardZoomStateCommandHandler struct {
	command             *WhiteboardZoomStateCommand
	ctx                 context.Context
	db                  database.Ext
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (w *WhiteboardZoomStateCommandHandler) pExecute(db database.Ext) error {
	if err := w.lessonRoomStateRepo.UpsertWhiteboardZoomState(w.ctx, db, w.command.LessonID, w.command.WhiteboardZoomState); err != nil {
		return fmt.Errorf("lessonRoomStateRepo.UpsertWhiteboardZoomState: %w", err)
	}

	return nil
}

func (w *WhiteboardZoomStateCommandHandler) Execute() error {
	switch w.db.(type) {
	case pgx.Tx:
		return w.pExecute(w.db)
	default:
		return database.ExecInTx(w.ctx, w.db, func(ctx context.Context, tx pgx.Tx) error {
			return w.pExecute(tx)
		})
	}
}
