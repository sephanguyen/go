package domain

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVirtualRoom_isValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	creator := "user-id-1"
	tcs := []struct {
		name        string
		virtualRoom *VirtualRoom
		setup       func(ctx context.Context) *mock.UserModulePortMock
		isValid     bool
	}{
		{
			name: "full fields",
			virtualRoom: &VirtualRoom{
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
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				userModulePortMock := &mock.UserModulePortMock{}
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{creator})
					return ids, nil
				}, 1)
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{"user-id-1", "user-id-2"})
					return ids, nil
				}, 1)
				return userModulePortMock
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			virtualRoom: &VirtualRoom{
				//ID: "room-id",
				StreamingProvider: &StreamingProvider{
					StreamingRoomID:     "streaming-room-id",
					TotalStreamingSlots: 2,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: true,
		},
		{
			name:        "miss streaming provider field",
			virtualRoom: &VirtualRoom{},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "have invalid streaming provider field",
			virtualRoom: &VirtualRoom{
				StreamingProvider: &StreamingProvider{
					TotalStreamingSlots: 2,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "have invalid material",
			virtualRoom: &VirtualRoom{
				StreamingProvider: &StreamingProvider{
					StreamingRoomID:     "streaming-room-id",
					TotalStreamingSlots: 2,
				},
				Materials: Materials{
					&VideoMaterial{
						Name:    "video",
						VideoID: "video-id",
					},
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "have invalid PresentMaterialState",
			virtualRoom: &VirtualRoom{
				StreamingProvider: &StreamingProvider{
					StreamingRoomID:     "streaming-room-id",
					TotalStreamingSlots: 2,
				},
				PresentMaterialState: &PDFPresentMaterialState{
					Material: &PDFMaterial{
						Name: "pdf",
						URL:  "https://example.com/random-path/name.pdf",
						ConvertedImageURL: &ConvertedImage{
							Width:    2,
							Height:   3,
							ImageURL: "https://example.com/random-path/name.png",
						},
					},
					UpdatedAt: now,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "have invalid currentPolling",
			virtualRoom: &VirtualRoom{
				StreamingProvider: &StreamingProvider{
					StreamingRoomID:     "streaming-room-id",
					TotalStreamingSlots: 2,
				},
				CurrentPolling: &CurrentPolling{
					Options: CurrentPollingOptions{
						{
							Answer:    "A",
							IsCorrect: false,
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
					Status: CurrentPollingStatusStarted,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "have invalid recording state",
			virtualRoom: &VirtualRoom{
				StreamingProvider: &StreamingProvider{
					StreamingRoomID:     "streaming-room-id",
					TotalStreamingSlots: 2,
				},
				RecordingState: &RecordingState{
					IsRecording: true,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "have invalid attendee state",
			virtualRoom: &VirtualRoom{
				StreamingProvider: &StreamingProvider{
					StreamingRoomID:     "streaming-room-id",
					TotalStreamingSlots: 2,
				},
				AttendeeStates: AttendeeStates{
					{
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
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			userModulePortMock := tc.setup(ctx)
			tc.virtualRoom.userModulePort = userModulePortMock
			err := tc.virtualRoom.isValid(ctx)
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.NoError(t, userModulePortMock.AllFuncBeCalledAsExpected())
		})
	}
}

func TestVirtualRoom_isValidPresentMaterialState(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name        string
		virtualRoom *VirtualRoom
		isValid     bool
	}{
		{
			name: "present material is a video",
			virtualRoom: &VirtualRoom{
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
			},
			isValid: true,
		},
		{
			name: "present material is a pdf",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &PDFPresentMaterialState{
					Material: &PDFMaterial{
						ID:   "media-id-2",
						Name: "pdf",
						URL:  "https://example.com/random-path/name.pdf",
						ConvertedImageURL: &ConvertedImage{
							Width:    2,
							Height:   3,
							ImageURL: "https://example.com/random-path/name.png",
						},
					},
					UpdatedAt: now,
				},
			},
			isValid: true,
		},
		{
			name: "present material is null",
			virtualRoom: &VirtualRoom{
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
			},
			isValid: true,
		},
		{
			name: "present material is not null but list material is null",
			virtualRoom: &VirtualRoom{
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
			},
			isValid: false,
		},
		{
			name: "present material is not exist in list materials",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &VideoPresentMaterialState{
					Material: &VideoMaterial{
						ID:      "media-id-3",
						Name:    "video",
						VideoID: "video-id-1",
					},
					UpdatedAt: now,
					VideoState: &VideoState{
						CurrentTime: Duration(2 * time.Minute),
						PlayerState: PlayerStatePlaying,
					},
				},
			},
			isValid: false,
		},
		{
			name: "present material have id is in list materials but different type",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &PDFPresentMaterialState{
					Material: &PDFMaterial{
						ID:   "media-id-1",
						Name: "pdf",
						URL:  "https://example.com/random-path/name.pdf",
						ConvertedImageURL: &ConvertedImage{
							Width:    2,
							Height:   3,
							ImageURL: "https://example.com/random-path/name.png",
						},
					},
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
		{
			name: "present material is a video but miss video state field",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &VideoPresentMaterialState{
					Material: &VideoMaterial{
						ID:      "media-id-1",
						Name:    "video",
						VideoID: "video-id-1",
					},
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
		{
			name: "present material is a video but miss material field",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &VideoPresentMaterialState{
					UpdatedAt: now,
					VideoState: &VideoState{
						CurrentTime: Duration(2 * time.Minute),
						PlayerState: PlayerStatePlaying,
					},
				},
			},
			isValid: false,
		},
		{
			name: "present material is an invalid video material",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &VideoPresentMaterialState{
					Material: &VideoMaterial{
						ID:   "media-id-1",
						Name: "video",
					},
					UpdatedAt: now,
					VideoState: &VideoState{
						CurrentTime: Duration(2 * time.Minute),
						PlayerState: PlayerStatePlaying,
					},
				},
			},
			isValid: false,
		},
		{
			name: "present material is a pdf but miss material field",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &PDFPresentMaterialState{
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
		{
			name: "present material is an invalid pdf material",
			virtualRoom: &VirtualRoom{
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
				PresentMaterialState: &PDFPresentMaterialState{
					Material: &PDFMaterial{
						ID:   "media-id-2",
						Name: "pdf",
						ConvertedImageURL: &ConvertedImage{
							Width:    2,
							Height:   3,
							ImageURL: "https://example.com/random-path/name.png",
						},
					},
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.virtualRoom.isValidPresentMaterialState()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestVirtualRoom_isValidRecordingState(t *testing.T) {
	t.Parallel()
	creator := "user-id-1"
	tcs := []struct {
		name        string
		virtualRoom *VirtualRoom
		setup       func(ctx context.Context) *mock.UserModulePortMock
		isValid     bool
	}{
		{
			name: "these is a creator is recording state",
			virtualRoom: &VirtualRoom{
				RecordingState: &RecordingState{
					IsRecording: true,
					Creator:     &creator,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				userModulePortMock := &mock.UserModulePortMock{}
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{creator})
					return ids, nil
				}, 1)
				return userModulePortMock
			},
			isValid: true,
		},
		{
			name: "recording state is not recording",
			virtualRoom: &VirtualRoom{
				RecordingState: &RecordingState{
					IsRecording: false,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: true,
		},
		{
			name:        "recording state is null",
			virtualRoom: &VirtualRoom{},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: true,
		},
		{
			name: "these is a creator is recording state but miss creator",
			virtualRoom: &VirtualRoom{
				RecordingState: &RecordingState{
					IsRecording: true,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "creator is not existed",
			virtualRoom: &VirtualRoom{
				RecordingState: &RecordingState{
					IsRecording: true,
					Creator:     &creator,
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				userModulePortMock := &mock.UserModulePortMock{}
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{creator})
					return nil, nil
				}, 1)
				return userModulePortMock
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			userModulePortMock := tc.setup(ctx)
			tc.virtualRoom.userModulePort = userModulePortMock
			err := tc.virtualRoom.isValidRecordingState(ctx)
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.NoError(t, userModulePortMock.AllFuncBeCalledAsExpected())
		})
	}
}

func TestVirtualRoom_isValidAttendeeStates(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name        string
		virtualRoom *VirtualRoom
		setup       func(ctx context.Context) *mock.UserModulePortMock
		isValid     bool
	}{
		{
			name: "attendee state have full fields",
			virtualRoom: &VirtualRoom{
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
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				userModulePortMock := &mock.UserModulePortMock{}
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{"user-id-1", "user-id-2"})
					return ids, nil
				}, 1)
				return userModulePortMock
			},
			isValid: true,
		},
		{
			name:        "attendee state is null",
			virtualRoom: &VirtualRoom{},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: true,
		},
		{
			name: "attendee state have valid data",
			virtualRoom: &VirtualRoom{
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
				AttendeeStates: AttendeeStates{
					{
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
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "attendee state's polling answer is not in current polling",
			virtualRoom: &VirtualRoom{
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
							Answer:    []string{"D"},
							UpdatedAt: now,
						},
					},
				},
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				return &mock.UserModulePortMock{}
			},
			isValid: false,
		},
		{
			name: "attendee state have user id is not existed",
			virtualRoom: &VirtualRoom{
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
			},
			setup: func(ctx context.Context) *mock.UserModulePortMock {
				userModulePortMock := &mock.UserModulePortMock{}
				userModulePortMock.SetCheckExistedUserIDs(func(ctx context.Context, ids []string) (existed []string, err error) {
					assert.ElementsMatch(t, ids, []string{"user-id-1", "user-id-2"})
					return []string{"user-id-1"}, nil
				}, 1)
				return userModulePortMock
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			userModulePortMock := tc.setup(ctx)
			tc.virtualRoom.userModulePort = userModulePortMock
			err := tc.virtualRoom.isValidAttendeeStates(ctx)
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			require.NoError(t, userModulePortMock.AllFuncBeCalledAsExpected())
		})
	}
}
