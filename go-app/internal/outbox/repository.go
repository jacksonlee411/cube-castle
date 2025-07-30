package outbox

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository 发件箱数据访问层
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository 创建新的 Repository 实例
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateEvent 创建事件
func (r *Repository) CreateEvent(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO outbox.events (id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query,
		event.ID,
		event.AggregateID,
		event.AggregateType,
		event.EventType,
		event.EventVersion,
		event.Payload,
		event.Metadata,
		event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetUnprocessedEvents 获取未处理的事件
func (r *Repository) GetUnprocessedEvents(ctx context.Context, limit int) ([]Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, processed_at, created_at
		FROM outbox.events 
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query unprocessed events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.EventVersion,
			&event.Payload,
			&event.Metadata,
			&event.ProcessedAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// MarkEventAsProcessed 标记事件为已处理
func (r *Repository) MarkEventAsProcessed(ctx context.Context, eventID uuid.UUID) error {
	query := `
		UPDATE outbox.events 
		SET processed_at = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query, time.Now(), eventID)
	if err != nil {
		return fmt.Errorf("failed to mark event as processed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found: %s", eventID)
	}

	return nil
}

// GetEventsByAggregateID 根据聚合ID获取事件
func (r *Repository) GetEventsByAggregateID(ctx context.Context, aggregateID uuid.UUID) ([]Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, processed_at, created_at
		FROM outbox.events 
		WHERE aggregate_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by aggregate ID: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.EventVersion,
			&event.Payload,
			&event.Metadata,
			&event.ProcessedAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// GetEventsByType 根据事件类型获取事件
func (r *Repository) GetEventsByType(ctx context.Context, eventType string, limit int) ([]Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, processed_at, created_at
		FROM outbox.events 
		WHERE event_type = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events by type: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.EventVersion,
			&event.Payload,
			&event.Metadata,
			&event.ProcessedAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// GetEventByID 根据ID获取事件
func (r *Repository) GetEventByID(ctx context.Context, eventID uuid.UUID) (*Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, processed_at, created_at
		FROM outbox.events 
		WHERE id = $1
	`

	var event Event
	err := r.db.QueryRow(ctx, query, eventID).Scan(
		&event.ID,
		&event.AggregateID,
		&event.AggregateType,
		&event.EventType,
		&event.EventVersion,
		&event.Payload,
		&event.Metadata,
		&event.ProcessedAt,
		&event.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get event by ID: %w", err)
	}

	return &event, nil
}

// DeleteProcessedEvents 删除已处理的事件（清理旧数据）
func (r *Repository) DeleteProcessedEvents(ctx context.Context, olderThan time.Duration) (int64, error) {
	query := `
		DELETE FROM outbox.events 
		WHERE processed_at IS NOT NULL 
		AND processed_at < $1
	`

	result, err := r.db.Exec(ctx, query, time.Now().Add(-olderThan))
	if err != nil {
		return 0, fmt.Errorf("failed to delete processed events: %w", err)
	}

	return result.RowsAffected(), nil
}

// GetEventStats 获取事件统计信息
func (r *Repository) GetEventStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT 
			COUNT(*) as total_events,
			COUNT(CASE WHEN processed_at IS NULL THEN 1 END) as unprocessed_events,
			COUNT(CASE WHEN processed_at IS NOT NULL THEN 1 END) as processed_events,
			MIN(created_at) as oldest_event,
			MAX(created_at) as newest_event
		FROM outbox.events
	`

	var stats struct {
		TotalEvents       int64      `db:"total_events"`
		UnprocessedEvents int64      `db:"unprocessed_events"`
		ProcessedEvents   int64      `db:"processed_events"`
		OldestEvent       *time.Time `db:"oldest_event"`
		NewestEvent       *time.Time `db:"newest_event"`
	}

	err := r.db.QueryRow(ctx, query).Scan(
		&stats.TotalEvents,
		&stats.UnprocessedEvents,
		&stats.ProcessedEvents,
		&stats.OldestEvent,
		&stats.NewestEvent,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get event stats: %w", err)
	}

	result := map[string]interface{}{
		"total_events":       stats.TotalEvents,
		"unprocessed_events": stats.UnprocessedEvents,
		"processed_events":   stats.ProcessedEvents,
	}

	// 处理可能为NULL的时间字段
	if stats.OldestEvent != nil {
		result["oldest_event"] = *stats.OldestEvent
	} else {
		result["oldest_event"] = nil
	}

	if stats.NewestEvent != nil {
		result["newest_event"] = *stats.NewestEvent
	} else {
		result["newest_event"] = nil
	}

	// 计算处理率
	if stats.TotalEvents > 0 {
		result["processing_rate"] = float64(stats.ProcessedEvents) / float64(stats.TotalEvents) * 100
	} else {
		result["processing_rate"] = 0.0
	}

	return result, nil
}

// GetEvents 获取所有事件
func (r *Repository) GetEvents(ctx context.Context, limit int) ([]Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, event_version, payload, metadata, processed_at, created_at
		FROM outbox.events 
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.AggregateID,
			&event.AggregateType,
			&event.EventType,
			&event.EventVersion,
			&event.Payload,
			&event.Metadata,
			&event.ProcessedAt,
			&event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}
