package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/organizationunit"
	"github.com/gaogu/cube-castle/go-app/ent/position"
	testutil "github.com/gaogu/cube-castle/go-app/internal/testing"
)

func setupPositionHandler(t *testing.T) (*PositionHandler, *ent.Client, func()) {
	// Create test database using configurable backend
	client, logger, cleanup := testutil.SetupTestHandler(t)
	
	// Create handler
	handler := NewPositionHandler(client, logger)
	
	return handler, client, cleanup
}

func createTestDepartment(t *testing.T, client *ent.Client, tenantID uuid.UUID) *ent.OrganizationUnit {
	dept, err := client.OrganizationUnit.Create().
		SetTenantID(tenantID).
		SetUnitType(organizationunit.UnitTypeDEPARTMENT).
		SetName("测试部门").
		SetDescription("测试用部门").
		SetProfile(map[string]interface{}{
			"functional_area": "ENGINEERING",
		}).
		Save(context.Background())
	require.NoError(t, err)
	return dept
}

func TestPositionHandler_CreatePosition(t *testing.T) {
	handler, client, cleanup := setupPositionHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	dept := createTestDepartment(t, client, tenantID)
	
	testCases := []struct {
		name           string
		payload        CreatePositionRequest
		expectedStatus int
		setupTenant    bool
	}{
		{
			name: "Valid FULL_TIME position creation",
			payload: CreatePositionRequest{
				PositionType:  "FULL_TIME",
				JobProfileID:  uuid.New(),
				DepartmentID:  dept.ID,
				Status:        "OPEN",
				BudgetedFTE:   1.0,
				Details: map[string]interface{}{
					"work_schedule": "standard",
					"location":      "office",
				},
			},
			expectedStatus: http.StatusCreated,
			setupTenant:    true,
		},
		{
			name: "Valid PART_TIME position creation",
			payload: CreatePositionRequest{
				PositionType: "PART_TIME",
				JobProfileID: uuid.New(),
				DepartmentID: dept.ID,
				BudgetedFTE:  0.5,
				Details: map[string]interface{}{
					"standard_hours_per_week": 20,
				},
			},
			expectedStatus: http.StatusCreated,
			setupTenant:    true,
		},
		{
			name: "Invalid position type",
			payload: CreatePositionRequest{
				PositionType: "INVALID_TYPE",
				JobProfileID: uuid.New(),
				DepartmentID: dept.ID,
			},
			expectedStatus: http.StatusBadRequest,
			setupTenant:    true,
		},
		{
			name: "Missing required fields",
			payload: CreatePositionRequest{
				PositionType: "FULL_TIME",
				// Missing JobProfileID and DepartmentID
			},
			expectedStatus: http.StatusBadRequest,
			setupTenant:    true,
		},
		{
			name: "Non-existing department",
			payload: CreatePositionRequest{
				PositionType: "FULL_TIME",
				JobProfileID: uuid.New(),
				DepartmentID: uuid.New(), // Non-existing department
			},
			expectedStatus: http.StatusBadRequest,
			setupTenant:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request
			payload, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/v1/positions", bytes.NewBuffer(payload))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Add tenant context
			if tc.setupTenant {
				ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
				req = req.WithContext(ctx)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.CreatePosition()(rr, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusCreated {
				var response PositionResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, tc.payload.PositionType, response.PositionType)
				assert.Equal(t, tc.payload.JobProfileID, response.JobProfileID)
				assert.Equal(t, tc.payload.DepartmentID, response.DepartmentID)
				assert.Equal(t, tenantID, response.TenantID)
				assert.NotEqual(t, uuid.Nil, response.ID)
				
				// Check default values
				if tc.payload.Status == "" {
					assert.Equal(t, "OPEN", response.Status)
				}
				if tc.payload.BudgetedFTE == 0 {
					assert.Equal(t, 1.0, response.BudgetedFTE)
				}
			}
		})
	}
}

