package validator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cube-castle/internal/types"
	"github.com/google/uuid"
)

const (
	maxOrganizationDepth  = 17
	depthWarningThreshold = 15
)

type organizationCreateSubject struct {
	TenantID uuid.UUID
	Request  *types.CreateOrganizationRequest
}

type organizationUpdateSubject struct {
	TenantID uuid.UUID
	Code     string
	Request  *types.UpdateOrganizationRequest
	Existing *types.Organization
}

func (v *BusinessRuleValidator) buildOrganizationCreateChain(req *types.CreateOrganizationRequest) *ValidationChain {
	chain := NewValidationChain(
		v.logger,
		WithOperationLabel("CreateOrganization"),
		WithBaseContext(map[string]interface{}{"operation": "CreateOrganization"}),
	)

	parentCode := ""
	if req.ParentCode != nil {
		parentCode = strings.TrimSpace(*req.ParentCode)
	}

	if parentCode != "" {
		chain.Register(&Rule{
			ID:           "ORG-DEPTH",
			Priority:     10,
			Severity:     SeverityHigh,
			ShortCircuit: true,
			Handler:      v.newOrgDepthRule(parentCode),
		})

		if req.EffectiveDate != nil {
			effective := req.EffectiveDate.Time
			chain.Register(&Rule{
				ID:           "ORG-TEMPORAL",
				Priority:     15,
				Severity:     SeverityHigh,
				ShortCircuit: true,
				Handler:      v.newOrgTemporalRule(parentCode, &effective),
			})
		}
	}

	return chain
}

func (v *BusinessRuleValidator) buildOrganizationUpdateChain(existing *types.Organization, req *types.UpdateOrganizationRequest) *ValidationChain {
	chain := NewValidationChain(
		v.logger,
		WithOperationLabel("UpdateOrganization"),
		WithBaseContext(map[string]interface{}{"operation": "UpdateOrganization"}),
	)

	parentCode := ""
	if req.ParentCode != nil {
		parentCode = strings.TrimSpace(*req.ParentCode)
	}

	if parentCode != "" {
		chain.Register(&Rule{
			ID:           "ORG-DEPTH",
			Priority:     10,
			Severity:     SeverityHigh,
			ShortCircuit: true,
			Handler:      v.newOrgDepthRule(parentCode),
		})

		chain.Register(&Rule{
			ID:           "ORG-CIRC",
			Priority:     20,
			Severity:     SeverityCritical,
			ShortCircuit: true,
			Handler:      v.newOrgCircularRule(parentCode),
		})
	}

	if req.Status != nil {
		chain.Register(&Rule{
			ID:       "ORG-STATUS",
			Priority: 30,
			Severity: SeverityCritical,
			Handler:  v.newOrgStatusRule(existing),
		})
	}

	if parentCode != "" {
		var effectiveAt *time.Time
		if req != nil && req.EffectiveDate != nil {
			t := req.EffectiveDate.Time
			effectiveAt = &t
		} else if existing != nil && existing.EffectiveDate != nil {
			t := existing.EffectiveDate.Time
			effectiveAt = &t
		}

		if effectiveAt != nil {
			chain.Register(&Rule{
				ID:           "ORG-TEMPORAL",
				Priority:     25,
				Severity:     SeverityHigh,
				ShortCircuit: true,
				Handler:      v.newOrgTemporalRule(parentCode, effectiveAt),
			})
		}
	}

	return chain
}

func (v *BusinessRuleValidator) newOrgDepthRule(parentCode string) RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		switch s := subject.(type) {
		case *organizationCreateSubject:
			return v.evaluateDepth(ctx, s.TenantID, parentCode)
		case *organizationUpdateSubject:
			return v.evaluateDepth(ctx, s.TenantID, parentCode)
		default:
			return nil, fmt.Errorf("ORG-DEPTH rule: unsupported subject type %T", subject)
		}
	}
}

