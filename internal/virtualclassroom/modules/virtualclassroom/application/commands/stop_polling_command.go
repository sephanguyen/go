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

type StopPollingCommand struct {
	*VirtualClassroomCommand
}

type StopPollingCommandHandler struct {
	command             *StopPollingCommand
	ctx                 context.Context
	db                  database.Ext
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (h *StopPollingCommandHandler) pExecute(db database.Ext) error {
	state, err := h.lessonRoomStateRepo.GetLessonRoomStateByLessonID(h.ctx, db, h.command.LessonID)
	if err != nil && err != domain.ErrLessonRoomStateNotFound {
		return fmt.Errorf("LessonRoomStateRepo.GetLessonRoomStateByLessonID: %w with lessonId %s", err, h.command.LessonID)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("the Polling not exists")
	}
	if state.CurrentPolling.Status != domain.CurrentPollingStatusStarted {
		return fmt.Errorf("permission denied: Can't stop polling when polling not start")
	}
	// update room lessonState
	now := time.Now()
	state.CurrentPolling.StoppedAt = &now
	state.CurrentPolling.Status = domain.CurrentPollingStatusStopped

	if err := h.lessonRoomStateRepo.UpsertCurrentPollingState(h.ctx, db, h.command.LessonID, state.CurrentPolling); err != nil {
		return fmt.Errorf("lessonRoomStateRepo.UpsertCurrentPollingState: %w", err)
	}
	return nil
}

func (h *StopPollingCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
