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

const errSubString = "error subscribing to subject"

func TestUserRegistrationSubscriber_Subscribe(t *testing.T) {
	t.Parallel()
	_, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	jsm := new(mock_nats.JetStreamManagement)
	subscriber := &UserRegistrationSubscriber{
		JSM:                     jsm,
		UserRegistrationService: &service.UserRegistrationService{},
	}

	testCases := []struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, constants.QueueSyncStudent, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("QueueSubscribe", mock.Anything, constants.QueueSyncStaff, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
			expectedErr: nil,
		},
		{
			name: "subscribe queue student failed",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
			expectedErr: fmt.Errorf("syncStudentSub.Subscribe: %w", fmt.Errorf(errSubString)),
		},
		{
			name: "subscribe queue staff failed",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf(errSubString))
			},
			expectedErr: fmt.Errorf("syncStaffSub.Subscribe: %w", fmt.Errorf(errSubString)),
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
