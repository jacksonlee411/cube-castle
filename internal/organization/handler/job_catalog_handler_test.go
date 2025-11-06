package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cube-castle/internal/organization/service"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
)

func TestJobCatalogHandlePreconditionFailed(t *testing.T) {
	handler := &JobCatalogHandler{
		logger: pkglogger.NewNoopLogger(),
	}

	req := httptest.NewRequest(http.MethodPut, "/api/v1/job-family-groups/PROF", nil)
	rec := httptest.NewRecorder()

	handler.handleServiceError(rec, req, service.ErrJobCatalogPreconditionFailed)

	if rec.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status 412, got %d", rec.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	errorField, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected error object in response: %v", body)
	}
	if code, _ := errorField["code"].(string); code != "PRECONDITION_FAILED" {
		t.Fatalf("expected error code PRECONDITION_FAILED, got %q", code)
	}
}

// TestValidateCreateJobLevelRequest tests the CreateJobLevel request validation
func TestValidateCreateJobLevelRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *types.CreateJobLevelRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid request",
			req: &types.CreateJobLevelRequest{
				Code:          "L-001",
				JobRoleCode:   "ROLE-001",
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				LevelRank:     "5",
				EffectiveDate: "2025-01-01",
			},
			expectError: false,
		},
		{
			name: "Missing Code",
			req: &types.CreateJobLevelRequest{
				Code:          "",
				JobRoleCode:   "ROLE-001",
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				LevelRank:     "5",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级代码不能为空",
		},
		{
			name: "Missing JobRoleCode",
			req: &types.CreateJobLevelRequest{
				Code:          "L-001",
				JobRoleCode:   "",
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				LevelRank:     "5",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职位角色代码不能为空",
		},
		{
			name: "Missing Name",
			req: &types.CreateJobLevelRequest{
				Code:          "L-001",
				JobRoleCode:   "ROLE-001",
				Name:          "",
				Status:        "ACTIVE",
				LevelRank:     "5",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级名称不能为空",
		},
		{
			name: "Missing Status",
			req: &types.CreateJobLevelRequest{
				Code:          "L-001",
				JobRoleCode:   "ROLE-001",
				Name:          "Senior Manager",
				Status:        "",
				LevelRank:     "5",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级状态不能为空",
		},
		{
			name: "Missing LevelRank",
			req: &types.CreateJobLevelRequest{
				Code:          "L-001",
				JobRoleCode:   "ROLE-001",
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				LevelRank:     "",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级排序号不能为空",
		},
		{
			name: "Missing EffectiveDate",
			req: &types.CreateJobLevelRequest{
				Code:          "L-001",
				JobRoleCode:   "ROLE-001",
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				LevelRank:     "5",
				EffectiveDate: "",
			},
			expectError:   true,
			errorContains: "生效日期不能为空",
		},
		{
			name: "Whitespace-only Code",
			req: &types.CreateJobLevelRequest{
				Code:          "   ",
				JobRoleCode:   "ROLE-001",
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				LevelRank:     "5",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级代码不能为空",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCreateJobLevelRequest(tc.req)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%v'", err)
				}
			}
		})
	}
}

// TestValidateUpdateJobLevelRequest tests the UpdateJobLevel request validation
func TestValidateUpdateJobLevelRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *types.UpdateJobLevelRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid request",
			req: &types.UpdateJobLevelRequest{
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				EffectiveDate: "2025-01-01",
			},
			expectError: false,
		},
		{
			name: "Missing Name",
			req: &types.UpdateJobLevelRequest{
				Name:          "",
				Status:        "ACTIVE",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级名称不能为空",
		},
		{
			name: "Missing Status",
			req: &types.UpdateJobLevelRequest{
				Name:          "Senior Manager",
				Status:        "",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级状态不能为空",
		},
		{
			name: "Missing EffectiveDate",
			req: &types.UpdateJobLevelRequest{
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				EffectiveDate: "",
			},
			expectError:   true,
			errorContains: "生效日期不能为空",
		},
		{
			name: "Whitespace-only Name",
			req: &types.UpdateJobLevelRequest{
				Name:          "   ",
				Status:        "ACTIVE",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "职级名称不能为空",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUpdateJobLevelRequest(tc.req)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%v'", err)
				}
			}
		})
	}
}

// TestValidateJobCatalogVersionRequest tests the JobCatalogVersion request validation
func TestValidateJobCatalogVersionRequest(t *testing.T) {
	tests := []struct {
		name          string
		req           *types.JobCatalogVersionRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "Valid request",
			req: &types.JobCatalogVersionRequest{
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				EffectiveDate: "2025-01-01",
			},
			expectError: false,
		},
		{
			name: "Missing Name",
			req: &types.JobCatalogVersionRequest{
				Name:          "",
				Status:        "ACTIVE",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "名称不能为空",
		},
		{
			name: "Missing Status",
			req: &types.JobCatalogVersionRequest{
				Name:          "Senior Manager",
				Status:        "",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "状态不能为空",
		},
		{
			name: "Missing EffectiveDate",
			req: &types.JobCatalogVersionRequest{
				Name:          "Senior Manager",
				Status:        "ACTIVE",
				EffectiveDate: "",
			},
			expectError:   true,
			errorContains: "生效日期不能为空",
		},
		{
			name: "Whitespace-only Name",
			req: &types.JobCatalogVersionRequest{
				Name:          "   ",
				Status:        "ACTIVE",
				EffectiveDate: "2025-01-01",
			},
			expectError:   true,
			errorContains: "名称不能为空",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateJobCatalogVersionRequest(tc.req)
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("expected error containing '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got '%v'", err)
				}
			}
		})
	}
}
