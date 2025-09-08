# 🛡️ P3企业级防控系统使用手册

**文档版本**: v1.0  
**创建日期**: 2025-09-07  
**更新时间**: 2025-09-07  
**状态**: 生产就绪  
**适用范围**: Cube Castle项目团队全员  

## 📋 使用手册概述

P3企业级防控系统是Cube Castle项目的核心质量保证机制，通过三大防控系统实现95%+重复代码消除和企业级架构标准维护。本手册为项目团队提供完整的防控系统使用指南。

## 🎯 防控系统架构

### 三大核心系统
- **P3.1 自动化重复检测系统**: 防止代码重复率回归
- **P3.2 架构守护规则系统**: 确保CQRS+端口+API契约一致性
- **P3.3 文档自动同步系统**: 维护核心文档一致性

### 三层纵深防御
1. **本地开发防护**: Pre-commit Hook + 本地质量工具
2. **CI/CD管道防护**: GitHub Actions + 企业级质量门禁
3. **持续监控防护**: 定时检查 + 趋势分析

## 🚀 快速上手指南

### 开发者日常使用

#### 1. 开发前质量检查
```bash
# 进入项目根目录
cd /path/to/cube-castle

# 完整质量检查套件（推荐）
bash scripts/quality/duplicate-detection.sh      # 重复代码检测
node scripts/quality/architecture-validator.js   # 架构一致性验证
node scripts/quality/document-sync.js           # 文档同步检查

# 快速检查前端代码质量
bash scripts/quality/duplicate-detection.sh -s frontend -t 5
```

#### 2. 提交前验证
```bash
# 正常提交 - 系统会自动运行防控检查
git add .
git commit -m "your changes"

# 如果Pre-commit Hook检查失败，根据提示修复问题：
# - 重复代码超标：bash scripts/quality/duplicate-detection.sh --fix
# - 架构违规：根据报告修复CQRS、端口、API契约问题
# - 文档不同步：node scripts/quality/document-sync.js --auto-sync
```

#### 3. 查看质量报告
```bash
# 查看重复代码详细报告
open reports/duplicate-code/html/index.html

# 查看架构违规详情
cat reports/architecture/architecture-validation.json | jq .

# 查看文档同步状态
cat reports/document-sync/document-sync-report.json | jq .
```

## 🔍 P3.1 自动化重复检测系统

### 核心功能
- **阈值控制**: 重复代码率>5%自动阻止提交
- **智能排除**: 自动排除node_modules、测试文件等
- **多格式报告**: HTML可视化 + JSON数据 + 控制台输出

### 使用命令

#### 基础检测
```bash
# 默认扫描（全项目，5%阈值）
bash scripts/quality/duplicate-detection.sh

# 扫描特定范围
bash scripts/quality/duplicate-detection.sh -s frontend    # 仅前端
bash scripts/quality/duplicate-detection.sh -s backend     # 仅后端
bash scripts/quality/duplicate-detection.sh -s config      # 仅配置

# 自定义阈值
bash scripts/quality/duplicate-detection.sh -t 3   # 3%阈值更严格
bash scripts/quality/duplicate-detection.sh -t 10  # 10%阈值更宽松
```

#### 报告生成
```bash
# 生成HTML报告
bash scripts/quality/duplicate-detection.sh -f html

# 生成JSON报告用于CI/CD
bash scripts/quality/duplicate-detection.sh -f json

# 详细输出模式
bash scripts/quality/duplicate-detection.sh -v

# 静默模式（仅显示结果）
bash scripts/quality/duplicate-detection.sh -q
```

#### 自动修复
```bash
# 自动修复重复代码（谨慎使用）
bash scripts/quality/duplicate-detection.sh --fix

# 预览修复效果（推荐）
bash scripts/quality/duplicate-detection.sh --fix --dry-run
```

### 配置自定义

#### 修改检测阈值
编辑 `.jscpdrc.json`:
```json
{
  "threshold": 3,          // 降低到3%更严格
  "minTokens": 30,         // 调整最小检测块大小
  "minLines": 8,           // 调整最小检测行数
  "maxLines": 2000,        // 调整最大检测行数
  "ignore": [              // 添加忽略模式
    "**/node_modules/**",
    "**/dist/**",
    "**/*.test.ts",
    "**/custom-ignore/**"
  ]
}
```

### CI/CD集成

#### GitHub Actions触发条件
- **自动触发**: push到任何分支，PR到主分支
- **定时扫描**: 每周一早上8点完整扫描
- **手动触发**: GitHub Actions页面手动运行

#### 检查失败处理
1. **查看Actions日志**: 点击失败的workflow查看详细信息
2. **本地复现**: 使用相同命令本地运行检测
3. **修复重复代码**: 重构或提取公共函数
4. **重新提交**: 修复后重新push触发检查

## 🏗️ P3.2 架构守护规则系统

### 核心功能
- **CQRS架构守护**: 禁止前端REST查询，强制GraphQL
- **端口配置守护**: 检测硬编码端口，强制统一配置
- **API契约守护**: camelCase命名，废弃字段检查

