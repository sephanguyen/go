package kafka

import "github.com/segmentio/kafka-go"

// kafkaerImpl implements Kafkaer.
type kafkaerImpl struct{}

func newKafkaImpl() *kafkaerImpl {
	return &kafkaerImpl{}
}

func (n *kafkaerImpl) NewWriter(config kafka.WriterConfig) Writer {
	writerClient := kafka.NewWriter(config)
	return writerClient
}

func (n *kafkaerImpl) NewReader(config kafka.ReaderConfig) Reader {
	readerClient := kafka.NewReader(config)
	return readerClient
}

func (n *kafkaerImpl) DialConn(network string, address string) (Conn, error) {
	return kafka.Dial(network, address)
}
