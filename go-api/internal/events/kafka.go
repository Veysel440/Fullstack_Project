package events

import (
	"context"
	"os"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Writer struct {
	w     *kafka.Writer
	topic string
}

func NewWriter() *Writer {
	brokers := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC_ITEMS")
	if brokers == "" || topic == "" {
		return nil
	}
	return &Writer{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: -1,
			BatchTimeout: 50 * time.Millisecond,
		},
		topic: topic,
	}
}

func (w *Writer) Close() {
	if w != nil && w.w != nil {
		_ = w.w.Close()
	}
}

func (w *Writer) Publish(ctx context.Context, key string, value []byte) error {
	if w == nil || w.w == nil {
		return nil
	}
	return w.w.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: value})
}
