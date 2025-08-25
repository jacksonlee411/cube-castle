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
- ADR management (`DOCS2/architecture-decisions/`)
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