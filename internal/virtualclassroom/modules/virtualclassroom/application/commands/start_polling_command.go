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

type StartPollingCommand struct {
	*VirtualClassroomCommand
	Options  domain.CurrentPollingOptions
	Question string
}

type StartPollingCommandHandler struct {
	command             *StartPollingCommand
	ctx                 context.Context
	db                  database.Ext
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (h *StartPollingCommandHandler) pExecute(db database.Ext) error {
	options := h.command.Options
	if err := options.ValidatePollingOptions([]string{}); err != nil {
		return err
	}
	state, err := h.lessonRoomStateRepo.GetLessonRoomStateByLessonID(h.ctx, db, h.command.LessonID)
	if err != nil && err != domain.ErrLessonRoomStateNotFound {
		return fmt.Errorf("LessonRoomStateRepo.GetLessonRoomStateByLessonID: %w with lessonId %s", err, h.command.LessonID)
	}
	if state.CurrentPolling != nil {
		return fmt.Errorf("the Polling already exists")
	}
	state.CurrentPolling = &domain.CurrentPolling{
		Options:   h.command.Options,
		Status:    domain.CurrentPollingStatusStarted,
		CreatedAt: time.Now(),
		IsShared:  false,
		Question:  h.command.Question,
	}
	if err := h.lessonRoomStateRepo.UpsertCurrentPollingState(h.ctx, db, h.command.LessonID, state.CurrentPolling); err != nil {
		return fmt.Errorf("VirtualLessonRepo.UpsertCurrentPollingState: %w", err)
	}

	return nil
}

func (h *StartPollingCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
