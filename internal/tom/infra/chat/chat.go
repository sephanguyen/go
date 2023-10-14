package chat

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/metrics"
	"github.com/manabie-com/backend/internal/golibs/nats"
	entities "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"firebase.google.com/go/v4/messaging"
	"github.com/gogo/protobuf/types"
	"github.com/jackc/pgtype"
	"go.uber.org/zap"
)

type UserDeviceTokenRepo interface {
	Find(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray) ([]string, error)
	FindByUserID(ctx context.Context, db database.QueryExecer, userID pgtype.Text) (c *entities.UserDeviceToken, err error)
}

type Pusher interface {
	SendTokens(ctx context.Context, msg *messaging.MulticastMessage, tokens []string) (successCount, failureCount int, err *firebase.SendTokensError)
	RetrievePushedMessages(ctx context.Context, deviceToken string, limit int, since *types.Timestamp) ([]*messaging.MulticastMessage, error)
}

type Notification struct {
	Pusher Pusher
}

type Server struct {
	logger              *zap.Logger
	JSM                 nats.JetStreamManagement
	hostName            string
	db                  database.Ext
	userDeviceTokenRepo UserDeviceTokenRepo
	onlineUserRepo      interface {
		Insert(ctx context.Context, db database.QueryExecer, e *entities.OnlineUser) error
		SetActive(ctx context.Context, db database.QueryExecer, ID pgtype.Text) error
		Delete(ctx context.Context, db database.QueryExecer, ID pgtype.Text) error
		Find(ctx context.Context, db database.QueryExecer, userIDs pgtype.TextArray, since pgtype.Timestamptz, msg *pb.Event) (mapNodeUserIDs map[pgtype.Text][]string, err error)
		DeleteByNode(ctx context.Context, db database.QueryExecer, node pgtype.Text) error
		InvalidateCache(ttl time.Duration)
	}

	Hubs                        []*Hub
	HubsStopCheckingForDeadlock chan bool
	notification                *Notification
	metrics                     struct {
		clientSideDis int64
		serverSideDis int64
	}
}

func NewChatServer(
	ctx context.Context,
	hostName string,
	logger *zap.Logger,
	wrapperDB database.Ext,
	jsm nats.JetStreamManagement,
	ur UserDeviceTokenRepo,
	onlineUserRepo *repositories.OnlineUserRepo,
	n *Notification,
	collector metrics.MetricCollector,
) *Server {
	svc := &Server{
		db:                  wrapperDB,
		logger:              logger,
		JSM:                 jsm,
		hostName:            hostName,
		userDeviceTokenRepo: ur,
		notification:        n,
		onlineUserRepo:      onlineUserRepo,
	}
	svc.RegisterMetric(collector)

	svc.HubStart()
	go svc.InvalidateCacheUserOnline(ctx)

	return svc
}

func (rcv *Server) DeleteOnlineUser(ctx context.Context) {
	err := rcv.onlineUserRepo.DeleteByNode(ctx, rcv.db, database.Text(rcv.hostName))
	if err != nil {
		rcv.logger.Error("err onlineUser DeleteByNode", zap.Error(err))
	}
}

func (rcv *Server) TotalConn() float64 {
	return float64(rcv.TotalConnections())
}

func (rcv *Server) clientDisconnections() float64 {
	return float64(atomic.LoadInt64(&rcv.metrics.clientSideDis))
}

func (rcv *Server) serverDisconnections() float64 {
	return float64(atomic.LoadInt64(&rcv.metrics.serverSideDis))
}
