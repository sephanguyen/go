package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"cloud.google.com/go/profiler"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/spf13/cobra"
	"go.opencensus.io/plugin/ochttp"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type bootstrapper[T any] struct {
	grpcserver *grpc.Server
	httpServer *http.Server
	config     *T
	tp         *tracesdk.TracerProvider

	unaryInterceptors []grpc.UnaryServerInterceptor
}

func newBootstrapper[T any]() *bootstrapper[T] {
	return &bootstrapper[T]{}
}

func (b *bootstrapper[T]) runJob(cmd *cobra.Command, runJob JobFunc[T]) error {
	ctx := cmd.Context()

	var err error
	b.config, err = b.commandConfig(cmd)
	if err != nil {
		return err
	}

	err = b.startProfiler(b.config)
	if err != nil {
		return fmt.Errorf("startProfiler failed: %w", err)
	}

	rsc, err := b.initDeps(ctx, b.config)
	if err != nil {
		return fmt.Errorf("initDeps failed: %w", err)
	}
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, os.Kill)
	defer cancel()

	defer b.shutdownPlatformDependencies(rsc)
	ctx = ctxzap.ToContext(ctx, rsc.Logger()) // so that we can use the logger from the context easily
	err = runJob(ctx, *b.config, rsc)
	return err
}

// run is the entrypoint function when running servers.
func (b *bootstrapper[T]) run(cmd *cobra.Command, s BaseServicer[T]) error {
	ctx := cmd.Context()

	var err error
	b.config, err = b.commandConfig(cmd)
	if err != nil {
		return err
	}

	err = b.startProfiler(b.config)
	if err != nil {
		return fmt.Errorf("startProfiler failed: %w", err)
	}

	rsc, err := b.initDeps(ctx, b.config)
	if err != nil {
		return fmt.Errorf("initDeps failed: %w", err)
	}

	err = b.initServer(ctx, s, rsc)
	if err != nil {
		return fmt.Errorf("initServer failed: %w", err)
	}
	return b.serve(ctx, s, rsc)
}

// commandConfig parses flags, loads and merged the service's configuration.
func (b *bootstrapper[T]) commandConfig(cmd *cobra.Command) (*T, error) {
	commonConfigPath, err := cmd.Flags().GetString(commonConfigPathFlag)
	if err != nil {
		return nil, err
	}
	configPath, err := cmd.Flags().GetString(configPathFlag)
	if err != nil {
		return nil, err
	}
	secretPath, err := cmd.Flags().GetString(secretPathFlag)
	if err != nil {
		return nil, err
	}
	return configs.LoadAll[T](commonConfigPath, configPath, secretPath)
}

// startProfiler starts the GCP profiler.
func (b *bootstrapper[T]) startProfiler(c *T) error {
	cc, err := extract[configs.CommonConfig](c, commonFieldName)
	if err != nil {
		return fmt.Errorf("failed to extract configs.Common: %w", err)
	}
	err = profiler.Start(profiler.Config{
		ProjectID:      cc.GoogleCloudProject,
		Service:        fmt.Sprintf("%s-%s-%s", cc.Environment, cc.Organization, cc.Name),
		ServiceVersion: cc.ImageTag,
	})
	if err != nil {
		// it's okay to not have profiler started
		log.Println("failed to start profiler: ", err)
	}
	return nil
}

func (b *bootstrapper[T]) initDeps(ctx context.Context, c *T) (*Resources, error) {
	rsc := NewResources()
	if err := b.initLogger(c, rsc); err != nil {
		return nil, err
	}

	if err := initDatabase(ctx, c, rsc); err != nil {
		return nil, err
	}

	if err := initnats(ctx, c, rsc); err != nil {
		return nil, err
	}

	if err := initKafka(ctx, c, rsc); err != nil {
		return nil, err
	}

	if err := initUnleash(ctx, c, rsc); err != nil {
		return nil, err
	}

	if err := initElastic(ctx, c, rsc); err != nil {
		return nil, err
	}

	if err := initStorage(c, rsc); err != nil {
		return nil, err
	}

	rsc.addresses = addresses

	return rsc, nil
}

