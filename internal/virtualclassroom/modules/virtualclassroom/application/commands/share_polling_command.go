package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type SharePollingCommand struct {
	*VirtualClassroomCommand
	IsShared bool
}

type SharePollingCommandHandler struct {
	command             *SharePollingCommand
	ctx                 context.Context
	db                  database.Ext
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (s *SharePollingCommandHandler) pExecute(db database.Ext) error {
	state, err := s.lessonRoomStateRepo.GetLessonRoomStateByLessonID(s.ctx, db, s.command.LessonID)
	if err != nil && err != domain.ErrLessonRoomStateNotFound {
		return fmt.Errorf("LessonRoomStateRepo.GetLessonRoomStateByLessonID: %w with lessonId %s", err, s.command.LessonID)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("lesson with ID %s does not have a polling state", s.command.LessonID)
	}
	if state.CurrentPolling.Status != domain.CurrentPollingStatusStopped {
		return fmt.Errorf("permission denied: Can't stop polling when polling not stopped")
	}
	state.CurrentPolling.IsShared = s.command.IsShared

	if err := s.lessonRoomStateRepo.UpsertCurrentPollingState(s.ctx, db, s.command.LessonID, state.CurrentPolling); err != nil {
		return fmt.Errorf("lessonRoomStateRepo.UpsertCurrentPollingState: %w", err)
	}
	return nil
}

func (s *SharePollingCommandHandler) Execute() error {
	switch s.db.(type) {
	case pgx.Tx:
		return s.pExecute(s.db)
	default:
		return database.ExecInTx(s.ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
