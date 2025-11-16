package dto

import (
	"testing"
	"time"
)

func TestClampToInt32(t *testing.T) {
	if got := clampToInt32(42); got != 42 {
		t.Fatalf("expected 42, got %d", got)
	}
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if got := clampToInt32(maxInt); got != 2147483647 {
		t.Fatalf("expected MaxInt32 clamp, got %d", got)
	}
	if got := clampToInt32(minInt); got != -2147483648 {
		t.Fatalf("expected MinInt32 clamp, got %d", got)
	}
}

func TestClampToInt32Ptr(t *testing.T) {
	if clampToInt32Ptr(nil) != nil {
		t.Fatalf("nil should return nil")
	}
	v := 7
	out := clampToInt32Ptr(&v)
	if out == nil || *out != 7 {
		t.Fatalf("expected 7, got %#v", out)
	}
}

func TestOrganization_Getters(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	end := now.Add(24 * time.Hour)
	name := "Org"
	parent := "1000000"
	profile := "p"
	desc := "d"
	sort := 3
	change := "reason"
	delBy := "userX"
	delReason := "cleanup"
	suspBy := "ops"
	suspReason := "policy"
	codePath := "/100/101"
	namePath := "/A/B"

	o := Organization{
		RecordIDField:         "rid",
		TenantIDField:         "tid",
		CodeField:             "1000100",
		ParentCodeField:       &parent,
		NameField:             name,
		UnitTypeField:         "DEPARTMENT",
		StatusField:           "ACTIVE",
		LevelField:            2,
		CodePathField:         codePath,
		NamePathField:         namePath,
		SortOrderField:        &sort,
		DescriptionField:      &desc,
		ProfileField:          &profile,
		CreatedAtField:        now,
		UpdatedAtField:        now,
		EffectiveDateField:    now,
		EndDateField:          &end,
		IsCurrentField:        true,
		ChangeReasonField:     &change,
		DeletedAtField:        &end,
		DeletedByField:        &delBy,
		DeletionReasonField:   &delReason,
		SuspendedAtField:      &end,
		SuspendedByField:      &suspBy,
		SuspensionReasonField: &suspReason,
		HierarchyDepthField:   5,
		ChildrenCountField:    10,
	}

	if o.RecordId() != "rid" || o.TenantId() != "tid" || o.Code() != "1000100" {
		t.Fatalf("basic id fields mismatch")
	}
	if o.ParentCode() != "1000000" {
		t.Fatalf("parentCode mismatch")
	}
	if o.Name() != "Org" || o.UnitType() != "DEPARTMENT" || o.Status() != "ACTIVE" {
		t.Fatalf("basic string getters mismatch")
	}
	if *o.SortOrder() != 3 {
		t.Fatalf("sortOrder mismatch")
	}
	if o.Description() == nil || *o.Description() != "d" {
		t.Fatalf("description mismatch")
	}
	if o.Profile() == nil || *o.Profile() != "p" {
		t.Fatalf("profile mismatch")
	}
	if o.Level() != 2 {
		t.Fatalf("level clamp mismatch")
	}
	if o.CreatedAt() == "" || o.UpdatedAt() == "" || o.EffectiveDate() == "" {
		t.Fatalf("created/updated/effective date should be formatted")
	}
	if o.Path() == nil || *o.Path() != codePath {
		t.Fatalf("path getter mismatch")
	}
	if o.CodePath() != codePath || o.NamePath() != namePath {
		t.Fatalf("path string getters mismatch")
	}
	if o.IsTemporal() != true {
		t.Fatalf("IsTemporal expected true")
	}
	// EffectiveDate = today -> IsFuture false
	if o.IsFuture() {
		t.Fatalf("IsFuture expected false for today")
	}
	if o.EndDate() == nil || *o.EndDate() == "" {
		t.Fatalf("enddate string expected non-empty")
	}
	if !o.IsCurrent() {
		t.Fatalf("isCurrent expected true")
	}
	if o.HierarchyDepth() != 5 || o.ChildrenCount() != 10 {
		t.Fatalf("hierarchy depth/count mismatch")
	}
	if o.ChangeReason() == nil || *o.ChangeReason() != "reason" {
		t.Fatalf("changeReason mismatch")
	}
	if o.DeletedAt() == nil || *o.DeletedAt() == "" {
		t.Fatalf("deletedAt expected non-empty")
	}
	if o.DeletedBy() == nil || *o.DeletedBy() != delBy {
		t.Fatalf("deletedBy mismatch")
	}
	if o.DeletionReason() == nil || *o.DeletionReason() != delReason {
		t.Fatalf("deletionReason mismatch")
	}
	if o.SuspendedAt() == nil || *o.SuspendedAt() == "" {
		t.Fatalf("suspendedAt expected non-empty")
	}
	if o.SuspendedBy() == nil || *o.SuspendedBy() != suspBy {
		t.Fatalf("suspendedBy mismatch")
	}
	if o.SuspensionReason() == nil || *o.SuspensionReason() != suspReason {
		t.Fatalf("suspensionReason mismatch")
	}
	_ = name // silence unused (document intent)
}

