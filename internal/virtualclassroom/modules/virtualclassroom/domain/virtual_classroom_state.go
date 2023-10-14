package domain

import (
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

type (
	LearnerStateType string
	PollingState     string
)

const (
	LearnerStateTypeHandsUp       LearnerStateType = "LEARNER_STATE_TYPE_HANDS_UP"
	LearnerStateTypeAnnotation    LearnerStateType = "LEARNER_STATE_TYPE_ANNOTATION"
	LearnerStateTypePollingAnswer LearnerStateType = "LEARNER_STATE_TYPE_POLLING_ANSWER"
	LearnerStateTypeChat          LearnerStateType = "LEARNER_STATE_TYPE_CHAT"

	PollingStateStarted PollingState = "POLLING_STATE_STARTED"
	PollingStateStopped PollingState = "POLLING_STATE_STOPPED"
	PollingStateEnded   PollingState = "POLLING_STATE_ENDED"
)

type VirtualClassroomState struct {
	LessonID   string
	RoomState  *OldLessonRoomState
	UserStates *UserStates
}

func NewVirtualClassroomState(lessonID string, roomState *OldLessonRoomState, learnersStates LessonMemberStates) (*VirtualClassroomState, error) {
	ls := VirtualClassroomState{
		LessonID: lessonID,
	}
	ls.SetRoomState(roomState)
	// check all learnerStates must belong to a lesson
	for _, v := range learnersStates {
		if v.LessonID != lessonID {
			return nil, fmt.Errorf(`learner %s not belong to lesson %s`, v.UserID, lessonID)
		}
	}

	ls.UserStates = NewUserState(learnersStates)
	return &ls, nil
}

func (ls *VirtualClassroomState) SetRoomState(src *OldLessonRoomState) {
	ls.RoomState = src
}

func UnmarshalRoomStateJSON(src pgtype.JSONB) OldLessonRoomState {
	state := OldLessonRoomState{}
	err := src.AssignTo(&state)
	if err != nil {
		return OldLessonRoomState{}
	}
	return state
}

type VirtualClassroomRoomState struct {
	CurrentMaterial *CurrentMaterial `json:"current_material,omitempty"`
	CurrentPolling  *CurrentPolling  `json:"current_polling,omitempty"`
	Recording       *RecordingState  `json:"recording,omitempty"`
}

type CurrentMaterial struct {
	MediaID    string      `json:"media_id"`
	UpdatedAt  time.Time   `json:"updated_at"`
	VideoState *VideoState `json:"video_state,omitempty"`
	AudioState *AudioState `json:"audio_state,omitempty"`
}

func (c *CurrentMaterial) IsValid() error {
	if len(c.MediaID) == 0 {
		return fmt.Errorf("media_id could not be empty")
	}

	if c.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at could not be zero")
	}

	return nil
}

type UserStates struct {
	LearnersState []*LearnerState
}

func NewUserState(learnersSt LessonMemberStates) *UserStates {
	res := UserStates{}
	if len(learnersSt) != 0 {
		statesByUser := learnersSt.GroupByUserID()
		res.LearnersState = make([]*LearnerState, 0, len(statesByUser))
		for userID, userStates := range statesByUser {
			res.LearnersState = append(res.LearnersState, NewLearnerState(userID, userStates))
		}
	}

	return &res
}

type LearnerState struct {
	UserID        string
	HandsUp       *UserHandsUp
	Annotation    *UserAnnotation
	PollingAnswer *UserPollingAnswer
	Chat          *UserChat
}

func NewLearnerState(userID string, states LessonMemberStates) *LearnerState {
	res := LearnerState{
		UserID:        userID,
		HandsUp:       &UserHandsUp{},
		Annotation:    &UserAnnotation{},
		PollingAnswer: &UserPollingAnswer{},
		Chat:          &UserChat{},
	}
	for _, state := range states {
		if state.UserID != res.UserID {
			continue
		}
		switch state.StateType {
		case string(LearnerStateTypeHandsUp):
			res.HandsUp = &UserHandsUp{
				Value:     state.BoolValue,
				UpdatedAt: state.UpdatedAt,
			}
		case string(LearnerStateTypeAnnotation):
			res.Annotation = &UserAnnotation{
				Value:     state.BoolValue,
				UpdatedAt: state.UpdatedAt,
			}
		case string(LearnerStateTypePollingAnswer):
			var arrayValue []string
			arrayValue = append(arrayValue, state.StringArrayValue...)
			res.PollingAnswer = &UserPollingAnswer{
				StringArrayValue: arrayValue,
				UpdatedAt:        state.UpdatedAt,
			}
		case string(LearnerStateTypeChat):
			res.Chat = &UserChat{
				Value:     state.BoolValue,
				UpdatedAt: state.UpdatedAt,
			}
		}
	}

	return &res
}

type UserHandsUp struct {
	Value     bool
	UpdatedAt time.Time
}

type UserAnnotation struct {
	Value     bool
	UpdatedAt time.Time
}

type UserPollingAnswer struct {
	StringArrayValue []string
	UpdatedAt        time.Time
}

type UserChat struct {
	Value     bool
	UpdatedAt time.Time
}
