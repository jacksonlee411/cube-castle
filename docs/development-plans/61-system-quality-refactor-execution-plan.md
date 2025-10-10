# 61号文档：系统级质量重构执行计划

**版本**: v1.0
**创建日期**: 2025-10-10
**维护团队**: 架构组 + 后端团队 + 前端团队 + 平台/DevOps团队
**状态**: 执行中
**遵循原则**: CLAUDE.md 资源唯一性与跨层一致性原则（最高优先级）
**关联计划**: 60号文档 - 系统级质量整合与重构计划

## 文档目的

本文档是 [60号文档](./60-system-wide-quality-refactor-plan.md) 的执行落地指南，提供从"规划"到"执行"的完整路线图，包含：
- 阶段零启动准备的详细步骤
- 第一阶段（契约与类型统一）的具体实施路径
- 后续阶段的时间线与里程碑
- 每日/每周可执行的任务清单
- 风险应对与资源需求

## 当前状态分析

### ✅ 已就绪资源
- **质量分析文档**: 50-59 号文档全部存在并可引用
- **后端类型文件**: `cmd/organization-command-service/internal/types/models.go`、`responses.go` 已存在
- **前端类型文件**: `frontend/src/shared/types/organization.ts`、`frontend/src/shared/validation/schemas.ts` 已存在
- **现有工具**: `scripts/check-api-naming.sh`、`scripts/generate-implementation-inventory.js` 可用
- **API 契约**: `docs/api/openapi.yaml`、`docs/api/schema.graphql` 为唯一真源

### ❌ 待建设施（第一阶段目标）
- `scripts/contract/` 目录及契约同步脚本体系
- `shared/contracts/organization.json` 中间契约文件
- CI Job `contract-sync` 和 `contract-snapshot`
- Go/TS 代码生成器 (`contract_gen.go`、`contract_gen.ts`)
- 跨层快照测试框架 (`tests/contract/*.snap`)

### ⚠️ 关键依赖
- 60 号计划文档需提交到主干并获架构组批准
- 各阶段负责人需明确指定
- staging 环境访问权限需就绪

---

## 阶段零：启动准备（预估 3-5 天）

### Step 0.1: 计划文档正式化（优先级：P0）

**执行时间**: Day 1 上午
**负责人**: 架构组负责人

**任务清单**:
- [ ] 确认 60 号计划已合入主干并完成评审（若尚未提交，按常规流程补交）。
- [ ] 在 `docs/development-plans/00-README.md` 的“活跃计划”中列出 60/61 号文档。
- [ ] 确认本执行计划（61 号）为最新版本并已提交。

**验收标准**:
- [ ] 主分支可查阅 60 号计划与本执行计划。
- [ ] 计划索引与实际文档保持一致。

---

### Step 0.2: 组建跨团队小组（优先级：P0）

**执行时间**: Day 1 下午
**负责人**: 项目经理 + 架构组负责人

**任务清单**:
- [ ] 明确各阶段责任人与时间投入，并将人员列表存档于共享工作区（而非本文档）。
- [ ] 召开启动会议（建议 1 小时），确认目标、分工、沟通节奏；会议纪要上传至协作空间。
- [ ] 建立沟通渠道（群组、双周同步会、共享文档空间等），通知所有参与者。

**验收标准**:
- [ ] 各阶段责任人已确认
- [ ] 启动会议已完成并有会议纪要
- [ ] 沟通渠道已建立并测试可用

---

### Step 0.3: 评估前置条件（优先级：P1）

**执行时间**: Day 2
**负责人**: 第一阶段责任人

- [ ] 验证 API 契约干净无未合并变更；检查 53、56 号计划列出的风险项已关闭或纳入本计划。
- [ ] 运行 `scripts/generate-implementation-inventory.js` 输出参考基线（保存为团队共享文件或 CI 工件，非必须提交到仓库）。
- [ ] 验证辅助脚本可执行（如 `scripts/check-api-naming.sh`），确认无运行错误。

**验收标准**:
- [ ] API 契约文件干净无变更
- [ ] 相关计划无阻塞项
- [ ] 基线文件已提交到 Git
- [ ] 现有工具测试通过

---

### Step 0.4: 建立迭代跟踪（优先级：P1）

**执行时间**: Day 2-3
**负责人**: 项目经理

