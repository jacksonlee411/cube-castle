package intelligencegateway

import (
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/common"
)

// Conversation 对话模型
type Conversation struct {
	common.TenantEntity
	UserID    *uuid.UUID `json:"user_id,omitempty" db:"user_id"`
	SessionID string     `json:"session_id" db:"session_id"`
	Status    string     `json:"status" db:"status"`
}

// Message 消息模型
type Message struct {
	common.BaseEntity
	ConversationID uuid.UUID       `json:"conversation_id" db:"conversation_id"`
	UserText       *string         `json:"user_text,omitempty" db:"user_text"`
	AIResponse     *string         `json:"ai_response,omitempty" db:"ai_response"`
	Intent         *string         `json:"intent,omitempty" db:"intent"`
	Entities       map[string]any  `json:"entities,omitempty" db:"entities"`
	Confidence     *float64        `json:"confidence,omitempty" db:"confidence"`
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID string     `json:"session_id" validate:"required"`
}

// CreateMessageRequest 创建消息请求
type CreateMessageRequest struct {
	ConversationID uuid.UUID      `json:"conversation_id" validate:"required"`
	UserText       *string        `json:"user_text,omitempty"`
	AIResponse     *string        `json:"ai_response,omitempty"`
	Intent         *string        `json:"intent,omitempty"`
	Entities       map[string]any `json:"entities,omitempty"`
	Confidence     *float64       `json:"confidence,omitempty"`
}

// InterpretRequest 意图识别请求
type InterpretRequest struct {
	UserText  string         `json:"user_text" validate:"required"`
	SessionId string         `json:"session_id" validate:"required"`  // 注意：使用SessionId而不是SessionID以匹配gRPC字段
	Context   map[string]any `json:"context,omitempty"`
}

// InterpretResponse 意图识别响应
type InterpretResponse struct {
	Intent              string         `json:"intent"`
	Confidence          float64        `json:"confidence"`
	Entities            map[string]any `json:"entities,omitempty"`
	StructuredData      map[string]any `json:"structured_data,omitempty"`
	StructuredDataJson  string         `json:"structured_data_json,omitempty"` // 添加JSON字符串字段
	SuggestedActions    []string       `json:"suggested_actions,omitempty"`
	ResponseMessage     string         `json:"response_message,omitempty"`
}

// InterpretUserQueryRequest 用户查询解释请求
type InterpretUserQueryRequest struct {
	Query    string    `json:"query" validate:"required"`
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

// InterpretUserQueryResponse 用户查询解释响应
type InterpretUserQueryResponse struct {
	Intent             string    `json:"intent"`
	StructuredDataJson string    `json:"structured_data_json,omitempty"`
	ProcessedAt        time.Time `json:"processed_at"`
}

// ConversationResponse 对话响应
type ConversationResponse struct {
	Conversation
	Messages []MessageResponse `json:"messages,omitempty"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	Message
	Conversation *ConversationResponse `json:"conversation,omitempty"`
}

// IntentDefinition 意图定义
type IntentDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]Parameter   `json:"parameters,omitempty"`
	Examples    []string               `json:"examples,omitempty"`
}

// Parameter 参数定义
type Parameter struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// ConversationSearchRequest 对话搜索请求
type ConversationSearchRequest struct {
	common.Pagination
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID *string    `json:"session_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	FromDate  *time.Time `json:"from_date,omitempty"`
	ToDate    *time.Time `json:"to_date,omitempty"`
}

// AIProviderConfig AI提供商配置
type AIProviderConfig struct {
	Provider    string            `json:"provider"`
	APIKey      string            `json:"api_key"`
	BaseURL     string            `json:"base_url"`
	Model       string            `json:"model"`
	MaxTokens   int               `json:"max_tokens"`
	Temperature float64           `json:"temperature"`
	Headers     map[string]string `json:"headers,omitempty"`
}

// ConversationStats 对话统计
type ConversationStats struct {
	TotalConversations int     `json:"total_conversations"`
	TotalMessages      int     `json:"total_messages"`
	AvgMessagesPerConv float64 `json:"avg_messages_per_conv"`
	TopIntents         []IntentStats `json:"top_intents"`
}

// IntentStats 意图统计
type IntentStats struct {
	Intent      string  `json:"intent"`
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
} 