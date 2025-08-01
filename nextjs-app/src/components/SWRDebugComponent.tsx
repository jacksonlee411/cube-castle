import React from 'react';
import useSWR from 'swr';

// 最简单的fetcher
const debugFetcher = async (url: string) => {
  console.log('🔬 DEBUG Fetcher被调用:', url);
  const response = await fetch(url);
  const data = await response.json();
  console.log('🔬 DEBUG Fetcher成功:', data);
  return data;
};

export function SWRDebugComponent() {
  console.log('🔬 SWRDebugComponent渲染');
  
  // Force an immediate effect to trigger manual fetch
  React.useEffect(() => {
    console.log('🔬 Debug component mounted - testing direct fetch');
    
    // Test direct fetch without SWR
    const testDirectFetch = async () => {
      try {
        console.log('🔬 Testing direct fetch to API...');
        const response = await fetch('/api/employees?page=1&page_size=3');
        const data = await response.json();
        console.log('🔬 Direct fetch SUCCESS:', data);
      } catch (error) {
        console.error('🔬 Direct fetch ERROR:', error);
      }
    };
    
    testDirectFetch();
  }, []);
  
  // 最简单的SWR调用，没有任何配置
  const { data, error, isLoading, mutate } = useSWR('/api/employees?page=1&page_size=3', debugFetcher);
  
  // Force SWR to trigger manually
  React.useEffect(() => {
    const timer = setTimeout(() => {
      console.log('🔬 Forcing SWR mutate...');
      mutate();
    }, 2000);
    
    return () => clearTimeout(timer);
  }, [mutate]);
  
  console.log('🔬 SWR状态:', { 
    hasData: !!data, 
    hasError: !!error, 
    isLoading,
    dataType: typeof data 
  });
  
  if (isLoading) {
    console.log('🔬 SWR正在加载...');
    return <div>🔬 调试加载中...</div>;
  }
  
  if (error) {
    console.log('🔬 SWR错误:', error);
    return <div>🔬 调试错误: {error.message}</div>;
  }
  
  if (data) {
    console.log('🔬 SWR成功获取数据:', data);
    return <div>🔬 调试成功: {data.employees?.length || 0} 员工</div>;
  }
  
  console.log('🔬 SWR无数据');
  return <div>🔬 调试: 无数据</div>;
}