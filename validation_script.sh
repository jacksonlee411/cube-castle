#!/bin/bash

echo "🔍 Employee-Organization-Position System Validation"
echo "=================================================="

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track validation results
PASS_COUNT=0
FAIL_COUNT=0

check_result() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ PASS${NC}: $1"
        PASS_COUNT=$((PASS_COUNT + 1))
    else
        echo -e "${RED}❌ FAIL${NC}: $1"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

echo -e "\n${YELLOW}Phase 1: Build Validation${NC}"
echo "--------------------------------"

# Test 1: Server compilation
echo "Testing server compilation..."
go build -o build/server ./cmd/server
check_result "Server builds without errors"

# Test 2: Ent code generation
echo "Testing Ent code generation..."
go generate ./ent > /dev/null 2>&1
check_result "Ent schemas generate successfully"

# Test 3: Migration script compilation
echo "Testing migration script compilation..."
go build -o build/migrate ./cmd/migrate/employee_migration.go
check_result "Migration script compiles"

# Test 4: Test program compilation
echo "Testing API test program compilation..."
go build -o build/test_employee_api ./test_employee_api.go
check_result "API test program compiles"

echo -e "\n${YELLOW}Phase 2: Schema Validation${NC}"
echo "--------------------------------"

# Test 5: Employee schema validation
echo "Validating Employee schema structure..."
if go run -tags=ignore ./cmd/schema/main.go employee > /dev/null 2>&1; then
    check_result "Employee schema structure is valid"
else
    check_result "Employee schema structure validation"
fi

# Test 6: Position schema validation
echo "Validating Position schema structure..."
if go run -tags=ignore ./cmd/schema/main.go position > /dev/null 2>&1; then
    check_result "Position schema structure is valid"
else
    check_result "Position schema structure validation"
fi

echo -e "\n${YELLOW}Phase 3: Handler Validation${NC}"
echo "--------------------------------"

# Test 7: Employee handler structure
echo "Validating Employee handler structure..."
if grep -q "func (h \*EmployeeHandler) CreateEmployee" internal/handler/employee_handler.go && \
   grep -q "func (h \*EmployeeHandler) GetEmployee" internal/handler/employee_handler.go && \
   grep -q "func (h \*EmployeeHandler) ListEmployees" internal/handler/employee_handler.go && \
   grep -q "func (h \*EmployeeHandler) UpdateEmployee" internal/handler/employee_handler.go && \
   grep -q "func (h \*EmployeeHandler) DeleteEmployee" internal/handler/employee_handler.go; then
    check_result "Employee handler has all CRUD methods"
else
    check_result "Employee handler CRUD methods"
fi

# Test 8: Position assignment functionality
echo "Validating position assignment functionality..."
if grep -q "func (h \*EmployeeHandler) AssignPosition" internal/handler/employee_handler.go && \
   grep -q "func (h \*EmployeeHandler) GetPositionHistory" internal/handler/employee_handler.go; then
    check_result "Employee handler has position assignment methods"
else
    check_result "Employee handler position assignment methods"
fi

echo -e "\n${YELLOW}Phase 4: API Integration Validation${NC}"
echo "-------------------------------------------"

# Test 9: API routes integration
echo "Validating API route integration..."
if grep -q "employeeHandler = handler.NewEmployeeHandler" cmd/server/main.go && \
   grep -q "/employees.*employeeHandler.ListEmployees" cmd/server/main.go && \
   grep -q "/employees.*employeeHandler.CreateEmployee" cmd/server/main.go; then
    check_result "Employee API routes are integrated"
else
    check_result "Employee API routes integration"
fi

echo -e "\n${YELLOW}Phase 5: Relationship Validation${NC}"
echo "--------------------------------------------"

# Test 10: Employee-Position relationship
echo "Validating Employee-Position relationship..."
if grep -q "current_position_id.*UUID" go-app/ent/schema/employee.go && \
   grep -q "current_incumbents.*Employee" go-app/ent/schema/position.go; then
    check_result "Employee-Position relationship is established"
else
    check_result "Employee-Position relationship"
fi

# Test 11: Position occupancy history relationship
echo "Validating Position occupancy history relationship..."
if grep -q "employee.*Employee.Type" go-app/ent/schema/position_occupancy_history.go && \
   grep -q "position_history.*PositionOccupancyHistory" go-app/ent/schema/employee.go; then
    check_result "Position occupancy history relationship is activated"
else
    check_result "Position occupancy history relationship"
fi

echo -e "\n${YELLOW}Summary${NC}"
echo "========"
echo -e "Total tests: $((PASS_COUNT + FAIL_COUNT))"
echo -e "${GREEN}Passed: ${PASS_COUNT}${NC}"
echo -e "${RED}Failed: ${FAIL_COUNT}${NC}"

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "\n🎉 ${GREEN}All validations passed!${NC} Employee-Organization-Position system is ready."
    exit 0
else
    echo -e "\n⚠️  ${YELLOW}Some validations failed.${NC} Please review the issues above."
    exit 1
fi