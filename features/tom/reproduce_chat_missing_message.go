package tom

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/ktr0731/grpc-web-go-client/grpcweb"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// user_id_1 subscribe conversations - LRUCache [user_id_1]
// after 1 second user_id_2 subscribe conversation - LRUCache [user_id_1, user_id_2]
// user_id_1 ping every 4s

// ----------------------------------------------
// second  | lru cache
// 0       | [user_id_1]            user_id_1 subscribe conversation
// 1       | [user_id_1, user_id_2] user_id_2 subscribe conversation
// 2       | [user_id_1, user_id_2]
// 3       | [user_id_1, user_id_2]
// 4       | [user_id_1, user_id_2] user_id_1 ping
// 5       | [user_id_1, user_id_2] user_id_2 ping
// 6       | [user_id_1, user_id_2]
// 7       | [user_id_1, user_id_2]
// 8       | [user_id_1, user_id_2] user_id_1 ping
// 9       | [user_id_1, user_id_2] user_id_2 ping
// 10      | [user_id_1, user_id_2]
// 11      | [user_id_2]            user_id_1 expire, send message at this time, this will raise error, user_id_2 cannot get the message.
//
//	Because in func (r *OnlineUserRepo) Find(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, since pgtype.Timestamptz, msg *pb.Event) (mapNodeUserIDs map[pgtype.Text][]string, err error)
//	we clear user_ids in cache because of this line
//	mapNodeUserIDs[node] = userIDs
type ChatClient struct {
	// either of them, not both
	grpcClient    pb.ChatService_SubscribeV2Client
	grpcWebClient grpcweb.ServerStream

	userID    string
	token     string
	sessionID string
}
type messageLostSuite struct {
	token   string
	userID  string
	clients []*ChatClient
	convID  string
}

func (s *messageLostSuite) toCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, messageLostSuiteKey{}, s)
}
func (s messageLostSuite) fromCtx(ctx context.Context) *messageLostSuite {
	return ctx.Value(messageLostSuiteKey{}).(*messageLostSuite)
}

type messageLostSuiteKey struct{}

const intervalPing = 3 * time.Second
const sendMessageDeadline = 22 * time.Second

func (s *suite) allMemberSubscribeToThisConversationAndChat(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()
	var mu sync.Mutex
	var wg sync.WaitGroup

	// from Background step, we now have 1 conversation with 1 teacher and 1 student
	// number of member in conversation = 2

	// user_id_1
	conversationID := s.conversationID
	m1 := s.ConversationMembers[conversationID][0]
	m2 := s.ConversationMembers[conversationID][1]

	userClientMap := make(map[string]*ChatClient)

	errCh := make(chan error)
	done1 := make(chan bool)
	done2 := make(chan bool)

	// member 1 enter to conversation
	wg.Add(1)
	t1 := time.AfterFunc(time.Second, func() {
		defer wg.Done()
		token, err := s.generateExchangeToken(m1.UserID.String, m1.Role.String, applicantID, s.getSchool(), s.ShamirConn)
		if err != nil {
			errCh <- err
			return
		}
		c, err := s.UserSubscribeV2(ctx2, m1.UserID.String, token)
		if err != nil {
			errCh <- err
			return
		}
		mu.Lock()
		userClientMap[m1.UserID.String] = c
		mu.Unlock()
		ctx, err = s.PingSubscribeV2(ctx2, c.grpcClient, token, c.sessionID, done1)
		if err != nil {
			errCh <- err
		}
	})
	defer t1.Stop()

	// member 2 enter to conversation
	wg.Add(1)
	t2 := time.AfterFunc(2*time.Second, func() {
		defer wg.Done()
		token, err := s.generateExchangeToken(m2.UserID.String, m1.Role.String, applicantID, s.getSchool(), s.ShamirConn)
		if err != nil {
			errCh <- err
			return
		}
		c, err := s.UserSubscribeV2(ctx2, m2.UserID.String, token)
		if err != nil {
			errCh <- err
			return
		}
		mu.Lock()
		userClientMap[m2.UserID.String] = c
		mu.Unlock()
		ctx, err = s.PingSubscribeV2(ctx2, c.grpcClient, token, c.sessionID, done2)
		if err != nil {
			errCh <- err
		}
	})
	defer t2.Stop()

	// 10 second after since member_1 enter conversation, member_1 is expired in cache
	// if we send message at this time and member_1 will not get the message
	wg.Add(1)
	t3 := time.AfterFunc(5*time.Second, func() {
		defer func() {
			wg.Done()
			done1 <- true
			done2 <- true
		}()

		c := userClientMap[m1.UserID.String]
		c2 := userClientMap[m2.UserID.String]
		i := 0

		chatFinishCh := make(chan bool)
		go func() {
			time.Sleep(sendMessageDeadline)
			chatFinishCh <- true
		}()

	loop:
		for {
			select {
			case <-chatFinishCh:
				break loop
			default:
				content := fmt.Sprintf("hello %v", i)
				// fmt.Println("content: ", content)
				resp, err := s.SendMessage(ctx2, c.token, conversationID, content)
				if err != nil {
					errCh <- err
					break loop
				}
				msg, err := s.getNewMsgFromGrpc(c.grpcClient)
				if err != nil {
					errCh <- fmt.Errorf("cannot receive message_id %v getNewMsg %v", resp.MessageId, err)
					break loop
				}
				msg2, err := s.getNewMsgFromGrpc(c2.grpcClient)
				if err != nil {
					errCh <- fmt.Errorf("cannot receive message_id %v getNewMsg %v", resp.MessageId, err)
					break loop
				}
				if msg.GetContent() != content || msg2.GetContent() != content {
					errCh <- fmt.Errorf("error receive unexpected message, expect %v but got %v and %v", content, msg.GetContent(), msg2.GetContent())
					break loop
				}
				i++
			}
		}
	})
	defer t3.Stop()

	go func() {
		wg.Wait()
		close(errCh)
	}()
	var err error
	for e := range errCh {
		err = multierr.Append(err, e)
	}
	return ctx, err
}

