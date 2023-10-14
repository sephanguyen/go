package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

func TestLockLessonSubscription_Subscribe(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	lessonRepo := new(mock_repositories.MockLessonRepo)

	s := &LockLessonSubscription{
		JSM:        jsm,
		LessonRepo: lessonRepo,
		Logger:     zap.NewNop(),
	}

	type TestCase struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}

	testCases := []TestCase{
		{
			name: "success",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:        "failed",
			expectedErr: fmt.Errorf(errSubString),
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxRP := golibs.ResourcePathToCtx(context.Background(), "school-id")
			tc.setup(ctxRP)

			err := s.Subscribe()
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLockLessonSubscription_handleLockLessonEvent(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")

	s := &LockLessonSubscription{
		JSM:               jsm,
		LessonRepo:        lessonRepo,
		Logger:            zap.NewNop(),
		wrapperConnection: wrapperConnection,
	}
	lessonIds := []string{"lesson-1", "lesson-2", "lesson-3"}
	lessonEvent := &pb.TimesheetLessonLockEvt{
		LessonIds: lessonIds,
	}
	msg, err := proto.Marshal(lessonEvent)
	assert.NoError(t, err)

	type TestCase struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}

	testCases := []TestCase{
		{
			name: "success",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("LockLesson", mock.Anything, mock.Anything, lessonIds).
					Once().Return(nil)

			},
		},
		{
			name:        "failed",
			expectedErr: fmt.Errorf("error when update lock lesson"),
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				lessonRepo.On("LockLesson", mock.Anything, mock.Anything, lessonIds).
					Once().Return(fmt.Errorf("error when update lock lesson"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxRP := golibs.ResourcePathToCtx(context.Background(), "school-id")
			tc.setup(ctxRP)

			_, err := s.handleLockLessonEvent(ctx, msg)
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}
}
