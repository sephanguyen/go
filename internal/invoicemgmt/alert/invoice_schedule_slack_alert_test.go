package alert

import (
	"context"
	"errors"
	"testing"

	mock_alert "github.com/manabie-com/backend/mock/golibs/alert"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type TestCase struct {
	ctx         context.Context
	name        string
	expectedErr error
	config      Config
	setup       func(ctx context.Context)
}

var testError = errors.New("test-error")

func TestInvoiceScheduleSlackAlert_SendSuccessNotification(t *testing.T) {
	t.Parallel()

	mockAlert := &mock_alert.SlackFactory{}
	slackChannel := "test-channel"

	testCases := []TestCase{
		{
			name:        "happy case - local env",
			expectedErr: nil,
			config: Config{
				Environment:  "local",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case - stag env",
			expectedErr: nil,
			config: Config{
				Environment:  "stag",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case - uat env",
			expectedErr: nil,
			config: Config{
				Environment:  "uat",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case - prod env",
			expectedErr: nil,
			config: Config{
				Environment:  "prod",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "negative case - empty slack channel",
			expectedErr: errors.New("slack channel is not provided"),
			config: Config{
				Environment:  "prod",
				SlackChannel: "  ",
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "negative case - error on SendByte",
			expectedErr: testError,
			config: Config{
				Environment:  "prod",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(testError)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			alertManager := NewInvoiceScheduleSlackAlert(mockAlert, testCase.config)
			err := alertManager.SendSuccessNotification()

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockAlert)
		})
	}
}

func TestInvoiceScheduleSlackAlert_SendFailNotification(t *testing.T) {
	t.Parallel()

	mockAlert := &mock_alert.SlackFactory{}
	slackChannel := "test-channel"

	testCases := []TestCase{
		{
			name:        "happy case - local env",
			expectedErr: nil,
			config: Config{
				Environment:  "local",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case - stag env",
			expectedErr: nil,
			config: Config{
				Environment:  "stag",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case - uat env",
			expectedErr: nil,
			config: Config{
				Environment:  "uat",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "happy case - prod env",
			expectedErr: nil,
			config: Config{
				Environment:  "prod",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(nil)
			},
		},
		{
			name:        "negative case - empty slack channel",
			expectedErr: errors.New("slack channel is not provided"),
			config: Config{
				Environment:  "prod",
				SlackChannel: "  ",
			},
			setup: func(ctx context.Context) {
			},
		},
		{
			name:        "negative case - error on SendByte",
			expectedErr: testError,
			config: Config{
				Environment:  "prod",
				SlackChannel: slackChannel,
			},
			setup: func(ctx context.Context) {
				mockAlert.On("SendByte", mock.Anything).Once().Return(testError)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)

			alertManager := NewInvoiceScheduleSlackAlert(mockAlert, testCase.config)
			err := alertManager.SendFailNotification(testError)

			if testCase.expectedErr == nil {
				assert.Equal(t, testCase.expectedErr, err)
			} else {
				assert.Contains(t, err.Error(), testCase.expectedErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockAlert)
		})
	}
}
