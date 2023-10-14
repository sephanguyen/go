package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"

	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

type SubmitPollingAnswerCommand struct {
	*VirtualClassroomCommand
	UserID  string // user who will be submit answer
	Answers []string
}

type SubmitPollingAnswerCommandHandler struct {
	command             *SubmitPollingAnswerCommand
	ctx                 context.Context
	db                  database.Ext
	lessonMemberRepo    infrastructure.LessonMemberRepo
	lessonRoomStateRepo infrastructure.LessonRoomStateRepo
}

func (h *SubmitPollingAnswerCommandHandler) pExecute(db database.Ext) error {
	if len(h.command.Answers) == 0 {
		return fmt.Errorf("at least 1 answer")
	}
	state, err := h.lessonRoomStateRepo.GetLessonRoomStateByLessonID(h.ctx, db, h.command.LessonID)
	if err != nil {
		return fmt.Errorf("LessonRoomStateRepo.GetLessonRoomStateByLessonID: %w with lessonId %s", err, h.command.LessonID)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("the Polling not exists")
	}
	if state.CurrentPolling.Status != domain.CurrentPollingStatusStarted {
		return fmt.Errorf("permission denied: Can't submit answer when polling not start")
	}
	options := state.CurrentPolling.Options
	if err := options.ValidatePollingOptions(h.command.Answers); err != nil {
		return err
	}

	var learnerStates repo.LessonMemberStateDTOs
	filter := repo.MemberStatesFilter{}
	errFilter := multierr.Combine(
		filter.LessonID.Set(h.command.LessonID),
		filter.UserID.Set(h.command.UserID),
		filter.StateType.Set(domain.LearnerStateTypePollingAnswer),
	)
	if errFilter != nil {
		return errFilter
	}

	learnerStates, err = h.lessonMemberRepo.GetLessonMemberStatesWithParams(h.ctx, db, &filter)
	if err != nil {
		return err
	}
	if len(learnerStates) > 0 && len(learnerStates[0].StringArrayValue.Elements) > 0 {
		return fmt.Errorf("permission denied: Only submit 1 time")
	}
	memberState := &repo.LessonMemberStateDTO{}
	database.AllNullEntity(memberState)

	now := time.Now()
	if err := multierr.Combine(
		memberState.LessonID.Set(h.command.LessonID),
		memberState.UserID.Set(h.command.UserID),
		memberState.StateType.Set(domain.LearnerStateTypePollingAnswer),
		memberState.CreatedAt.Set(now),
		memberState.UpdatedAt.Set(now),
		memberState.BoolValue.Set(false),
		memberState.StringArrayValue.Set(h.command.Answers),
	); err != nil {
		return err
	}

	if err := h.lessonMemberRepo.UpsertLessonMemberState(
		h.ctx,
		db,
		memberState,
	); err != nil {
		return fmt.Errorf("LessonMemberRepo.UpsertLessonMemberState: %w", err)
	}
	return nil
}

func (h *SubmitPollingAnswerCommandHandler) Execute() error {
	switch h.db.(type) {
	case pgx.Tx:
		return h.pExecute(h.db)
	default:
		return database.ExecInTx(h.ctx, h.db, func(ctx context.Context, tx pgx.Tx) error {
			return h.pExecute(tx)
		})
	}
}
