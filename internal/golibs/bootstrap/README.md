## Why
- DRY on common logic (resource initialization), setting up a new binary right now (new service, new batchjob) is really painful, and contains a lot of boilerplate=> save us a lot of LOC
- Centralized infrastructure management => reusable features/enrichment. For example we enforce usage of library for Postgres or Elasticsearch must be from platform provided libraries, so no team can accidentally violates the constraint such as RLS/DLS (row level security/Document level security)

### Resources

Platform automatically initializes some resources, such as:

```go
func (r *Resources) Logger() *zap.Logger {}

func (r *Resources) DB() *database.DBTrace {}

func (r *Resources) NATS() nats.JetStreamManagement {}

func (r *Resources) Unleash() unleashclient.ClientInstance {}

func (r *Resources) Elastic() *elastic.SearchFactoryImpl {}

```

=> We reduce a lot logic initializing common dependencies.

## How to setup new service
### Mandatory interface
```go

// Every service must implement this interface
type BaseServicer[T any] interface {
	// Service name and cmdline name
	ServerName() string

	// Setup all customized dependencies of your service
	// not including those managed by platform, aka whatever
	// available in Resource (Agora,Firebase,S3),
	// maybe create some complex service obj that may share between grpc/nats/http
	InitDependencies(T, *Resources) error

	// Shutdown your own dependencies, not including those managed
	// by platform
	GracefulShutdown(context.Context)
}

```
This interface is generic over the configuration struct (you must define this struct for your service).

After provide an implementation, you register your service using this syntax

```go

import "github.com/manabie-com/backend/internal/golibs/bootstrap"

type MyConfig struct{

}
type myserver struct{}

func init() {
	s := &myserver{}
	bootstrap.
		WithGRPC[MyConfig](s).
		// WithNatsServicer(s).
		// WithMonitorServicer(s).
		Register(s)
}
```

This syntax allows the bootstrap package to know about your server and bootstrap dynamically based on the server name (also the command line name), for example your server is bob, then it impl something like
```go

func (*server) ServerName() string {
	return "bob"
}
```

And the container argument will init this service using the cmd:

```
gserver bob ...
```

There are other interfaces you can implement, depends on which protocol your server uses. Read more about the interface in our code
```go
type GRPCServicer[T any] interface {
	// Register your implementation to the provided grpc server
	// You don't need to start any net Listener
	SetupGRPC(*grpc.Server, T, *Resources) error
	WithUnaryServerInterceptors(T, *Resources) []grpc.UnaryServerInterceptor
	WithStreamServerInterceptors(T, *Resources) []grpc.StreamServerInterceptor
	WithServerOptions() []grpc.ServerOption
}
type NatsServicer[T any] interface{}
type HTTPServicer[T any] interface{}
```
Notice that Nats is also considered a "protocol", it is also some way the outside world comes into your application

For detailed implementation, check this file:
https://github.com/manabie-com/backend/blob/ae4b3a8771897e4f1a8ca9437200c31a19dc50dd/cmd/server/shamir/gserver.go
### Customization

Monitoring service
```go

// MonitorServicer Implement this interface to expose
// a metrics api to monitor your service (usually exposed at port 8888)
type MonitorServicer[T any] interface {
	// Provide us with your own metric collector, if any
	WithPrometheusCollectors(*Resources) []prometheus.Collector
	// Provide us with your own opencensus, if any
	WithOpencensusViews() []*view.View
}
```

Most of the time you will want to implement this interface, if you don't know how to implement it, just embed this struct into your server like

```go
type MyConfig struct{}

type myserver struct {
    bootstrap.DefaultMonitorService[MyConfig] 
}
```
This will still make your app expose metrics, but metrics defined defaul by platform. You can.provide your own anytime using this interface

## How to create a new batch job

Batch job is still a binary but it has a short life time, we used alot of migration script using batch job, so using generics can help us reduce a lot of boilerplating code.

### Registration syntax

```go
	bootstrap.RegisterJob("bob_sync_lesson_conversations", RunSyncLessonConversation).
		StringVar(&schoolID, "schoolID", "", "sync for specific school").
		StringVar(&schoolName, "schoolName", "", "should match with school name in secret config, for sanity check")
```

The syntax is straightforward, you only needs to provide the name of the job, and the Execute function to run the job, we provide all the basic resources you need to run the job

```go
func RunSyncLessonConversation(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
```

You can further specify the flag your job can receive. This types of registration decentralize all the variable bindings, making the code more readable. In the old style we used to have a bunch of global variable inside the main.go file like

```
var rootCmd = &cobra.Command{}

func makeRootCmd() {
	var (
		configPath       string
		commonConfigPath string
		secretsPath      string
		migratePath      string
		renewESIndex     bool

		// only batch job commands use those
		schoolID, schoolName string

		// only used for migrate student enrollment status
		newStatus, originStatus, resourcePath string

		// only used for migrate resource_path in eureka db
		bobConfigPath       string
		bobCommonConfigPath string
		bobSecretPath       string

		userID, createdAt, organizationID string
```
