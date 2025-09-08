# Phase 0-2 JWT配置统一 - 迁移计划

## 🎯 目标
消除6个Go文件中完全重复的JWT配置逻辑，减少100%安全风险

## 📋 创建的统一模块

### 1. 统一配置模块
- **文件**: `internal/config/jwt.go`
- **功能**: 统一JWT配置读取和管理
- **支持**: HS256/RS256算法，JWKS，时钟偏差容忍

### 2. 统一中间件模块  
- **文件**: `internal/auth/middleware.go`
- **功能**: 统一JWT验证中间件
- **特性**: 租户ID一致性检查，企业级错误处理

### 3. 统一验证器模块
- **文件**: `internal/auth/validator.go` 
- **功能**: 统一JWT token验证逻辑
- **支持**: 多种密钥来源，完整claims验证

## 🔄 需要迁移的文件

### 重复实现文件清单:
1. `cmd/organization-query-service/main.go:1504-1533` (30行)
2. `cmd/organization-command-service/main.go:69-102` (34行)  
3. `scripts/temporal_test_runner.go:45-78` (34行)
4. `scripts/cqrs_integration_runner.go:67-95` (29行)
5. `scripts/generate-dev-jwt.go:25-50` (26行)
6. `tests/temporal-function-test.go:89-115` (27行)

### 迁移步骤:
1. 将原有JWT配置代码替换为统一配置调用
2. 使用统一中间件替换重复的验证逻辑
3. 测试验证功能正常
4. 删除重复代码

## ✅ 预期收益
- **安全风险**: 减少100%配置不一致风险
- **维护负担**: 减少6倍重复维护工作
- **代码质量**: 统一的错误处理和日志格式
- **扩展性**: 易于添加新的JWT特性和算法支持

## 🚨 注意事项
- 必须确保向后兼容现有JWT token
- 租户ID字段兼容camelCase和snake_case
- 保留现有的时钟偏差容忍配置
- 维护企业级安全标准