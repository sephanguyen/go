package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/consumers"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/lesson/repositories"
	mock_repo "github.com/manabie-com/backend/mock/mastermgmt/modules/class/infrastructure/repo"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentChangeClassSubscriber_subscribeStudentChangeClass(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	lessonRepo := new(mock_repositories.MockLessonRepo)
	db := &mock_database.Ext{}
	lessonMemberRepo := new(mock_repositories.MockLessonMemberRepo)
	classMemberRepo := new(mock_repo.MockClassMemberRepo)

	subscriberHandler := &consumers.StudentChangeClassHandler{
		Logger:           ctxzap.Extract(ctx),
		DB:               db,
		JSM:              jsm,
		LessonRepo:       lessonRepo,
		LessonMemberRepo: lessonMemberRepo,
		ClassMemberRepo:  classMemberRepo,
	}
	s := &StudentChangeClassSubscriber{
		Logger:            ctxzap.Extract(ctx),
		JSM:               jsm,
		SubscriberHandler: subscriberHandler,
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

			err := s.subscribeStudentChangeClass()
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
