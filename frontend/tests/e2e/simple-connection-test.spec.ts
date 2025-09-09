import { test, expect } from '@playwright/test';
import { validateTestEnvironment } from './config/test-environment';

let BASE_URL: string;

test('简单的服务器连接测试', async ({ page }) => {
  console.log('开始测试服务器连接...');
  
  // 环境验证和动态端口配置
  const envValidation = await validateTestEnvironment();
  if (!envValidation.isValid) {
    console.error('🚨 环境验证失败:', envValidation.errors);
    throw new Error('测试环境不可用');
  }
  BASE_URL = envValidation.frontendUrl;
  console.log(`✅ 使用前端基址: ${BASE_URL}`);
  
  try {
    await page.goto(BASE_URL + '/', { 
      waitUntil: 'load',
      timeout: 30000 
    });
    
    console.log('页面加载成功');
    console.log('当前URL:', page.url());
    
    // 获取页面标题
    const title = await page.title();
    console.log('页面标题:', title);
    
    // 截图
    await page.screenshot({ 
      path: 'test-results/connection-test.png',
      fullPage: true 
    });
    
    // 基本断言 - 使用动态端口
    expect(page.url()).toContain(BASE_URL.replace('http://', ''));
    
    console.log('测试完成成功');
    
  } catch (error) {
    console.error('测试失败:', error);
    
    // 即使失败也尝试截图
    try {
      await page.screenshot({ path: 'test-results/error-screenshot.png' });
    } catch (screenshotError) {
      console.error('截图也失败了:', screenshotError);
    }
    
    throw error;
  }
});