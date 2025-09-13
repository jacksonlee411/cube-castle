# Agent Management System

Specialized agent framework for Cube Castle HR System management and development.

## Agent Architecture

### Core Agent Types

#### 1. **Backend Agent** (`backend-agent`)
**Domain**: Go backend services, database operations, CQRS architecture
**Specializations**:
- API endpoint development and maintenance
- Database schema and migration management
- CQRS event handling and command processing
- Neo4j graph database operations
- PostgreSQL relational operations
- Business ID standardization and validation

**Tools**: Go, Ent ORM, PostgreSQL, Neo4j, Kafka, Temporal
**Responsibilities**:
- Employee management API (`go-app/internal/handler/employee_handler.go`)
- Organization unit management (`go-app/internal/handler/organization_unit_handler.go`)
- Position management with CQRS (`go-app/internal/cqrs/`)
- Event sourcing and CDC pipeline (`go-app/internal/events/`)

#### 2. **Frontend Agent** (`frontend-agent`)
**Domain**: Next.js frontend, UI/UX, component development
**Specializations**:
- React component development with TypeScript
- Next.js page routing and SSR optimization
- Ant Design and custom UI component integration
- GraphQL client operations with Apollo
- State management with SWR and Zustand
- Responsive design and mobile optimization

**Tools**: Next.js, React, TypeScript, Ant Design, Apollo GraphQL, SWR
**Responsibilities**:
- Employee management pages (`nextjs-app/src/pages/employees/`)
- Organization chart visualization (`nextjs-app/src/components/business/organization-tree.tsx`)
- Form handling and validation (`nextjs-app/src/components/business/employee-create-dialog.tsx`)
- API integration layers (`nextjs-app/src/lib/api-client.ts`)

#### 3. **AI Agent** (`ai-agent`)
**Domain**: Python AI services, intelligent features
**Specializations**:
- Natural language processing for HR queries
- gRPC service implementation
- Dialogue state management
- Intelligence gateway integration
- ML model integration for HR analytics

**Tools**: Python, gRPC, FastAPI, scikit-learn
**Responsibilities**:
- Intelligence gateway service (`python-ai/main.py`)
- Dialogue state management (`python-ai/dialogue_state.py`)
- AI-powered HR recommendations and insights

#### 4. **DevOps Agent** (`devops-agent`)
**Domain**: Infrastructure, deployment, monitoring
**Specializations**:
- Docker containerization and orchestration
- Multi-service deployment coordination
- Database migration and backup management
- Performance monitoring and logging
- CI/CD pipeline optimization

**Tools**: Docker, Docker Compose, PostgreSQL, Neo4j, Kafka, Temporal
**Responsibilities**:
- Service orchestration (`docker-compose.yml`)
- Database migrations (`go-app/deployments/migrations/`)
- Deployment scripts (`scripts/`)
- Health monitoring and service status

#### 5. **QA Agent** (`qa-agent`)
**Domain**: Testing, quality assurance, validation
**Specializations**:
- E2E testing with Playwright
- Integration testing across services
- Performance testing and benchmarking
- UAT planning and execution
- Test automation framework development

**Tools**: Playwright, Go testing, Jest, Test automation frameworks
**Responsibilities**:
- E2E test suites (`nextjs-app/tests/e2e/`)
- Integration tests (`go-app/tests/`)
- UAT test plans (`docs/testing/`)
- Performance benchmarks

#### 6. **Architecture Agent** (`architecture-agent`)
**Domain**: System design, technical decisions, documentation
**Specializations**:
- System architecture design and review
- API design and specification
- Technical documentation maintenance
- Architecture decision records (ADRs)
- Cross-service integration patterns

**Tools**: OpenAPI, GraphQL schemas, Architecture diagrams
**Responsibilities**:
- API specifications (`docs/api/openapi.yaml` - Organization API)
- Architecture documentation (`docs/architecture/`)
- ADR management (架构决策记录已整合至主要文档系统)
- Integration patterns and best practices

## Agent Communication Protocol

### Command Interface
```bash
# Direct agent invocation
/agent <agent-type> <command> [options]

# Examples:
/agent backend-agent implement-employee-api --with-cqrs
/agent frontend-agent create-component --type=dialog --feature=employee
/agent qa-agent run-e2e-tests --suite=employee-management
/agent devops-agent deploy --environment=staging
```

