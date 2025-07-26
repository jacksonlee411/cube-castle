package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	StatusPending   WorkflowStatus = "pending"
	StatusRunning   WorkflowStatus = "running"
	StatusCompleted WorkflowStatus = "completed"
	StatusFailed    WorkflowStatus = "failed"
	StatusCanceled  WorkflowStatus = "canceled"
)

// WorkflowExecution 工作流执行实例
type WorkflowExecution struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      WorkflowStatus         `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Steps       []WorkflowStep         `json:"steps"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Status      WorkflowStatus         `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration_ms"`
}

// WorkflowDefinition 工作流定义
type WorkflowDefinition struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Steps       []string `json:"steps"`
}

// ActivityFunc 活动函数类型
type ActivityFunc func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

// Engine 工作流引擎
type Engine struct {
	executions map[string]*WorkflowExecution
	workflows  map[string]*WorkflowDefinition
	activities map[string]ActivityFunc
	mu         sync.RWMutex
}

// NewEngine 创建新的工作流引擎
func NewEngine() *Engine {
	engine := &Engine{
		executions: make(map[string]*WorkflowExecution),
		workflows:  make(map[string]*WorkflowDefinition),
		activities: make(map[string]ActivityFunc),
	}
	
	// 注册默认活动
	engine.registerDefaultActivities()
	
	return engine
}

// RegisterWorkflow 注册工作流定义
func (e *Engine) RegisterWorkflow(definition *WorkflowDefinition) error {
	if definition == nil {
		return fmt.Errorf("workflow definition cannot be nil")
	}
	
	if definition.ID == "" {
		return fmt.Errorf("workflow ID cannot be empty")
	}
	
	if definition.Name == "" {
		return fmt.Errorf("workflow name cannot be empty")
	}
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.workflows[definition.ID] = definition
	return nil
}

// RegisterActivity 注册活动
func (e *Engine) RegisterActivity(name string, activity ActivityFunc) error {
	if name == "" {
		return fmt.Errorf("activity name cannot be empty")
	}
	
	if activity == nil {
		return fmt.Errorf("activity function cannot be nil")
	}
	
	e.mu.Lock()
	defer e.mu.Unlock()
	
	e.activities[name] = activity
	return nil
}

// StartWorkflow 启动工作流
func (e *Engine) StartWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowExecution, error) {
	e.mu.RLock()
	workflow, exists := e.workflows[workflowID]
	e.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("workflow '%s' not found", workflowID)
	}
	
	execution := &WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     StatusPending,
		Input:      input,
		StartedAt:  time.Now(),
		Steps:      make([]WorkflowStep, 0),
	}
	
	e.mu.Lock()
	e.executions[execution.ID] = execution
	e.mu.Unlock()
	
	// 异步执行工作流
	go e.executeWorkflow(ctx, execution, workflow)
	
	return execution, nil
}

// GetExecution 获取工作流执行实例
func (e *Engine) GetExecution(executionID string) (*WorkflowExecution, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	execution, exists := e.executions[executionID]
	if !exists {
		return nil, fmt.Errorf("execution '%s' not found", executionID)
	}
	
	return execution, nil
}

// ListExecutions 列出所有执行实例
func (e *Engine) ListExecutions() []*WorkflowExecution {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	executions := make([]*WorkflowExecution, 0, len(e.executions))
	for _, execution := range e.executions {
		executions = append(executions, execution)
	}
	
	return executions
}

// GetWorkflowStats 获取工作流统计信息
func (e *Engine) GetWorkflowStats() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()
	
	stats := map[string]interface{}{
		"total_workflows":  len(e.workflows),
		"total_executions": len(e.executions),
		"total_activities": len(e.activities),
	}
	
	// 按状态统计执行实例
	statusCounts := make(map[WorkflowStatus]int)
	for _, execution := range e.executions {
		statusCounts[execution.Status]++
	}
	
	stats["executions_by_status"] = statusCounts
	
	return stats
}

// CancelExecution 取消工作流执行
func (e *Engine) CancelExecution(executionID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	
	execution, exists := e.executions[executionID]
	if !exists {
		return fmt.Errorf("execution '%s' not found", executionID)
	}
	
	if execution.Status == StatusCompleted || execution.Status == StatusFailed {
		return fmt.Errorf("cannot cancel execution in status '%s'", execution.Status)
	}
	
	execution.Status = StatusCanceled
	now := time.Now()
	execution.CompletedAt = &now
	
	return nil
}

// 私有方法

