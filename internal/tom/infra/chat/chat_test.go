package chat

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"
	mocks "github.com/manabie-com/backend/mock/tom/pb"
	mock_repositories "github.com/manabie-com/backend/mock/tom/repositories"
	mock_services "github.com/manabie-com/backend/mock/tom/services"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgtype"
	natsJS "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type TestCase struct {
	name         string
	ctx          context.Context
	req          interface{}
	customCtx    func(context.Context) context.Context
	expectedResp interface{}
	expectedErr  error
	setup        func(ctx context.Context)
}

func customStudentCtx(ctxt context.Context) context.Context {
	claims := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			DefaultRole: cpb.UserGroup_USER_GROUP_STUDENT.String(),
		},
	}
	return interceptors.ContextWithJWTClaims(ctxt, claims)
}

func mockChatServerCtx(stream *mocks.ChatService_SubscribeV2Server, commonUserCtx context.Context, userID string, rp string) (context.Context, context.CancelFunc) {
	userCtx, cancel := context.WithCancel(interceptors.ContextWithUserID(commonUserCtx, userID))

	ctxWithRP := interceptors.ContextWithJWTClaims(userCtx, &interceptors.CustomClaims{Manabie: &interceptors.ManabieClaims{ResourcePath: rp}})
	stream.On("Context").Times(3).Return(ctxWithRP)
	return ctxWithRP, cancel
}

func TestChatService_RemoveConnectionOnError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	onlineUserRepo := &mock_repositories.MockOnlineUserRepo{}

	s := &Server{
		hostName:       idutil.ULIDNow(),
		logger:         zap.NewNop(),
		onlineUserRepo: onlineUserRepo,
	}
	s.HubStart()
	defer s.HubStop()

	totalConnection := 100
	type user struct {
		ctx    context.Context
		cancel context.CancelFunc
		userID string
		srv    *mocks.ChatService_SubscribeV2Server
	}

	onlineUserRepo.On("Insert", mock.Anything, mock.Anything, mock.AnythingOfType("*core.OnlineUser")).
		Times(100).Return(fmt.Errorf("dummy"))
	onlineUserRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
	onlineUserRepo.AssertNotCalled(t, "SetActive", mock.Anything, mock.Anything, mock.Anything)

	commonUserCtx, commonUserCancel := context.WithCancel(ctx)
	defer commonUserCancel()
	users := make([]*user, 0, totalConnection)
	commonResourcePath := "1"
	for i := 0; i < totalConnection; i++ {
		stream := new(mocks.ChatService_SubscribeV2Server)
		userID := idutil.ULIDNow()
		ctx, cancel := mockChatServerCtx(stream, commonUserCtx, userID, commonResourcePath)

		stream.AssertNotCalled(t, "Send", mock.Anything)

		users = append(users, &user{
			ctx:    ctx,
			userID: userID,
			srv:    stream,
			cancel: cancel,
		})
	}

	var wg sync.WaitGroup
	for _, e := range users {
		wg.Add(1)
		go func(u *user) {
			defer wg.Done()
			err := s.SubscribeV2(&pb.SubscribeV2Request{}, u.srv)
			assert.EqualError(t, err, "rpc error: code = Internal desc = dummy")
			u.srv.AssertNumberOfCalls(t, "Context", 3)
		}(e)
	}

	// here, to wait for all goroutines to spawn, we periodically check the number of connections
	// the timeout is determined by the context

	wg.Wait()
	success := checkCondition(ctx, func() bool { return s.TotalConnections() == 0 })
	if !success {
		for _, h := range s.Hubs {
			h.stop <- struct{}{}
		}
	}
	assert.True(t, success, "want no connection, but has %d", s.TotalConnections())
}

