const http = require('http');
const fs = require('fs');

// 测试职位数据
const testPositions = [
  {
    code: '3000001',
    organizationCode: '2000001',
    positionType: 'FULL_TIME',
    jobProfileId: 'JP001',
    status: 'OPEN',
    budgetedFte: 1.0,
    tenantId: '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    title: '高级软件工程师',
    description: '负责后端系统开发和架构设计'
  },
  {
    code: '3000002',
    organizationCode: '2000001',
    positionType: 'FULL_TIME',
    jobProfileId: 'JP002',
    status: 'FILLED',
    budgetedFte: 1.0,
    tenantId: '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    title: '前端工程师',
    description: '负责前端界面开发和用户体验优化'
  },
  {
    code: '3000003',
    organizationCode: '2000002',
    positionType: 'PART_TIME',
    jobProfileId: 'JP003',
    status: 'OPEN',
    budgetedFte: 0.5,
    tenantId: '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    title: '兼职UI设计师',
    description: '负责界面设计和用户体验研究'
  },
  {
    code: '3000004',
    organizationCode: '2000001',
    positionType: 'INTERN',
    jobProfileId: 'JP004',
    status: 'OPEN',
    budgetedFte: 0.5,
    tenantId: '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9',
    title: '软件开发实习生',
    description: '参与项目开发，学习软件开发流程'
  }
];

// 创建职位的GraphQL mutation
function createPositionMutation(position) {
  return JSON.stringify({
    query: `
      mutation CreatePosition($input: CreatePositionInput!) {
        createPosition(input: $input) {
          code
          organizationCode
          status
          positionType
          budgetedFte
          details {
            title
            description
          }
        }
      }
    `,
    variables: {
      input: {
        code: position.code,
        organizationCode: position.organizationCode,
        positionType: position.positionType,
        jobProfileId: position.jobProfileId,
        status: position.status,
        budgetedFte: position.budgetedFte,
        details: {
          title: position.title,
          description: position.description
        }
      }
    }
  });
}

async function createTestData() {
  console.log('开始创建测试职位数据...\n');
  
  for (let i = 0; i < testPositions.length; i++) {
    const position = testPositions[i];
    console.log(`创建职位 ${i + 1}/${testPositions.length}: ${position.title} (${position.code})`);
    
    const postData = createPositionMutation(position);
    
    const options = {
      hostname: 'localhost',
      port: 8091,
      path: '/graphql',
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Tenant-ID': position.tenantId,
        'Content-Length': Buffer.byteLength(postData)
      }
    };
    
    try {
      await new Promise((resolve, reject) => {
        const req = http.request(options, (res) => {
          let data = '';
          res.on('data', chunk => data += chunk);
          res.on('end', () => {
            try {
              const result = JSON.parse(data);
              if (result.errors) {
                console.log(`  ❌ 创建失败: ${result.errors.map(e => e.message).join(', ')}`);
              } else if (result.data && result.data.createPosition) {
                console.log(`  ✅ 创建成功: ${result.data.createPosition.code}`);
              } else {
                console.log(`  ⚠️  响应格式异常: ${data}`);
              }
            } catch (e) {
              console.log(`  ⚠️  JSON解析错误: ${e.message}`);
              console.log(`  原始响应: ${data}`);
            }
            resolve();
          });
        });
        
        req.on('error', (e) => {
          console.log(`  ❌ 请求错误: ${e.message}`);
          reject(e);
        });
        
        req.write(postData);
        req.end();
      });
    } catch (error) {
      console.log(`  ❌ 创建职位失败: ${error.message}`);
    }
    
    // 稍微延迟，避免请求过快
    await new Promise(resolve => setTimeout(resolve, 100));
  }
  
  console.log('\n测试数据创建完成！');
}

createTestData().catch(console.error);