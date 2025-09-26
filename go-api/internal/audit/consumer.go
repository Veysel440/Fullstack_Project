package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/segmentio/kafka-go"
)

type Config struct {
	Brokers   []string
	Topic     string
	Group     string
	DeadTopic string
	DB        *sql.DB
}

type Consumer struct {
	r   *kafka.Reader
	dlq *kafka.Writer
	db  *sql.DB
}

func NewConsumer(c Config) (*Consumer, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.Brokers,
		GroupID:  c.Group,
		Topic:    c.Topic,
		MinBytes: 1e3, MaxBytes: 10e6,
		MaxWait: 2 * time.Second,
	})
	w := &kafka.Writer{Addr: kafka.TCP(c.Brokers...), Topic: c.DeadTopic, RequiredAcks: kafka.RequireAll}
	return &Consumer{r: r, dlq: w, db: c.DB}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	defer c.r.Close()
	defer c.dlq.Close()

	for {
		m, err := c.r.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		if err := c.handle(ctx, m.Value); err != nil {
			_ = c.dlq.WriteMessages(ctx, kafka.Message{Value: m.Value})
		}
	}
}

type evt struct {
	Type string          `json:"type"`
	Any  json.RawMessage `json:"item"`
	ID   any             `json:"id"`
}

func (c *Consumer) handle(ctx context.Context, b []byte) error {
	var e evt
	if err := json.Unmarshal(b, &e); err != nil {
		return err
	}
	_, err := c.db.ExecContext(ctx,
		`INSERT INTO app.item_audit(evt_type, payload) VALUES($1,$2)`,
		e.Type, json.RawMessage(b),
	)
	return err
}