func TestChatService_SubscribeV2WithMultipleConnection(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	onlineUserRepo := &mock_repositories.MockOnlineUserRepo{}

	s := &Server{
		hostName:       idutil.ULIDNow(),
		logger:         zap.NewNop(),
		onlineUserRepo: onlineUserRepo,
	}
	s.HubStart()
	defer s.HubStop()

	totalConnection := 100
	type user struct {
		ctx    context.Context
		cancel context.CancelFunc
		userID string
		srv    *mocks.ChatService_SubscribeV2Server
	}

	commonResourcePath := "1"

	onlineUserRepo.On("Insert", mock.Anything, mock.Anything, mock.AnythingOfType("*core.OnlineUser")).
		Times(100).Return(nil)

	onlineUserRepo.On("Delete", mock.MatchedBy(func(ctx context.Context) bool {
		rpinctx, err := interceptors.ResourcePathFromContext(ctx)
		if err != nil {
			return false
		}
		return rpinctx == commonResourcePath
	}), mock.Anything, mock.AnythingOfType("pgtype.Text")).
		Times(100).Return(nil)

	onlineUserRepo.On("SetActive", mock.Anything, mock.Anything, mock.AnythingOfType("pgtype.Text")).
		Times(200).Return(nil)

	commonUserCtx, commonUserCancel := context.WithCancel(ctx)
	defer commonUserCancel()
	users := make([]*user, 0, totalConnection)
	for i := 0; i < totalConnection; i++ {
		userID := idutil.ULIDNow()
		stream := new(mocks.ChatService_SubscribeV2Server)
		ctx, cancel := mockChatServerCtx(stream, commonUserCtx, userID, commonResourcePath)

		// send ping event to client for the first time
		stream.On("Send", mock.MatchedBy(func(evt *pb.SubscribeV2Response) bool {
			return evt.GetEvent().GetEventPing() != nil
		})).Once().Return(nil)

		users = append(users, &user{
			ctx:    ctx,
			userID: userID,
			srv:    stream,
			cancel: cancel,
		})
	}

	var wg sync.WaitGroup
	for _, e := range users {
		wg.Add(1)
		go func(u *user) {
			defer wg.Done()
			err := s.SubscribeV2(&pb.SubscribeV2Request{}, u.srv)
			assert.EqualError(t, err, "ctx.Done: context canceled")
			u.srv.AssertNumberOfCalls(t, "Context", 3)
		}(e)
	}

	// here, to wait for all goroutines to spawn, we periodically check the number of connections
	// the timeout is determined by the context
	success := checkCondition(ctx, func() bool { return s.TotalConnections() == totalConnection })
	require.True(t, success)
	time.Sleep(time.Millisecond * 50) // another sleep to check if the number of connections still stays the same after a bit
	assert.Equal(t, totalConnection, s.TotalConnections())

	// cancel the main context to cancel all the spawned goroutines
	commonUserCancel()
	wg.Wait()

	onlineUserRepo.AssertNumberOfCalls(t, "Insert", totalConnection)
	onlineUserRepo.AssertNumberOfCalls(t, "Delete", totalConnection)
	success = checkCondition(ctx, func() bool { return s.TotalConnections() == 0 })
	require.True(t, success)
}

func checkCondition(ctx context.Context, f func() bool) (success bool) {
	tmr := time.NewTimer(time.Second * 30) // to prevent running forever
	defer tmr.Stop()
	tck := time.NewTicker(time.Millisecond * 50)
	defer tck.Stop()
	for {
		select {
		case <-tck.C:
			if f() {
				return true
			}
		case <-ctx.Done():
			return false
		case <-tmr.C:
			return false
		}
	}
}

