package domain

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualClassRoom_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name             string
		virtualClassRoom *VirtualClassroom
		setup            func(ctx context.Context) (*mock.UserModulePortMock, *mock.LessonPortMock)
		isValid          bool
	}{
		{
			name: "full fields",
			virtualClassRoom: &VirtualClassroom{
				ID:        "virtual-class-room-id",
				CreatedAt: now,
				UpdatedAt: now,
				Room: &VirtualRoom{
					StreamingProvider: &StreamingProvider{
						StreamingRoomID:     "streaming-room-id",
						TotalStreamingSlots: 2,
					},
					AttendeeStates: AttendeeStates{
						{
							UserID:           "user-id-1",
							RaisingHandState: &AttendeeRaisingHandState{},
							AnnotationState:  &AttendeeAnnotationState{},
							PollingAnswer:    nil,
						},
						{
							UserID:           "user-id-2",
							RaisingHandState: &AttendeeRaisingHandState{},
							AnnotationState:  &AttendeeAnnotationState{},
							PollingAnswer:    nil,
						},
					},
				},
				Lesson: &VirtualLesson{
					LessonID: "lesson-id-1",
				},
			},
			setup: func(ctx context.Context) (*mock.UserModulePortMock, *mock.LessonPortMock) {
				userModulePortMock := &mock.UserModulePortMock{}
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{"user-id-1", "user-id-2"})
					return ids, nil
				}, 1)

				lessonPortMock := &mock.LessonPortMock{}
				lessonPortMock.SetIsLessonMediumOnline(func(ctx context.Context, lessonID string) (bool, error) {
					assert.Equal(t, "lesson-id-1", lessonID)
					return true, nil
				}, 1)
				lessonPortMock.SetCheckLessonMemberIDs(func(ctx context.Context, lessonID string, userIDs []string) (memberIDs []string, err error) {
					assert.Equal(t, "lesson-id-1", lessonID)
					assert.ElementsMatch(t, []string{"user-id-1", "user-id-2"}, userIDs)
					return userIDs, nil
				}, 1)
				return userModulePortMock, lessonPortMock
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			virtualClassRoom: &VirtualClassroom{
				ID: "virtual-class-room-id",
				Room: &VirtualRoom{
					StreamingProvider: &StreamingProvider{
						StreamingRoomID:     "streaming-room-id",
						TotalStreamingSlots: 2,
					},
				},
				Lesson: &VirtualLesson{
					LessonID: "lesson-id-1",
				},
			},
			setup: func(ctx context.Context) (*mock.UserModulePortMock, *mock.LessonPortMock) {
				lessonPortMock := &mock.LessonPortMock{}
				lessonPortMock.SetIsLessonMediumOnline(func(ctx context.Context, lessonID string) (bool, error) {
					assert.Equal(t, "lesson-id-1", lessonID)
					return true, nil
				}, 1)
				return &mock.UserModulePortMock{}, lessonPortMock
			},
			isValid: true,
		},
		{
			name: "missing id field",
			virtualClassRoom: &VirtualClassroom{
				Room: &VirtualRoom{
					StreamingProvider: &StreamingProvider{
						StreamingRoomID:     "streaming-room-id",
						TotalStreamingSlots: 2,
					},
				},
				Lesson: &VirtualLesson{
					LessonID: "lesson-id-1",
				},
			},
			setup: func(ctx context.Context) (*mock.UserModulePortMock, *mock.LessonPortMock) {
				return &mock.UserModulePortMock{}, &mock.LessonPortMock{}
			},
			isValid: false,
		},
		{
			name: "missing room field",
			virtualClassRoom: &VirtualClassroom{
				ID: "virtual-class-room-id",
				Lesson: &VirtualLesson{
					LessonID: "lesson-id-1",
				},
			},
			setup: func(ctx context.Context) (*mock.UserModulePortMock, *mock.LessonPortMock) {
				return &mock.UserModulePortMock{}, &mock.LessonPortMock{}
			},
			isValid: false,
		},
		{
			name: "missing lesson field",
			virtualClassRoom: &VirtualClassroom{
				ID: "virtual-class-room-id",
				Room: &VirtualRoom{
					StreamingProvider: &StreamingProvider{
						StreamingRoomID:     "streaming-room-id",
						TotalStreamingSlots: 2,
					},
				},
			},
			setup: func(ctx context.Context) (*mock.UserModulePortMock, *mock.LessonPortMock) {
				return &mock.UserModulePortMock{}, &mock.LessonPortMock{}
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			userModulePortMock, lessonPortMock := tc.setup(context.Background())
			if tc.virtualClassRoom.Lesson != nil {
				tc.virtualClassRoom.Lesson.virtualLessonPort = lessonPortMock
			}
			if tc.virtualClassRoom.Room != nil {
				tc.virtualClassRoom.Room.userModulePort = userModulePortMock
			}
			err := tc.virtualClassRoom.IsValid(context.Background())
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.NoError(t, userModulePortMock.AllFuncBeCalledAsExpected())
			require.NoError(t, lessonPortMock.AllFuncBeCalledAsExpected())
		})
	}
}

