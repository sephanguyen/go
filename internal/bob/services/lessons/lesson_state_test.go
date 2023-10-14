package services

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/bob/entities"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLiveLessonState(t *testing.T) {
	marshalToJSONB := func(src string, isNull bool) pgtype.JSONB {
		res := pgtype.JSONB{}
		var err error
		if isNull {
			err = res.Set(nil)
		} else {
			err = res.Set(src)
		}
		require.NoError(t, err)
		return res
	}
	now := time.Now().UTC()
	nowString, err := now.MarshalText()
	require.NoError(t, err)

	tcs := []struct {
		name           string
		lessonID       pgtype.Text
		roomState      pgtype.JSONB
		learnersStates entities.LessonMemberStates
		expected       *LiveLessonState
		hasError       bool
	}{
		{
			name:     "create new live lesson state with full data",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				},
				"current_polling": {
					"options": [
						{
							"answer": "A",
							"is_correct": true
						},
						{
							"answer": "B",
							"is_correct": false
						},
						{
							"answer": "C",
							"is_correct": false
						}
					],
					"status": "POLLING_STATE_STARTED",
					"created_at": "`+string(nowString)+`"
				}
			}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:         database.Text("lesson-1"),
					UserID:           database.Text("user-3"),
					StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
					CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
					StringArrayValue: database.TextArray([]string{"A"}),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
					CurrentPolling: &CurrentPolling{
						Options: []*PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
						Status:    PollingStateStarted,
						CreatedAt: now,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-20 * time.Hour),
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state with full data and audio state",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"audio_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				},
				"current_polling": {
					"options": [
						{
							"answer": "A",
							"is_correct": true
						},
						{
							"answer": "B",
							"is_correct": false
						},
						{
							"answer": "C",
							"is_correct": false
						}
					],
					"status": "POLLING_STATE_STARTED",
					"created_at": "`+string(nowString)+`"
				}
			}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:         database.Text("lesson-1"),
					UserID:           database.Text("user-3"),
					StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
					CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
					StringArrayValue: database.TextArray([]string{"A"}),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						AudioState: &AudioState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
					CurrentPolling: &CurrentPolling{
						Options: []*PollingOption{
							{
								Answer:    "A",
								IsCorrect: true,
							},
							{
								Answer:    "B",
								IsCorrect: false,
							},
							{
								Answer:    "C",
								IsCorrect: false,
							},
						},
						Status:    PollingStateStarted,
						CreatedAt: now,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-20 * time.Hour),
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state without current_time",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"video_state": {
						"player_state": "PLAYER_STATE_ENDED"
					}
				}
			}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							PlayerState: PlayerStateEnded,
						},
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state without video_state and audio_state",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`"
				}
			}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state without current_material",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(
				`{}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID:  "lesson-1",
				RoomState: &LessonRoomState{},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state without current_polling",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:         database.Text("lesson-1"),
					UserID:           database.Text("user-3"),
					StateType:        database.Text("LEARNER_STATE_TYPE_POLLING_ANSWER"),
					CreatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt:        database.Timestamptz(now.Add(-20 * time.Hour)),
					StringArrayValue: database.TextArray([]string{"A"}),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
				},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{
								StringArrayValue: []string{"A"},
								UpdatedAt:        now.Add(-20 * time.Hour),
							},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state without any room's states",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(
				`{}`,
				true,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text(string(LearnerStateTypeChat)),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			expected: &LiveLessonState{
				LessonID:  "lesson-1",
				RoomState: &LessonRoomState{},
				UserStates: &UserStates{
					LearnersState: []*LearnerState{
						{
							UserID: "user-1",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-2",
							HandsUp: &UserHandsUp{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							Annotation: &UserAnnotation{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     true,
								UpdatedAt: now.Add(-20 * time.Minute),
							},
						},
						{
							UserID: "user-3",
							HandsUp: &UserHandsUp{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							Annotation: &UserAnnotation{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
							PollingAnswer: &UserPollingAnswer{},
							Chat: &UserChat{
								Value:     false,
								UpdatedAt: now.Add(-20 * time.Hour),
							},
						},
					},
				},
			},
		},
		{
			name:     "create new live lesson state without any learner's state",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`,
				false,
			),
			expected: &LiveLessonState{
				LessonID: "lesson-1",
				RoomState: &LessonRoomState{
					CurrentMaterial: &CurrentMaterial{
						MediaID:   "media-1",
						UpdatedAt: now,
						VideoState: &VideoState{
							CurrentTime: Duration(23 * time.Minute),
							PlayerState: PlayerStatePlaying,
						},
					},
				},
				UserStates: &UserStates{},
			},
		},
		{
			name:     "create new live lesson state with wrong format room's state",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					},
				}
			}`,
				false,
			),
			hasError: true,
		},
		{
			name:     "create new live lesson state with learner's states not belong same lesson",
			lessonID: database.Text("lesson-1"),
			roomState: marshalToJSONB(`
			{
				"current_material": {
					"media_id": "media-1",
					"updated_at": "`+string(nowString)+`",
					"video_state": {
						"current_time": "23m",
						"player_state": "PLAYER_STATE_PLAYING"
					}
				}
			}`,
				false,
			),
			learnersStates: entities.LessonMemberStates{
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-2"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-2 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_HANDS_UP"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-1"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-2"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Minute)),
					BoolValue: database.Bool(true),
				},
				{
					LessonID:  database.Text("lesson-1"),
					UserID:    database.Text("user-3"),
					StateType: database.Text("LEARNER_STATE_TYPE_ANNOTATION"),
					CreatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					UpdatedAt: database.Timestamptz(now.Add(-20 * time.Hour)),
					BoolValue: database.Bool(false),
				},
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := NewLiveLessonState(tc.lessonID, tc.roomState, tc.learnersStates)
			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				if tc.expected.UserStates != nil && tc.expected.UserStates.LearnersState != nil {
					lsExpected := tc.expected.UserStates.LearnersState
					tc.expected.UserStates.LearnersState = nil
					lsActual := actual.UserStates.LearnersState
					actual.UserStates.LearnersState = nil
					assert.ElementsMatch(t, lsExpected, lsActual)
				}
				assert.EqualValues(t, *tc.expected, *actual)
			}
		})
	}
}
