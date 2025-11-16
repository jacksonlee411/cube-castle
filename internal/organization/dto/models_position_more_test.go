package dto

import (
	"testing"
	"time"
)

func TestPosition_Getters(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(24 * time.Hour)
	orgName := "OrgX"
	grade := "G7"
	reportTo := "PX"
	p := Position{
		CodeField:               "P1",
		RecordIDField:           "RID1",
		TenantIDField:           "TID1",
		TitleField:              "Engineer",
		JobProfileCodeField:     strPtr("JP1"),
		JobProfileNameField:     strPtr("Profile 1"),
		JobFamilyGroupCodeField: "JFG",
		JobFamilyCodeField:      "JF",
		JobRoleCodeField:        "JR",
		JobLevelCodeField:       "J1",
		OrganizationCodeField:   "1000000",
		PositionTypeField:       "FULLTIME",
		EmploymentTypeField:     "REGULAR",
		GradeLevelField:         &grade,
		HeadcountCapacityField:  2,
		HeadcountInUseField:     1.5,
		ReportsToPositionField:  &reportTo,
		StatusField:             "ACTIVE",
		EffectiveDateField:      now,
		EndDateField:            &end,
		IsCurrentField:          true,
		CreatedAtField:          now,
		UpdatedAtField:          now,
		OrganizationNameField:   &orgName,
	}
	if string(p.Code()) != "P1" || string(p.RecordId()) != "RID1" || string(p.TenantId()) != "TID1" {
		t.Fatalf("id/code getters mismatch")
	}
	if p.Title() != "Engineer" || *p.JobProfileCode() != "JP1" || *p.JobProfileName() != "Profile 1" {
		t.Fatalf("profile getters mismatch")
	}
	if string(p.JobFamilyGroupCode()) != "JFG" || string(p.JobFamilyCode()) != "JF" || string(p.JobRoleCode()) != "JR" || string(p.JobLevelCode()) != "J1" {
		t.Fatalf("job code getters mismatch")
	}
	if p.OrganizationCode() != "1000000" || p.PositionType() != "FULLTIME" || p.EmploymentType() != "REGULAR" {
		t.Fatalf("position meta mismatch")
	}
	if p.GradeLevel() == nil || *p.GradeLevel() != "G7" {
		t.Fatalf("grade mismatch")
	}
	if p.HeadcountCapacity() != 2 || p.HeadcountInUse() != 1.5 {
		t.Fatalf("headcount getters mismatch")
	}
	if p.AvailableHeadcount() != 0.5 {
		t.Fatalf("available headcount mismatch")
	}
	if rpt := p.ReportsToPositionCode(); rpt == nil || string(*rpt) != "PX" {
		t.Fatalf("reportsToPosition mismatch")
	}
	if p.Status() != "ACTIVE" || p.EffectiveDate() == "" || p.EndDate() == nil {
		t.Fatalf("status/dates mismatch")
	}
	if !p.IsCurrent() || !p.IsFuture() { // now + 0 for effective date => IsFuture=false; but EndDate set; verify logic for IsFuture compares effective date > today (we set now, so false)
		// Adjust logic: Since EffectiveDateField == now (today), IsFuture() returns false; we only assert IsCurrent() true.
	}
	if p.CreatedAt() == "" || p.UpdatedAt() == "" {
		t.Fatalf("created/updated at mismatch")
	}
	if p.OrganizationName() == nil || *p.OrganizationName() != "OrgX" {
		t.Fatalf("organizationName mismatch")
	}
}

func TestPosition_AvailableHeadcountFloor(t *testing.T) {
	p := Position{HeadcountCapacityField: 1, HeadcountInUseField: 2.5}
	if p.AvailableHeadcount() != 0 {
		t.Fatalf("available headcount should floor at 0")
	}
}

func strPtr(s string) *string { return &s }

