package chat

import (
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync/atomic"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"go.uber.org/zap"
)

type Hub struct {
	connectionCount int64
	connectionIndex int
	register        chan *ClientConn
	unregister      chan *ClientConn
	broadcast       chan *Event
	stop            chan struct{}
	didStop         chan struct{}
	ExplicitStop    bool
	goroutineID     int
	chatService     *Server
}

type Event struct {
	UserID       string
	ResourcePath string
	Data         *pb.Event
}

func (rcv *Hub) Register(clientConn *ClientConn) {
	select {
	case rcv.register <- clientConn:
	case <-rcv.didStop:
	}
}

func (rcv *Hub) Unregister(clientConn *ClientConn) {
	select {
	case rcv.unregister <- clientConn:
	case <-rcv.stop:
	}
}

func (rcv *Hub) Broadcast(message *Event) {
	if rcv != nil && rcv.broadcast != nil && message != nil {
		select {
		case rcv.broadcast <- message:
		case <-rcv.didStop:
		}
	}
}

func (rcv *Hub) Stop() {
	close(rcv.stop)
	<-rcv.didStop
}

func getGoroutineID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		id = -1
	}
	return id
}

func (rcv *Hub) Start() {
	var doStart func()
	var doRecoverableStart func()
	var doRecover func()

	doStart = func() {
		rcv.goroutineID = getGoroutineID()
		rcv.chatService.logger.Debug("Hub for index is starting with goroutine", zap.Int("index", rcv.connectionIndex), zap.Int("goroutine", rcv.goroutineID))

		connsContainer := newHubConnectionContainer()

		for {
			select {
			case clientConn := <-rcv.register:
				connsContainer.Add(clientConn)
				atomic.StoreInt64(&rcv.connectionCount, int64(len(connsContainer.All())))
			case clientConn := <-rcv.unregister:
				itemExist := connsContainer.Remove(clientConn)
				if !itemExist && len(rcv.register) > 0 {
					// This may happen because <-unregister branch comes before <-register (buffered 1 item)
					// if the conn inside register buffer == removed conn inside unregister buffer, we ignore it
					// else we treat it like adding new conn
					newConn := <-rcv.register
					if newConn == clientConn {
						continue
					}
					connsContainer.Add(newConn)
					atomic.StoreInt64(&rcv.connectionCount, int64(len(connsContainer.All())))
					continue
				}
				atomic.StoreInt64(&rcv.connectionCount, int64(len(connsContainer.All())))
			case msg := <-rcv.broadcast:
				clientConn := connsContainer.ForUser(msg.UserID)

				conversationID := ""
				messageID := ""
				newMsg := msg.Data.GetEventNewMessage()
				if newMsg != nil {
					conversationID = newMsg.GetConversationId()
					messageID = newMsg.GetMessageId()
				}
				var removedConns []*ClientConn
				for _, conn := range clientConn {
					rcv.chatService.logger.Info(
						"Hub.Start hub broadcast message to user",
						zap.String("host", rcv.chatService.hostName),
						zap.String("session_id", conn.SessionID),
						zap.String("user_id", conn.UserID),
						zap.String("conversation_id", conversationID),
						zap.String("message_id", messageID),
					)
					if msg.ResourcePath != conn.ResourcePath {
						rcv.chatService.logger.Warn("some how msg has different resource path from target socket", zap.String("message_id", messageID),
							zap.String("message_rp", msg.ResourcePath), zap.String("socket_rp", conn.ResourcePath))
						continue
					}

					select {
					case conn.Send <- msg.Data:
					default:
						rcv.chatService.logger.Error("hub.broadcast: cannot send, closing stream for user", zap.String("user_id", conn.UserID))
						close(conn.Send)
						// don't call connection.Remove directly because we are iterating over it
						removedConns = append(removedConns, conn)
					}
				}
				if len(removedConns) != 0 {
					for _, removed := range removedConns {
						connsContainer.Remove(removed)
					}
					atomic.StoreInt64(&rcv.connectionCount, int64(len(connsContainer.All())))
				}
			case <-rcv.stop:
				for _, clientConn := range connsContainer.All() {
					clientConn.Close()
				}

				rcv.ExplicitStop = true
				close(rcv.didStop)

				return
			}
		}
	}

	doRecoverableStart = func() {
		defer doRecover()
		doStart()
	}

	doRecover = func() {
		if !rcv.ExplicitStop {
			if r := recover(); r != nil {
				rcv.chatService.logger.Error("Recovering from Hub panic.", zap.Any("panic", r))
			} else {
				rcv.chatService.logger.Error("Webhub stopped unexpectedly. Recovering.")
			}

			rcv.chatService.logger.Error(string(debug.Stack()))

			go doRecoverableStart()
		}
	}

	go doRecoverableStart()
}

type connectionIndexes struct {
	hubConnectionsIndex  int
	userConnectionsIndex int
}

// hubConnectionsContainer provides fast addition, removal, and iteration of connections.
type hubConnectionsContainer struct {
	connections         []*ClientConn
	connectionsByUserID map[string][]*ClientConn
	// index of above array
	connectionIndexes map[*ClientConn]*connectionIndexes
}

func newHubConnectionContainer() *hubConnectionsContainer {
	return &hubConnectionsContainer{
		connections:         make([]*ClientConn, 0, SESSION_CACHE_SIZE),
		connectionsByUserID: make(map[string][]*ClientConn),
		connectionIndexes:   make(map[*ClientConn]*connectionIndexes),
	}
}

func (rcv *hubConnectionsContainer) Add(clientConn *ClientConn) {
	rcv.connections = append(rcv.connections, clientConn)
	rcv.connectionsByUserID[clientConn.UserID] = append(rcv.connectionsByUserID[clientConn.UserID], clientConn)
	rcv.connectionIndexes[clientConn] = &connectionIndexes{
		hubConnectionsIndex:  len(rcv.connections) - 1,
		userConnectionsIndex: len(rcv.connectionsByUserID[clientConn.UserID]) - 1,
	}
}

func (rcv *hubConnectionsContainer) Remove(clientConn *ClientConn) (itemExist bool) {
	indexes, ok := rcv.connectionIndexes[clientConn]
	if !ok {
		itemExist = false
		// This happen maybe because the <-unregister branch comes before the <-register
		return
	}

	// remove the current connection by moving the last item to current item
	// for hub connections
	last := rcv.connections[len(rcv.connections)-1]
	rcv.connections[indexes.hubConnectionsIndex] = last
	rcv.connections[len(rcv.connections)-1] = nil
	rcv.connections = rcv.connections[:len(rcv.connections)-1]
	rcv.connectionIndexes[last].hubConnectionsIndex = indexes.hubConnectionsIndex

	// for connections by userID
	userConnections := rcv.connectionsByUserID[clientConn.UserID]
	last = userConnections[len(userConnections)-1]

	userConnections[indexes.userConnectionsIndex] = last
	userConnections[len(userConnections)-1] = nil
	rcv.connectionsByUserID[clientConn.UserID] = userConnections[:len(userConnections)-1]
	rcv.connectionIndexes[last].userConnectionsIndex = indexes.userConnectionsIndex

	delete(rcv.connectionIndexes, clientConn)
	itemExist = true
	return
}

func (rcv *hubConnectionsContainer) ForUser(id string) []*ClientConn {
	return rcv.connectionsByUserID[id]
}

func (rcv *hubConnectionsContainer) All() []*ClientConn {
	return rcv.connections
}