func TestVirtualClassRoomBuilder_BuildDraft(t *testing.T) {
	now := time.Now()
	creator := "user-id-1"
	data := &VirtualClassroom{
		ID:        "virtual-class-room-id",
		CreatedAt: now,
		UpdatedAt: now,
		Room: &VirtualRoom{
			EndedAt: &now,
			StreamingProvider: &StreamingProvider{
				StreamingRoomID:     "streaming-room-id",
				TotalStreamingSlots: 2,
			},
			Materials: Materials{
				&VideoMaterial{
					ID:      "media-id-1",
					Name:    "video",
					VideoID: "video-id-1",
				},
				&PDFMaterial{
					ID:   "media-id-2",
					Name: "pdf",
					URL:  "https://example.com/random-path/name.pdf",
					ConvertedImageURL: &ConvertedImage{
						Width:    2,
						Height:   3,
						ImageURL: "https://example.com/random-path/name.png",
					},
				},
			},
			AttendeeStates: AttendeeStates{
				{
					UserID: "user-id-1",
					RaisingHandState: &AttendeeRaisingHandState{
						IsRaisingHand: true,
						UpdatedAt:     now,
					},
					AnnotationState: &AttendeeAnnotationState{
						BeAllowed: true,
						UpdatedAt: now,
					},
					PollingAnswer: &AttendeePollingAnswerState{
						Answer:    []string{"A"},
						UpdatedAt: now,
					},
				},
				{
					UserID: "user-id-2",
					RaisingHandState: &AttendeeRaisingHandState{
						IsRaisingHand: true,
						UpdatedAt:     now,
					},
					AnnotationState: &AttendeeAnnotationState{
						BeAllowed: false,
						UpdatedAt: now,
					},
					PollingAnswer: &AttendeePollingAnswerState{
						Answer:    []string{"C"},
						UpdatedAt: now,
					},
				},
			},
			PresentMaterialState: &VideoPresentMaterialState{
				Material: &VideoMaterial{
					ID:      "media-id-1",
					Name:    "video",
					VideoID: "video-id-1",
				},
				UpdatedAt: now,
				VideoState: &VideoState{
					CurrentTime: Duration(2 * time.Minute),
					PlayerState: PlayerStatePlaying,
				},
			},
			CurrentPolling: &CurrentPolling{
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			RecordingState: &RecordingState{
				IsRecording: true,
				Creator:     &creator,
			},
		},
		Lesson: &VirtualLesson{
			LessonID: "lesson-id-1",
		},
	}

	builder := NewVirtualClassRoomBuilder()
	actual := builder.
		WithID(data.ID).
		WithModificationTime(data.CreatedAt, data.UpdatedAt).
		WithLessonID(data.Lesson.LessonID).
		WithStreamingProvider(data.Room.StreamingProvider).
		WithEndedAt(*data.Room.EndedAt).
		WithMaterials(data.Room.Materials).
		WithPresentMaterialState(data.Room.PresentMaterialState).
		WithCurrentPolling(data.Room.CurrentPolling).
		WithRecordingState(data.Room.RecordingState).
		WithAttendeeStates(data.Room.AttendeeStates).
		BuildDraft()

	assert.EqualValues(t, data, actual)
}
