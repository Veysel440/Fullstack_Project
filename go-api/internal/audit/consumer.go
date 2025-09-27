package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers   []string
	Topic     string
	Group     string
	DeadTopic string
	DB        *sql.DB
	Logger    *slog.Logger
}

type Consumer struct {
	r      *kafka.Reader
	dlq    *kafka.Writer
	db     *sql.DB
	logger *slog.Logger
}

type itemEvent struct {
	Type string          `json:"type"`
	Item json.RawMessage `json:"item,omitempty"`
	ID   int64           `json:"id,omitempty"`
}

func NewConsumer(c Config) (*Consumer, error) {
	lg := c.Logger
	if lg == nil {
		lg = slog.Default()
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        c.Brokers,
		GroupID:        c.Group,
		Topic:          c.Topic,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	w := &kafka.Writer{
		Addr:         kafka.TCP(c.Brokers...),
		Topic:        c.DeadTopic,
		RequiredAcks: kafka.RequireAll,
	}
	return &Consumer{r: r, dlq: w, db: c.DB, logger: lg}, nil
}

func (c *Consumer) Close() {
	_ = c.r.Close()
	_ = c.dlq.Close()
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.Close()
	for {
		m, err := c.r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("kafka_read", "err", err)
			continue
		}
		if err := c.handle(ctx, m.Value); err != nil {
			c.logger.Error("audit_handle", "err", err)
			_ = c.dlq.WriteMessages(ctx, kafka.Message{Value: m.Value})
			continue
		}
	}
}

func (c *Consumer) handle(ctx context.Context, b []byte) error {
	var ev itemEvent
	if err := json.Unmarshal(b, &ev); err != nil {
		return err
	}
	_, err := c.db.ExecContext(ctx,
		`INSERT INTO app.item_audit(evt_type, payload) VALUES ($1,$2)`,
		ev.Type, json.RawMessage(b),
	)
	return err
}
