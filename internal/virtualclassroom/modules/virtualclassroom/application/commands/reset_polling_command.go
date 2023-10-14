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

type ResetPollingCommand struct {
	*VirtualClassroomCommand
}

type ResetPollingCommandHandler struct {
	command             *ResetPollingCommand
	ctx                 context.Context
	db                  database.Ext
	lessonMemberRepo    infrastructure.LessonMemberRepo
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (r *ResetPollingCommandHandler) pExecute(db database.Ext) error {
	state, err := r.lessonRoomStateRepo.GetLessonRoomStateByLessonID(r.ctx, db, r.command.LessonID)
	if err != nil && err != domain.ErrLessonRoomStateNotFound {
		return fmt.Errorf("LessonRoomStateRepo.GetLessonRoomStateByLessonID %s: %w", r.command.LessonID, err)
	}

	if state.CurrentPolling != nil {
		state.CurrentPolling = nil

		if err := r.lessonRoomStateRepo.UpsertCurrentPollingState(r.ctx, db, r.command.LessonID, state.CurrentPolling); err != nil {
			return fmt.Errorf("lessonRoomStateRepo.UpsertCurrentPollingState: %w", err)
		}

		// update learner state
		if err := r.lessonMemberRepo.UpsertAllLessonMemberStateByStateType(
			r.ctx,
			db,
			r.command.LessonID,
			domain.LearnerStateTypePollingAnswer,
			&repo.StateValueDTO{
				BoolValue:        database.Bool(false),
				StringArrayValue: database.TextArray([]string{}),
			},
		); err != nil {
			return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
		}
	}

	return nil
}

func (r *ResetPollingCommandHandler) Execute() error {
	switch r.db.(type) {
	case pgx.Tx:
		return r.pExecute(r.db)
	default:
		return database.ExecInTx(r.ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
			return r.pExecute(tx)
		})
	}
}
