package nats

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTimesheetActionLogEventSubscription_Subscribe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mockJsm := &mock_nats.JetStreamManagement{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	s := &TimesheetActionLogEventSubscription{
		Logger:        nil,
		JSM:           mockJsm,
		UnleashClient: mockUnleashClient,
		Env:           "local",
	}

	tests := []struct {
		name    string
		wantErr bool
		ctx     context.Context
		setup   func(ctx context.Context)
	}{
		{
			name:    "Feature Flag for Auto Create is enabled",
			wantErr: false,
			ctx:     ctx,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(true, nil)
				mockJsm.On("QueueSubscribe", constants.SubjectTimesheetActionLog, constants.QueueTimesheetActionLog, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:    "Feature Flag for Auto Create is disabled",
			wantErr: false,
			ctx:     ctx,
			setup: func(ctx context.Context) {
				mockUnleashClient.On("IsFeatureEnabled", mock.Anything, mock.Anything).Once().Return(false, nil)
			},
		},
	}

	// run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(tt.ctx)

			err := s.Subscribe()
			assert.NoError(t, err)
			mock.AssertExpectationsForObjects(t, mockUnleashClient, mockJsm)
		})
	}

}
