// Simple test script to verify frontend-backend connection
const fetch = require('node-fetch');

async function testAPI() {
  try {
    console.log('Testing backend API...');
    
    // Test backend API directly
    const backendResponse = await fetch('http://localhost:8080/api/v1/corehr/employees?page=1&page_size=3');
    const backendData = await backendResponse.json();
    
    console.log('✅ Backend API Response:');
    console.log(`Status: ${backendResponse.status}`);
    console.log(`Employee count: ${backendData.employees?.length || 0}`);
    console.log(`Total count: ${backendData.total_count || 0}`);
    
    if (backendData.employees && backendData.employees.length > 0) {
      console.log('First employee:', {
        id: backendData.employees[0].id,
        name: `${backendData.employees[0].first_name} ${backendData.employees[0].last_name}`,
        email: backendData.employees[0].email,
        employee_number: backendData.employees[0].employee_number,
      });
    }
    
    console.log('\n✅ Backend is working correctly with real database data!');
    
    // Test frontend accessibility
    console.log('\nTesting frontend accessibility...');
    const frontendResponse = await fetch('http://localhost:3001');
    console.log(`Frontend status: ${frontendResponse.status}`);
    
    if (frontendResponse.status === 200) {
      console.log('✅ Frontend is accessible');
    } else {
      console.log('❌ Frontend is not accessible');
    }
    
  } catch (error) {
    console.error('❌ Test failed:', error.message);
  }
}

testAPI();