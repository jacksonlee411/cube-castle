# P3企业级防控系统 - 快速使用指南

## 🛡️ 三层纵深防御架构

### 系统概述
P3防控系统提供三层纵深防御，确保项目代码质量、架构一致性和文档同步：

```yaml
第一层 - 本地开发防护:
  工具: Pre-commit Hook + 本地质量工具
  覆盖: 架构一致性 + 重复代码 + 文档同步
  执行: 每次git commit时自动触发

第二层 - CI/CD管道防护:
  工具: GitHub Actions + ESLint + jscpd
  覆盖: 企业级质量门禁 + 回归检测
  执行: 每次push/PR时自动运行

第三层 - 持续监控防护:
  工具: 定时检查 + 报告生成 + 趋势分析
  覆盖: 长期质量趋势 + 技术债务追踪
  执行: 每日定时运行 + 手动触发
```

## 🚀 快速启动命令

### 完整质量检查 (推荐)
```bash
# 运行所有三个防控系统
bash scripts/quality/duplicate-detection.sh      # P3.1 重复代码检测
node scripts/quality/architecture-validator.js   # P3.2 架构一致性验证  
node scripts/quality/document-sync.js           # P3.3 文档同步检查
```

### 自动修复模式
```bash
# 自动修复重复代码和文档同步问题
bash scripts/quality/duplicate-detection.sh --fix
node scripts/quality/document-sync.js --auto-sync
```

### 查看详细报告
```bash
# 查看HTML格式重复代码报告
open reports/duplicate-code/html/index.html

# 查看架构验证JSON报告
cat reports/architecture/architecture-validation.json | jq

# 查看文档同步报告
cat reports/document-sync/document-sync-report.json | jq
```

## 📊 质量指标监控

### 当前质量状态
- **重复代码率**: 2.11% (目标 < 5%) ✅
- **架构违规数**: 25个已识别 (需修复)
- **文档同步率**: 20% (目标 > 80%)
- **自动化程度**: 100%流程覆盖

### 质量门禁阈值
```yaml
重复代码检测:
  阈值: 5%重复率
  最小令牌: 50个
  最小行数: 10行
  
架构守护规则:
  CQRS架构: 禁止前端REST查询
  端口配置: 检测硬编码端口
  API契约: camelCase命名强制
  
文档同步监控:
  同步对: 5个核心文档对
  不一致类型: 8个检测类型
  自动修复: --auto-sync支持
```

## 🔧 分场景使用指南

### 开发前检查
```bash
# 在开始开发新功能前运行
bash scripts/quality/duplicate-detection.sh -s frontend
node scripts/quality/architecture-validator.js --scope frontend
```

### 提交前验证
```bash
# Pre-commit Hook会自动运行，也可手动执行
node scripts/quality/architecture-validator.js
```

### CI/CD集成状态
```bash
# 查看GitHub Actions工作流状态
# - duplicate-code-detection.yml (P3.1)
# - architecture-validation.yml (P3.2) 
# - document-sync.yml (P3.3)
```

### 问题修复流程
```bash
# 1. 识别问题
node scripts/quality/architecture-validator.js

# 2. 查看详细报告
cat reports/architecture/architecture-validation.json | jq '.violations'

# 3. 修复代码问题
# (根据报告手动修复架构违规)

# 4. 验证修复结果
node scripts/quality/architecture-validator.js
```

## 🎯 工具专用参数

### P3.1 重复代码检测
```bash
# 基础检测
bash scripts/quality/duplicate-detection.sh

# 指定扫描范围
bash scripts/quality/duplicate-detection.sh -s frontend
bash scripts/quality/duplicate-detection.sh -s backend

# 自动修复模式
bash scripts/quality/duplicate-detection.sh --fix

# 生成详细报告
bash scripts/quality/duplicate-detection.sh --verbose
```

### P3.2 架构守护验证
```bash
# 完整架构验证
node scripts/quality/architecture-validator.js

# 指定验证范围
node scripts/quality/architecture-validator.js --scope frontend
node scripts/quality/architecture-validator.js --scope cmd

# 详细模式输出
node scripts/quality/architecture-validator.js --verbose

# 仅检查特定规则
node scripts/quality/architecture-validator.js --rule cqrs
node scripts/quality/architecture-validator.js --rule ports
```

### P3.3 文档同步引擎
```bash
# 检查文档同步状态
node scripts/quality/document-sync.js

# 自动同步模式
node scripts/quality/document-sync.js --auto-sync

# 详细差异分析
node scripts/quality/document-sync.js --verbose

# 检查特定文档对
node scripts/quality/document-sync.js --pair version
node scripts/quality/document-sync.js --pair ports
```

## ⚙️ 配置文件说明

### jscpd配置 (.jscpdrc.json)
```json
{
  "threshold": 5,           # 5%重复率阈值
  "minTokens": 50,          # 最小令牌数
  "minLines": 10,           # 最小行数
  "reporters": ["html", "console", "json"]
}
```

### ESLint自定义规则配置
- `scripts/eslint-rules/no-rest-queries.js` - 禁止前端REST查询
- `scripts/eslint-rules/no-hardcoded-ports.js` - 检测硬编码端口
- `scripts/eslint-rules/enforce-camelcase.js` - 强制camelCase命名

## 🚨 故障排除

### 常见问题
1. **Windows行结束符问题**:
   ```bash
   sed -i 's/\r$//' scripts/quality/*.sh
   ```

2. **jscpd安装超时**:
   ```bash
   npm install jscpd --save-dev
   ```

3. **权限问题**:
   ```bash
   chmod +x scripts/quality/*.sh
   chmod +x scripts/git-hooks/*
   ```

4. **Node.js版本兼容**:
   ```bash
   node --version  # 需要 >= 16.0
   ```

### 获取帮助
- 查看详细文档: `docs/P3-Defense-System-Manual.md`
- 检查CI/CD状态: GitHub Actions页面
- 查看质量报告: `reports/` 目录下的各类报告文件

---

## 📈 集成状态
✅ P3.1 自动化重复检测系统 - 运行正常  
✅ P3.2 架构守护规则系统 - 运行正常  
✅ P3.3 文档自动同步系统 - 运行正常  
✅ GitHub Actions CI/CD集成 - 运行正常  
✅ Pre-commit Hook本地防护 - 运行正常  

**P3企业级防控系统已全面上线，为项目提供三层纵深防御保护！**