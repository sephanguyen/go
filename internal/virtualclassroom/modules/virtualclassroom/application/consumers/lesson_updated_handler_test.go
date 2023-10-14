package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLessonUpdateHandler(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := &mock_database.Ext{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	liveLessonSentNotificationRepo := &mock_virtual_repo.MockLiveLessonSentNotificationRepo{}
	lessonID := "lesson-id-1"

	tcs := []struct {
		name     string
		data     []byte
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "happy case",
			data: func() []byte {
				evt := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_UpdateLesson_{
						UpdateLesson: &bpb.EvtLesson_UpdateLesson{
							LessonId:               lessonID,
							StartAtBefore:          timestamppb.New(time.Now().Add(-1 * time.Hour)),
							StartAtAfter:           timestamppb.New(time.Now().Add(1 * time.Hour)),
							SchedulingStatusBefore: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
							SchedulingStatusAfter:  cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						},
					},
				}
				msg, _ := proto.Marshal(evt)
				return msg
			}(),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				liveLessonSentNotificationRepo.On("SoftDeleteLiveLessonSentNotificationRecord", mock.Anything, db, lessonID).Return(nil).Once()
			},
		},
		{
			name: "error",
			data: func() []byte {
				r := &bpb.EvtLesson{
					Message: &bpb.EvtLesson_UpdateLesson_{
						UpdateLesson: &bpb.EvtLesson_UpdateLesson{
							LessonId:               lessonID,
							StartAtBefore:          timestamppb.New(time.Now().Add(-1 * time.Hour)),
							StartAtAfter:           timestamppb.New(time.Now().Add(1 * time.Hour)),
							SchedulingStatusBefore: cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
							SchedulingStatusAfter:  cpb.LessonSchedulingStatus_LESSON_SCHEDULING_STATUS_PUBLISHED,
						},
					},
				}
				b, _ := proto.Marshal(r)
				return b
			}(),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				liveLessonSentNotificationRepo.On("SoftDeleteLiveLessonSentNotificationRecord", mock.Anything, db, lessonID).Return(fmt.Errorf("error")).Once()
			},
			hasError: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			l := &LessonUpdatedHandler{
				Logger:                         ctxzap.Extract(ctx),
				WrapperConnection:              wrapperConnection,
				JSM:                            jsm,
				LiveLessonSentNotificationRepo: liveLessonSentNotificationRepo,
			}

			isLoop, err := l.Handle(ctx, tc.data)
			if tc.hasError {
				require.Error(t, err)
				require.False(t, isLoop)
			} else {
				require.NoError(t, err)
				require.True(t, isLoop)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
