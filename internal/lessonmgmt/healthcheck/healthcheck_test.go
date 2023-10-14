package healthcheck

import (
	"context"
	"testing"

	mockhc "github.com/manabie-com/backend/mock/golibs/healthcheck"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	health "google.golang.org/grpc/health/grpc_health_v1"
)

func TestHealthCheckService_Check(t *testing.T) {
	db := &mockhc.Pinger{}
	s := Service{
		DBBob:    db,
		DBLesson: db,
	}
	req := &health.HealthCheckRequest{
		Service: "virtualclassroom",
	}
	db.On("Ping", mock.Anything).Return(nil)

	res, err := s.Check(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, health.HealthCheckResponse_SERVING, res.Status)
}
