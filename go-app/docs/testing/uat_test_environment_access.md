# 测试环境访问信息

**环境名称**: Cube Castle UAT测试环境  
**更新时间**: 2025-07-30 16:45  
**状态**: ✅ 运行中

## 🌐 服务地址

### API服务器 (Go后端)
- **地址**: `http://localhost:8080`
- **健康检查**: `http://localhost:8080/health`
- **API文档**: API接口遵循RESTful设计
- **状态**: ✅ 正常运行

### Web应用 (Next.js前端)
- **地址**: `http://localhost:3000`
- **状态**: ✅ 已修复ES模块兼容性问题，正常运行
- **版本信息**: Next.js 14.1.4 + Ant Design 5.20.6 (稳定版本组合)
- **兼容性测试页面**: `http://localhost:3000/test-antd`

## 🔧 前端修复进展 (2025-07-30)

### ✅ ES模块兼容性问题已解决
- **问题**: Next.js 14.2.30 与 Ant Design 5.26.7 ES模块冲突
- **解决方案**: 版本降级 + Webpack配置优化
- **修复时间**: 2025-07-30 16:30
- **详细报告**: `nextjs-app/ES_MODULE_COMPATIBILITY_FIX_REPORT.md`

### 🎯 已验证功能页面
- **Ant Design兼容性测试**: `http://localhost:3000/test-antd` ✅
- **员工管理**: `http://localhost:3000/employees` ✅
- **组织架构图**: `http://localhost:3000/organization/chart` ✅
- **职位管理**: `http://localhost:3000/positions` ✅

### 📦 稳定版本组合
```json
{
  "next": "14.1.4",
  "antd": "5.20.6", 
  "@ant-design/icons": "5.3.7",
  "rc-util": "5.38.2",
  "@rc-component/util": "1.1.0"
}
```

## 🔑 测试账号信息

### 系统级别测试
**租户ID (X-Tenant-ID header)**: `550e8400-e29b-41d4-a716-446655440000`

### 数据库连接
- **主机**: localhost:5432
- **数据库**: cube_castle_uat
- **用户名**: cube_user
- **密码**: cube_password_123
- **状态**: ✅ 已连接

## 📋 API测试端点

### 核心业务API

#### 1. 健康检查
```bash
curl http://localhost:8080/health
```

#### 2. 组织单元管理
```bash
# 获取组织列表
curl http://localhost:8080/api/v1/organization-units

# 创建组织单元
curl -X POST http://localhost:8080/api/v1/organization-units \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "name": "测试部门",
    "unit_type": "DEPARTMENT",
    "description": "用于UAT测试的部门"
  }'
```

#### 3. 职位管理 (已修复的核心功能)
```bash
# 获取职位列表
curl http://localhost:8080/api/v1/positions

# 创建职位 (展示结构化错误处理)
curl -X POST http://localhost:8080/api/v1/positions \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "position_type": "FULL_TIME",
    "job_profile_id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "department_id": "ec3afce7-4466-420d-bfa8-b569880b984a",
    "status": "OPEN",
    "budgeted_fte": 1.0,
    "details": {
      "title": "高级软件工程师",
      "level": "L4"
    }
  }'

# 测试错误处理 (展示改进的验证消息)
curl -X POST http://localhost:8080/api/v1/positions \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "position_type": "INVALID_TYPE",
    "job_profile_id": "",
    "department_id": "invalid-uuid"
  }'
```

#### 4. 员工管理 (CoreHR)
```bash
# 获取员工列表 (已修复)
curl http://localhost:8080/api/v1/corehr/employees

# 创建员工
curl -X POST http://localhost:8080/api/v1/corehr/employees \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: 550e8400-e29b-41d4-a716-446655440000" \
  -d '{
    "name": "张测试",
    "email": "zhangtest@company.com",
    "position": "软件工程师",
    "department": "技术部",
    "hire_date": "2025-07-30"
  }'
```

#### 5. 权限验证测试 (已修复)
```bash
# 测试缺失租户ID (展示改进的权限验证)
curl -X POST http://localhost:8080/api/v1/positions \
  -H "Content-Type: application/json" \
  -d '{
    "position_type": "FULL_TIME",
    "job_profile_id": "test",
    "department_id": "test"
  }'

# 测试无效租户ID格式
curl -X POST http://localhost:8080/api/v1/positions \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: invalid-uuid" \
  -d '{
    "position_type": "FULL_TIME",
    "job_profile_id": "test",
    "department_id": "test"
  }'
```

## 🧪 自动化测试

### UAT Stage 2 测试套件
```bash
# 进入Go应用目录
cd /home/shangmeilin/cube-castle/go-app

# 执行完整的UAT Stage 2测试
./uat_stage2_test.sh
```

**预期结果**: 9/11测试通过 (81.8%通过率)

## 🔍 验证重点

### 1. TC2006a - 员工列表接口 ✅
- 验证API返回正确的JSON结构
- 确认包含 `"employees"` 和 `"total_count"` 字段

### 2. TC2009 - 错误处理增强 ✅
- 测试无效数据时返回结构化错误信息
- 验证字段级验证错误详情
- 确认错误消息包含具体修复建议

### 3. TC2010 - 权限验证改进 ✅
- 测试缺失X-Tenant-ID头的情况
- 验证无效UUID格式的错误处理
- 确认返回用户友好的错误消息

## 📊 监控和日志

### 应用日志
- **位置**: `/home/shangmeilin/cube-castle/go-app/server.log`
- **格式**: 结构化JSON日志
- **包含**: API请求、数据库操作、性能指标、错误信息

### 性能指标
- **API响应时间**: <200ms
- **内存使用**: ~1.6MB (稳定)
- **数据库连接**: 正常
- **错误处理**: 结构化响应

## 🚀 启动前端应用 (可选)

如需测试完整的Web界面:

```bash
cd /home/shangmeilin/cube-castle/nextjs-app
npm run dev
```

然后访问 `http://localhost:3000`

## 📞 技术支持

如在测试过程中遇到问题:
1. 检查API服务健康状态: `curl http://localhost:8080/health`
2. 查看实时日志: `tail -f /home/shangmeilin/cube-castle/go-app/server.log`
3. 重启API服务: `./cube-castle-api` (在go-app目录下)

---

**环境就绪状态**: ✅ 可以开始测试  
**建议测试顺序**: 健康检查 → 权限验证 → 错误处理 → 业务功能 → 自动化测试