# GitHub分支保护规则配置指南

**文档编号**: CI/CD-002  
**创建日期**: 2025-08-24  
**适用项目**: Cube Castle契约测试自动化体系  
**配置目标**: 建立合并阻塞门禁，确保代码质量  

---

## 📋 配置概述

GitHub分支保护规则是契约测试自动化体系的**最后一道防线**，通过强制要求所有状态检查通过来阻止不符合契约标准的代码合并到主分支。

### 🎯 保护目标
- ✅ 防止未通过契约测试的代码合并
- ✅ 确保所有API变更经过充分验证  
- ✅ 维持主分支的稳定性和生产就绪状态
- ✅ 强制执行代码审查和质量门禁

---

## ⚙️ 必需的分支保护配置

### 🛡️ **主分支保护设置** (master/main)

```yaml
分支保护规则配置:
  branch_name: "master"  # 或 "main"
  
  # 🚨 必需状态检查 - 核心门禁
  required_status_checks:
    strict: true  # 分支必须是最新的
    contexts:
      # 契约测试自动化验证工作流
      - "契约测试验证"           # 核心契约测试套件
      - "契约合规性门禁"         # 合规性验证和阻塞逻辑
      - "Schema变更检测"        # API Schema变更影响分析
      - "性能影响分析"          # 构建和执行性能基准
      
      # 可选的额外检查（如果启用）
      - "build"                # 构建验证
      - "test"                 # 单元测试
      - "lint"                 # 代码风格检查
  
  # 🔒 推送限制
  restrict_pushes: true
  restrictions:
    users: []      # 允许直接推送的用户列表（通常为空）
    teams: []      # 允许直接推送的团队列表（通常为空）
    apps: []       # 允许直接推送的应用列表

  # 👥 Pull Request要求
  required_pull_request_reviews:
    required_approving_review_count: 1    # 至少1个批准审查
    dismiss_stale_reviews: true           # 代码更新时重置审查
    require_code_owner_reviews: false     # 不强制要求代码所有者审查
    restrict_review_dismissal: false      # 不限制审查撤销
    require_review_dismissal_approval: false  # 不需要审查撤销批准

  # 🛡️ 管理员强制执行
  enforce_admins: false  # 管理员不受限制（紧急情况下可绕过）
  
  # 🚫 强制推送和删除保护
  allow_force_pushes: false  # 禁止强制推送
  allow_deletions: false     # 禁止分支删除

  # 📋 其他保护选项
  required_linear_history: false      # 不要求线性历史
  allow_squash_merge: true           # 允许压缩合并
  allow_merge_commit: true           # 允许合并提交
  allow_rebase_merge: true           # 允许变基合并
```

---

## 🔧 配置实施步骤

### **Step 1: 访问仓库设置**
1. 导航到GitHub仓库主页
2. 点击 **Settings** 选项卡
3. 在左侧菜单选择 **Branches**

### **Step 2: 创建分支保护规则**
1. 点击 **Add rule** 按钮
2. 在 **Branch name pattern** 输入: `master` 或 `main`
3. 按照上述配置启用各项保护

### **Step 3: 配置必需状态检查**
在 **Require status checks to pass** 部分：
1. ✅ 选中 **Require branches to be up to date before merging**
2. 搜索并添加以下状态检查：
   ```
   契约测试验证
   契约合规性门禁  
   Schema变更检测
   性能影响分析
   ```

### **Step 4: 配置Pull Request保护**
在 **Require pull request reviews** 部分：
1. ✅ 选中 **Require pull request reviews before merging**
2. 设置 **Required approving reviews**: `1`
3. ✅ 选中 **Dismiss stale PR reviews when new commits are pushed**

### **Step 5: 配置推送限制**  
在 **Restrict pushes** 部分：
1. ✅ 选中 **Restrict pushes that create files larger than 100 MB**
2. 🚫 不选中 **Allow force pushes**
3. 🚫 不选中 **Allow deletions**

---

## 🚨 状态检查详解

### **契约测试验证** (必需)
- **检查内容**: GraphQL Schema语法验证、字段命名规范、契约测试套件执行
- **失败条件**: 任何契约测试失败、构建错误、类型检查失败
- **阻塞影响**: 直接阻止合并，必须修复后才能继续

### **契约合规性门禁** (必需)
- **检查内容**: 汇总所有契约验证结果，执行最终合规性判断
- **失败条件**: 任何上游检查失败、合规性标准不达标
- **阻塞影响**: 最后防线，确保所有质量标准满足