func TestOrganization_PathEmpty(t *testing.T) {
	o := Organization{CodePathField: ""}
	if o.Path() != nil || o.CodePath() != "" {
		t.Fatalf("empty path should return nil/*empty")
	}
}

func TestOrganizationHierarchyData_Getters(t *testing.T) {
	cp := "/100/101"
	np := "/A/B"
	data := OrganizationHierarchyData{
		CodeField:           "100",
		NameField:           "A",
		LevelField:          1,
		HierarchyDepthField: 2,
		CodePathField:       &cp,
		NamePathField:       &np,
		ParentChainField:    []string{"0"},
		ChildrenCountField:  3,
		IsRootField:         true,
		IsLeafField:         false,
		ChildrenField:       []OrganizationHierarchyData{{CodeField: "101"}},
	}
	if data.Code() != "100" || data.Name() != "A" || data.Level() != 1 {
		t.Fatalf("basic fields mismatch")
	}
	if data.CodePath() != cp || data.NamePath() != np {
		t.Fatalf("path fields mismatch")
	}
	if !data.IsRoot() || data.IsLeaf() {
		t.Fatalf("root/leaf mismatch")
	}
	if len(data.Children()) != 1 {
		t.Fatalf("children length mismatch")
	}
}

func TestOrganizationSubtreeData_NilPaths(t *testing.T) {
	data := OrganizationSubtreeData{
		CodeField: "200",
		NameField: "B",
	}
	if data.CodePath() != "" || data.NamePath() != "" {
		t.Fatalf("nil paths should return empty strings")
	}
}

func TestStatsAndCounts_Getters(t *testing.T) {
	stats := OrganizationStats{
		TotalCountField:    10,
		ActiveCountField:   9,
		InactiveCountField: 1,
		PlannedCountField:  0,
		DeletedCountField:  0,
		ByTypeField:        []TypeCount{{UnitTypeField: "DEPARTMENT", CountField: 10}},
		ByStatusField:      []StatusCount{{StatusField: "ACTIVE", CountField: 10}},
		ByLevelField:       []LevelCount{{LevelField: 1, CountField: 10}},
		TemporalStatsField: TemporalStats{TotalVersionsField: 5, AverageVersionsPerOrgField: 1.0, OldestEffectiveDateField: "2025-01-01", NewestEffectiveDateField: "2025-01-02"},
	}
	if stats.TotalCount() != 10 || stats.ActiveCount() != 9 || stats.InactiveCount() != 1 {
		t.Fatalf("stats counts mismatch")
	}
	if stats.TemporalStats().TotalVersions() != 5 {
		t.Fatalf("temporal stats mismatch")
	}
	if stats.PlannedCount() != 0 || stats.DeletedCount() != 0 {
		t.Fatalf("planned/deleted counts mismatch")
	}
	ts := stats.TemporalStats()
	if ts.AverageVersionsPerOrg() != 1.0 || ts.OldestEffectiveDate() != "2025-01-01" || ts.NewestEffectiveDate() != "2025-01-02" {
		t.Fatalf("temporal stats detail getters mismatch")
	}
	if len(stats.ByType()) != 1 || len(stats.ByStatus()) != 1 || len(stats.ByLevel()) != 1 {
		t.Fatalf("stats By* getters mismatch")
	}
	tc := TypeCount{UnitTypeField: "DEPARTMENT", CountField: 2}
	if tc.UnitType() != "DEPARTMENT" || tc.Count() != 2 {
		t.Fatalf("type count getters mismatch")
	}
	lc := LevelCount{LevelField: 3, CountField: 4}
	if lc.Level() != 3 || lc.Count() != 4 {
		t.Fatalf("level count getters mismatch")
	}
	sc := StatusCount{StatusField: "ACTIVE", CountField: 10}
	if sc.Status() != "ACTIVE" || sc.Count() != 10 {
		t.Fatalf("status count getters mismatch")
	}
}

