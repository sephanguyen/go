package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type EndPollingCommand struct {
	*VirtualClassroomCommand
}

type EndPollingCommandHandler struct {
	command                  *EndPollingCommand
	ctx                      context.Context
	db                       database.Ext
	lessonMemberRepo         infrastructure.LessonMemberRepo
	virtualLessonPollingRepo infrastructure.VirtualLessonPollingRepo
	lessonRoomStateRepo      infrastructure.LessonRoomStateRepo
}

func (h *EndPollingCommandHandler) pExecute(db database.Ext) error {
	state, err := h.lessonRoomStateRepo.GetLessonRoomStateByLessonID(h.ctx, db, h.command.LessonID)
	if err != nil && err != domain.ErrLessonRoomStateNotFound {
		return fmt.Errorf("LessonRoomStateRepo.GetLessonRoomStateByLessonID: %w with lessonId %s", err, h.command.LessonID)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("the Polling not exists")
	}
	if state.CurrentPolling.Status != domain.CurrentPollingStatusStopped {
		return fmt.Errorf("permission denied: Can't end polling when polling not stop")
	}

	// save polling
	var learnerStates repo.LessonMemberStateDTOs
	filter := &repo.MemberStatesFilter{}
	errFilter := multierr.Combine(
		filter.LessonID.Set(h.command.LessonID),
		filter.UserID.Set(nil),
		filter.StateType.Set(domain.LearnerStateTypePollingAnswer),
	)
	if errFilter != nil {
		return fmt.Errorf("could not filter to get lesson member states: %w", errFilter)
	}
	learnerStates, err = h.lessonMemberRepo.GetLessonMemberStatesWithParams(h.ctx, db, filter)
	if err != nil {
		return err
	}
	e := &repo.VirtualLessonPolling{}
	database.AllNullEntity(e)
	srcOptions := pgtype.JSONB{}
	if err := srcOptions.Set(state.CurrentPolling.Options); err != nil {
		return fmt.Errorf("could not marshal options to jsonb: %w", err)
	}
	srcAnswers := pgtype.JSONB{}
	if err := srcAnswers.Set(learnerStates); err != nil {
		return fmt.Errorf("could not marshal answers to jsonb: %w", err)
	}

	pollID := idutil.ULIDNow()
	if err := multierr.Combine(
		e.PollID.Set(pollID),
		e.LessonID.Set(h.command.LessonID),
		e.Options.Set(srcOptions),
		e.StudentsAnswers.Set(srcAnswers),
		e.CreatedAt.Set(state.CurrentPolling.CreatedAt),
		e.StoppedAt.Set(state.CurrentPolling.StoppedAt),
		e.UpdatedAt.Set(state.CurrentPolling.StoppedAt),
		e.EndedAt.Set(time.Now()),
	); err != nil {
		return err
	}
	_, err = h.virtualLessonPollingRepo.Create(h.ctx, db, e)
	if err != nil {
		return fmt.Errorf("LessonPollingRepo.Create: %v", err)
	}

	// update room state
	state.CurrentPolling = nil

	if err := h.lessonRoomStateRepo.UpsertCurrentPollingState(h.ctx, db, h.command.LessonID, state.CurrentPolling); err != nil {
		return fmt.Errorf("virtualLessonRepo.UpdateLessonRoomState: %w", err)
	}

	// update user state
	if err := h.lessonMemberRepo.UpsertAllLessonMemberStateByStateType(
		h.ctx,
		db,
		h.command.LessonID,
		domain.LearnerStateTypePollingAnswer,
		&repo.StateValueDTO{
			BoolValue:        database.Bool(false),
			StringArrayValue: database.TextArray([]string{}),
		},
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertAllLessonMemberStateByStateType: %w", err)
	}
	return nil
}

func (h *EndPollingCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
