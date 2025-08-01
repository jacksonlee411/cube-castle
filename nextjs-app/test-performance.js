#!/usr/bin/env node

// Phase 2 性能和稳定性测试脚本
console.log('⚡ Phase 2性能和稳定性测试开始...\n');

async function runPerformanceTests() {
  const results = {
    apiPerformance: [],
    concurrency: [],
    reliability: [],
    memory: []
  };
  
  // Test 1: API响应时间基准测试
  console.log('⏱️  测试1: API响应时间基准测试');
  for (let i = 0; i < 10; i++) {
    const start = Date.now();
    try {
      const response = await fetch('http://localhost:3000/api/employees?page=1&page_size=20');
      const data = await response.json();
      const duration = Date.now() - start;
      
      results.apiPerformance.push({
        iteration: i + 1,
        duration,
        success: response.ok && data.employees?.length > 0,
        employeeCount: data.employees?.length || 0
      });
      
      process.stdout.write(`${i + 1}/10 (${duration}ms) `);
    } catch (error) {
      results.apiPerformance.push({
        iteration: i + 1,
        duration: Date.now() - start,
        success: false,
        error: error.message
      });
      process.stdout.write(`${i + 1}/10 (ERROR) `);
    }
  }
  console.log('\n');
  
  // Test 2: 并发请求测试
  console.log('🚀 测试2: 并发请求测试');
  const concurrentRequests = Array.from({ length: 5 }, (_, i) => {
    const start = Date.now();
    return fetch(`http://localhost:3000/api/employees?page=${i + 1}&page_size=10`)
      .then(response => response.json())
      .then(data => ({
        page: i + 1,
        duration: Date.now() - start,
        success: !!data.employees,
        employeeCount: data.employees?.length || 0
      }))
      .catch(error => ({
        page: i + 1,
        duration: Date.now() - start,
        success: false,
        error: error.message
      }));
  });
  
  const concurrentResults = await Promise.all(concurrentRequests);
  results.concurrency = concurrentResults;
  console.log(`完成 ${concurrentResults.length} 个并发请求`);
  
  // Test 3: 错误恢复测试
  console.log('🔄 测试3: 错误恢复和稳定性测试');
  const errorRecoveryTests = [
    { name: '无效页码', url: 'http://localhost:3000/api/employees?page=-1&page_size=10' },
    { name: '过大页面', url: 'http://localhost:3000/api/employees?page=1&page_size=1000' },
    { name: '无效参数', url: 'http://localhost:3000/api/employees?page=abc&page_size=def' },
    { name: '空参数', url: 'http://localhost:3000/api/employees?' }
  ];
  
  for (const test of errorRecoveryTests) {
    const start = Date.now();
    try {
      const response = await fetch(test.url);
      const data = await response.json();
      results.reliability.push({
        test: test.name,
        duration: Date.now() - start,
        status: response.status,
        handled: !response.ok && (data.error || data.message),
        graceful: !response.ok && response.status < 500
      });
    } catch (error) {
      results.reliability.push({
        test: test.name,
        duration: Date.now() - start,
        status: 'NETWORK_ERROR',
        handled: false,
        error: error.message
      });
    }
  }
  
  // Test 4: 内存和资源使用测试
  console.log('💾 测试4: 内存使用监控');
  const memoryBefore = process.memoryUsage();
  
  // 执行大量请求来测试内存泄漏
  const heavyRequests = Array.from({ length: 20 }, async (_, i) => {
    const response = await fetch('http://localhost:3000/api/employees?page=1&page_size=50');
    return response.json();
  });
  
  await Promise.all(heavyRequests);
  
  // 强制垃圾回收 (如果可用)
  if (global.gc) {
    global.gc();
  }
  
  const memoryAfter = process.memoryUsage();
  results.memory = {
    before: memoryBefore,
    after: memoryAfter,
    heapGrowth: memoryAfter.heapUsed - memoryBefore.heapUsed,
    rssGrowth: memoryAfter.rss - memoryBefore.rss
  };
  
  // 分析结果
  console.log('\n📊 性能测试结果分析:');
  console.log('================================================');
  
  // API性能分析
  const successfulRequests = results.apiPerformance.filter(r => r.success);
  const avgResponseTime = successfulRequests.reduce((sum, r) => sum + r.duration, 0) / successfulRequests.length;
  const maxResponseTime = Math.max(...successfulRequests.map(r => r.duration));
  const minResponseTime = Math.min(...successfulRequests.map(r => r.duration));
  
  console.log('🏃 API性能基准:');
  console.log(`  成功率: ${successfulRequests.length}/${results.apiPerformance.length} (${((successfulRequests.length / results.apiPerformance.length) * 100).toFixed(1)}%)`);
  console.log(`  平均响应时间: ${avgResponseTime.toFixed(1)}ms`);
  console.log(`  最快响应: ${minResponseTime}ms`);
  console.log(`  最慢响应: ${maxResponseTime}ms`);
  
  // 并发性能分析
  const concurrentSuccess = results.concurrency.filter(r => r.success);
  const avgConcurrentTime = concurrentSuccess.reduce((sum, r) => sum + r.duration, 0) / concurrentSuccess.length;
  
  console.log('\n🚀 并发处理能力:');
  console.log(`  并发成功率: ${concurrentSuccess.length}/${results.concurrency.length} (${((concurrentSuccess.length / results.concurrency.length) * 100).toFixed(1)}%)`);
  console.log(`  平均并发响应时间: ${avgConcurrentTime.toFixed(1)}ms`);
  
  // 错误处理分析
  const gracefulErrors = results.reliability.filter(r => r.graceful).length;
  const handledErrors = results.reliability.filter(r => r.handled).length;
  
  console.log('\n🛡️ 错误处理和稳定性:');
  console.log(`  优雅错误处理: ${gracefulErrors}/${results.reliability.length}`);
  console.log(`  错误信息完整性: ${handledErrors}/${results.reliability.length}`);
  
  // 内存使用分析
  console.log('\n💾 内存使用情况:');
  console.log(`  堆内存增长: ${(results.memory.heapGrowth / 1024 / 1024).toFixed(2)} MB`);
  console.log(`  常驻内存增长: ${(results.memory.rssGrowth / 1024 / 1024).toFixed(2)} MB`);
  
  // 性能评估
  console.log('\n🎯 性能评估:');
  const performanceScore = {
    api: avgResponseTime < 200 ? 'EXCELLENT' : avgResponseTime < 500 ? 'GOOD' : avgResponseTime < 1000 ? 'FAIR' : 'POOR',
    concurrency: concurrentSuccess.length === results.concurrency.length ? 'EXCELLENT' : 'GOOD',
    reliability: gracefulErrors === results.reliability.length ? 'EXCELLENT' : handledErrors >= results.reliability.length * 0.8 ? 'GOOD' : 'FAIR',
    memory: Math.abs(results.memory.heapGrowth) < 10 * 1024 * 1024 ? 'EXCELLENT' : 'GOOD'
  };
  
  console.log(`  API响应速度: ${performanceScore.api} (${avgResponseTime.toFixed(1)}ms平均)`);
  console.log(`  并发处理: ${performanceScore.concurrency}`);
  console.log(`  错误处理: ${performanceScore.reliability}`);
  console.log(`  内存管理: ${performanceScore.memory}`);
  
  return results;
}

runPerformanceTests();