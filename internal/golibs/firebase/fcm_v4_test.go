package firebase

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/multierr"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTenantFCMClient_SendToken(t *testing.T) {
	t.Parallel()
	fcmClient := new(MockFCMClientV4)

	s := &notificationPusherImpl{
		Client: fcmClient,
	}

	token := "fcm-token"

	msg := &messaging.Message{}
	msg.Token = token

	fcmClient.On("Send", mock.Anything, msg).Once().Return("", nil)
	s.SendToken(context.Background(), msg, token)

	mock.AssertExpectationsForObjects(t, fcmClient)
	fcmClient.AssertCalled(t, "Send", mock.Anything, msg)
}

func TestTenantSendTokens(t *testing.T) {
	t.Parallel()
	fcmClient := new(MockFCMClientV4)

	s := &notificationPusherImpl{
		Client: fcmClient,
	}

	tokens := []string{"fcm-token-1", "fcm-token-2"}
	notFoundError := fmt.Errorf("token not found")

	msg := &messaging.MulticastMessage{}
	msg.Tokens = tokens

	ctx := context.Background()

	t.Run("success send 2 tokens", func(t *testing.T) {
		resp := &messaging.BatchResponse{
			SuccessCount: 2,
			FailureCount: 0,
			Responses: []*messaging.SendResponse{
				{
					Success: true,
					Error:   nil,
				},
				{
					Success: true,
					Error:   nil,
				},
			},
		}
		fcmClient.On("SendMulticast", ctx, msg).Once().Return(resp, nil)
		successCount, failureCount, err := s.SendTokens(context.Background(), msg, tokens)

		mock.AssertExpectationsForObjects(t, fcmClient)
		fcmClient.AssertCalled(t, "SendMulticast", ctx, msg)

		assert.Equal(t, 2, successCount)
		assert.Equal(t, 0, failureCount)
		assert.Nil(t, err)
	})

	t.Run("internal error", func(t *testing.T) {
		internalFCMError := fmt.Errorf("internal error")
		fcmClient.On("SendMulticast", ctx, msg).Once().Return(nil, internalFCMError)
		successCount, failureCount, err := s.SendTokens(context.Background(), msg, tokens)

		mock.AssertExpectationsForObjects(t, fcmClient)
		fcmClient.AssertCalled(t, "SendMulticast", ctx, msg)

		assert.Equal(t, 0, successCount)
		assert.Equal(t, 2, failureCount)
		assert.Nil(t, err.BatchCombinedError)
		assert.ErrorIs(t, err.DirectError, internalFCMError)
	})

	t.Run("batch error, full failed", func(t *testing.T) {
		resp := &messaging.BatchResponse{
			SuccessCount: 0,
			FailureCount: 2,
			Responses: []*messaging.SendResponse{
				{
					Success: false,
					Error:   notFoundError,
				},
				{
					Success: false,
					Error:   notFoundError,
				},
			},
		}
		fcmClient.On("SendMulticast", ctx, msg).Once().Return(resp, nil)
		successCount, failureCount, err := s.SendTokens(context.Background(), msg, tokens)

		mock.AssertExpectationsForObjects(t, fcmClient)
		fcmClient.AssertCalled(t, "SendMulticast", ctx, msg)

		assert.Equal(t, 0, successCount)
		assert.Equal(t, 2, failureCount)
		assert.Nil(t, err.DirectError)

		expectedBatchErr := multierr.Combine(
			notFoundError,
			notFoundError,
		)
		assert.Equal(t, expectedBatchErr.Error(), err.BatchCombinedError.Error())
	})

	t.Run("batch error, partial failed", func(t *testing.T) {
		resp := &messaging.BatchResponse{
			SuccessCount: 1,
			FailureCount: 1,
			Responses: []*messaging.SendResponse{
				{
					Success: false,
					Error:   notFoundError,
				},
				{
					Success: true,
					Error:   nil,
				},
			},
		}
		fcmClient.On("SendMulticast", ctx, msg).Once().Return(resp, nil)
		successCount, failureCount, err := s.SendTokens(context.Background(), msg, tokens)

		mock.AssertExpectationsForObjects(t, fcmClient)
		fcmClient.AssertCalled(t, "SendMulticast", ctx, msg)

		assert.Equal(t, 1, successCount)
		assert.Equal(t, 1, failureCount)
		assert.Nil(t, err.DirectError)

		expectedBatchErr := multierr.Combine(
			notFoundError,
		)
		assert.Equal(t, expectedBatchErr.Error(), err.BatchCombinedError.Error())
	})
}
