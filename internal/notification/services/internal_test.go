package services

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/notification/services/utils"
	mock_infra "github.com/manabie-com/backend/mock/notification/infra"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInternal_RetrievePushedNotificationMessage(t *testing.T) {
	t.Parallel()
	pushSvc := &mock_infra.PushNotificationService{}

	svc := &InternalService{
		PushNotificationService: pushSvc,
	}
	mockMsges := []utils.RetrievedPushNotificationMsg{
		{
			Data:  map[string]string{"custom_key": "value"},
			Title: "fake title",
			Body:  "fake body",
		},
	}
	pushSvc.On("RetrievePushedMessages", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockMsges, nil)
	ret, err := svc.RetrievePushedNotificationMessages(context.Background(), &npb.RetrievePushedNotificationMessageRequest{
		Since: timestamppb.Now(),
	})
	assert.NoError(t, err)
	assert.Equal(t, "fake title", ret.Messages[0].Title)
	assert.Equal(t, "fake body", ret.Messages[0].Body)
	customField := ret.Messages[0].Data.Fields["custom_key"]
	assert.Equal(t, structpb.NewStringValue("value"), customField)
}
