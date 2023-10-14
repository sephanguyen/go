package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type UpsertSessionTimeCommand struct {
	*VirtualClassroomCommand
}

type UpsertSessionTimeCommandHandler struct {
	command *UpsertSessionTimeCommand
	ctx     context.Context
	db      database.Ext

	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (u *UpsertSessionTimeCommandHandler) pExecute(db database.Ext) error {
	if err := u.lessonRoomStateRepo.UpsertLiveLessonSessionTime(
		u.ctx,
		db,
		u.command.LessonID,
		time.Now(),
	); err != nil {
		return fmt.Errorf("error in LessonRoomStateRepo.UpsertLiveLessonSessionTime, lesson %s: %w",
			u.command.LessonID,
			err,
		)
	}

	return nil
}

func (u *UpsertSessionTimeCommandHandler) Execute() error {
	switch u.db.(type) {
	case pgx.Tx:
		return u.pExecute(u.db)
	default:
		return database.ExecInTx(u.ctx, u.db, func(ctx context.Context, tx pgx.Tx) error {
			return u.pExecute(tx)
		})
	}
}