**任务清单**:
- [ ] 创建执行跟踪文档
  ```bash
  cat > docs/development-plans/60-execution-tracker.md <<'EOF'
  # 60号计划执行跟踪

  **启动日期**: 2025-10-10
  **当前阶段**: 阶段零（启动准备）
  **预计完成**: 2025-12-20（10周）

  ## 进度看板

  ### 阶段零：启动准备（3-5天）
  - [ ] Step 0.1: 计划文档正式化
  - [ ] Step 0.2: 组建跨团队小组
  - [ ] Step 0.3: 评估前置条件
  - [ ] Step 0.4: 建立迭代跟踪

  ### 第一阶段：契约与类型统一（2周）
  - [ ] Week 1: 契约同步脚本开发
  - [ ] Week 2: 代码生成与集成

  ### 第二阶段：后端服务与中间件收敛（3周）
  - [ ] 待启动

  ### 第三阶段：前端 API/Hooks/配置整治（2-3周）
  - [ ] 待启动

  ### 第四阶段：工具与验证体系巩固（1-2周）
  - [ ] 待启动

  ## 本周进展（Week 41, 2025-10-10）

  ### 已完成
  - 创建 60 号计划文档 v1.1
  - 创建 61 号执行计划

  ### 进行中
  - 组建跨团队小组
  - 评估前置条件

  ### 下周计划
  - 搭建契约脚本框架
  - 实现 OpenAPI 解析器

  ## 风险与问题日志

  | ID | 风险/问题 | 影响 | 状态 | 负责人 | 应对措施 |
  |----|----------|------|------|--------|---------|
  | R01 | 契约脚本开发延期 | 中 | 监控中 | _______ | 保留人工校对备选 |

  ## 变更记录

  - 2025-10-10: 初始化跟踪文档
  EOF

  git add docs/development-plans/60-execution-tracker.md
  git commit -m "docs: 初始化60号计划执行跟踪看板

  ref: plan-60"
  ```

- [ ] （可选）在项目管理工具中创建任务
  - 创建 Epic: "60号系统级质量重构"
  - 创建 4 个 Story（对应四个阶段）
  - 为第一阶段创建详细 Task

**验收标准**:
- [ ] 跟踪文档已创建并提交
- [ ] 进度看板初始化完成
- [ ] （可选）项目管理工具任务已创建

---

### 阶段零验收（Day 3-5）

**验收会议**: 由架构组负责人主持，所有阶段责任人参加

**验收清单**:
- [ ] 60 号计划文档已合并主干
- [ ] 61 号执行计划已合并主干
- [ ] 跨团队小组已组建，责任人明确
- [ ] 前置条件已评估，无阻塞项
- [ ] 实现清单基线已记录
- [ ] 执行跟踪机制已建立

**输出物**:
- [ ] 阶段零验收报告（简短邮件或会议纪要）
- [ ] 更新 `60-execution-tracker.md` 状态

**通过标准**: 所有清单项 ✓，可进入第一阶段

---

## 第一阶段：契约与类型统一（Week 1-2）

### Week 1: 契约同步脚本开发

#### Day 1-2: 搭建脚本框架

**执行时间**: 第一阶段 Week 1, Day 1-2
**负责人**: 第一阶段责任人

**任务清单**:
- [ ] 创建目录结构
  ```bash
  cd /home/shangmeilin/cube-castle
  mkdir -p scripts/contract
  mkdir -p shared/contracts
  mkdir -p tests/contract

  # 创建主同步脚本
  cat > scripts/contract/sync.sh <<'EOF'
  #!/bin/bash
  # 契约同步主脚本
  # 用途：从 OpenAPI/GraphQL 契约生成统一中间层与 Go/TS 类型
  # 维护：架构组

  set -e

  PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
  cd "$PROJECT_ROOT"

  echo "📋 [契约同步] 开始..."
  echo "  工作目录: $PROJECT_ROOT"

  # 1. 从 OpenAPI 提取契约
  echo "  → 提取 OpenAPI 契约..."
  node scripts/contract/openapi-to-json.js

  # 2. 从 GraphQL 提取契约
  echo "  → 提取 GraphQL 契约..."
  node scripts/contract/graphql-to-json.js

  # 3. 生成 Go 类型
  echo "  → 生成 Go 类型..."
  node scripts/contract/generate-go-types.js

  # 4. 生成 TypeScript 类型
  echo "  → 生成 TypeScript 类型..."
  node scripts/contract/generate-ts-types.js

  echo "✅ [契约同步] 完成"
  echo "  输出文件:"
  echo "    - shared/contracts/organization.json"
  echo "    - cmd/organization-command-service/internal/types/contract_gen.go"
  echo "    - frontend/src/shared/types/contract_gen.ts"
  EOF

  chmod +x scripts/contract/sync.sh
  ```

- [ ] 创建辅助脚本占位文件
  ```bash
  touch scripts/contract/openapi-to-json.js
  touch scripts/contract/graphql-to-json.js
  touch scripts/contract/generate-go-types.js
  touch scripts/contract/generate-ts-types.js

  # 添加基础注释
  for file in scripts/contract/*.js; do
    cat > "$file" <<EOF
  #!/usr/bin/env node
  // $(basename "$file")
  // 用途：[待实现]
  // 维护：架构组

  console.log('[TODO] $(basename "$file") 待实现');
  EOF
  done

  chmod +x scripts/contract/*.js
  ```

