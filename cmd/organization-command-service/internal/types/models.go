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
	RecordID    string    `json:"recordId" db:"record_id"`
	TenantID    string    `json:"tenantId" db:"tenant_id"`
	Code        string    `json:"code" db:"code"`
	ParentCode  *string   `json:"parentCode,omitempty" db:"parent_code"`
	Name        string    `json:"name" db:"name"`
	UnitType    string    `json:"unitType" db:"unit_type"`
	Status      string    `json:"status" db:"status"`
	Level       int       `json:"level" db:"level"`
	Path        string    `json:"path" db:"path"`
	CodePath    string    `json:"codePath" db:"code_path"`
	NamePath    string    `json:"namePath" db:"name_path"`
	SortOrder   int       `json:"sortOrder" db:"sort_order"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effectiveDate,omitempty" db:"effective_date"`
	EndDate       *Date   `json:"endDate,omitempty" db:"end_date"`
	ChangeReason  *string `json:"changeReason,omitempty" db:"change_reason"`
	IsCurrent     bool    `json:"isCurrent" db:"is_current"`
}

// CreateOrganizationRequest 创建组织请求
type CreateOrganizationRequest struct {
	Code        *string `json:"code,omitempty"` // 可选：指定组织代码（用于时态记录）
	Name        string  `json:"name" validate:"required,max=100"`
	UnitType    string  `json:"unitType" validate:"required"`
	ParentCode  *string `json:"parentCode,omitempty"`
	SortOrder   int     `json:"sortOrder"`
	Description string  `json:"description"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date  `json:"effectiveDate,omitempty"`
	EndDate       *Date  `json:"endDate,omitempty"`
	ChangeReason  string `json:"changeReason,omitempty"`
}

// UpdateOrganizationRequest 更新组织请求
type UpdateOrganizationRequest struct {
	Name        *string `json:"name,omitempty"`
	UnitType    *string `json:"unitType,omitempty"`
	Status      *string `json:"status,omitempty"` // 添加状态字段
	SortOrder   *int    `json:"sortOrder,omitempty"`
	Description *string `json:"description,omitempty"`
	ParentCode  *string `json:"parentCode,omitempty"` // 通过修改parent_code来改变层级
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effectiveDate,omitempty"`
	EndDate       *Date   `json:"endDate,omitempty"`
	ChangeReason  *string `json:"changeReason,omitempty"`
}

// OrganizationResponse 组织响应
type OrganizationResponse struct {
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	UnitType    string    `json:"unitType"`
	Status      string    `json:"status"`
	Level       int       `json:"level"`
	Path        string    `json:"path"`
	CodePath    string    `json:"codePath"`
	NamePath    string    `json:"namePath"`
	SortOrder   int       `json:"sortOrder"`
	Description string    `json:"description"`
	ParentCode  *string   `json:"parentCode,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	// 时态管理字段 (使用Date类型)
	EffectiveDate *Date   `json:"effectiveDate,omitempty"`
	EndDate       *Date   `json:"endDate,omitempty"`
	ChangeReason  *string `json:"changeReason,omitempty"`
}

// CreateVersionRequest 为现有组织创建新时态版本的请求 (基于OpenAPI契约v4.4.0)
type CreateVersionRequest struct {
	Name            string  `json:"name" validate:"required,max=255"`
	UnitType        string  `json:"unitType" validate:"required"`
	ParentCode      *string `json:"parentCode,omitempty" validate:"omitempty,len=7"`
	Description     *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	SortOrder       *int    `json:"sortOrder,omitempty"`
	Profile         *string `json:"profile,omitempty"` // JSON string
	EffectiveDate   string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	EndDate         *string `json:"endDate,omitempty" validate:"omitempty,datetime=2006-01-02"`
	OperationReason string  `json:"operationReason" validate:"omitempty,max=500"`
}

// 组织历史版本请求 (旧版本，保持兼容性)
type CreateOrganizationVersionRequest struct {
	BasedOnVersion int     `json:"basedOnVersion"`
	Name           *string `json:"name,omitempty"`
	UnitType       *string `json:"unitType,omitempty"`
	Status         *string `json:"status,omitempty"`
	SortOrder      *int    `json:"sortOrder,omitempty"`
	Description    *string `json:"description,omitempty"`
	ParentCode     *string `json:"parentCode,omitempty"`
	EffectiveDate  Date    `json:"effectiveDate" validate:"required"`
	EndDate        *Date   `json:"endDate,omitempty"`
	ChangeReason   string  `json:"changeReason" validate:"required"`
}

// 时态查询响应（包含时间线信息）
type TemporalOrganizationResponse struct {
	*OrganizationResponse
	TemporalStatus string                    `json:"temporalStatus"`
	Timeline       []TemporalTimelineEvent   `json:"timeline,omitempty"`
	Versions       []OrganizationVersionInfo `json:"versions,omitempty"`
}

// 时间线事件
type TemporalTimelineEvent struct {
	EventType     string                 `json:"eventType"`
	EventDate     time.Time              `json:"eventDate"`
	EffectiveDate *Date                  `json:"effectiveDate,omitempty"`
	Status        string                 `json:"status"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// 版本信息
type OrganizationVersionInfo struct {
	Version       int       `json:"version"`
	EffectiveFrom Date      `json:"effectiveFrom"`
	EffectiveTo   *Date     `json:"effectiveTo,omitempty"`
	ChangeReason  string    `json:"changeReason"`
	CreatedAt     time.Time `json:"createdAt"`
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
	EventType     string                 `json:"eventType" validate:"required"`
	RecordID      string                 `json:"recordId,omitempty"` // 用于精确定位记录（作废时必需）
	EffectiveDate string                 `json:"effectiveDate" validate:"required"`
	ChangeData    map[string]interface{} `json:"changeData,omitempty"` // UPDATE时必需
	ChangeReason  string                 `json:"changeReason" validate:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}
