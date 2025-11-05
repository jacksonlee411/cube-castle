package validator

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
)

// jobCatalogRepository 定义职位验证所需的 Job Catalog 查询接口。
type jobCatalogRepository interface {
	GetCurrentFamilyGroup(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamilyGroup, error)
	GetCurrentJobFamily(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobFamily, error)
	GetCurrentJobRole(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobRole, error)
	GetCurrentJobLevel(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.JobLevel, error)
}

// positionRepository 定义职位验证所需的职位查询接口。
type positionRepository interface {
	GetCurrentPosition(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string) (*types.Position, error)
}

// positionAssignmentRepository 定义任职验证所需的仓储接口。
type positionAssignmentRepository interface {
	SumActiveFTE(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, positionCode string) (float64, error)
	GetByID(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, assignmentID uuid.UUID) (*types.PositionAssignment, error)
}

// positionAssignmentValidationService 实现职位与任职验证器接口。
type positionAssignmentValidationService struct {
	orgRepo        organizationRepository
	jobCatalogRepo jobCatalogRepository
	positionRepo   positionRepository
	assignmentRepo positionAssignmentRepository
	logger         pkglogger.Logger
}

// NewPositionAssignmentValidationService 构建职位与任职业务规则验证器。
func NewPositionAssignmentValidationService(
	orgRepo organizationRepository,
	jobCatalogRepo jobCatalogRepository,
	positionRepo positionRepository,
	assignmentRepo positionAssignmentRepository,
	baseLogger pkglogger.Logger,
) (PositionValidationService, AssignmentValidationService) {
	logger := baseLogger
	if logger == nil {
		logger = pkglogger.NewNoopLogger()
	}

	service := &positionAssignmentValidationService{
		orgRepo:        orgRepo,
		jobCatalogRepo: jobCatalogRepo,
		positionRepo:   positionRepo,
		assignmentRepo: assignmentRepo,
		logger: logger.WithFields(pkglogger.Fields{
			"component": "validator",
			"module":    "position-assignment",
		}),
	}
	return service, service
}

// ValidateCreatePosition 校验创建职位请求。
func (s *positionAssignmentValidationService) ValidateCreatePosition(ctx context.Context, tenantID uuid.UUID, req *types.PositionRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	subject := &positionCreateSubject{
		TenantID: tenantID,
		Request:  req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation": "CreatePosition",
	}))
	s.registerPositionCreateRules(chain)

	result := chain.Execute(ctx, subject)
	s.mergeJobCatalogContext(ctx, result, tenantID, req)
	return result
}

// ValidateReplacePosition 校验替换职位请求。
func (s *positionAssignmentValidationService) ValidateReplacePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	subject := &positionUpdateSubject{
		TenantID: tenantID,
		Code:     strings.TrimSpace(code),
		Request:  req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":     "ReplacePosition",
		"positionCode":  subject.Code,
		"tenantId":      tenantID.String(),
		"targetVersion": "CURRENT",
	}))
	s.registerPositionCreateRules(chain)

	result := chain.Execute(ctx, subject)
	s.mergeJobCatalogContext(ctx, result, tenantID, req)
	return result
}

// ValidateCreateVersion 校验新增职位版本请求。
func (s *positionAssignmentValidationService) ValidateCreateVersion(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionVersionRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	subject := &positionVersionSubject{
		TenantID: tenantID,
		Code:     strings.TrimSpace(code),
		Request:  req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":    "CreatePositionVersion",
		"positionCode": subject.Code,
		"tenantId":     tenantID.String(),
	}))

	chain.Register(&Rule{
		ID:           "POS-ORG",
		Priority:     10,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newPosOrgRule(),
	})

	result := chain.Execute(ctx, subject)
	return result
}

