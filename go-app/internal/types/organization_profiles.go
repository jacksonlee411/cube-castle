package types

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

// OrganizationUnitProfile defines the base interface for all organization unit profiles
// This implements the polymorphic "profile slot" design pattern from Meta-Contract v6.0
type OrganizationUnitProfile interface {
	GetType() string
	Validate() error
}

// DepartmentProfile represents a standard functional or business unit profile
// Maps to DEPARTMENT unit_type in the discriminator pattern
type DepartmentProfile struct {
	// Head of unit person reference
	HeadOfUnitPersonID *uuid.UUID `json:"head_of_unit_person_id,omitempty"`
	
	// Functional area classification
	FunctionalArea string `json:"functional_area,omitempty"`
	
	// Associated cost center for financial reporting
	CostCenter string `json:"cost_center,omitempty"`
	
	// Department-specific capabilities or focus areas
	Capabilities []string `json:"capabilities,omitempty"`
	
	// Budget allocation responsibility flag
	HasBudgetResponsibility bool `json:"has_budget_responsibility,omitempty"`
}

func (d DepartmentProfile) GetType() string {
	return "DEPARTMENT"
}

func (d DepartmentProfile) Validate() error {
	if d.FunctionalArea == "" {
		return fmt.Errorf("functional_area is required for department profile")
	}
	return nil
}

// CostCenterProfile represents a financial responsibility unit profile
// Maps to COST_CENTER unit_type in the discriminator pattern
type CostCenterProfile struct {
	// Unique business code for the cost center
	CostCenterCode string `json:"cost_center_code"`
	
	// Financial owner/manager person reference
	FinancialOwnerID *uuid.UUID `json:"financial_owner_id,omitempty"`
	
	// Budget allocation amount
	BudgetAllocation *float64 `json:"budget_allocation,omitempty"`
	
	// Currency for budget allocation
	Currency string `json:"currency,omitempty"`
	
	// Fiscal year for budget period
	FiscalYear string `json:"fiscal_year,omitempty"`
	
	// Cost center category (e.g., "OPERATIONAL", "STRATEGIC", "SUPPORT")
	Category string `json:"category,omitempty"`
	
	// Whether this cost center can be charged to projects
	IsChargeable bool `json:"is_chargeable,omitempty"`
}

func (c CostCenterProfile) GetType() string {
	return "COST_CENTER"
}

func (c CostCenterProfile) Validate() error {
	if c.CostCenterCode == "" {
		return fmt.Errorf("cost_center_code is required for cost center profile")
	}
	if c.Currency != "" && len(c.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// CompanyProfile represents a legal entity profile
// Maps to COMPANY unit_type in the discriminator pattern
type CompanyProfile struct {
	// Official legal entity registration number
	LegalEntityID string `json:"legal_entity_id,omitempty"`
	
	// Tax registration identifier
	TaxID string `json:"tax_id,omitempty"`
	
	// Legal entity name (may differ from display name)
	LegalName string `json:"legal_name,omitempty"`
	
	// Primary business address
	BusinessAddress map[string]interface{} `json:"business_address,omitempty"`
	
	// Incorporation date
	IncorporationDate string `json:"incorporation_date,omitempty"`
	
	// Legal structure (e.g., "LLC", "CORP", "LTD")
	LegalStructure string `json:"legal_structure,omitempty"`
	
	// Primary business jurisdiction
	Jurisdiction string `json:"jurisdiction,omitempty"`
	
	// Industry classification codes
	IndustryCodes []string `json:"industry_codes,omitempty"`
}

func (c CompanyProfile) GetType() string {
	return "COMPANY"
}

func (c CompanyProfile) Validate() error {
	if c.LegalEntityID == "" && c.TaxID == "" {
		return fmt.Errorf("either legal_entity_id or tax_id is required for company profile")
	}
	return nil
}

// ProjectTeamProfile represents a temporary, cross-functional team profile
// Maps to PROJECT_TEAM unit_type in the discriminator pattern
type ProjectTeamProfile struct {
	// Project lead person reference
	ProjectLeadPersonID *uuid.UUID `json:"project_lead_person_id,omitempty"`
	
	// Project start date
	ProjectStartDate string `json:"project_start_date,omitempty"`
	
	// Project expected end date
	ProjectEndDate string `json:"project_end_date,omitempty"`
	
	// Project status (e.g., "PLANNING", "ACTIVE", "ON_HOLD", "COMPLETED")
	ProjectStatus string `json:"project_status,omitempty"`
	
	// Project budget allocation
	ProjectBudget *float64 `json:"project_budget,omitempty"`
	
	// Project priority level
	Priority string `json:"priority,omitempty"`
	
	// Stakeholder references
	Stakeholders []uuid.UUID `json:"stakeholders,omitempty"`
	
	// Expected team size
	TargetTeamSize int `json:"target_team_size,omitempty"`
	
	// Project objectives
	Objectives []string `json:"objectives,omitempty"`
	
	// Success criteria
	SuccessCriteria []string `json:"success_criteria,omitempty"`
}

func (p ProjectTeamProfile) GetType() string {
	return "PROJECT_TEAM"
}

func (p ProjectTeamProfile) Validate() error {
	if p.ProjectLeadPersonID == nil {
		return fmt.Errorf("project_lead_person_id is required for project team profile")
	}
	if p.ProjectStartDate == "" {
		return fmt.Errorf("project_start_date is required for project team profile")
	}
	return nil
}

// ProfileFactory creates the appropriate profile type based on unit_type discriminator
// This implements the polymorphic instantiation pattern
func ProfileFactory(unitType string, profileData json.RawMessage) (OrganizationUnitProfile, error) {
	switch unitType {
	case "DEPARTMENT":
		var profile DepartmentProfile
		if err := json.Unmarshal(profileData, &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal department profile: %w", err)
		}
		return profile, nil
		
	case "COST_CENTER":
		var profile CostCenterProfile
		if err := json.Unmarshal(profileData, &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal cost center profile: %w", err)
		}
		return profile, nil
		
	case "COMPANY":
		var profile CompanyProfile
		if err := json.Unmarshal(profileData, &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal company profile: %w", err)
		}
		return profile, nil
		
	case "PROJECT_TEAM":
		var profile ProjectTeamProfile
		if err := json.Unmarshal(profileData, &profile); err != nil {
			return nil, fmt.Errorf("failed to unmarshal project team profile: %w", err)
		}
		return profile, nil
		
	default:
		return nil, fmt.Errorf("unsupported unit type: %s", unitType)
	}
}

// ProfileToJSON converts a profile back to JSON for storage
func ProfileToJSON(profile OrganizationUnitProfile) (json.RawMessage, error) {
	data, err := json.Marshal(profile)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal profile: %w", err)
	}
	return json.RawMessage(data), nil
}