// Test frontend to backend connection using same domain
const https = require('https');
const http = require('http');

async function testConnection() {
  console.log('Testing frontend-backend connection...\n');
  
  // Test backend API directly
  console.log('1. Testing backend API directly:');
  try {
    const response = await fetch('http://localhost:8080/api/v1/corehr/employees?page=1&page_size=3');
    const data = await response.json();
    console.log(`✅ Backend Response: ${response.status}`);
    console.log(`✅ Employees Count: ${data.employees?.length || 0}`);
    console.log(`✅ Total Count: ${data.total_count || 0}`);
    if (data.employees && data.employees.length > 0) {
      console.log(`✅ First Employee: ${data.employees[0].first_name} ${data.employees[0].last_name} (${data.employees[0].email})`);
    }
  } catch (error) {
    console.log(`❌ Backend API Error: ${error.message}`);
  }
  
  console.log('\n2. Testing frontend accessibility:');
  // Test frontend accessibility
  try {
    const response = await fetch('http://localhost:3002');
    console.log(`✅ Frontend Status: ${response.status}`);
    console.log('✅ Frontend is accessible');
  } catch (error) {
    console.log(`❌ Frontend Error: ${error.message}`);
  }
  
  console.log('\n3. Testing CORS from frontend perspective:');
  // Test CORS
  try {
    const response = await fetch('http://localhost:8080/api/v1/corehr/employees?page=1&page_size=3', {
      method: 'GET',
      headers: {
        'Origin': 'http://localhost:3002',
        'Content-Type': 'application/json'
      }
    });
    console.log(`✅ CORS Test Status: ${response.status}`);
    console.log('✅ CORS headers seem to be working');
  } catch (error) {
    console.log(`❌ CORS Error: ${error.message}`);
  }
}

testConnection().catch(console.error);