func (e *Engine) executeWorkflow(ctx context.Context, execution *WorkflowExecution, workflow *WorkflowDefinition) {
	e.mu.Lock()
	execution.Status = StatusRunning
	e.mu.Unlock()
	
	defer func() {
		if r := recover(); r != nil {
			e.mu.Lock()
			execution.Status = StatusFailed
			execution.Error = fmt.Sprintf("workflow panicked: %v", r)
			now := time.Now()
			execution.CompletedAt = &now
			e.mu.Unlock()
		}
	}()
	
	currentInput := execution.Input
	
	for i, stepName := range workflow.Steps {
		// 检查是否被取消
		if execution.Status == StatusCanceled {
			return
		}
		
		step := WorkflowStep{
			ID:        fmt.Sprintf("step-%d", i+1),
			Name:      stepName,
			Status:    StatusRunning,
			Input:     currentInput,
			StartedAt: time.Now(),
		}
		
		e.mu.Lock()
		execution.Steps = append(execution.Steps, step)
		stepIndex := len(execution.Steps) - 1
		e.mu.Unlock()
		
		// 执行步骤
		output, err := e.executeStep(ctx, stepName, currentInput)
		
		e.mu.Lock()
		now := time.Now()
		execution.Steps[stepIndex].CompletedAt = &now
		execution.Steps[stepIndex].Duration = now.Sub(execution.Steps[stepIndex].StartedAt)
		
		if err != nil {
			execution.Steps[stepIndex].Status = StatusFailed
			execution.Steps[stepIndex].Error = err.Error()
			execution.Status = StatusFailed
			execution.Error = fmt.Sprintf("step '%s' failed: %v", stepName, err)
			execution.CompletedAt = &now
			e.mu.Unlock()
			return
		}
		
		execution.Steps[stepIndex].Status = StatusCompleted
		execution.Steps[stepIndex].Output = output
		e.mu.Unlock()
		
		// 将当前步骤的输出作为下一步骤的输入
		currentInput = output
	}
	
	// 工作流完成
	e.mu.Lock()
	execution.Status = StatusCompleted
	execution.Output = currentInput
	now := time.Now()
	execution.CompletedAt = &now
	e.mu.Unlock()
}

func (e *Engine) executeStep(ctx context.Context, stepName string, input map[string]interface{}) (map[string]interface{}, error) {
	e.mu.RLock()
	activity, exists := e.activities[stepName]
	e.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("activity '%s' not found", stepName)
	}
	
	return activity(ctx, input)
}

func (e *Engine) registerDefaultActivities() {
	// 验证活动
	e.activities["validate"] = func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(time.Millisecond * 100) // 模拟处理时间
		
		output := make(map[string]interface{})
		for k, v := range input {
			output[k] = v
		}
		output["validated"] = true
		output["validation_time"] = time.Now()
		
		return output, nil
	}
	
	// 处理活动
	e.activities["process"] = func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(time.Millisecond * 200) // 模拟处理时间
		
		output := make(map[string]interface{})
		for k, v := range input {
			output[k] = v
		}
		output["processed"] = true
		output["process_time"] = time.Now()
		
		return output, nil
	}
	
	// 通知活动
	e.activities["notify"] = func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(time.Millisecond * 50) // 模拟处理时间
		
		output := make(map[string]interface{})
		for k, v := range input {
			output[k] = v
		}
		output["notified"] = true
		output["notification_time"] = time.Now()
		
		return output, nil
	}
	
	// AI查询活动
	e.activities["ai_query"] = func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(time.Millisecond * 300) // 模拟AI处理时间
		
		query, ok := input["query"].(string)
		if !ok {
			return nil, fmt.Errorf("query field is required and must be string")
		}
		
		output := make(map[string]interface{})
		for k, v := range input {
			output[k] = v
		}
		output["ai_response"] = fmt.Sprintf("AI processed: %s", query)
		output["ai_confidence"] = 0.95
		output["ai_time"] = time.Now()
		
		return output, nil
	}
	
	// 批处理活动
	e.activities["batch_process"] = func(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
		time.Sleep(time.Millisecond * 500) // 模拟批处理时间
		
		batchSize, ok := input["batch_size"].(float64)
		if !ok {
			batchSize = 10 // 默认批大小
		}
		
		output := make(map[string]interface{})
		for k, v := range input {
			output[k] = v
		}
		output["batch_processed"] = true
		output["processed_count"] = int(batchSize)
		output["batch_time"] = time.Now()
		
		return output, nil
	}
}