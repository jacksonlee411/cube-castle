// internal/service/temporal_types.go  
package service

import (
	"time"

	"github.com/google/uuid"
)

// PositionChangeEvent represents a position change event with context
type PositionChangeEvent struct {
	PositionHistoryID uuid.UUID        `json:"position_history_id"`
	EmployeeID       uuid.UUID        `json:"employee_id"`
	ChangeType       string           `json:"change_type"` // INITIAL_HIRE, PROMOTION, TRANSFER, MANAGER_CHANGE, INFORMATION_UPDATE
	EffectiveDate    time.Time        `json:"effective_date"`
	PreviousPosition *PositionSnapshot `json:"previous_position,omitempty"`
	NewPosition      *PositionSnapshot `json:"new_position"`
	IsRetroactive    bool             `json:"is_retroactive"`
	ChangeReason     string           `json:"change_reason,omitempty"`
}

// TemporalConsistencyReport identifies gaps and overlaps in position timeline
type TemporalConsistencyReport struct {
	TenantID    uuid.UUID         `json:"tenant_id"`
	EmployeeIDs []uuid.UUID       `json:"employee_ids"`
	GeneratedAt time.Time         `json:"generated_at"`
	Gaps        []PositionGap     `json:"gaps"`
	Overlaps    []PositionOverlap `json:"overlaps"`
	Warnings    []string          `json:"warnings"`
	Summary     ConsistencySummary `json:"summary"`
}

// PositionGap represents a gap in employment history
type PositionGap struct {
	EmployeeID       uuid.UUID     `json:"employee_id"`
	GapStart         time.Time     `json:"gap_start"`
	GapEnd           time.Time     `json:"gap_end"`
	GapDuration      time.Duration `json:"gap_duration"`
	PreviousPosition uuid.UUID     `json:"previous_position"`
	NextPosition     uuid.UUID     `json:"next_position"`
	Severity         string        `json:"severity"` // LOW, MEDIUM, HIGH
}

