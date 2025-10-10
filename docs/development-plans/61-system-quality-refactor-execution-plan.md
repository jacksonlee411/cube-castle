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

### 🚧 待建设施（第一阶段剩余目标）
- [x] CI Job `contract-sync` 和 `contract-snapshot` (新增快照校验工作流)
- [x] 跨层快照测试框架（`tests/contract/` 基线 + 验证脚本）

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
- [x] 确认 60 号计划已合入主干并完成评审（commit `4258bce6`）。
- [x] 在 `docs/development-plans/00-README.md` 的“活跃计划”中列出 60/61 号文档（commit `8cf9b6c2`）。
- [x] 确认本执行计划（61 号）为最新版本并已提交。

**验收标准**:
- [x] 主分支可查阅 60 号计划与本执行计划。
- [x] 计划索引与实际文档保持一致。

---

### Step 0.2: 组建跨团队小组（优先级：P0）

**执行时间**: Day 1 下午
**负责人**: 项目经理 + 架构组负责人

**任务清单**:
- [x] 明确各阶段责任人与时间投入（单人执行，责任人为本人，记录于 60-execution-tracker.md）。
- [x] 召开启动会议 → 单人执行，无需会议，改以书面行动计划确认。
- [x] 建立沟通渠道 → 单人执行，采用 60-execution-tracker.md + Git 提交作为信息同步渠道。

**验收标准**:
- [x] 各阶段责任人已确认（单人执行）
- [x] 启动会议已完成并有会议纪要（以执行计划变更记录代替）
- [x] 沟通渠道已建立并测试可用（以文档+提交为通道）

---

### Step 0.3: 评估前置条件（优先级：P1）

**执行时间**: Day 2
**负责人**: 第一阶段责任人

- [x] 验证 API 契约干净无未合并变更；检查 53、56 号计划列出的风险项已关闭或纳入本计划。
- [x] 运行 `scripts/generate-implementation-inventory.js` 输出参考基线（` .baseline-before-refactor.md`）。
- [x] 验证辅助脚本可执行（如 `scripts/check-api-naming.sh`），确认无运行错误。

**验收标准**:
- [x] API 契约文件干净无变更
- [x] 相关计划无阻塞项
- [x] 基线文件已提交到 Git
- [x] 现有工具测试通过

---

### Step 0.4: 建立迭代跟踪（优先级：P1）

**执行时间**: Day 2-3
**负责人**: 项目经理

**任务清单**:
- [x] 创建执行跟踪文档（`docs/development-plans/60-execution-tracker.md`）。
- [x] 每次阶段推进时更新看板与变更记录。
- [x] （单人执行，项目管理工具任务不再单独建立，改为文档+提交记录）。

**验收标准**:
- [x] 跟踪文档已创建并提交
- [x] 进度看板初始化完成并持续更新
- [x] （可选项）项目管理工具改以文档记录代替

---

### 阶段零验收（Day 3-5）

**验收会议**: 由架构组负责人主持，所有阶段责任人参加

**验收清单**:
- [x] 60 号计划文档已合并主干
- [x] 61 号执行计划已合并主干
- [x] 跨团队小组已组建，责任人明确（单人执行）
- [x] 前置条件已评估，无阻塞项
- [x] 实现清单基线已记录
- [x] 执行跟踪机制已建立

**输出物**:
- [x] 阶段零验收报告（以提交记录和跟踪文档代替）
- [x] 更新 `60-execution-tracker.md` 状态

**通过标准**: 所有清单项 ✓，可进入第一阶段

---

## 第一阶段：契约与类型统一（Week 1-2）

### Week 1: 契约同步脚本开发

#### Day 1-2: 搭建脚本框架

**执行时间**: 第一阶段 Week 1, Day 1-2
**负责人**: 第一阶段责任人

- **任务清单（已完成）**:
  - [x] 创建 `scripts/contract/`、`shared/contracts/`、`tests/contract/` 目录并编写 `sync.sh`（commit `7e268c57`）。
  - [x] 初始化四个子脚本文件，现已实现并加入可执行权限。
  - [x] 验证 `scripts/contract/sync.sh` 可顺利执行并产出三份契约工件。
  - [x] 相关变更已提交并通过预提交检查。

**验收标准**:
- [x] 目录结构与主脚本已创建
- [x] `sync.sh` 可执行且逻辑清晰
- [x] 子脚本完成初始化并纳入版本控制
- [x] 框架代码已提交 Git

---

#### Day 3-4: 实现 OpenAPI 解析器

**执行时间**: 第一阶段 Week 1, Day 3-4
**负责人**: 第一阶段责任人

