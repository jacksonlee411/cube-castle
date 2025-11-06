package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/organization/repository"
	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

const (
	jobCatalogTemporalRuleID = "JC-TEMPORAL"
	jobCatalogSequenceRuleID = "JC-SEQUENCE"
	dateLayout               = "2006-01-02"

	errorCodeTemporalConflict      = "JOB_CATALOG_TEMPORAL_CONFLICT"
	errorCodeSequenceMismatch      = "JOB_CATALOG_SEQUENCE_MISMATCH"
	errorCodeSequenceMissingParent = "JOB_CATALOG_SEQUENCE_MISSING_PARENT"
	errorCodeSequenceMissingBase   = "JOB_CATALOG_SEQUENCE_MISSING_BASE"
	errorCodeTimelineUnavailable   = "JOB_CATALOG_TIMELINE_UNAVAILABLE"
)

type jobCatalogTimelineRepository interface {
	ListFamilyGroupTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]repository.JobCatalogTimelineEntry, error)
	ListJobFamilyTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]repository.JobCatalogTimelineEntry, error)
	ListJobRoleTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]repository.JobCatalogTimelineEntry, error)
	ListJobLevelTimeline(ctx context.Context, tenantID uuid.UUID, code string) ([]repository.JobCatalogTimelineEntry, error)
}

type jobCatalogValidationService struct {
	repo   jobCatalogTimelineRepository
	logger pkglogger.Logger
}

// NewJobCatalogValidationService 构建 Job Catalog 业务规则验证器。
func NewJobCatalogValidationService(repo jobCatalogTimelineRepository, baseLogger pkglogger.Logger) JobCatalogValidationService {
	logger := baseLogger
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}

	return &jobCatalogValidationService{
		repo: repo,
		logger: logger.WithFields(pkglogger.Fields{
			"component": "validator",
			"module":    "job-catalog",
		}),
	}
}

func (s *jobCatalogValidationService) ValidateCreateFamilyGroupVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest) *ValidationResult {
	cfg := jobCatalogVersionConfig{
		operation: "CreateJobFamilyGroupVersion",
		entity:    "JOB_FAMILY_GROUP",
		tenantID:  tenantID,
		code:      code,
		request:   req,
		loader:    s.repo.ListFamilyGroupTimeline,
	}
	return s.validateVersion(ctx, cfg)
}

func (s *jobCatalogValidationService) ValidateCreateJobFamilyVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *ValidationResult {
	cfg := jobCatalogVersionConfig{
		operation:      "CreateJobFamilyVersion",
		entity:         "JOB_FAMILY",
		tenantID:       tenantID,
		code:           code,
		request:        req,
		loader:         s.repo.ListJobFamilyTimeline,
		requireParent:  true,
		parentRecordID: &parentRecordID,
	}
	return s.validateVersion(ctx, cfg)
}

func (s *jobCatalogValidationService) ValidateCreateJobRoleVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *ValidationResult {
	cfg := jobCatalogVersionConfig{
		operation:      "CreateJobRoleVersion",
		entity:         "JOB_ROLE",
		tenantID:       tenantID,
		code:           code,
		request:        req,
		loader:         s.repo.ListJobRoleTimeline,
		requireParent:  true,
		parentRecordID: &parentRecordID,
	}
	return s.validateVersion(ctx, cfg)
}

func (s *jobCatalogValidationService) ValidateCreateJobLevelVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.JobCatalogVersionRequest, parentRecordID uuid.UUID) *ValidationResult {
	cfg := jobCatalogVersionConfig{
		operation:      "CreateJobLevelVersion",
		entity:         "JOB_LEVEL",
		tenantID:       tenantID,
		code:           code,
		request:        req,
		loader:         s.repo.ListJobLevelTimeline,
		requireParent:  true,
		parentRecordID: &parentRecordID,
	}
	return s.validateVersion(ctx, cfg)
}

type jobCatalogVersionConfig struct {
	operation      string
	entity         string
	tenantID       uuid.UUID
	code           string
	request        *types.JobCatalogVersionRequest
	loader         func(ctx context.Context, tenantID uuid.UUID, code string) ([]repository.JobCatalogTimelineEntry, error)
	requireParent  bool
	parentRecordID *uuid.UUID
}

