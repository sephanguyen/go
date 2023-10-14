package consumers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain/constant"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure/repo"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"gotest.tools/assert"
)

func TestStudentCourseSlotInfoHandler_Handle(t *testing.T) {
	t.Parallel()

	db := &mock_database.Ext{}
	tx := &mock_database.Tx{}
	jsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)

	stateValue := &repo.StateValueDTO{
		BoolValue:        database.Bool(true),
		StringArrayValue: database.TextArray([]string{}),
	}
	lessonID := "lesson-id1"
	teacherID := "user-id1"
	studentID := "student-id1"

	tcs := []struct {
		name     string
		data     *bpb.EvtLesson
		setup    func(ctx context.Context)
		hasError bool
	}{
		{
			name: "successful handler with teacher user",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_JoinLesson_{
					JoinLesson: &bpb.EvtLesson_JoinLesson{
						LessonId:  lessonID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupTeacher]),
						UserId:    teacherID,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("InsertMissingLessonMemberStateByState", mock.Anything, mock.Anything, lessonID, domain.LearnerStateTypeChat, stateValue).Once().
					Return(nil)
			},
			hasError: false,
		},
		{
			name: "failed handler with teacher user",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_JoinLesson_{
					JoinLesson: &bpb.EvtLesson_JoinLesson{
						LessonId:  lessonID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupTeacher]),
						UserId:    teacherID,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("InsertMissingLessonMemberStateByState", mock.Anything, mock.Anything, lessonID, domain.LearnerStateTypeChat, stateValue).Once().
					Return(fmt.Errorf("test error"))
			},
			hasError: true,
		},
		{
			name: "successful handler with student user",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_JoinLesson_{
					JoinLesson: &bpb.EvtLesson_JoinLesson{
						LessonId:  lessonID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupStudent]),
						UserId:    studentID,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("InsertLessonMemberState", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dto := args.Get(2).(*repo.LessonMemberStateDTO)

					assert.Equal(t, lessonID, dto.LessonID.String)
					assert.Equal(t, studentID, dto.UserID.String)
					assert.Equal(t, string(domain.LearnerStateTypeChat), dto.StateType.String)
					assert.Equal(t, true, dto.BoolValue.Bool)
				}).Once().Return(nil)
			},
			hasError: false,
		},
		{
			name: "failed handler with student user",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_JoinLesson_{
					JoinLesson: &bpb.EvtLesson_JoinLesson{
						LessonId:  lessonID,
						UserGroup: cpb.UserGroup(cpb.UserGroup_value[constant.UserGroupStudent]),
						UserId:    studentID,
					},
				},
			},
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonMemberRepo.On("InsertLessonMemberState", mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					dto := args.Get(2).(*repo.LessonMemberStateDTO)

					assert.Equal(t, lessonID, dto.LessonID.String)
					assert.Equal(t, studentID, dto.UserID.String)
					assert.Equal(t, string(domain.LearnerStateTypeChat), dto.StateType.String)
					assert.Equal(t, true, dto.BoolValue.Bool)
				}).Once().Return(fmt.Errorf("test error"))
			},
			hasError: true,
		},
		{
			name: "successful handler but message is unsupported type",
			data: &bpb.EvtLesson{
				Message: &bpb.EvtLesson_CreateLessons_{
					CreateLessons: &bpb.EvtLesson_CreateLessons{},
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

			handler := LessonDefaultChatStateHandler{
				Logger:            ctxzap.Extract(ctx),
				WrapperConnection: wrapperConnection,
				JSM:               jsm,
				LessonMemberRepo:  lessonMemberRepo,
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

			mock.AssertExpectationsForObjects(t, db, tx, lessonMemberRepo, mockUnleashClient)
		})
	}
}
