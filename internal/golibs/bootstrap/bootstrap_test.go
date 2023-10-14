package bootstrap

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func mockResource() *Resources {
	r := &Resources{
		logger: zap.NewNop(),
	}
	return r
}

type Config struct {
	Common configs.CommonConfig
}

func Test_bootstrapper_initServer(t *testing.T) {
	t.Parallel()
	mockSvc := &MockAllService[Config]{}
	conf := &Config{Common: configs.CommonConfig{Name: "test_service"}}
	rsc := mockResource()
	ctx := context.Background()
	mockSvc.On("InitDependencies", *conf, rsc).Once().Return(nil)
	mockSvc.On("WithOpencensusViews").Once().Return(nil)
	mockSvc.On("WithStreamServerInterceptors", *conf, rsc).Once().Return(nil)
	mockSvc.On("WithUnaryServerInterceptors", *conf, rsc).Once().Return(nil)
	mockSvc.On("WithServerOptions").Once().Return(nil)
	mockSvc.On("SetupGRPC", ctx, mock.Anything, *conf, rsc).Once().Return(nil)
	mockSvc.On("SetupHTTP", *conf, mock.Anything, rsc).Once().Return(nil)
	mockSvc.On("RegisterNatsSubscribers", ctx, *conf, rsc).Once().Return(nil)
	mockSvc.On("InitKafkaConsumers", ctx, *conf, rsc).Once().Return(nil)
	b := &bootstrapper[Config]{
		config: conf,
	}
	err := b.initServer(ctx, mockSvc, rsc)
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, mockSvc)
}

func Test_bootstrapper_serve(t *testing.T) {
	t.Parallel()
	run := func(t *testing.T, shutdownType string) {
		rsc := mockResource()
		b := &bootstrapper[Config]{
			config: &Config{Common: configs.CommonConfig{Name: "test_service"}},
		}
		mockServer := &MockAllService[Config]{}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockServer.On("GracefulShutdown", mock.Anything).Once().Return()

		httpServer := &http.Server{Addr: ":8081", Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		})}
		b.httpServer = httpServer
		grpcserver := grpc.NewServer()
		b.grpcserver = grpcserver
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := b.serve(ctx, mockServer, rsc)
			// grpc server on shutdown does not return err
			assert.NoError(t, err)
		}()
		time.Sleep(3 * time.Second)

		// graceful shutdown signal from os
		switch shutdownType {
		case "os":
			cancel()
		case "http":
			// httpserver shutdown somehow
			httpServer.Shutdown(context.Background())
		case "grpc":
			// grpserver shutdown somehow
			b.grpcserver.GracefulStop()
		default:
			assert.Fail(t, "unknown shutdown type", "%s", shutdownType)
		}
		wg.Wait()
		mock.AssertExpectationsForObjects(t, mockServer)
	}
	t.Run("serve shutdown with os signal", func(t *testing.T) { run(t, "os") })
	t.Run("serve shutdown with error from http", func(t *testing.T) { run(t, "http") })
	t.Run("serve shutdown with error from grpc", func(t *testing.T) { run(t, "grpc") })
}
