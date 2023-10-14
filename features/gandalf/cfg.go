package gandalf

import (
	"context"
	"fmt"

	config "github.com/manabie-com/backend/internal/gandalf/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Deprecated directly use cfg from Gandalf or features/common instead
type Config struct {
	config.Config
}

// ConnectGRPC uses parameters from Config and connects to GRPC services for Bob and Tom
// at bobConn and tomConn, respectively.
func (c *Config) ConnectGRPC(
	ctx context.Context,
	credentials grpc.DialOption,
	bobConn **grpc.ClientConn,
	tomConn **grpc.ClientConn,
	yasuoConn **grpc.ClientConn,
	eurekaConn **grpc.ClientConn,
	fatimaConn **grpc.ClientConn,
	shamirConn **grpc.ClientConn,
	userMgmtConn **grpc.ClientConn,
	entryExitMgmtConn **grpc.ClientConn,
) error {
	rsc := bootstrap.NewResources().WithLoggerC(&c.Common)
	conn, err := grpc.DialContext(ctx, rsc.GetAddress("bob"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot connect to Bob at %s: %s", rsc.GetAddress("bob"), err)
	}
	*bobConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("tom"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Tom at %s: %s", rsc.GetAddress("tom"), err)
	}
	*tomConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("yasuo"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Yasuo at %s: %s", rsc.GetAddress("yasuo"), err)
	}
	*yasuoConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("eureka"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Eureka at %s: %s", rsc.GetAddress("eureka"), err)
	}
	*eurekaConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("fatima"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Fatima at %s: %s", rsc.GetAddress("fatima"), err)
	}
	*fatimaConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("shamir"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Shamir at %s: %s", rsc.GetAddress("shamir"), err)
	}
	*shamirConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("usermgmt"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Usermgmt at %s: %s", rsc.GetAddress("usermgmt"), err)
	}
	*userMgmtConn = conn

	conn, err = grpc.DialContext(ctx, rsc.GetAddress("entryexitmgmt"), credentials, grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("cannot to connect to Entryexitmgmt at %s: %s", rsc.GetAddress("entryexitmgmt"), err)
	}
	*entryExitMgmtConn = conn

	return nil
}

// ConnectGRPCInsecure is similar to ConnectGRPC, but uses grpc.Insecure() instead of
// an actual credential.
func (c *Config) ConnectGRPCInsecure(
	ctx context.Context,
	bobConn **grpc.ClientConn,
	tomConn **grpc.ClientConn,
	yasuoConn **grpc.ClientConn,
	eurekaConn **grpc.ClientConn,
	fatimaConn **grpc.ClientConn,
	shamirConn **grpc.ClientConn,
	userMgmtConn **grpc.ClientConn,
	entryExitMgmtConn **grpc.ClientConn,
) error {
	return c.ConnectGRPC(ctx, grpc.WithInsecure(), bobConn, tomConn, yasuoConn, eurekaConn, fatimaConn, shamirConn, userMgmtConn, entryExitMgmtConn)
}

// ConnectDB uses parameters from Config and connects to Postgres databases for Bob and
// Tom at bobDB and tomDB, respectively.
func (c *Config) ConnectDB(ctx context.Context, bobDB **pgxpool.Pool, tomDB **pgxpool.Pool, eurekaDB **pgxpool.Pool, fatimaDB **pgxpool.Pool, zeusDB **pgxpool.Pool) {
	*bobDB = getDBConnectionDemo(ctx, c.PostgresV2.Databases["bob"])
	*tomDB = getDBConnectionDemo(ctx, c.PostgresV2.Databases["tom"])
	*eurekaDB = getDBConnectionDemo(ctx, c.PostgresV2.Databases["eureka"])
	*fatimaDB = getDBConnectionDemo(ctx, c.PostgresV2.Databases["fatima"])
	*zeusDB = getDBConnectionDemo(ctx, c.PostgresV2.Databases["zeus"])
}

func (c *Config) ConnectSpecificDB(ctx context.Context, bobPostgresDB **pgxpool.Pool) {
	bobPostgres := c.PostgresV2.Databases["bob"]
	bobPostgres.User = "postgres"
	bobPostgres.Password = c.PostgresMigrate.Database.Password
	*bobPostgresDB = getDBConnectionDemo(ctx, bobPostgres)
}

func getDBConnectionDemo(ctx context.Context, pgCfg configs.PostgresDatabaseConfig) *pgxpool.Pool {
	db, _, _ := database.NewPool(ctx, zap.NewNop(), pgCfg)
	return db
}
