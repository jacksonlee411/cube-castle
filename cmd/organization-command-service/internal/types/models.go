package types

import (
	"time"

	"github.com/google/uuid"
)

// 默认租户配置
const (
	DefaultTenantIDString = "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
	DefaultTenantName     = "高谷集团"
)

var DefaultTenantID = uuid.MustParse(DefaultTenantIDString)

// Organization 组织业务实体
type Organization struct {
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	Code        string    `json:"code" db:"code"`
	ParentCode  *string   `json:"parent_code,omitempty" db:"parent_code"`
	Name        string    `json:"name" db:"name"`
	UnitType    string    `json:"unit_type" db:"unit_type"`
	Status      string    `json:"status" db:"status"`
	Level       int       `json:"level" db:"level"`
	Path        string    `json:"path" db:"path"`
	SortOrder   int       `json:"sort_order" db:"sort_order"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effective_date,omitempty" db:"effective_date"`
	EndDate       *Date   `json:"end_date,omitempty" db:"end_date"`
	IsTemporal    bool    `json:"is_temporal" db:"is_temporal"`
	ChangeReason  *string `json:"change_reason,omitempty" db:"change_reason"`
	IsCurrent     bool    `json:"is_current" db:"is_current"`
}

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Code        *string `json:"code,omitempty"`          // 可选：指定组织代码（用于时态记录）
	Name        string  `json:"name" validate:"required,max=100"`
	UnitType    string  `json:"unit_type" validate:"required"`
	ParentCode  *string `json:"parent_code,omitempty"`
	SortOrder   int     `json:"sort_order"`
	Description string  `json:"description"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date  `json:"effective_date,omitempty"`
	EndDate       *Date  `json:"end_date,omitempty"`
	IsTemporal    bool   `json:"is_temporal"`
	ChangeReason  string `json:"change_reason,omitempty"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty"`
	UnitType    *string `json:"unit_type,omitempty"`
	Status      *string `json:"status,omitempty"`      // 添加状态字段
	SortOrder   *int    `json:"sort_order,omitempty"`
	Description *string `json:"description,omitempty"`
	ParentCode  *string `json:"parent_code,omitempty"` // 通过修改parent_code来改变层级
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effective_date,omitempty"`
	EndDate       *Date   `json:"end_date,omitempty"`
	IsTemporal    *bool   `json:"is_temporal,omitempty"`
	ChangeReason  *string `json:"change_reason,omitempty"`
}

// OrganizationResponse 组织响应
type OrganizationResponse struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	UnitType    string    `json:"unit_type"`
	Status      string    `json:"status"`
	Level       int       `json:"level"`
	Path        string    `json:"path"`
	SortOrder   int       `json:"sort_order"`
	Description string    `json:"description"`
	ParentCode  *string   `json:"parent_code,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effective_date,omitempty"`
	EndDate       *Date   `json:"end_date,omitempty"`
	IsTemporal    bool    `json:"is_temporal"`
	ChangeReason  *string `json:"change_reason,omitempty"`
}

// 组织历史版本请求
type CreateOrganizationVersionRequest struct {
	BasedOnVersion int     `json:"based_on_version"`
	Name           *string `json:"name,omitempty"`
	UnitType       *string `json:"unit_type,omitempty"`
	Status         *string `json:"status,omitempty"`
	SortOrder      *int    `json:"sort_order,omitempty"`
	Description    *string `json:"description,omitempty"`
	ParentCode     *string `json:"parent_code,omitempty"`
	EffectiveDate  Date    `json:"effective_date" validate:"required"`
	EndDate        *Date   `json:"end_date,omitempty"`
	ChangeReason   string  `json:"change_reason" validate:"required"`
}

// 时态查询响应（包含时间线信息）
type TemporalOrganizationResponse struct {
	*OrganizationResponse
	TemporalStatus string                    `json:"temporal_status"`
	Timeline       []TemporalTimelineEvent   `json:"timeline,omitempty"`
	Versions       []OrganizationVersionInfo `json:"versions,omitempty"`
}

// 时间线事件
type TemporalTimelineEvent struct {
	EventType     string                 `json:"event_type"`
	EventDate     time.Time              `json:"event_date"`
	EffectiveDate *Date                  `json:"effective_date,omitempty"`
	Status        string                 `json:"status"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// 版本信息
type OrganizationVersionInfo struct {
	Version       int       `json:"version"`
	EffectiveFrom Date      `json:"effective_from"`
	EffectiveTo   *Date     `json:"effective_to,omitempty"`
	ChangeReason  string    `json:"change_reason"`
	CreatedAt     time.Time `json:"created_at"`
}

// 组织操作请求类型
type SuspendOrganizationRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type ReactivateOrganizationRequest struct {
	Reason string `json:"reason" validate:"required"`
}

// 组织事件请求类型 (用于时态版本管理)
type OrganizationEventRequest struct {
	EventType     string                 `json:"event_type" validate:"required"`
	RecordID      string                 `json:"record_id,omitempty"`       // 用于精确定位记录（作废时必需）
	EffectiveDate string                 `json:"effective_date" validate:"required"`
	ChangeData    map[string]interface{} `json:"change_data,omitempty"`     // UPDATE时必需
	ChangeReason  string                 `json:"change_reason" validate:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}