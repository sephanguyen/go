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

type DisableAllAnnotationCommand struct {
	*VirtualClassroomCommand
}

type DisableAllAnnotationCommandHandler struct {
	command          *DisableAllAnnotationCommand
	ctx              context.Context
	db               database.Ext
	lessonMemberRepo infrastructure.LessonMemberRepo
}

func (d *DisableAllAnnotationCommandHandler) pExecute(db database.Ext) error {
	if err := d.lessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		d.ctx,
		db,
		d.command.LessonID,
		domain.LearnerStateTypeAnnotation,
		&repo.StateValueDTO{
			BoolValue:        database.Bool(false),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}

func (d *DisableAllAnnotationCommandHandler) Execute() error {
	switch d.db.(type) {
	case pgx.Tx:
		return d.pExecute(d.db)
	default:
		return database.ExecInTx(d.ctx, d.db, func(ctx context.Context, tx pgx.Tx) error {
			return d.pExecute(tx)
		})
	}
}
