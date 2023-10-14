package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/consumers"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/controller"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mock_liveroom_repo "github.com/manabie-com/backend/mock/virtualclassroom/liveroom/repositories"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLiveRoomSubscriber_Subscribe(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	db := &mock_database.Ext{}
	jsm := &mock_nats.JetStreamManagement{}
	liveRoomMemberRepo := &mock_liveroom_repo.MockLiveRoomMemberStateRepo{}

	handler := &consumers.LiveRoomHandler{
		Logger:                  ctxzap.Extract(ctx),
		LessonmgmtDB:            db,
		JSM:                     jsm,
		LiveRoomMemberStateRepo: liveRoomMemberRepo,
	}

	subscriber := &controller.LiveRoomSubscriber{
		Logger:            ctxzap.Extract(ctx),
		JSM:               jsm,
		SubscriberHandler: handler,
	}

	tcs := []struct {
		name        string
		expectedErr error
		setup       func(ctx context.Context)
	}{
		{
			name: "successful subscribe",
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, nil)

			},
		},
		{
			name:        "failed subscribe",
			expectedErr: errors.New("error"),
			setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, errors.New("error"))
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctxRP := golibs.ResourcePathToCtx(context.Background(), "school-id")
			tc.setup(ctxRP)

			err := subscriber.Subscribe()
			if tc.expectedErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
