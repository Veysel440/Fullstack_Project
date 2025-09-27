package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

type DBExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type RetryConfig struct {
	Brokers      []string
	DLQTopic     string
	Group        string
	MaxAttempts  int
	BackoffStart time.Duration
	Logger       *slog.Logger
	DB           DBExec
}

type RetryWorker struct {
	r   *kafka.Reader
	w   *kafka.Writer
	log *slog.Logger
	rc  RetryConfig
}

func NewRetryWorker(rc RetryConfig) *RetryWorker {
	lg := rc.Logger
	if lg == nil {
		lg = slog.Default()
	}
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        rc.Brokers,
		GroupID:        rc.Group,
		Topic:          rc.DLQTopic,
		MinBytes:       1e3,
		MaxBytes:       10e6,
		CommitInterval: time.Second,
	})
	w := &kafka.Writer{
		Addr:         kafka.TCP(rc.Brokers...),
		Topic:        rc.DLQTopic,
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 200 * time.Millisecond,
	}
	return &RetryWorker{r: r, w: w, log: lg, rc: rc}
}

func (w *RetryWorker) Close() {
	_ = w.r.Close()
	_ = w.w.Close()
}

func (w *RetryWorker) Run(ctx context.Context) error {
	defer w.Close()
	for {
		msg, err := w.r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			w.log.Error("dlq_read", "err", err)
			continue
		}
		attempt := headerInt(msg.Headers, "x-attempts")
		if attempt < 1 {
			attempt = 1
		}

		if err := w.processOnce(ctx, msg.Value); err != nil {
			if attempt >= w.rc.MaxAttempts {
				w.log.Error("dlq_giveup", "attempts", attempt, "err", err)
				continue
			}
			delay := w.rc.BackoffStart * (1 << (attempt - 1))
			time.Sleep(delay)
			attempt++
			msg.Headers = upsertHeader(msg.Headers, "x-attempts", strconv.Itoa(attempt))
			_ = w.w.WriteMessages(ctx, msg)
			continue
		}
	}
}

func (w *RetryWorker) processOnce(ctx context.Context, payload []byte) error {
	var ev itemEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}
	if w.rc.DB == nil {
		return nil
	}
	_, err := w.rc.DB.ExecContext(ctx,
		`INSERT INTO app.item_audit(evt_type, payload) VALUES ($1,$2)`,
		ev.Type, json.RawMessage(payload),
	)
	return err
}

func headerInt(h []kafka.Header, key string) int {
	for _, x := range h {
		if x.Key == key {
			i, _ := strconv.Atoi(string(x.Value))
			return i
		}
	}
	return 0
}

func upsertHeader(h []kafka.Header, k, v string) []kafka.Header {
	for i := range h {
		if h[i].Key == k {
			h[i].Value = []byte(v)
			return h
		}
	}
	return append(h, kafka.Header{Key: k, Value: []byte(v)})
}
