import React, { useState, useEffect } from 'react';

const TestApiPage: React.FC = () => {
  const [data, setData] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [logs, setLogs] = useState<string[]>([]);

  const addLog = (message: string) => {
    const timestamp = new Date().toLocaleTimeString();
    setLogs(prev => [...prev, `[${timestamp}] ${message}`]);
    console.log(`[${timestamp}] ${message}`);
  };

  const testApiCall = async () => {
    addLog('🚀 开始测试API调用...');
    setLoading(true);
    setError(null);
    
    try {
      const url = 'http://localhost:8080/api/v1/corehr/employees?page=1&page_size=5';
      addLog(`📡 请求URL: ${url}`);
      
      const response = await fetch(url);
      addLog(`📨 响应状态: ${response.status} ${response.statusText}`);
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const result = await response.json();
      addLog(`✅ 成功获取数据: ${result.employees?.length || 0} 个员工`);
      addLog(`📊 总数: ${result.total_count}`);
      
      setData(result);
    } catch (err: any) {
      const errorMsg = err.message || '未知错误';
      addLog(`❌ 请求失败: ${errorMsg}`);
      setError(errorMsg);
    } finally {
      setLoading(false);
    }
  };

  // 检查是否在客户端
  const [isClient, setIsClient] = useState(false);
  
  useEffect(() => {
    setIsClient(true);
    // 添加直接的JavaScript确认
    if (typeof window !== 'undefined') {
      addLog('🔧 组件已挂载在客户端，开始自动测试');
      // 延迟执行以确保DOM已准备好
      setTimeout(() => {
        testApiCall();
      }, 1000);
    } else {
      addLog('⚠️ 当前在服务器端，跳过API调用');
    }
  }, []);

  return (
    <div style={{ padding: '20px', maxWidth: '800px', margin: '0 auto' }}>
      <h1>API连接测试页面</h1>
      
      <div style={{ marginBottom: '20px' }}>
        <button onClick={testApiCall} disabled={loading}>
          {loading ? '测试中...' : '重新测试API'}
        </button>
      </div>

      {/* 日志显示 */}
      <div style={{ marginBottom: '20px' }}>
        <h3>执行日志:</h3>
        <div style={{ 
          backgroundColor: '#f5f5f5', 
          padding: '10px', 
          borderRadius: '4px',
          maxHeight: '200px',
          overflowY: 'auto',
          fontFamily: 'monospace',
          fontSize: '12px'
        }}>
          {logs.map((log, index) => (
            <div key={index}>{log}</div>
          ))}
        </div>
      </div>

      {/* 错误显示 */}
      {error && (
        <div style={{ 
          backgroundColor: '#ffe6e6', 
          color: '#d63031', 
          padding: '10px', 
          borderRadius: '4px', 
          marginBottom: '20px' 
        }}>
          <strong>错误:</strong> {error}
        </div>
      )}

      {/* 数据显示 */}
      {data && (
        <div style={{ marginBottom: '20px' }}>
          <h3>API响应数据:</h3>
          <div style={{ 
            backgroundColor: '#e8f5e8', 
            padding: '10px', 
            borderRadius: '4px',
            marginBottom: '10px'
          }}>
            <strong>总员工数:</strong> {data.total_count}
          </div>
          
          {data.employees && data.employees.length > 0 && (
            <div>
              <h4>员工列表 (前{data.employees.length}个):</h4>
              <ul>
                {data.employees.map((emp: any, index: number) => (
                  <li key={emp.id || index}>
                    {emp.first_name} {emp.last_name} ({emp.employee_number}) - {emp.email}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {/* 调试信息 */}
      <div style={{ 
        marginTop: '20px', 
        padding: '10px', 
        backgroundColor: '#f0f8ff', 
        borderRadius: '4px',
        fontSize: '12px'
      }}>
        <h4>调试信息:</h4>
        <p><strong>是否在客户端:</strong> {isClient ? '是' : '否 (SSR模式)'}</p>
        <p><strong>当前URL:</strong> {typeof window !== 'undefined' ? window.location.href : 'SSR模式'}</p>
        <p><strong>User Agent:</strong> {typeof navigator !== 'undefined' ? navigator.userAgent : 'N/A'}</p>
        <p><strong>是否支持fetch:</strong> {typeof fetch !== 'undefined' ? '是' : '否'}</p>
      </div>
    </div>
  );
};

export default TestApiPage;