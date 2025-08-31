// 测试审计数据重建逻辑
const testAuditData = {
  auditId: "6ab3b6d9-b21c-43c0-90d2-c49edb839bf1",
  operationType: "UPDATE", 
  beforeData: "{}",  // GraphQL返回的空数据
  afterData: "{}",   // GraphQL返回的空数据
  changesSummary: '[{"field": "name", "dataType": "string", "newValue": "测试部门-API验证-已更新", "oldValue": "测试部门-API验证"}, {"field": "description", "dataType": "string", "newValue": "更新后的描述：验证UPDATE API功能", "oldValue": "用于验证CUD API和审计日志功能"}, {"field": "sortOrder", "dataType": "int", "newValue": 200, "oldValue": 100}]'
};

// 模拟前端重建逻辑
function reconstructDataFromChanges(audit) {
  const beforeData = (() => {
    try {
      // 先尝试解析原始的beforeData
      if (audit.beforeData && audit.beforeData !== 'null' && audit.beforeData !== '{}') {
        const parsed = JSON.parse(audit.beforeData);
        if (Object.keys(parsed).length > 0) return parsed;
      }
      
      // 如果beforeData为空，但有changesSummary，尝试从中重建
      if (audit.changesSummary && audit.changesSummary !== 'null' && audit.changesSummary !== '[]') {
        const changes = JSON.parse(audit.changesSummary);
        if (Array.isArray(changes) && changes.length > 0 && changes[0].oldValue !== undefined) {
          const reconstructed = {};
          changes.forEach((change) => {
            if (change.field && change.oldValue !== undefined) {
              reconstructed[change.field] = change.oldValue;
            }
          });
          return Object.keys(reconstructed).length > 0 ? reconstructed : undefined;
        }
      }
      return undefined;
    } catch (error) {
      console.warn('Failed to parse beforeData:', error);
      return undefined;
    }
  })();

  const afterData = (() => {
    try {
      // 先尝试解析原始的afterData  
      if (audit.afterData && audit.afterData !== 'null' && audit.afterData !== '{}') {
        const parsed = JSON.parse(audit.afterData);
        if (Object.keys(parsed).length > 0) return parsed;
      }
      
      // 如果afterData为空，但有changesSummary，尝试从中重建
      if (audit.changesSummary && audit.changesSummary !== 'null' && audit.changesSummary !== '[]') {
        const changes = JSON.parse(audit.changesSummary);
        if (Array.isArray(changes) && changes.length > 0 && changes[0].newValue !== undefined) {
          const reconstructed = {};
          changes.forEach((change) => {
            if (change.field && change.newValue !== undefined) {
              reconstructed[change.field] = change.newValue;
            }
          });
          return Object.keys(reconstructed).length > 0 ? reconstructed : undefined;
        }
      }
      return undefined;
    } catch (error) {
      console.warn('Failed to parse afterData:', error);
      return undefined;
    }
  })();

  return { beforeData, afterData };
}

// 执行测试
const result = reconstructDataFromChanges(testAuditData);
console.log('🔍 测试结果：');
console.log('变更前数据:', JSON.stringify(result.beforeData, null, 2));
console.log('变更后数据:', JSON.stringify(result.afterData, null, 2));

// 预期结果验证
const expectedBefore = {
  name: "测试部门-API验证",
  description: "用于验证CUD API和审计日志功能", 
  sortOrder: 100
};

const expectedAfter = {
  name: "测试部门-API验证-已更新",
  description: "更新后的描述：验证UPDATE API功能",
  sortOrder: 200  
};

console.log('\n✅ 验证结果:');
console.log('变更前重建是否正确:', JSON.stringify(result.beforeData) === JSON.stringify(expectedBefore));
console.log('变更后重建是否正确:', JSON.stringify(result.afterData) === JSON.stringify(expectedAfter));