// ValidateFillPosition 校验填充职位请求（创建任职）。
func (s *positionAssignmentValidationService) ValidateFillPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.FillPositionRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	position, organization, currentFTE, err := s.loadPositionContext(ctx, tenantID, code)
	if err != nil {
		s.logger.WithFields(pkglogger.Fields{"error": err}).Error("load position context failed for FillPosition")
		return NewValidationResult()
	}
	if position == nil || organization == nil {
		return NewValidationResult()
	}

	requestedFTE := resolveRequestedFTE(req.FTE)

	subject := &positionFillSubject{
		TenantID:     tenantID,
		Position:     position,
		Organization: organization,
		CurrentFTE:   currentFTE,
		RequestedFTE: requestedFTE,
		Request:      req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":    "FillPosition",
		"positionCode": position.Code,
		"tenantId":     tenantID.String(),
	}))
	s.registerAssignmentCreationRules(chain)

	return chain.Execute(ctx, subject)
}

// ValidateVacatePosition 校验清空职位请求。
func (s *positionAssignmentValidationService) ValidateVacatePosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.VacatePositionRequest) *ValidationResult {
	return NewValidationResult()
}

// ValidateTransferPosition 校验职位转移请求。
func (s *positionAssignmentValidationService) ValidateTransferPosition(ctx context.Context, tenantID uuid.UUID, code string, req *types.TransferPositionRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	targetOrg := strings.TrimSpace(req.TargetOrganizationCode)
	if targetOrg == "" {
		return NewValidationResult()
	}

	subject := &positionTransferSubject{
		TenantID: tenantID,
		Code:     strings.TrimSpace(code),
		Target:   targetOrg,
		Request:  req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":       "TransferPosition",
		"positionCode":    subject.Code,
		"targetOrg":       targetOrg,
		"tenantId":        tenantID.String(),
		"requestedAt":     req.EffectiveDate,
		"crossDomainRule": true,
	}))
	chain.Register(&Rule{
		ID:           "POS-ORG",
		Priority:     10,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newPosOrgRule(),
	})
	return chain.Execute(ctx, subject)
}

// ValidateApplyEvent 校验职位事件请求。
func (s *positionAssignmentValidationService) ValidateApplyEvent(ctx context.Context, tenantID uuid.UUID, code string, req *types.PositionEventRequest) *ValidationResult {
	return NewValidationResult()
}

// ValidateCreateAssignment 校验创建任职请求（Position Assignment API）。
func (s *positionAssignmentValidationService) ValidateCreateAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, req *types.CreateAssignmentRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	position, organization, currentFTE, err := s.loadPositionContext(ctx, tenantID, positionCode)
	if err != nil {
		s.logger.WithFields(pkglogger.Fields{"error": err}).Error("load position context failed for CreateAssignment")
		return NewValidationResult()
	}
	if position == nil || organization == nil {
		return NewValidationResult()
	}

	requestedFTE := resolveRequestedFTE(req.FTE)

	subject := &assignmentCreateSubject{
		TenantID:     tenantID,
		Position:     position,
		Organization: organization,
		CurrentFTE:   currentFTE,
		RequestedFTE: requestedFTE,
		Request:      req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":    "CreateAssignment",
		"positionCode": position.Code,
		"tenantId":     tenantID.String(),
	}))
	s.registerAssignmentCreationRules(chain)

	return chain.Execute(ctx, subject)
}

// ValidateUpdateAssignment 校验更新任职请求。
func (s *positionAssignmentValidationService) ValidateUpdateAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID uuid.UUID, req *types.UpdateAssignmentRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	position, organization, currentFTE, err := s.loadPositionContext(ctx, tenantID, positionCode)
	if err != nil {
		s.logger.WithFields(pkglogger.Fields{"error": err}).Error("load position context failed for UpdateAssignment")
		return NewValidationResult()
	}
	if position == nil || organization == nil {
		return NewValidationResult()
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, nil, tenantID, assignmentID)
	if err != nil {
		s.logger.WithFields(pkglogger.Fields{"error": err}).Error("load assignment failed for UpdateAssignment")
		return NewValidationResult()
	}
	if assignment == nil {
		return NewValidationResult()
	}

	requestedFTE := assignment.FTE
	if req.FTE != nil {
		requestedFTE = *req.FTE
	}

	subject := &assignmentUpdateSubject{
		TenantID:         tenantID,
		Position:         position,
		Organization:     organization,
		Assignment:       assignment,
		CurrentFTE:       currentFTE,
		RequestedFTE:     requestedFTE,
		Request:          req,
		AssignmentID:     assignmentID,
		OriginalFTE:      assignment.FTE,
		AssignmentStatus: strings.ToUpper(strings.TrimSpace(assignment.AssignmentStatus)),
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":      "UpdateAssignment",
		"positionCode":   position.Code,
		"assignmentId":   assignmentID.String(),
		"tenantId":       tenantID.String(),
		"currentStatus":  assignment.AssignmentStatus,
		"requestedFTE":   requestedFTE,
		"existingFTE":    assignment.FTE,
		"existingStatus": assignment.AssignmentStatus,
	}))
	s.registerAssignmentUpdateRules(chain)

	return chain.Execute(ctx, subject)
}

