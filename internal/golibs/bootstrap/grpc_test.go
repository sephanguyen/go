package bootstrap

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetupGRPCService(t *testing.T) {
	t.Parallel()
	conf := &Config{}
	rsc := NewResources()
	b := bootstrapper[Config]{}
	ctx := context.Background()

	mockSvc := new(MockAllService[Config])
	mockSvc.On("WithUnaryServerInterceptors", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("WithStreamServerInterceptors", mock.Anything, mock.Anything).Return(nil)
	mockSvc.On("WithServerOptions").Return(nil)

	t.Run("init success", func(t *testing.T) {
		mockSvc.On("SetupGRPC", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil)
		s, err := b.setupGRPCService(ctx, mockSvc, conf, rsc)

		assert.NoError(t, err)
		assert.NotNil(t, s)
		mock.AssertExpectationsForObjects(t, mockSvc)
	})

	t.Run("init fail", func(t *testing.T) {
		mockSvc.On("SetupGRPC", ctx, mock.Anything, mock.Anything, mock.Anything).Once().Return(fmt.Errorf("init fail"))
		s, err := b.setupGRPCService(ctx, mockSvc, conf, rsc)

		assert.Error(t, err)
		assert.EqualError(t, err, "setupGRPCService: init fail")
		assert.Nil(t, s)
		mock.AssertExpectationsForObjects(t, mockSvc)
	})
}
