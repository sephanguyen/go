package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type EndPollingCommand struct {
	*ModifyLiveRoomCommand
}

type EndPollingCommandHandler struct {
	command      *EndPollingCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
	LiveRoomPoll            infrastructure.LiveRoomPoll
}

func (e *EndPollingCommandHandler) pExecute(db database.Ext) error {
	channelID := e.command.ChannelID
	state, err := e.LiveRoomStateRepo.GetLiveRoomStateByChannelID(e.ctx, db, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		return fmt.Errorf("error in LiveRoomStateRepo.GetLiveRoomStateByChannelID, channel %s: %w", channelID, err)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("the polling does not exist in live room %s", channelID)
	}
	if state.CurrentPolling.Status != vc_domain.CurrentPollingStatusStopped {
		return fmt.Errorf("cannot end polling in live room %s as polling is not in stopped status", channelID)
	}

	// get students and student answers
	stateType := vc_domain.LearnerStateTypePollingAnswer
	filter := &domain.SearchLiveRoomMemberStateParams{
		ChannelID: channelID,
		StateType: string(stateType),
	}
	liveRoomMemberStates, err := e.LiveRoomMemberStateRepo.GetLiveRoomMemberStatesWithParams(e.ctx, db, filter)
	if err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.GetLiveRoomMemberStatesWithParams, params %v: %w", filter, err)
	}
	studentAnswersList := liveRoomMemberStates.GetStudentAnswersList()
	userList := liveRoomMemberStates.GetStudentList()

	// save polling
	now := time.Now()
	liveRoomPoll := &domain.LiveRoomPoll{
		ChannelID:      channelID,
		Options:        &state.CurrentPolling.Options,
		StudentAnswers: studentAnswersList,
		CreatedAt:      state.CurrentPolling.CreatedAt,
		StoppedAt:      state.CurrentPolling.StoppedAt,
		UpdatedAt:      *state.CurrentPolling.StoppedAt,
		EndedAt:        &now,
	}
	if err = e.LiveRoomPoll.CreateLiveRoomPoll(e.ctx, db, liveRoomPoll); err != nil {
		return fmt.Errorf("error in LiveRoomPoll.CreateLiveRoomPoll, channel %s: %w", channelID, err)
	}

	// update live room state
	state.CurrentPolling = nil
	if err := e.LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState(e.ctx, db, channelID, state.CurrentPolling); err != nil {
		return fmt.Errorf("error in LiveRoomStateRepo.UpsertLiveRoomCurrentPollingState, channel %s: %w", channelID, err)
	}

	// update live room member state
	if err := e.LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState(
		e.ctx,
		db,
		channelID,
		userList,
		stateType,
		&vc_domain.StateValue{
			BoolValue:        false,
			StringArrayValue: []string{},
		},
	); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState, channel %s, users %v, state %s: %w",
			channelID,
			userList,
			stateType,
			err,
		)
	}

	return nil
}

func (e *EndPollingCommandHandler) Execute() error {
	switch e.lessonmgmtDB.(type) {
	case pgx.Tx:
		return e.pExecute(e.lessonmgmtDB)
	default:
		return database.ExecInTx(e.ctx, e.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return e.pExecute(tx)
		})
	}
}
