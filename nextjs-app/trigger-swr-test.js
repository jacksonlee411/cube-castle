// Simple test to trigger SWR refresh by accessing the page
console.log('Testing SWR refresh...');

// Create a test request to trigger page refresh
fetch('http://localhost:3000/employees')
  .then(() => console.log('Page refresh triggered'))
  .catch(err => console.error('Test failed:', err));