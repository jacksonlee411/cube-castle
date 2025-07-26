package workflow

import (
	"context"
	"testing"
	"time"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	
	if engine == nil {
		t.Fatal("Expected engine to be created, got nil")
	}
	
	if engine.executions == nil {
		t.Error("Expected executions map to be initialized")
	}
	
	if engine.workflows == nil {
		t.Error("Expected workflows map to be initialized")
	}
	
	if engine.activities == nil {
		t.Error("Expected activities map to be initialized")
	}
	
	// Check that default activities are registered
	expectedActivities := []string{"validate", "process", "notify", "ai_query", "batch_process"}
	for _, activity := range expectedActivities {
		if _, exists := engine.activities[activity]; !exists {
			t.Errorf("Expected default activity '%s' to be registered", activity)
		}
	}
}

func TestEngine_RegisterWorkflow(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name       string
		definition *WorkflowDefinition
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "nil definition",
			definition: nil,
			wantErr:    true,
			errMsg:     "workflow definition cannot be nil",
		},
		{
			name: "empty ID",
			definition: &WorkflowDefinition{
				ID:   "",
				Name: "Test Workflow",
			},
			wantErr: true,
			errMsg:  "workflow ID cannot be empty",
		},
		{
			name: "empty name",
			definition: &WorkflowDefinition{
				ID:   "test-workflow",
				Name: "",
			},
			wantErr: true,
			errMsg:  "workflow name cannot be empty",
		},
		{
			name: "valid definition",
			definition: &WorkflowDefinition{
				ID:          "test-workflow",
				Name:        "Test Workflow",
				Description: "A test workflow",
				Steps:       []string{"validate", "process", "notify"},
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterWorkflow(tt.definition)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				
				// Check that workflow was registered
				if workflow, exists := engine.workflows[tt.definition.ID]; !exists {
					t.Error("Expected workflow to be registered")
				} else {
					if workflow.Name != tt.definition.Name {
						t.Errorf("Expected workflow name '%s', got '%s'", tt.definition.Name, workflow.Name)
					}
				}
			}
		})
	}
}

func TestEngine_RegisterActivity(t *testing.T) {
	engine := NewEngine()
	
	tests := []struct {
		name     string
		actName  string
		activity ActivityFunc
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "empty name",
			actName: "",
			activity: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				return input, nil
			},
			wantErr: true,
			errMsg:  "activity name cannot be empty",
		},
		{
			name:     "nil activity",
			actName:  "test-activity",
			activity: nil,
			wantErr:  true,
			errMsg:   "activity function cannot be nil",
		},
		{
			name:    "valid activity",
			actName: "test-activity",
			activity: func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
				return input, nil
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.RegisterActivity(tt.actName, tt.activity)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got none")
				} else if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				
				// Check that activity was registered
				if _, exists := engine.activities[tt.actName]; !exists {
					t.Error("Expected activity to be registered")
				}
			}
		})
	}
}

func TestEngine_StartWorkflow(t *testing.T) {
	engine := NewEngine()
	
	// Register a test workflow
	workflow := &WorkflowDefinition{
		ID:    "test-workflow",
		Name:  "Test Workflow",
		Steps: []string{"validate", "process"},
	}
	
	err := engine.RegisterWorkflow(workflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}
	
	ctx := context.Background()
	input := map[string]interface{}{
		"test_data": "hello world",
	}
	
	execution, err := engine.StartWorkflow(ctx, "test-workflow", input)
	
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if execution == nil {
		t.Fatal("Expected execution to be returned, got nil")
	}
	
	if execution.ID == "" {
		t.Error("Expected execution ID to be generated")
	}
	
	if execution.WorkflowID != "test-workflow" {
		t.Errorf("Expected WorkflowID 'test-workflow', got '%s'", execution.WorkflowID)
	}
	
	if execution.Status != StatusPending {
		t.Errorf("Expected initial status 'pending', got '%s'", execution.Status)
	}
	
	if execution.Input["test_data"] != "hello world" {
		t.Errorf("Expected input data to be preserved")
	}
	
	// Wait a bit for workflow execution to start
	time.Sleep(time.Millisecond * 100)
	
	// Get updated execution
	updatedExecution, err := engine.GetExecution(execution.ID)
	if err != nil {
		t.Fatalf("Failed to get execution: %v", err)
	}
	
	// Status should have changed from pending
	if updatedExecution.Status == StatusPending {
		t.Error("Expected status to change from pending")
	}
}

