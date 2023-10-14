package subscriber

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStudentRegistrationSubscriber_Subscribe(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	subscriber := &StudentRegistrationSubscriber{
		JSM:                        jsm,
		StudentRegistrationService: &service.StudentRegistrationService{},
	}

	testCases := []struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, constants.QueueOrderEventLogCreated, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "subscribe queue order failed",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
			expectedErr: fmt.Errorf("syncOrderSub.Subscribe: %w", fmt.Errorf(errSubString)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctxWithResourcePath := golibs.ResourcePathToCtx(context.Background(), fmt.Sprint(constants.JPREPSchool))
			tc.setup(ctxWithResourcePath)

			err := subscriber.Subscribe()
			assert.Equal(t, tc.expectedErr, err)
		})
	}
}