// ValidateCloseAssignment 校验关闭任职请求。
func (s *positionAssignmentValidationService) ValidateCloseAssignment(ctx context.Context, tenantID uuid.UUID, positionCode string, assignmentID uuid.UUID, req *types.CloseAssignmentRequest) *ValidationResult {
	if req == nil {
		return NewValidationResult()
	}

	assignment, err := s.assignmentRepo.GetByID(ctx, nil, tenantID, assignmentID)
	if err != nil {
		s.logger.WithFields(pkglogger.Fields{"error": err}).Error("load assignment failed for CloseAssignment")
		return NewValidationResult()
	}
	if assignment == nil {
		return NewValidationResult()
	}

	subject := &assignmentCloseSubject{
		TenantID:   tenantID,
		Assignment: assignment,
		Request:    req,
	}

	chain := NewValidationChain(s.logger, WithBaseContext(map[string]interface{}{
		"operation":    "CloseAssignment",
		"positionCode": strings.TrimSpace(positionCode),
		"assignmentId": assignmentID.String(),
		"tenantId":     tenantID.String(),
	}))

	chain.Register(&Rule{
		ID:           "ASSIGN-STATE",
		Priority:     10,
		Severity:     SeverityCritical,
		ShortCircuit: true,
		Handler:      s.newAssignStateRule(),
	})

	return chain.Execute(ctx, subject)
}

// ===== 规则注册 & Subject 定义 =====

func (s *positionAssignmentValidationService) registerPositionCreateRules(chain *ValidationChain) {
	_ = chain.Register(&Rule{
		ID:           "POS-ORG",
		Priority:     10,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newPosOrgRule(),
	})

	_ = chain.Register(&Rule{
		ID:       "POS-JC-LINK",
		Priority: 20,
		Severity: SeverityMedium,
		Handler:  s.newPosJobCatalogRule(),
	})
}

func (s *positionAssignmentValidationService) registerAssignmentCreationRules(chain *ValidationChain) {
	_ = chain.Register(&Rule{
		ID:           "ASSIGN-FTE",
		Priority:     5,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newAssignFTERule(),
	})

	_ = chain.Register(&Rule{
		ID:           "ASSIGN-STATE",
		Priority:     8,
		Severity:     SeverityCritical,
		ShortCircuit: true,
		Handler:      s.newAssignStateRule(),
	})

	_ = chain.Register(&Rule{
		ID:           "CROSS-ACTIVE",
		Priority:     10,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newCrossActiveRule(),
	})

	_ = chain.Register(&Rule{
		ID:           "POS-HEADCOUNT",
		Priority:     20,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newPosHeadcountRule(),
	})
}

func (s *positionAssignmentValidationService) registerAssignmentUpdateRules(chain *ValidationChain) {
	_ = chain.Register(&Rule{
		ID:           "ASSIGN-FTE",
		Priority:     5,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newAssignFTERule(),
	})

	_ = chain.Register(&Rule{
		ID:           "ASSIGN-STATE",
		Priority:     8,
		Severity:     SeverityCritical,
		ShortCircuit: true,
		Handler:      s.newAssignStateRule(),
	})

	_ = chain.Register(&Rule{
		ID:           "POS-HEADCOUNT",
		Priority:     15,
		Severity:     SeverityHigh,
		ShortCircuit: true,
		Handler:      s.newPosHeadcountRule(),
	})
}