func TestPositionHandler_GetPosition(t *testing.T) {
	handler, client, cleanup := setupPositionHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	dept := createTestDepartment(t, client, tenantID)
	
	// Create test position
	pos, err := client.Position.Create().
		SetTenantID(tenantID).
		SetPositionType(position.PositionTypeFULL_TIME).
		SetJobProfileID(uuid.New()).
		SetDepartmentID(dept.ID).
		SetStatus(position.StatusOPEN).
		SetBudgetedFte(1.0).
		Save(context.Background())
	require.NoError(t, err)

	testCases := []struct {
		name           string
		positionID     string
		tenantID       uuid.UUID
		expectedStatus int
	}{
		{
			name:           "Valid request",
			positionID:     pos.ID.String(),
			tenantID:       tenantID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID",
			positionID:     "invalid-uuid",
			tenantID:       tenantID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not found",
			positionID:     uuid.New().String(),
			tenantID:       tenantID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with chi URL parameters
			req, err := http.NewRequest(http.MethodGet, "/api/v1/positions/"+tc.positionID, nil)
			require.NoError(t, err)

			// Add tenant context
			ctx := context.WithValue(req.Context(), "tenant_id", tc.tenantID)
			req = req.WithContext(ctx)

			// Create chi context with URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.positionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.GetPosition()(rr, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response PositionResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, pos.ID, response.ID)
				assert.Equal(t, pos.PositionType.String(), response.PositionType)
				assert.Equal(t, pos.JobProfileID, response.JobProfileID)
				assert.Equal(t, pos.DepartmentID, response.DepartmentID)
			}
		})
	}
}

func TestPositionHandler_ListPositions(t *testing.T) {
	handler, client, cleanup := setupPositionHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	dept := createTestDepartment(t, client, tenantID)
	
	// Create test positions
	_, err := client.Position.Create().
		SetTenantID(tenantID).
		SetPositionType(position.PositionTypeFULL_TIME).
		SetJobProfileID(uuid.New()).
		SetDepartmentID(dept.ID).
		SetStatus(position.StatusOPEN).
		SetBudgetedFte(1.0).
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.Position.Create().
		SetTenantID(tenantID).
		SetPositionType(position.PositionTypePART_TIME).
		SetJobProfileID(uuid.New()).
		SetDepartmentID(dept.ID).
		SetStatus(position.StatusFILLED).
		SetBudgetedFte(0.5).
		Save(context.Background())
	require.NoError(t, err)

	t.Run("List all positions", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/positions", nil)
		require.NoError(t, err)

		// Add tenant context
		ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		handler.ListPositions()(rr, req)

		// Assert response
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		data, ok := response["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 2)
	})

	t.Run("Filter by position type", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/positions?position_type=FULL_TIME", nil)
		require.NoError(t, err)

		// Add tenant context
		ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		handler.ListPositions()(rr, req)

		// Assert response
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		data, ok := response["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 1)
	})

	t.Run("Filter by status", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/positions?status=OPEN", nil)
		require.NoError(t, err)

		// Add tenant context
		ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		handler.ListPositions()(rr, req)

		// Assert response
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		data, ok := response["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 1)
	})
}

func TestPositionHandler_UpdatePosition(t *testing.T) {
	handler, client, cleanup := setupPositionHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	dept := createTestDepartment(t, client, tenantID)
	
	// Create test position
	pos, err := client.Position.Create().
		SetTenantID(tenantID).
		SetPositionType(position.PositionTypeFULL_TIME).
		SetJobProfileID(uuid.New()).
		SetDepartmentID(dept.ID).
		SetStatus(position.StatusOPEN).
		SetBudgetedFte(1.0).
		Save(context.Background())
	require.NoError(t, err)

	testCases := []struct {
		name           string
		positionID     string
		payload        UpdatePositionRequest
		expectedStatus int
		setupTenant    bool
	}{
		{
			name:       "Valid update",
			positionID: pos.ID.String(),
			payload: UpdatePositionRequest{
				Status: stringPtr("FILLED"),
				BudgetedFTE: float64Ptr(0.8),
			},
			expectedStatus: http.StatusOK,
			setupTenant:    true,
		},
		{
			name:       "Invalid status",
			positionID: pos.ID.String(),
			payload: UpdatePositionRequest{
				Status: stringPtr("INVALID_STATUS"),
			},
			expectedStatus: http.StatusBadRequest,
			setupTenant:    true,
		},
		{
			name:           "Position not found",
			positionID:     uuid.New().String(),
			payload:        UpdatePositionRequest{},
			expectedStatus: http.StatusNotFound,
			setupTenant:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare request
			payload, err := json.Marshal(tc.payload)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPut, "/api/v1/positions/"+tc.positionID, bytes.NewBuffer(payload))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Add tenant context
			if tc.setupTenant {
				ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
				req = req.WithContext(ctx)
			}

			// Create chi context with URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.positionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.UpdatePosition()(rr, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response PositionResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				
				if tc.payload.Status != nil {
					assert.Equal(t, *tc.payload.Status, response.Status)
				}
				if tc.payload.BudgetedFTE != nil {
					assert.Equal(t, *tc.payload.BudgetedFTE, response.BudgetedFTE)
				}
			}
		})
	}
}

