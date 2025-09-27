package events

import (
	"context"
	"os"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
)

type Writer struct {
	w     *kafka.Writer
	topic string
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func NewWriter() *Writer {
	brokers := splitCSV(os.Getenv("KAFKA_BROKERS"))
	topic := os.Getenv("KAFKA_TOPIC_ITEMS")
	if topic == "" {
		topic = os.Getenv("KAFKA_ITEMS_TOPIC")
	}
	if len(brokers) == 0 || topic == "" {
		return nil
	}
	return &Writer{
		w: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
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
