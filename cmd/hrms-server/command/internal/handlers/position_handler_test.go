package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"cube-castle/cmd/hrms-server/command/internal/middleware"
	"cube-castle/cmd/hrms-server/command/internal/services"
	"cube-castle/internal/types"
	"cube-castle/cmd/hrms-server/command/internal/utils"
)

type stubPositionService struct {
	createFunc           func(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	replaceFunc          func(ctx context.Context, tenantID uuid.UUID, code string, ifMatch *string, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	versionFunc          func(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	fillFunc             func(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	vacateFunc           func(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	transferFunc         func(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	applyEventFunc       func(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest, operator types.OperatedByInfo) (*types.PositionResponse, error)
	listAssignmentsFunc  func(ctx context.Context, tenantID uuid.UUID, code string, opts types.AssignmentListOptions) ([]types.PositionAssignmentResponse, int, error)
	listCapturedOpts     types.AssignmentListOptions
	createAssignmentFunc func(ctx context.Context, tenantID uuid.UUID, code string, req *types.CreateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error)
	updateAssignmentFunc func(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error)
	closeAssignmentFunc  func(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error)
	createCallTenant     uuid.UUID
	createCallOperator   types.OperatedByInfo
	replaceIfMatch       *string
}

var _ PositionService = (*stubPositionService)(nil)

func (s *stubPositionService) CreatePosition(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	if s.createFunc == nil {
		return nil, nil
	}
	s.createCallTenant = tenantID
	s.createCallOperator = operator
	return s.createFunc(ctx, tenantID, req, operator)
}

func (s *stubPositionService) ReplacePosition(ctx context.Context, tenantID uuid.UUID, code string, ifMatch *string, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	s.replaceIfMatch = ifMatch
	if s.replaceFunc == nil {
		return nil, nil
	}
	return s.replaceFunc(ctx, tenantID, code, ifMatch, req, operator)
}

func (s *stubPositionService) CreatePositionVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	if s.versionFunc == nil {
		return nil, nil
	}
	return s.versionFunc(ctx, tenantID, code, req, operator)
}

func (s *stubPositionService) FillPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	if s.fillFunc == nil {
		return nil, nil
	}
	return s.fillFunc(ctx, tenantID, code, req, operator)
}

func (s *stubPositionService) VacatePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	if s.vacateFunc == nil {
		return nil, nil
	}
	return s.vacateFunc(ctx, tenantID, code, req, operator)
}

func (s *stubPositionService) TransferPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	if s.transferFunc == nil {
		return nil, nil
	}
	return s.transferFunc(ctx, tenantID, code, req, operator)
}

func (s *stubPositionService) ApplyEvent(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
	if s.applyEventFunc == nil {
		return nil, nil
	}
	return s.applyEventFunc(ctx, tenantID, code, req, operator)
}

func (s *stubPositionService) ListAssignments(ctx context.Context, tenantID uuid.UUID, code string, opts types.AssignmentListOptions) ([]types.PositionAssignmentResponse, int, error) {
	s.listCapturedOpts = opts
	if s.listAssignmentsFunc == nil {
		return []types.PositionAssignmentResponse{}, 0, nil
	}
	return s.listAssignmentsFunc(ctx, tenantID, code, opts)
}

func (s *stubPositionService) CreateAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, req *types.CreateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
	if s.createAssignmentFunc == nil {
		return nil, nil
	}
	return s.createAssignmentFunc(ctx, tenantID, code, req, operator)
}

func (s *stubPositionService) UpdateAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
	if s.updateAssignmentFunc == nil {
		return nil, nil
	}
	return s.updateAssignmentFunc(ctx, tenantID, code, assignmentID, req, operator)
}

func (s *stubPositionService) CloseAssignmentRecord(ctx context.Context, tenantID uuid.UUID, code string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
	if s.closeAssignmentFunc == nil {
		return nil, nil
	}
	return s.closeAssignmentFunc(ctx, tenantID, code, assignmentID, req, operator)
}

