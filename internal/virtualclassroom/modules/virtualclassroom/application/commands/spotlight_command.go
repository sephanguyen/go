package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"github.com/jackc/pgx/v4"
)

type SpotlightCommand struct {
	*VirtualClassroomCommand
	SpotlightedUser string
	IsEnable        bool
}

type SpotlightCommandHandler struct {
	command             *SpotlightCommand
	ctx                 context.Context
	db                  database.Ext
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (s *SpotlightCommandHandler) pExecute(db database.Ext) error {
	if s.command.IsEnable {
		if err := s.lessonRoomStateRepo.UpsertSpotlightState(s.ctx, db, s.command.LessonID, s.command.SpotlightedUser); err != nil {
			return fmt.Errorf("lessonRoomStateRepo.UpsertSpotlightState, lesson %s: %w", s.command.LessonID, err)
		}
	} else {
		if err := s.lessonRoomStateRepo.UnSpotlight(s.ctx, db, s.command.LessonID); err != nil {
			return fmt.Errorf("lessonRoomStateRepo.UnSpotlight, lesson %s: %w", s.command.LessonID, err)
		}
	}

	return nil
}

func (s *SpotlightCommandHandler) Execute() error {
	switch s.db.(type) {
	case pgx.Tx:
		return s.pExecute(s.db)
	default:
		return database.ExecInTx(s.ctx, s.db, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