- [ ] 测试框架可执行性
  ```bash
  bash scripts/contract/sync.sh
  # 预期输出：所有子脚本输出 [TODO] 待实现
  ```

- [ ] 提交框架代码
  ```bash
  git add scripts/contract/ shared/contracts/ tests/contract/
  git commit -m "feat(contract): 建立契约同步脚本框架

  - 创建主同步脚本 sync.sh
  - 建立 OpenAPI/GraphQL 解析器占位
  - 建立 Go/TS 代码生成器占位
  - 准备测试目录结构

  ref: plan-60 stage-1"
  ```

**验收标准**:
- [ ] 目录结构已创建
- [ ] `sync.sh` 可执行且逻辑清晰
- [ ] 子脚本占位文件已创建
- [ ] 框架代码已提交 Git

---

#### Day 3-4: 实现 OpenAPI 解析器

**执行时间**: 第一阶段 Week 1, Day 3-4
**负责人**: 第一阶段责任人

**任务清单**:
- [ ] 安装依赖
  ```bash
  cd /home/shangmeilin/cube-castle
  npm install --save-dev js-yaml
  ```

- [ ] 实现 OpenAPI 解析器
  ```javascript
  // scripts/contract/openapi-to-json.js
  #!/usr/bin/env node
  const yaml = require('js-yaml');
  const fs = require('fs');
  const path = require('path');

  const PROJECT_ROOT = path.resolve(__dirname, '../..');
  const OPENAPI_PATH = path.join(PROJECT_ROOT, 'docs/api/openapi.yaml');
  const OUTPUT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');

  console.log('[OpenAPI] 解析契约...');

  try {
    // 读取 OpenAPI 规范
    const openapi = yaml.load(fs.readFileSync(OPENAPI_PATH, 'utf8'));

    // 提取枚举
    const schemas = openapi.components.schemas;
    const organizationUnit = schemas.OrganizationUnit || {};
    const properties = organizationUnit.properties || {};

    const contract = {
      version: '1.0.0',
      source: 'openapi',
      timestamp: new Date().toISOString(),
      enums: {
        UnitType: properties.unitType?.enum || [],
        Status: properties.status?.enum || []
      },
      constraints: {
        hierarchy: {
          maxLevel: 17,
          description: '组织层级上限'
        },
        name: {
          maxLength: properties.name?.maxLength || 100,
          pattern: properties.name?.pattern || ''
        },
        code: {
          pattern: properties.code?.pattern || ''
        }
      }
    };

    // 确保输出目录存在
    const outputDir = path.dirname(OUTPUT_PATH);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    // 写入文件
    fs.writeFileSync(OUTPUT_PATH, JSON.stringify(contract, null, 2));

    console.log('[OpenAPI] ✓ 契约已提取');
    console.log(`  → ${OUTPUT_PATH}`);
    console.log(`  → UnitType: ${contract.enums.UnitType.length} 个枚举值`);
    console.log(`  → Status: ${contract.enums.Status.length} 个枚举值`);

  } catch (error) {
    console.error('[OpenAPI] ✗ 解析失败:', error.message);
    process.exit(1);
  }
  ```

- [ ] 测试 OpenAPI 解析器
  ```bash
  node scripts/contract/openapi-to-json.js
  cat shared/contracts/organization.json
  # 验证输出格式正确
  ```

- [ ] 提交实现代码
  ```bash
  git add scripts/contract/openapi-to-json.js shared/contracts/organization.json package.json
  git commit -m "feat(contract): 实现 OpenAPI 契约解析器

  - 从 openapi.yaml 提取 UnitType/Status 枚举
  - 提取组织层级约束（maxLevel: 17）
  - 提取字段校验规则（name/code pattern）
  - 输出统一中间契约文件

  ref: plan-60 stage-1"
  ```

**验收标准**:
- [ ] `openapi-to-json.js` 执行成功
- [ ] `organization.json` 包含正确枚举与约束
- [ ] 输出格式规范（带时间戳、版本号）
- [ ] 代码已提交 Git

---

#### Day 5: 实现 GraphQL 解析器 + 人工验收

**执行时间**: 第一阶段 Week 1, Day 5
**负责人**: 第一阶段责任人