### **Schema变更检测** (必需)
- **检查内容**: API Schema变更检测、向后兼容性分析、影响评估
- **触发条件**: PR中包含 `docs/api/schema.graphql` 变更
- **阻塞影响**: Schema变更必须经过完整验证流程

### **性能影响分析** (建议)
- **检查内容**: Bundle大小分析、构建性能基准、契约测试执行时间
- **失败条件**: 性能回归超过阈值、构建时间异常增长
- **阻塞影响**: 防止性能退化影响生产环境

---

## 🛡️ 分支保护策略

### **开发分支保护** (develop/feature/*)
```yaml
开发分支推荐配置:
  required_status_checks:
    strict: false  # 允许不同步的分支
    contexts:
      - "契约测试验证"  # 仅基础契约验证
  
  required_pull_request_reviews:
    required_approving_review_count: 0  # 开发阶段不强制审查
  
  restrict_pushes: false  # 允许直接推送到开发分支
```

### **发布分支保护** (release/*)
```yaml
发布分支推荐配置:
  required_status_checks:
    strict: true
    contexts:
      - "契约测试验证"
      - "契约合规性门禁"
      - "Schema变更检测" 
      - "性能影响分析"
      - "构建验证"
      - "集成测试"
  
  required_pull_request_reviews:
    required_approving_review_count: 2  # 发布前需要2个审查
  
  enforce_admins: true  # 发布分支管理员也受限制
```

---

## 📊 配置验证清单

### ✅ **配置完成检查**
- [ ] 主分支保护规则已创建
- [ ] 必需状态检查已配置（4个核心检查项）
- [ ] Pull Request审查已启用（至少1个批准）
- [ ] 强制推送已禁用
- [ ] 分支删除已禁用
- [ ] GitHub Actions工作流状态检查已识别

### ✅ **功能验证测试**
- [ ] 创建测试PR验证状态检查触发
- [ ] 故意让契约测试失败验证阻塞效果
- [ ] 验证审查批准要求正常工作
- [ ] 测试直接推送被正确阻止
- [ ] 确认所有状态检查通过后可以正常合并

---

## 🚨 应急处理机制

### **紧急情况绕过规则**
当遇到生产环境紧急修复需要时：

1. **管理员临时调整**:
   ```bash
   # 临时禁用分支保护（仅管理员）
   # 推送紧急修复
   # 立即恢复分支保护规则
   ```

2. **紧急修复标准流程**:
   - 创建 `hotfix/*` 分支
   - 应用最小化修复
   - 快速通过契约验证（如果可能）
   - 合并后立即创建正式PR补充完整验证

3. **事后审查要求**:
   - 24小时内补充完整的契约测试
   - 分析根本原因防止类似问题
   - 更新应急处理文档

---

## 📈 监控和维护

### **分支保护效果监控**
- **阻塞率统计**: 记录因契约测试失败被阻塞的合并请求数量
- **通过率趋势**: 监控契约测试通过率变化
- **修复时间**: 统计从阻塞到修复的平均时间
- **绕过次数**: 监控紧急绕过分支保护的频率

### **定期维护任务**
- **月度审查**: 检查分支保护规则配置是否需要调整
- **状态检查更新**: 根据CI/CD工作流变更更新必需检查项
- **权限审查**: 确认具有绕过权限的用户列表是否合理
- **配置备份**: 定期备份分支保护配置以便快速恢复

---

## 🎯 最佳实践建议

### **配置原则**
1. **严格但实用**: 保证质量同时不影响开发效率
2. **渐进实施**: 新项目可以先启用基础保护，逐步增加检查项
3. **明确责任**: 确保团队理解每个检查项的目的和要求
4. **应急预案**: 准备紧急情况下的处理流程

### **团队协作**
1. **文档共享**: 确保所有团队成员了解分支保护规则
2. **培训支持**: 提供故障排除和问题解决的培训
3. **反馈机制**: 建立团队对分支保护规则的反馈和改进渠道
4. **工具支持**: 提供本地预检查工具减少CI/CD中的失败

---

**配置完成后，Cube Castle项目将具备企业级的代码质量门禁，确保所有合并到主分支的代码都经过严格的契约验证！** 🎉

---

**文档维护**: 2025-08-24 创建  
**下次更新**: 根据实际使用情况调整配置参数  
**负责团队**: DevOps + 前端团队