// - Create a chat socket
func TestChatService_DenyMsgWithDifferentResourcePath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	onlineUserRepo := &mock_repositories.MockOnlineUserRepo{}
	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	notiPusher := &mock_services.Pusher{}

	jsm := &mock_nats.JetStreamManagement{}
	s := &Server{
		hostName:            idutil.ULIDNow(),
		JSM:                 jsm,
		logger:              zap.NewNop(),
		onlineUserRepo:      onlineUserRepo,
		userDeviceTokenRepo: userDeviceTokenRepo,
		notification: &Notification{
			Pusher: notiPusher,
		},
	}
	s.HubStart()
	defer s.HubStop()

	totalConnection := 100
	type user struct {
		ctx    context.Context
		cancel context.CancelFunc
		userID string
		srv    *mocks.ChatService_SubscribeV2Server
	}

	userIDs := make([]string, 0, totalConnection)
	for i := 0; i < totalConnection; i++ {
		userID := idutil.ULIDNow()
		userIDs = append(userIDs, userID)
	}

	onlineUserRepo.On("Insert", mock.Anything, mock.Anything, mock.AnythingOfType("*core.OnlineUser")).
		Times(100).Return(nil)

	onlineUserRepo.On("Delete", mock.Anything, mock.Anything, mock.AnythingOfType("pgtype.Text")).
		Times(100).Return(nil)

	// validUserIDs := userIDs[:len(userIDs)/2]
	invalidUserIDs := userIDs[len(userIDs)/2:]
	nodeOnlineUsers := map[pgtype.Text][]string{
		database.Text(s.hostName): userIDs,
	}
	invalidIDCheckList := map[string]struct{}{}
	for _, id := range invalidUserIDs {
		invalidIDCheckList[id] = struct{}{}
	}

	onlineUserRepo.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("pgtype.TextArray"), mock.AnythingOfType("pgtype.Timestamptz"), mock.Anything).
		Once().Return(nodeOnlineUsers, nil)

	onlineUserRepo.On("SetActive", mock.Anything, mock.Anything, mock.AnythingOfType("pgtype.Text")).
		Times(200).Return(nil)

	commonUserCtx, commonUserCancel := context.WithCancel(ctx)
	defer commonUserCancel()
	mockUsers := make([]*user, 0, totalConnection)
	commonResourcePath := "1"
	invalidResourcePath := "2"
	for _, userID := range userIDs {
		var (
			ctxWithRP context.Context
			cancel    context.CancelFunc
		)

		stream := new(mocks.ChatService_SubscribeV2Server)
		_, isinvalid := invalidIDCheckList[userID]
		if isinvalid {
			ctxWithRP, cancel = mockChatServerCtx(stream, commonUserCtx, userID, invalidResourcePath)
		} else {
			ctxWithRP, cancel = mockChatServerCtx(stream, commonUserCtx, userID, commonResourcePath)
		}

		// send ping msg once
		stream.On("Send", mock.MatchedBy(func(res *pb.SubscribeV2Response) bool {
			return res.GetEvent().GetEventPing() != nil
		})).Once().Return(nil)

		if isinvalid {
			stream.AssertNotCalled(t, "Send", mock.MatchedBy(func(res *pb.SubscribeV2Response) bool {
				return res.GetEvent().GetEventNewMessage() != nil
			}))
		} else {
			// send real msg once
			stream.On("Send", mock.MatchedBy(func(res *pb.SubscribeV2Response) bool {
				return res.GetEvent().GetEventNewMessage() != nil
			})).Once().Return(nil)
		}

		mockUsers = append(mockUsers, &user{
			ctx:    ctxWithRP,
			userID: userID,
			srv:    stream,
			cancel: cancel,
		})
	}

	var wg sync.WaitGroup
	for _, e := range mockUsers {
		wg.Add(1)
		go func(u *user) {
			defer wg.Done()
			err := s.SubscribeV2(&pb.SubscribeV2Request{}, u.srv)
			assert.EqualError(t, err, "ctx.Done: context canceled")
			u.srv.AssertNumberOfCalls(t, "Context", 3)
		}(e)
	}

	// here, to wait for all goroutines to spawn, we periodically check the number of connections
	// the timeout is determined by the context
	success := checkCondition(ctx, func() bool { return s.TotalConnections() == len(userIDs) })
	require.True(t, success)
	time.Sleep(time.Millisecond * 50) // another sleep to check if the number of connections still stays the same after a bit
	assert.Equal(t, len(userIDs), s.TotalConnections())

	// send 1 message
	userSend := mockUsers[0]

	conversationID := idutil.ULIDNow()
	var cID pgtype.Text
	_ = cID.Set(conversationID)

	// we have no user on other node
	jsm.AssertNotCalled(t, "PublishContext", mock.Anything, mock.Anything, mock.Anything)
	// we have no offline user
	userDeviceTokenRepo.AssertNotCalled(t, "Find", mock.Anything, mock.Anything, mock.Anything)
	notiPusher.AssertNotCalled(t, "SendTokens", mock.Anything, mock.Anything, mock.Anything)

	s.PushMessage(userSend.ctx, userIDs, &pb.Event{
		Event: &pb.Event_EventNewMessage{
			EventNewMessage: &pb.MessageResponse{
				ConversationId: conversationID,
				Content:        "Hello",
				UrlMedia:       "image",
				Type:           pb.MESSAGE_TYPE_IMAGE,
				LocalMessageId: "",
			},
		},
	}, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			Enabled:      true,
			IgnoredUsers: []string{userSend.userID},
		},
	})

	// wait all message send
	time.Sleep(500 * time.Millisecond)

	commonUserCancel()
	wg.Wait()

	onlineUserRepo.AssertNumberOfCalls(t, "Insert", len(userIDs))
	onlineUserRepo.AssertNumberOfCalls(t, "Delete", len(userIDs))
	success = checkCondition(ctx, func() bool { return s.TotalConnections() == 0 })
	require.True(t, success)
}