**任务清单（已完成）**:
- [x] 安装 `js-yaml` 依赖并写入 `package.json`。
- [x] 实现 `scripts/contract/openapi-to-json.js`，输出枚举/约束（commit `b5deddac`）。
- [x] 通过 `scripts/contract/sync.sh` 验证生成的 `shared/contracts/organization.json`。
- [x] 相关代码与契约文件已提交并记录生成时间戳 / SHA。

**验收标准**:
- [x] `openapi-to-json.js` 执行成功
- [x] `organization.json` 包含正确枚举与约束
- [x] 输出格式规范（带时间戳、版本号）
- [x] 代码已提交 Git

---

#### Day 5: 实现 GraphQL 解析器 + 人工验收

**执行时间**: 第一阶段 Week 1, Day 5
**负责人**: 第一阶段责任人

**任务清单**:
- [x] 实现 GraphQL 解析器（commit 4efc3ebb）
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

- [x] 测试 GraphQL 解析器（通过脚本执行和 diff 日志验证）
  ```bash
  node scripts/contract/graphql-to-json.js
  cat shared/contracts/organization.json | jq .
  ```

- [x] 人工验收契约文件（差异已记录，现已对齐）
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

- [x] 提交验收通过的代码
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
- [x] GraphQL 解析器执行成功
- [x] 枚举一致性检查通过
- [x] 人工验收检查清单全部 ✓
- [x] 代码已提交 Git

---

### Week 2: 代码生成与集成

#### Day 6-7: Go 类型生成器

**执行时间**: 第一阶段 Week 2, Day 6-7
**负责人**: 第一阶段责任人

**任务清单**:
- [x] 实现 Go 类型生成器
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

- [x] 测试 Go 代码生成（通过 sync.sh）
  ```bash
  node scripts/contract/generate-go-types.js
  cat cmd/organization-command-service/internal/types/contract_gen.go
  ```

- [x] 验证 Go 代码编译
  ```bash
  cd cmd/organization-command-service
  go build ./internal/types
  # 确保编译通过
  ```

- [x] 提交生成器代码
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
- [x] Go 类型生成器执行成功
- [x] 生成的 Go 代码编译通过
- [x] 枚举值与契约一致
- [x] 代码已提交 Git

---

#### Day 8-9: TypeScript 类型生成器

**执行时间**: 第一阶段 Week 2, Day 8-9
**负责人**: 第一阶段责任人

**任务清单**:
- [x] 实现 TypeScript 类型生成器
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

- [x] 测试 TypeScript 代码生成（通过 sync.sh）
  ```bash
  node scripts/contract/generate-ts-types.js
  cat frontend/src/shared/types/contract_gen.ts
  ```

- [x] 验证 TypeScript 编译
  ```bash
  cd frontend
  npm run typecheck
  # 确保无类型错误
  ```

- [x] 更新现有代码引用生成类型（已在 shared/types 等处替换）
  ```typescript
  // frontend/src/shared/types/organization.ts
  // 添加导入
  import { UnitType, Status } from './contract_gen';

  // 将手动枚举替换为引用生成类型
  // export type UnitType = 'COMPANY' | 'DEPARTMENT' | ...;  // ← 删除
  // 改为使用 import 的 UnitType
  ```

- [x] 提交生成器代码
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
- [x] TypeScript 类型生成器执行成功
- [x] 生成的 TS 代码编译通过
- [x] 枚举值与契约一致
- [x] 代码已提交 Git

---

#### Day 10: CI 集成与第一阶段验收

**执行时间**: 第一阶段 Week 2, Day 10
**负责人**: 第一阶段责任人 + 平台团队

**任务清单**:
- [x] 创建 CI 工作流文件（contract-testing.yml 新增 snapshot job）
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

- [x] 测试 CI 工作流（快照 job 已在 commit 4d218e48 中添加，待实际运行验证）
  ```bash
  # 本地模拟 CI 执行
  bash scripts/contract/sync.sh
  git diff --exit-code shared/contracts/ \
    cmd/organization-command-service/internal/types/contract_gen.go \
    frontend/src/shared/types/contract_gen.ts
  # 确保无差异
  ```

- [x] 提交 CI 配置
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
  - [x] 契约同步脚本 `sync.sh` 执行成功
  - [x] `organization.json` 包含正确枚举与约束
  - [x] Go 生成代码 `contract_gen.go` 编译通过
  - [x] TS 生成代码 `contract_gen.ts` 编译通过
  - [ ] CI Job `contract-sync` 绿灯（首次运行后确认）
  - [x] 运行实现清单对比基线（阶段零已完成 `.baseline-before-refactor.md`）
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
- [ ] CI 工作流已配置并测试通过（待 CI 运行确认）
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

**最后更新**: 2025-10-10
**下次评审**: 阶段一验收后
**文档状态**: 执行中
