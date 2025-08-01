import React from 'react';
import useSWR from 'swr';

const forceFetcher = async (url: string) => {
  console.log('🟢 FORCE Fetcher被调用:', url);
  const response = await fetch(url);
  const data = await response.json();
  console.log('🟢 FORCE Fetcher成功:', data);
  return data;
};

export function ForceSWRComponent() {
  const [mounted, setMounted] = React.useState(false);
  
  // 确保只在客户端运行
  React.useEffect(() => {
    setMounted(true);
    console.log('🟢 ForceSWRComponent已挂载，开始SWR调用');
  }, []);
  
  // 只在客户端挂载后才调用SWR
  const { data, error, isLoading, mutate } = useSWR(
    mounted ? '/api/employees?page=1&page_size=3' : null,
    forceFetcher,
    {
      revalidateOnMount: true,
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      dedupingInterval: 0,  // 禁用去重
      fallbackData: undefined,
    }
  );
  
  // 强制触发
  React.useEffect(() => {
    if (mounted && !data && !isLoading && !error) {
      console.log('🟢 强制触发SWR mutate');
      mutate();
    }
  }, [mounted, data, isLoading, error, mutate]);
  
  if (!mounted) {
    return <div>🟢 等待客户端挂载...</div>;
  }
  
  if (isLoading) {
    return <div>🟢 强制SWR加载中...</div>;
  }
  
  if (error) {
    return <div>🟢 强制SWR错误: {error.message}</div>;
  }
  
  if (data) {
    return <div>🟢 强制SWR成功: {data.employees?.length || 0} 员工</div>;
  }
  
  return <div>🟢 强制SWR: 无数据</div>;
}