func TestPositionHandler_DeletePosition(t *testing.T) {
	handler, client, cleanup := setupPositionHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	dept := createTestDepartment(t, client, tenantID)
	
	// Create test position
	pos, err := client.Position.Create().
		SetTenantID(tenantID).
		SetPositionType(position.PositionTypeFULL_TIME).
		SetJobProfileID(uuid.New()).
		SetDepartmentID(dept.ID).
		SetStatus(position.StatusOPEN).
		SetBudgetedFte(1.0).
		Save(context.Background())
	require.NoError(t, err)

	// Create another position that reports to the first one (manager-subordinate relationship)
	subordinatePos, err := client.Position.Create().
		SetTenantID(tenantID).
		SetPositionType(position.PositionTypeFULL_TIME).
		SetJobProfileID(uuid.New()).
		SetDepartmentID(dept.ID).
		SetManagerPositionID(pos.ID).
		SetStatus(position.StatusOPEN).
		SetBudgetedFte(1.0).
		Save(context.Background())
	require.NoError(t, err)

	testCases := []struct {
		name           string
		positionID     string
		expectedStatus int
		setupTenant    bool
	}{
		{
			name:           "Delete position with subordinates should fail",
			positionID:     pos.ID.String(),
			expectedStatus: http.StatusConflict,
			setupTenant:    true,
		},
		{
			name:           "Delete position without subordinates should succeed",
			positionID:     subordinatePos.ID.String(),
			expectedStatus: http.StatusNoContent,
			setupTenant:    true,
		},
		{
			name:           "Position not found",
			positionID:     uuid.New().String(),
			expectedStatus: http.StatusNotFound,
			setupTenant:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, "/api/v1/positions/"+tc.positionID, nil)
			require.NoError(t, err)

			// Add tenant context
			if tc.setupTenant {
				ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
				req = req.WithContext(ctx)
			}

			// Create chi context with URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.positionID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.DeletePosition()(rr, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestPositionHandler_NoTenantContext(t *testing.T) {
	handler, _, cleanup := setupPositionHandler(t)
	defer cleanup()

	testCases := []struct {
		name    string
		method  string
		path    string
		handler http.HandlerFunc
	}{
		{
			name:    "Create without tenant",
			method:  http.MethodPost,
			path:    "/api/v1/positions",
			handler: handler.CreatePosition(),
		},
		{
			name:    "Get without tenant",
			method:  http.MethodGet,
			path:    "/api/v1/positions/" + uuid.New().String(),
			handler: handler.GetPosition(),
		},
		{
			name:    "List without tenant",
			method:  http.MethodGet,
			path:    "/api/v1/positions",
			handler: handler.ListPositions(),
		},
		{
			name:    "Update without tenant",
			method:  http.MethodPut,
			path:    "/api/v1/positions/" + uuid.New().String(),
			handler: handler.UpdatePosition(),
		},
		{
			name:    "Delete without tenant",
			method:  http.MethodDelete,
			path:    "/api/v1/positions/" + uuid.New().String(),
			handler: handler.DeletePosition(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			tc.handler(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
		})
	}
}