#!/usr/bin/env node

// 前后端集成测试脚本
console.log('🧪 Phase 2前后端集成测试开始...\n');

async function runIntegrationTests() {
  const tests = [];
  
  // Test 1: 后端服务健康检查
  try {
    console.log('📡 测试1: 后端服务健康检查');
    const response = await fetch('http://localhost:8080/health');
    const health = await response.json();
    tests.push({
      name: '后端服务健康检查',
      status: response.ok ? 'PASS' : 'FAIL',
      details: `状态: ${health.status}, 时间: ${health.timestamp}`
    });
  } catch (error) {
    tests.push({
      name: '后端服务健康检查',
      status: 'FAIL',
      details: `错误: ${error.message}`
    });
  }
  
  // Test 2: Next.js API路由测试
  try {
    console.log('🔗 测试2: Next.js API路由');
    const response = await fetch('http://localhost:3000/api/employees?page=1&page_size=3');
    const data = await response.json();
    tests.push({
      name: 'Next.js API路由',
      status: response.ok && data.employees ? 'PASS' : 'FAIL',
      details: `状态码: ${response.status}, 员工数: ${data.employees?.length || 0}, 总数: ${data.total_count}`
    });
  } catch (error) {
    tests.push({
      name: 'Next.js API路由',
      status: 'FAIL',
      details: `错误: ${error.message}`
    });
  }
  
  // Test 3: 数据格式验证
  try {
    console.log('📊 测试3: 数据格式验证');
    const response = await fetch('http://localhost:3000/api/employees?page=1&page_size=1');
    const data = await response.json();
    const employee = data.employees?.[0];
    
    const requiredFields = ['id', 'employee_number', 'first_name', 'last_name', 'email', 'status'];
    const missingFields = requiredFields.filter(field => !employee?.[field]);
    
    tests.push({
      name: '数据格式验证',
      status: missingFields.length === 0 ? 'PASS' : 'FAIL',
      details: missingFields.length === 0 
        ? `所有必需字段存在: ${requiredFields.join(', ')}`
        : `缺失字段: ${missingFields.join(', ')}`
    });
  } catch (error) {
    tests.push({
      name: '数据格式验证',
      status: 'FAIL',
      details: `错误: ${error.message}`
    });
  }
  
  // Test 4: 错误处理测试
  try {
    console.log('❌ 测试4: 错误处理机制');
    const response = await fetch('http://localhost:3000/api/employees', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ invalid: 'data' })
    });
    const data = await response.json();
    
    tests.push({
      name: '错误处理机制',
      status: !response.ok && data.error ? 'PASS' : 'FAIL',
      details: `状态码: ${response.status}, 错误信息: ${data.error || data.message || 'N/A'}`
    });
  } catch (error) {
    tests.push({
      name: '错误处理机制',
      status: 'PARTIAL',
      details: `网络错误: ${error.message}`
    });
  }
  
  // Test 5: 分页功能测试
  try {
    console.log('📄 测试5: 分页功能');
    const response = await fetch('http://localhost:3000/api/employees?page=2&page_size=10');
    const data = await response.json();
    
    tests.push({
      name: '分页功能',
      status: data.pagination && data.pagination.page === 2 ? 'PASS' : 'FAIL',
      details: `页码: ${data.pagination?.page}, 页大小: ${data.pagination?.page_size}, 总页数: ${data.pagination?.total_pages}`
    });
  } catch (error) {
    tests.push({
      name: '分页功能',
      status: 'FAIL',
      details: `错误: ${error.message}`
    });
  }
  
  // 输出测试结果
  console.log('\n🎯 集成测试结果汇总:');
  console.log('================================================');
  
  let passed = 0, failed = 0, partial = 0;
  
  tests.forEach((test, index) => {
    const status = test.status === 'PASS' ? '✅' 
                  : test.status === 'FAIL' ? '❌' 
                  : '⚠️';
    console.log(`${index + 1}. ${status} ${test.name}`);
    console.log(`   详情: ${test.details}\n`);
    
    if (test.status === 'PASS') passed++;
    else if (test.status === 'FAIL') failed++;
    else partial++;
  });
  
  console.log('================================================');
  console.log(`总结: ${passed} 通过, ${failed} 失败, ${partial} 部分通过`);
  console.log(`成功率: ${((passed / tests.length) * 100).toFixed(1)}%`);
  
  return tests;
}

runIntegrationTests();