func TestPaginationAndConnection(t *testing.T) {
	p := PaginationInfo{TotalField: 100, PageField: 2, PageSizeField: 10, HasNextField: true, HasPreviousField: false}
	if p.Total() != 100 || p.Page() != 2 || p.PageSize() != 10 || !p.HasNext() || p.HasPrevious() {
		t.Fatalf("pagination getters mismatch")
	}
	edge := PositionEdge{CursorField: "c1", NodeField: Position{CodeField: "P1"}}
	if edge.Cursor() != "c1" || string(edge.Node().Code()) != "P1" {
		t.Fatalf("position edge getters mismatch")
	}
	pc := PositionConnection{
		EdgesField:      []PositionEdge{edge},
		DataField:       []Position{{CodeField: "P1"}},
		PaginationField: p,
		TotalCountField: 1,
	}
	if len(pc.Edges()) != 1 || len(pc.Data()) != 1 || pc.Pagination().Page() != 2 || pc.TotalCount() != 1 {
		t.Fatalf("position connection getters mismatch")
	}
	c := OrganizationConnection{
		DataField:       []Organization{{CodeField: "1000"}},
		PaginationField: p,
		TemporalField:   TemporalInfo{AsOfDateField: "2025-11-15", CurrentCountField: 1},
	}
	if len(c.Data()) != 1 || c.Pagination().Page() != 2 || c.Temporal().CurrentCount() != 1 {
		t.Fatalf("connection getters mismatch")
	}
}

func TestUnmarshalGraphQL_Inputs(t *testing.T) {
	// OrganizationFilter
	of := &OrganizationFilter{}
	err := of.UnmarshalGraphQL(map[string]interface{}{
		"asOfDate":      "2025-11-15",
		"includeFuture": true,
		"onlyFuture":    false,
		"unitType":      "DEPARTMENT",
		"status":        "ACTIVE",
		"parentCode":    "1000000",
		"hasChildren":   true,
		"hasProfile":    false,
	})
	if err != nil {
		t.Fatalf("OrganizationFilter.UnmarshalGraphQL error: %v", err)
	}
	if of.UnitType == nil || *of.UnitType != "DEPARTMENT" || of.HasChildren == nil || *of.HasChildren != true {
		t.Fatalf("OrganizationFilter fields mismatch")
	}

	// PositionFilterInput
	pf := &PositionFilterInput{}
	err = pf.UnmarshalGraphQL(map[string]interface{}{
		"codes":           []interface{}{"P1", "P2"},
		"organization":    "1000000",
		"positionTypes":   []interface{}{"FULLTIME"},
		"employmentTypes": []interface{}{"REGULAR"},
		"status":          "ACTIVE",
		"asOfDate":        "2025-11-15",
	})
	if err != nil {
		t.Fatalf("PositionFilterInput.UnmarshalGraphQL error: %v", err)
	}
	if pf.Status == nil || *pf.Status != "ACTIVE" {
		t.Fatalf("PositionFilterInput status mismatch")
	}

	// PositionSortInput
	ps := &PositionSortInput{}
	err = ps.UnmarshalGraphQL(map[string]interface{}{"field": "code", "direction": "ASC"})
	if err != nil || ps.Field != "code" || ps.Direction != "ASC" {
		t.Fatalf("PositionSortInput.UnmarshalGraphQL mismatch: %#v, err=%v", ps, err)
	}

	// VacantPositionFilterInput
	vf := &VacantPositionFilterInput{}
	err = vf.UnmarshalGraphQL(map[string]interface{}{
		"organizationCodes": []interface{}{"1000000"},
		"jobFamilyCodes":    []interface{}{"JF1"},
		"jobRoleCodes":      []interface{}{"JR1"},
		"jobLevelCodes":     []interface{}{"J1"},
		"positionTypes":     []interface{}{"FULLTIME"},
		"minimumVacantDays": 10,
		"asOfDate":          "2025-11-15",
	})
	if err != nil {
		t.Fatalf("VacantPositionFilterInput.UnmarshalGraphQL error: %v", err)
	}
	if vf.MinimumVacantDays == nil || *vf.MinimumVacantDays != 10 {
		t.Fatalf("VacantPositionFilterInput minimumVacantDays mismatch")
	}

	// VacantPositionSortInput
	vs := &VacantPositionSortInput{}
	err = vs.UnmarshalGraphQL(map[string]interface{}{"field": "code", "direction": "DESC"})
	if err != nil || vs.Field != "code" || vs.Direction != "DESC" {
		t.Fatalf("VacantPositionSortInput.UnmarshalGraphQL mismatch: %#v, err=%v", vs, err)
	}

	// PositionAssignmentFilterInput
	paf := &PositionAssignmentFilterInput{}
	err = paf.UnmarshalGraphQL(map[string]interface{}{
		"employeeId":        "emp1",
		"status":            "ACTIVE",
		"assignmentTypes":   []interface{}{"PRIMARY"},
		"includeHistorical": true,
	})
	if err != nil {
		t.Fatalf("PositionAssignmentFilterInput.UnmarshalGraphQL error: %v", err)
	}
	if paf.EmployeeID == nil || *paf.EmployeeID != "emp1" || paf.IncludeHistorical != true {
		t.Fatalf("PositionAssignmentFilterInput fields mismatch")
	}

	// PositionAssignmentSortInput
	pas := &PositionAssignmentSortInput{}
	err = pas.UnmarshalGraphQL(map[string]interface{}{"field": "effectiveDate", "direction": "ASC"})
	if err != nil || pas.Field != "effectiveDate" || pas.Direction != "ASC" {
		t.Fatalf("PositionAssignmentSortInput.UnmarshalGraphQL mismatch: %#v, err=%v", pas, err)
	}
}

