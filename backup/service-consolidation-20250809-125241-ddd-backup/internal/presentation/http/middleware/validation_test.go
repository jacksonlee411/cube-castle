package validation

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateOrganizationRequestValidate(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateOrganizationRequest
		expectError bool
		errorFields []string
	}{
		{
			name: "Valid complete request",
			request: CreateOrganizationRequest{
				Name:        "Test Department",
				UnitType:    "DEPARTMENT",
				Status:      "ACTIVE",
				Level:       2,
				ParentCode:  stringPtr("1000000"),
				SortOrder:   intPtr(1),
				Description: stringPtr("Test description"),
			},
			expectError: false,
		},
		{
			name: "Valid minimal request",
			request: CreateOrganizationRequest{
				Name:     "Test Department",
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectError: false,
		},
		{
			name: "Empty name",
			request: CreateOrganizationRequest{
				Name:     "",
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "Name too long",
			request: CreateOrganizationRequest{
				Name:     string(make([]byte, 101)), // 101 characters
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "Invalid unit type",
			request: CreateOrganizationRequest{
				Name:     "Test Department",
				UnitType: "INVALID_TYPE",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectError: true,
			errorFields: []string{"unit_type"},
		},
		{
			name: "Invalid status",
			request: CreateOrganizationRequest{
				Name:     "Test Department",
				UnitType: "DEPARTMENT",
				Status:   "INVALID_STATUS",
				Level:    2,
			},
			expectError: true,
			errorFields: []string{"status"},
		},
		{
			name: "Level too low",
			request: CreateOrganizationRequest{
				Name:     "Test Department",
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    0,
			},
			expectError: true,
			errorFields: []string{"level"},
		},
		{
			name: "Level too high",
			request: CreateOrganizationRequest{
				Name:     "Test Department",
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    15,
			},
			expectError: true,
			errorFields: []string{"level"},
		},
		{
			name: "Invalid code format",
			request: CreateOrganizationRequest{
				Code:     stringPtr("123"), // Too short
				Name:     "Test Department",
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectError: true,
			errorFields: []string{"code"},
		},
		{
			name: "Invalid parent code format",
			request: CreateOrganizationRequest{
				Name:       "Test Department",
				UnitType:   "DEPARTMENT",
				Status:     "ACTIVE",
				Level:      2,
				ParentCode: stringPtr("123"), // Too short
			},
			expectError: true,
			errorFields: []string{"parent_code"},
		},
		{
			name: "Negative sort order",
			request: CreateOrganizationRequest{
				Name:      "Test Department",
				UnitType:  "DEPARTMENT",
				Status:    "ACTIVE",
				Level:     2,
				SortOrder: intPtr(-1),
			},
			expectError: true,
			errorFields: []string{"sort_order"},
		},
		{
			name: "Description too long",
			request: CreateOrganizationRequest{
				Name:        "Test Department",
				UnitType:    "DEPARTMENT",
				Status:      "ACTIVE",
				Level:       2,
				Description: stringPtr(string(make([]byte, 501))), // 501 characters
			},
			expectError: true,
			errorFields: []string{"description"},
		},
		{
			name: "Multiple validation errors",
			request: CreateOrganizationRequest{
				Name:     "", // Empty name
				UnitType: "INVALID", // Invalid unit type
				Status:   "INVALID", // Invalid status
				Level:    0, // Invalid level
			},
			expectError: true,
			errorFields: []string{"name", "unit_type", "status", "level"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			
			if test.expectError {
				if err == nil {
					t.Error("Expected validation error, but got none")
					return
				}
				
				// Check if it's a ValidationErrors type
				if validationErrors, ok := err.(*ValidationErrors); ok {
					actualFields := make(map[string]bool)
					for _, e := range validationErrors.Errors() {
						actualFields[e.Field] = true
					}
					
					for _, expectedField := range test.errorFields {
						if !actualFields[expectedField] {
							t.Errorf("Expected error for field %q, but not found in: %v", 
								expectedField, validationErrors.Errors())
						}
					}
				} else {
					t.Errorf("Expected ValidationErrors type, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestUpdateOrganizationRequestValidate(t *testing.T) {
	tests := []struct {
		name        string
		request     UpdateOrganizationRequest
		expectError bool
		errorFields []string
	}{
		{
			name: "Valid complete update",
			request: UpdateOrganizationRequest{
				Name:        stringPtr("Updated Department"),
				Status:      stringPtr("INACTIVE"),
				SortOrder:   intPtr(5),
				Description: stringPtr("Updated description"),
			},
			expectError: false,
		},
		{
			name: "Valid partial update",
			request: UpdateOrganizationRequest{
				Name: stringPtr("Updated Department"),
			},
			expectError: false,
		},
		{
			name: "Valid empty update",
			request: UpdateOrganizationRequest{},
			expectError: false,
		},
		{
			name: "Empty name",
			request: UpdateOrganizationRequest{
				Name: stringPtr(""),
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "Name too long",
			request: UpdateOrganizationRequest{
				Name: stringPtr(string(make([]byte, 101))), // 101 characters
			},
			expectError: true,
			errorFields: []string{"name"},
		},
		{
			name: "Invalid status",
			request: UpdateOrganizationRequest{
				Status: stringPtr("INVALID_STATUS"),
			},
			expectError: true,
			errorFields: []string{"status"},
		},
		{
			name: "Negative sort order",
			request: UpdateOrganizationRequest{
				SortOrder: intPtr(-1),
			},
			expectError: true,
			errorFields: []string{"sort_order"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.request.Validate()
			
			if test.expectError {
				if err == nil {
					t.Error("Expected validation error, but got none")
					return
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}

func TestValidateCreateOrganizationRequestMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectContext  bool
	}{
		{
			name: "Valid request",
			requestBody: CreateOrganizationRequest{
				Name:     "Test Department",
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectedStatus: http.StatusOK,
			expectContext:  true,
		},
		{
			name:           "Invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectContext:  false,
		},
		{
			name: "Validation error",
			requestBody: CreateOrganizationRequest{
				Name:     "", // Empty name will cause validation error
				UnitType: "DEPARTMENT",
				Status:   "ACTIVE",
				Level:    2,
			},
			expectedStatus: http.StatusBadRequest,
			expectContext:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create request body
			var requestBody []byte
			if str, ok := test.requestBody.(string); ok {
				requestBody = []byte(str)
			} else {
				var err error
				requestBody, err = json.Marshal(test.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(requestBody))
			req.Header.Set("Content-Type", "application/json")
			
			// Create response recorder
			recorder := httptest.NewRecorder()
			
			// Create next handler that checks context
			var contextChecked bool
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, ok := GetValidatedCreateRequest(r.Context())
				contextChecked = ok
				w.WriteHeader(http.StatusOK)
			})
			
			// Create middleware and execute
			middleware := ValidateCreateOrganizationRequest(nextHandler)
			middleware.ServeHTTP(recorder, req)
			
			// Check response status
			if recorder.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, recorder.Code)
			}
			
			// Check context availability
			if test.expectContext && !contextChecked {
				t.Error("Expected validated request in context, but not found")
			} else if !test.expectContext && contextChecked {
				t.Error("Did not expect validated request in context, but found")
			}
		})
	}
}

func TestValidateUpdateOrganizationRequestMiddleware(t *testing.T) {
	// Create valid request
	updateRequest := UpdateOrganizationRequest{
		Name:   stringPtr("Updated Department"),
		Status: stringPtr("INACTIVE"),
	}
	
	requestBody, err := json.Marshal(updateRequest)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("PUT", "/test", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	recorder := httptest.NewRecorder()
	
	var contextChecked bool
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		validatedReq, ok := GetValidatedUpdateRequest(r.Context())
		contextChecked = ok
		if ok && validatedReq.Name != nil && *validatedReq.Name != "Updated Department" {
			t.Errorf("Expected name 'Updated Department', got %q", *validatedReq.Name)
		}
		w.WriteHeader(http.StatusOK)
	})
	
	middleware := ValidateUpdateOrganizationRequest(nextHandler)
	middleware.ServeHTTP(recorder, req)
	
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}
	
	if !contextChecked {
		t.Error("Expected validated request in context, but not found")
	}
}

func TestGetValidatedRequests(t *testing.T) {
	// Test GetValidatedCreateRequest
	createReq := &CreateOrganizationRequest{
		Name: "Test Department",
	}
	
	ctx := context.WithValue(context.Background(), "validated_create_request", createReq)
	
	result, ok := GetValidatedCreateRequest(ctx)
	if !ok {
		t.Error("Expected to find validated create request in context")
	}
	if result.Name != "Test Department" {
		t.Errorf("Expected name 'Test Department', got %q", result.Name)
	}
	
	// Test with empty context
	emptyCtx := context.Background()
	_, ok = GetValidatedCreateRequest(emptyCtx)
	if ok {
		t.Error("Should not find validated request in empty context")
	}
	
	// Test GetValidatedUpdateRequest
	updateReq := &UpdateOrganizationRequest{
		Name: stringPtr("Updated Department"),
	}
	
	updateCtx := context.WithValue(context.Background(), "validated_update_request", updateReq)
	
	updateResult, ok := GetValidatedUpdateRequest(updateCtx)
	if !ok {
		t.Error("Expected to find validated update request in context")
	}
	if updateResult.Name == nil || *updateResult.Name != "Updated Department" {
		t.Error("Expected name 'Updated Department' in update request")
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}