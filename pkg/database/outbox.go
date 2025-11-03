package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// OutboxEvent 表示 outbox_events 表中的一条记录。
type OutboxEvent struct {
	ID            int64
	EventID       string
	AggregateID   string
	AggregateType string
	EventType     string
	Payload       string
	RetryCount    int
	Published     bool
	PublishedAt   *time.Time
	AvailableAt   time.Time
	CreatedAt     time.Time
}

// NewOutboxEvent 创建带默认值的 OutboxEvent。
func NewOutboxEvent() *OutboxEvent {
	return &OutboxEvent{
		EventID:    uuid.NewString(),
		RetryCount: 0,
		Published:  false,
		CreatedAt:  time.Now(),
	}
}

// OutboxRepository 定义 outbox 的持久化接口。
type OutboxRepository interface {
	Save(ctx context.Context, tx Transaction, event *OutboxEvent) error
	GetUnpublishedForUpdate(ctx context.Context, tx Transaction, limit int) ([]*OutboxEvent, error)
	MarkPublished(ctx context.Context, eventID string) error
	IncrementRetryCount(ctx context.Context, eventID string, nextAvailable time.Time) error
}

type outboxRepository struct {
	db *Database
}

// NewOutboxRepository 返回默认实现。
func NewOutboxRepository(db *Database) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Save(ctx context.Context, tx Transaction, event *OutboxEvent) error {
	if event == nil {
		return ErrNilOutboxEvent
	}
	if event.EventID == "" {
		return ErrEmptyEventID
	}
	if tx == nil {
		return fmt.Errorf("transaction is nil")
	}

	if event.AvailableAt.IsZero() {
		event.AvailableAt = time.Now()
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO outbox_events
		(event_id, aggregate_id, aggregate_type, event_type, payload, retry_count, published, available_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, available_at, created_at
	`,
		event.EventID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.Payload,
		0,
		false,
		event.AvailableAt,
		time.Now(),
	)

	var (
		availableAt time.Time
		createdAt   time.Time
	)
	if err := row.Scan(&event.ID, &availableAt, &createdAt); err != nil {
		return fmt.Errorf("failed to save outbox event: %w", err)
	}
	event.CreatedAt = createdAt
	event.AvailableAt = availableAt
	event.RetryCount = 0
	event.Published = false
	event.PublishedAt = nil

	return nil
}

func (r *outboxRepository) GetUnpublishedForUpdate(ctx context.Context, tx Transaction, limit int) ([]*OutboxEvent, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT id, event_id, aggregate_id, aggregate_type, event_type, payload,
		       retry_count, published, published_at, available_at, created_at
		FROM outbox_events
		WHERE published = FALSE
		  AND available_at <= NOW()
		ORDER BY created_at ASC
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query outbox events: %w", err)
	}
	defer rows.Close()

	var events []*OutboxEvent
	for rows.Next() {
		var (
			publishedAt sql.NullTime
			availableAt time.Time
		)
		event := &OutboxEvent{}
		if err := rows.Scan(
			&event.ID,
			&event.EventID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.Payload,
			&event.RetryCount,
			&event.Published,
			&publishedAt,
			&availableAt,
			&event.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event: %w", err)
		}

		if publishedAt.Valid {
			t := publishedAt.Time
			event.PublishedAt = &t
		}
		event.AvailableAt = availableAt

		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *outboxRepository) MarkPublished(ctx context.Context, eventID string) error {
	if r.db == nil || r.db.db == nil {
		return ErrDatabaseNotInitialized
	}
	if eventID == "" {
		return ErrEmptyEventID
	}

	res, err := r.db.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET published = TRUE, published_at = NOW()
		WHERE event_id = $1
	`, eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as published: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return ErrEventNotFound
	}

	return nil
}

func (r *outboxRepository) IncrementRetryCount(ctx context.Context, eventID string, nextAvailable time.Time) error {
	if r.db == nil || r.db.db == nil {
		return ErrDatabaseNotInitialized
	}
	if eventID == "" {
		return ErrEmptyEventID
	}

	res, err := r.db.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET retry_count = retry_count + 1,
		    available_at = $2
		WHERE event_id = $1
	`, eventID, nextAvailable)
	if err != nil {
		return fmt.Errorf("failed to increment retry count: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to read affected rows: %w", err)
	}
	if rows == 0 {
		return ErrEventNotFound
	}

	return nil
}
