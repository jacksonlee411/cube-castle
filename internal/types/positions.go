package types

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// PositionRequest 用于创建/更新职位的请求载荷
type PositionRequest struct {
	Title                  string   `json:"title" validate:"required,max=120"`
	JobProfileCode         *string  `json:"jobProfileCode,omitempty" validate:"omitempty,max=64"`
	JobProfileName         *string  `json:"jobProfileName,omitempty" validate:"omitempty,max=255"`
	JobFamilyGroupCode     string   `json:"jobFamilyGroupCode" validate:"required"`
	JobFamilyGroupRecordID *string  `json:"jobFamilyGroupRecordId,omitempty"`
	JobFamilyCode          string   `json:"jobFamilyCode" validate:"required"`
	JobFamilyRecordID      *string  `json:"jobFamilyRecordId,omitempty"`
	JobRoleCode            string   `json:"jobRoleCode" validate:"required"`
	JobRoleRecordID        *string  `json:"jobRoleRecordId,omitempty"`
	JobLevelCode           string   `json:"jobLevelCode" validate:"required"`
	JobLevelRecordID       *string  `json:"jobLevelRecordId,omitempty"`
	OrganizationCode       string   `json:"organizationCode" validate:"required,len=7"`
	PositionType           string   `json:"positionType" validate:"required"`
	Status                 *string  `json:"status,omitempty" validate:"omitempty"`
	EmploymentType         string   `json:"employmentType" validate:"required"`
	HeadcountCapacity      float64  `json:"headcountCapacity" validate:"required"`
	HeadcountInUse         *float64 `json:"headcountInUse,omitempty"`
	GradeLevel             *string  `json:"gradeLevel,omitempty" validate:"omitempty,max=20"`
	CostCenterCode         *string  `json:"costCenterCode,omitempty" validate:"omitempty,max=50"`
	ReportsToPositionCode  *string  `json:"reportsToPositionCode,omitempty" validate:"omitempty,len=8"`
	Profile                *string  `json:"profile,omitempty"`
	EffectiveDate          string   `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	OperationReason        string   `json:"operationReason" validate:"required"`
}

// PositionVersionRequest 插入职位版本的请求
type PositionVersionRequest struct {
	Title                  string   `json:"title" validate:"required,max=120"`
	JobProfileCode         *string  `json:"jobProfileCode,omitempty" validate:"omitempty,max=64"`
	JobProfileName         *string  `json:"jobProfileName,omitempty" validate:"omitempty,max=255"`
	JobFamilyGroupCode     string   `json:"jobFamilyGroupCode" validate:"required"`
	JobFamilyGroupRecordID *string  `json:"jobFamilyGroupRecordId,omitempty"`
	JobFamilyCode          string   `json:"jobFamilyCode" validate:"required"`
	JobFamilyRecordID      *string  `json:"jobFamilyRecordId,omitempty"`
	JobRoleCode            string   `json:"jobRoleCode" validate:"required"`
	JobRoleRecordID        *string  `json:"jobRoleRecordId,omitempty"`
	JobLevelCode           string   `json:"jobLevelCode" validate:"required"`
	JobLevelRecordID       *string  `json:"jobLevelRecordId,omitempty"`
	PositionType           *string  `json:"positionType,omitempty"`
	EmploymentType         *string  `json:"employmentType,omitempty"`
	HeadcountCapacity      *float64 `json:"headcountCapacity,omitempty"`
	HeadcountInUse         *float64 `json:"headcountInUse,omitempty"`
	GradeLevel             *string  `json:"gradeLevel,omitempty"`
	CostCenterCode         *string  `json:"costCenterCode,omitempty"`
	ReportsTo              *string  `json:"reportsToPositionCode,omitempty"`
	Profile                *string  `json:"profile,omitempty"`
	EffectiveDate          string   `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	OperationReason        string   `json:"operationReason" validate:"required"`
}

// PositionEventRequest 职位事件（停用/激活/删除）
type PositionEventRequest struct {
	EventType       string  `json:"eventType" validate:"required"`
	RecordID        *string `json:"recordId,omitempty"`
	EffectiveDate   string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	OperationReason string  `json:"operationReason" validate:"required"`
}

