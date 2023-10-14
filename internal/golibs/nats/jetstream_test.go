package nats

import (
	"errors"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_CheckDurName(t *testing.T) {
	t.Run("test with valid durable name format", func(t *testing.T) {
		err := checkDurName("durable-test")
		assert.NoError(t, err)
	})
	t.Run("test with invalid durable name format", func(t *testing.T) {
		err := checkDurName("durable-test.something")
		assert.Equal(t, err, errors.New("nats: invalid durable name"))
	})
}

// test to make sure that each element of JetStreamOptions is match with ConsumerConfig
func TestSubOption(t *testing.T) {
	t.Run("Create consumer with common config", func(t *testing.T) {
		option := Option{
			JetStreamOptions: []JSSubOption{
				ManualAck(),
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				DeliverSubject("deliver.subjecttest"),
				AckExplicit(),
				AckWait(2 * time.Second),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckExplicitPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, 2*time.Second, o.consumerConfig.AckWait)
		assert.Equal(t, nats.DeliverAllPolicy, o.consumerConfig.DeliverPolicy)
	})
	t.Run("Create consumer with Starttime config", func(t *testing.T) {
		now := time.Now()
		option := Option{
			JetStreamOptions: []JSSubOption{
				ManualAck(),
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				DeliverSubject("deliver.subjecttest"),
				StartTime(now),
				AckExplicit(),
				AckWait(2 * time.Second),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckExplicitPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, &now, o.consumerConfig.OptStartTime)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, 2*time.Second, o.consumerConfig.AckWait)
		assert.Equal(t, nats.DeliverByStartTimePolicy, o.consumerConfig.DeliverPolicy)
	})

	t.Run("Create consumer with AckAll policy", func(t *testing.T) {
		now := time.Now()
		option := Option{
			JetStreamOptions: []JSSubOption{
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				AckAll(),
				DeliverSubject("deliver.subjecttest"),
				StartTime(now),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckAllPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, &now, o.consumerConfig.OptStartTime)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, nats.DeliverByStartTimePolicy, o.consumerConfig.DeliverPolicy)
	})

	t.Run("Create consumer with DeliverNew policy", func(t *testing.T) {
		option := Option{
			JetStreamOptions: []JSSubOption{
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				AckAll(),
				DeliverSubject("deliver.subjecttest"),
				DeliverNew(),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckAllPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, nats.DeliverNewPolicy, o.consumerConfig.DeliverPolicy)
	})

	t.Run("Create consumer with DeliverAll policy", func(t *testing.T) {
		option := Option{
			JetStreamOptions: []JSSubOption{
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				AckAll(),
				DeliverSubject("deliver.subjecttest"),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckAllPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, nats.DeliverAllPolicy, o.consumerConfig.DeliverPolicy)
	})

	t.Run("Create a consumer with DeliverLastPolicy policy", func(t *testing.T) {
		option := Option{
			JetStreamOptions: []JSSubOption{
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				AckAll(),
				DeliverSubject("deliver.subjecttest"),
				DeliverLast(),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckAllPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, nats.DeliverLastPolicy, o.consumerConfig.DeliverPolicy)
	})

	t.Run("Create a consumer with DeliverByStartSequencePolicy policy", func(t *testing.T) {
		seq := 1000
		option := Option{
			JetStreamOptions: []JSSubOption{
				MaxDeliver(10),
				Bind("streamtest", "consumertest"),
				AckAll(),
				DeliverSubject("deliver.subjecttest"),
				StartSequence(uint64(seq)),
			},
		}

		consumerConfig := &nats.ConsumerConfig{}
		o := jsSubOptions{consumerConfig: consumerConfig}

		for _, v := range option.JetStreamOptions {
			if err := v.configureSubscribeOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, "consumertest", o.consumerConfig.Durable)
		assert.Equal(t, "deliver.subjecttest", o.consumerConfig.DeliverSubject)
		assert.Equal(t, 10, o.consumerConfig.MaxDeliver)
		assert.Equal(t, nats.AckAllPolicy, o.consumerConfig.AckPolicy)
		assert.Equal(t, "streamtest", o.streamName)
		assert.Equal(t, nats.DeliverByStartSequencePolicy, o.consumerConfig.DeliverPolicy)
		assert.Equal(t, uint64(seq), consumerConfig.OptStartSeq)
	})
}

func TestNewJetStreamManagement(t *testing.T) {
	t.Run("Fail due to miss url", func(t *testing.T) {
		_, err := NewJetStreamManagement("", "user", "password", 1, time.Second, false, &zap.Logger{})
		assert.Equal(t, "missing url", err.Error())
	})

	t.Run("Fail due to miss user", func(t *testing.T) {
		_, err := NewJetStreamManagement("url", "", "password", 1, time.Second, false, &zap.Logger{})
		assert.Equal(t, "missing user", err.Error())
	})
	t.Run("Fail due to miss password", func(t *testing.T) {
		_, err := NewJetStreamManagement("url", "user", "", 1, time.Second, false, &zap.Logger{})
		assert.Equal(t, "missing password", err.Error())
	})
	t.Run("Fail due to miss logger", func(t *testing.T) {
		_, err := NewJetStreamManagement("url", "user", "password", 1, time.Second, false, nil)
		assert.Equal(t, "missing logger", err.Error())
	})
}