func (v *BusinessRuleValidator) evaluateDepth(ctx context.Context, tenantID uuid.UUID, parentCode string) (*RuleOutcome, error) {
	depth, err := v.hierarchyRepo.GetOrganizationDepth(ctx, parentCode, tenantID)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "depth not found") {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "INVALID_PARENT",
					Message:  fmt.Sprintf("Parent organization %s does not exist", parentCode),
					Field:    "parentCode",
					Severity: string(SeverityHigh),
					Context: map[string]interface{}{
						"ruleId": "ORG-DEPTH",
					},
				}},
			}, nil
		}
		return nil, fmt.Errorf("fetch parent depth failed: %w", err)
	}

	attemptedDepth := depth + 1

	if attemptedDepth > maxOrganizationDepth {
		return &RuleOutcome{
			Errors: []ValidationError{{
				Code:     "ORG_DEPTH_LIMIT",
				Message:  fmt.Sprintf("Organization depth exceeds maximum of %d levels", maxOrganizationDepth),
				Field:    "parentCode",
				Severity: string(SeverityHigh),
				Context: map[string]interface{}{
					"ruleId":         "ORG-DEPTH",
					"maxDepth":       maxOrganizationDepth,
					"attemptedDepth": attemptedDepth,
				},
			}},
		}, nil
	}

	if attemptedDepth >= depthWarningThreshold {
		return &RuleOutcome{
			Warnings: []ValidationWarning{{
				Code:    "ORG_DEPTH_NEAR_LIMIT",
				Message: fmt.Sprintf("Organization depth is near the limit (%d/%d)", attemptedDepth, maxOrganizationDepth),
				Field:   "parentCode",
				Value:   attemptedDepth,
			}},
			Context: map[string]interface{}{
				"ruleId":         "ORG-DEPTH",
				"attemptedDepth": attemptedDepth,
			},
		}, nil
	}

	return nil, nil
}

func (v *BusinessRuleValidator) newOrgCircularRule(parentCode string) RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		update, ok := subject.(*organizationUpdateSubject)
		if !ok {
			return nil, fmt.Errorf("ORG-CIRC rule expects organizationUpdateSubject, got %T", subject)
		}

		if parentCode == "" {
			return nil, nil
		}
		if strings.EqualFold(parentCode, update.Code) {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "ORG_CYCLE_DETECTED",
					Message:  "Organization cannot be its own parent",
					Field:    "parentCode",
					Severity: string(SeverityCritical),
					Context: map[string]interface{}{
						"ruleId":          "ORG-CIRC",
						"attemptedParent": parentCode,
					},
				}},
			}, nil
		}

		ancestors, err := v.hierarchyRepo.GetAncestorChain(ctx, parentCode, update.TenantID)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "not found") {
				return &RuleOutcome{
					Errors: []ValidationError{{
						Code:     "INVALID_PARENT",
						Message:  fmt.Sprintf("Parent organization %s does not exist", parentCode),
						Field:    "parentCode",
						Severity: string(SeverityHigh),
						Context: map[string]interface{}{
							"ruleId": "ORG-CIRC",
						},
					}},
				}, nil
			}
			return nil, fmt.Errorf("fetch ancestor chain failed: %w", err)
		}

		for _, ancestor := range ancestors {
			if strings.EqualFold(ancestor.Code, update.Code) {
				return &RuleOutcome{
					Errors: []ValidationError{{
						Code:     "ORG_CYCLE_DETECTED",
						Message:  fmt.Sprintf("Detected circular reference: %s -> %s", update.Code, parentCode),
						Field:    "parentCode",
						Severity: string(SeverityCritical),
						Context: map[string]interface{}{
							"ruleId":           "ORG-CIRC",
							"attemptedParent":  parentCode,
							"ancestorDetected": ancestor.Code,
						},
					}},
				}, nil
			}
		}

		return nil, nil
	}
}