**任务清单**:
- [ ] 实现 GraphQL 解析器
  ```javascript
  // scripts/contract/graphql-to-json.js
  #!/usr/bin/env node
  const fs = require('fs');
  const path = require('path');

  const PROJECT_ROOT = path.resolve(__dirname, '../..');
  const GRAPHQL_SCHEMA_PATH = path.join(PROJECT_ROOT, 'docs/api/schema.graphql');
  const CONTRACT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');

  console.log('[GraphQL] 解析 Schema...');

  try {
    const schema = fs.readFileSync(GRAPHQL_SCHEMA_PATH, 'utf8');

    // 提取 UnitType 枚举（简单正则匹配）
    const unitTypeMatch = schema.match(/enum UnitType \{([^}]+)\}/);
    const unitTypeValues = unitTypeMatch
      ? unitTypeMatch[1].trim().split(/\s+/).filter(v => v && !v.startsWith('#'))
      : [];

    // 提取 Status 枚举
    const statusMatch = schema.match(/enum Status \{([^}]+)\}/);
    const statusValues = statusMatch
      ? statusMatch[1].trim().split(/\s+/).filter(v => v && !v.startsWith('#'))
      : [];

    // 读取现有契约
    let contract = {};
    if (fs.existsSync(CONTRACT_PATH)) {
      contract = JSON.parse(fs.readFileSync(CONTRACT_PATH, 'utf8'));
    }

    // 合并 GraphQL 信息
    contract.graphql = {
      source: 'schema.graphql',
      timestamp: new Date().toISOString(),
      enums: {
        UnitType: unitTypeValues,
        Status: statusValues
      }
    };

    // 写回文件
    fs.writeFileSync(CONTRACT_PATH, JSON.stringify(contract, null, 2));

    console.log('[GraphQL] ✓ Schema 已解析');
    console.log(`  → UnitType: ${unitTypeValues.length} 个枚举值`);
    console.log(`  → Status: ${statusValues.length} 个枚举值`);

    // 一致性检查
    if (contract.enums) {
      const openApiUnitType = contract.enums.UnitType || [];
      const graphqlUnitType = unitTypeValues;

      if (JSON.stringify(openApiUnitType.sort()) !== JSON.stringify(graphqlUnitType.sort())) {
        console.warn('[GraphQL] ⚠ UnitType 枚举与 OpenAPI 不一致');
        console.warn(`  OpenAPI: ${openApiUnitType.join(', ')}`);
        console.warn(`  GraphQL: ${graphqlUnitType.join(', ')}`);
      } else {
        console.log('[GraphQL] ✓ UnitType 枚举与 OpenAPI 一致');
      }
    }

  } catch (error) {
    console.error('[GraphQL] ✗ 解析失败:', error.message);
    process.exit(1);
  }
  ```

- [ ] 测试 GraphQL 解析器
  ```bash
  node scripts/contract/graphql-to-json.js
  cat shared/contracts/organization.json | jq .
  ```

- [ ] 人工验收契约文件
  ```bash
  # 验收检查清单
  echo "## 契约文件人工验收"
  echo "1. UnitType 枚举值是否完整？"
  cat shared/contracts/organization.json | jq '.enums.UnitType'

  echo "2. Status 枚举值是否完整？"
  cat shared/contracts/organization.json | jq '.enums.Status'

  echo "3. GraphQL 与 OpenAPI 枚举是否一致？"
  cat shared/contracts/organization.json | jq '{openapi: .enums, graphql: .graphql.enums}'

  echo "4. 约束条件是否正确？"
  cat shared/contracts/organization.json | jq '.constraints'
  ```

- [ ] 提交验收通过的代码
  ```bash
  git add scripts/contract/graphql-to-json.js shared/contracts/organization.json
  git commit -m "feat(contract): 实现 GraphQL Schema 解析器

  - 从 schema.graphql 提取枚举定义
  - 与 OpenAPI 契约合并到统一文件
  - 增加跨源一致性检查
  - 人工验收通过

  ref: plan-60 stage-1"
  ```

**验收标准**:
- [ ] GraphQL 解析器执行成功
- [ ] 枚举一致性检查通过
- [ ] 人工验收检查清单全部 ✓
- [ ] 代码已提交 Git

---

### Week 2: 代码生成与集成

#### Day 6-7: Go 类型生成器

**执行时间**: 第一阶段 Week 2, Day 6-7
**负责人**: 第一阶段责任人

