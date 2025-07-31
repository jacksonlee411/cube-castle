package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// Test structures matching the API
type CreateEmployeeRequest struct {
	EmployeeType     string    `json:"employee_type"`
	EmployeeNumber   string    `json:"employee_number"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	Email            string    `json:"email"`
	EmploymentStatus string    `json:"employment_status"`
	HireDate         time.Time `json:"hire_date"`
}

type EmployeeResponse struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenant_id"`
	EmployeeType     string    `json:"employee_type"`
	EmployeeNumber   string    `json:"employee_number"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	FullName         string    `json:"full_name"`
	Email            string    `json:"email"`
	EmploymentStatus string    `json:"employment_status"`
	HireDate         time.Time `json:"hire_date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ListEmployeesResponse struct {
	Data   []EmployeeResponse `json:"data"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
	Total  int                `json:"total"`
}

const baseURL = "http://localhost:8080/api/v1/employees"

func main() {
	fmt.Println("🧪 Testing Employee API endpoints...")

	// Test 1: Create a new employee
	fmt.Println("\n1. Testing CREATE Employee...")
	employee := CreateEmployeeRequest{
		EmployeeType:     "FULL_TIME",
		EmployeeNumber:   "EMP001",
		FirstName:        "John",
		LastName:         "Doe",
		Email:            "john.doe@example.com",
		EmploymentStatus: "ACTIVE",
		HireDate:         time.Now(),
	}

	createdEmployee, err := createEmployee(employee)
	if err != nil {
		log.Printf("❌ Create employee failed: %v", err)
	} else {
		fmt.Printf("✅ Employee created successfully: %s %s (ID: %s)\n", 
			createdEmployee.FirstName, createdEmployee.LastName, createdEmployee.ID)
	}

	// Test 2: List employees
	fmt.Println("\n2. Testing LIST Employees...")
	employees, err := listEmployees()
	if err != nil {
		log.Printf("❌ List employees failed: %v", err)
	} else {
		fmt.Printf("✅ Found %d employees\n", employees.Total)
		for _, emp := range employees.Data {
			fmt.Printf("   - %s (%s)\n", emp.FullName, emp.EmployeeNumber)
		}
	}

	// Test 3: Get specific employee (if we created one)
	if createdEmployee != nil {
		fmt.Println("\n3. Testing GET Employee...")
		fetchedEmployee, err := getEmployee(createdEmployee.ID.String())
		if err != nil {
			log.Printf("❌ Get employee failed: %v", err)
		} else {
			fmt.Printf("✅ Employee retrieved: %s %s (Status: %s)\n", 
				fetchedEmployee.FirstName, fetchedEmployee.LastName, fetchedEmployee.EmploymentStatus)
		}
	}

	// Test 4: Search employees
	fmt.Println("\n4. Testing SEARCH Employees...")
	searchResults, err := searchEmployees("john")
	if err != nil {
		log.Printf("❌ Search employees failed: %v", err)
	} else {
		fmt.Printf("✅ Search found %d employees\n", searchResults.Total)
	}

	fmt.Println("\n🎉 Employee API testing completed!")
}

func createEmployee(req CreateEmployeeRequest) (*EmployeeResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var employee EmployeeResponse
	if err := json.NewDecoder(resp.Body).Decode(&employee); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &employee, nil
}

func listEmployees() (*ListEmployeesResponse, error) {
	resp, err := http.Get(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var employees ListEmployeesResponse
	if err := json.NewDecoder(resp.Body).Decode(&employees); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &employees, nil
}

func getEmployee(id string) (*EmployeeResponse, error) {
	resp, err := http.Get(baseURL + "/" + id)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var employee EmployeeResponse
	if err := json.NewDecoder(resp.Body).Decode(&employee); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &employee, nil
}

func searchEmployees(query string) (*ListEmployeesResponse, error) {
	url := fmt.Sprintf("%s?search=%s", baseURL, query)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var employees ListEmployeesResponse
	if err := json.NewDecoder(resp.Body).Decode(&employees); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &employees, nil
}