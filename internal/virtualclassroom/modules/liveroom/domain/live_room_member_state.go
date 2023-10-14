package domain

import (
	"fmt"
	"time"

	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type SearchLiveRoomMemberStateParams struct {
	ChannelID string
	UserIDs   []string
	StateType string
}

type LiveRoomMemberState struct {
	ChannelID        string
	UserID           string
	StateType        string
	StringArrayValue []string
	BoolValue        bool
	UpdatedAt        time.Time
	CreatedAt        time.Time
	DeletedAt        *time.Time
}

type LiveRoomMemberStates []*LiveRoomMemberState

func (ls LiveRoomMemberStates) ValidInChannel(channelID string) error {
	for _, learnerState := range ls {
		if learnerState.ChannelID != channelID {
			return fmt.Errorf(`got learner %s state that does not belong to channel %s, got state for channel %s`,
				learnerState.UserID,
				channelID,
				learnerState.ChannelID,
			)
		}
	}
	return nil
}

func (ls LiveRoomMemberStates) GroupByUserID() map[string]LiveRoomMemberStates {
	res := make(map[string]LiveRoomMemberStates)

	for _, state := range ls {
		userID := state.UserID
		if v, ok := res[userID]; !ok {
			res[userID] = LiveRoomMemberStates{state}
		} else {
			v = append(v, state)
			res[userID] = v
		}
	}

	return res
}

func (ls LiveRoomMemberStates) ConvertToLearnerStateByUserID(userID string) *vc_domain.LearnerState {
	learnerState := &vc_domain.LearnerState{
		UserID:        userID,
		HandsUp:       &vc_domain.UserHandsUp{},
		Annotation:    &vc_domain.UserAnnotation{},
		PollingAnswer: &vc_domain.UserPollingAnswer{},
		Chat:          &vc_domain.UserChat{},
	}
	for _, state := range ls {
		if state.UserID != learnerState.UserID {
			continue
		}
		switch state.StateType {
		case string(vc_domain.LearnerStateTypeHandsUp):
			learnerState.HandsUp = &vc_domain.UserHandsUp{
				Value:     state.BoolValue,
				UpdatedAt: state.UpdatedAt,
			}
		case string(vc_domain.LearnerStateTypeAnnotation):
			learnerState.Annotation = &vc_domain.UserAnnotation{
				Value:     state.BoolValue,
				UpdatedAt: state.UpdatedAt,
			}
		case string(vc_domain.LearnerStateTypePollingAnswer):
			var arrayValue []string
			arrayValue = append(arrayValue, state.StringArrayValue...)
			learnerState.PollingAnswer = &vc_domain.UserPollingAnswer{
				StringArrayValue: arrayValue,
				UpdatedAt:        state.UpdatedAt,
			}
		case string(vc_domain.LearnerStateTypeChat):
			learnerState.Chat = &vc_domain.UserChat{
				Value:     state.BoolValue,
				UpdatedAt: state.UpdatedAt,
			}
		}
	}

	return learnerState
}

func (ls LiveRoomMemberStates) ConvertToUserState() *vc_domain.UserStates {
	userStates := vc_domain.UserStates{}
	if len(ls) > 0 {
		statesByUser := ls.GroupByUserID()
		userStates.LearnersState = make([]*vc_domain.LearnerState, 0, len(statesByUser))
		for userID, memberStates := range statesByUser {
			userStates.LearnersState = append(
				userStates.LearnersState,
				memberStates.ConvertToLearnerStateByUserID(userID),
			)
		}
	}

	return &userStates
}

func (ls LiveRoomMemberStates) GetStudentAnswersList() StudentAnswersList {
	list := StudentAnswersList{}

	for _, memberState := range ls {
		if memberState.StateType != string(vc_domain.LearnerStateTypePollingAnswer) {
			continue
		}

		list = append(list, &StudentAnswers{
			UserID:    memberState.UserID,
			Answers:   memberState.StringArrayValue,
			CreatedAt: memberState.CreatedAt,
			UpdatedAt: memberState.UpdatedAt,
		})
	}

	return list
}

func (ls LiveRoomMemberStates) GetStudentList() []string {
	list := make([]string, 0, len(ls))
	listed := make(map[string]bool)

	for _, memberState := range ls {
		if listed[memberState.UserID] {
			continue
		}
		list = append(list, memberState.UserID)
		listed[memberState.UserID] = true
	}

	return list
}
