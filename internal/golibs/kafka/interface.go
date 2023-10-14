package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

type Writer interface {
	Close() error
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
}

type Reader interface {
	Close() error
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	FetchMessage(ctx context.Context) (kafka.Message, error)
	ReadMessage(ctx context.Context) (kafka.Message, error)
}

type Conn interface {
	ApiVersions() ([]kafka.ApiVersion, error)
	Close() error
	Controller() (broker kafka.Broker, err error)
	CreateTopics(topics ...kafka.TopicConfig) error
	ReadPartitions(topics ...string) (partitions []kafka.Partition, err error)
}

type Kafkaer interface {
	DialConn(network string, address string) (Conn, error)
	NewWriter(config kafka.WriterConfig) Writer
	NewReader(config kafka.ReaderConfig) Reader
}