### 使用命令

#### 基础验证
```bash
# 完整架构验证
node scripts/quality/architecture-validator.js

# 验证特定范围
node scripts/quality/architecture-validator.js --scope frontend
node scripts/quality/architecture-validator.js --scope backend
node scripts/quality/architecture-validator.js --scope config
```

#### 详细分析
```bash
# 详细输出模式
VERBOSE=true node scripts/quality/architecture-validator.js

# 生成JSON报告
node scripts/quality/architecture-validator.js > reports/architecture-custom.json
```

#### 使用架构守护脚本（高级）
```bash
# 使用完整架构守护脚本
bash scripts/quality/architecture-guard.sh -s frontend -v

# 自动修复模式
bash scripts/quality/architecture-guard.sh --fix

# 生成HTML报告
bash scripts/quality/architecture-guard.sh -r html
```

### 常见违规类型及修复方法

#### 1. CQRS架构违规
**问题**: 前端使用REST API进行查询
```typescript
// ❌ 违规：前端使用fetch进行GET查询
const data = await fetch('/api/v1/organizations').then(r => r.json());

// ✅ 正确：使用GraphQL进行查询
const { data } = await apolloClient.query({
  query: gql`query GetOrganizations { organizations { code name } }`
});
```

#### 2. 端口配置违规
**问题**: 硬编码端口号
```typescript
// ❌ 违规：硬编码端口
const API_URL = 'http://localhost:9090/api/v1';

// ✅ 正确：使用统一配置
import { CQRS_ENDPOINTS } from '@shared/config/ports';
const API_URL = CQRS_ENDPOINTS.COMMAND_API;
```

#### 3. API契约违规
**问题**: 使用snake_case字段名
```typescript
// ❌ 违规：snake_case字段名
interface Organization {
  unit_type: string;
  created_at: string;
  parent_unit_id: string;
}

// ✅ 正确：camelCase字段名
interface Organization {
  unitType: string;
  createdAt: string;
  parentCode: string;
}
```

### Pre-commit Hook配置

#### 检查Hook状态
```bash
# 检查Pre-commit Hook是否正确安装
ls -la .git/hooks/pre-commit

# 手动测试Hook
bash scripts/git-hooks/pre-commit-architecture.sh
```

#### Hook失败处理
1. **查看错误信息**: Hook会显示具体的违规类型
2. **本地修复**: 根据提示修复架构问题
3. **重新提交**: 修复后再次commit

## 📝 P3.3 文档自动同步系统

### 核心功能
- **5个同步对监控**: API规范、端口配置、项目状态、依赖版本、架构成果
- **智能冲突检测**: 基于内容分析的不一致性识别
- **自动同步修复**: 支持一键修复文档不一致问题

### 使用命令

#### 基础同步检查
```bash
# 检查所有文档同步状态
node scripts/quality/document-sync.js

# 预览同步更改（推荐）
node scripts/quality/document-sync.js --dry-run

# 执行自动同步
node scripts/quality/document-sync.js --auto-sync
```

#### 高级选项
```bash
# 详细输出模式
VERBOSE=true node scripts/quality/document-sync.js

# 仅检查模式（不修复）
node scripts/quality/document-sync.js --check-only

# 强制同步（跳过安全检查）
node scripts/quality/document-sync.js --auto-sync --force
```

### 同步对详解

#### 1. API规范版本同步
- **源文件**: `docs/api/openapi.yaml`
- **目标文件**: 前端类型定义、技术架构文档
- **同步内容**: API版本号保持一致

#### 2. 端口配置同步
- **源文件**: `frontend/src/shared/config/ports.ts`
- **目标文件**: vite.config.ts、playwright.config.ts、README文档
- **同步内容**: 服务端口配置一致性

#### 3. 项目状态同步
- **源文件**: `CLAUDE.md`
- **目标文件**: README.md、18号计划文档
- **同步内容**: 项目当前状态描述

#### 4. 依赖版本同步
- **源文件**: `frontend/package.json`
- **目标文件**: README、技术架构文档
- **同步内容**: 关键依赖版本号

#### 5. 架构成果同步
- **源文件**: `docs/development-plans/18-duplicate-code-elimination-plan.md`
- **目标文件**: README.md、CLAUDE.md
- **同步内容**: 重复代码消除成果

### 冲突处理

#### 自动修复流程
1. **备份创建**: 自动创建原文件备份
2. **内容比较**: 智能识别不一致部分
3. **安全更新**: 仅更新不一致的特定内容
4. **验证检查**: 更新后再次验证同步状态

#### 手动冲突解决
```bash
# 查看具体冲突详情
cat reports/document-sync/document-sync-report.json | jq '.violations'

# 查看备份文件
ls -la reports/document-sync/backups/

# 手动编辑冲突文件后重新检查
node scripts/quality/document-sync.js
```

## 📊 质量指标监控

