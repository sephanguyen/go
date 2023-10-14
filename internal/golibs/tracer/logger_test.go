package tracer

import (
	"testing"

	"github.com/manabie-com/backend/internal/golibs/nats"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestUnaryActivityLogRequestInterceptor(t *testing.T) {
	var jsm nats.JetStreamManagement
	resp := UnaryActivityLogRequestInterceptor(jsm, &zap.Logger{}, "test_rp")
	assert.NotEmpty(t, resp)
}
