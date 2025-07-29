// test/fixtures/test_data.go
package fixtures

import (
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// TestDataBuilder provides builder pattern for creating test data
type TestDataBuilder struct {
	employees []EmployeeFixture
	positions []PositionFixture
}

// EmployeeFixture represents test employee data
type EmployeeFixture struct {
	ID             uuid.UUID
	EmployeeID     string
	LegalName      string
	PreferredName  *string
	Email          string
	Status         string
	HireDate       time.Time
	TerminationDate *time.Time
}

// PositionFixture represents test position data
type PositionFixture struct {
	ID                  uuid.UUID
	EmployeeID          uuid.UUID
	PositionTitle       string
	Department          string
	JobLevel            string
	Location            *string
	EmploymentType      string
	ReportsToEmployeeID *uuid.UUID
	EffectiveDate       time.Time
	EndDate             *time.Time
	ChangeReason        string
	IsRetroactive       bool
	MinSalary           *float64
	MaxSalary           *float64
	Currency            *string
}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder() *TestDataBuilder {
	return &TestDataBuilder{
		employees: make([]EmployeeFixture, 0),
		positions: make([]PositionFixture, 0),
	}
}

// Standard test data sets

// GetStandardEmployees returns a set of standard test employees
func (b *TestDataBuilder) GetStandardEmployees() *TestDataBuilder {
	employees := []EmployeeFixture{
		{
			ID:         uuid.New(),
			EmployeeID: "EMP001",
			LegalName:  "张三",
			Email:      "zhang.san@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:         uuid.New(),
			EmployeeID: "EMP002",
			LegalName:  "李四",
			Email:      "li.si@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2019, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:         uuid.New(),
			EmployeeID: "EMP003",
			LegalName:  "王五",
			Email:      "wang.wu@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:         uuid.New(),
			EmployeeID: "EMP004",
			LegalName:  "赵六",
			Email:      "zhao.liu@company.com",
			Status:     "TERMINATED",
			HireDate:   time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC),
			TerminationDate: &[]time.Time{time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)}[0],
		},
		{
			ID:         uuid.New(),
			EmployeeID: "EMP005",
			LegalName:  "孙七",
			Email:      "sun.qi@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2017, 4, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	
	b.employees = employees
	return b
}

// GetStandardPositions returns standard position data for the employees
func (b *TestDataBuilder) GetStandardPositions() *TestDataBuilder {
	if len(b.employees) == 0 {
		b.GetStandardEmployees()
	}
	
	positions := []PositionFixture{
		// EMP001 - 张三 的职位历史
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[0].ID,
			PositionTitle:  "软件工程师",
			Department:     "技术部",
			JobLevel:       "INTERMEDIATE",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        &[]time.Time{time.Date(2021, 6, 30, 23, 59, 59, 0, time.UTC)}[0],
			ChangeReason:   "入职",
			IsRetroactive:  false,
			MinSalary:      &[]float64{8000}[0],
			MaxSalary:      &[]float64{12000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[0].ID,
			PositionTitle:  "高级软件工程师",
			Department:     "技术部",
			JobLevel:       "SENIOR",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2021, 7, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:   "晋升",
			IsRetroactive:  false,
			MinSalary:      &[]float64{12000}[0],
			MaxSalary:      &[]float64{18000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		
		// EMP002 - 李四 的职位历史
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[1].ID,
			PositionTitle:  "产品经理",
			Department:     "产品部",
			JobLevel:       "INTERMEDIATE",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2019, 6, 15, 0, 0, 0, 0, time.UTC),
			EndDate:        &[]time.Time{time.Date(2022, 3, 31, 23, 59, 59, 0, time.UTC)}[0],
			ChangeReason:   "入职",
			IsRetroactive:  false,
			MinSalary:      &[]float64{10000}[0],
			MaxSalary:      &[]float64{15000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[1].ID,
			PositionTitle:  "高级产品经理",
			Department:     "产品部",
			JobLevel:       "SENIOR",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2022, 4, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:   "晋升",
			IsRetroactive:  false,
			MinSalary:      &[]float64{15000}[0],
			MaxSalary:      &[]float64{22000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		
		// EMP003 - 王五 的职位历史
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[2].ID,
			PositionTitle:  "UI设计师",
			Department:     "设计部",
			JobLevel:       "INTERMEDIATE",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2021, 3, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:   "入职",
			IsRetroactive:  false,
			MinSalary:      &[]float64{9000}[0],
			MaxSalary:      &[]float64{14000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		
		// EMP004 - 赵六 的职位历史（已离职）
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[3].ID,
			PositionTitle:  "人事专员",
			Department:     "人力资源部",
			JobLevel:       "JUNIOR",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2018, 9, 1, 0, 0, 0, 0, time.UTC),
			EndDate:        &[]time.Time{time.Date(2023, 12, 31, 23, 59, 59, 0, time.UTC)}[0],
			ChangeReason:   "入职",
			IsRetroactive:  false,
			MinSalary:      &[]float64{6000}[0],
			MaxSalary:      &[]float64{9000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
		
		// EMP005 - 孙七 的职位历史（管理层）
		{
			ID:                  uuid.New(),
			EmployeeID:          b.employees[4].ID,
			PositionTitle:       "软件架构师",
			Department:          "技术部",
			JobLevel:            "SENIOR",
			EmploymentType:      "FULL_TIME",
			EffectiveDate:       time.Date(2017, 4, 1, 0, 0, 0, 0, time.UTC),
			EndDate:             &[]time.Time{time.Date(2020, 12, 31, 23, 59, 59, 0, time.UTC)}[0],
			ChangeReason:        "入职",
			IsRetroactive:       false,
			MinSalary:           &[]float64{18000}[0],
			MaxSalary:           &[]float64{25000}[0],
			Currency:            &[]string{"CNY"}[0],
		},
		{
			ID:             uuid.New(),
			EmployeeID:     b.employees[4].ID,
			PositionTitle:  "技术总监",
			Department:     "技术部",
			JobLevel:       "MANAGER",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:   "晋升",
			IsRetroactive:  false,
			MinSalary:      &[]float64{25000}[0],
			MaxSalary:      &[]float64{35000}[0],
			Currency:       &[]string{"CNY"}[0],
		},
	}
	
	b.positions = positions
	return b
}

// GetComplexHierarchy returns test data with complex reporting relationships
func (b *TestDataBuilder) GetComplexHierarchy() *TestDataBuilder {
	// Create a more complex organizational hierarchy
	managerID := uuid.New()
	directorID := uuid.New()
	
	employees := []EmployeeFixture{
		// CEO
		{
			ID:         directorID,
			EmployeeID: "CEO001",
			LegalName:  "首席执行官",
			Email:      "ceo@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		// Director
		{
			ID:         managerID,
			EmployeeID: "DIR001",
			LegalName:  "技术总监",
			Email:      "director@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		// Managers
		{
			ID:         uuid.New(),
			EmployeeID: "MGR001",
			LegalName:  "后端团队经理",
			Email:      "backend.manager@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2017, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:         uuid.New(),
			EmployeeID: "MGR002",
			LegalName:  "前端团队经理",
			Email:      "frontend.manager@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2018, 3, 1, 0, 0, 0, 0, time.UTC),
		},
		// Team Members
		{
			ID:         uuid.New(),
			EmployeeID: "DEV001",
			LegalName:  "后端开发工程师",
			Email:      "backend.dev@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2019, 9, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			ID:         uuid.New(),
			EmployeeID: "DEV002",
			LegalName:  "前端开发工程师",
			Email:      "frontend.dev@company.com",
			Status:     "ACTIVE",
			HireDate:   time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	
	positions := []PositionFixture{
		// CEO Position
		{
			ID:             uuid.New(),
			EmployeeID:     employees[0].ID,
			PositionTitle:  "首席执行官",
			Department:     "管理层",
			JobLevel:       "EXECUTIVE",
			EmploymentType: "FULL_TIME",
			EffectiveDate:  time.Date(2015, 1, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:   "入职",
			IsRetroactive:  false,
		},
		// Director Position
		{
			ID:                  uuid.New(),
			EmployeeID:          employees[1].ID,
			PositionTitle:       "技术总监",
			Department:          "技术部",
			JobLevel:            "MANAGER",
			EmploymentType:      "FULL_TIME",
			ReportsToEmployeeID: &employees[0].ID,
			EffectiveDate:       time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:        "入职",
			IsRetroactive:       false,
		},
		// Backend Manager Position
		{
			ID:                  uuid.New(),
			EmployeeID:          employees[2].ID,
			PositionTitle:       "后端团队经理",
			Department:          "技术部",
			JobLevel:            "MANAGER",
			EmploymentType:      "FULL_TIME",
			ReportsToEmployeeID: &employees[1].ID,
			EffectiveDate:       time.Date(2017, 6, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:        "入职",
			IsRetroactive:       false,
		},
		// Frontend Manager Position
		{
			ID:                  uuid.New(),
			EmployeeID:          employees[3].ID,
			PositionTitle:       "前端团队经理",
			Department:          "技术部",
			JobLevel:            "MANAGER",
			EmploymentType:      "FULL_TIME",
			ReportsToEmployeeID: &employees[1].ID,
			EffectiveDate:       time.Date(2018, 3, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:        "入职",
			IsRetroactive:       false,
		},
		// Backend Developer Position
		{
			ID:                  uuid.New(),
			EmployeeID:          employees[4].ID,
			PositionTitle:       "后端开发工程师",
			Department:          "技术部",
			JobLevel:            "SENIOR",
			EmploymentType:      "FULL_TIME",
			ReportsToEmployeeID: &employees[2].ID,
			EffectiveDate:       time.Date(2019, 9, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:        "入职",
			IsRetroactive:       false,
		},
		// Frontend Developer Position
		{
			ID:                  uuid.New(),
			EmployeeID:          employees[5].ID,
			PositionTitle:       "前端开发工程师",
			Department:          "技术部",
			JobLevel:            "INTERMEDIATE",
			EmploymentType:      "FULL_TIME",
			ReportsToEmployeeID: &employees[3].ID,
			EffectiveDate:       time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC),
			ChangeReason:        "入职",
			IsRetroactive:       false,
		},
	}
	
	b.employees = employees
	b.positions = positions
	return b
}

// GetPerformanceTestData returns large dataset for performance testing
func (b *TestDataBuilder) GetPerformanceTestData(employeeCount int) *TestDataBuilder {
	employees := make([]EmployeeFixture, employeeCount)
	positions := make([]PositionFixture, 0, employeeCount*2) // Average 2 positions per employee
	
	departments := []string{"技术部", "产品部", "设计部", "市场部", "销售部", "人力资源部", "财务部"}
	jobLevels := []string{"JUNIOR", "INTERMEDIATE", "SENIOR", "MANAGER"}
	positionTitles := map[string][]string{
		"技术部":    {"软件工程师", "高级工程师", "架构师", "技术总监"},
		"产品部":    {"产品专员", "产品经理", "高级产品经理", "产品总监"},
		"设计部":    {"UI设计师", "UX设计师", "设计总监"},
		"市场部":    {"市场专员", "市场经理", "市场总监"},
		"销售部":    {"销售代表", "销售经理", "销售总监"},
		"人力资源部": {"人事专员", "招聘经理", "HR总监"},
		"财务部":    {"会计", "财务经理", "财务总监"},
	}
	
	for i := 0; i < employeeCount; i++ {
		empID := uuid.New()
		hireDate := time.Date(2018+i%6, 1+i%12, 1, 0, 0, 0, 0, time.UTC)
		
		employees[i] = EmployeeFixture{
			ID:         empID,
			EmployeeID: fmt.Sprintf("PERF%05d", i+1),
			LegalName:  fmt.Sprintf("性能测试员工%d", i+1),
			Email:      fmt.Sprintf("perf%d@company.com", i+1),
			Status:     "ACTIVE",
			HireDate:   hireDate,
		}
		
		// Create 1-3 positions per employee
		positionCount := 1 + i%3
		dept := departments[i%len(departments)]
		titles := positionTitles[dept]
		
		for j := 0; j < positionCount; j++ {
			effectiveDate := hireDate.AddDate(j*2, 0, 0) // Every 2 years
			var endDate *time.Time
			if j < positionCount-1 {
				end := effectiveDate.AddDate(2, 0, -1)
				endDate = &end
			}
			
			positions = append(positions, PositionFixture{
				ID:             uuid.New(),
				EmployeeID:     empID,
				PositionTitle:  titles[j%len(titles)],
				Department:     dept,
				JobLevel:       jobLevels[j%len(jobLevels)],
				EmploymentType: "FULL_TIME",
				EffectiveDate:  effectiveDate,
				EndDate:        endDate,
				ChangeReason:   map[int]string{0: "入职", 1: "晋升", 2: "调岗"}[j%3],
				IsRetroactive:  false,
				MinSalary:      &[]float64{float64(6000 + j*3000 + i%5000)}[0],
				MaxSalary:      &[]float64{float64(9000 + j*4500 + i%7000)}[0],
				Currency:       &[]string{"CNY"}[0],
			})
		}
	}
	
	b.employees = employees
	b.positions = positions
	return b
}

// CreateInDatabase persists the test data to the database
func (b *TestDataBuilder) CreateInDatabase(entClient *ent.Client) error {
	ctx := context.Background()
	
	// Create employees first
	employeeMap := make(map[uuid.UUID]*ent.Employee)
	for _, empFixture := range b.employees {
		employee := entClient.Employee.Create().
			SetID(empFixture.ID).
			SetEmployeeID(empFixture.EmployeeID).
			SetLegalName(empFixture.LegalName).
			SetEmail(empFixture.Email).
			SetStatus(empFixture.Status).
			SetHireDate(empFixture.HireDate)
		
		if empFixture.PreferredName != nil {
			employee = employee.SetPreferredName(*empFixture.PreferredName)
		}
		if empFixture.TerminationDate != nil {
			employee = employee.SetTerminationDate(*empFixture.TerminationDate)
		}
		
		emp, err := employee.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create employee %s: %w", empFixture.EmployeeID, err)
		}
		employeeMap[empFixture.ID] = emp
	}
	
	// Create positions
	for _, posFixture := range b.positions {
		position := entClient.PositionHistory.Create().
			SetID(posFixture.ID).
			SetEmployeeID(posFixture.EmployeeID).
			SetPositionTitle(posFixture.PositionTitle).
			SetDepartment(posFixture.Department).
			SetJobLevel(posFixture.JobLevel).
			SetEmploymentType(posFixture.EmploymentType).
			SetEffectiveDate(posFixture.EffectiveDate).
			SetChangeReason(posFixture.ChangeReason).
			SetIsRetroactive(posFixture.IsRetroactive)
		
		if posFixture.Location != nil {
			position = position.SetLocation(*posFixture.Location)
		}
		if posFixture.ReportsToEmployeeID != nil {
			position = position.SetReportsToEmployeeID(*posFixture.ReportsToEmployeeID)
		}
		if posFixture.EndDate != nil {
			position = position.SetEndDate(*posFixture.EndDate)
		}
		if posFixture.MinSalary != nil {
			position = position.SetMinSalary(*posFixture.MinSalary)
		}
		if posFixture.MaxSalary != nil {
			position = position.SetMaxSalary(*posFixture.MaxSalary)
		}
		if posFixture.Currency != nil {
			position = position.SetCurrency(*posFixture.Currency)
		}
		
		_, err := position.Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create position for employee %v: %w", posFixture.EmployeeID, err)
		}
	}
	
	return nil
}

// GetEmployees returns the employee fixtures
func (b *TestDataBuilder) GetEmployees() []EmployeeFixture {
	return b.employees
}

// GetPositions returns the position fixtures
func (b *TestDataBuilder) GetPositions() []PositionFixture {
	return b.positions
}

// Quick access methods for common test scenarios

// QuickStandardDataset returns a builder with standard test dataset
func QuickStandardDataset() *TestDataBuilder {
	return NewTestDataBuilder().GetStandardEmployees().GetStandardPositions()
}

// QuickHierarchyDataset returns a builder with hierarchical test dataset
func QuickHierarchyDataset() *TestDataBuilder {
	return NewTestDataBuilder().GetComplexHierarchy()
}

// QuickPerformanceDataset returns a builder with performance test dataset
func QuickPerformanceDataset(size int) *TestDataBuilder {
	return NewTestDataBuilder().GetPerformanceTestData(size)
}