package consume

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

var _ Consumer = (*KafkaConsumer)(nil)

// Consumer is an abstraction for consuming messages from a topic.
type Consumer interface {
	Consume(ctx context.Context, offset int64) (chan *kafka.Message, error)
}

// KafkaConsumer implements the Consumer interface.
type KafkaConsumer struct {
	reader *kafka.Reader
}

// NewKafkaConsumer creates a new KafkaConsumer.
func NewKafkaConsumer(brokers []string, topic string) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:   brokers,
			Topic:     topic,
			Partition: 0,
			MinBytes:  10e6, // 10MB
		}),
	}
}

func (x *KafkaConsumer) Consume(
	ctx context.Context,
	offset int64,
) (chan *kafka.Message, error) {
	if err := x.reader.SetOffset(offset); err != nil {
		return nil, err
	}
	mChan := make(chan *kafka.Message, 1)

	go func() {
		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("consume: context cancelled")
				close(mChan)
				x.reader.Close()
			default:
			}

			message, err := x.reader.ReadMessage(ctx)
			if err != nil {
				log.Error().Err(err).Msg("consume: failed to read message")
				continue
			}
			mChan <- &message
		}
	}()

	return mChan, nil
}
