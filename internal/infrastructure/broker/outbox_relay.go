package broker

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type OutboxRelay struct {
	db       *sql.DB
	producer Producer
	log      *zap.Logger
	interval time.Duration
}

type Producer interface {
	Publish(ctx context.Context, topic string, payload []byte) error
}

func NewOutboxRelay(db *sql.DB, producer Producer, log *zap.Logger) *OutboxRelay {
	return &OutboxRelay{db: db, producer: producer, log: log, interval: 3 * time.Second}
}

// Start menjalankan polling loop. Hentikan via context cancellation.
func (r *OutboxRelay) Start(ctx context.Context) error {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	r.log.Info("outbox relay started", zap.Duration("interval", r.interval), zap.Int("batch_size", 50))
	for {
		select {
		case <-ctx.Done():
			r.log.Info("outbox relay stopped")
			return ctx.Err()
		case <-ticker.C:
			r.processPending(ctx)
		}
	}
}

func (r *OutboxRelay) processPending(ctx context.Context) {
	start := time.Now()
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, topic, payload FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT 50
		FOR UPDATE SKIP LOCKED
	`)
	if err != nil {
		r.log.Error("query outbox failed", zap.Error(err))
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id, topic, payload string
		if err := rows.Scan(&id, &topic, &payload); err != nil {
			r.log.Warn("scan outbox row failed", zap.Error(err))
			continue
		}
		ids = append(ids, id)

		if err := r.producer.Publish(ctx, topic, []byte(payload)); err != nil {
			r.log.Error("publish outbox failed",
				zap.String("id", id),
				zap.String("topic", topic),
				zap.Error(err))
			continue
		}

		r.db.ExecContext(ctx, `UPDATE outbox_events SET status = 'processed', processed_at = NOW() WHERE id = $1`, id)
		r.log.Info("outbox event published",
			zap.String("id", id),
			zap.String("topic", topic),
			zap.String("payload_size", fmt.Sprintf("%d bytes", len(payload))))
	}

	if len(ids) > 0 {
		r.log.Info("outbox batch complete",
			zap.Int("processed", len(ids)),
			zap.String("duration", time.Since(start).String()))
	}
}