func (s *suite) aUserMakeSubscribesStreamUsingGrpcConnectionsAndGrpcWebConnections(ctx context.Context, grpcnum, grpcwebnum int) (context.Context, error) {
	clients, userID, token, convID, err := s.connectUsersUsingMultipleConnections(ctx, grpcnum, grpcwebnum)
	if err != nil {
		return ctx, err
	}

	state := &messageLostSuite{
		clients: clients,
		userID:  userID,
		convID:  convID,
		token:   token,
	}
	return state.toCtx(ctx), nil
}

func (s *suite) spammingIntoConversation(ctx context.Context, msgCount int) (context.Context, error) {
	st := messageLostSuite{}
	msgLostState := st.fromCtx(ctx)
	convID, token := msgLostState.convID, msgLostState.token
	for i := 0; i < msgCount; i++ {
		content := strconv.Itoa(i)
		_, err := s.SendMessage(ctx, token, convID, content)
		if err != nil {
			return ctx, err
		}
		time.Sleep(100 * time.Millisecond)
	}
	return ctx, nil
}

func (s *suite) allConnectionsReceiveMsgInOrder(ctx context.Context, msgCount int) (context.Context, error) {
	st := messageLostSuite{}
	msgLostState := st.fromCtx(ctx)
	clients := msgLostState.clients
	gr := &errgroup.Group{}
	for i := 0; i < len(clients); i++ {
		client := clients[i]
		gr.Go(func() error {
			receivedBuffer := make([]string, 0, msgCount)
			for count := 0; count < msgCount; count++ {
				var (
					msg *pb.MessageResponse
					err error
				)
				if client.grpcClient != nil {
					msg, err = s.getNewMsgFromGrpc(client.grpcClient)
				} else {
					msg, err = s.getNewMsgFromGrpcWeb(ctx, client.grpcWebClient)
				}
				if err != nil {
					return err
				}
				countInMsg, err := strconv.Atoi(msg.Content)
				if err != nil {
					return err
				}
				receivedBuffer = append(receivedBuffer, msg.Content)
				if countInMsg != count {
					msgLeft := countInMsg - count + 1
					for i := 0; i < msgLeft; i++ {
						if client.grpcClient != nil {
							msg, err = s.getNewMsgFromGrpc(client.grpcClient)
						} else {
							msg, err = s.getNewMsgFromGrpcWeb(ctx, client.grpcWebClient)
						}
						if err != nil {
							return multierr.Combine(err, fmt.Errorf("msg receive out of order: %v", receivedBuffer))
						}
						receivedBuffer = append(receivedBuffer, msg.Content)
					}
					return fmt.Errorf("msg receive out of order: %v", receivedBuffer)
				}
			}
			return nil
		})
	}
	return ctx, gr.Wait()
}

