package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	vc_domain "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_liveroom_repo "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	vpb "github.com/manabie-com/backend/pkg/manabuf/virtualclassroom/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStudentCourseSlotInfoHandler_Handle(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	liveRoomMemberRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}

	stateValue := &vc_domain.StateValue{
		BoolValue: true,
	}
	stateTypeChat := vc_domain.LearnerStateTypeChat
	stateTypeAnnotation := vc_domain.LearnerStateTypeAnnotation

	channelID := "channel-id1"
	teacherID := "user-id1"
	studentID := "student-id1"

	tcs := []struct {
		name     string
		data     *vpb.LiveRoomEvent
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "successful handler with student user",
			data: &vpb.LiveRoomEvent{
				Message: &vpb.LiveRoomEvent_JoinLiveRoom_{
					JoinLiveRoom: &vpb.LiveRoomEvent_JoinLiveRoom{
						ChannelId: channelID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupStudent]),
						UserId:    studentID,
					},
				},
			},
			setup: func(ctx context.Context) {
				liveRoomMemberRepo.On("CreateLiveRoomMemberState", mock.Anything, mock.Anything, channelID, studentID, stateTypeChat, stateValue).Once().
					Return(nil)

				liveRoomMemberRepo.On("CreateLiveRoomMemberState", mock.Anything, mock.Anything, channelID, studentID, stateTypeAnnotation, stateValue).Once().
					Return(nil)
			},
			hasError: false,
		},
		{
			name: "failed handler with student user",
			data: &vpb.LiveRoomEvent{
				Message: &vpb.LiveRoomEvent_JoinLiveRoom_{
					JoinLiveRoom: &vpb.LiveRoomEvent_JoinLiveRoom{
						ChannelId: channelID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupStudent]),
						UserId:    studentID,
					},
				},
			},
			setup: func(ctx context.Context) {
				liveRoomMemberRepo.On("CreateLiveRoomMemberState", mock.Anything, mock.Anything, channelID, studentID, stateTypeChat, stateValue).Once().
					Return(fmt.Errorf("test error"))
			},
			hasError: true,
		},
		{
			name: "successful handler with teacher user",
			data: &vpb.LiveRoomEvent{
				Message: &vpb.LiveRoomEvent_JoinLiveRoom_{
					JoinLiveRoom: &vpb.LiveRoomEvent_JoinLiveRoom{
						ChannelId: channelID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupTeacher]),
						UserId:    teacherID,
					},
				},
			},
			setup:    func(ctx context.Context) {},
			hasError: false,
		},
		{
			name: "successful handler but message is unsupported type",
			data: &vpb.LiveRoomEvent{
				Message: &vpb.LiveRoomEvent_EndLiveRoom_{},
			},
			setup:    func(ctx context.Context) {},
			hasError: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tc.setup(ctx)

			handler := LiveRoomHandler{
				Logger:                  ctxzap.Extract(ctx),
				LessonmgmtDB:            db,
				JSM:                     jsm,
				LiveRoomMemberStateRepo: liveRoomMemberRepo,
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

			mock.AssertExpectationsForObjects(t, db, tx, liveRoomMemberRepo)
		})
	}
}
