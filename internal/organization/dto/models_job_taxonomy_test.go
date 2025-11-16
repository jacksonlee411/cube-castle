package dto

import (
	"testing"
	"time"
)

func TestHeadcountStatsAndBreakdowns_Getters(t *testing.T) {
	lv := LevelHeadcount{JobLevelCodeField: "J1", CapacityField: 10, UtilizedField: 7, AvailableField: 3}
	ty := TypeHeadcount{PositionTypeField: "FULLTIME", CapacityField: 5, FilledField: 4, AvailableField: 1}
	famName := "FamilyA"
	fm := FamilyHeadcount{JobFamilyCodeField: "JF", JobFamilyNameField: &famName, CapacityField: 8, UtilizedField: 6, AvailableField: 2}
	h := HeadcountStats{
		OrganizationCodeField: "1000000",
		OrganizationNameField: "Org",
		TotalCapacityField:    20,
		TotalFilledField:      15,
		TotalAvailableField:   5,
		LevelBreakdownField:   []LevelHeadcount{lv},
		TypeBreakdownField:    []TypeHeadcount{ty},
		FamilyBreakdownField:  []FamilyHeadcount{fm},
	}
	if h.OrganizationCode() != "1000000" || h.OrganizationName() != "Org" || h.TotalCapacity() != 20 || h.TotalFilled() != 15 || h.TotalAvailable() != 5 {
		t.Fatalf("headcount stats totals mismatch")
	}
	// simple coverage of computed methods
	_ = h.FillRate()
	if len(h.ByLevel()) != 1 || len(h.ByType()) != 1 || len(h.ByFamily()) != 1 {
		t.Fatalf("headcount breakdown getters mismatch")
	}
	if string(lv.JobLevelCode()) != "J1" || lv.Capacity() != 10 || lv.Utilized() != 7 || lv.Available() != 3 {
		t.Fatalf("level headcount getters mismatch")
	}
	if ty.PositionType() != "FULLTIME" || ty.Capacity() != 5 || ty.Filled() != 4 || ty.Available() != 1 {
		t.Fatalf("type headcount getters mismatch")
	}
	if string(fm.JobFamilyCode()) != "JF" || fm.JobFamilyName() == nil || *fm.JobFamilyName() != "FamilyA" {
		t.Fatalf("family headcount getters mismatch")
	}
}

func TestJobTaxonomyEntities_Getters(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(24 * time.Hour)
	desc := "desc"

	g := JobFamilyGroup{
		RecordIDField:      "RIDG",
		TenantIDField:      "TID",
		CodeField:          "JFG",
		NameField:          "Group",
		DescriptionField:   &desc,
		StatusField:        "ACTIVE",
		EffectiveDateField: now,
		EndDateField:       &end,
		IsCurrentField:     true,
	}
	if string(g.RecordId()) == "" || string(g.TenantId()) == "" || string(g.Code()) != "JFG" || g.Name() != "Group" {
		t.Fatalf("JobFamilyGroup basic getters mismatch")
	}
	_ = g.EffectiveDate()
	_ = g.EndDate()
	if !g.IsCurrent() {
		t.Fatalf("JobFamilyGroup isCurrent mismatch")
	}

	f := JobFamily{
		RecordIDField:        "RIDF",
		TenantIDField:        "TID",
		CodeField:            "JF",
		NameField:            "Family",
		DescriptionField:     &desc,
		StatusField:          "ACTIVE",
		EffectiveDateField:   now,
		EndDateField:         &end,
		IsCurrentField:       true,
		FamilyGroupCodeField: "JFG",
	}
	if string(f.RecordId()) == "" || string(f.TenantId()) == "" || string(f.Code()) != "JF" || f.Name() != "Family" {
		t.Fatalf("JobFamily basic getters mismatch")
	}
	_ = f.EffectiveDate()
	_ = f.EndDate()
	if !f.IsCurrent() || string(f.GroupCode()) != "JFG" {
		t.Fatalf("JobFamily group/isCurrent mismatch")
	}

	r := JobRole{
		RecordIDField:      "RIDR",
		TenantIDField:      "TID",
		CodeField:          "JR",
		NameField:          "Role",
		DescriptionField:   &desc,
		StatusField:        "ACTIVE",
		EffectiveDateField: now,
		EndDateField:       &end,
		IsCurrentField:     true,
		FamilyCodeField:    "JF",
	}
	if string(r.RecordId()) == "" || string(r.TenantId()) == "" || string(r.Code()) != "JR" || r.Name() != "Role" {
		t.Fatalf("JobRole basic getters mismatch")
	}
	_ = r.EffectiveDate()
	_ = r.EndDate()
	if !r.IsCurrent() || string(r.FamilyCode()) != "JF" {
		t.Fatalf("JobRole family/isCurrent mismatch")
	}

	l := JobLevel{
		RecordIDField:      "RIDL",
		TenantIDField:      "TID",
		CodeField:          "JL",
		NameField:          "Level",
		DescriptionField:   &desc,
		StatusField:        "ACTIVE",
		EffectiveDateField: now,
		EndDateField:       &end,
		IsCurrentField:     true,
		RoleCodeField:      "JR",
		LevelRankField:     "3",
	}
	if string(l.RecordId()) == "" || string(l.TenantId()) == "" || string(l.Code()) != "JL" || l.Name() != "Level" {
		t.Fatalf("JobLevel basic getters mismatch")
	}
	_ = l.EffectiveDate()
	_ = l.EndDate()
	if !l.IsCurrent() || string(l.RoleCode()) != "JR" || l.LevelRank() != 3 {
		t.Fatalf("JobLevel role/levelRank mismatch")
	}
}
