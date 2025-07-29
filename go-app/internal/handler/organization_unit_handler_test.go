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
	testutil "github.com/gaogu/cube-castle/go-app/internal/testing"
)

func setupOrganizationUnitHandler(t *testing.T) (*OrganizationUnitHandler, *ent.Client, func()) {
	// Create test database using configurable backend
	client, logger, cleanup := testutil.SetupTestHandler(t)
	
	// Create handler
	handler := NewOrganizationUnitHandler(client, logger)
	
	return handler, client, cleanup
}

func TestOrganizationUnitHandler_CreateOrganizationUnit(t *testing.T) {
	handler, _, cleanup := setupOrganizationUnitHandler(t)
	defer cleanup()

	// Test data
	tenantID := uuid.New()
	
	testCases := []struct {
		name           string
		payload        CreateOrganizationUnitRequest
		expectedStatus int
		setupTenant    bool
	}{
		{
			name: "Valid DEPARTMENT creation",
			payload: CreateOrganizationUnitRequest{
				UnitType:    "DEPARTMENT",
				Name:        "工程技术部",
				Description: stringPtr("负责技术研发工作"),
				Profile: map[string]interface{}{
					"functional_area": "ENGINEERING",
					"cost_center":     "CC001",
					"budget_code":     "TECH-001",
				},
			},
			expectedStatus: http.StatusCreated,
			setupTenant:    true,
		},
		{
			name: "Valid COST_CENTER creation",
			payload: CreateOrganizationUnitRequest{
				UnitType:    "COST_CENTER",
				Name:        "成本中心",
				Description: stringPtr("成本控制中心"),
				Profile: map[string]interface{}{
					"cost_center_code": "CC001",
					"budget_amount":    1000000,
				},
			},
			expectedStatus: http.StatusCreated,
			setupTenant:    true,
		},
		{
			name: "Invalid unit type",
			payload: CreateOrganizationUnitRequest{
				UnitType: "INVALID_TYPE",
				Name:     "测试部门",
			},
			expectedStatus: http.StatusBadRequest,
			setupTenant:    true,
		},
		{
			name: "Missing required fields",
			payload: CreateOrganizationUnitRequest{
				UnitType: "DEPARTMENT",
				// Missing name
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

			req, err := http.NewRequest(http.MethodPost, "/api/v1/organization-units", bytes.NewBuffer(payload))
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
			handler.CreateOrganizationUnit()(rr, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusCreated {
				var response OrganizationUnitResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, tc.payload.UnitType, response.UnitType)
				assert.Equal(t, tc.payload.Name, response.Name)
				assert.Equal(t, tc.payload.Description, response.Description)
				assert.Equal(t, tenantID, response.TenantID)
				assert.NotEqual(t, uuid.Nil, response.ID)
			}
		})
	}
}

func TestOrganizationUnitHandler_GetOrganizationUnit(t *testing.T) {
	handler, client, cleanup := setupOrganizationUnitHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	
	// Create test organization unit
	orgUnit, err := client.OrganizationUnit.Create().
		SetTenantID(tenantID).
		SetUnitType(organizationunit.UnitTypeDEPARTMENT).
		SetName("测试部门").
		SetDescription("测试用部门").
		SetProfile(map[string]interface{}{
			"functional_area": "ENGINEERING",
		}).
		Save(context.Background())
	require.NoError(t, err)

	testCases := []struct {
		name           string
		orgUnitID      string
		tenantID       uuid.UUID
		expectedStatus int
	}{
		{
			name:           "Valid request",
			orgUnitID:      orgUnit.ID.String(),
			tenantID:       tenantID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID",
			orgUnitID:      "invalid-uuid",
			tenantID:       tenantID,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Not found",
			orgUnitID:      uuid.New().String(),
			tenantID:       tenantID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with chi URL parameters
			req, err := http.NewRequest(http.MethodGet, "/api/v1/organization-units/"+tc.orgUnitID, nil)
			require.NoError(t, err)

			// Add tenant context
			ctx := context.WithValue(req.Context(), "tenant_id", tc.tenantID)
			req = req.WithContext(ctx)

			// Create chi context with URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tc.orgUnitID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.GetOrganizationUnit()(rr, req)

			// Assert response
			assert.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusOK {
				var response OrganizationUnitResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				
				assert.Equal(t, orgUnit.ID, response.ID)
				assert.Equal(t, orgUnit.Name, response.Name)
				assert.Equal(t, orgUnit.UnitType.String(), response.UnitType)
			}
		})
	}
}

func TestOrganizationUnitHandler_ListOrganizationUnits(t *testing.T) {
	handler, client, cleanup := setupOrganizationUnitHandler(t)
	defer cleanup()

	tenantID := uuid.New()
	
	// Create test organization units
	_, err := client.OrganizationUnit.Create().
		SetTenantID(tenantID).
		SetUnitType(organizationunit.UnitTypeDEPARTMENT).
		SetName("部门1").
		SetDescription("测试部门1").
		Save(context.Background())
	require.NoError(t, err)

	_, err = client.OrganizationUnit.Create().
		SetTenantID(tenantID).
		SetUnitType(organizationunit.UnitTypeCOST_CENTER).
		SetName("成本中心1").
		SetDescription("测试成本中心1").
		Save(context.Background())
	require.NoError(t, err)

	t.Run("List all organization units", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/organization-units", nil)
		require.NoError(t, err)

		// Add tenant context
		ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		handler.ListOrganizationUnits()(rr, req)

		// Assert response
		assert.Equal(t, http.StatusOK, rr.Code)

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)

		data, ok := response["data"].([]interface{})
		require.True(t, ok)
		assert.Len(t, data, 2)
	})

	t.Run("Filter by unit type", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/api/v1/organization-units?unit_type=DEPARTMENT", nil)
		require.NoError(t, err)

		// Add tenant context
		ctx := context.WithValue(req.Context(), "tenant_id", tenantID)
		req = req.WithContext(ctx)

		// Create response recorder
		rr := httptest.NewRecorder()

		// Call handler
		handler.ListOrganizationUnits()(rr, req)

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

func TestOrganizationUnitHandler_NoTenantContext(t *testing.T) {
	handler, _, cleanup := setupOrganizationUnitHandler(t)
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
			path:    "/api/v1/organization-units",
			handler: handler.CreateOrganizationUnit(),
		},
		{
			name:    "Get without tenant",
			method:  http.MethodGet,
			path:    "/api/v1/organization-units/" + uuid.New().String(),
			handler: handler.GetOrganizationUnit(),
		},
		{
			name:    "List without tenant",
			method:  http.MethodGet,
			path:    "/api/v1/organization-units",
			handler: handler.ListOrganizationUnits(),
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

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}