**任务清单**:
- [ ] 实现 Go 类型生成器
  ```javascript
  // scripts/contract/generate-go-types.js
  #!/usr/bin/env node
  const fs = require('fs');
  const path = require('path');

  const PROJECT_ROOT = path.resolve(__dirname, '../..');
  const CONTRACT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');
  const OUTPUT_PATH = path.join(PROJECT_ROOT, 'cmd/organization-command-service/internal/types/contract_gen.go');

  console.log('[Go] 生成类型定义...');

  try {
    const contract = JSON.parse(fs.readFileSync(CONTRACT_PATH, 'utf8'));

    // 生成 Go 代码
    const goCode = `// Code generated by scripts/contract/generate-go-types.js. DO NOT EDIT.
  // Source: shared/contracts/organization.json
  // Generated: ${new Date().toISOString()}

  package types

  // UnitType 组织单元类型（契约生成）
  type UnitType string

  const (
  ${contract.enums.UnitType.map((v, i) => {
    const constName = v.charAt(0) + v.slice(1).toLowerCase().replace(/_([a-z])/g, (_, c) => c.toUpperCase());
    return `\tUnitType${constName} UnitType = "${v}"`;
  }).join('\n')}
  )

  // Status 组织状态（契约生成）
  type Status string

  const (
  ${contract.enums.Status.map(v => {
    const constName = v.charAt(0) + v.slice(1).toLowerCase().replace(/_([a-z])/g, (_, c) => c.toUpperCase());
    return `\tStatus${constName} Status = "${v}"`;
  }).join('\n')}
  )

  // OrganizationConstraints 组织约束（契约生成）
  const (
  	// MaxOrganizationLevel 组织层级上限
  	MaxOrganizationLevel = ${contract.constraints.hierarchy.maxLevel}

  	// MaxOrganizationNameLength 组织名称最大长度
  	MaxOrganizationNameLength = ${contract.constraints.name.maxLength}
  )
  `;

    // 确保输出目录存在
    const outputDir = path.dirname(OUTPUT_PATH);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    fs.writeFileSync(OUTPUT_PATH, goCode);

    console.log('[Go] ✓ 类型已生成');
    console.log(`  → ${OUTPUT_PATH}`);

  } catch (error) {
    console.error('[Go] ✗ 生成失败:', error.message);
    process.exit(1);
  }
  ```

- [ ] 测试 Go 代码生成
  ```bash
  node scripts/contract/generate-go-types.js
  cat cmd/organization-command-service/internal/types/contract_gen.go
  ```

- [ ] 验证 Go 代码编译
  ```bash
  cd cmd/organization-command-service
  go build ./internal/types
  # 确保编译通过
  ```

- [ ] 提交生成器代码
  ```bash
  git add scripts/contract/generate-go-types.js \
         cmd/organization-command-service/internal/types/contract_gen.go
  git commit -m "feat(contract): 实现 Go 类型生成器

  - 从契约文件生成 UnitType/Status 枚举
  - 生成组织约束常量（MaxLevel 等）
  - 添加代码生成标记（DO NOT EDIT）
  - Go 编译验证通过

  ref: plan-60 stage-1"
  ```

**验收标准**:
- [ ] Go 类型生成器执行成功
- [ ] 生成的 Go 代码编译通过
- [ ] 枚举值与契约一致
- [ ] 代码已提交 Git

---

#### Day 8-9: TypeScript 类型生成器

**执行时间**: 第一阶段 Week 2, Day 8-9
**负责人**: 第一阶段责任人

**任务清单**:
- [ ] 实现 TypeScript 类型生成器
  ```javascript
  // scripts/contract/generate-ts-types.js
  #!/usr/bin/env node
  const fs = require('fs');
  const path = require('path');

  const PROJECT_ROOT = path.resolve(__dirname, '../..');
  const CONTRACT_PATH = path.join(PROJECT_ROOT, 'shared/contracts/organization.json');
  const OUTPUT_PATH = path.join(PROJECT_ROOT, 'frontend/src/shared/types/contract_gen.ts');

  console.log('[TypeScript] 生成类型定义...');

  try {
    const contract = JSON.parse(fs.readFileSync(CONTRACT_PATH, 'utf8'));

    // 生成 TypeScript 代码
    const tsCode = `// Code generated by scripts/contract/generate-ts-types.js. DO NOT EDIT.
  // Source: shared/contracts/organization.json
  // Generated: ${new Date().toISOString()}

  /**
   * 组织单元类型（契约生成）
   */
  export enum UnitType {
  ${contract.enums.UnitType.map(v => {
    const enumKey = v.charAt(0) + v.slice(1).toLowerCase().replace(/_([a-z])/g, (_, c) => c.toUpperCase());
    return `  ${enumKey} = '${v}',`;
  }).join('\n')}
  }

  /**
   * 组织状态（契约生成）
   */
  export enum Status {
  ${contract.enums.Status.map(v => {
    const enumKey = v.charAt(0) + v.slice(1).toLowerCase().replace(/_([a-z])/g, (_, c) => c.toUpperCase());
    return `  ${enumKey} = '${v}',`;
  }).join('\n')}
  }

  /**
   * 组织约束常量（契约生成）
   */
  export const OrganizationConstraints = {
    /** 组织层级上限 */
    MAX_LEVEL: ${contract.constraints.hierarchy.maxLevel},

    /** 组织名称最大长度 */
    MAX_NAME_LENGTH: ${contract.constraints.name.maxLength},
  } as const;

  /**
   * UnitType 类型守卫
   */
  export function isUnitType(value: unknown): value is UnitType {
    return typeof value === 'string' && Object.values(UnitType).includes(value as UnitType);
  }

  /**
   * Status 类型守卫
   */
  export function isStatus(value: unknown): value is Status {
    return typeof value === 'string' && Object.values(Status).includes(value as Status);
  }
  `;

    // 确保输出目录存在
    const outputDir = path.dirname(OUTPUT_PATH);
    if (!fs.existsSync(outputDir)) {
      fs.mkdirSync(outputDir, { recursive: true });
    }

    fs.writeFileSync(OUTPUT_PATH, tsCode);

    console.log('[TypeScript] ✓ 类型已生成');
    console.log(`  → ${OUTPUT_PATH}`);

  } catch (error) {
    console.error('[TypeScript] ✗ 生成失败:', error.message);
    process.exit(1);
  }
  ```

