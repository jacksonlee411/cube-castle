# 端到端测试配置文档

## 测试套件结构

### 1. 架构完整性验证 (architecture-e2e.spec.ts)
- **目标**: 验证6服务→2服务的架构优化
- **测试内容**:
  - 双核心服务健康检查
  - GraphQL统一查询接口
  - 冗余服务移除验证

### 2. 业务流程测试 (business-flow-e2e.spec.ts)
- **目标**: 验证核心CRUD业务流程
- **测试内容**:
  - 完整CRUD操作流程
  - 分页和筛选功能
  - 性能响应时间验证
  - 错误处理和恢复机制
  - 数据一致性验证

### 3. 优化效果验证 (optimization-verification-e2e.spec.ts)
- **目标**: 验证Phase 2-3优化效果
- **测试内容**:
  - 简化验证体系验证
  - DDD简化效果验证  
  - 性能改善量化
  - 系统稳定性验证
  - 监控指标验证

### 4. 回归兼容性测试 (regression-e2e.spec.ts)
- **目标**: 确保重构不破坏现有功能
- **测试内容**:
  - 关键功能回归测试
  - API兼容性验证
  - 数据迁移完整性
  - 跨浏览器兼容性
  - 性能回归验证
  - 异常处理边界测试

### 5. Canvas UI测试 (canvas-e2e.spec.ts)
- **目标**: 验证前端UI组件正常工作
- **测试内容**: [现有测试保持]

### 6. Schema验证测试 (schema-validation.spec.ts)
- **目标**: 验证数据验证机制
- **测试内容**: [现有测试保持]

## 执行方式

### 快速执行
```bash
cd /home/shangmeilin/cube-castle
./run-e2e-tests.sh
```

### 单个测试套件
```bash
cd /home/shangmeilin/cube-castle/frontend
npx playwright test tests/e2e/architecture-e2e.spec.ts
```

### 调试模式
```bash
cd /home/shangmeilin/cube-castle/frontend
npx playwright test --debug
```

## 测试环境要求

### 服务依赖
- 查询服务: http://localhost:8090 
- 命令服务: http://localhost:9090
- 前端服务: http://localhost:3001
- PostgreSQL: localhost:5432
- Neo4j: localhost:7474

### 数据要求
- 测试数据已初始化
- 基础组织结构存在(高谷集团等)

## 测试报告

- **位置**: frontend/playwright-report/
- **访问**: http://localhost:9323 (自动启动)
- **内容**: 详细的测试执行结果、截图、视频

## 失败处理

### 常见问题
1. **服务未启动**: 执行 `./start_optimized_services.sh`
2. **端口冲突**: 检查端口占用情况
3. **数据缺失**: 检查数据库连接和数据完整性
4. **超时问题**: 增加等待时间或检查性能

### 调试步骤
1. 查看控制台输出
2. 检查测试报告详情
3. 运行单个失败的测试
4. 检查服务日志

## 性能基准

### 期望指标
- 页面加载: < 3秒
- API响应: < 1秒
- 成功率: > 80%
- 内存使用: < 100MB

### 优化收益验证
- 服务数量: 6 → 2 (减少67%)
- 验证代码: 889行 → 100行 (减少89%)  
- 打包体积: 预期减少50KB+

这套测试体系确保重构后的系统在功能完整性、性能优化和稳定性方面都达到预期目标。