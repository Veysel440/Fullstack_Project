package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strings"
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
	brokers := c.Brokers
	if len(brokers) == 1 && strings.Contains(brokers[0], ",") {
		brokers = strings.Split(brokers[0], ",")
		for i := range brokers {
			brokers[i] = strings.TrimSpace(brokers[i])
		}
	}

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        c.Group,
		Topic:          c.Topic,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
		MaxWait:        2 * time.Second,
	})

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        c.DeadTopic,
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 200 * time.Millisecond,
	}

	lg := c.Logger
	if lg == nil {
		lg = slog.Default()
	}

	return &Consumer{r: r, dlq: w, db: c.DB, logger: lg}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.r.Close()
	defer c.dlq.Close()

	for {
		m, err := c.r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			c.logger.Error("kafka_read", "err", err)
			continue
		}
		if err := c.process(ctx, m.Value); err != nil {
			c.logger.Error("audit_process", "err", err)
			_ = c.dlq.WriteMessages(ctx, kafka.Message{Value: m.Value})
		}
	}
}

func (c *Consumer) process(ctx context.Context, payload []byte) error {
	var ev itemEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}

	if c.db == nil {
		c.logger.Info("audit", "type", ev.Type, "payload", string(payload))
		return nil
	}

	_, err := c.db.ExecContext(ctx,
		`INSERT INTO app.item_audit(evt_type, payload) VALUES ($1,$2)`,
		ev.Type, json.RawMessage(payload),
	)
	return err
}
