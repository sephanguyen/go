package subscriptions

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	nats_golib "github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/zeus/configurations"
	mock_nats "github.com/manabie-com/backend/mock/golibs/nats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestActivityLogCreatedEventSubscriber_Subscribe(t *testing.T) {
	t.Parallel()
	jsm := &mock_nats.JetStreamManagement{}
	s := ActivityLogCreatedEventSubscriber{
		CentralizeLogsService: nil,
		JSM:                   jsm,
		Logger:                nil,
		Configs: &configurations.Config{
			NatsJS: configs.NatsJetStreamConfig{
				DefaultAckWait: time.Minute,
				MaxRedelivery:  5,
			},
		},
	}

	jsm.
		On("QueueSubscribe",
			constants.SubjectActivityLogCreated,
			constants.QueueActivityLogCreated,
			mock.Anything,
			mock.Anything,
		).
		Once().
		Return(&nats_golib.Subscription{}, nil)

	err := s.Subscribe()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, jsm)
}

func TestActivityLogCreatedEventSubscriber_Pull(t *testing.T) {
	t.Parallel()
	jsm := &mock_nats.JetStreamManagement{}
	s := ActivityLogCreatedEventSubscriber{
		CentralizeLogsService: nil,
		JSM:                   jsm,
		Logger:                nil,
		Configs: &configurations.Config{
			NatsJS: configs.NatsJetStreamConfig{
				DefaultAckWait: time.Minute,
				MaxRedelivery:  5,
			},
		},
	}

	jsm.On("PullSubscribe",
		constants.SubjectActivityLogCreated,
		constants.DurableActivityLogCreatedPull,
		mock.Anything,
		mock.Anything).Return(nil)

	err := s.PullConsumer()
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, jsm)
}