func TestVacantAndTransferGetters(t *testing.T) {
	// VacantPosition getters + edges/connection
	vacDate := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	orgName := "OrgA"
	v := VacantPosition{
		PositionCodeField:       "P-V-1",
		OrganizationCodeField:   "1000000",
		OrganizationNameField:   &orgName,
		JobFamilyCodeField:      "JF",
		JobRoleCodeField:        "JR",
		JobLevelCodeField:       "JL",
		VacantSinceField:        vacDate,
		HeadcountCapacityField:  2,
		HeadcountAvailableField: 1,
		TotalAssignmentsField:   0,
	}
	if string(v.PositionCode()) != "P-V-1" || v.OrganizationCode() != "1000000" {
		t.Fatalf("VacantPosition basic getters mismatch")
	}
	if v.OrganizationName() == nil || *v.OrganizationName() != "OrgA" {
		t.Fatalf("VacantPosition organizationName mismatch")
	}
	if string(v.JobFamilyCode()) != "JF" || string(v.JobRoleCode()) != "JR" || string(v.JobLevelCode()) != "JL" {
		t.Fatalf("VacantPosition job codes mismatch")
	}
	if v.VacantSince() == "" || v.HeadcountCapacity() != 2 || v.HeadcountAvailable() != 1 || v.TotalAssignments() != 0 {
		t.Fatalf("VacantPosition numeric/date getters mismatch")
	}
	ve := VacantPositionEdge{CursorField: "vc1", NodeField: v}
	if ve.Cursor() != "vc1" || string(ve.Node().PositionCode()) != "P-V-1" {
		t.Fatalf("VacantPositionEdge mismatch")
	}
	vc := VacantPositionConnection{
		EdgesField:      []VacantPositionEdge{ve},
		DataField:       []VacantPosition{v},
		PaginationField: PaginationInfo{TotalField: 1},
		TotalCountField: 1,
	}
	if len(vc.Edges()) != 1 || len(vc.Data()) != 1 || vc.TotalCount() != 1 {
		t.Fatalf("VacantPositionConnection mismatch")
	}
}

func TestPositionAssignmentAudit_Getters(t *testing.T) {
	now := time.Now().UTC()
	end := now.Add(24 * time.Hour)
	changes := map[string]interface{}{"field": "value"}
	a := PositionAssignmentAudit{
		AssignmentIDField:  "A1",
		EventTypeField:     "CREATED",
		EffectiveDateField: now,
		EndDateField:       &end,
		ActorField:         "dev",
		ChangesField:       changes,
		CreatedAtField:     now,
	}
	if string(a.AssignmentId()) != "A1" || a.EventType() != "CREATED" {
		t.Fatalf("audit basic getters mismatch")
	}
	_ = a.EffectiveDate()
	if a.EndDate() == nil || a.Actor() != "dev" {
		t.Fatalf("audit endDate/actor mismatch")
	}
	if a.Changes() == nil || (*a.Changes())["field"] != "value" {
		t.Fatalf("audit changes mismatch")
	}
	if a.CreatedAt() == "" {
		t.Fatalf("audit createdAt mismatch")
	}
	conn := PositionAssignmentAuditConnection{
		DataField:       []PositionAssignmentAudit{a},
		PaginationField: PaginationInfo{TotalField: 1},
		TotalCountField: 1,
	}
	if len(conn.Data()) != 1 || conn.Pagination().Total() != 1 || conn.TotalCount() != 1 {
		t.Fatalf("audit connection getters mismatch")
	}
}