- [ ] 测试 TypeScript 代码生成
  ```bash
  node scripts/contract/generate-ts-types.js
  cat frontend/src/shared/types/contract_gen.ts
  ```

- [ ] 验证 TypeScript 编译
  ```bash
  cd frontend
  npm run typecheck
  # 确保无类型错误
  ```

- [ ] 更新现有代码引用生成类型（示例）
  ```typescript
  // frontend/src/shared/types/organization.ts
  // 添加导入
  import { UnitType, Status } from './contract_gen';

  // 将手动枚举替换为引用生成类型
  // export type UnitType = 'COMPANY' | 'DEPARTMENT' | ...;  // ← 删除
  // 改为使用 import 的 UnitType
  ```

- [ ] 提交生成器代码
  ```bash
  git add scripts/contract/generate-ts-types.js \
         frontend/src/shared/types/contract_gen.ts
  git commit -m "feat(contract): 实现 TypeScript 类型生成器

  - 从契约文件生成 UnitType/Status 枚举
  - 生成组织约束常量
  - 提供类型守卫函数
  - TypeScript 编译验证通过

  ref: plan-60 stage-1"
  ```

**验收标准**:
- [ ] TypeScript 类型生成器执行成功
- [ ] 生成的 TS 代码编译通过
- [ ] 枚举值与契约一致
- [ ] 代码已提交 Git

---

#### Day 10: CI 集成与第一阶段验收

**执行时间**: 第一阶段 Week 2, Day 10
**负责人**: 第一阶段责任人 + 平台团队

**任务清单**:
- [ ] 创建 CI 工作流文件
  ```yaml
  # .github/workflows/contract-sync.yml
  name: Contract Sync Check

  on:
    pull_request:
      paths:
        - 'docs/api/openapi.yaml'
        - 'docs/api/schema.graphql'
        - 'scripts/contract/**'
    push:
      branches:
        - master

  jobs:
    contract-sync:
      runs-on: ubuntu-latest

      steps:
        - name: Checkout code
          uses: actions/checkout@v3

        - name: Setup Node.js
          uses: actions/setup-node@v3
          with:
            node-version: '18'

        - name: Install dependencies
          run: npm install --save-dev js-yaml

        - name: Run contract sync
          run: bash scripts/contract/sync.sh

        - name: Check for uncommitted changes
          run: |
            if ! git diff --exit-code shared/contracts/ \
              cmd/organization-command-service/internal/types/contract_gen.go \
              frontend/src/shared/types/contract_gen.ts; then
              echo "❌ 契约文件有未提交的变更，请运行 scripts/contract/sync.sh 并提交"
              exit 1
            fi
            echo "✅ 契约文件与仓库一致"
  ```

- [ ] 测试 CI 工作流
  ```bash
  # 本地模拟 CI 执行
  bash scripts/contract/sync.sh
  git diff --exit-code shared/contracts/ \
    cmd/organization-command-service/internal/types/contract_gen.go \
    frontend/src/shared/types/contract_gen.ts
  # 确保无差异
  ```

- [ ] 提交 CI 配置
  ```bash
  git add .github/workflows/contract-sync.yml
  git commit -m "ci: 添加契约同步检查工作流

  - 监控 OpenAPI/GraphQL 契约变更
  - 自动执行契约同步脚本
  - 验证生成文件与仓库一致性
  - 阻止不同步的代码合并

  ref: plan-60 stage-1"
  git push origin master
  ```

- [ ] 执行第一阶段验收
  **验收会议**: 第一阶段责任人主持，架构组参与

  **验收清单**:
  - [ ] 契约同步脚本 `sync.sh` 执行成功
  - [ ] `organization.json` 包含正确枚举与约束
  - [ ] Go 生成代码 `contract_gen.go` 编译通过
  - [ ] TS 生成代码 `contract_gen.ts` 编译通过
  - [ ] CI Job `contract-sync` 绿灯
  - [ ] 运行实现清单对比基线
    ```bash
    node scripts/generate-implementation-inventory.js > .after-stage1.md
    diff .baseline-before-refactor.md .after-stage1.md
    # 确认新增了契约相关实现，无重复
    ```
  - [ ] 更新 `docs/reference/` 相关表格（如有需要）

