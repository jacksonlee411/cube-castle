package audit

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

type jsonContains struct {
	fragment string
}

func (m jsonContains) Match(v driver.Value) bool {
	data, ok := normalizeArgument(v)
	if !ok {
		return strings.Contains(fmt.Sprint(v), m.fragment)
	}
	return strings.Contains(string(data), m.fragment)
}

func normalizeArgument(v driver.Value) ([]byte, bool) {
	switch val := v.(type) {
	case string:
		return []byte(val), true
	case []byte:
		return val, true
	default:
		bytes, err := json.Marshal(val)
		if err != nil {
			return nil, false
		}
		return bytes, true
	}
}

type fieldChangeHas struct {
	field    string
	newValue string
}

func (m fieldChangeHas) Match(v driver.Value) bool {
	data, ok := normalizeArgument(v)
	if !ok {
		return false
	}
	var entries []map[string]interface{}
	if err := json.Unmarshal(data, &entries); err != nil {
		return false
	}
	for _, entry := range entries {
		if fmt.Sprint(entry["field"]) == m.field {
			if fmt.Sprint(entry["newValue"]) == m.newValue {
				return true
			}
		}
	}
	return false
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

func TestLogOrganizationDeleteUsesRecordId(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	recordID := uuid.New()
	org := &types.Organization{
		RecordID: recordID.String(),
		TenantID: tenantID.String(),
		Code:     "1000001",
		Name:     "删除前",
		Status:   "ACTIVE",
		Level:    1,
	}

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeDelete,
		ResourceTypeOrganization,
		recordID.String(),
		"user-1",
		ActorTypeUser,
		"DeleteOrganization",
		"req-del",
		"cleanup",
		sqlmock.AnyArg(),
		true,
		"",
		"",
		sqlmock.AnyArg(),
		sqlmock.AnyArg(),
		sqlmock.AnyArg(),
		sqlmock.AnyArg(),
		recordID,
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := auditLogger.LogOrganizationDelete(context.Background(), tenantID, org.Code, org, "user-1", "req-del", "cleanup"); err != nil {
		t.Fatalf("LogOrganizationDelete returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestLogOrganizationDeleteRequiresSnapshot(t *testing.T) {
	auditLogger := NewAuditLogger(nil, pkglogger.NewNoopLogger())
	err := auditLogger.LogOrganizationDelete(context.Background(), uuid.New(), "1000001", nil, "user-1", "req-del", "cleanup")
	if err == nil || !strings.Contains(err.Error(), "organization snapshot") {
		t.Fatalf("expected snapshot error, got %v", err)
	}

	org := &types.Organization{
		Code: "1000001",
		Name: "missing record",
	}
	err = auditLogger.LogOrganizationDelete(context.Background(), uuid.New(), org.Code, org, "user-1", "req-del", "cleanup")
	if err == nil || !strings.Contains(err.Error(), "recordId") {
		t.Fatalf("expected recordId error, got %v", err)
	}
}

func TestLogOrganizationCreateCapturesNewValues(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	recordID := uuid.New()
	parentCode := "PARENT01"
	result := &types.Organization{
		RecordID:   recordID.String(),
		TenantID:   tenantID.String(),
		Code:       "ORG001",
		Name:       "测试组织",
		UnitType:   "DEPARTMENT",
		Status:     "ACTIVE",
		Level:      1,
		ParentCode: &parentCode,
	}

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeCreate,
		ResourceTypeOrganization,
		recordID.String(),
		"user-1",
		ActorTypeUser,
		"CreateOrganization",
		"req-create",
		"reason",
		sqlmock.AnyArg(),
		true,
		"",
		"",
		"{}",
		jsonContains{`"code":"ORG001"`},
		sqlmock.AnyArg(),
		fieldChangeHas{field: "code", newValue: "ORG001"},
		recordID,
		jsonContains{`"entityCode":"ORG001"`},
	).WillReturnResult(sqlmock.NewResult(1, 1))

	req := &types.CreateOrganizationRequest{}
	if err := auditLogger.LogOrganizationCreate(context.Background(), req, result, "user-1", "req-create", "reason"); err != nil {
		t.Fatalf("LogOrganizationCreate returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestLogOrganizationUpdate_EmitsChanges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()

	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	recordID := uuid.New()
	oldOrg := &types.Organization{
		RecordID: recordID.String(),
		TenantID: tenantID.String(),
		Code:     "ORG001",
		Name:     "Old",
		Status:   "ACTIVE",
		Level:    1,
	}
	newOrg := &types.Organization{
		RecordID: recordID.String(),
		TenantID: tenantID.String(),
		Code:     "ORG001",
		Name:     "New",
		Status:   "ACTIVE",
		Level:    1,
	}

	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeUpdate,
		ResourceTypeOrganization,
		recordID.String(),
		"user-1",
		ActorTypeUser,
		"UpdateOrganization",
		"req-upd",
		"reason",
		sqlmock.AnyArg(),
		true,
		"",
		"",
		sqlmock.AnyArg(),             // before
		jsonContains{`"name":"New"`}, // after
		sqlmock.AnyArg(),             // modified
		fieldChangeHas{field: "name", newValue: "New"},
		recordID,
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))

	if err := auditLogger.LogOrganizationUpdate(context.Background(), "ORG001", &types.UpdateOrganizationRequest{}, oldOrg, newOrg, "user-1", "req-upd", "reason"); err != nil {
		t.Fatalf("LogOrganizationUpdate returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestLogOrganizationSuspendActivate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer db.Close()
	auditLogger := NewAuditLogger(db, pkglogger.NewNoopLogger())
	tenantID := uuid.New()
	recordID := uuid.New()
	org := &types.Organization{
		RecordID: recordID.String(),
		TenantID: tenantID.String(),
		Code:     "ORG001",
		Name:     "Name",
		Status:   "ACTIVE",
		Level:    1,
	}
	// Suspend: expect status INACTIVE
	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeSuspend,
		ResourceTypeOrganization,
		recordID.String(),
		"user-1",
		ActorTypeUser,
		"SuspendOrganization",
		"req-sus",
		"maint",
		sqlmock.AnyArg(),
		true,
		"",
		"",
		sqlmock.AnyArg(),
		jsonContains{`"status":"INACTIVE"`},
		sqlmock.AnyArg(),
		fieldChangeHas{field: "status", newValue: "INACTIVE"},
		recordID,
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := auditLogger.LogOrganizationSuspend(context.Background(), "ORG001", org, "user-1", "req-sus", "maint"); err != nil {
		t.Fatalf("LogOrganizationSuspend returned error: %v", err)
	}
	// Activate: expect status ACTIVE
	org.Status = "INACTIVE"
	mock.ExpectExec("INSERT INTO audit_logs").WithArgs(
		sqlmock.AnyArg(),
		tenantID,
		EventTypeActivate,
		ResourceTypeOrganization,
		recordID.String(),
		"user-1",
		ActorTypeUser,
		"ActivateOrganization",
		"req-act",
		"reopen",
		sqlmock.AnyArg(),
		true,
		"",
		"",
		sqlmock.AnyArg(),
		jsonContains{`"status":"ACTIVE"`},
		sqlmock.AnyArg(),
		fieldChangeHas{field: "status", newValue: "ACTIVE"},
		recordID,
		sqlmock.AnyArg(),
	).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := auditLogger.LogOrganizationActivate(context.Background(), "ORG001", org, "user-1", "req-act", "reopen"); err != nil {
		t.Fatalf("LogOrganizationActivate returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations not met: %v", err)
	}
}

func TestCalculateFieldChanges_Simple(t *testing.T) {
	a := NewAuditLogger(nil, pkglogger.NewNoopLogger())
	oldOrg := &types.Organization{Name: "A", Status: "ACTIVE"}
	newOrg := &types.Organization{Name: "B", Status: "INACTIVE"}
	changes := a.calculateFieldChanges(oldOrg, newOrg)
	if len(changes) == 0 {
		t.Fatalf("expected some changes")
	}
	foundName, foundStatus := false, false
	for _, c := range changes {
		if c.Field == "name" {
			foundName = true
		}
		if c.Field == "status" {
			foundStatus = true
		}
	}
	if !foundName || !foundStatus {
		t.Fatalf("expected name and status changes; got %#v", changes)
	}
}