func (s *jobCatalogValidationService) validateVersion(ctx context.Context, cfg jobCatalogVersionConfig) *ValidationResult {
	result := NewValidationResult()
	normalizedCode := strings.ToUpper(strings.TrimSpace(cfg.code))

	result.Context["operation"] = cfg.operation
	result.Context["catalogCode"] = normalizedCode
	result.Context["tenantId"] = cfg.tenantID.String()
	if cfg.parentRecordID != nil {
		result.Context["parentRecordId"] = cfg.parentRecordID.String()
	}

	if cfg.request == nil {
		return result
	}

	effectiveDate, err := s.parseEffectiveDate(cfg.request.EffectiveDate)
	if err != nil {
		return s.invalidEffectiveDateResult(cfg.operation, normalizedCode, cfg.request.EffectiveDate)
	}

	timeline, err := cfg.loader(ctx, cfg.tenantID, normalizedCode)
	if err != nil {
		s.logger.WithFields(pkglogger.Fields{
			"operation": cfg.operation,
			"code":      normalizedCode,
			"error":     err,
		}).Error("加载 Job Catalog 时态版本失败")
		return s.timelineFailureResult(cfg.operation, normalizedCode, err)
	}

	subject := &jobCatalogVersionSubject{
		TenantID:       cfg.tenantID,
		Code:           normalizedCode,
		Entity:         cfg.entity,
		EffectiveDate:  effectiveDate,
		Timeline:       timeline,
		ParentRecordID: cfg.parentRecordID,
		RequireParent:  cfg.requireParent,
	}

	chain := s.newJobCatalogChain(cfg.operation, normalizedCode, cfg.tenantID, cfg.parentRecordID)
	if err := chain.Register(&Rule{
		ID:           jobCatalogTemporalRuleID,
		Priority:     10,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newJobCatalogTemporalRule(),
	}); err != nil {
		s.logger.WithFields(pkglogger.Fields{
			"operation": cfg.operation,
			"code":      normalizedCode,
			"error":     err,
		}).Error("注册 JC-TEMPORAL 规则失败")
		return s.timelineFailureResult(cfg.operation, normalizedCode, err)
	}

	if cfg.requireParent {
		if err := chain.Register(&Rule{
			ID:           jobCatalogSequenceRuleID,
			Priority:     20,
			Severity:     SeverityHigh,
			ShortCircuit: true,
			Handler:      s.newJobCatalogSequenceRule(),
		}); err != nil {
			s.logger.WithFields(pkglogger.Fields{
				"operation": cfg.operation,
				"code":      normalizedCode,
				"error":     err,
			}).Error("注册 JC-SEQUENCE 规则失败")
			return s.timelineFailureResult(cfg.operation, normalizedCode, err)
		}
	}

	return chain.Execute(ctx, subject)
}

func (s *jobCatalogValidationService) parseEffectiveDate(raw string) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return time.Time{}, fmt.Errorf("effective date is required")
	}
	return time.Parse(dateLayout, trimmed)
}

func (s *jobCatalogValidationService) newJobCatalogChain(operation, code string, tenantID uuid.UUID, parent *uuid.UUID) *ValidationChain {
	ctx := map[string]interface{}{
		"operation":   operation,
		"catalogCode": code,
		"tenantId":    tenantID.String(),
	}
	if parent != nil {
		ctx["parentRecordId"] = parent.String()
	}
	return NewValidationChain(
		s.logger,
		WithBaseContext(ctx),
		WithOperationLabel(operation),
	)
}

func (s *jobCatalogValidationService) invalidEffectiveDateResult(operation, code, attempted string) *ValidationResult {
	result := NewValidationResult()
	result.Valid = false
	result.Context["operation"] = operation
	result.Context["catalogCode"] = code
	result.Context["executedRules"] = []string{}
	result.Errors = append(result.Errors, ValidationError{
		Code:     "INVALID_EFFECTIVE_DATE",
		Message:  fmt.Sprintf("effectiveDate must follow format %s", dateLayout),
		Field:    "effectiveDate",
		Value:    strings.TrimSpace(attempted),
		Severity: string(SeverityHigh),
		Context: map[string]interface{}{
			"ruleId":        jobCatalogTemporalRuleID,
			"catalogCode":   code,
			"attemptedDate": strings.TrimSpace(attempted),
		},
	})
	return result
}

func (s *jobCatalogValidationService) timelineFailureResult(operation, code string, err error) *ValidationResult {
	result := NewValidationResult()
	result.Valid = false
	result.Context["operation"] = operation
	result.Context["catalogCode"] = code
	result.Context["executedRules"] = []string{}
	result.Errors = append(result.Errors, ValidationError{
		Code:     errorCodeTimelineUnavailable,
		Message:  "无法加载 Job Catalog 时间线以完成校验",
		Severity: string(SeverityCritical),
		Context: map[string]interface{}{
			"ruleId":      jobCatalogTemporalRuleID,
			"catalogCode": code,
			"internal":    true,
			"error":       err.Error(),
		},
	})
	return result
}

