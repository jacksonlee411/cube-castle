#!/usr/bin/env node

// SWR Configuration Test Script
console.log('🧪 SWR配置测试开始...');

async function testSWRConfig() {
  try {
    // Test 1: Direct API call
    console.log('\n📡 测试1: 直接API调用');
    const response = await fetch('http://localhost:3000/api/employees?page=1&page_size=5');
    
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}: ${response.statusText}`);
    }
    
    const data = await response.json();
    console.log('✅ API调用成功:', {
      员工数量: data.employees?.length || 0,
      总数: data.total_count,
      分页: data.pagination
    });
    
    // Test 2: Check if data structure is correct
    console.log('\n🔍 测试2: 数据结构验证');
    if (data && data.employees && Array.isArray(data.employees)) {
      console.log('✅ 数据结构正确');
      console.log('📊 第一个员工样本:', JSON.stringify(data.employees[0], null, 2));
    } else {
      console.log('❌ 数据结构异常:', {
        hasData: !!data,
        hasEmployees: !!data?.employees,
        isArray: Array.isArray(data?.employees)
      });
    }
    
  } catch (error) {
    console.error('❌ 测试失败:', error.message);
  }
}

testSWRConfig();