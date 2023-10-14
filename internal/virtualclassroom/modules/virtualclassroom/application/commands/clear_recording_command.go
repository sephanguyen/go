package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ClearRecordingCommand struct {
	*VirtualClassroomCommand
}

type ClearRecordingCommandHandler struct {
	command             *ClearRecordingCommand
	ctx                 context.Context
	db                  database.Ext
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (c *ClearRecordingCommandHandler) pExecute(db database.Ext) error {
	if err := c.lessonRoomStateRepo.UpsertRecordingState(c.ctx, db, c.command.LessonID, nil); err != nil {
		return fmt.Errorf("lessonRoomStateRepo.UpsertRecordingState, lesson %s: %w", c.command.LessonID, err)
	}

	return nil
}

func (c *ClearRecordingCommandHandler) Execute() error {
	switch c.db.(type) {
	case pgx.Tx:
		return c.pExecute(c.db)
	default:
		return database.ExecInTx(c.ctx, c.db, func(ctx context.Context, tx pgx.Tx) error {
			return c.pExecute(tx)
		})
	}
}