// PositionOverlap represents an overlap in position records
type PositionOverlap struct {
	EmployeeID      uuid.UUID     `json:"employee_id"`
	OverlapStart    time.Time     `json:"overlap_start"`
	OverlapEnd      time.Time     `json:"overlap_end"`
	OverlapDuration time.Duration `json:"overlap_duration"`
	Position1       uuid.UUID     `json:"position1"`
	Position2       uuid.UUID     `json:"position2"`
	Severity        string        `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
}

// ConsistencySummary provides summary statistics for consistency report
type ConsistencySummary struct {
	TotalEmployees       int                    `json:"total_employees"`
	EmployeesWithGaps    int                    `json:"employees_with_gaps"`
	EmployeesWithOverlaps int                   `json:"employees_with_overlaps"`
	TotalGaps            int                    `json:"total_gaps"`
	TotalOverlaps        int                    `json:"total_overlaps"`
	SeverityBreakdown    map[string]int         `json:"severity_breakdown"`
	AverageGapDuration   time.Duration          `json:"average_gap_duration"`
	LongestGap           time.Duration          `json:"longest_gap"`
}

// BatchPositionSnapshotData represents data for batch position creation
type BatchPositionSnapshotData struct {
	EmployeeID          uuid.UUID  `json:"employee_id"`
	PositionTitle       string     `json:"position_title"`
	Department          string     `json:"department"`
	JobLevel            *string    `json:"job_level,omitempty"`
	Location            *string    `json:"location,omitempty"`
	EmploymentType      string     `json:"employment_type"`
	ReportsToEmployeeID *uuid.UUID `json:"reports_to_employee_id,omitempty"`
	EffectiveDate       time.Time  `json:"effective_date"`
	EndDate             *time.Time `json:"end_date,omitempty"`
	ChangeReason        *string    `json:"change_reason,omitempty"`
	IsRetroactive       bool       `json:"is_retroactive"`
	MinSalary           *float64   `json:"min_salary,omitempty"`
	MaxSalary           *float64   `json:"max_salary,omitempty"`
	Currency            *string    `json:"currency,omitempty"`
	CreatedBy           uuid.UUID  `json:"created_by"`
}

// BatchCreateResult represents the result of a batch create operation
type BatchCreateResult struct {
	TotalRequested int          `json:"total_requested"`
	SuccessCount   int          `json:"success_count"`
	FailureCount   int          `json:"failure_count"`
	Successful     []uuid.UUID  `json:"successful"`
	Failed         []BatchError `json:"failed"`
	ExecutionTime  time.Duration `json:"execution_time"`
}

// BatchError represents an error in batch processing
type BatchError struct {
	Index      int       `json:"index"`
	EmployeeID uuid.UUID `json:"employee_id"`
	Error      string    `json:"error"`
	ErrorType  string    `json:"error_type"` // TEMPORAL_CONFLICT, VALIDATION_ERROR, CREATE_FAILED
}

// TemporalQueryRequest represents a generic temporal query request
type TemporalQueryRequest struct {
	TenantID       uuid.UUID             `json:"tenant_id"`
	QueryType      string                `json:"query_type"` // TIMELINE, AS_OF_DATE, CHANGES_IN_PERIOD
	Parameters     map[string]interface{} `json:"parameters"`
	ExecutionMode  string                `json:"execution_mode"` // STANDARD, OPTIMIZED, CACHED
	ResponseFormat string                `json:"response_format"` // FULL, SUMMARY, IDS_ONLY
}

// TemporalQueryResponse represents a generic temporal query response
type TemporalQueryResponse struct {
	QueryID       uuid.UUID              `json:"query_id"`
	QueryType     string                 `json:"query_type"`
	Results       interface{}            `json:"results"`
	Metadata      *QueryExecutionMetrics `json:"metadata"`
	GeneratedAt   time.Time              `json:"generated_at"`
	CacheStatus   string                 `json:"cache_status"` // HIT, MISS, BYPASS
}

// PositionAnalytics represents analytical data about positions
type PositionAnalytics struct {
	TenantID             uuid.UUID                    `json:"tenant_id"`
	AnalysisPeriod       DateRange                    `json:"analysis_period"`
	TotalPositions       int                          `json:"total_positions"`
	ActivePositions      int                          `json:"active_positions"`
	RetroactiveChanges   int                          `json:"retroactive_changes"`
	DepartmentBreakdown  map[string]int              `json:"department_breakdown"`
	JobLevelBreakdown    map[string]int              `json:"job_level_breakdown"`
	ChangeTypeBreakdown  map[string]int              `json:"change_type_breakdown"`
	AveragePositionTenure time.Duration              `json:"average_position_tenure"`
	TopDepartmentChanges []DepartmentChangeStats     `json:"top_department_changes"`
	PositionTrends       []PositionTrendData         `json:"position_trends"`
}

// DepartmentChangeStats represents change statistics for a department
type DepartmentChangeStats struct {
	Department     string  `json:"department"`
	TotalChanges   int     `json:"total_changes"`
	HireCount      int     `json:"hire_count"`
	PromotionCount int     `json:"promotion_count"`
	TransferIn     int     `json:"transfer_in"`
	TransferOut    int     `json:"transfer_out"`
	TurnoverRate   float64 `json:"turnover_rate"`
}

// PositionTrendData represents position trends over time
type PositionTrendData struct {
	Date            time.Time `json:"date"`
	TotalPositions  int       `json:"total_positions"`
	NewHires        int       `json:"new_hires"`
	Promotions      int       `json:"promotions"`
	Transfers       int       `json:"transfers"`
	Terminations    int       `json:"terminations"`
}

// RetroactiveChangeAnalysis represents analysis of retroactive changes
type RetroactiveChangeAnalysis struct {
	TenantID              uuid.UUID                  `json:"tenant_id"`
	AnalysisPeriod        DateRange                  `json:"analysis_period"`
	TotalRetroactiveChanges int                      `json:"total_retroactive_changes"`
	ByTimePeriod          map[string]int             `json:"by_time_period"` // days_ago ranges
	ByChangeType          map[string]int             `json:"by_change_type"`
	ByDepartment          map[string]int             `json:"by_department"`
	AverageRetroactiveDays int                       `json:"average_retroactive_days"`
	MostRetroactiveChanges []RetroactiveChangeDetail `json:"most_retroactive_changes"`
	ComplianceRisks       []ComplianceRisk           `json:"compliance_risks"`
}

// RetroactiveChangeDetail represents detailed information about a retroactive change
type RetroactiveChangeDetail struct {
	PositionHistoryID uuid.UUID `json:"position_history_id"`
	EmployeeID       uuid.UUID `json:"employee_id"`
	ChangeType       string    `json:"change_type"`
	EffectiveDate    time.Time `json:"effective_date"`
	CreatedAt        time.Time `json:"created_at"`
	RetroactiveDays  int       `json:"retroactive_days"`
	Reason           string    `json:"reason"`
	RiskLevel        string    `json:"risk_level"` // LOW, MEDIUM, HIGH, CRITICAL
}

// ComplianceRisk represents a compliance risk identified in retroactive changes
type ComplianceRisk struct {
	RiskType        string    `json:"risk_type"` // AUDIT_TRAIL, PAYROLL_IMPACT, REGULATORY
	Description     string    `json:"description"`
	AffectedRecords int       `json:"affected_records"`
	Severity        string    `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	Recommendation  string    `json:"recommendation"`
	DeadlineDate    *time.Time `json:"deadline_date,omitempty"`
}

