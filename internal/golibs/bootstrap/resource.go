package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/elastic"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/tracer"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opencensus.io/plugin/ocgrpc"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"k8s.io/utils/strings/slices"
)

var addresses = map[string]configs.ListenerConfig{
	"auth":             {GRPC: ":7550", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"bob":              {GRPC: ":5050", HTTP: ":5080", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"calendar":         {GRPC: ":7050", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"discount":         {GRPC: ":7450", HTTP: ":7480", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"draft":            {GRPC: ":6050", HTTP: ":6080", MigratedEnvironments: []string{"stag"}},
	"enigma":           {HTTP: ":5380", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"entryexitmgmt":    {GRPC: ":6350", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"eureka":           {GRPC: ":5550", HTTP: ":5580", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"fatima":           {GRPC: ":5450", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"hephaestus":       {GRPC: ":7150", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"conversationmgmt": {GRPC: ":7350", HTTP: ":7380", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"invoicemgmt":      {GRPC: ":6650", HTTP: ":6680", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"jerry":            {HTTP: ":8081", MigratedEnvironments: []string{"local", "stag"}},
	"lessonmgmt":       {GRPC: ":6550", HTTP: ":6580", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"mastermgmt":       {GRPC: ":6450", HTTP: ":6480", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"notificationmgmt": {GRPC: ":6950", HTTP: ":6980", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"payment":          {GRPC: ":6250", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"shamir":           {GRPC: ":5650", HTTP: ":5680", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"spike":            {GRPC: ":7450", HTTP: ":7480", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"timesheet":        {GRPC: ":6850", HTTP: ":6880", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"tom":              {GRPC: ":5150", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"usermgmt":         {GRPC: ":6150", HTTP: ":6180", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"virtualclassroom": {GRPC: ":6750", HTTP: ":6760", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"yasuo":            {GRPC: ":5250", HTTP: ":5280", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"zeus":             {GRPC: ":5950", MigratedEnvironments: []string{"local", "stag", "uat", "dorp", "prod"}},
	"test_service":     {GRPC: ":8080", HTTP: ":8081"},
	"gandalf":          {MigratedEnvironments: []string{"local"}},
}

// Resources contains the platform service clients which can be
// used by the server.
//
// Cleanup must be called when the program finishes to clean up
// the underlying resources.
//
// Note that most operations require a zap.Logger, so remember to
// add a logger first before adding other resources.
//
// TODO @anhpngt: for With<Resource>C function, we might want to
// reset the sync.Once too.
type Resources struct {
	// svcName is retrieved from config's Common.Name.
	svcName string

	storage *configs.StorageConfig

	// databases databaseInitMap
	databases        map[string]*database.DBTrace
	databaseCtx      context.Context
	databaseConfigs  map[string]configs.PostgresDatabaseConfig
	databaseOnce     sync.Once
	databaseCleanups map[string]func() error
	databaser        Databaser

	// logger-related attributes
	logger         *zap.Logger
	loggerConfig   *configs.CommonConfig
	loggerOnce     sync.Once
	logLevelServer *http.Server

	// elastic-related config
	elastic       *elastic.SearchFactoryImpl
	elasticConfig *configs.ElasticSearchConfig
	elasticOnce   sync.Once
	elasticer     Elasticer

	// nats-jetstream-related attributes
	natsjs       nats.JetStreamManagement
	natsjsConfig *configs.NatsJetStreamConfig
	natsjsOnce   sync.Once
	natsjser     NATSJetstreamer

	// kafka attributes
	kafkaMgmt   kafka.KafkaManagement
	kafkaConfig *configs.KafkaClusterConfig
	kafkaOnce   sync.Once
	kafkaer     Kafkaer

	// unleash-related attributes
	unleash       unleashclient.ClientInstance
	unleashConfig *configs.UnleashClientConfig
	unleashOnce   sync.Once

	addresses          map[string]configs.ListenerConfig
	connectionCleanups map[string]func() error
}

type ResourcesOption func(resources *Resources)

// NewResources creates a new Resources.
func NewResources(opts ...ResourcesOption) *Resources {
	r := &Resources{
		databases:          make(map[string]*database.DBTrace),
		databaseCleanups:   make(map[string]func() error),
		databaser:          newDatabaseImpl(),
		elasticer:          newElasticImpl(),
		natsjser:           newNATSJetstreamImpl(),
		kafkaer:            newKafkaImpl(),
		connectionCleanups: make(map[string]func() error),
		addresses:          addresses,
	}
	for _, opt := range opts {
		opt(r)
	}

	return r
}

func WithStorage(storage *configs.StorageConfig) ResourcesOption {
	return ResourcesOption(func(r *Resources) {
		r.storage = storage
	})
}

// WithServiceName specifies name of the service this resource is intended for.
func (r *Resources) WithServiceName(name string) *Resources {
	r.svcName = name
	return r
}

// WithLogger sets the logger for this resource.
func (r *Resources) WithLogger(logger *zap.Logger) *Resources {
	r.logger = logger
	r.loggerConfig = nil
	return r
}

// WithLoggerC registers the config for this resource's logger object.
func (r *Resources) WithLoggerC(c *configs.CommonConfig) *Resources {
	r.loggerConfig = c
	r.logger = nil
	return r
}

func (r *Resources) WithDatabase(databases map[string]*database.DBTrace) *Resources {
	r.databases = databases
	r.databaseCtx = nil
	r.databaseConfigs = nil
	return r
}

// WithDatabaseC registers the config for this resource's database connections.
// Use DB or DBWith to init and retrieve the database client.
func (r *Resources) WithDatabaseC(ctx context.Context, databaseConfigs map[string]configs.PostgresDatabaseConfig) *Resources {
	r.databaseCtx = ctx
	r.databaseConfigs = databaseConfigs
	r.databases = make(map[string]*database.DBTrace)
	return r
}

// WithLogger sets the Elastic client for this resource.
func (r *Resources) WithElastic(e *elastic.SearchFactoryImpl) *Resources {
	r.elastic = e
	r.elasticConfig = nil
	return r
}

// WithElasticC registers the config for this resource's elastic client.
// Use Elastic to init and retrieve the new elastic client.
func (r *Resources) WithElasticC(c *configs.ElasticSearchConfig) *Resources {
	r.elasticConfig = c
	r.elastic = nil
	return r
}

// WithNATS sets the NATS Jetstream client for this resource.
func (r *Resources) WithNATS(n nats.JetStreamManagement) *Resources {
	r.natsjs = n
	r.natsjsConfig = nil
	return r
}

// WithNATSC registers the config for this resource's nats jetstream client.
// Use NATS to init and retrieve the new client.
func (r *Resources) WithNATSC(c *configs.NatsJetStreamConfig) *Resources {
	r.natsjsConfig = c
	r.natsjs = nil
	return r
}

// WithKafka sets the kafka client for this resource.
func (r *Resources) WithKafka(k kafka.KafkaManagement) *Resources {
	r.kafkaMgmt = k
	r.kafkaConfig = nil
	return r
}

// WithKafkaC registers the config for this resource's kafka client.
// Use Kafka to init and retrieve the new client.
func (r *Resources) WithKafkaC(c *configs.KafkaClusterConfig) *Resources {
	r.kafkaConfig = c
	r.kafkaMgmt = nil
	return r
}

// WithLogger sets the Unleash client for this resource.
func (r *Resources) WithUnleash(u unleashclient.ClientInstance) *Resources {
	r.unleash = u
	r.unleashConfig = nil
	return r
}

// WithUnleashC registers the config for this resource's Unleash client.
// Use Unleash to init and retrieve the new Unleash client.
func (r *Resources) WithUnleashC(c *configs.UnleashClientConfig) *Resources {
	r.unleashConfig = c
	r.unleash = nil
	return r
}

func (r *Resources) initLogger(c *configs.CommonConfig) *zap.Logger {
	return logger.NewZapLogger(c.Log.ApplicationLevel, c.Environment == "local")
}

// ServiceName returns the the service name. Use WithServiceName to register the service name.
func (r *Resources) ServiceName() string {
	if r.svcName == "" {
		panic(fmt.Errorf("undefined service name (hint: use WithServiceName when initializing Resources)"))
	}
	return r.svcName
}

// Logger returns the global logger. It also initializes
// the client if necessary.
func (r *Resources) Logger() *zap.Logger {
	if r.logger == nil {
		if r.loggerConfig == nil {
			panic("logger not initialized")
		}
		r.loggerOnce.Do(func() { r.logger = r.initLogger(r.loggerConfig) })
	}
	return r.logger
}

// DB is similar to DBWith, but assumes the dbname to be the same as
// the service name (from r.ServiceName).
func (r *Resources) DB() *database.DBTrace {
	if r.svcName == "" {
		panic(fmt.Errorf("undefined service name (hint: use WithServiceName when initializing Resources)"))
	}
	return r.DBWith(r.ServiceName())
}

// DBWith returns a database.DBTrace created from `postgres_v2` config. It uses
// the context provided from WithDatabaseC to initialize the connection, if necessary.
//
// It panics if database with dbname does not exist for the provided configuration.
func (r *Resources) DBWith(dbname string) *database.DBTrace {
	if len(r.databases) == 0 {
		if r.databaseConfigs == nil {
			panic("unable to init database connection: database config is not provided")
		}
		r.databaseOnce.Do(func() {
			c := r.databaseConfigs
			l := r.Logger()
			for dbname, dbconf := range c {
				dbpool, dbcancel, err := r.databaser.ConnectV2(r.databaseCtx, l, dbconf)
				if err != nil {
					panic(fmt.Errorf("failed to connect to database %s: %s", dbname, err))
				}

				r.databases[dbname] = &database.DBTrace{DB: dbpool}
				r.databaseCleanups[dbname] = dbcancel
			}
		})
	}

	db, ok := r.databases[dbname]
	if !ok {
		panic(fmt.Errorf("database %q not yet initialized", dbname))
	}
	return db
}

func (r *Resources) GRPCDial(svcName string) *grpc.ClientConn {
	svcConn, err := grpc.Dial(r.GetAddress(svcName), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(&tracer.B3Handler{
		ClientHandler: &ocgrpc.ClientHandler{},
	}))
	if err != nil {
		panic(fmt.Errorf("grpc.Dial : %s", err))
	}
	r.connectionCleanups[svcName] = func() error {
		return svcConn.Close()
	}
	return svcConn
}

func UnaryClientAttachHeaderInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	token := md.Get("token")
	if len(token) == 0 {
		return status.Error(codes.Unauthenticated, "wrong token")
	}
	newMd := metadata.New(map[string]string{"token": token[0], "version": "1.0.0", "pkg": "com.manabie.liz"})
	ctxOutGoing := metadata.NewOutgoingContext(ctx, newMd)
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

func (r *Resources) GRPCDialContext(ctx context.Context, svcName string, retryOptions configs.RetryOptions) *grpc.ClientConn {
	commonDialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock()}
	commonDialOptions = append(commonDialOptions, retryPolicy(retryOptions))

	svcConn, err := grpc.DialContext(ctx, r.GetAddress(svcName), commonDialOptions...)
	if err != nil {
		panic(fmt.Errorf("grpc.Dial : %s", err))
	}
	r.connectionCleanups[svcName] = func() error {
		return svcConn.Close()
	}
	return svcConn
}

func (r *Resources) GetAddress(svcName string) string {
	addr := getAddress(r.addresses, r.loggerConfig, svcName, "GRPC", false)
	r.Logger().Debug("GetAddress",
		zap.String("service", svcName), zap.String("address", addr))
	return addr
}

func (r *Resources) GetHTTPAddress(svcName string) string {
	force := false
	if svcName == "shamir" {
		force = true
	}
	addr := getAddress(r.addresses, r.loggerConfig, svcName, "HTTP", force)
	r.Logger().Debug("GetHTTPAddress",
		zap.String("service", svcName), zap.String("address", addr), zap.Bool("force", force))
	return addr
}

func (r *Resources) GetGRPCPort(svcName string) string {
	addr, ok := r.addresses[svcName]
	if !ok {
		panic(fmt.Errorf("not found: %s address", svcName))
	}
	if addr.GRPC == "" {
		panic(fmt.Errorf("service %s does not listen any GRPC port", svcName))
	}
	return addr.GRPC
}

func getAddress(addresses map[string]configs.ListenerConfig, c *configs.CommonConfig, svcName string, protocol string, force bool) string {
	if c == nil {
		panic("error: r.loggerConfig is nil")
	}
	addr, ok := addresses[svcName]
	if !ok {
		panic(fmt.Errorf("not found: %s address", svcName))
	}
	port, err := extract[string](addr, protocol)
	if err != nil {
		panic(fmt.Errorf("getAddress error: %s", err.Error()))
	}
	if *port == "" {
		panic(fmt.Errorf("service %s does not listen any %s port", svcName, protocol))
	}
	currentAddr, ok := addresses[c.Name]
	if !ok {
		panic(fmt.Errorf("not found: %s address", c.Name))
	}
	if !force {
		if svcName == c.Name {
			return fmt.Sprintf("%s%s", svcName, *port)
		}
		if len(addr.MigratedEnvironments) == 0 && len(currentAddr.MigratedEnvironments) == 0 {
			return fmt.Sprintf("%s%s", svcName, *port)
		}
		if slices.Contains(addr.MigratedEnvironments, c.Environment) && slices.Contains(currentAddr.MigratedEnvironments, c.Environment) {
			return fmt.Sprintf("%s%s", svcName, *port)
		}
		if !slices.Contains(addr.MigratedEnvironments, c.Environment) && !slices.Contains(currentAddr.MigratedEnvironments, c.Environment) {
			return fmt.Sprintf("%s%s", svcName, *port)
		}
	}
	var namespace string
	if c.Environment == "local" {
		if slices.Contains(addr.MigratedEnvironments, c.Environment) {
			namespace = fmt.Sprintf("%s-%s-%s", c.ActualEnvironment, c.Organization, "backend")
		} else {
			namespace = "backend"
		}
	} else {
		if slices.Contains(addr.MigratedEnvironments, c.Environment) {
			namespace = fmt.Sprintf("%s-%s-%s", c.ActualEnvironment, c.Organization, "backend")
		} else {
			namespace = fmt.Sprintf("%s-%s-%s", c.ActualEnvironment, c.Organization, "services")
		}
	}
	return fmt.Sprintf("%s.%s.svc.cluster.local%s", svcName, namespace, *port)
}

// NATS returns a NATS client created from `natsjs` config.
// It panics if not available.
func (r *Resources) NATS() nats.JetStreamManagement {
	if r.natsjs == nil {
		if r.natsjsConfig == nil {
			panic("unable to init NAT Jetstream client: NATS config is not provided")
		}
		r.natsjsOnce.Do(func() {
			c := r.natsjsConfig
			l := r.Logger()
			natsjs, err := r.natsjser.NewJetStreamManagement(l, c)
			if err != nil {
				panic(err)
			}
			natsjs.ConnectToJS()
			r.natsjs = natsjs
		})
	}
	return r.natsjs
}

// Kafka returns a Kafka client created from `kafka_cluster` config.
// It panics if not available.
// NOTED [IMPORTANT]: If using it on PROD, be careful for synersia cluster because it doesn't have any Kafka deployment (at this time this comment are written)
func (r *Resources) Kafka() kafka.KafkaManagement {
	if r.kafkaMgmt == nil {
		if r.kafkaConfig == nil {
			panic("unable to init Kafka client: Kafka config is not provided")
		}
		r.kafkaOnce.Do(func() {
			c := r.kafkaConfig
			l := r.Logger()
			kafkaMgmt, err := r.kafkaer.NewKafkaManagement(l, c)
			if err != nil {
				panic(err)
			}
			kafkaMgmt.ConnectToKafka()
			r.kafkaMgmt = kafkaMgmt
		})
	}
	return r.kafkaMgmt
}

// Unleash returns this resource's Unleash client. It also initializes
// the client if necessary.
func (r *Resources) Unleash() unleashclient.ClientInstance {
	if r.unleash == nil {
		if r.unleashConfig == nil {
			panic("unleash not initialized")
		}
		r.unleashOnce.Do(func() {
			c := r.unleashConfig
			uc, err := unleashInitF(c.URL, c.AppName, c.APIToken, r.Logger())
			if err != nil {
				panic(err)
			}
			if err := uc.ConnectToUnleashClient(); err != nil {
				panic(err)
			}
			r.unleash = uc
		})
	}
	return r.unleash
}

// Elastic returns this resource's Elastic client. It also initializes
// the client if necessary.
func (r *Resources) Elastic() *elastic.SearchFactoryImpl {
	if r.elastic == nil {
		if r.elasticConfig == nil {
			panic("unable to return a new elastic client: elastic config is not provided")
		}
		r.elasticOnce.Do(func() {
			c := r.elasticConfig
			client, err := r.elasticer.Init(r.Logger(), c.Addresses, c.Username, c.Password, "", "")
			if err != nil {
				panic(fmt.Errorf("failed to initialize elastic: %s", err))
			}
			r.elastic = client
		})
	}
	return r.elastic
}

func (r *Resources) Storage() *configs.StorageConfig {
	if r.storage == nil {
		panic("storage not initialized")
	}
	return r.storage
}

// Cleanup cleans up the underlying resources.
// It must be called when the program finishes.
func (r *Resources) Cleanup() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30) // 30s is the default grace period on k8s
	defer cancel()
	var combinedErr error

	if r.logLevelServer != nil {
		r.Logger().Debug("shutting down log level server...")
		if err := r.logLevelServer.Shutdown(ctx); err != nil {
			combinedErr = multierr.Append(combinedErr, fmt.Errorf("failed to shutdown log level server: %s", err))
		}
	}

	for dbname, f := range r.databaseCleanups {
		if err := f(); err != nil {
			combinedErr = multierr.Append(combinedErr, fmt.Errorf("failed to cleanup database %s: %s", dbname, err))
		}
	}

	for svcName, f := range r.connectionCleanups {
		if err := f(); err != nil {
			combinedErr = multierr.Append(combinedErr, fmt.Errorf("failed to cleanup connection %s: %s", svcName, err))
		}
	}

	return combinedErr
}