type positionCreateSubject struct {
	TenantID uuid.UUID
	Request  *types.PositionRequest
}

type positionUpdateSubject struct {
	TenantID uuid.UUID
	Code     string
	Request  *types.PositionRequest
}

type positionVersionSubject struct {
	TenantID uuid.UUID
	Code     string
	Request  *types.PositionVersionRequest
}

type positionFillSubject struct {
	TenantID     uuid.UUID
	Position     *types.Position
	Organization *types.Organization
	CurrentFTE   float64
	RequestedFTE float64
	Request      *types.FillPositionRequest
}

type positionTransferSubject struct {
	TenantID uuid.UUID
	Code     string
	Target   string
	Request  *types.TransferPositionRequest
}

type assignmentCreateSubject struct {
	TenantID     uuid.UUID
	Position     *types.Position
	Organization *types.Organization
	CurrentFTE   float64
	RequestedFTE float64
	Request      *types.CreateAssignmentRequest
}

type assignmentUpdateSubject struct {
	TenantID         uuid.UUID
	Position         *types.Position
	Organization     *types.Organization
	Assignment       *types.PositionAssignment
	CurrentFTE       float64
	RequestedFTE     float64
	Request          *types.UpdateAssignmentRequest
	AssignmentID     uuid.UUID
	OriginalFTE      float64
	AssignmentStatus string
}

type assignmentCloseSubject struct {
	TenantID   uuid.UUID
	Assignment *types.PositionAssignment
	Request    *types.CloseAssignmentRequest
}

// ===== Rule Handlers =====

func (s *positionAssignmentValidationService) newPosOrgRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		tenantID, orgCode := s.resolveOrgContext(subject)
		if tenantID == uuid.Nil || orgCode == "" {
			return nil, nil
		}

		org, err := s.orgRepo.GetByCode(ctx, tenantID, orgCode)
		if err != nil {
			return nil, fmt.Errorf("pos-org: fetch organization %s failed: %w", orgCode, err)
		}
		if org == nil {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "POS_ORG_INACTIVE",
					Message:  fmt.Sprintf("Organization %s does not exist or is inactive", orgCode),
					Field:    "organizationCode",
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId":           "POS-ORG",
						"organizationCode": orgCode,
						"status":           "UNKNOWN",
					},
				}},
			}, nil
		}
		status := strings.ToUpper(strings.TrimSpace(org.Status))
		if status != string(types.OrganizationStatusActive) {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "POS_ORG_INACTIVE",
					Message:  fmt.Sprintf("Organization %s status %s is not ACTIVE", orgCode, status),
					Field:    "organizationCode",
					Value:    status,
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId":           "POS-ORG",
						"organizationCode": orgCode,
						"status":           status,
					},
				}},
			}, nil
		}

		return &RuleOutcome{
			Context: map[string]interface{}{
				"organizationCode": orgCode,
				"status":           status,
			},
		}, nil
	}
}

func (s *positionAssignmentValidationService) newPosJobCatalogRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		req := s.extractPositionRequest(subject)
		if req == nil {
			return nil, nil
		}

		tenantID := s.extractTenant(subject)
		if tenantID == uuid.Nil {
			return nil, nil
		}

		group, err := s.jobCatalogRepo.GetCurrentFamilyGroup(ctx, nil, tenantID, strings.TrimSpace(req.JobFamilyGroupCode))
		if err != nil {
			return nil, fmt.Errorf("pos-jc-link: fetch job family group failed: %w", err)
		}
		if group == nil || !strings.EqualFold(group.Status, "ACTIVE") {
			return jobCatalogViolation("JobFamilyGroup", req.JobFamilyGroupCode), nil
		}

		family, err := s.jobCatalogRepo.GetCurrentJobFamily(ctx, nil, tenantID, strings.TrimSpace(req.JobFamilyCode))
		if err != nil {
			return nil, fmt.Errorf("pos-jc-link: fetch job family failed: %w", err)
		}
		if family == nil || !strings.EqualFold(family.Status, "ACTIVE") {
			return jobCatalogViolation("JobFamily", req.JobFamilyCode), nil
		}

		role, err := s.jobCatalogRepo.GetCurrentJobRole(ctx, nil, tenantID, strings.TrimSpace(req.JobRoleCode))
		if err != nil {
			return nil, fmt.Errorf("pos-jc-link: fetch job role failed: %w", err)
		}
		if role == nil || !strings.EqualFold(role.Status, "ACTIVE") {
			return jobCatalogViolation("JobRole", req.JobRoleCode), nil
		}

		level, err := s.jobCatalogRepo.GetCurrentJobLevel(ctx, nil, tenantID, strings.TrimSpace(req.JobLevelCode))
		if err != nil {
			return nil, fmt.Errorf("pos-jc-link: fetch job level failed: %w", err)
		}
		if level == nil || !strings.EqualFold(level.Status, "ACTIVE") {
			return jobCatalogViolation("JobLevel", req.JobLevelCode), nil
		}

		return &RuleOutcome{
			Context: map[string]interface{}{
				"jobFamilyGroup": req.JobFamilyGroupCode,
				"jobFamily":      req.JobFamilyCode,
				"jobRole":        req.JobRoleCode,
				"jobLevel":       req.JobLevelCode,
			},
		}, nil
	}
}