// PositionValidationRule represents a rule for validating position data
type PositionValidationRule struct {
	RuleID          string                 `json:"rule_id"`
	RuleName        string                 `json:"rule_name"`
	RuleType        string                 `json:"rule_type"` // TEMPORAL, BUSINESS, COMPLIANCE
	IsActive        bool                   `json:"is_active"`
	Priority        int                    `json:"priority"`
	Parameters      map[string]interface{} `json:"parameters"`
	ErrorMessage    string                 `json:"error_message"`
	CreatedAt       time.Time              `json:"created_at"`
	LastModified    time.Time              `json:"last_modified"`
}

// ValidationResult represents the result of position validation
type ValidationResult struct {
	IsValid         bool                   `json:"is_valid"`
	ValidationScore float64                `json:"validation_score"` // 0.0 to 1.0
	Violations      []ValidationViolation  `json:"violations"`
	Warnings        []ValidationWarning    `json:"warnings"`
	Recommendations []string               `json:"recommendations"`
	ValidatedAt     time.Time              `json:"validated_at"`
}

// ValidationViolation represents a validation rule violation
type ValidationViolation struct {
	RuleID      string    `json:"rule_id"`
	RuleName    string    `json:"rule_name"`
	Severity    string    `json:"severity"` // ERROR, WARNING, INFO
	Message     string    `json:"message"`
	FieldName   string    `json:"field_name,omitempty"`
	FieldValue  interface{} `json:"field_value,omitempty"`
	Suggestion  string    `json:"suggestion,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	WarningType string    `json:"warning_type"`
	Message     string    `json:"message"`
	Impact      string    `json:"impact"` // LOW, MEDIUM, HIGH
	ActionRequired bool   `json:"action_required"`
}

// TemporalQueryCache represents cached query results
type TemporalQueryCache struct {
	CacheKey      string                 `json:"cache_key"`
	QueryHash     string                 `json:"query_hash"`
	Results       interface{}            `json:"results"`
	Metadata      *QueryExecutionMetrics `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	ExpiresAt     time.Time              `json:"expires_at"`
	AccessCount   int                    `json:"access_count"`
	LastAccessedAt time.Time             `json:"last_accessed_at"`
}

// TemporalQueryOptimization represents query optimization suggestions
type TemporalQueryOptimization struct {
	QueryPattern     string                 `json:"query_pattern"`
	OptimizationType string                 `json:"optimization_type"` // INDEX, CACHE, PARTITION
	Description      string                 `json:"description"`
	ExpectedImprovement float64             `json:"expected_improvement"` // percentage
	ImplementationCost   string             `json:"implementation_cost"` // LOW, MEDIUM, HIGH
	Parameters       map[string]interface{} `json:"parameters"`
	EstimatedSavings time.Duration          `json:"estimated_savings"`
}