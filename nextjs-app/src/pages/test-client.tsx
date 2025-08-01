import React, { useEffect, useState } from 'react';

const TestClientSidePage: React.FC = () => {
  const [mounted, setMounted] = useState(false);
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setMounted(true);
    console.log('🔥 TestClientSide: 组件已挂载，开始测试客户端API调用');
    
    const testAPI = async () => {
      setLoading(true);
      try {
        console.log('📡 TestClientSide: 发送API请求');
        const response = await fetch('/api/employees?page=1&page_size=3');
        console.log('📨 TestClientSide: 收到响应', response.status);
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        
        const result = await response.json();
        console.log('✅ TestClientSide: 成功获取数据', result);
        setData(result);
      } catch (err: any) {
        console.error('❌ TestClientSide: 请求失败', err);
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    testAPI();
  }, []);

  if (!mounted) {
    return <div>Server Side Rendering...</div>;
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl mb-4">客户端API调用测试</h1>
      
      {loading && (
        <div className="bg-blue-100 p-4 rounded mb-4">
          <p className="text-blue-700">正在加载...</p>
        </div>
      )}
      
      {error && (
        <div className="bg-red-100 p-4 rounded mb-4">
          <p className="text-red-700">错误: {error}</p>
        </div>
      )}
      
      {data && (
        <div className="bg-green-100 p-4 rounded mb-4">
          <p className="text-green-700">
            成功！获取到 {(data as any).employees?.length} / {(data as any).total_count} 个员工
          </p>
        </div>
      )}
      
      <div className="mt-4">
        <p><strong>Mounted:</strong> {mounted ? '是' : '否'}</p>
        <p><strong>Loading:</strong> {loading ? '是' : '否'}</p>
        <p><strong>Error:</strong> {error || '无'}</p>
        <p><strong>Data:</strong> {data ? '有数据' : '无数据'}</p>
      </div>
    </div>
  );
};

export default TestClientSidePage;