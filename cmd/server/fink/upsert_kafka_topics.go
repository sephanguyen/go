package fink

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/fink/topics"
	"github.com/manabie-com/backend/internal/fink/configurations"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/kafka"

	kafka_lib "github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

func init() {
	bootstrap.RegisterJob("upsert_kafka_topics", RunUpsertTopics)
}

func RunUpsertTopics(_ context.Context, _ configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()

	err := upsertKafkaTopics(rsc.Kafka())
	if err != nil {
		zapLogger.Fatal("error when upsert topic", zap.Error(err))
		return err
	}

	zapLogger.Info("Upsert all topics have succeed")
	return nil
}

func upsertKafkaTopics(kafkaMgmt kafka.KafkaManagement) error {
	arrTopicConfig := []*kafka_lib.TopicConfig{}

	arrTopicConfig = append(arrTopicConfig, topics.GetTopicConfigSpike()...)
	arrTopicConfig = append(arrTopicConfig, topics.GetTopicConfigNotificationmgmt()...)

	// make topic name like: env.org.name, for example: prod.tokyo.topic-A
	for i := 0; i < len(arrTopicConfig); i++ {
		realTopicName := arrTopicConfig[i].Topic
		arrTopicConfig[i].Topic = kafka.GetTopicNameWithPrefix(realTopicName, kafkaMgmt.GetObjectNamePrefix())
	}

	var err error
	for i := range arrTopicConfig {
		err = kafkaMgmt.UpsertTopic(arrTopicConfig[i])
		if err != nil {
			return fmt.Errorf("failed to UpsertTopic: %s. Detail: %v", arrTopicConfig[i].Topic, err)
		}
	}
	return nil
}
