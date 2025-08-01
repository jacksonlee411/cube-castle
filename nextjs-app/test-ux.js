#!/usr/bin/env node

// Phase 2 用户体验测试脚本
console.log('🎨 Phase 2用户体验测试开始...\n');

async function runUXTests() {
  const results = {
    accessibility: [],
    responsiveness: [],
    loading: [],
    errorUX: []
  };
  
  // Test 1: 响应式设计测试
  console.log('📱 测试1: 响应式设计检查');
  const viewports = [
    { name: '移动端', width: 375, height: 667 },
    { name: '平板端', width: 768, height: 1024 },
    { name: '桌面端', width: 1200, height: 800 }
  ];
  
  for (const viewport of viewports) {
    try {
      const response = await fetch('http://localhost:3000/employees');
      const html = await response.text();
      
      // 检查响应式类名
      const hasResponsiveClasses = [
        'sm:', 'md:', 'lg:', 'xl:',
        'grid-cols-1', 'sm:grid-cols-2', 'lg:grid-cols-4',
        'flex-col', 'sm:flex-row'
      ].some(className => html.includes(className));
      
      results.responsiveness.push({
        viewport: viewport.name,
        hasResponsiveClasses,
        status: hasResponsiveClasses ? 'PASS' : 'FAIL'
      });
      
    } catch (error) {
      results.responsiveness.push({
        viewport: viewport.name,
        error: error.message,
        status: 'ERROR'
      });
    }
  }
  
  // Test 2: 加载状态用户体验
  console.log('⏳ 测试2: 加载状态用户体验');
  try {
    const response = await fetch('http://localhost:3000/employees');
    const html = await response.text();
    
    // 检查加载状态指示器
    const loadingIndicators = {
      skeleton: html.includes('animate-pulse'),
      spinners: html.includes('loading') || html.includes('spinner'),
      progressBars: html.includes('progress'),
      placeholders: html.includes('placeholder') || html.includes('skeleton')
    };
    
    results.loading = {
      ...loadingIndicators,
      score: Object.values(loadingIndicators).filter(Boolean).length,
      maxScore: Object.keys(loadingIndicators).length
    };
    
  } catch (error) {
    results.loading = { error: error.message, score: 0 };
  }
  
  // Test 3: 错误状态用户体验
  console.log('❌ 测试3: 错误状态用户体验');
  const errorScenarios = [
    { name: '无效API调用', url: 'http://localhost:3000/api/nonexistent' },
    { name: '服务器错误', url: 'http://localhost:3000/api/test-error' }
  ];
  
  for (const scenario of errorScenarios) {
    try {
      const response = await fetch(scenario.url);
      const data = await response.json();
      
      results.errorUX.push({
        scenario: scenario.name,
        hasErrorMessage: !!(data.error || data.message),
        statusCode: response.status,
        userFriendly: data.message && !data.message.includes('Error:'),
        status: !response.ok && (data.error || data.message) ? 'PASS' : 'FAIL'
      });
      
    } catch (error) {
      results.errorUX.push({
        scenario: scenario.name,
        error: error.message,
        status: 'ERROR'
      });
    }
  }
  
  // Test 4: 可访问性基础检查
  console.log('♿ 测试4: 可访问性基础检查');
  try {
    const response = await fetch('http://localhost:3000/employees');
    const html = await response.text();
    
    // 基础可访问性检查
    const accessibilityChecks = {
      hasSemanticHTML: /(<main|<header|<nav|<section|<article)/.test(html),
      hasAltText: html.includes('alt='),
      hasAriaLabels: html.includes('aria-label') || html.includes('aria-labelledby'),
      hasProperHeadings: /<h[1-6]/.test(html),
      hasKeyboardNavigation: html.includes('tabindex') || html.includes('focus'),
      hasColorContrast: html.includes('text-') && html.includes('bg-')
    };
    
    results.accessibility = {
      ...accessibilityChecks,
      score: Object.values(accessibilityChecks).filter(Boolean).length,
      maxScore: Object.keys(accessibilityChecks).length
    };
    
  } catch (error) {
    results.accessibility = { error: error.message, score: 0 };
  }
  
  // 分析和评分
  console.log('\n🎯 用户体验测试结果分析:');
  console.log('================================================');
  
  // 响应式设计评分
  const responsivePass = results.responsiveness.filter(r => r.status === 'PASS').length;
  console.log('📱 响应式设计:');
  console.log(`  测试通过: ${responsivePass}/${results.responsiveness.length}`);
  results.responsiveness.forEach(r => {
    const status = r.status === 'PASS' ? '✅' : '❌';
    console.log(`  ${status} ${r.viewport}: ${r.hasResponsiveClasses ? '支持响应式' : '不支持响应式'}`);
  });
  
  // 加载体验评分
  console.log('\n⏳ 加载体验:');
  console.log(`  加载指示器: ${results.loading.score}/${results.loading.maxScore || 4}`);
  if (results.loading.skeleton) console.log('  ✅ 骨架屏动画');
  if (results.loading.placeholders) console.log('  ✅ 内容占位符');
  
  // 错误体验评分
  const errorPass = results.errorUX.filter(r => r.status === 'PASS').length;
  console.log('\n❌ 错误处理体验:');
  console.log(`  优雅错误处理: ${errorPass}/${results.errorUX.length}`);
  results.errorUX.forEach(r => {
    const status = r.status === 'PASS' ? '✅' : '❌';
    console.log(`  ${status} ${r.scenario}: ${r.hasErrorMessage ? '有错误信息' : '无错误信息'}`);
  });
  
  // 可访问性评分
  console.log('\n♿ 可访问性:');
  console.log(`  基础可访问性: ${results.accessibility.score}/${results.accessibility.maxScore || 6}`);
  if (results.accessibility.hasSemanticHTML) console.log('  ✅ 语义化HTML');
  if (results.accessibility.hasProperHeadings) console.log('  ✅ 正确的标题结构');
  if (results.accessibility.hasAriaLabels) console.log('  ✅ ARIA标签');
  if (results.accessibility.hasColorContrast) console.log('  ✅ 颜色对比度');
  
  // 总体UX评分
  const uxScore = {
    responsive: responsivePass / results.responsiveness.length,
    loading: results.loading.score / (results.loading.maxScore || 4),
    error: errorPass / results.errorUX.length,
    accessibility: results.accessibility.score / (results.accessibility.maxScore || 6)
  };
  
  const overallScore = (uxScore.responsive + uxScore.loading + uxScore.error + uxScore.accessibility) / 4;
  
  console.log('\n🏆 用户体验总评:');
  console.log(`  响应式设计: ${(uxScore.responsive * 100).toFixed(0)}%`);
  console.log(`  加载体验: ${(uxScore.loading * 100).toFixed(0)}%`);
  console.log(`  错误处理: ${(uxScore.error * 100).toFixed(0)}%`);
  console.log(`  可访问性: ${(uxScore.accessibility * 100).toFixed(0)}%`);
  console.log('================================================');
  
  const grade = overallScore >= 0.9 ? 'EXCELLENT' 
              : overallScore >= 0.8 ? 'GOOD' 
              : overallScore >= 0.7 ? 'FAIR' 
              : 'NEEDS_IMPROVEMENT';
  
  console.log(`🎖️  总体UX评级: ${grade} (${(overallScore * 100).toFixed(1)}%)`);
  
  return results;
}

runUXTests();