package database

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestOutboxRepositorySave(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/outbox-save"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	repo := NewOutboxRepository(db)

	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO outbox_events").
		WithArgs("00000000-0000-0000-0000-000000000001", "agg-1", "organization", "organization.created", `{"id":"agg-1"}`, 0, false, sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "available_at", "created_at"}).AddRow(int64(100), time.Now(), time.Now()))
	mock.ExpectCommit()

	event := &OutboxEvent{
		EventID:       "00000000-0000-0000-0000-000000000001",
		AggregateID:   "agg-1",
		AggregateType: "organization",
		EventType:     "organization.created",
		Payload:       `{"id":"agg-1"}`,
	}

	err := db.WithTx(context.Background(), func(ctx context.Context, tx Transaction) error {
		return repo.Save(ctx, tx, event)
	})
	require.NoError(t, err)
	require.Equal(t, int64(100), event.ID)
	require.False(t, event.Published)
	require.Equal(t, 0, event.RetryCount)
	require.Nil(t, event.PublishedAt)
	require.WithinDuration(t, time.Now(), event.AvailableAt, time.Second)
}

func TestOutboxRepositorySaveValidation(t *testing.T) {
	repo := NewOutboxRepository(&Database{})
	err := repo.Save(context.Background(), nil, nil)
	require.ErrorIs(t, err, ErrNilOutboxEvent)

	err = repo.Save(context.Background(), nil, &OutboxEvent{})
	require.ErrorIs(t, err, ErrEmptyEventID)
}

func TestNewOutboxEventDefaults(t *testing.T) {
	event := NewOutboxEvent()
	require.NotEmpty(t, event.EventID)
	require.False(t, event.Published)
	require.Equal(t, 0, event.RetryCount)
	require.WithinDuration(t, time.Now(), event.CreatedAt, time.Second)
}

func TestOutboxRepositoryGetUnpublished(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/outbox-fetch"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	repo := NewOutboxRepository(db)

	rows := sqlmock.NewRows([]string{
		"id", "event_id", "aggregate_id", "aggregate_type", "event_type", "payload",
		"retry_count", "published", "published_at", "available_at", "created_at",
	}).AddRow(
		1, "event-1", "agg-1", "organization", "created", `{"id":"agg-1"}`,
		0, false, nil, time.Now(), time.Now(),
	).AddRow(
		2, "event-2", "agg-2", "organization", "updated", `{"id":"agg-2"}`,
		2, false, time.Now(), time.Now().Add(time.Minute), time.Now(),
	)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id, event_id").
		WithArgs(50).
		WillReturnRows(rows)
	mock.ExpectCommit()

	var events []*OutboxEvent
	err := db.WithTx(context.Background(), func(ctx context.Context, tx Transaction) error {
		var err error
		events, err = repo.GetUnpublishedForUpdate(ctx, tx, 50)
		return err
	})

	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, "event-1", events[0].EventID)
	require.Nil(t, events[0].PublishedAt)
	require.NotNil(t, events[1].PublishedAt)
	require.WithinDuration(t, time.Now(), events[0].AvailableAt, time.Second)
}

func TestOutboxRepositoryMarkPublished(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/outbox-publish"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	repo := NewOutboxRepository(db)

	mock.ExpectExec("UPDATE outbox_events").
		WithArgs("event-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.MarkPublished(context.Background(), "event-1")
	require.NoError(t, err)
}

func TestOutboxRepositoryIncrementRetry(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/outbox-retry"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	repo := NewOutboxRepository(db)

	mock.ExpectExec("UPDATE outbox_events").
		WithArgs("event-1", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.IncrementRetryCount(context.Background(), "event-1", time.Now().Add(time.Minute))
	require.NoError(t, err)
}

func TestOutboxRepositoryErrors(t *testing.T) {
	cfg := ConnectionConfig{DSN: "postgres://test/outbox-errors"}
	db, mock, cleanup := newMockDatabase(t, cfg)
	defer cleanup()

	repo := NewOutboxRepository(db)

	mock.ExpectExec("UPDATE outbox_events").
		WithArgs("event-404").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repo.MarkPublished(context.Background(), "event-404")
	require.ErrorIs(t, err, ErrEventNotFound)

	mock.ExpectExec("UPDATE outbox_events").
		WithArgs("event-404", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.IncrementRetryCount(context.Background(), "event-404", time.Now().Add(time.Minute))
	require.ErrorIs(t, err, ErrEventNotFound)
}
