package validation

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/services/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ValidateNotification(t *testing.T) {
	randomNotiAndMsg := func() (*entities.InfoNotification, *entities.InfoNotificationMsg) {
		var noti entities.InfoNotification
		var notiMsg entities.InfoNotificationMsg
		database.AllRandomEntity(&noti)
		database.AllRandomEntity(&notiMsg)
		return &noti, &notiMsg
	}
	t.Run("empty title", func(t *testing.T) {
		noti, notiMsg := randomNotiAndMsg()
		notiMsg.Title.Set("")
		assert.Equal(t, fmt.Errorf("validateNotification.NotificationMessage.Title is empty"), ValidateNotification(notiMsg, noti))
	})
	t.Run("invalid target group json", func(t *testing.T) {
		noti, notiMsg := randomNotiAndMsg()
		noti.TargetGroups.Set("invalid target group json")
		assert.Error(t, ValidateNotification(notiMsg, noti))
	})

}

func Test_ValidateMsgRequiredField(t *testing.T) {
	type tcase struct {
		name  string
		setup func(noti *cpb.Notification)
		err   error
	}
	tcases := []tcase{
		{
			name: "nil msg",
			setup: func(noti *cpb.Notification) {
				noti.Message = nil
			},
			err: fmt.Errorf("request Notification.Message is null"),
		},
		{
			name: "empty title",
			err:  fmt.Errorf("request Notification.Message.Title is empty"),
			setup: func(noti *cpb.Notification) {
				noti.Message.Title = ""
			},
		},
		{
			name: "empty content",
			setup: func(noti *cpb.Notification) {
				noti.Message.Content = makeEmptyContent()
			},
			err: fmt.Errorf("request Notification.Message.Content is null"),
		},
	}
	for _, tcas := range tcases {
		t.Run(tcas.name, func(t *testing.T) {
			noti := &cpb.Notification{}
			assert.NoError(t, faker.FakeData(noti))
			tcas.setup(noti)
			err := ValidateMessageRequiredField(noti)
			assert.Equal(t, tcas.err, err)
		})
	}
}

func Test_ValidateUpsertNotificationRequest(t *testing.T) {
	type testCase struct {
		Name    string
		Req     *npb.UpsertNotificationRequest
		RespErr error
		Setup   func(ctx context.Context, this *testCase)
	}
	testCases := []*testCase{
		{
			Name: "happy case",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: nil,
			Setup: func(ctx context.Context, this *testCase) {
			},
		}, {
			Name: "missing notification",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("request notification is null"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification = nil
			},
		},
		{
			Name: "missing message",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("request Notification.Message is null"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Message = nil
			},
		},
		{
			Name: "missing title",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("request Notification.Message.Title is empty"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Message.Title = ""
			},
		},
		{
			Name: "missing content",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("request Notification.Message.Content is null"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Message.Content = nil
			},
		},
		{
			Name: "error when notification status is none",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("do not allow req notication status is %v", cpb.NotificationStatus_NOTIFICATION_STATUS_NONE),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_NONE
			},
		},
		{
			Name: "error when notification status is sent",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("do not allow req notication status is %v", cpb.NotificationStatus_NOTIFICATION_STATUS_SENT),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SENT
			},
		},
		{
			Name: "error when notification status is discard",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("do not allow req notication status is %v", cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_DISCARD
			},
		},
	}
	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx, tc)
		err := ValidateUpsertNotificationRequest(tc.Req)
		assert.Equal(t, tc.RespErr, err)
	}
}

func Test_ValidateScheduledNotification(t *testing.T) {
	type testCase struct {
		Name    string
		Req     *npb.UpsertNotificationRequest
		RespErr error
		Setup   func(ctx context.Context, this *testCase)
	}
	testCases := []*testCase{
		{
			Name: "happy case",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: nil,
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
				this.Req.Notification.ScheduledAt = timestamppb.New(time.Now().Add(3 * time.Minute))
			},
		},
		{
			Name: "error when notification scheduled at is before current time",
			Req: &npb.UpsertNotificationRequest{
				Notification: utils.GenSampleNotification(),
			},
			RespErr: fmt.Errorf("you cannot schedule at a time in the past"),
			Setup: func(ctx context.Context, this *testCase) {
				this.Req.Notification.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
				this.Req.Notification.ScheduledAt = timestamppb.New(time.Now().Add(-1 * time.Minute))
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		tc.Setup(ctx, tc)
		err := ValidateScheduledNotification(tc.Req)
		assert.Equal(t, tc.RespErr, err)
	}
}

func makeEmptyContent() *cpb.RichText {
	if rand.Intn(2) == 0 {
		return nil
	}
	return &cpb.RichText{}
}