// Have 2 nodes: this node and other node, each has 100 online users
// have some offline user as well
// when sending message:
// - socket of users on this node must receive msg
// - jsm must send broadcasted msg to 100 users online on other node
// - fcm must push msg for offline user
func TestChatService_SubscribeV2PushMessage(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	onlineUserRepo := &mock_repositories.MockOnlineUserRepo{}
	userDeviceTokenRepo := &mock_repositories.MockUserDeviceTokenRepo{}
	notiPusher := &mock_services.Pusher{}

	jsm := &mock_nats.JetStreamManagement{}
	s := &Server{
		hostName:            idutil.ULIDNow(),
		JSM:                 jsm,
		logger:              zap.NewNop(),
		onlineUserRepo:      onlineUserRepo,
		userDeviceTokenRepo: userDeviceTokenRepo,
		notification: &Notification{
			Pusher: notiPusher,
		},
	}
	s.HubStart()
	defer s.HubStop()

	totalOnlineUsers := 200
	type user struct {
		ctx    context.Context
		cancel context.CancelFunc
		userID string
		srv    *mocks.ChatService_SubscribeV2Server
	}

	onlineUserIDs := make([]string, 0, totalOnlineUsers)
	for i := 0; i < totalOnlineUsers; i++ {
		userID := idutil.ULIDNow()
		onlineUserIDs = append(onlineUserIDs, userID)
	}

	commonResourcePath := "1"
	userIDsOnThisNode := onlineUserIDs[:len(onlineUserIDs)/2]
	userIDsOnOtherNode := onlineUserIDs[len(onlineUserIDs)/2:]
	otherNodeName := fmt.Sprintf("%s-2", s.hostName)
	nodeOnlineUsers := map[pgtype.Text][]string{
		database.Text(s.hostName):    userIDsOnThisNode,
		database.Text(otherNodeName): userIDsOnOtherNode,
	}

	offlineUsers := []string{
		idutil.ULIDNow(), idutil.ULIDNow(),
	}
	offlineUserTokens := []string{
		idutil.ULIDNow(), idutil.ULIDNow(),
	}
	totalusers := append(onlineUserIDs, offlineUsers...)

	onlineUserRepo.On("Find", mock.Anything, mock.Anything, mock.AnythingOfType("pgtype.TextArray"), mock.AnythingOfType("pgtype.Timestamptz"), mock.Anything).
		Once().Return(nodeOnlineUsers, nil)

	onlineUserRepo.On("SetActive", mock.Anything, mock.Anything, mock.AnythingOfType("pgtype.Text")).
		Times(200).Return(nil)
	onlineUserRepo.On("Insert", mock.Anything, mock.Anything, mock.AnythingOfType("*core.OnlineUser")).
		Times(100).Return(nil)

	onlineUserRepo.On("Delete", mock.MatchedBy(func(ctx context.Context) bool {
		rp, err := interceptors.ResourcePathFromContext(ctx)
		return err == nil && rp == commonResourcePath
	}), mock.Anything, mock.AnythingOfType("pgtype.Text")).
		Times(100).Return(nil)

	commonUserCtx, commonUserCancel := context.WithCancel(ctx)
	defer commonUserCancel()
	mockUsers := make([]*user, 0, totalOnlineUsers)
	for _, userID := range userIDsOnThisNode {

		stream := new(mocks.ChatService_SubscribeV2Server)
		ctx, cancel := mockChatServerCtx(stream, commonUserCtx, userID, commonResourcePath)

		// send ping msg once
		stream.On("Send", mock.MatchedBy(func(res *pb.SubscribeV2Response) bool {
			return res.GetEvent().GetEventPing() != nil
		})).Once().Return(nil)
		// send real msg once
		stream.On("Send", mock.MatchedBy(func(res *pb.SubscribeV2Response) bool {
			return res.GetEvent().GetEventNewMessage() != nil
		})).Once().Return(nil)

		mockUsers = append(mockUsers, &user{
			ctx:    ctx,
			userID: userID,
			srv:    stream,
			cancel: cancel,
		})
	}

	var wg sync.WaitGroup
	for _, e := range mockUsers {
		wg.Add(1)
		go func(u *user) {
			defer wg.Done()
			err := s.SubscribeV2(&pb.SubscribeV2Request{}, u.srv)
			assert.EqualError(t, err, "ctx.Done: context canceled")
			u.srv.AssertNumberOfCalls(t, "Context", 3)
		}(e)
	}

	// here, to wait for all goroutines to spawn, we periodically check the number of connections
	// the timeout is determined by the context
	success := checkCondition(ctx, func() bool { return s.TotalConnections() == len(userIDsOnThisNode) })
	require.True(t, success)
	time.Sleep(time.Millisecond * 50) // another sleep to check if the number of connections still stays the same after a bit
	assert.Equal(t, len(userIDsOnThisNode), s.TotalConnections())

	// send 1 message
	userSend := mockUsers[0]

	conversationID := idutil.ULIDNow()
	var cID pgtype.Text
	_ = cID.Set(conversationID)

	// Must broadcast msg to users online on other node
	jsm.On("PublishContext", mock.Anything, constants.SubjectSendChatMessageCreated, mock.MatchedBy(func(raw []byte) bool {
		broadCastedToOtherNode := &pb.InternalSendMessageRequest{}
		err := broadCastedToOtherNode.Unmarshal(raw)
		if err != nil {
			return true
		}
		return stringutil.SliceElementsMatch(broadCastedToOtherNode.GetNodeReceive()[otherNodeName].Ids, userIDsOnOtherNode)
	})).Once().Return(&natsJS.PubAck{}, nil)
	userDeviceTokenRepo.On("Find", userSend.ctx, mock.Anything, mock.AnythingOfType("pgtype.TextArray")).Once().Return(offlineUserTokens, nil)

	// for offline users, must send fcm notification
	notiPusher.On("SendTokens", mock.Anything, mock.Anything, offlineUserTokens).Once().Return(0, 0, nil)

	s.PushMessage(userSend.ctx, totalusers, &pb.Event{
		Event: &pb.Event_EventNewMessage{
			EventNewMessage: &pb.MessageResponse{
				ConversationId: conversationID,
				Content:        "Hello",
				UrlMedia:       "image",
				Type:           pb.MESSAGE_TYPE_IMAGE,
				LocalMessageId: "",
			},
		},
	}, domain.MessageToUserOpts{
		Notification: domain.NotificationOpts{
			Enabled:      true,
			IgnoredUsers: []string{userSend.userID},
		},
	})

	// wait all message send
	time.Sleep(500 * time.Millisecond)

	commonUserCancel()
	wg.Wait()

	onlineUserRepo.AssertNumberOfCalls(t, "Insert", len(userIDsOnThisNode))
	onlineUserRepo.AssertNumberOfCalls(t, "Delete", len(userIDsOnThisNode))
	success = checkCondition(ctx, func() bool { return s.TotalConnections() == 0 })
	require.True(t, success)
}