func newPositionResponse() *types.PositionResponse {
	now := time.Now().UTC()
	return &types.PositionResponse{
		Code:               "P1000001",
		Title:              "HR Manager",
		JobFamilyGroupCode: "OPER",
		JobFamilyGroupName: "Operations",
		JobFamilyCode:      "OPER-HR",
		JobFamilyName:      "Human Resources",
		JobRoleCode:        "OPER-HR-SUP",
		JobRoleName:        "Supervisor",
		JobLevelCode:       "P1",
		JobLevelName:       "Level 1",
		OrganizationCode:   "1000001",
		PositionType:       "FULL_TIME",
		Status:             "ACTIVE",
		EmploymentType:     "FULL_TIME",
		HeadcountCapacity:  1,
		HeadcountInUse:     0,
		AvailableHeadcount: 1,
		EffectiveDate:      now,
		IsCurrent:          true,
		RecordID:           uuid.New(),
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func withRequestID(req *http.Request) *http.Request {
	ctx := context.WithValue(req.Context(), middleware.RequestIDKey, "request-123")
	return req.WithContext(ctx)
}

func withChiURLParam(req *http.Request, key, value string) *http.Request {
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add(key, value)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx)
	return req.WithContext(ctx)
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder) utils.APIResponse {
	t.Helper()
	var resp utils.APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return resp
}

func TestPositionHandler_CreatePosition_Success(t *testing.T) {
	service := &stubPositionService{
		createFunc: func(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			if req.Title != "HR Manager" {
				t.Fatalf("unexpected title: %s", req.Title)
			}
			return newPositionResponse(), nil
		},
	}
	logger := logDiscard()
	handler := NewPositionHandler(service, logger)

	payload := `{
		"title":"HR Manager",
		"jobFamilyGroupCode":"OPER",
		"jobFamilyCode":"OPER-HR",
		"jobRoleCode":"OPER-HR-SUP",
		"jobLevelCode":"P1",
		"organizationCode":"1000001",
		"effectiveDate":"2025-01-01",
		"employmentType":"FULL_TIME",
		"positionType":"FULL_TIME",
		"headcountCapacity":1
	}`

	tenant := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", tenant.String())
	req.Header.Set("X-Actor-Name", "Jane Doe")
	req = withRequestID(req)

	recorder := httptest.NewRecorder()
	handler.CreatePosition(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success=true")
	}
	if resp.Message != "Position created successfully" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
	if service.createCallTenant != tenant {
		t.Fatalf("tenant mismatch: %s", service.createCallTenant)
	}
	if service.createCallOperator.Name != "Jane Doe" {
		t.Fatalf("operator name mismatch: %s", service.createCallOperator.Name)
	}
	if data, ok := resp.Data.(map[string]interface{}); ok {
		if data["code"] != "P1000001" {
			t.Fatalf("unexpected code: %v", data["code"])
		}
	} else {
		t.Fatalf("expected map data, got %T", resp.Data)
	}
}

func TestPositionHandler_ReplacePosition_VersionConflict(t *testing.T) {
	service := &stubPositionService{
		replaceFunc: func(ctx context.Context, tenantID uuid.UUID, code string, ifMatch *string, req *types.PositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			if code != "P2000001" {
				t.Fatalf("unexpected code: %s", code)
			}
			if ifMatch == nil || *ifMatch != "abc" {
				t.Fatalf("expected if-match abc, got %v", ifMatch)
			}
			return nil, services.ErrVersionConflict
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"title":"Update","jobFamilyGroupCode":"OPER","jobFamilyCode":"OPER-HR","jobRoleCode":"OPER-HR-SUP","jobLevelCode":"P1","organizationCode":"1000001","effectiveDate":"2025-01-01","employmentType":"FULL_TIME","positionType":"FULL_TIME","headcountCapacity":1}`

	req := httptest.NewRequest(http.MethodPut, "/api/v1/positions/P2000001", strings.NewReader(payload))
	req.Header.Set("If-Match", `W/"abc"`)
	req.Header.Set("Content-Type", "application/json")
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P2000001")

	recorder := httptest.NewRecorder()
	handler.ReplacePosition(recorder, req)

	if recorder.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status 412, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if resp.Success {
		t.Fatalf("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "PRECONDITION_FAILED" {
		t.Fatalf("unexpected error: %#v", resp.Error)
	}
}

func TestPositionHandler_CreatePositionVersion_Success(t *testing.T) {
	service := &stubPositionService{
		versionFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			if req.EffectiveDate != "2025-02-01" {
				t.Fatalf("unexpected effective date: %s", req.EffectiveDate)
			}
			return newPositionResponse(), nil
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"effectiveDate":"2025-02-01","title":"HR Manager","jobFamilyGroupCode":"OPER","jobFamilyCode":"OPER-HR","jobRoleCode":"OPER-HR-SUP","jobLevelCode":"P1","organizationCode":"1000001","operationReason":"Planning"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/P1000001/versions", strings.NewReader(payload))
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P1000001")

	recorder := httptest.NewRecorder()
	handler.CreatePositionVersion(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success=true")
	}
}

func TestPositionHandler_CreatePositionVersion_Conflict(t *testing.T) {
	service := &stubPositionService{
		versionFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			return nil, services.ErrPositionVersionExists
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"effectiveDate":"2025-02-01","title":"HR Manager","jobFamilyGroupCode":"OPER","jobFamilyCode":"OPER-HR","jobRoleCode":"OPER-HR-SUP","jobLevelCode":"P1","organizationCode":"1000001","operationReason":"Planning"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/P1000001/versions", strings.NewReader(payload))
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P1000001")

	recorder := httptest.NewRecorder()
	handler.CreatePositionVersion(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if resp.Success {
		t.Fatalf("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "POSITION_VERSION_EXISTS" {
		t.Fatalf("unexpected error payload: %#v", resp.Error)
	}
}

func TestPositionHandler_FillPosition_Success(t *testing.T) {
	service := &stubPositionService{
		fillFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			if req.EmployeeID != "11111111-1111-1111-1111-111111111111" {
				t.Fatalf("unexpected employee id: %s", req.EmployeeID)
			}
			if req.EmployeeName != "张三" {
				t.Fatalf("unexpected employee name: %s", req.EmployeeName)
			}
			return newPositionResponse(), nil
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"employeeId":"11111111-1111-1111-1111-111111111111","employeeName":"张三","assignmentType":"PRIMARY","effectiveDate":"2025-01-01","operationReason":"Initial fill"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/P1000001/fill", strings.NewReader(payload))
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P1000001")

	recorder := httptest.NewRecorder()
	handler.FillPosition(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success=true")
	}
	if resp.Message != "Position filled successfully" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestPositionHandler_ListAssignments_Success(t *testing.T) {
	employeeID := uuid.New()
	assignmentID := uuid.New()
	enforcedTenant := uuid.New()
	sampleEffective := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	service := &stubPositionService{
		listAssignmentsFunc: func(ctx context.Context, tenantID uuid.UUID, code string, opts types.AssignmentListOptions) ([]types.PositionAssignmentResponse, int, error) {
			if tenantID != enforcedTenant {
				t.Fatalf("unexpected tenant: %s", tenantID)
			}
			if code != "P1000001" {
				t.Fatalf("unexpected code: %s", code)
			}
			return []types.PositionAssignmentResponse{
				{
					AssignmentID:     assignmentID,
					PositionCode:     code,
					PositionRecordID: uuid.New(),
					EmployeeID:       employeeID,
					EmployeeName:     "代理经理",
					AssignmentType:   "ACTING",
					AssignmentStatus: "ACTIVE",
					FTE:              1,
					EffectiveDate:    sampleEffective,
					IsCurrent:        true,
				},
			}, 5, nil
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/positions/P1000001/assignments?assignmentStatus=ACTIVE&page=2&pageSize=10&includeHistorical=false&asOfDate=2025-01-01&assignmentTypes=ACTING", nil)
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P1000001")
	req.Header.Set("X-Tenant-ID", enforcedTenant.String())
	recorder := httptest.NewRecorder()

	handler.ListAssignments(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success=true")
	}
	if resp.Message != "Assignments retrieved successfully" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
	if service.listCapturedOpts.Page != 2 {
		t.Fatalf("expected page=2, got %d", service.listCapturedOpts.Page)
	}
	if service.listCapturedOpts.PageSize != 10 {
		t.Fatalf("expected pageSize=10, got %d", service.listCapturedOpts.PageSize)
	}
	if service.listCapturedOpts.Filter.IncludeHistorical {
		t.Fatalf("expected includeHistorical=false")
	}
}

func TestPositionHandler_CreateAssignment_Success(t *testing.T) {
	assignmentID := uuid.New()
	service := &stubPositionService{
		createAssignmentFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.CreateAssignmentRequest, operator types.OperatedByInfo) (*types.PositionAssignmentResponse, error) {
			if req.EmployeeName != "李四" {
				t.Fatalf("unexpected employee name: %s", req.EmployeeName)
			}
			return &types.PositionAssignmentResponse{
				AssignmentID:     assignmentID,
				PositionCode:     code,
				PositionRecordID: uuid.New(),
				EmployeeID:       uuid.New(),
				EmployeeName:     req.EmployeeName,
				AssignmentType:   req.AssignmentType,
				AssignmentStatus: "ACTIVE",
				FTE:              1,
				EffectiveDate:    time.Now(),
				IsCurrent:        true,
			}, nil
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"employeeId":"11111111-1111-1111-1111-111111111112","employeeName":"李四","assignmentType":"ACTING","effectiveDate":"2025-02-01","operationReason":"代理","autoRevert":true,"actingUntil":"2025-03-01"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/P1000002/assignments", strings.NewReader(payload))
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P1000002")
	recorder := httptest.NewRecorder()

	handler.CreateAssignment(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", recorder.Code)
	}
	resp := decodeResponse(t, recorder)
	if !resp.Success {
		t.Fatalf("expected success=true")
	}
	if resp.Message != "Assignment created successfully" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestPositionHandler_ApplyPositionEvent_NotFound(t *testing.T) {
	service := &stubPositionService{
		applyEventFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			return nil, services.ErrPositionNotFound
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"eventType":"SUSPEND","effectiveDate":"2025-01-01","operationReason":"Audit"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/P999/events", strings.NewReader(payload))
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P999")

	recorder := httptest.NewRecorder()
	handler.ApplyPositionEvent(recorder, req)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", recorder.Code)
	}

	resp := decodeResponse(t, recorder)
	if resp.Success {
		t.Fatalf("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "POSITION_NOT_FOUND" {
		t.Fatalf("unexpected error code: %#v", resp.Error)
	}
}

func TestPositionHandler_VacatePosition_MissingCode(t *testing.T) {
	service := &stubPositionService{
		vacateFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			t.Fatalf("service should not be called")
			return nil, nil
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions//vacate", strings.NewReader(`{}`))
	req = withRequestID(req)

	recorder := httptest.NewRecorder()
	handler.VacatePosition(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Code)
	}
	resp := decodeResponse(t, recorder)
	if resp.Success {
		t.Fatalf("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "MISSING_CODE" {
		t.Fatalf("unexpected error code: %#v", resp.Error)
	}
}

func TestPositionHandler_TransferPosition_Error(t *testing.T) {
	service := &stubPositionService{
		transferFunc: func(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest, operator types.OperatedByInfo) (*types.PositionResponse, error) {
			return nil, services.ErrJobCatalogMismatch
		},
	}
	handler := NewPositionHandler(service, logDiscard())

	payload := `{"targetOrganizationCode":"2000001","effectiveDate":"2025-01-01","operationReason":"Org restructure"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/positions/P1000001/transfer", strings.NewReader(payload))
	req = withRequestID(req)
	req = withChiURLParam(req, "code", "P1000001")

	recorder := httptest.NewRecorder()
	handler.TransferPosition(recorder, req)

	if recorder.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", recorder.Code)
	}
	resp := decodeResponse(t, recorder)
	if resp.Success {
		t.Fatalf("expected success=false")
	}
	if resp.Error == nil || resp.Error.Code != "JOB_CATALOG_MISMATCH" {
		t.Fatalf("unexpected error code: %#v", resp.Error)
	}
}

func logDiscard() *log.Logger {
	return log.New(io.Discard, "", log.LstdFlags)
}
