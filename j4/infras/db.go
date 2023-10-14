// Every thing you need to talk to our clusters
package infras

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"

	"github.com/jackc/pgx/v4/pgxpool"
	grpcpool "github.com/processout/grpc-go-pool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Connections struct {
	// key by svcname
	DBConnPools map[string]*pgxpool.Pool
	dbCleanUps  map[string]func() error

	grpcPool      map[string]*grpcpool.Pool
	GrpcConns     map[string]*grpc.ClientConn
	hasura        map[string]*Hasura
	Kafka         kafka.KafkaManagement
	PoolToGateWay *grpcpool.Pool
}

func (c *Connections) ConnectDB(ctx context.Context, cfg configs.PostgresConfigV2) error {
	c.DBConnPools = make(map[string]*pgxpool.Pool)
	c.dbCleanUps = make(map[string]func() error)
	for svcName, dbconfig := range cfg.Databases {
		dbpool, dbcancel, err := database.NewPool(ctx, zap.NewNop(), dbconfig)
		if err != nil {
			return fmt.Errorf("failed to connect to database %q: %s", svcName, err)
		}
		c.DBConnPools[svcName] = dbpool
		c.dbCleanUps[svcName] = dbcancel
	}
	return nil
}

func (c *Connections) CloseAllConnections() {
	for _, cleanup := range c.dbCleanUps {
		_ = cleanup()
	}
	for _, clientPool := range c.grpcPool {
		clientPool.Close()
	}
}
