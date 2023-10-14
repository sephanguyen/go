package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/consumers"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_repositories "github.com/manabie-com/backend/mock/lessonmgmt/user/repositories"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentCourseSlotInfoSubscriber_Subscribe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	db := &mock_database.Ext{}
	jsm := &mock_nats.JetStreamManagement{}
	userRepo := new(mock_repositories.MockUserRepo)
	studentSubRepo := new(mock_repositories.MockStudentSubscriptionRepo)
	studentSubAccessPathRepo := new(mock_repositories.MockStudentSubscriptionAccessPathRepo)

	handler := &consumers.StudentCourseSlotInfoHandler{
		Logger:                            ctxzap.Extract(ctx),
		DB:                                db,
		JSM:                               jsm,
		UserRepo:                          userRepo,
		StudentSubscriptionRepo:           studentSubRepo,
		StudentSubscriptionAccessPathRepo: studentSubAccessPathRepo,
	}

	s := &StudentCourseSlotInfoSubscriber{
		Logger:            ctxzap.Extract(ctx),
		JSM:               jsm,
		SubscriberHandler: handler,
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
			expectedErr: errors.New("error"),
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("error"))
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
