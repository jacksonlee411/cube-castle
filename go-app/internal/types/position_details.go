package types

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

// PositionDetails defines the base interface for all position detail profiles
// This implements the polymorphic "details slot" design pattern from Meta-Contract v6.0
type PositionDetails interface {
	GetType() string
	Validate() error
}

// FullTimePositionDetails represents a regular full-time position profile
// Maps to FULL_TIME position_type in the discriminator pattern
type FullTimePositionDetails struct {
	// Salary band or grade level
	SalaryBand string `json:"salary_band,omitempty"`
	
	// Annual salary range
	SalaryRange map[string]float64 `json:"salary_range,omitempty"` // {"min": 50000, "max": 80000}
	
	// Currency for salary
	Currency string `json:"currency,omitempty"`
	
	// Bonus eligibility flag
	BonusEligible bool `json:"bonus_eligible,omitempty"`
	
	// Stock options eligibility
	StockOptionsEligible bool `json:"stock_options_eligible,omitempty"`
	
	// Benefits package reference
	BenefitsPackageID *uuid.UUID `json:"benefits_package_id,omitempty"`
	
	// Career ladder or progression path
	CareerLadder string `json:"career_ladder,omitempty"`
	
	// Performance review cycle (e.g., "ANNUAL", "SEMI_ANNUAL")
	ReviewCycle string `json:"review_cycle,omitempty"`
}

func (f FullTimePositionDetails) GetType() string {
	return "FULL_TIME"
}