- [ ] 输出第一阶段验收报告
  ```markdown
  # 第一阶段验收报告

  **阶段**: 契约与类型统一
  **完成日期**: 2025-10-XX
  **负责人**: ________
  **状态**: ✅ 通过

  ## 交付物
  - [x] 契约同步脚本体系（`scripts/contract/`）
  - [x] 统一契约文件（`shared/contracts/organization.json`）
  - [x] Go 类型生成代码（`contract_gen.go`）
  - [x] TypeScript 类型生成代码（`contract_gen.ts`）
  - [x] CI 契约检查工作流

  ## 关键指标
  - 契约枚举值数量：UnitType 5个，Status 4个
  - 组织层级上限：17层
  - Go 编译：✅ 通过
  - TS 编译：✅ 通过
  - CI 状态：✅ 绿灯

  ## 风险与问题
  - 无

  ## 下一步
  - 进入第二阶段：后端服务与中间件收敛
  - 预计启动时间：2025-10-XX
  ```

- [ ] 更新执行跟踪文档
  ```bash
  # 在 60-execution-tracker.md 中标记第一阶段完成
  # 更新进度看板
  ```

**验收标准**:
- [ ] CI 工作流已配置并测试通过
- [ ] 第一阶段所有验收清单项 ✓
- [ ] 验收报告已输出
- [ ] 执行跟踪文档已更新

---

## 后续阶段时间线（概览）

### 第二阶段：后端服务与中间件收敛（Week 3-5）

**关键里程碑**:
- **Week 3**: 抽取共享事务与审计封装，实现双写+比对日志
- **Week 4**: 定义统一响应/错误结构，制定 Dev/Operational 白名单
- **Week 5**: 集成 Prometheus/Otel 中间件，灰度验证

**输出物**:
- `internal/services/temporal_transaction.go` 共享封装
- 统一响应/错误结构体
- Dev/Operational 白名单配置
- Prometheus 指标定义

**验收标准**:
- 双写期间新旧数据 diff = 0
- Prometheus 延迟 < 200ms
- 安全测试通过

**详细执行计划**: 待第一阶段验收通过后制定

---

### 第三阶段：前端 API/Hooks/配置整治（Week 6-8）

**关键里程碑**:
- **Week 6**: 统一 React Query 客户端，建立标准错误包装
- **Week 7**: Hooks 迁移（先查询后写操作），临时桥接层
- **Week 8**: 端口/环境助手重写，QA 关键路径巡检

**输出物**:
- `shared/api/queryClient.ts` 统一客户端
- 重构后的 Hooks（`useOrganizationsQuery` 等）
- `legacyOrganizationApi` 桥接层
- 新的端口/环境助手

**验收标准**:
- Vitest 覆盖率 ≥ 75%
- Playwright 冒烟场景全绿
- 运行时代码包体积下降 ≥ 5%

**详细执行计划**: 待第二阶段验收通过后制定

---

### 第四阶段：工具与验证体系巩固（Week 9-10）

**关键里程碑**:
- **Week 9**: Temporal/Validation 工具折叠，审计字段完善
- **Week 10**: 新增 CI 守护任务，最终验收

**输出物**:
- 单一 Temporal/Validation 实现
- 结构化审计 DTO
- CI 新增 `lint-contract`、`lint-audit`、`doc-archive-check`

**验收标准**:
- 审计记录含完整字段
- CI 守护任务全绿
- 所有旧别名标记废弃

**详细执行计划**: 待第三阶段验收通过后制定

---

## 风险与问题应对

### 风险矩阵

| ID | 风险描述 | 概率 | 影响 | 应对措施 | 负责人 |
|----|---------|------|------|---------|--------|
| R01 | 契约脚本开发延期 | 中 | 中 | 保留人工校对备选方案 | 第一阶段负责人 |
| R02 | 跨团队协调困难 | 中 | 高 | 建立异步沟通机制 | 项目经理 |
| R03 | staging 环境不稳定 | 低 | 高 | 提前申请专用环境 | 平台团队 |
| R04 | 生产环境影响 | 低 | 极高 | 严格执行阶段门禁 | 所有阶段负责人 |
| R05 | 关键人员离职 | 低 | 高 | 文档化所有关键知识 | 架构组 |

### 问题升级机制

**Level 1 - 团队内解决**:
- 执行过程中的技术问题
- 由阶段负责人协调解决
- 解决时限：2 个工作日

**Level 2 - 跨团队协调**:
- 需要多团队配合的问题
- 由项目经理协调解决
- 解决时限：5 个工作日

**Level 3 - 架构组决策**:
- 涉及架构变更的问题
- 由架构组负责人决策
- 解决时限：按紧急程度定

**Level 4 - 管理层介入**:
- 资源不足、优先级冲突
- 由管理层介入协调
- 解决时限：按紧急程度定

---

## 资源需求清单

### 人力资源

