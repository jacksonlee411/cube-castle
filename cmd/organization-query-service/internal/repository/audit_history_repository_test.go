package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type auditMockDriver struct{}

func (d *auditMockDriver) Open(_ string) (driver.Conn, error) {
	return &auditMockConn{}, nil
}

type auditMockConn struct{}

func (c *auditMockConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}

func (c *auditMockConn) Close() error { return nil }

func (c *auditMockConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions not supported")
}

func (c *auditMockConn) CheckNamedValue(nv *driver.NamedValue) error {
	switch v := nv.Value.(type) {
	case uuid.UUID:
		nv.Value = v.String()
	case int, int32, int64, uint, uint32, uint64:
		nv.Value = fmt.Sprint(v)
	default:
		// keep original value
	}
	return nil
}

func (c *auditMockConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if currentExpectation == nil {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	defer func() { currentExpectation = nil }()

	if !strings.Contains(query, currentExpectation.querySubstring) {
		return nil, fmt.Errorf("query mismatch, got %q", query)
	}

	if len(args) != len(currentExpectation.expectedArgs) {
		return nil, fmt.Errorf("arg length mismatch: got %d want %d", len(args), len(currentExpectation.expectedArgs))
	}

	for idx, named := range args {
		if fmt.Sprint(named.Value) != currentExpectation.expectedArgs[idx] {
			return nil, fmt.Errorf("arg[%d] mismatch: got %v want %s", idx, named.Value, currentExpectation.expectedArgs[idx])
		}
	}

	return &auditMockRows{
		columns: currentExpectation.columns,
		data:    currentExpectation.rows,
		index:   -1,
	}, nil
}

type auditMockRows struct {
	columns []string
	data    [][]driver.Value
	index   int
}

func (r *auditMockRows) Columns() []string {
	return r.columns
}

func (r *auditMockRows) Close() error { return nil }

func (r *auditMockRows) Next(dest []driver.Value) error {
	r.index++
	if r.index >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.index])
	return nil
}

type auditExpectation struct {
	querySubstring string
	expectedArgs   []string
	columns        []string
	rows           [][]driver.Value
}

var (
	registerAuditMock  sync.Once
	currentExpectation *auditExpectation
)

func openAuditMockDB() (*sql.DB, error) {
	registerAuditMock.Do(func() {
		sql.Register("auditmock", &auditMockDriver{})
	})
	return sql.Open("auditmock", "")
}

func TestGetAuditHistoryReturnsStructuredRecords(t *testing.T) {
	ctx := context.Background()
	db, err := openAuditMockDB()
	if err != nil {
		t.Fatalf("failed to open auditmock db: %v", err)
	}
	defer db.Close()

	repo := &PostgreSQLRepository{
		db:          db,
		redisClient: nil,
		logger:      log.New(io.Discard, "", 0),
		auditConfig: AuditHistoryConfig{LegacyMode: false},
	}

	tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
	recordID := uuid.New()
	beforeJSON := "{\"name\":\"旧名称\"}"
	afterJSON := "{\"name\":\"新名称\"}"
	modifiedJSON := "[\"name\"]"
	changesJSON := "[{\"field\":\"name\",\"dataType\":\"string\",\"oldValue\":\"旧名称\",\"newValue\":\"新名称\"}]"
	now := time.Now().UTC().Format(time.RFC3339)

	currentExpectation = &auditExpectation{
		querySubstring: "FROM audit_logs",
		expectedArgs:   []string{tenantID.String(), recordID.String(), "50"},
		columns: []string{
			"audit_id",
			"record_id",
			"operation_type",
			"operated_by_id",
			"operated_by_name",
			"changes_summary",
			"operation_reason",
			"operation_timestamp",
			"before_data",
			"after_data",
			"modified_fields",
			"detailed_changes",
		},
		rows: [][]driver.Value{
			{
				"audit-1",
				recordID.String(),
				"UPDATE",
				"user-1",
				"Auditor",
				"{\"operationSummary\":\"UpdateOrganization\",\"totalChanges\":1,\"keyChanges\":[{\"field\":\"name\"}]}",
				"修复组织名称",
				now,
				beforeJSON,
				afterJSON,
				modifiedJSON,
				changesJSON,
			},
		},
	}

	result, err := repo.GetAuditHistory(ctx, tenantID, recordID.String(), nil, nil, nil, nil, 50)
	if err != nil {
		t.Fatalf("GetAuditHistory returned error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(result))
	}

	record := result[0]
	if record.AuditIDField != "audit-1" {
		t.Errorf("unexpected auditId: %s", record.AuditIDField)
	}
	if record.RecordIDField != recordID.String() {
		t.Errorf("unexpected recordId: %s", record.RecordIDField)
	}
	if record.OperationTypeField != "UPDATE" {
		t.Errorf("unexpected operationType: %s", record.OperationTypeField)
	}
	if record.OperationReasonField == nil || *record.OperationReasonField != "修复组织名称" {
		t.Errorf("operationReason mismatch: %v", record.OperationReasonField)
	}
	if record.BeforeDataField == nil || *record.BeforeDataField != beforeJSON {
		t.Errorf("beforeData mismatch: %v", record.BeforeDataField)
	}
	if record.AfterDataField == nil || *record.AfterDataField != afterJSON {
		t.Errorf("afterData mismatch: %v", record.AfterDataField)
	}
	if len(record.ModifiedFieldsField) != 1 || record.ModifiedFieldsField[0] != "name" {
		t.Errorf("modifiedFields mismatch: %v", record.ModifiedFieldsField)
	}
	if len(record.ChangesField) != 1 || record.ChangesField[0].FieldField != "name" {
		t.Errorf("changes mismatch: %+v", record.ChangesField)
	}
	if record.TimestampField != now {
		t.Errorf("timestamp mismatch: %s", record.TimestampField)
	}

	if currentExpectation != nil {
		t.Fatalf("expectation was not consumed")
	}
}