func (f FullTimePositionDetails) Validate() error {
	if f.Currency != "" && len(f.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// PartTimePositionDetails represents a part-time position profile
// Maps to PART_TIME position_type in the discriminator pattern
type PartTimePositionDetails struct {
	// Standard hours per week
	StandardHoursPerWeek float64 `json:"standard_hours_per_week"`
	
	// Hourly wage rate
	HourlyRate *float64 `json:"hourly_rate,omitempty"`
	
	// Currency for hourly rate
	Currency string `json:"currency,omitempty"`
	
	// Flexible schedule allowed
	FlexibleSchedule bool `json:"flexible_schedule,omitempty"`
	
	// Minimum hours per week required
	MinHoursPerWeek *float64 `json:"min_hours_per_week,omitempty"`
	
	// Maximum hours per week allowed
	MaxHoursPerWeek *float64 `json:"max_hours_per_week,omitempty"`
	
	// Benefits eligibility (often prorated)
	BenefitsEligible bool `json:"benefits_eligible,omitempty"`
	
	// Overtime eligibility
	OvertimeEligible bool `json:"overtime_eligible,omitempty"`
	
	// Work days pattern (e.g., ["MON", "TUE", "WED"])
	WorkDaysPattern []string `json:"work_days_pattern,omitempty"`
}

func (p PartTimePositionDetails) GetType() string {
	return "PART_TIME"
}

func (p PartTimePositionDetails) Validate() error {
	if p.StandardHoursPerWeek <= 0 {
		return fmt.Errorf("standard_hours_per_week must be greater than 0")
	}
	if p.StandardHoursPerWeek >= 40 {
		return fmt.Errorf("standard_hours_per_week for part-time position should be less than 40")
	}
	if p.Currency != "" && len(p.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// ContingentWorkerDetails represents an external contractor or temporary worker position
// Maps to CONTINGENT_WORKER position_type in the discriminator pattern
type ContingentWorkerDetails struct {
	// Vendor/supplier company ID
	VendorID *uuid.UUID `json:"vendor_id,omitempty"`
	
	// Hourly rate for contractor
	HourlyRate float64 `json:"hourly_rate"`
	
	// Currency for rate
	Currency string `json:"currency,omitempty"`
	
	// Contract start date
	ContractStartDate string `json:"contract_start_date,omitempty"`
	
	// Contract end date
	ContractEndDate string `json:"contract_end_date"`
	
	// Maximum hours per week allowed
	MaxHoursPerWeek *float64 `json:"max_hours_per_week,omitempty"`
	
	// Contract type (e.g., "STATEMENT_OF_WORK", "TIME_AND_MATERIALS")
	ContractType string `json:"contract_type,omitempty"`
	
	// Purchase order number
	PurchaseOrderNumber string `json:"purchase_order_number,omitempty"`
	
	// Statement of work reference
	StatementOfWorkID string `json:"statement_of_work_id,omitempty"`
	
	// Security clearance required
	SecurityClearanceRequired bool `json:"security_clearance_required,omitempty"`
	
	// On-site work required
	OnsiteRequired bool `json:"onsite_required,omitempty"`
	
	// Equipment provided by company
	EquipmentProvided bool `json:"equipment_provided,omitempty"`
	
	// Skills or certifications required
	RequiredSkills []string `json:"required_skills,omitempty"`
}

func (c ContingentWorkerDetails) GetType() string {
	return "CONTINGENT_WORKER"
}

func (c ContingentWorkerDetails) Validate() error {
	if c.HourlyRate <= 0 {
		return fmt.Errorf("hourly_rate must be greater than 0")
	}
	if c.ContractEndDate == "" {
		return fmt.Errorf("contract_end_date is required for contingent worker")
	}
	if c.Currency != "" && len(c.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// InternPositionDetails represents an internship position profile
// Maps to INTERN position_type in the discriminator pattern
type InternPositionDetails struct {
	// Internship program ID reference
	InternshipProgramID *uuid.UUID `json:"internship_program_id,omitempty"`
	
	// Mentor person ID
	MentorPersonID *uuid.UUID `json:"mentor_person_id"`
	
	// Academic institution
	AcademicInstitution string `json:"academic_institution,omitempty"`
	
	// Expected graduation date
	ExpectedGraduationDate string `json:"expected_graduation_date,omitempty"`
	
	// Major/field of study
	FieldOfStudy string `json:"field_of_study,omitempty"`
	
	// Internship duration in weeks
	DurationWeeks int `json:"duration_weeks,omitempty"`
	
	// Stipend amount (if paid internship)
	Stipend *float64 `json:"stipend,omitempty"`
	
	// Currency for stipend
	Currency string `json:"currency,omitempty"`
	
	// Academic credit offered
	AcademicCredit bool `json:"academic_credit,omitempty"`
	
	// Hours per week expected
	HoursPerWeek float64 `json:"hours_per_week,omitempty"`
	
	// Learning objectives
	LearningObjectives []string `json:"learning_objectives,omitempty"`
	
	// Evaluation criteria
	EvaluationCriteria []string `json:"evaluation_criteria,omitempty"`
	
	// Final presentation required
	FinalPresentationRequired bool `json:"final_presentation_required,omitempty"`
}

func (i InternPositionDetails) GetType() string {
	return "INTERN"
}

func (i InternPositionDetails) Validate() error {
	if i.MentorPersonID == nil {
		return fmt.Errorf("mentor_person_id is required for intern position")
	}
	if i.HoursPerWeek <= 0 {
		return fmt.Errorf("hours_per_week must be greater than 0")
	}
	if i.Currency != "" && len(i.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// PositionDetailsFactory creates the appropriate details type based on position_type discriminator
func PositionDetailsFactory(positionType string, detailsData json.RawMessage) (PositionDetails, error) {
	switch positionType {
	case "FULL_TIME":
		var details FullTimePositionDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal full-time position details: %w", err)
		}
		return details, nil
		
	case "PART_TIME":
		var details PartTimePositionDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal part-time position details: %w", err)
		}
		return details, nil
		
	case "CONTINGENT_WORKER":
		var details ContingentWorkerDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal contingent worker details: %w", err)
		}
		return details, nil
		
	case "INTERN":
		var details InternPositionDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal intern position details: %w", err)
		}
		return details, nil
		
	default:
		return nil, fmt.Errorf("unsupported position type: %s", positionType)
	}
}

// PositionDetailsToJSON converts position details back to JSON for storage
func PositionDetailsToJSON(details PositionDetails) (json.RawMessage, error) {
	data, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal position details: %w", err)
	}
	return json.RawMessage(data), nil
}