func (b *bootstrapper[T]) initServer(ctx context.Context, s BaseServicer[T], rsc *Resources) error {
	var err error
	baseServicer, ok := interface{}(s).(BaseServicer[T])
	if !ok {
		return errNotBaseService
	}
	err = baseServicer.InitDependencies(*b.config, rsc)
	if err != nil {
		return err
	}

	monitorServicer, ok := interface{}(s).(MonitorServicer[T])
	if ok {
		err = b.setupMonitor(monitorServicer, b.config, rsc)
		if err != nil {
			return err
		}
	}

	grpcServicer, ok := interface{}(s).(GRPCServicer[T])

	if ok {
		b.grpcserver, err = b.setupGRPCService(ctx, grpcServicer, b.config, rsc)
		if err != nil {
			return err
		}
	}

	httpServicer, ok := interface{}(s).(HTTPServicer[T])

	if ok {
		gin, err := b.setupHTTPService(httpServicer, b.config, rsc)
		if err != nil {
			return err
		}
		httpAddr, err := b.httpAddr(b.config)
		if err != nil {
			return err
		}
		httpServer := &http.Server{Addr: httpAddr, Handler: &ochttp.Handler{Handler: gin}, ReadHeaderTimeout: 5 * time.Second}
		b.httpServer = httpServer
	}

	natsServicer, ok := interface{}(s).(NatsServicer[T])
	if ok {
		err = natsServicer.RegisterNatsSubscribers(ctx, *b.config, rsc)
		if err != nil {
			return err
		}
	}

	kafkaServicer, ok := interface{}(s).(KafkaServicer[T])
	if ok {
		err = kafkaServicer.InitKafkaConsumers(ctx, *b.config, rsc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *bootstrapper[T]) shutdownPlatformDependencies(rsc *Resources) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	// shutdown telemetry
	if b.tp != nil {
		err := b.tp.Shutdown(ctx)
		if err != nil {
			rsc.Logger().Error("failed to shutdown telemetry", zap.Error(err))
		}
	}

	if rsc.natsjs != nil {
		rsc.natsjs.Close()
	}

	if rsc.kafkaMgmt != nil {
		rsc.kafkaMgmt.Close()
	}

	if err := rsc.Cleanup(); err != nil {
		rsc.Logger().Error("failed to clean up resources", zap.Error(err))
	}
	if b.grpcserver != nil {
		b.grpcserver.GracefulStop()
	}

	if b.httpServer != nil {
		err := b.httpServer.Shutdown(ctx)
		if err != nil {
			rsc.Logger().Error("b.httpServer.Shutdown", zap.Error(err))
		}
	}
}

func (b *bootstrapper[T]) serve(ctx context.Context, s BaseServicer[T], rsc *Resources) error {
	rsc.Logger().Info("starting generic server")

	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	go func() {
		<-ctx.Done()
		rsc.Logger().Warn("shutting down server...")

		// graceful shutdown for each service
		b.gracefulShutdown(ctx, s, rsc)
		b.shutdownPlatformDependencies(rsc)
	}()
	var errGr errgroup.Group

	if b.httpServer != nil {
		errGr.Go(func() error {
			defer cancel()
			err := b.httpServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}
			rsc.Logger().Error("httpServer.ListenAndServe finished", zap.Error(err))
			return err
		})
	}

	if b.grpcserver != nil {
		errGr.Go(func() error {
			defer cancel()
			grpcAddr, err := b.grpcAddr(b.config)
			if err != nil {
				return err
			}
			lis, err := net.Listen("tcp", grpcAddr)
			if err != nil {
				return fmt.Errorf("failed to listen: %w", err)
			}
			return b.grpcserver.Serve(lis)
		})
	}
	// if we have 2 servers grpc and http, and one fails with some error,
	// this function wait for the other to stop as well, to do that, we must
	// call cancel() for all 2 goroutines, so it can call the "shutdown goroutines"
	// and stops the other
	return errGr.Wait()
}

func (b *bootstrapper[T]) httpAddr(config *T) (string, error) {
	cc, err := extract[configs.CommonConfig](config, commonFieldName)
	if err != nil {
		return "", fmt.Errorf("failed to extract CommonConfig: %w", err)
	}
	addr, ok := addresses[cc.Name]
	if !ok {
		return "", fmt.Errorf("not found: %s http address", cc.Name)
	}
	if addr.HTTP == "" {
		return "", fmt.Errorf("service %s does not listen any HTTP port", cc.Name)
	}
	return addr.HTTP, nil
}

func (b *bootstrapper[T]) grpcAddr(config *T) (string, error) {
	cc, err := extract[configs.CommonConfig](config, commonFieldName)
	if err != nil {
		return "", fmt.Errorf("failed to extract CommonConfig: %w", err)
	}
	addr, ok := addresses[cc.Name]
	if !ok {
		return "", fmt.Errorf("not found: %s grpc address", cc.Name)
	}
	if addr.GRPC == "" {
		return "", fmt.Errorf("service %s does not listen any GRPC port", cc.Name)
	}
	return addr.GRPC, nil
}
