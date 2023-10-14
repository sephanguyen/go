package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"

	"github.com/jackc/pgtype"
)

type (
	PlayerState      string
	LearnerStateType string
	PollingState     string
)

const (
	PlayerStatePause   PlayerState = "PLAYER_STATE_PAUSE"
	PlayerStatePlaying PlayerState = "PLAYER_STATE_PLAYING"
	PlayerStateEnded   PlayerState = "PLAYER_STATE_ENDED"

	LearnerStateTypeHandsUp       LearnerStateType = "LEARNER_STATE_TYPE_HANDS_UP"
	LearnerStateTypeAnnotation    LearnerStateType = "LEARNER_STATE_TYPE_ANNOTATION"
	LearnerStateTypePollingAnswer LearnerStateType = "LEARNER_STATE_TYPE_POLLING_ANSWER"
	LearnerStateTypeChat          LearnerStateType = "LEARNER_STATE_TYPE_CHAT"

	PollingStateStarted PollingState = "POLLING_STATE_STARTED"
	PollingStateStopped PollingState = "POLLING_STATE_STOPPED"
	PollingStateEnded   PollingState = "POLLING_STATE_ENDED"
)

type LiveLessonState struct {
	LessonID   string
	RoomState  *LessonRoomState
	UserStates *UserStates
}

func NewLiveLessonState(lessonID pgtype.Text, roomState pgtype.JSONB, learnersStates entities.LessonMemberStates) (*LiveLessonState, error) {
	ls := LiveLessonState{
		LessonID: lessonID.String,
	}

	err := ls.SetRoomState(roomState)
	if err != nil {
		return nil, err
	}

	// check all learnerStates must belong to a lesson
	for _, v := range learnersStates {
		if v.LessonID.String != lessonID.String {
			return nil, fmt.Errorf(`learner %s not belong to lesson %s`, v.UserID.String, lessonID.String)
		}
	}

	ls.UserStates = NewUserState(learnersStates)
	return &ls, nil
}

func (ls *LiveLessonState) SetRoomState(src pgtype.JSONB) error {
	state := &LessonRoomState{}
	err := src.AssignTo(state)
	if err != nil {
		return fmt.Errorf("could to unmarshal roomstate: %v", err)
	}
	ls.RoomState = state
	return nil
}

type LessonRoomState struct {
	CurrentMaterial *CurrentMaterial `json:"current_material,omitempty"`
	CurrentPolling  *CurrentPolling  `json:"current_polling,omitempty"`
	Recording       *RecordingState  `json:"recording,omitempty"`
}

func (l *LessonRoomState) IsValid() error {
	if l.CurrentMaterial != nil {
		if err := l.CurrentMaterial.IsValid(); err != nil {
			return fmt.Errorf("invalid current_material: %v", err)
		}
	}
	if l.CurrentPolling != nil {
		if err := l.CurrentPolling.IsValid(); err != nil {
			return fmt.Errorf("invalid current_polling: %v", err)
		}
	}

	return nil
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

	if c.VideoState != nil {
		if err := c.VideoState.IsValid(); err != nil {
			return fmt.Errorf("invalid video_state: %v", err)
		}
	}

	if c.AudioState != nil {
		if err := c.AudioState.IsValid(); err != nil {
			return fmt.Errorf("invalid audio_state: %v", err)
		}
	}

	return nil
}

type CurrentPolling struct {
	Options   []*PollingOption `json:"options"`
	Status    PollingState     `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	StoppedAt time.Time        `json:"stopped_at"`
}

type PollingOption struct {
	Answer    string `json:"answer"`
	IsCorrect bool   `json:"is_correct"`
}

func (c *CurrentPolling) IsValid() error {
	if len(c.Options) == 0 {
		return fmt.Errorf("options could not be empty")
	}
	if !c.StoppedAt.IsZero() && c.StoppedAt.Before(c.CreatedAt) {
		return fmt.Errorf("stopped_at could not before created_at")
	}

	return nil
}

type RecordingState struct {
	IsRecording bool    `json:"is_recording"`
	Creator     *string `json:"creator,omitempty"`
}

type VideoState struct {
	CurrentTime Duration    `json:"current_time"`
	PlayerState PlayerState `json:"player_state"`
}

func (v *VideoState) IsValid() error {
	if len(v.PlayerState) == 0 {
		return fmt.Errorf("invalid player_state %s", v.PlayerState)
	}

	if v.PlayerState == PlayerStatePlaying || v.PlayerState == PlayerStatePause {
		if v.CurrentTime < 0 {
			return fmt.Errorf("invalid current_time %v", v.CurrentTime)
		}
	}

	return nil
}

type AudioState struct {
	CurrentTime Duration    `json:"current_time"`
	PlayerState PlayerState `json:"player_state"`
}

func (a *AudioState) IsValid() error {
	if len(a.PlayerState) == 0 {
		return fmt.Errorf("invalid player_state %s", a.PlayerState)
	}

	if a.PlayerState == PlayerStatePlaying || a.PlayerState == PlayerStatePause {
		if a.CurrentTime < 0 {
			return fmt.Errorf("invalid current_time %v", a.CurrentTime)
		}
	}

	return nil
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return fmt.Errorf("invalid duration")
	}
}

func (d *Duration) Duration() time.Duration {
	return time.Duration(*d)
}

type UserStates struct {
	LearnersState []*LearnerState
}

func NewUserState(learnersSt entities.LessonMemberStates) *UserStates {
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

func NewLearnerState(userID string, states entities.LessonMemberStates) *LearnerState {
	res := LearnerState{
		UserID:        userID,
		HandsUp:       &UserHandsUp{},
		Annotation:    &UserAnnotation{},
		PollingAnswer: &UserPollingAnswer{},
		Chat:          &UserChat{},
	}
	for _, state := range states {
		if state.UserID.String != res.UserID {
			continue
		}
		switch state.StateType.String {
		case string(LearnerStateTypeHandsUp):
			res.HandsUp = &UserHandsUp{
				Value:     state.BoolValue.Bool,
				UpdatedAt: state.UpdatedAt.Time,
			}
		case string(LearnerStateTypeAnnotation):
			res.Annotation = &UserAnnotation{
				Value:     state.BoolValue.Bool,
				UpdatedAt: state.UpdatedAt.Time,
			}
		case string(LearnerStateTypePollingAnswer):
			var arrayValue []string
			for _, r := range state.StringArrayValue.Elements {
				arrayValue = append(arrayValue, r.String)
			}
			res.PollingAnswer = &UserPollingAnswer{
				StringArrayValue: arrayValue,
				UpdatedAt:        state.UpdatedAt.Time,
			}
		case string(LearnerStateTypeChat):
			res.Chat = &UserChat{
				Value:     state.BoolValue.Bool,
				UpdatedAt: state.UpdatedAt.Time,
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