func (s *suite) getNewMsgFromGrpcWeb(ctx context.Context, stream grpcweb.ServerStream) (*pb.MessageResponse, error) {
	msg := &pb.SubscribeV2Response{}
	for try := 0; try < 5; try++ {
		err := stream.Receive(ctx, msg)
		if err != nil {
			fmt.Printf("debug %v\n", stream.Trailer())
			return nil, err
		}
		if msg.GetEvent().GetEventNewMessage() == nil {
			continue
		}
		return msg.GetEvent().GetEventNewMessage(), nil
	}
	return nil, fmt.Errorf("no signal from upstream")
}

func (s *suite) connectUsersUsingMultipleConnections(ctx context.Context, grpcConn int, grpcWebConn int) (clients []*ChatClient, userID, token, convID string, err error) {
	clients = make([]*ChatClient, grpcConn+grpcWebConn)
	ctx, err = s.createAValidStudentConversationInDBWithATeacherAndAStudent(ctx)
	if err != nil {
		return
	}

	convID = s.conversationID
	chosenUser := s.ConversationMembers[convID][0]
	userID = chosenUser.UserID.String
	token, err = s.generateExchangeToken(userID, chosenUser.Role.String, applicantID, s.getSchool(), s.ShamirConn)
	if err != nil {
		return
	}
	gr := &errgroup.Group{}
	for i := 0; i < grpcConn; i++ {
		curIDx := i
		gr.Go(func() error {
			c, err := s.UserSubscribeV2(ctx, userID, token)
			if err != nil {
				return err
			}

			go func() {
				// having at least 1 pingger to remain connection
				_, err := s.PingSubscribeV2(ctx, c.grpcClient, c.token, c.sessionID, nil)
				if err != nil {
					fmt.Printf("pingSubscribeV2 %v\n", err)
				}
			}()
			clients[curIDx] = c
			return nil
		})
	}
	for i := grpcConn; i < grpcWebConn+grpcConn; i++ {
		curIDx := i
		gr.Go(func() error {
			c, err := s.UserSubscribeV2GrpcWeb(ctx, userID, token)
			if err != nil {
				return err
			}
			go func() {
				// having at least 1 pingger to remain connection
				_, err := s.PingSubscribeV2(ctx, c.grpcClient, c.token, c.sessionID, nil)
				if err != nil {
					fmt.Printf("pingSubscribeV2 %v\n", err)
				}
			}()
			clients[curIDx] = c
			return nil
		})
	}
	err = gr.Wait()
	return
}

func (s *suite) grpcMetadataWithKeyInContext(ctx context.Context, key string) (context.Context, error) {
	return metadata.AppendToOutgoingContext(ctx, key, idutil.ULIDNow()), nil
}

func (s *suite) allConnectionsAreRoutedToMultipleNodes(ctx context.Context) (context.Context, error) {
	var count int
	st := messageLostSuite{}
	msgLostState := st.fromCtx(ctx)
	userID := msgLostState.userID
	err := s.DB.QueryRow(ctx, `select count(*) from (select distinct user_id, node_name from online_users where user_id=$1) as conn_count`, database.Text(userID)).Scan(&count)
	if err != nil {
		return ctx, err
	}
	if count == 1 {
		return ctx, fmt.Errorf("all connections connected to 1 node only, consistent hash not applied")
	}
	return ctx, nil
}

func (s *suite) allConnectionsAreRoutedToOneNode(ctx context.Context) (context.Context, error) {
	var count int
	st := messageLostSuite{}
	msgLostState := st.fromCtx(ctx)
	userID := msgLostState.userID
	err := s.DB.QueryRow(ctx, `select count(*) from (select distinct user_id, node_name from online_users where user_id=$1) as conn_count`, database.Text(userID)).Scan(&count)
	if err != nil {
		return ctx, err
	}
	if count != 1 {
		return ctx, fmt.Errorf("connections connected to %d nodes, consistent hash not applied", count)
	}
	return ctx, nil
}