### Agent Coordination
```yaml
orchestration:
  primary_agent: "Leads the task execution"
  supporting_agents: "Provide specialized assistance"
  coordination_pattern:
    - task_analysis: "Determine required agent types"
    - agent_selection: "Choose primary and supporting agents"
    - task_delegation: "Distribute work based on specialization"
    - progress_sync: "Coordinate progress across agents"
    - result_integration: "Combine outputs from multiple agents"
```

### Inter-Agent Communication
- **Shared Context**: Common understanding of project state
- **Event Broadcasting**: Notify relevant agents of changes
- **Resource Coordination**: Prevent conflicts during concurrent operations
- **Knowledge Sharing**: Cross-pollinate domain expertise

## 项目管理与执行原则（对齐 CLAUDE.md）

以下原则为本仓库 Agent 的强制工作规范，来源：`CLAUDE.md` 项目记忆文档。所有 Agent 在执行分析、设计、实现、测试与文档工作时必须遵循。

### 开发与沟通原则
- 诚实原则：状态、性能、风险一律基于可验证事实，不夸大，不隐瞒。
- 悲观谨慎：按最坏情况评估方案与风险，预留缓冲并分阶段验证。
- 健壮优先：拒绝权宜之计，优先根因修复与可维护性，配套测试与文档。
- 中文沟通：与用户交流、提交物描述、变更说明优先使用中文；专业、准确、清晰。

### 临时方案管控
- 必须显式标注：仅在确有必要时使用临时方案，并使用 `// TODO-TEMPORARY:` 注明原因、改进计划与最后期限（不超过一个迭代）。
- 建立清单：对临时方案进行台账管理，定期清理与回收。
- 严禁：无标注的临时实现、削减错误处理、绕过一致性校验。

### 新功能审批
- 强制审批：任何新增 API 端点、页面/组件、服务、数据库表，先分析设计，实施前需用户明确同意。
- 修复除外：修复现有功能缺陷无需审批，但需如实记录变更与验证方式。

### PostgreSQL 原生 CQRS 规则（2025-08-22 修订）
- 协议分工：查询统一走 GraphQL；命令统一走 REST API。
- 单一数据源：读写均基于 PostgreSQL；禁止引入 Neo4j/Kafka CDC 同步等额外数据源。
- 性能与一致性：利用索引/窗口函数优化；保持强一致与零同步延迟的假设前提。

### API 一致性规范（2025-08-23 修订）
- JSON 字段一律使用 camelCase（如 `parentCode`, `effectiveDate`, `recordId`）。
- 路径参数统一 `{code}`，禁止 `{id}` 用于组织单元路径参数。
- 查询参数使用 camelCase；数据库层可维持 snake_case，但 API 层必须转换为 camelCase。
- 代码审查必查：响应中不得出现 snake_case；跨层字段命名需对齐（前端 TS ↔ 后端 Go）。

### API 优先授权管理（2025-08-31 重大新增）
- 单一事实来源：权限定义优先以 `docs/api/openapi.yaml` 为准；先契约、后实现。
- 实施顺序：先在契约中定义端点与 scopes → 后端路由与中间件校验 → 前端 UI 权限控制。
- 禁止：脱离契约的私有权限、硬编码未声明权限、前后端权限不一致。

### 表述与数据
- 禁止绝对化/夸张表述（如“革命性/完全/100%/0错误/所有/一键解决”等）。
- 任何性能与效果陈述需有可复现实验数据支撑，并注明测试条件与范围。

### 早期/阶段性聚焦（如 CLAUDE.md 规定）
- 按项目阶段调整优先级：核心功能与一致性优先；避免早期过度工程化。
- 若阶段说明与 README 存在差异，以 CLAUDE.md 标注的最新阶段要求为准。

### 提交与验证清单（面向所有 PR/变更）
- 设计对齐：是否符合 CQRS 分工（GraphQL 查询 / REST 命令）。
- 命名一致：API 输入/输出字段均为 camelCase；路径参数为 `{code}`。
- 权限契约：相关端点已在 OpenAPI 中定义权限并对齐实现与前端控制。
- 测试具备：为核心路径提供可运行的最小验证（单测/集成/契约测试）。
- 文档更新：在相应文档中反映接口与行为变化；临时方案标注到位。
- 运行可证：提供可复现实验步骤与数据，不夸大效果。