func (s *positionAssignmentValidationService) newAssignFTERule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		requested := s.extractRequestedFTE(subject)
		if requested <= 0 || requested > 1.0 {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "ASSIGN_FTE_LIMIT",
					Message:  fmt.Sprintf("Assignment FTE %.2f must be between 0 and 1", requested),
					Field:    "fte",
					Value:    requested,
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId":        "ASSIGN-FTE",
						"requestedFTE":  requested,
						"allowedRange":  "[0,1]",
						"shortCircuit":  true,
						"operationType": s.extractOperation(subject),
					},
				}},
			}, nil
		}
		return nil, nil
	}
}

func (s *positionAssignmentValidationService) newAssignStateRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		operation := s.extractOperation(subject)
		switch sub := subject.(type) {
		case *positionFillSubject:
			if isInactivePositionStatus(sub.Position.Status) {
				return assignStateViolation(sub.Position.Status, operation), nil
			}
		case *assignmentCreateSubject:
			if isInactivePositionStatus(sub.Position.Status) {
				return assignStateViolation(sub.Position.Status, operation), nil
			}
		case *assignmentUpdateSubject:
			if strings.EqualFold(sub.AssignmentStatus, "ENDED") {
				return assignStateViolation(sub.AssignmentStatus, operation), nil
			}
		case *assignmentCloseSubject:
			if !strings.EqualFold(sub.Assignment.AssignmentStatus, "ACTIVE") {
				return assignStateViolation(sub.Assignment.AssignmentStatus, operation), nil
			}
		default:
			return nil, nil
		}
		return nil, nil
	}
}

func (s *positionAssignmentValidationService) newCrossActiveRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		position := s.extractPosition(subject)
		org := s.extractOrganization(subject)
		if position == nil || org == nil {
			return nil, nil
		}

		status := strings.ToUpper(strings.TrimSpace(position.Status))
		if status == "INACTIVE" || status == "DELETED" {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "CROSS_ACTIVATION_CONFLICT",
					Message:  fmt.Sprintf("Position %s status %s does not allow assignment operations", position.Code, status),
					Field:    "status",
					Value:    status,
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId":         "CROSS-ACTIVE",
						"positionCode":   position.Code,
						"positionStatus": status,
					},
				}},
			}, nil
		}

		orgStatus := strings.ToUpper(strings.TrimSpace(org.Status))
		if orgStatus != string(types.OrganizationStatusActive) {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "CROSS_ACTIVATION_CONFLICT",
					Message:  fmt.Sprintf("Organization %s status %s does not allow assignment operations", org.Code, orgStatus),
					Field:    "organizationCode",
					Value:    orgStatus,
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId":             "CROSS-ACTIVE",
						"organizationCode":   org.Code,
						"organizationStatus": orgStatus,
					},
				}},
			}, nil
		}

		return nil, nil
	}
}