func (s *suite) UserSubscribeV2GrpcWeb(ctx context.Context, userID string, tok string) (*ChatClient, error) {
	// tok, err := generateValidAuthenticationToken(userID)
	// if err != nil {
	// 	return nil, err
	// }
	ctx = contextWithToken(ctx, tok)

	conn := s.GrpcWebConn
	stream, err := conn.NewServerStream(&grpc.StreamDesc{
		StreamName:    "SubscribeV2",
		ServerStreams: true,
	}, "/manabie.tom.ChatService/SubscribeV2")

	if err != nil {
		return nil, err
	}
	err = stream.Send(ctx, &pb.SubscribeV2Request{})
	if err != nil {
		return nil, fmt.Errorf("sending stream grpc-web request %w", err)
	}

	newmsg := pb.SubscribeV2Response{}
	err = stream.Receive(ctx, &newmsg)
	if err != nil {
		return nil, err
	}
	msg := newmsg.GetEvent().GetEventPing()
	if msg == nil {
		return nil, fmt.Errorf("expect first message is a ping event")
	}
	return &ChatClient{
		grpcWebClient: stream,
		sessionID:     msg.GetSessionId(),
		userID:        userID,
		token:         tok,
	}, nil
}

func (s *suite) UserSubscribeV2(ctx context.Context, userID string, token string) (*ChatClient, error) {
	subClient, err := pb.NewChatServiceClient(s.Conn).SubscribeV2(contextWithToken(ctx, token), &pb.SubscribeV2Request{})
	if err != nil {
		return nil, err
	}
	msg, err := s.getPingMsg(subClient)
	if err != nil {
		return nil, err
	}
	c := &ChatClient{grpcClient: subClient, userID: userID, sessionID: msg.GetSessionId(), token: token}
	return c, err
}
func (s *suite) getPingMsg(client pb.ChatService_SubscribeV2Client) (*pb.EventPing, error) {
	var msg *pb.EventPing
	for msg == nil {
		resp, err := client.Recv()
		if err != nil {
			return nil, err
		}
		if err == io.EOF {
			break
		}
		evt := resp.GetEvent()
		if evt == nil {
			return nil, fmt.Errorf("cannot get event message")
		}
		msg = evt.GetEventPing()
	}
	return msg, nil
}

func (s *suite) getNewMsgFromGrpc(client pb.ChatService_SubscribeV2Client) (*pb.MessageResponse, error) {
	msgCh := make(chan *pb.MessageResponse)
	errCh := make(chan error)
	var wg sync.WaitGroup
	wg.Add(1)
	foo := func() {
		defer wg.Done()
		var msg *pb.MessageResponse
		for msg == nil {
			resp, err := client.Recv()
			if err != nil {
				errCh <- err
				return
			}
			if err == io.EOF {
				return
			}
			evt := resp.GetEvent()
			if evt == nil {
				errCh <- fmt.Errorf("cannot get event message")
				return
			}
			msg = evt.GetEventNewMessage()
		}
		msgCh <- msg
	}
	go foo()
	go func() { wg.Wait(); close(msgCh); close(errCh) }()
	var msg *pb.MessageResponse
	var err error
	select {
	case msg = <-msgCh:
	case err = <-errCh:
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("time out after 5s")
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}
func (s *suite) signedTokenCtx(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}
func (s *suite) PingSubscribeV2(ctx context.Context, client pb.ChatService_SubscribeV2Client, token string, sessionID string, done chan bool) (context.Context, error) {
	var err error
	// token, err := generateValidAuthenticationToken(userID)
	// if err != nil {
	// 	return ctx, err
	// }
	ticker := time.NewTicker(intervalPing)

loop:
	for {
		select {
		case <-ticker.C:
			_, err = pb.NewChatServiceClient(s.Conn).PingSubscribeV2(
				s.signedTokenCtx(ctx, token),
				&pb.PingSubscribeV2Request{
					SessionId: sessionID,
				})
			if err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case <-done:
			break loop
		}
	}
	return ctx, err
}

func (s *suite) SendMessage(ctx context.Context, token string, conversationID, msg string) (*pb.SendMessageResponse, error) {
	resp, err := pb.NewChatServiceClient(s.Conn).SendMessage(
		s.signedTokenCtx(ctx, token),
		&pb.SendMessageRequest{
			ConversationId: conversationID,
			Message:        msg,
			Type:           pb.MESSAGE_TYPE_TEXT,
		})
	if err != nil {
		return nil, err
	}
	return resp, err
}
