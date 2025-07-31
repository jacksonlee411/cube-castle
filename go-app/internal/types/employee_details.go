package types

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
)

// EmployeeDetails defines the base interface for all employee detail profiles
// This implements the polymorphic "details slot" design pattern from Meta-Contract v6.0
type EmployeeDetails interface {
	GetType() string
	Validate() error
}

// FullTimeEmployeeDetails represents a regular full-time employee profile
// Maps to FULL_TIME employee_type in the discriminator pattern
type FullTimeEmployeeDetails struct {
	// Employment classification
	EmploymentLevel string `json:"employment_level,omitempty"` // ENTRY, MID, SENIOR, EXECUTIVE

	// Compensation details
	AnnualSalary *float64 `json:"annual_salary,omitempty"`
	Currency     string   `json:"currency,omitempty"`

	// Benefits and perks
	BenefitsPackageID    *uuid.UUID `json:"benefits_package_id,omitempty"`
	VacationDaysPerYear  int        `json:"vacation_days_per_year,omitempty"`
	SickDaysPerYear      int        `json:"sick_days_per_year,omitempty"`
	BonusEligible        bool       `json:"bonus_eligible,omitempty"`
	StockOptionsEligible bool       `json:"stock_options_eligible,omitempty"`

	// Performance and career
	PerformanceReviewCycle string `json:"performance_review_cycle,omitempty"` // ANNUAL, SEMI_ANNUAL, QUARTERLY
	CareerTrack            string `json:"career_track,omitempty"`             // INDIVIDUAL_CONTRIBUTOR, MANAGEMENT, TECHNICAL_LEADERSHIP

	// Work arrangement
	WorkLocation string `json:"work_location,omitempty"` // ON_SITE, REMOTE, HYBRID
	TimeZone     string `json:"time_zone,omitempty"`
}

func (f FullTimeEmployeeDetails) GetType() string {
	return "FULL_TIME"
}

