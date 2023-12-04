package produce

import (
	"context"

	"github.com/segmentio/kafka-go"
)

var _ Producer = (*KafkaProducer)(nil)

// Producer is an abstract interface for producing messages.
type Producer interface {
	Produce(ctx context.Context, topic string, partition int, key, value []byte) error
}

// KafkaProducer implements the Producer interface.
type KafkaProducer struct {
	writer *kafka.Writer
}

// NewKafkaProducer creates a new KafkaProducer.
func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
	return &KafkaProducer{
		writer: &kafka.Writer{
			Addr:                   kafka.TCP(brokers...),
			Topic:                  topic,
			Balancer:               &kafka.LeastBytes{},
			BatchSize:              1,
			Compression:            kafka.Snappy,
			AllowAutoTopicCreation: true,
		},
	}
}

func (x *KafkaProducer) Produce(ctx context.Context, topic string, partition int, key, value []byte) error {
	return x.writer.WriteMessages(ctx, kafka.Message{
		Topic:     topic,
		Partition: partition,
		Key:       key,
		Value:     value,
	})
}
