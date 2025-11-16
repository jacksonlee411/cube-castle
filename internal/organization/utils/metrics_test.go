package utils

import "testing"

func TestMetricsRecorders_NoPanic(t *testing.T) {
	// Exercise registration and counters
	RecordTemporalOperation(OperationCreate, nil)
	RecordTemporalOperation(OperationUpdate, assertError{})
	RecordAuditWrite(nil)
	RecordAuditWrite(assertError{})
	RecordHTTPRequest("POST", "/api/v1/organization-units", 201)
	RecordHTTPRequest("PUT", "/api/v1/organization-units/1000001", 200)
	RecordOutboxDispatch("success", "org.created")
	RecordOutboxDispatch("error", "")
}

type assertError struct{}

func (assertError) Error() string { return "err" }