func TestEngine_StartWorkflow_NonExistentWorkflow(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()
	
	execution, err := engine.StartWorkflow(ctx, "non-existent", map[string]interface{}{})
	
	if err == nil {
		t.Error("Expected error for non-existent workflow")
	}
	
	if execution != nil {
		t.Error("Expected no execution for non-existent workflow")
	}
	
	expectedErr := "workflow 'non-existent' not found"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message '%s', got '%s'", expectedErr, err.Error())
	}
}

func TestEngine_GetExecution(t *testing.T) {
	engine := NewEngine()
	
	// Test getting non-existent execution
	execution, err := engine.GetExecution("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent execution")
	}
	if execution != nil {
		t.Error("Expected nil execution for non-existent execution")
	}
	
	// Register and start a workflow
	workflow := &WorkflowDefinition{
		ID:    "test-workflow",
		Name:  "Test Workflow",
		Steps: []string{"validate"},
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	startedExecution, _ := engine.StartWorkflow(ctx, "test-workflow", map[string]interface{}{})
	
	// Get the execution
	execution, err = engine.GetExecution(startedExecution.ID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if execution == nil {
		t.Fatal("Expected execution, got nil")
	}
	
	if execution.ID != startedExecution.ID {
		t.Errorf("Expected execution ID '%s', got '%s'", startedExecution.ID, execution.ID)
	}
}

func TestEngine_ListExecutions(t *testing.T) {
	engine := NewEngine()
	
	// Initially should be empty
	executions := engine.ListExecutions()
	if len(executions) != 0 {
		t.Errorf("Expected 0 executions initially, got %d", len(executions))
	}
	
	// Register and start workflows
	workflow := &WorkflowDefinition{
		ID:    "test-workflow",
		Name:  "Test Workflow",
		Steps: []string{"validate"},
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	
	// Start multiple executions
	for i := 0; i < 3; i++ {
		_, err := engine.StartWorkflow(ctx, "test-workflow", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to start workflow %d: %v", i, err)
		}
	}
	
	// List executions
	executions = engine.ListExecutions()
	if len(executions) != 3 {
		t.Errorf("Expected 3 executions, got %d", len(executions))
	}
}

func TestEngine_GetWorkflowStats(t *testing.T) {
	engine := NewEngine()
	
	// Get initial stats
	stats := engine.GetWorkflowStats()
	
	if stats["total_workflows"] != 0 {
		t.Errorf("Expected 0 workflows initially, got %v", stats["total_workflows"])
	}
	
	if stats["total_executions"] != 0 {
		t.Errorf("Expected 0 executions initially, got %v", stats["total_executions"])
	}
	
	// Should have default activities
	if stats["total_activities"].(int) <= 0 {
		t.Error("Expected some default activities to be registered")
	}
	
	// Register workflow and start executions
	workflow := &WorkflowDefinition{
		ID:    "test-workflow",
		Name:  "Test Workflow",
		Steps: []string{"validate"},
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	engine.StartWorkflow(ctx, "test-workflow", map[string]interface{}{})
	
	// Get updated stats
	stats = engine.GetWorkflowStats()
	
	if stats["total_workflows"] != 1 {
		t.Errorf("Expected 1 workflow, got %v", stats["total_workflows"])
	}
	
	if stats["total_executions"] != 1 {
		t.Errorf("Expected 1 execution, got %v", stats["total_executions"])
	}
	
	// Check executions by status
	if statusCounts, ok := stats["executions_by_status"].(map[WorkflowStatus]int); ok {
		totalStatusCount := 0
		for _, count := range statusCounts {
			totalStatusCount += count
		}
		if totalStatusCount != 1 {
			t.Errorf("Expected total status count 1, got %d", totalStatusCount)
		}
	} else {
		t.Error("Expected executions_by_status to be present")
	}
}

func TestEngine_CancelExecution(t *testing.T) {
	engine := NewEngine()
	
	// Test canceling non-existent execution
	err := engine.CancelExecution("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent execution")
	}
	
	// Register and start a workflow
	workflow := &WorkflowDefinition{
		ID:    "test-workflow",
		Name:  "Test Workflow",
		Steps: []string{"validate", "process", "notify"}, // Long workflow
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	execution, _ := engine.StartWorkflow(ctx, "test-workflow", map[string]interface{}{})
	
	// Cancel the execution
	err = engine.CancelExecution(execution.ID)
	if err != nil {
		t.Errorf("Expected no error canceling execution, got %v", err)
	}
	
	// Check that execution was canceled
	updatedExecution, _ := engine.GetExecution(execution.ID)
	if updatedExecution.Status != StatusCanceled {
		t.Errorf("Expected status 'canceled', got '%s'", updatedExecution.Status)
	}
	
	if updatedExecution.CompletedAt == nil {
		t.Error("Expected CompletedAt to be set when canceled")
	}
}

func TestEngine_WorkflowExecution_Complete(t *testing.T) {
	engine := NewEngine()
	
	// Register a simple workflow
	workflow := &WorkflowDefinition{
		ID:    "simple-workflow",
		Name:  "Simple Workflow",
		Steps: []string{"validate", "process"},
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	input := map[string]interface{}{
		"test_data": "hello",
	}
	
	execution, _ := engine.StartWorkflow(ctx, "simple-workflow", input)
	
	// Wait for workflow to complete
	maxWait := time.Second * 2
	start := time.Now()
	
	for {
		if time.Since(start) > maxWait {
			t.Fatal("Workflow did not complete in time")
		}
		
		updatedExecution, _ := engine.GetExecution(execution.ID)
		if updatedExecution.Status == StatusCompleted || updatedExecution.Status == StatusFailed {
			// Check completion
			if updatedExecution.Status != StatusCompleted {
				t.Errorf("Expected workflow to complete successfully, got status '%s'", updatedExecution.Status)
				if updatedExecution.Error != "" {
					t.Errorf("Workflow error: %s", updatedExecution.Error)
				}
			}
			
			if updatedExecution.CompletedAt == nil {
				t.Error("Expected CompletedAt to be set")
			}
			
			if len(updatedExecution.Steps) != 2 {
				t.Errorf("Expected 2 steps, got %d", len(updatedExecution.Steps))
			}
			
			// Check that all steps completed
			for i, step := range updatedExecution.Steps {
				if step.Status != StatusCompleted {
					t.Errorf("Step %d: Expected status 'completed', got '%s'", i, step.Status)
				}
				if step.CompletedAt == nil {
					t.Errorf("Step %d: Expected CompletedAt to be set", i)
				}
				if step.Duration <= 0 {
					t.Errorf("Step %d: Expected positive duration, got %v", i, step.Duration)
				}
			}
			
			// Check that output is populated
			if updatedExecution.Output == nil {
				t.Error("Expected output to be populated")
			}
			
			break
		}
		
		time.Sleep(time.Millisecond * 50)
	}
}

func TestEngine_ActivityExecution(t *testing.T) {
	engine := NewEngine()
	
	// Test default activities
	ctx := context.Background()
	
	tests := []struct {
		name     string
		activity string
		input    map[string]interface{}
		checkFn  func(t *testing.T, output map[string]interface{}, err error)
	}{
		{
			name:     "validate activity",
			activity: "validate",
			input:    map[string]interface{}{"data": "test"},
			checkFn: func(t *testing.T, output map[string]interface{}, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if output["validated"] != true {
					t.Error("Expected validated to be true")
				}
				if output["data"] != "test" {
					t.Error("Expected input data to be preserved")
				}
			},
		},
		{
			name:     "ai_query activity",
			activity: "ai_query",
			input:    map[string]interface{}{"query": "test query"},
			checkFn: func(t *testing.T, output map[string]interface{}, err error) {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if output["query"] != "test query" {
					t.Error("Expected query to be preserved")
				}
				if response, ok := output["ai_response"].(string); !ok || response == "" {
					t.Error("Expected ai_response to be populated")
				}
				if confidence, ok := output["ai_confidence"].(float64); !ok || confidence <= 0 {
					t.Error("Expected ai_confidence to be populated")
				}
			},
		},
		{
			name:     "ai_query activity missing query",
			activity: "ai_query",
			input:    map[string]interface{}{"data": "test"},
			checkFn: func(t *testing.T, output map[string]interface{}, err error) {
				if err == nil {
					t.Error("Expected error for missing query")
				}
				expectedErr := "query field is required and must be string"
				if err.Error() != expectedErr {
					t.Errorf("Expected error '%s', got '%s'", expectedErr, err.Error())
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			activity := engine.activities[tt.activity]
			output, err := activity(ctx, tt.input)
			tt.checkFn(t, output, err)
		})
	}
}

func BenchmarkEngine_StartWorkflow(b *testing.B) {
	engine := NewEngine()
	
	workflow := &WorkflowDefinition{
		ID:    "bench-workflow",
		Name:  "Benchmark Workflow",
		Steps: []string{"validate"},
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	input := map[string]interface{}{"data": "benchmark"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.StartWorkflow(ctx, "bench-workflow", input)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

func BenchmarkEngine_GetWorkflowStats(b *testing.B) {
	engine := NewEngine()
	
	// Create some test data
	workflow := &WorkflowDefinition{
		ID:    "bench-workflow",
		Name:  "Benchmark Workflow",
		Steps: []string{"validate"},
	}
	
	engine.RegisterWorkflow(workflow)
	
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		engine.StartWorkflow(ctx, "bench-workflow", map[string]interface{}{})
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.GetWorkflowStats()
	}
}