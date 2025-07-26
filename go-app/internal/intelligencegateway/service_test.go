package intelligencegateway

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestNewService(t *testing.T) {
	service := NewService()
	
	if service == nil {
		t.Fatal("Expected service to be created, got nil")
	}
	
	if service.contextStore == nil {
		t.Error("Expected contextStore to be initialized")
	}
	
	if len(service.contextStore) != 0 {
		t.Error("Expected contextStore to be empty initially")
	}
}

func TestService_InterpretUserQuery(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	
	userID := uuid.New()
	tenantID := uuid.New()
	
	req := &InterpretUserQueryRequest{
		Query:    "Test query",
		UserID:   userID,
		TenantID: tenantID,
	}
	
	resp, err := service.InterpretUserQuery(ctx, req)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	
	if resp.Intent != "general_query" {
		t.Errorf("Expected intent 'general_query', got %s", resp.Intent)
	}
	
	if resp.StructuredDataJson == "" {
		t.Error("Expected structured data json to be populated")
	}
	
	if resp.ProcessedAt.IsZero() {
		t.Error("Expected ProcessedAt to be set")
	}
	
	// Check that context was created
	context, err := service.GetConversationContext(ctx, userID, tenantID)
	if err != nil {
		t.Errorf("Expected context to be created, got error: %v", err)
	}
	
	if len(context.History) != 2 { // user message + assistant response
		t.Errorf("Expected 2 messages in history, got %d", len(context.History))
	}
}

func TestService_InterpretUserQuery_Validation(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	
	tests := []struct {
		name    string
		req     *InterpretUserQueryRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty query",
			req: &InterpretUserQueryRequest{
				Query:    "",
				UserID:   uuid.New(),
				TenantID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "query cannot be empty",
		},
		{
			name: "nil user ID",
			req: &InterpretUserQueryRequest{
				Query:    "test",
				UserID:   uuid.Nil,
				TenantID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "user_id cannot be empty",
		},
		{
			name: "nil tenant ID",
			req: &InterpretUserQueryRequest{
				Query:    "test",
				UserID:   uuid.New(),
				TenantID: uuid.Nil,
			},
			wantErr: true,
			errMsg:  "tenant_id cannot be empty",
		},
		{
			name: "query too long",
			req: &InterpretUserQueryRequest{
				Query:    string(make([]byte, 5001)), // 5001 characters
				UserID:   uuid.New(),
				TenantID: uuid.New(),
			},
			wantErr: true,
			errMsg:  "query is too long",
		},
		{
			name: "valid request",
			req: &InterpretUserQueryRequest{
				Query:    "valid query",
				UserID:   uuid.New(),
				TenantID: uuid.New(),
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := service.InterpretUserQuery(ctx, tt.req)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
				if resp != nil {
					t.Error("Expected no response when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if resp == nil {
					t.Error("Expected response, got nil")
				}
			}
		})
	}
}

func TestService_ProcessBatchRequests(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	
	userID1 := uuid.New()
	userID2 := uuid.New()
	tenantID := uuid.New()
	
	batchReq := &BatchRequest{
		BatchID: "test-batch-123",
		Requests: []InterpretUserQueryRequest{
			{
				Query:    "First query",
				UserID:   userID1,
				TenantID: tenantID,
			},
			{
				Query:    "Second query",
				UserID:   userID2,
				TenantID: tenantID,
			},
		},
	}
	
	resp, err := service.ProcessBatchRequests(ctx, batchReq)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	
	if resp.BatchID != "test-batch-123" {
		t.Errorf("Expected BatchID 'test-batch-123', got %s", resp.BatchID)
	}
	
	if resp.Status != "completed" {
		t.Errorf("Expected status 'completed', got %s", resp.Status)
	}
	
	if len(resp.Responses) != 2 {
		t.Errorf("Expected 2 responses, got %d", len(resp.Responses))
	}
}

func TestService_GetContextStats(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	
	// Initial stats should be empty
	stats := service.GetContextStats()
	
	if stats["total_contexts"] != 0 {
		t.Errorf("Expected 0 total contexts, got %v", stats["total_contexts"])
	}
	
	// Create some contexts
	userID := uuid.New()
	tenantID := uuid.New()
	
	service.InterpretUserQuery(ctx, &InterpretUserQueryRequest{
		Query: "Test query", UserID: userID, TenantID: tenantID,
	})
	
	stats = service.GetContextStats()
	
	if stats["total_contexts"] != 1 {
		t.Errorf("Expected 1 total context, got %v", stats["total_contexts"])
	}
}