// FillPositionRequest 填充职位请求
type FillPositionRequest struct {
	EmployeeID         string   `json:"employeeId" validate:"required,uuid4"`
	EmployeeName       string   `json:"employeeName" validate:"required,max=120"`
	EmployeeNumber     *string  `json:"employeeNumber,omitempty" validate:"omitempty,max=64"`
	AssignmentType     string   `json:"assignmentType" validate:"required"`
	FTE                *float64 `json:"fte,omitempty"`
	EffectiveDate      string   `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	AnticipatedEndDate *string  `json:"anticipatedEndDate,omitempty" validate:"omitempty,datetime=2006-01-02"`
	AutoRevert         *bool    `json:"autoRevert,omitempty"`
	OperationReason    string   `json:"operationReason" validate:"required"`
	Notes              *string  `json:"notes,omitempty" validate:"omitempty,max=500"`
}

type CreateAssignmentRequest struct {
	EmployeeID      string   `json:"employeeId" validate:"required,uuid4"`
	EmployeeName    string   `json:"employeeName" validate:"required,max=120"`
	EmployeeNumber  *string  `json:"employeeNumber,omitempty" validate:"omitempty,max=64"`
	AssignmentType  string   `json:"assignmentType" validate:"required"`
	FTE             *float64 `json:"fte,omitempty"`
	EffectiveDate   string   `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	ActingUntil     *string  `json:"actingUntil,omitempty" validate:"omitempty,datetime=2006-01-02"`
	AutoRevert      *bool    `json:"autoRevert,omitempty"`
	OperationReason string   `json:"operationReason" validate:"required"`
	Notes           *string  `json:"notes,omitempty" validate:"omitempty,max=500"`
}

type UpdateAssignmentRequest struct {
	FTE             *float64 `json:"fte,omitempty"`
	ActingUntil     *string  `json:"actingUntil,omitempty" validate:"omitempty,datetime=2006-01-02"`
	AutoRevert      *bool    `json:"autoRevert,omitempty"`
	Notes           *string  `json:"notes,omitempty" validate:"omitempty,max=500"`
	OperationReason string   `json:"operationReason" validate:"required"`
}

