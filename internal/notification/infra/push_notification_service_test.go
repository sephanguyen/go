package infra

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/infra/metrics"
	mock_firebase "github.com/manabie-com/backend/mock/golibs/firebase"
	mock_metrics "github.com/manabie-com/backend/mock/notification/infra/metrics"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newMockPushNotificationService() (*pushNotificationServiceImpl, *mock_firebase.NotificationPusher, *mock_metrics.NotificationMetrics) {
	notificationPusher := &mock_firebase.NotificationPusher{}
	mockMetric := &mock_metrics.NotificationMetrics{}
	pushNotificationService := &pushNotificationServiceImpl{
		notificationPusher:  notificationPusher,
		NotificationMetrics: mockMetric,
	}
	return pushNotificationService, notificationPusher, mockMetric
}

func newRandomNotiAndMsgs() (*entities.InfoNotification, *entities.InfoNotificationMsg) {
	var notification entities.InfoNotification
	var notificationMsg entities.InfoNotificationMsg
	database.AllRandomEntity(&notification)
	database.AllRandomEntity(&notificationMsg)
	return &notification, &notificationMsg
}

func newRandomUserDeviceTokens(count int) []*entities.UserDeviceToken {
	var res []*entities.UserDeviceToken
	for i := 0; i < count; i++ {
		udt := &entities.UserDeviceToken{}
		database.AllRandomEntity(udt)
		udt.AllowNotification.Set(true)
		res = append(res, udt)
	}
	return res
}

func TestPushNotificationService_PushNotificationForUser(t *testing.T) {
	var dummyFcmError = errors.New("Dummy FCM error")

	var multiFCMSendTokenBatchErr = &firebase.SendTokensError{
		BatchCombinedError: dummyFcmError,
	}

	var multiFCMSendTokenDirectErr = &firebase.SendTokensError{
		DirectError: dummyFcmError,
	}

	t.Parallel()

	t.Run("0 token", func(t *testing.T) {
		pushNotificationService, _, mockMetric := newMockPushNotificationService()
		notification, notificationMsg := newRandomNotiAndMsgs()
		userDeviceTokens := newRandomUserDeviceTokens(0)
		mockMetric.On("RecordPushNotificationErrors", mock.Anything, mock.Anything)
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		assert.NoError(t, err)
		assert.Equal(t, 0, success)
		assert.Equal(t, 0, failure)
	})
	t.Run("fcm success 1 token", func(t *testing.T) {
		pushNotificationService, notificationPusher, mockMetric := newMockPushNotificationService()
		notificationPusher.On("SendToken", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		notification, notificationMsg := newRandomNotiAndMsgs()
		userDeviceTokens := newRandomUserDeviceTokens(1)
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusOK, float64(1))
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		assert.NoError(t, err)
		assert.Equal(t, 1, success)
		assert.Equal(t, 0, failure)
	})
	t.Run("fcm success multiple tokens", func(t *testing.T) {
		pushNotificationService, notificationPusher, mockMetric := newMockPushNotificationService()
		notificationPusher.On("SendTokens", mock.Anything, mock.Anything, mock.Anything).Once().Return(2, 0, nil)
		userDeviceTokens := newRandomUserDeviceTokens(2)
		notification, notificationMsg := newRandomNotiAndMsgs()
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusFail, float64(0))
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusOK, float64(2))
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		assert.NoError(t, err)
		assert.Equal(t, 2, success)
		assert.Equal(t, 0, failure)
	})
	t.Run("fcm partial fails with batch error", func(t *testing.T) {
		pushNotificationService, notificationPusher, mockMetric := newMockPushNotificationService()
		notificationPusher.On("SendTokens", mock.Anything, mock.Anything, mock.Anything).Once().Return(1, 1, multiFCMSendTokenBatchErr)
		userDeviceTokens := newRandomUserDeviceTokens(2)
		notification, notificationMsg := newRandomNotiAndMsgs()
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusFail, float64(1))
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusOK, float64(1))
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		// Only return error in case of full failure
		assert.NoError(t, err, nil)
		assert.Equal(t, 1, success)
		assert.Equal(t, 1, failure)
	})
	t.Run("fcm partial fails with direct error", func(t *testing.T) {
		pushNotificationService, notificationPusher, mockMetric := newMockPushNotificationService()
		notificationPusher.On("SendTokens", mock.Anything, mock.Anything, mock.Anything).Once().Return(1, 1, multiFCMSendTokenDirectErr)
		userDeviceTokens := newRandomUserDeviceTokens(2)
		notification, notificationMsg := newRandomNotiAndMsgs()
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusFail, float64(1))
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusOK, float64(1))
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		// Only return error in case of full failure
		assert.Equal(t, err.Error(), fmt.Sprintf("svc.PushNotificationService.SendTokens - SendMulticast error: %v", multiFCMSendTokenDirectErr.DirectError))
		assert.Equal(t, 1, success)
		assert.Equal(t, 1, failure)
	})
	t.Run("fcm full fails with batch error", func(t *testing.T) {
		pushNotificationService, notificationPusher, mockMetric := newMockPushNotificationService()
		notificationPusher.On("SendTokens", mock.Anything, mock.Anything, mock.Anything).Once().Return(0, 2, multiFCMSendTokenBatchErr)
		userDeviceTokens := newRandomUserDeviceTokens(2)
		notification, notificationMsg := newRandomNotiAndMsgs()
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusOK, float64(0))
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusFail, float64(2))
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		assert.NoError(t, err, nil)
		assert.Equal(t, 0, success)
		assert.Equal(t, 2, failure)
	})
	t.Run("fcm full fails with direct error", func(t *testing.T) {
		pushNotificationService, notificationPusher, mockMetric := newMockPushNotificationService()
		notificationPusher.On("SendTokens", mock.Anything, mock.Anything, mock.Anything).Once().Return(0, 2, multiFCMSendTokenDirectErr)
		userDeviceTokens := newRandomUserDeviceTokens(2)
		notification, notificationMsg := newRandomNotiAndMsgs()
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusOK, float64(0))
		mockMetric.On("RecordPushNotificationErrors", metrics.StatusFail, float64(2))
		success, failure, err := pushNotificationService.PushNotificationForUser(context.Background(), userDeviceTokens, notification, notificationMsg)
		assert.Equal(t, err.Error(), fmt.Sprintf("svc.PushNotificationService.SendTokens - SendMulticast error: %v", multiFCMSendTokenDirectErr.DirectError))
		assert.Equal(t, 0, success)
		assert.Equal(t, 2, failure)
	})
}

