package chat

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mocks "github.com/manabie-com/backend/mock/tom/pb"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestHub_RemoveStallConn(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	s := &Server{
		hostName: idutil.ULIDNow(),
		logger:   zap.NewNop(),
	}
	hub := s.NewHub()
	hub.Start()
	stream := new(mocks.ChatService_SubscribeV2Server)
	stream.On("Context").Return(ctx)
	blockingConn := s.newClientConnV2(stream, "test_user", "")
	healthyConn := s.newClientConnV2(stream, "test_user", "")

	// We want this channel to be unbuffered,
	// if hub's loop sends msg into it, it will block
	blockingConn.Send = make(chan *pb.Event)
	hub.register <- blockingConn
	hub.register <- healthyConn
	require.True(t, checkCondition(context.Background(), func() bool {
		return atomic.LoadInt64(&hub.connectionCount) == 2
	}))
	testMsg := &Event{
		UserID: "test_user",
	}
	hub.broadcast <- testMsg
	require.True(t, checkCondition(ctx, func() bool {
		return atomic.LoadInt64(&hub.connectionCount) == 1
	}))
	// sleep a bit to make sure healthyConn is not removed
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int64(1), atomic.LoadInt64(&hub.connectionCount))
}

func TestHub_PushMessageDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	jsm := &mock_nats.JetStreamManagement{}

	onlineUserRepo := new(mock_repositories.MockOnlineUserRepo)
	hostname := idutil.ULIDNow()
	s := &Server{
		hostName:       hostname,
		logger:         zap.NewNop(),
		JSM:            jsm,
		onlineUserRepo: onlineUserRepo,
	}

	userID := idutil.ULIDNow()
	// userIDS := []string{userID}
	msg := &pb.Event{}
	ctx = interceptors.ContextWithUserID(ctx, userID)

	pushMsgOpts := domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			Enabled:      true,
			IgnoredUsers: []string{userID},
		},
	}

	// nodeID := idutil.ULIDNow()
	userIDs := []string{
		idutil.ULIDNow(),
	}
	nodeID := idutil.ULIDNow()
	var (
		hID pgtype.Text
		nID pgtype.Text
		uID pgtype.Text
	)
	_ = hID.Set(hostname)
	_ = nID.Set(nodeID)
	_ = uID.Set(userID)

	testCases := []TestCase{
		{
			name:         "one node only",
			ctx:          ctx,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				onlineUserRepo.On("Find", ctx, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text][]string{hID: userIDs}, nil)
				jsm.On("PublishContext", ctx, mock.Anything, mock.Anything).Once().Return(nil, nil)
			},
		},
		{
			name:         "multiple nodes",
			ctx:          ctx,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				onlineUserRepo.On("Find", ctx, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text][]string{nID: userIDs, hID: userIDs}, nil)
				jsm.On("PublishContext", ctx, mock.Anything, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
		{
			name:         "multiple nodes but failed in publishContext",
			ctx:          ctx,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				onlineUserRepo.On("Find", ctx, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text][]string{nID: userIDs, hID: userIDs}, nil)
				jsm.On("PublishContext", ctx, mock.Anything, mock.Anything).Once().Return(nil, errors.New("publish error"))
			},
		},
		{
			name:         "error db online user",
			ctx:          ctx,
			expectedResp: nil,
			expectedErr:  nil,
			setup: func(ctx context.Context) {
				onlineUserRepo.On("Find", ctx, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(map[pgtype.Text][]string{nID: userIDs}, pgx.ErrTxClosed)
				jsm.On("PublishContext", ctx, mock.Anything, mock.Anything).Once().Return(&nats.PubAck{}, nil)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			resp := s.PushMessageDeleted(testCase.ctx, userIDs, msg, pushMsgOpts)
			if testCase.expectedResp == nil {
				assert.Nil(t, testCase.expectedResp, resp)
			} else {
				assert.Equal(t, testCase.expectedResp, resp)
			}
		})
	}
}
