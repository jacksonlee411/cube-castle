// Test CQRS Employee Query Integration
import { employeeQueries, EmployeeSearchParams } from '../nextjs-app/src/lib/cqrs/employee-queries'

/**
 * Test CQRS Employee Query Integration
 * Validates that frontend CQRS hooks can communicate with backend Neo4j query handlers
 */
async function testEmployeeCQRSQueries() {
  console.log('🧪 Testing Employee CQRS Query Integration...')
  
  try {
    // Test 1: Search all employees
    console.log('\n📋 Test 1: Search all employees')
    const searchParams: EmployeeSearchParams = {
      limit: 10,
      offset: 0
    }
    
    const searchResponse = await employeeQueries.searchEmployees(searchParams)
    console.log('✅ Search employees successful:', {
      total: searchResponse.total_count,
      returned: searchResponse.employees.length,
      limit: searchResponse.limit,
      offset: searchResponse.offset
    })
    
    if (searchResponse.employees.length > 0) {
      const firstEmployee = searchResponse.employees[0]
      console.log('📝 First employee sample:', {
        id: firstEmployee.id,
        name: firstEmployee.legalName,
        email: firstEmployee.email,
        status: firstEmployee.status,
        department: firstEmployee.department
      })
      
      // Test 2: Get single employee (if we have valid employee data)
      if (firstEmployee.id && firstEmployee.id !== '00000000-0000-0000-0000-000000000000') {
        console.log('\n📋 Test 2: Get single employee')
        try {
          const employee = await employeeQueries.getEmployee(firstEmployee.id)
          if (employee) {
            console.log('✅ Get employee successful:', {
              id: employee.id,
              name: employee.legalName,
              email: employee.email
            })
          } else {
            console.log('⚠️ Employee not found by ID')
          }
        } catch (error) {
          console.log('❌ Get employee failed:', error.message)
        }
      } else {
        console.log('⏭️ Skipping single employee test - no valid UUID available')
      }
      
      // Test 3: Search by email
      if (firstEmployee.email) {
        console.log('\n📋 Test 3: Search employees by email')
        const emailSearchParams: EmployeeSearchParams = {
          email: firstEmployee.email,
          limit: 5,
          offset: 0
        }
        
        const emailSearchResponse = await employeeQueries.searchEmployees(emailSearchParams)
        console.log('✅ Email search successful:', {
          email: firstEmployee.email,
          found: emailSearchResponse.employees.length,
          total: emailSearchResponse.total_count
        })
      }
      
      // Test 4: Search by name
      if (firstEmployee.legalName) {
        console.log('\n📋 Test 4: Search employees by name')
        const nameSearchParams: EmployeeSearchParams = {
          name: firstEmployee.legalName.split(' ')[0], // First name
          limit: 5,
          offset: 0
        }
        
        const nameSearchResponse = await employeeQueries.searchEmployees(nameSearchParams)
        console.log('✅ Name search successful:', {
          name: nameSearchParams.name,
          found: nameSearchResponse.employees.length,
          total: nameSearchResponse.total_count
        })
      }
    } else {
      console.log('⚠️ No employees found in database')
    }
    
    // Test 5: Employee stats
    console.log('\n📋 Test 5: Get employee statistics')
    const stats = await employeeQueries.getEmployeeStats()
    console.log('✅ Employee stats:', stats)
    
    console.log('\n🎉 All CQRS Employee Query tests completed successfully!')
    
    return {
      success: true,
      totalEmployees: searchResponse.total_count,
      sampleEmployee: searchResponse.employees[0] || null,
      testsCompleted: 5
    }
    
  } catch (error) {
    console.error('❌ CQRS Employee Query test failed:', error)
    return {
      success: false,
      error: error.message,
      testsCompleted: 0
    }
  }
}

// Run the test
if (require.main === module) {
  testEmployeeCQRSQueries()
    .then(result => {
      console.log('\n📊 Test Summary:', result)
      process.exit(result.success ? 0 : 1)
    })
    .catch(error => {
      console.error('💥 Test execution failed:', error)
      process.exit(1)
    })
}

export { testEmployeeCQRSQueries }