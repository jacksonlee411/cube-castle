package audit

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"
	"testing"

	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

type jsonContains struct {
	fragment string
}

func (m jsonContains) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	return strings.Contains(s, m.fragment)
}

func TestLogEvent_FallbackResourceID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeError,
		ResourceType: ResourceTypeSystem,
		ResourceID:   "",
		ActorID:      "auth-service",
		ActorType:    ActorTypeService,
		ActionName:   "REFRESH",
		RequestID:    "req-123",
		Success:      false,
	}

	expectedResourceID := fmt.Sprintf("%s_%s_%s", event.EventType, event.ResourceType, event.ActionName)

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(), // id
		tenantID,         // tenant_id
		event.EventType,  // event_type
		event.ResourceType,
		expectedResourceID,
		event.ActorID,
		event.ActorType,
		event.ActionName,
		event.RequestID,
		event.OperationReason,
		sqlmock.AnyArg(), // timestamp
		event.Success,
		event.ErrorCode,
		event.ErrorMessage,
		"{}",
		"{}",
		"[]",
		"[]",
		nil,              // record_id
		sqlmock.AnyArg(), // business_context
	).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := auditLogger.LogEvent(context.Background(), event); err != nil {
		t.Fatalf("LogEvent returned error: %v", err)
	}

	if event.ResourceID != expectedResourceID {
		t.Fatalf("expected event.ResourceID to fallback to %s, got %s", expectedResourceID, event.ResourceID)
	}
	if event.RecordID != uuid.Nil {
		t.Fatalf("expected RecordID to remain zero, got %s", event.RecordID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestLogEventInTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	recordID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeCreate,
		ResourceTypeOrganization,
		recordID.String(),
		"user-1",
		ActorTypeUser,
		"CreateOrganization",
		"req-456",
		"",
		sqlmock.AnyArg(),
		true,
		"",
		"",
		"{}",
		"{\"code\":\"ORG001\"}",
		"[]",
		"[]",
		recordID,
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectRollback()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	event := &AuditEvent{
		TenantID:     tenantID,
		EventType:    EventTypeCreate,
		ResourceType: ResourceTypeOrganization,
		ResourceID:   recordID.String(),
		RecordID:     recordID,
		ActorID:      "user-1",
		ActorType:    ActorTypeUser,
		ActionName:   "CreateOrganization",
		RequestID:    "req-456",
		Success:      true,
		AfterData: map[string]interface{}{
			"code": "ORG001",
		},
	}

	if err := auditLogger.LogEventInTransaction(context.Background(), tx, event); err != nil {
		t.Fatalf("LogEventInTransaction returned error: %v", err)
	}

	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback tx: %v", err)
	}

	if event.BusinessContext == nil || event.BusinessContext["entityCode"] != "ORG001" {
		t.Fatalf("expected business context entityCode to be set")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestLogErrorSetsPayload(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	reqData := map[string]interface{}{"code": "ERR"}

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeError,
		ResourceTypeSystem,
		"entity-1",
		"system",
		ActorTypeUser,
		"DoSomething",
		"req-err",
		"",
		sqlmock.AnyArg(),
		false,
		"E001",
		"boom",
		"{\"code\":\"ERR\"}",
		"{}",
		"[]",
		"[]",
		nil,
		jsonContains{`"payload":{"code":"ERR"}`},
	).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := auditLogger.LogError(context.Background(), tenantID, ResourceTypeSystem, "entity-1", "DoSomething", "system", "req-err", "E001", "boom", reqData); err != nil {
		t.Fatalf("LogError returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