### 当前质量状态
- **重复代码率**: 2.11% (目标 < 5%) ✅
- **架构违规数**: 25个已识别 (目标 0个) ⚠️
- **文档同步率**: 20% (目标 > 80%) ⚠️
- **自动化程度**: 100%流程覆盖 ✅

### 质量趋势监控
```bash
# 查看历史质量数据
cat reports/duplicate-code/jscpd-report.json | jq '.timestamp'
cat reports/architecture/architecture-validation.json | jq '.summary'
cat reports/document-sync/sync-history.json | jq '.syncRecords[-5:]'
```

### 质量改善建议

#### 提升架构一致性
1. **逐步修复违规**: 每次提交修复2-3个架构违规
2. **团队培训**: 组织CQRS架构和API契约培训
3. **代码审查**: 在代码审查中关注架构一致性

#### 提升文档同步率
1. **定期检查**: 每周运行一次完整文档同步检查
2. **自动化修复**: 使用--auto-sync自动修复简单冲突
3. **文档规范**: 建立统一的文档更新流程

## ⚠️ 故障排除指南

### 常见问题及解决方案

#### 1. jscpd工具未找到
```bash
# 检查Node.js和npm安装
node --version && npm --version

# 安装jscpd（如果未安装）
npm install -g jscpd

# 或在项目中使用npx
npx jscpd --version
```

#### 2. Pre-commit Hook不工作
```bash
# 检查Hook权限
chmod +x .git/hooks/pre-commit

# 检查Hook内容
cat .git/hooks/pre-commit

# 重新安装Hook
cp scripts/git-hooks/pre-commit-architecture.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

#### 3. GitHub Actions失败
1. **检查工作流状态**: GitHub → Actions → 查看失败的workflow
2. **查看详细日志**: 点击失败的步骤查看错误信息
3. **本地复现**: 使用相同命令在本地运行测试
4. **修复后重试**: push修复代码触发重新运行

#### 4. 报告文件不存在
```bash
# 创建报告目录
mkdir -p reports/{duplicate-code,architecture,document-sync}

# 运行检测生成报告
bash scripts/quality/duplicate-detection.sh
node scripts/quality/architecture-validator.js
node scripts/quality/document-sync.js
```

## 🔧 高级配置

### 自定义ESLint架构规则

#### 编辑架构规则配置
文件: `.eslintrc.architecture.js`
```javascript
rules: {
  // 自定义阈值
  'architecture/no-hardcoded-ports': ['error', {
    allowedPorts: [80, 443, 3000], // 添加允许的端口
    configModule: '@shared/config/ports'
  }],
  
  // 添加项目特定规则
  'architecture/project-specific-rule': ['warn', {
    customConfig: 'your-config'
  }]
}
```

### 自定义文档同步规则

#### 添加新的同步对
编辑: `scripts/quality/document-sync.js`
```javascript
const config = {
  syncPairs: [
    // 现有同步对...
    
    // 添加新的同步对
    {
      name: '新文档同步',
      source: 'source/file.md',
      targets: ['target1.md', 'target2.md'],
      syncType: 'custom',
      pattern: /your-pattern/g,
      description: '自定义同步规则描述'
    }
  ]
};
```

### CI/CD工作流自定义

#### 修改触发条件
编辑: `.github/workflows/duplicate-code-detection.yml`
```yaml
on:
  push:
    branches: [ main, develop ]    # 限制触发分支
    paths:                        # 仅特定文件变更时触发
      - 'frontend/**'
      - 'scripts/**'
  schedule:
    - cron: '0 2 * * 1'          # 每周一凌晨2点运行
```

## 🎓 团队协作指南

### 开发者职责
- **日常检查**: 每次开发前运行质量检查
- **及时修复**: 发现质量问题立即修复
- **知识分享**: 将防控系统最佳实践分享给团队

### 团队负责人职责
- **监控指标**: 定期查看质量指标趋势
- **流程优化**: 根据团队反馈优化防控流程
- **培训支持**: 为团队提供防控系统培训

### 代码审查集成
在代码审查过程中关注：
- **质量报告**: 检查CI/CD中的质量检查结果
- **架构一致性**: 确保新代码符合架构标准
- **文档更新**: 验证相关文档是否同步更新

## 📞 支持与反馈

### 获取帮助
- **文档查阅**: 本手册涵盖90%+常见使用场景
- **命令帮助**: 所有脚本都支持 `-h` 或 `--help` 参数
- **报告分析**: 查看详细的JSON报告了解具体问题

### 问题报告
如发现防控系统问题，请提供：
1. **错误信息**: 完整的错误输出
2. **重现步骤**: 详细的操作步骤
3. **环境信息**: Node.js版本、操作系统等
4. **相关文件**: 问题相关的配置和代码文件

### 改进建议
欢迎提出以下改进建议：
- **新的质量检查规则**
- **更好的自动修复逻辑**
- **增强的报告格式**
- **更智能的冲突解决**

---

**版本历史**:
- v1.0 (2025-09-07): 初始版本，覆盖P3.1+P3.2+P3.3完整功能

**维护团队**: Cube Castle项目组  
**最后更新**: 2025-09-07