func (v *BusinessRuleValidator) newOrgStatusRule(existing *types.Organization) RuleHandler {
	return func(_ context.Context, subject interface{}) (*RuleOutcome, error) {
		update, ok := subject.(*organizationUpdateSubject)
		if !ok {
			return nil, fmt.Errorf("ORG-STATUS rule expects organizationUpdateSubject, got %T", subject)
		}
		if update.Request == nil || update.Request.Status == nil {
			return nil, nil
		}

		currentStatus := strings.ToUpper(strings.TrimSpace(existing.Status))
		newStatus := strings.ToUpper(strings.TrimSpace(*update.Request.Status))

		if currentStatus == "" || newStatus == "" {
			return nil, nil
		}
		if currentStatus == newStatus {
			return nil, nil
		}

		validTransitions := map[string][]string{
			"ACTIVE":   {"INACTIVE", "DELETED"},
			"INACTIVE": {"ACTIVE", "DELETED"},
			"PLANNED":  {"ACTIVE", "DELETED"},
		}

		targets, ok := validTransitions[currentStatus]
		if !ok {
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "ORG_STATUS_GUARD",
					Message:  fmt.Sprintf("Unsupported current status %s for transition", currentStatus),
					Field:    "status",
					Severity: string(SeverityCritical),
					Context: map[string]interface{}{
						"ruleId":          "ORG-STATUS",
						"currentStatus":   currentStatus,
						"requestedStatus": newStatus,
					},
				}},
			}, nil
		}

		for _, allowed := range targets {
			if allowed == newStatus {
				return nil, nil
			}
		}

		return &RuleOutcome{
			Errors: []ValidationError{{
				Code:     "ORG_STATUS_GUARD",
				Message:  fmt.Sprintf("Cannot transition from %s to %s", currentStatus, newStatus),
				Field:    "status",
				Severity: string(SeverityCritical),
				Context: map[string]interface{}{
					"ruleId":          "ORG-STATUS",
					"currentStatus":   currentStatus,
					"requestedStatus": newStatus,
				},
			}},
		}, nil
	}
}

func (v *BusinessRuleValidator) newOrgTemporalRule(parentCode string, effectiveDate *time.Time) RuleHandler {
	return func(ctx context.Context, subject interface{}) (*RuleOutcome, error) {
		if parentCode == "" || effectiveDate == nil {
			return nil, nil
		}

		var tenantID uuid.UUID
		switch s := subject.(type) {
		case *organizationCreateSubject:
			tenantID = s.TenantID
		case *organizationUpdateSubject:
			tenantID = s.TenantID
		default:
			return nil, fmt.Errorf("ORG-TEMPORAL rule: unsupported subject type %T", subject)
		}

		parent, err := v.hierarchyRepo.GetOrganizationAtDate(ctx, parentCode, tenantID, *effectiveDate)
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "not found") {
				return &RuleOutcome{
					Errors: []ValidationError{{
						Code:     "INVALID_PARENT",
						Message:  fmt.Sprintf("Parent organization %s does not exist", parentCode),
						Field:    "parentCode",
						Severity: string(SeverityHigh),
						Context: map[string]interface{}{
							"ruleId": "ORG-TEMPORAL",
						},
					}},
				}, nil
			}
			return nil, fmt.Errorf("fetch temporal parent failed: %w", err)
		}

		if parent == nil || !strings.EqualFold(parent.Status, "ACTIVE") {
			ctxMap := map[string]interface{}{
				"ruleId":     "ORG-TEMPORAL",
				"parentCode": parentCode,
				"effective":  effectiveDate.Format(time.RFC3339),
			}
			return &RuleOutcome{
				Errors: []ValidationError{{
					Code:     "ORG_TEMPORAL_PARENT_INACTIVE",
					Message:  fmt.Sprintf("Parent %s is not active at %s", parentCode, effectiveDate.Format("2006-01-02")),
					Field:    "parentCode",
					Severity: string(SeverityHigh),
					Context:  ctxMap,
				}},
			}, nil
		}

		return nil, nil
	}
}
