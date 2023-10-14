package grpcclient

import (
	"context"
	"reflect"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/objectutils"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ConnectionsGrpc struct {
	BobConn              *grpc.ClientConn
	TomConn              *grpc.ClientConn
	YasuoConn            *grpc.ClientConn
	EurekaConn           *grpc.ClientConn
	FatimaConn           *grpc.ClientConn
	ShamirConn           *grpc.ClientConn
	UserMgmtConn         *grpc.ClientConn
	NotificationConn     *grpc.ClientConn
	PaymentConn          *grpc.ClientConn
	EntryExitConn        *grpc.ClientConn
	MasterMgmtConn       *grpc.ClientConn
	InvoiceMgmtConn      *grpc.ClientConn
	LessonMgmtConn       *grpc.ClientConn
	EnigmaConn           *grpc.ClientConn
	VirtualClassroomConn *grpc.ClientConn
	CalendarConn         *grpc.ClientConn
	TimesheetConn        *grpc.ClientConn
	NotificationMgmtConn *grpc.ClientConn
}

func UnaryClientAttachHeaderInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	ctxOutGoing := metadata.NewOutgoingContext(ctx, md)
	return invoker(ctxOutGoing, method, req, reply, cc, opts...)
}

func retryPolicy(options configs.RetryOptions) grpc.DialOption {
	retriableErrors := []codes.Code{codes.Unavailable, codes.DataLoss}
	durationTimeout := time.Duration(options.RetryTimeout) * time.Millisecond
	return grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		UnaryClientAttachHeaderInterceptor, grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(uint(options.MaxCall)),
			grpc_retry.WithBackoff(grpc_retry.BackoffLinear(durationTimeout)),
			grpc_retry.WithCodes(retriableErrors...),
			grpc_retry.WithPerRetryTimeout(durationTimeout),
		)))
}

func ConnectGRPC(ctx context.Context, cf *configs.GRPCClientsConfig) (*ConnectionsGrpc, error) {
	commonDialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()}
	commonDialOptions = append(commonDialOptions, retryPolicy(cf.RetryOptions))

	grpConnMap, err := objectutils.ExtractFieldMapWithSuffix[string](*cf, "SrvAddr")
	if err != nil {
		return nil, err
	}
	connectionsGrpc := &ConnectionsGrpc{}
	for serviceName, addr := range grpConnMap {
		if addr != "" {
			connField := reflect.ValueOf(connectionsGrpc).Elem().FieldByName(serviceName + "Conn")
			if connField.CanSet() {
				conn, err := grpc.DialContext(ctx, addr, commonDialOptions...)
				if err != nil {
					return nil, err
				}
				connField.Set(reflect.ValueOf(conn))
			}
		}
	}

	return connectionsGrpc, nil
}

func (c *ConnectionsGrpc) Close() error {
	grpConnMap, err := objectutils.ExtractFieldMapWithSuffix[*grpc.ClientConn](*c, "Conn")
	if err != nil {
		return err
	}
	for _, connection := range grpConnMap {
		if connection != nil {
			err := connection.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