type CloseAssignmentRequest struct {
	EndDate         string  `json:"endDate" validate:"required,datetime=2006-01-02"`
	OperationReason string  `json:"operationReason" validate:"required"`
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

type AssignmentListFilter struct {
	AssignmentTypes   []string
	AssignmentStatus  *string
	AsOfDate          *time.Time
	IncludeHistorical bool
	IncludeActingOnly bool
}

type AssignmentListOptions struct {
	Filter   AssignmentListFilter
	Page     int
	PageSize int
}

type AssignmentUpdateParams struct {
	FTE               *float64
	ActingUntil       *time.Time
	ClearActingUntil  bool
	AutoRevert        *bool
	Notes             *string
	ReminderSentAt    *time.Time
	ClearReminderSent bool
}

type PaginationMeta struct {
	Total       int  `json:"total"`
	Page        int  `json:"page"`
	PageSize    int  `json:"pageSize"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

type PositionAssignmentListResponse struct {
	Data       []PositionAssignmentResponse `json:"data"`
	Pagination PaginationMeta               `json:"pagination"`
	TotalCount int                          `json:"totalCount"`
}

// VacatePositionRequest 清空职位请求
type VacatePositionRequest struct {
	AssignmentID    string  `json:"assignmentId" validate:"required,uuid4"`
	EffectiveDate   string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	OperationReason string  `json:"operationReason" validate:"required"`
	Notes           *string `json:"notes,omitempty" validate:"omitempty,max=500"`
}

// TransferPositionRequest 转移职位请求
type TransferPositionRequest struct {
	TargetOrganizationCode string `json:"targetOrganizationCode" validate:"required,len=7"`
	EffectiveDate          string `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	OperationReason        string `json:"operationReason" validate:"required"`
	ReassignReports        *bool  `json:"reassignReports,omitempty"`
}

// Job catalog requests
type CreateJobFamilyGroupRequest struct {
	Code          string  `json:"code" validate:"required"`
	Name          string  `json:"name" validate:"required"`
	Description   *string `json:"description,omitempty"`
	Status        string  `json:"status" validate:"required"`
	EffectiveDate string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
}

type CreateJobFamilyRequest struct {
	Code               string  `json:"code" validate:"required"`
	JobFamilyGroupCode string  `json:"jobFamilyGroupCode" validate:"required"`
	Name               string  `json:"name" validate:"required"`
	Description        *string `json:"description,omitempty"`
	Status             string  `json:"status" validate:"required"`
	EffectiveDate      string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
}

type CreateJobRoleRequest struct {
	Code            string                 `json:"code" validate:"required"`
	JobFamilyCode   string                 `json:"jobFamilyCode" validate:"required"`
	Name            string                 `json:"name" validate:"required"`
	Description     *string                `json:"description,omitempty"`
	Status          string                 `json:"status" validate:"required"`
	EffectiveDate   string                 `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	CompetencyModel map[string]interface{} `json:"competencyModel,omitempty"`
}

type CreateJobLevelRequest struct {
	Code          string                 `json:"code" validate:"required"`
	JobRoleCode   string                 `json:"jobRoleCode" validate:"required"`
	Name          string                 `json:"name" validate:"required"`
	Description   *string                `json:"description,omitempty"`
	Status        string                 `json:"status" validate:"required"`
	LevelRank     string                 `json:"levelRank" validate:"required"`
	EffectiveDate string                 `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	SalaryBand    map[string]interface{} `json:"salaryBand,omitempty"`
}

type UpdateJobFamilyGroupRequest struct {
	Name          string  `json:"name" validate:"required"`
	Status        string  `json:"status" validate:"required"`
	EffectiveDate string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	Description   *string `json:"description,omitempty"`
}

type UpdateJobFamilyRequest struct {
	Name               string  `json:"name" validate:"required"`
	Status             string  `json:"status" validate:"required"`
	EffectiveDate      string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	Description        *string `json:"description,omitempty"`
	JobFamilyGroupCode *string `json:"jobFamilyGroupCode,omitempty"`
}

type UpdateJobRoleRequest struct {
	Name          string  `json:"name" validate:"required"`
	Status        string  `json:"status" validate:"required"`
	EffectiveDate string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	Description   *string `json:"description,omitempty"`
	JobFamilyCode *string `json:"jobFamilyCode,omitempty"`
}

type UpdateJobLevelRequest struct {
	Name          string  `json:"name" validate:"required"`
	Status        string  `json:"status" validate:"required"`
	EffectiveDate string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	Description   *string `json:"description,omitempty"`
	JobRoleCode   *string `json:"jobRoleCode,omitempty"`
	LevelRank     *int    `json:"levelRank,omitempty"`
}

type JobCatalogVersionRequest struct {
	Name           string  `json:"name" validate:"required"`
	Status         string  `json:"status" validate:"required"`
	EffectiveDate  string  `json:"effectiveDate" validate:"required,datetime=2006-01-02"`
	Description    *string `json:"description,omitempty"`
	ParentRecordID *string `json:"parentRecordId,omitempty"`
}

// PositionResponse 响应结构
type PositionResponse struct {
	Code                  string                       `json:"code"`
	Title                 string                       `json:"title"`
	JobProfileCode        *string                      `json:"jobProfileCode,omitempty"`
	JobProfileName        *string                      `json:"jobProfileName,omitempty"`
	JobFamilyGroupCode    string                       `json:"jobFamilyGroupCode"`
	JobFamilyGroupName    string                       `json:"jobFamilyGroupName"`
	JobFamilyCode         string                       `json:"jobFamilyCode"`
	JobFamilyName         string                       `json:"jobFamilyName"`
	JobRoleCode           string                       `json:"jobRoleCode"`
	JobRoleName           string                       `json:"jobRoleName"`
	JobLevelCode          string                       `json:"jobLevelCode"`
	JobLevelName          string                       `json:"jobLevelName"`
	OrganizationCode      string                       `json:"organizationCode"`
	OrganizationName      *string                      `json:"organizationName,omitempty"`
	PositionType          string                       `json:"positionType"`
	Status                string                       `json:"status"`
	EmploymentType        string                       `json:"employmentType"`
	HeadcountCapacity     float64                      `json:"headcountCapacity"`
	HeadcountInUse        float64                      `json:"headcountInUse"`
	AvailableHeadcount    float64                      `json:"availableHeadcount"`
	GradeLevel            *string                      `json:"gradeLevel,omitempty"`
	CostCenterCode        *string                      `json:"costCenterCode,omitempty"`
	ReportsToPositionCode *string                      `json:"reportsToPositionCode,omitempty"`
	EffectiveDate         time.Time                    `json:"effectiveDate"`
	EndDate               *time.Time                   `json:"endDate,omitempty"`
	IsCurrent             bool                         `json:"isCurrent"`
	IsFuture              bool                         `json:"isFuture"`
	RecordID              uuid.UUID                    `json:"recordId"`
	CreatedAt             time.Time                    `json:"createdAt"`
	UpdatedAt             time.Time                    `json:"updatedAt"`
	CurrentAssignment     *PositionAssignmentResponse  `json:"currentAssignment,omitempty"`
	AssignmentHistory     []PositionAssignmentResponse `json:"assignmentHistory,omitempty"`
}

type PositionAssignmentResponse struct {
	AssignmentID     uuid.UUID  `json:"assignmentId"`
	PositionCode     string     `json:"positionCode"`
	PositionRecordID uuid.UUID  `json:"positionRecordId"`
	EmployeeID       uuid.UUID  `json:"employeeId"`
	EmployeeName     string     `json:"employeeName"`
	EmployeeNumber   *string    `json:"employeeNumber,omitempty"`
	AssignmentType   string     `json:"assignmentType"`
	AssignmentStatus string     `json:"assignmentStatus"`
	FTE              float64    `json:"fte"`
	EffectiveDate    time.Time  `json:"effectiveDate"`
	EndDate          *time.Time `json:"endDate,omitempty"`
	ActingUntil      *time.Time `json:"actingUntil,omitempty"`
	AutoRevert       bool       `json:"autoRevert"`
	ReminderSentAt   *time.Time `json:"reminderSentAt,omitempty"`
	IsCurrent        bool       `json:"isCurrent"`
	Notes            *string    `json:"notes,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// Position 数据实体，用于数据库映射
type Position struct {
	RecordID             uuid.UUID      `db:"record_id"`
	TenantID             uuid.UUID      `db:"tenant_id"`
	Code                 string         `db:"code"`
	Title                string         `db:"title"`
	JobProfileCode       sql.NullString `db:"job_profile_code"`
	JobProfileName       sql.NullString `db:"job_profile_name"`
	JobFamilyGroupCode   string         `db:"job_family_group_code"`
	JobFamilyGroupName   string         `db:"job_family_group_name"`
	JobFamilyGroupRecord uuid.UUID      `db:"job_family_group_record_id"`
	JobFamilyCode        string         `db:"job_family_code"`
	JobFamilyName        string         `db:"job_family_name"`
	JobFamilyRecord      uuid.UUID      `db:"job_family_record_id"`
	JobRoleCode          string         `db:"job_role_code"`
	JobRoleName          string         `db:"job_role_name"`
	JobRoleRecord        uuid.UUID      `db:"job_role_record_id"`
	JobLevelCode         string         `db:"job_level_code"`
	JobLevelName         string         `db:"job_level_name"`
	JobLevelRecord       uuid.UUID      `db:"job_level_record_id"`
	OrganizationCode     string         `db:"organization_code"`
	OrganizationName     sql.NullString `db:"organization_name"`
	PositionType         string         `db:"position_type"`
	Status               string         `db:"status"`
	EmploymentType       string         `db:"employment_type"`
	HeadcountCapacity    float64        `db:"headcount_capacity"`
	HeadcountInUse       float64        `db:"headcount_in_use"`
	GradeLevel           sql.NullString `db:"grade_level"`
	CostCenterCode       sql.NullString `db:"cost_center_code"`
	ReportsToPosition    sql.NullString `db:"reports_to_position_code"`
	Profile              []byte         `db:"profile"`
	EffectiveDate        time.Time      `db:"effective_date"`
	EndDate              sql.NullTime   `db:"end_date"`
	IsCurrent            bool           `db:"is_current"`
	CreatedAt            time.Time      `db:"created_at"`
	UpdatedAt            time.Time      `db:"updated_at"`
	DeletedAt            sql.NullTime   `db:"deleted_at"`
	OperationType        string         `db:"operation_type"`
	OperatedByID         uuid.UUID      `db:"operated_by_id"`
	OperatedByName       string         `db:"operated_by_name"`
	OperationReason      sql.NullString `db:"operation_reason"`
}

// PositionAssignment 数据实体
type PositionAssignment struct {
	AssignmentID     uuid.UUID      `db:"assignment_id"`
	TenantID         uuid.UUID      `db:"tenant_id"`
	PositionCode     string         `db:"position_code"`
	PositionRecordID uuid.UUID      `db:"position_record_id"`
	EmployeeID       uuid.UUID      `db:"employee_id"`
	EmployeeName     string         `db:"employee_name"`
	EmployeeNumber   sql.NullString `db:"employee_number"`
	AssignmentType   string         `db:"assignment_type"`
	AssignmentStatus string         `db:"assignment_status"`
	FTE              float64        `db:"fte"`
	EffectiveDate    time.Time      `db:"effective_date"`
	EndDate          sql.NullTime   `db:"end_date"`
	ActingUntil      sql.NullTime   `db:"acting_until"`
	AutoRevert       bool           `db:"auto_revert"`
	ReminderSentAt   sql.NullTime   `db:"reminder_sent_at"`
	IsCurrent        bool           `db:"is_current"`
	Notes            sql.NullString `db:"notes"`
	CreatedAt        time.Time      `db:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at"`
}

// PositionTimelineEntry 用于构建时间线或版本列表
type PositionTimelineEntry struct {
	RecordID      uuid.UUID    `db:"record_id"`
	EffectiveDate time.Time    `db:"effective_date"`
	EndDate       sql.NullTime `db:"end_date"`
	IsCurrent     bool         `db:"is_current"`
	Status        string       `db:"status"`
}

// Job catalog 实体定义
type JobFamilyGroup struct {
	RecordID      uuid.UUID      `db:"record_id"`
	TenantID      uuid.UUID      `db:"tenant_id"`
	Code          string         `db:"family_group_code"`
	Name          string         `db:"name"`
	Description   sql.NullString `db:"description"`
	Status        string         `db:"status"`
	EffectiveDate time.Time      `db:"effective_date"`
	EndDate       sql.NullTime   `db:"end_date"`
	IsCurrent     bool           `db:"is_current"`
}

type JobFamily struct {
	RecordID        uuid.UUID      `db:"record_id"`
	TenantID        uuid.UUID      `db:"tenant_id"`
	Code            string         `db:"family_code"`
	FamilyGroupCode string         `db:"family_group_code"`
	ParentRecord    uuid.UUID      `db:"parent_record_id"`
	Name            string         `db:"name"`
	Description     sql.NullString `db:"description"`
	Status          string         `db:"status"`
	EffectiveDate   time.Time      `db:"effective_date"`
	EndDate         sql.NullTime   `db:"end_date"`
	IsCurrent       bool           `db:"is_current"`
}

type JobRole struct {
	RecordID      uuid.UUID      `db:"record_id"`
	TenantID      uuid.UUID      `db:"tenant_id"`
	Code          string         `db:"role_code"`
	FamilyCode    string         `db:"family_code"`
	ParentRecord  uuid.UUID      `db:"parent_record_id"`
	Name          string         `db:"name"`
	Description   sql.NullString `db:"description"`
	Competency    []byte         `db:"competency_model"`
	Status        string         `db:"status"`
	EffectiveDate time.Time      `db:"effective_date"`
	EndDate       sql.NullTime   `db:"end_date"`
	IsCurrent     bool           `db:"is_current"`
}

type JobLevel struct {
	RecordID      uuid.UUID      `db:"record_id"`
	TenantID      uuid.UUID      `db:"tenant_id"`
	Code          string         `db:"level_code"`
	RoleCode      string         `db:"role_code"`
	ParentRecord  uuid.UUID      `db:"parent_record_id"`
	LevelRank     string         `db:"level_rank"`
	Name          string         `db:"name"`
	Description   sql.NullString `db:"description"`
	SalaryBand    []byte         `db:"salary_band"`
	Status        string         `db:"status"`
	EffectiveDate time.Time      `db:"effective_date"`
	EndDate       sql.NullTime   `db:"end_date"`
	IsCurrent     bool           `db:"is_current"`
}
