package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const errSubString = "error subscribing to subject"

func TestLessonDeletedSubscription(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	jsm := new(mock_nats.JetStreamManagement)

	s := &LessonDeletedSubscription{
		Logger: ctxzap.Extract(ctx),
		JSM:    jsm,
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
				jsm.On("QueueSubscribe", constants.SubjectLessonDeleted, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

			},
		},
		{
			name:        "failed",
			expectedErr: fmt.Errorf(errSubString),
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", constants.SubjectLessonDeleted, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)
			err := s.Subscribe()
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
