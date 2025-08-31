// 测试修复后的前端认证流程
async function testAuthFix() {
  console.log('🔧 测试前端认证修复...');
  
  try {
    // 测试开发令牌获取
    console.log('1. 测试开发令牌获取...');
    const tokenResponse = await fetch('http://localhost:9090/auth/dev-token', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        userID: 'frontend-user',
        tenantID: 'frontend-tenant',
        roles: ['ADMIN', 'USER'],
        duration: '1h'
      })
    });
    
    const tokenData = await tokenResponse.json();
    console.log('令牌响应:', tokenData);
    
    if (tokenData.success && tokenData.data.token) {
      console.log('✅ 开发令牌获取成功');
      const token = tokenData.data.token;
      
      // 测试GraphQL查询
      console.log('2. 测试GraphQL组织查询...');
      const graphqlResponse = await fetch('http://localhost:8090/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        },
        body: JSON.stringify({
          query: `query { 
            organizations { 
              data { 
                code 
                name 
                unitType 
                status 
              } 
              pagination { 
                total 
              } 
            } 
          }`
        })
      });
      
      const graphqlData = await graphqlResponse.json();
      console.log('GraphQL查询结果:', graphqlData);
      
      if (graphqlData.success && graphqlData.data.organizations) {
        console.log('✅ GraphQL组织查询成功');
        console.log(`📊 共找到 ${graphqlData.data.organizations.pagination.total} 个组织单元`);
        
        // 显示前几个组织
        graphqlData.data.organizations.data.slice(0, 3).forEach((org, index) => {
          console.log(`${index + 1}. ${org.name} (${org.code}) - ${org.status}`);
        });
        
        console.log('🎉 认证修复成功！前端应该可以正常加载组织列表了。');
      } else {
        console.log('❌ GraphQL查询失败:', graphqlData);
      }
    } else {
      console.log('❌ 开发令牌获取失败:', tokenData);
    }
    
  } catch (error) {
    console.error('❌ 测试过程中出现错误:', error.message);
  }
}

// 在Node.js环境中运行
if (typeof require !== 'undefined') {
  // 对于Node.js，需要导入fetch
  const fetch = require('node-fetch');
  global.fetch = fetch;
}

testAuthFix();