func (s *positionAssignmentValidationService) newPosHeadcountRule() RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		position := s.extractPosition(subject)
		if position == nil {
			return nil, nil
		}

		current := s.extractCurrentFTE(subject)
		requested := s.extractRequestedFTE(subject)
		original := s.extractOriginalFTE(subject)

		projected := current + requested
		if original > 0 {
			projected = current - original + requested
		}

		limit := position.HeadcountCapacity
		if projected > limit+1e-9 {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "POS_HEADCOUNT_EXCEEDED",
					Message:  fmt.Sprintf("Projected headcount %.2f exceeds capacity %.2f", projected, limit),
					Field:    "fte",
					Value:    requested,
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId":         "POS-HEADCOUNT",
						"positionCode":   position.Code,
						"headcountLimit": limit,
						"currentFTE":     current,
						"requestedFTE":   requested,
						"projectedFTE":   projected,
					},
				}},
			}, nil
		}

		return &RuleOutcome{
			Context: map[string]interface{}{
				"currentFTE":     current,
				"requestedFTE":   requested,
				"projectedFTE":   projected,
				"headcountLimit": limit,
			},
		}, nil
	}
}

// ===== Helper Methods =====

func (s *positionAssignmentValidationService) loadPositionContext(ctx context.Context, tenantID uuid.UUID, code string) (*types.Position, *types.Organization, float64, error) {
	position, err := s.positionRepo.GetCurrentPosition(ctx, nil, tenantID, strings.TrimSpace(code))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("load current position failed: %w", err)
	}
	if position == nil {
		return nil, nil, 0, nil
	}

	org, err := s.orgRepo.GetByCode(ctx, tenantID, strings.TrimSpace(position.OrganizationCode))
	if err != nil {
		return nil, nil, 0, fmt.Errorf("load position organization failed: %w", err)
	}
	if org == nil {
		return position, nil, 0, nil
	}

	currentFTE, err := s.assignmentRepo.SumActiveFTE(ctx, nil, tenantID, position.Code)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("calculate active assignment FTE failed: %w", err)
	}

	return position, org, currentFTE, nil
}

func (s *positionAssignmentValidationService) mergeJobCatalogContext(ctx context.Context, result *ValidationResult, tenantID uuid.UUID, req *types.PositionRequest) {
	if result == nil || result.Context == nil || req == nil {
		return
	}

	result.Context["jobCatalog"] = map[string]string{
		"group":  req.JobFamilyGroupCode,
		"family": req.JobFamilyCode,
		"role":   req.JobRoleCode,
		"level":  req.JobLevelCode,
	}
	result.Context["tenantId"] = tenantID.String()
}

func (s *positionAssignmentValidationService) resolveOrgContext(subject interface{}) (uuid.UUID, string) {
	switch sub := subject.(type) {
	case *positionCreateSubject:
		return sub.TenantID, strings.TrimSpace(sub.Request.OrganizationCode)
	case *positionUpdateSubject:
		return sub.TenantID, strings.TrimSpace(sub.Request.OrganizationCode)
	case *positionFillSubject:
		if sub.Position != nil {
			return sub.TenantID, strings.TrimSpace(sub.Position.OrganizationCode)
		}
	case *assignmentCreateSubject:
		if sub.Position != nil {
			return sub.TenantID, strings.TrimSpace(sub.Position.OrganizationCode)
		}
	case *positionTransferSubject:
		return sub.TenantID, strings.TrimSpace(sub.Target)
	case *assignmentUpdateSubject:
		if sub.Position != nil {
			return sub.TenantID, strings.TrimSpace(sub.Position.OrganizationCode)
		}
	}
	return uuid.Nil, ""
}

func (s *positionAssignmentValidationService) extractPositionRequest(subject interface{}) *types.PositionRequest {
	switch sub := subject.(type) {
	case *positionCreateSubject:
		return sub.Request
	case *positionUpdateSubject:
		return sub.Request
	default:
		return nil
	}
}

func (s *positionAssignmentValidationService) extractTenant(subject interface{}) uuid.UUID {
	switch sub := subject.(type) {
	case *positionCreateSubject:
		return sub.TenantID
	case *positionUpdateSubject:
		return sub.TenantID
	case *positionVersionSubject:
		return sub.TenantID
	case *positionFillSubject:
		return sub.TenantID
	case *assignmentCreateSubject:
		return sub.TenantID
	case *assignmentUpdateSubject:
		return sub.TenantID
	case *assignmentCloseSubject:
		return sub.TenantID
	default:
		return uuid.Nil
	}
}