func TestPushNotificationService_RetrievePushedMessages(t *testing.T) {
	t.Parallel()

	pushNotificationService, notificationPusher, _ := newMockPushNotificationService()

	t.Run("empty msg", func(t *testing.T) {
		notificationPusher.On("RetrievePushedMessages", context.Background(), mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)
		now := time.Now()
		msg, err := pushNotificationService.RetrievePushedMessages(context.Background(), idutil.ULIDNow(), 1, &types.Timestamp{
			Seconds: int64(now.Second()),
			Nanos:   int32(now.Nanosecond()),
		})
		assert.NoError(t, err)
		assert.Nil(t, msg)
	})
	t.Run("msg", func(t *testing.T) {
		reqMsgs := []*messaging.MulticastMessage{
			{
				Tokens: []string{idutil.ULIDNow()},
				Data:   map[string]string{},
				Notification: &messaging.Notification{
					Title: "title",
					Body:  "body",
				},
			},
		}

		notificationPusher.On("RetrievePushedMessages", context.Background(), mock.Anything, mock.Anything, mock.Anything).Once().Return(reqMsgs, nil)
		now := time.Now()
		msgs, err := pushNotificationService.RetrievePushedMessages(context.Background(), idutil.ULIDNow(), 1, &types.Timestamp{
			Seconds: int64(now.Second()),
			Nanos:   int32(now.Nanosecond()),
		})
		assert.NoError(t, err)

		for index, msg := range msgs {
			assert.Equal(t, msg.Tokens, reqMsgs[index].Tokens)
			assert.Equal(t, msg.Data, reqMsgs[index].Data)
			assert.Equal(t, msg.Title, reqMsgs[index].Notification.Title)
			assert.Equal(t, msg.Body, reqMsgs[index].Notification.Body)
		}
	})
}