func (f FullTimeEmployeeDetails) Validate() error {
	if f.Currency != "" && len(f.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	if f.VacationDaysPerYear < 0 || f.VacationDaysPerYear > 365 {
		return fmt.Errorf("vacation_days_per_year must be between 0 and 365")
	}
	if f.SickDaysPerYear < 0 || f.SickDaysPerYear > 365 {
		return fmt.Errorf("sick_days_per_year must be between 0 and 365")
	}
	return nil
}

// PartTimeEmployeeDetails represents a part-time employee profile
// Maps to PART_TIME employee_type in the discriminator pattern
type PartTimeEmployeeDetails struct {
	// Work schedule
	HoursPerWeek     float64  `json:"hours_per_week"`
	WorkDaysPattern  []string `json:"work_days_pattern,omitempty"` // ["MON", "TUE", "WED"]
	FlexibleSchedule bool     `json:"flexible_schedule,omitempty"`

	// Compensation
	HourlyRate *float64 `json:"hourly_rate,omitempty"`
	Currency   string   `json:"currency,omitempty"`

	// Benefits (often prorated)
	BenefitsEligible     bool       `json:"benefits_eligible,omitempty"`
	BenefitsPackageID    *uuid.UUID `json:"benefits_package_id,omitempty"`
	VacationDaysPerYear  int        `json:"vacation_days_per_year,omitempty"`
	OvertimeEligible     bool       `json:"overtime_eligible,omitempty"`

	// Work arrangement
	WorkLocation string `json:"work_location,omitempty"`
	TimeZone     string `json:"time_zone,omitempty"`
}

func (p PartTimeEmployeeDetails) GetType() string {
	return "PART_TIME"
}

func (p PartTimeEmployeeDetails) Validate() error {
	if p.HoursPerWeek <= 0 {
		return fmt.Errorf("hours_per_week must be greater than 0")
	}
	if p.HoursPerWeek >= 40 {
		return fmt.Errorf("hours_per_week for part-time employee should be less than 40")
	}
	if p.Currency != "" && len(p.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// ContractorEmployeeDetails represents a contractor employee profile
// Maps to CONTRACTOR employee_type in the discriminator pattern
type ContractorEmployeeDetails struct {
	// Contract information
	ContractorCompany   string `json:"contractor_company,omitempty"`
	ContractNumber      string `json:"contract_number,omitempty"`
	ContractStartDate   string `json:"contract_start_date,omitempty"`
	ContractEndDate     string `json:"contract_end_date"`
	ContractType        string `json:"contract_type,omitempty"` // FIXED_TERM, PROJECT_BASED, ONGOING

	// Compensation
	HourlyRate       float64 `json:"hourly_rate"`
	Currency         string  `json:"currency,omitempty"`
	MaxHoursPerWeek  *int    `json:"max_hours_per_week,omitempty"`
	MaxHoursPerMonth *int    `json:"max_hours_per_month,omitempty"`

	// Work requirements
	OnSiteRequired           bool     `json:"on_site_required,omitempty"`
	SecurityClearanceLevel   string   `json:"security_clearance_level,omitempty"`
	RequiredSkills           []string `json:"required_skills,omitempty"`
	EquipmentProvided        bool     `json:"equipment_provided,omitempty"`

	// Invoice and payment
	InvoicingSchedule    string `json:"invoicing_schedule,omitempty"`    // WEEKLY, MONTHLY, PROJECT_MILESTONE
	PaymentTerms         string `json:"payment_terms,omitempty"`         // NET_15, NET_30, NET_45
	PurchaseOrderNumber  string `json:"purchase_order_number,omitempty"`
}

func (c ContractorEmployeeDetails) GetType() string {
	return "CONTRACTOR"
}

func (c ContractorEmployeeDetails) Validate() error {
	if c.HourlyRate <= 0 {
		return fmt.Errorf("hourly_rate must be greater than 0")
	}
	if c.ContractEndDate == "" {
		return fmt.Errorf("contract_end_date is required for contractor")
	}
	if c.Currency != "" && len(c.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	return nil
}

// InternEmployeeDetails represents an intern employee profile
// Maps to INTERN employee_type in the discriminator pattern
type InternEmployeeDetails struct {
	// Academic information
	AcademicInstitution    string `json:"academic_institution,omitempty"`
	FieldOfStudy           string `json:"field_of_study,omitempty"`
	ExpectedGraduationDate string `json:"expected_graduation_date,omitempty"`
	AcademicYear           string `json:"academic_year,omitempty"` // FRESHMAN, SOPHOMORE, JUNIOR, SENIOR, GRADUATE

	// Internship program
	InternshipProgramID *uuid.UUID `json:"internship_program_id,omitempty"`
	MentorEmployeeID    *uuid.UUID `json:"mentor_employee_id"`
	DurationWeeks       int        `json:"duration_weeks,omitempty"`

	// Schedule and compensation
	HoursPerWeek       float64  `json:"hours_per_week,omitempty"`
	Stipend            *float64 `json:"stipend,omitempty"`
	Currency           string   `json:"currency,omitempty"`
	AcademicCredit     bool     `json:"academic_credit,omitempty"`

	// Learning objectives
	LearningObjectives          []string `json:"learning_objectives,omitempty"`
	EvaluationCriteria          []string `json:"evaluation_criteria,omitempty"`
	FinalPresentationRequired   bool     `json:"final_presentation_required,omitempty"`
	AcademicSupervisorContact   string   `json:"academic_supervisor_contact,omitempty"`
}

func (i InternEmployeeDetails) GetType() string {
	return "INTERN"
}

func (i InternEmployeeDetails) Validate() error {
	if i.MentorEmployeeID == nil {
		return fmt.Errorf("mentor_employee_id is required for intern")
	}
	if i.HoursPerWeek != 0 && i.HoursPerWeek <= 0 {
		return fmt.Errorf("hours_per_week must be greater than 0 if specified")
	}
	if i.Currency != "" && len(i.Currency) != 3 {
		return fmt.Errorf("currency must be a valid 3-letter ISO code")
	}
	if i.DurationWeeks <= 0 {
		return fmt.Errorf("duration_weeks must be greater than 0")
	}
	return nil
}

// EmployeeDetailsFactory creates the appropriate details type based on employee_type discriminator
func EmployeeDetailsFactory(employeeType string, detailsData json.RawMessage) (EmployeeDetails, error) {
	switch employeeType {
	case "FULL_TIME":
		var details FullTimeEmployeeDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal full-time employee details: %w", err)
		}
		return details, nil

	case "PART_TIME":
		var details PartTimeEmployeeDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal part-time employee details: %w", err)
		}
		return details, nil

	case "CONTRACTOR":
		var details ContractorEmployeeDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal contractor employee details: %w", err)
		}
		return details, nil

	case "INTERN":
		var details InternEmployeeDetails
		if err := json.Unmarshal(detailsData, &details); err != nil {
			return nil, fmt.Errorf("failed to unmarshal intern employee details: %w", err)
		}
		return details, nil

	default:
		return nil, fmt.Errorf("unsupported employee type: %s", employeeType)
	}
}

// EmployeeDetailsToJSON converts employee details back to JSON for storage
func EmployeeDetailsToJSON(details EmployeeDetails) (json.RawMessage, error) {
	data, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal employee details: %w", err)
	}
	return json.RawMessage(data), nil
}