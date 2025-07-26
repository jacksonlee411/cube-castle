package intelligencegateway

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service IntelligenceGateway服务
type Service struct {
	contextStore map[string]*ConversationContext
	mu           sync.RWMutex
}

// ConversationContext 对话上下文
type ConversationContext struct {
	UserID    uuid.UUID   `json:"user_id"`
	SessionID string      `json:"session_id"`
	History   []InternalMessage `json:"history"`
	Context   any `json:"context"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// InternalMessage 内部消息结构（用于对话上下文）
type InternalMessage struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// BatchRequest 批处理请求
type BatchRequest struct {
	Requests []InterpretUserQueryRequest `json:"requests"`
	BatchID  string                       `json:"batch_id"`
}

// BatchResponse 批处理响应
type BatchResponse struct {
	BatchID   string                        `json:"batch_id"`
	Responses []InterpretUserQueryResponse `json:"responses"`
	Status    string                        `json:"status"`
	StartedAt time.Time                     `json:"started_at"`
}

// NewService 创建新的IntelligenceGateway服务
func NewService() *Service {
	return &Service{
		contextStore: make(map[string]*ConversationContext),
	}
}

// InterpretUserQuery 解释用户查询
func (s *Service) InterpretUserQuery(ctx context.Context, req *InterpretUserQueryRequest) (*InterpretUserQueryResponse, error) {
	// 验证请求
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	// 更新对话上下文
	s.updateConversationContext(req)

	// 模拟AI处理
	response := &InterpretUserQueryResponse{
		Intent:             "general_query",
		StructuredDataJson: fmt.Sprintf(`{"query": "%s", "processed": true}`, req.Query),
		ProcessedAt:        time.Now(),
	}

	// 添加助手响应到历史
	s.addAssistantResponse(req.UserID, req.TenantID, response.StructuredDataJson)

	return response, nil
}

// GetConversationContext 获取对话上下文
func (s *Service) GetConversationContext(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (*ConversationContext, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	sessionKey := fmt.Sprintf("%s-%s", userID.String(), tenantID.String())
	context, exists := s.contextStore[sessionKey]
	if !exists {
		return nil, fmt.Errorf("conversation context not found")
	}
	return context, nil
}

// ProcessBatchRequests 批量处理请求
func (s *Service) ProcessBatchRequests(ctx context.Context, batchReq *BatchRequest) (*BatchResponse, error) {
	if len(batchReq.Requests) == 0 {
		return nil, fmt.Errorf("batch requests cannot be empty")
	}

	responses := make([]InterpretUserQueryResponse, 0, len(batchReq.Requests))
	
	for _, req := range batchReq.Requests {
		resp, err := s.InterpretUserQuery(ctx, &req)
		if err != nil {
			// 记录错误但继续处理其他请求
			resp = &InterpretUserQueryResponse{
				Intent:             "error",
				StructuredDataJson: fmt.Sprintf(`{"error": "%s"}`, err.Error()),
				ProcessedAt:        time.Now(),
			}
		}
		responses = append(responses, *resp)
	}

	return &BatchResponse{
		BatchID:   batchReq.BatchID,
		Responses: responses,
		Status:    "completed",
		StartedAt: time.Now(),
	}, nil
}

// ClearContext 清除用户上下文
func (s *Service) ClearContext(userID uuid.UUID, tenantID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	sessionKey := fmt.Sprintf("%s-%s", userID.String(), tenantID.String())
	delete(s.contextStore, sessionKey)
	return nil
}

// GetContextStats 获取上下文统计
func (s *Service) GetContextStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	totalContexts := len(s.contextStore)
	totalMessages := 0
	
	for _, context := range s.contextStore {
		totalMessages += len(context.History)
	}
	
	avgMessages := 0.0
	if totalContexts > 0 {
		avgMessages = float64(totalMessages) / float64(totalContexts)
	}
	
	return map[string]interface{}{
		"total_contexts":        totalContexts,
		"total_messages":        totalMessages,
		"avg_messages_per_context": avgMessages,
	}
}

// 私有方法

func (s *Service) validateRequest(req *InterpretUserQueryRequest) error {
	if req.Query == "" {
		return fmt.Errorf("query cannot be empty")
	}
	if req.UserID == uuid.Nil {
		return fmt.Errorf("user_id cannot be empty")
	}
	if req.TenantID == uuid.Nil {
		return fmt.Errorf("tenant_id cannot be empty")
	}
	if len(req.Query) > 5000 {
		return fmt.Errorf("query is too long")
	}
	return nil
}

func (s *Service) updateConversationContext(req *InterpretUserQueryRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	sessionKey := fmt.Sprintf("%s-%s", req.UserID.String(), req.TenantID.String())
	
	context, exists := s.contextStore[sessionKey]
	if !exists {
		context = &ConversationContext{
			UserID:    req.UserID,
			SessionID: sessionKey,
			History:   []InternalMessage{},
			CreatedAt: time.Now(),
		}
		s.contextStore[sessionKey] = context
	}

	// 添加用户消息到历史记录
	context.History = append(context.History, InternalMessage{
		Role:      "user",
		Content:   req.Query,
		Timestamp: time.Now(),
	})
	context.UpdatedAt = time.Now()

	// 限制历史记录长度
	if len(context.History) > 50 {
		context.History = context.History[len(context.History)-50:]
	}
}

func (s *Service) addAssistantResponse(userID, tenantID uuid.UUID, response string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	sessionKey := fmt.Sprintf("%s-%s", userID.String(), tenantID.String())
	
	if context, exists := s.contextStore[sessionKey]; exists {
		context.History = append(context.History, InternalMessage{
			Role:      "assistant",
			Content:   response,
			Timestamp: time.Now(),
		})
		context.UpdatedAt = time.Now()
		
		// 限制历史记录长度
		if len(context.History) > 50 {
			context.History = context.History[len(context.History)-50:]
		}
	}
}