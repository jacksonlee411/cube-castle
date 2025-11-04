package audit

import (
	"context"
	"fmt"
	"testing"

	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

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
		sqlmock.AnyArg(),      // id
		tenantID,              // tenant_id
		event.EventType,       // event_type
		event.ResourceType,    // resource_type
		expectedResourceID,    // resource_id (fallback)
		event.ActorID,         // actor_id
		event.ActorType,       // actor_type
		event.ActionName,      // action_name
		event.RequestID,       // request_id
		event.OperationReason, // operation_reason
		sqlmock.AnyArg(),      // timestamp
		event.Success,         // success
		event.ErrorCode,       // error_code
		event.ErrorMessage,    // error_message
		"{}",                  // request_data
		"{}",                  // response_data
		"[]",                  // modified_fields
		"[]",                  // changes
	).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := auditLogger.LogEvent(context.Background(), event); err != nil {
		t.Fatalf("LogEvent returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}
