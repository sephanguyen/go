package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"
	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgx/v4"
)

type SubmitPollingAnswerCommand struct {
	*ModifyLiveRoomCommand
	UserID  string // user whose answer was submitted
	Answers []string
}

type SubmitPollingAnswerCommandHandler struct {
	command      *SubmitPollingAnswerCommand
	ctx          context.Context
	lessonmgmtDB database.Ext

	LiveRoomStateRepo       infrastructure.LiveRoomStateRepo
	LiveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo
}

func (s *SubmitPollingAnswerCommandHandler) pExecute(db database.Ext) error {
	if len(s.command.Answers) == 0 {
		return fmt.Errorf("user needs to select at least 1 answer")
	}

	// validate if polling exist
	channelID := s.command.ChannelID
	userID := s.command.UserID
	state, err := s.LiveRoomStateRepo.GetLiveRoomStateByChannelID(s.ctx, db, channelID)
	if err != nil && err != domain.ErrChannelNotFound {
		return fmt.Errorf("error in LiveRoomStateRepo.GetLiveRoomStateByChannelID, channel %s: %w", channelID, err)
	}
	if state.CurrentPolling == nil {
		return fmt.Errorf("the polling does not exist in live room %s", channelID)
	}
	if state.CurrentPolling.Status != vc_domain.CurrentPollingStatusStarted {
		return fmt.Errorf("cannot submit polling answer in live room %s as polling is not in started status", channelID)
	}

	pollingOptions := state.CurrentPolling.Options
	if err := pollingOptions.ValidatePollingOptions(s.command.Answers); err != nil {
		return err
	}

	// get if there's an existing answer
	stateType := vc_domain.LearnerStateTypePollingAnswer
	userIDs := []string{userID}
	filter := &domain.SearchLiveRoomMemberStateParams{
		ChannelID: channelID,
		UserIDs:   userIDs,
		StateType: string(stateType),
	}
	liveRoomMemberStates, err := s.LiveRoomMemberStateRepo.GetLiveRoomMemberStatesWithParams(s.ctx, db, filter)
	if err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.GetLiveRoomMemberStatesWithParams, params %v: %w", filter, err)
	}
	if len(liveRoomMemberStates) > 0 && len(liveRoomMemberStates[0].StringArrayValue) > 0 {
		return fmt.Errorf("the user %s can only submit polling answer one time for live room %s", userID, channelID)
	}

	// save answer
	stateValue := &vc_domain.StateValue{
		BoolValue:        false,
		StringArrayValue: s.command.Answers,
	}
	if err := s.LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState(s.ctx, db, channelID, userIDs, stateType, stateValue); err != nil {
		return fmt.Errorf("error in LiveRoomMemberStateRepo.BulkUpsertLiveRoomMembersState, channel %s, users %v, state %s: %w",
			channelID,
			userIDs,
			stateType,
			err,
		)
	}

	return nil
}

func (s *SubmitPollingAnswerCommandHandler) Execute() error {
	switch s.lessonmgmtDB.(type) {
	case pgx.Tx:
		return s.pExecute(s.lessonmgmtDB)
	default:
		return database.ExecInTx(s.ctx, s.lessonmgmtDB, func(ctx context.Context, tx pgx.Tx) error {
			return s.pExecute(tx)
		})
	}
}