| 角色 | 投入时间 | 关键阶段 | 技能要求 |
|------|---------|---------|---------|
| 第一阶段负责人 | 100% 2周 | 阶段一 | OpenAPI/GraphQL、Node.js |
| 第二阶段负责人 | 100% 3周 | 阶段二 | Go、Temporal、Prometheus |
| 第三阶段负责人 | 100% 2-3周 | 阶段三 | React、TypeScript、React Query |
| 第四阶段负责人 | 100% 1-2周 | 阶段四 | CI/CD、质量工具 |
| 项目经理 | 20% 10周 | 全程 | 项目管理、协调 |
| QA 工程师 | 50% 每阶段末 | 全程 | 测试、验收 |
| 平台工程师 | 按需 | 阶段一/四 | CI/CD、环境配置 |

### 环境资源

- **开发环境**: 本地开发环境，无额外需求
- **staging 环境**:
  - 需要专用 staging 环境用于灰度验证
  - 需要 Prometheus/Grafana 监控接入
  - 需要 7x24 小时访问权限
- **CI/CD**:
  - GitHub Actions 执行时间配额（预估每次 PR 5-10 分钟）
  - 需要 npm 依赖缓存加速

### 工具资源

- **开发工具**:
  - Node.js 18+
  - Go 1.21+
  - npm/yarn
  - Git

- **依赖包**:
  - `js-yaml`（已在 package.json）
  - 其他无额外依赖

---

## 沟通与报告机制

### 定期会议

| 会议名称 | 频率 | 参与人 | 时长 | 议程 |
|---------|------|-------|------|------|
| 双周同步会 | 每两周 | 所有阶段负责人 + 项目经理 | 1小时 | 进度汇报、风险讨论、下周计划 |
| 阶段验收会 | 每阶段末 | 责任人 + 架构组 + QA | 1.5小时 | 验收清单复核、问题总结 |
| 临时问题会 | 按需 | 相关方 | 30分钟 | 问题升级、决策 |

### 进度报告

**周报**（由各阶段负责人提交）:
- 本周完成任务
- 下周计划任务
- 风险与问题
- 需要的支持

**月报**（由项目经理汇总）:
- 整体进度概览
- 关键里程碑状态
- 资源使用情况
- 风险趋势分析

### 文档更新

- **执行跟踪文档** (`60-execution-tracker.md`): 每周更新
- **验收报告**: 每阶段末输出
- **FAQ 文档**: 遇到问题时及时补充

---

## 附录

### A. 快速参考命令

> 下列命令用于参考演练，实际执行时请根据当期分支与流程酌情取舍。

```bash
# 阶段零启动
git add docs/development-plans/60-*.md docs/development-plans/61-*.md
git commit -m "docs: 启动60号质量重构计划"
node scripts/generate-implementation-inventory.js > .baseline-before-refactor.md

# 第一阶段开发
mkdir -p scripts/contract shared/contracts tests/contract
bash scripts/contract/sync.sh
node scripts/contract/openapi-to-json.js
node scripts/contract/generate-go-types.js
node scripts/contract/generate-ts-types.js

# 验证
cd cmd/organization-command-service && go build ./internal/types
cd frontend && npm run typecheck

# 提交
git add scripts/contract/ shared/contracts/ \
  cmd/organization-command-service/internal/types/contract_gen.go \
  frontend/src/shared/types/contract_gen.ts
git commit -m "feat(contract): 第一阶段完成"
```

### B. 相关文档索引

- **60号文档**: [系统级质量整合与重构计划](./60-system-wide-quality-refactor-plan.md)
- **执行跟踪**: [60号执行跟踪](./60-execution-tracker.md)（待创建）
- **质量分析**: [50-59号文档](./00-README.md)
- **开发者手册**: [docs/reference/01-DEVELOPER-QUICK-REFERENCE.md](../reference/01-DEVELOPER-QUICK-REFERENCE.md)
- **API 契约**: [docs/api/openapi.yaml](../api/openapi.yaml), [docs/api/schema.graphql](../api/schema.graphql)

### C. 联系人清单

| 角色 | 姓名 | 邮箱 | 职责 |
|------|------|------|------|
| 架构组负责人 | _______ | _______@example.com | 技术决策、验收 |
| 项目经理 | _______ | _______@example.com | 进度协调、资源分配 |
| 第一阶段负责人 | _______ | _______@example.com | 契约与类型统一 |
| 第二阶段负责人 | _______ | _______@example.com | 后端服务整合 |
| 第三阶段负责人 | _______ | _______@example.com | 前端整治 |
| 第四阶段负责人 | _______ | _______@example.com | 工具巩固 |
| 平台团队 Lead | _______ | _______@example.com | CI/CD、环境支持 |
| QA Lead | _______ | _______@example.com | 测试、验收 |

---

**最后更新**: 2025-10-10
**下次评审**: 阶段一验收后
**文档状态**: 执行中