func (s *jobCatalogValidationService) newJobCatalogTemporalRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		version, ok := subject.(*jobCatalogVersionSubject)
		if !ok {
			return nil, fmt.Errorf("JC-TEMPORAL rule expects jobCatalogVersionSubject, got %T", subject)
		}

		outcome := &RuleOutcome{
			Context: map[string]interface{}{
				"ruleId":        jobCatalogTemporalRuleID,
				"timelineSize":  len(version.Timeline),
				"catalogCode":   version.Code,
				"entity":        version.Entity,
				"effectiveDate": version.EffectiveDate.Format(dateLayout),
			},
		}

		if len(version.Timeline) == 0 {
			return outcome, nil
		}

		latest := version.Timeline[len(version.Timeline)-1]
		outcome.Context["latestEffective"] = latest.EffectiveDate.Format(dateLayout)
		outcome.Context["latestRecordId"] = latest.RecordID.String()

		if !version.EffectiveDate.After(latest.EffectiveDate) {
			outcome.Errors = append(outcome.Errors, ValidationError{
				Code:     errorCodeTemporalConflict,
				Message:  fmt.Sprintf("effectiveDate %s must be after latest version %s", version.EffectiveDate.Format(dateLayout), latest.EffectiveDate.Format(dateLayout)),
				Field:    "effectiveDate",
				Severity: string(SeverityHigh),
				Context: map[string]interface{}{
					"ruleId":             jobCatalogTemporalRuleID,
					"catalogCode":        version.Code,
					"latestEffective":    latest.EffectiveDate.Format(dateLayout),
					"attemptedEffective": version.EffectiveDate.Format(dateLayout),
					"latestRecordId":     latest.RecordID.String(),
				},
			})
		}

		return outcome, nil
	}
}

func (s *jobCatalogValidationService) newJobCatalogSequenceRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		version, ok := subject.(*jobCatalogVersionSubject)
		if !ok {
			return nil, fmt.Errorf("JC-SEQUENCE rule expects jobCatalogVersionSubject, got %T", subject)
		}

		outcome := &RuleOutcome{
			Context: map[string]interface{}{
				"ruleId":         jobCatalogSequenceRuleID,
				"catalogCode":    version.Code,
				"requiresParent": version.RequireParent,
			},
		}

		if !version.RequireParent {
			return outcome, nil
		}

		if len(version.Timeline) == 0 {
			outcome.Errors = append(outcome.Errors, ValidationError{
				Code:     errorCodeSequenceMissingBase,
				Message:  "cannot create version without existing timeline entries",
				Severity: string(SeverityHigh),
				Context: map[string]interface{}{
					"ruleId":      jobCatalogSequenceRuleID,
					"catalogCode": version.Code,
				},
			})
			return outcome, nil
		}

		expected := version.Timeline[len(version.Timeline)-1].RecordID
		outcome.Context["expectedParentRecordId"] = expected.String()

		if version.ParentRecordID == nil {
			outcome.Errors = append(outcome.Errors, ValidationError{
				Code:     errorCodeSequenceMissingParent,
				Message:  "parentRecordId is required to link version sequence",
				Field:    "parentRecordId",
				Severity: string(SeverityHigh),
				Context: map[string]interface{}{
					"ruleId":                 jobCatalogSequenceRuleID,
					"expectedParentRecordId": expected.String(),
				},
			})
			return outcome, nil
		}

		outcome.Context["providedParentRecordId"] = version.ParentRecordID.String()

		if expected != *version.ParentRecordID {
			outcome.Errors = append(outcome.Errors, ValidationError{
				Code:     errorCodeSequenceMismatch,
				Message:  "parentRecordId does not match latest version record",
				Field:    "parentRecordId",
				Value:    version.ParentRecordID.String(),
				Severity: string(SeverityHigh),
				Context: map[string]interface{}{
					"ruleId":                 jobCatalogSequenceRuleID,
					"expectedParentRecordId": expected.String(),
					"providedParentRecordId": version.ParentRecordID.String(),
				},
			})
		}

		return outcome, nil
	}
}

type jobCatalogVersionSubject struct {
	TenantID       uuid.UUID
	Code           string
	Entity         string
	EffectiveDate  time.Time
	Timeline       []repository.JobCatalogTimelineEntry
	ParentRecordID *uuid.UUID
	RequireParent  bool
}

var _ JobCatalogValidationService = (*jobCatalogValidationService)(nil)