func (s *positionAssignmentValidationService) extractOperation(subject interface{}) string {
	switch subject.(type) {
	case *positionFillSubject:
		return "FillPosition"
	case *assignmentCreateSubject:
		return "CreateAssignment"
	case *assignmentUpdateSubject:
		return "UpdateAssignment"
	case *assignmentCloseSubject:
		return "CloseAssignment"
	default:
		return "Unknown"
	}
}

func (s *positionAssignmentValidationService) extractRequestedFTE(subject interface{}) float64 {
	switch sub := subject.(type) {
	case *positionFillSubject:
		return sub.RequestedFTE
	case *assignmentCreateSubject:
		return sub.RequestedFTE
	case *assignmentUpdateSubject:
		return sub.RequestedFTE
	default:
		return 1.0
	}
}

func (s *positionAssignmentValidationService) extractOriginalFTE(subject interface{}) float64 {
	if sub, ok := subject.(*assignmentUpdateSubject); ok {
		return sub.OriginalFTE
	}
	return 0
}

func (s *positionAssignmentValidationService) extractCurrentFTE(subject interface{}) float64 {
	switch sub := subject.(type) {
	case *positionFillSubject:
		return sub.CurrentFTE
	case *assignmentCreateSubject:
		return sub.CurrentFTE
	case *assignmentUpdateSubject:
		return sub.CurrentFTE
	default:
		return 0
	}
}

func (s *positionAssignmentValidationService) extractPosition(subject interface{}) *types.Position {
	switch sub := subject.(type) {
	case *positionFillSubject:
		return sub.Position
	case *assignmentCreateSubject:
		return sub.Position
	case *assignmentUpdateSubject:
		return sub.Position
	default:
		return nil
	}
}

func (s *positionAssignmentValidationService) extractOrganization(subject interface{}) *types.Organization {
	switch sub := subject.(type) {
	case *positionFillSubject:
		return sub.Organization
	case *assignmentCreateSubject:
		return sub.Organization
	case *assignmentUpdateSubject:
		return sub.Organization
	default:
		return nil
	}
}

func jobCatalogViolation(entity, code string) *RuleOutcome {
	return &RuleOutcome{
		Errors: []ValidationError{{
			Code:     "JOB_CATALOG_NOT_FOUND",
			Message:  fmt.Sprintf("%s %s is inactive or missing", entity, strings.TrimSpace(code)),
			Field:    strings.ToLower(entity) + "Code",
			Value:    strings.TrimSpace(code),
			Severity: string(SeverityMedium),
			Context: map[string]interface{}{
				"ruleId":        "POS-JC-LINK",
				"catalogEntity": entity,
				"referenceCode": strings.TrimSpace(code),
			},
		}},
	}
}

func assignStateViolation(status string, operation string) *RuleOutcome {
	state := strings.ToUpper(strings.TrimSpace(status))
	return &RuleOutcome{
		Errors: []ValidationError{{
			Code:     "ASSIGN_INVALID_STATE",
			Message:  fmt.Sprintf("Assignment state %s does not allow %s", state, operation),
			Field:    "assignmentStatus",
			Value:    state,
			Severity: string(SeverityCritical),
			Context: map[string]interface{}{
				"ruleId":       "ASSIGN-STATE",
				"currentState": state,
				"operation":    operation,
			},
		}},
	}
}

func resolveRequestedFTE(ftePtr *float64) float64 {
	if ftePtr == nil {
		return 1.0
	}
	return *ftePtr
}

func isInactivePositionStatus(status string) bool {
	value := strings.ToUpper(strings.TrimSpace(status))
	return value == "INACTIVE" || value == "DELETED"
}

// Enforce interface compliance.
var (
	_ PositionValidationService   = (*positionAssignmentValidationService)(nil)
	_ AssignmentValidationService = (*positionAssignmentValidationService)(nil)
)
