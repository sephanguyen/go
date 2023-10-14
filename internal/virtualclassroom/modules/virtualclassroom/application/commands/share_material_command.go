package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type ShareMaterialCommand struct {
	*VirtualClassroomCommand
	State *domain.CurrentMaterial
}

type ShareMaterialCommandHandler struct {
	command             *ShareMaterialCommand
	ctx                 context.Context
	db                  database.Ext
	virtualLessonRepo   infrastructure.VirtualLessonRepo
	lessonGroupRepo     infrastructure.LessonGroupRepo
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
	dispatcher          Dispatcher
}

func (h *ShareMaterialCommandHandler) pExecute(db database.Ext) error {
	lesson, err := h.virtualLessonRepo.GetVirtualLessonByID(h.ctx, db, h.command.LessonID)
	if err != nil {
		return fmt.Errorf("error in VirtualLessonRepo.GetVirtualLessonByID, lesson %s: %w", h.command.LessonID, err)
	}
	// check media belong to lesson group of lesson or not
	if h.command.State != nil {
		gr, err := h.lessonGroupRepo.GetByIDAndCourseID(h.ctx, db, lesson.LessonGroupID, lesson.CourseID)

		if err != nil {
			return fmt.Errorf("error in LessonGroupRepo.GetByIDAndCourseID, lesson %s: %w", h.command.LessonID, err)
		}
		isValid := false
		for _, media := range gr.MediaIDs.Elements {
			if media.String == h.command.State.MediaID {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("media %s not belong to lesson %s", h.command.State.MediaID, h.command.LessonID)
		}
	}

	newCurrentMaterial := h.command.State
	if newCurrentMaterial != nil {
		newCurrentMaterial.UpdatedAt = time.Now()

		if err := newCurrentMaterial.IsValid(); err != nil {
			return fmt.Errorf("invalid current material state: %w", err)
		}
	}
	if err := h.lessonRoomStateRepo.UpsertCurrentMaterialState(h.ctx, db, lesson.LessonID, newCurrentMaterial); err != nil {
		return fmt.Errorf("error in lessonRoomStateRepo.UpsertCurrentMaterialState, lesson %s: %w", lesson.LessonID, err)
	}

	return nil
}

func (h *ShareMaterialCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