> 注：如发现历史文档、脚本或目录与上述规则不一致（例如遗留 Neo4j/Kafka 引用、snake_case 示例等），实现时一律以本节规范为准，同时在变更说明中记录差异与迁移路径。

## Agent Specialization Matrix

| Agent Type | Primary Domain | Secondary Skills | Tools & Frameworks |
|------------|----------------|------------------|-------------------|
| Backend | Go services, databases | API design, event sourcing | Go, PostgreSQL, Neo4j, Kafka |
| Frontend | React/Next.js, UI/UX | State management, performance | TypeScript, Ant Design, Apollo |
| AI | Python services, ML | NLP, gRPC services | Python, gRPC, FastAPI |
| DevOps | Infrastructure, deployment | Monitoring, automation | Docker, CI/CD, monitoring tools |
| QA | Testing, validation | Performance testing, automation | Playwright, Jest, testing frameworks |
| Architecture | System design, documentation | Integration patterns, ADRs | OpenAPI, diagrams, documentation |

## Agent Delegation Patterns

### Task Classification
```yaml
simple_tasks:
  description: "Single-domain, straightforward operations"
  delegation: "Single specialized agent"
  examples: ["Fix bug in employee API", "Update UI component styling"]

complex_tasks:
  description: "Multi-domain operations requiring coordination"
  delegation: "Primary agent + supporting agents"
  examples: ["Implement new employee onboarding workflow", "Add real-time notifications"]

cross_cutting_tasks:
  description: "Tasks affecting multiple services/domains"
  delegation: "Architecture agent coordinates multiple specialists"
  examples: ["Implement business ID standardization", "Add new authentication method"]
```

### Agent Selection Logic
1. **Domain Analysis**: Identify primary technical domain
2. **Complexity Assessment**: Determine if single or multi-agent approach needed
3. **Dependency Mapping**: Identify which services/components are affected
4. **Expertise Matching**: Select agents with relevant specializations
5. **Coordination Planning**: Define primary agent and supporting roles

## Usage Examples

### Employee Management Feature
```bash
# Architecture agent analyzes requirements
/agent architecture-agent analyze-requirements "Add employee performance tracking"

# Backend agent implements API layer
/agent backend-agent implement-api --feature=performance-tracking --with-cqrs

# Frontend agent creates UI components
/agent frontend-agent create-performance-dashboard --with-charts --responsive

# QA agent develops test suite
/agent qa-agent create-test-suite --feature=performance-tracking --include-e2e

# DevOps agent handles deployment
/agent devops-agent deploy-feature --feature=performance-tracking --environment=staging
```

### System-Wide Improvements
```bash
# Architecture agent leads system optimization
/agent architecture-agent optimize-system --focus=performance --coordinate

# Multiple agents work in parallel:
# - Backend agent optimizes database queries
# - Frontend agent implements lazy loading
# - DevOps agent optimizes container resources
# - QA agent validates performance improvements
```

## Agent State Management

### Context Sharing
- **Project State**: Current system status and recent changes
- **Task History**: Record of completed and ongoing tasks
- **Resource Status**: Database, service, and infrastructure state
- **Known Issues**: Current bugs, technical debt, and limitations

### Progress Tracking
- **Task Assignment**: Which agent is responsible for what
- **Dependency Tracking**: Task interdependencies and blockers
- **Completion Status**: Progress indicators and milestone tracking
- **Quality Gates**: Validation checkpoints and approval processes

## Best Practices

### Agent Selection
1. **Single Responsibility**: Each agent focuses on their core domain
2. **Clear Ownership**: Primary agent takes responsibility for task completion
3. **Collaborative Support**: Supporting agents provide expertise without taking over
4. **Context Preservation**: Maintain project understanding across agent switches

### Task Execution
1. **Plan Before Action**: Architecture analysis before implementation
2. **Incremental Progress**: Break large tasks into manageable chunks
3. **Continuous Validation**: Test and validate at each step
4. **Documentation**: Update relevant documentation as work progresses

### Quality Assurance
1. **Multi-Agent Review**: Cross-domain validation of changes
2. **Integration Testing**: Verify changes work across all affected services
3. **Performance Impact**: Assess performance implications of changes
4. **Security Review**: Security validation for sensitive changes

This agent system provides specialized expertise while maintaining coordination and project coherence across the complex Cube Castle HR system.
