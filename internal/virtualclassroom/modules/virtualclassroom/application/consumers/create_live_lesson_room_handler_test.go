package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_whiteboard "github.com/manabie-com/backend/mock/golibs/whiteboard"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestCreateLiveLessonRoomHandler_Handle(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonRepo := new(mock_repositories.MockVirtualLessonRepo)
	whiteboardSvc := new(mock_whiteboard.MockService)

	lessonID1 := "lesson-id1"
	lessonID2 := "lesson-id1"
	roomID1 := "room-id1"
	roomID2 := "room-id1"

	tcs := []struct {
		name     string
		data     *bpb.EvtLesson
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "successful handler",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_CreateLessons_{
					CreateLessons: &bpb.EvtLesson_CreateLessons{
						Lessons: []*bpb.EvtLesson_Lesson{
							{
								LessonId: lessonID1,
							},
							{
								LessonId: lessonID2,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				whiteboardSvc.On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     lessonID1,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{UUID: roomID1}, nil)

				lessonRepo.On("UpdateRoomID", mock.Anything, mock.Anything, lessonID1, roomID1).Once().
					Return(nil)

				whiteboardSvc.On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     lessonID2,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{UUID: roomID2}, nil)

				lessonRepo.On("UpdateRoomID", mock.Anything, mock.Anything, lessonID2, roomID2).Once().
					Return(nil)
			},
			hasError: false,
		},
		{
			name: "no rows affected",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_CreateLessons_{
					CreateLessons: &bpb.EvtLesson_CreateLessons{
						Lessons: []*bpb.EvtLesson_Lesson{
							{
								LessonId: lessonID1,
							},
							{
								LessonId: lessonID2,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				whiteboardSvc.On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     lessonID1,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{UUID: roomID1}, nil)

				lessonRepo.On("UpdateRoomID", mock.Anything, mock.Anything, lessonID1, roomID1).Once().
					Return(nil)

				whiteboardSvc.On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     lessonID2,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{UUID: roomID2}, nil)

				lessonRepo.On("UpdateRoomID", mock.Anything, mock.Anything, lessonID2, roomID2).Once().
					Return(fmt.Errorf("cannot update lesson %s room ID", lessonID2))
			},
			hasError: false,
		},
		{
			name: "failed handler",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_CreateLessons_{
					CreateLessons: &bpb.EvtLesson_CreateLessons{
						Lessons: []*bpb.EvtLesson_Lesson{
							{
								LessonId: lessonID1,
							},
							{
								LessonId: lessonID2,
							},
						},
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				whiteboardSvc.On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     lessonID1,
					IsRecord: false,
				}).Once().Return(nil, fmt.Errorf("error"))

				whiteboardSvc.On("CreateRoom", mock.Anything, &whiteboard.CreateRoomRequest{
					Name:     lessonID2,
					IsRecord: false,
				}).Once().Return(&whiteboard.CreateRoomResponse{UUID: roomID2}, nil)

				lessonRepo.On("UpdateRoomID", mock.Anything, mock.Anything, lessonID2, roomID2).Once().
					Return(fmt.Errorf("error"))
			},
			hasError: true,
		},
		{
			name: "successful handler but message is unsupported type",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_JoinLesson_{
					JoinLesson: &bpb.EvtLesson_JoinLesson{},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
			},
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tc.setup(ctx)

			handler := CreateLiveLessonRoomHandler{
				Logger:            ctxzap.Extract(ctx),
				WrapperConnection: wrapperConnection,
				JSM:               jsm,
				LessonRepo:        lessonRepo,
				WhiteboardService: whiteboardSvc,
			}

			msgEvnt, _ := proto.Marshal(tc.data)
			res, err := handler.Handle(ctx, msgEvnt)
			if tc.hasError {
				require.Error(t, err)
				require.False(t, res)
			} else {
				require.NoError(t, err)
				require.True(t, res)
			}

			mock.AssertExpectationsForObjects(t, db, lessonRepo, mockUnleashClient)
		})
	}
}
