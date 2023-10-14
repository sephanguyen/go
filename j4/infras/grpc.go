package infras

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/manabie-com/backend/internal/golibs/tracer"

	grpcpool "github.com/processout/grpc-go-pool"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	TlsSkipVerifyDialingOpts = []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})), //nolint:gosec
	}

	InsecureDialingOpts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
)

func (c *Connections) GetGrpcConnByAddr(addr string) *grpc.ClientConn {
	return c.GrpcConns[addr]
}

func (c *Connections) ConnecGrpc(ctx context.Context, addr string, opts []grpc.DialOption) error {
	if c.GrpcConns == nil {
		c.GrpcConns = map[string]*grpc.ClientConn{}
	}
	commonDialOptions := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithStatsHandler(&tracer.B3Handler{
			ClientHandler: &ocgrpc.ClientHandler{},
		}),
	}

	commonDialOptions = append(commonDialOptions, opts...)
	ctx, cancel := context.WithTimeout(ctx, time.Second*15)
	defer cancel()
	conn, err := grpc.DialContext(ctx, addr, commonDialOptions...)
	if err != nil {
		return err
	}

	c.GrpcConns[addr] = conn
	return nil
}

// To simulate multiple clients from multiple machine calling to the gateway, instead of one single client shared
// between scenarios like Gandalf
func (c *Connections) ConnectGrpcPool(ctx context.Context, addr string, isgatewayddr bool, opts []grpc.DialOption) error {
	if c.grpcPool == nil {
		c.grpcPool = map[string]*grpcpool.Pool{}
	}
	factory := func() (*grpc.ClientConn, error) {
		commonDialOptions := []grpc.DialOption{
			grpc.WithBlock(),
			grpc.WithStatsHandler(&tracer.B3Handler{
				ClientHandler: &ocgrpc.ClientHandler{},
			}),
		}

		commonDialOptions = append(commonDialOptions, opts...)
		ctx, cancel := context.WithTimeout(ctx, time.Second*15)
		defer cancel()
		return grpc.DialContext(ctx, addr, commonDialOptions...)
	}

	// TODO: let user decide connection pool setting
	pool, err := grpcpool.New(factory, 5, 20, 5*time.Minute, 10*time.Minute)
	if err != nil {
		return err
	}

	// in local, we can have individual pool for individual service
	c.grpcPool[addr] = pool

	if isgatewayddr {
		c.PoolToGateWay = pool
	}
	return nil
}
