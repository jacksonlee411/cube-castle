import React from 'react';

const TestCrud: React.FC = () => {
  const handleTestCreate = async () => {
    try {
      const response = await fetch('http://localhost:9090/api/v1/organization-units', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          name: `测试部门_${Date.now()}`,
          unit_type: 'DEPARTMENT',
          description: '这是一个测试部门'
        })
      });
      
      if (response.ok) {
        const result = await response.json();
        console.log('创建成功:', result);
        alert('创建成功! 查看控制台获取详细信息');
      } else {
        const error = await response.text();
        console.error('创建失败:', error);
        alert('创建失败: ' + error);
      }
    } catch (error) {
      console.error('请求错误:', error);
      alert('请求错误: ' + error);
    }
  };

  return (
    <div style={{ padding: '20px' }}>
      <h2>🧪 CRUD功能测试页面</h2>
      
      <div style={{ marginBottom: '20px' }}>
        <h3>📊 后端服务状态</h3>
        <p>✅ GraphQL服务 (端口 8090) - 数据查询正常</p>
        <p>✅ 命令端服务 (端口 9090) - 等待验证</p>
      </div>

      <div>
        <h3>🔨 CRUD操作测试</h3>
        <button 
          onClick={handleTestCreate}
          style={{
            padding: '10px 20px',
            backgroundColor: '#007bff',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            marginRight: '10px'
          }}
        >
          测试创建组织单元
        </button>
        
        <div style={{ marginTop: '20px', fontSize: '14px', color: '#666' }}>
          <p>点击按钮测试后端CRUD API</p>
          <p>查看浏览器控制台获取详细响应信息</p>
        </div>
      </div>
    </div>
  );
};

export default TestCrud;