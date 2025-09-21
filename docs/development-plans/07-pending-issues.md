# 07 — 时态组织架构用户体验优化

最后更新：2025-09-21
责任团队：前端组（主责）+ 后端组
当前状态：需求分析完成，待实施

---

## 1. 问题定位与根因分析

### 1.1 真正的问题

通过深入测试和验证，发现组织层级修改失败的根本原因是**时态约束违反**：

**核心问题**：尝试在2025-01-01将组织单元1000057的上级设为1000006，但1000006的生效日期是2025-04-01，在指定的修改生效日期时该组织尚不存在，违反了时态数据的逻辑一致性。

**问题验证过程**：
1. 初步怀疑路径数据不一致 → 修复后仍然失败
2. 检查时态数据结构 → 发现1000056存在多个历史版本（正常）
3. 深入分析验证逻辑 → 发现时态约束验证正常工作

**结论**：这是一个**设计正确的业务规则验证**，防止了时态数据的逻辑错误。系统行为符合预期，但用户体验需要优化。

### 1.2 相关数据结构

```sql
-- 组织1000006的时态信息
code: 1000006, name: E2E一体化测试部
effective_date: 2025-04-01, parent_code: 1000000

-- 组织1000057的当前记录
code: 1000057, name: 人力资源部
effective_date: 2025-01-01, parent_code: 1000056

-- 尝试的修改：在2025-01-01生效日期下，将1000057的上级改为1000006
-- 失败原因：1000006在2025-01-01时不存在（2025-04-01才生效）
```

---

## 2. 用户体验优化需求

### 2.1 优化点1：时态感知的上级组织筛选

**问题**：上级组织下拉框显示所有组织，不考虑指定生效日期的可用性。

**优化目标**：
- 根据当前记录的生效日期，只显示在该日期有效且状态为ACTIVE的组织
- 避免用户选择在时态上不合法的组织

**实施策略**：坚持 CQRS 原则，不新增 REST 端点，通过扩展现有 `organizations` GraphQL 查询实现所有过滤能力，最大化契约复用。

**技术实现方案**：

1. **GraphQL契约扩展**：
   - 在 `docs/api/schema.graphql` 的 `OrganizationFilter` 中新增可选字段 `excludeCodes`、`excludeDescendantsOf`（数组/单值均可），以复用既有 `organizations(filter: ...)` 查询。
   - 更新 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 相关条目，保持契约单一来源。

2. **查询服务复用实现**：
   ```go
   // cmd/organization-query-service/main.go
   type OrganizationFilter struct {
       // ...现有字段
       ExcludeCodes          *[]string `json:"excludeCodes"`
       ExcludeDescendantsOf  *string   `json:"excludeDescendantsOf"`
   }

   func (r *PostgreSQLRepository) GetOrganizations(ctx context.Context, tenantID uuid.UUID, filter *OrganizationFilter, pagination *PaginationInput) (*OrganizationConnection, error) {
       // 在生成 SQL 时追加 WHERE NOT (code = ANY(excludeCodes))
       // 且基于 hierarchy_repo.GetAncestorChain/现有码路径缓存排除子孙节点
   }
   ```

3. **客户端查询示例**：
   ```graphql
   query AvailableParents($asOfDate: String!, $excludeCode: String!) {
     organizations(
       filter: {
         asOfDate: $asOfDate
         status: ACTIVE
         excludeCodes: [$excludeCode]
         excludeDescendantsOf: $excludeCode
       }
       pagination: { page: 1, pageSize: 50 }
     ) {
       nodes { code name parentCode effectiveDate }
     }
   }
   ```

2. **前端组件优化**：
   ```typescript
   // 在 ParentOrganizationCombobox 组件中复用 GraphQL 查询
   const effectiveDate = form.watch('effectiveDate');

   const { data: availableParents, isFetching } = useQuery({
     queryKey: ['availableParents', organizationCode, effectiveDate],
     enabled: Boolean(effectiveDate),
     queryFn: () => organizationGraphQL.fetchOrganizations({
       filter: {
         asOfDate: effectiveDate,
         status: 'ACTIVE',
         excludeCodes: [organizationCode],
         excludeDescendantsOf: organizationCode,
       },
       pagination: { page: 1, pageSize: 50 },
     }),
   });

   useEffect(() => {
     setAvailableOrganizations(availableParents?.nodes ?? []);
   }, [availableParents]);
  ```

3. **用户界面提示**：
   ```tsx
   <ComboboxInput
       placeholder="搜索并选择上级组织..."
       helperText={`显示在 ${effectiveDate} 有效且状态为 ACTIVE 的组织`}
       emptyText="在指定日期没有可用的上级组织"
       loading={isFetching}
   />
   ```

### 2.2 优化点2：友好的错误提示

**问题**：当前错误提示过于技术化，用户无法理解失败原因。

**当前提示**：
- "操作失败修改失败，请检查网络连接"
- "业务规则校验失败"

**优化目标**：
- 提供具体、可操作的错误信息
- 指导用户如何正确操作

**技术实现方案**：

1. **后端错误详细化**：
   ```go
   func (v *BusinessRuleValidator) validateTemporalParentChange(ctx context.Context, tenantID uuid.UUID, parentCode string, effectiveDate time.Time, result *ValidationResult) error {
       parentAtDate, err := v.hierarchyRepo.GetOrganizationAtDate(ctx, parentCode, effectiveDate, tenantID)
       if err != nil {
           return fmt.Errorf("查询父组织时态失败: %w", err)
       }

       if parentAtDate == nil {
           latestParent, _ := v.hierarchyRepo.GetOrganization(ctx, parentCode, tenantID)

           message := fmt.Sprintf("上级组织 %s 在指定生效日期 %s 不存在或未激活。",
               parentCode, effectiveDate.Format("2006-01-02"))
           context := map[string]interface{}{}

           if latestParent != nil && latestParent.EffectiveDate != nil {
               nextDate := latestParent.EffectiveDate.String()
               message += fmt.Sprintf(" 可选择在 %s 之后生效，或更换上级组织。", nextDate)
               context["suggestedDate"] = nextDate
               context["parentName"] = latestParent.Name
           }

           result.Errors = append(result.Errors, ValidationError{
               Code:     "TEMPORAL_PARENT_UNAVAILABLE",
               Message:  message,
               Field:    "parentCode",
               Value:    parentCode,
               Severity: "HIGH",
               Context:  context,
           })

           return nil
       }

       return nil
   }
   ```

2. **前端错误处理优化**：
   ```typescript
   // 在 useOrganizationMutation 中
   const getErrorMessage = (error: ApiError) => {
       if (error.code === 'TEMPORAL_PARENT_UNAVAILABLE') {
           const actions = [];
           if (error.context?.suggestedDate) {
               actions.push({
                   label: `调整至 ${error.context.suggestedDate}`,
                   onClick: () => form.setValue('effectiveDate', error.context.suggestedDate),
               });
           }
           actions.push({
               label: '重新选择上级组织',
               onClick: () => form.setValue('parentCode', null),
           });

           return {
               title: '上级组织不可用',
               message: error.message,
               type: 'warning',
               actions,
           };
       }
       // 处理其他错误类型...
   };
   ```

3. **用户界面优化**：
   ```tsx
   <ErrorAlert
       title={errorInfo.title}
       message={errorInfo.message}
       type={errorInfo.type}
       actions={errorInfo.actions}
       dismissible
   />
   ```

---

## 3. 实施计划

### 3.1 Phase 1: 后端API优化（预估2天）

**负责团队**：后端组
**优先级**：高

**任务清单**：
1. 在 `docs/api/schema.graphql` 中扩展 `OrganizationFilter`，并同步实现清单/契约文档
2. 调整 `organization-query-service` 的 `OrganizationFilter` 及 SQL 生成逻辑，复用现有 `GetOrganizations`
3. 优化 `BusinessRuleValidator` 时态父级校验，复用现有仓储并补充安全提示
4. 编写/更新单测与集成测试覆盖过滤与错误信息

**验收标准**：
- `organizations(filter: { asOfDate: "2025-01-01", status: ACTIVE, excludeCodes: ["1000057"], excludeDescendantsOf: "1000057" })` 仅返回在指定日期有效且非自身/非子孙的组织
- 时态约束验证失败时返回具体的错误信息和建议操作，GraphQL/REST 返回结构保持契约一致

### 3.2 Phase 2: 前端体验优化（预估3天）

**负责团队**：前端组
**优先级**：高

**任务清单**：
1. 重构 ParentOrganizationCombobox 组件，复用 GraphQL organizations 查询实现时态感知筛选
2. 实现智能错误处理和用户引导
3. 添加加载状态和空状态处理
4. 优化表单验证和用户反馈

**验收标准**：
- 上级组织下拉框只显示在指定生效日期可用的组织
- 错误提示具体明确，说明失败原因和操作建议
- 用户界面响应流畅，加载状态清晰

### 3.3 Phase 3: 集成测试与优化（预估1天）

**负责团队**：测试组 + 前后端组
**优先级**：中

**任务清单**：
1. 端到端测试时态约束场景
2. 用户体验走查和反馈收集
3. 性能优化和缓存策略
4. 文档更新和知识分享

---

## 4. 技术细节

### 4.1 时态查询优化

```sql
-- 复用 organizations 查询的过滤模板，新增排除逻辑
WITH latest_versions AS (
    SELECT DISTINCT ON (code)
        code,
        parent_code,
        name,
        level,
        COALESCE(code_path, code) AS code_path,
        effective_date,
        end_date,
        status
    FROM organization_units
    WHERE tenant_id = $1
      AND effective_date <= $2
      AND (end_date IS NULL OR end_date > $2)
    ORDER BY code, effective_date DESC
)
SELECT *
FROM latest_versions
WHERE status = 'ACTIVE'
  AND ( $3 IS NULL OR code <> $3 )
  AND ( $4 IS NULL OR code_path NOT LIKE $4 || '/%')
ORDER BY name;
```

> 注：参数 `$4` 为待排除组织的 `code_path` 前缀，直接复用 `HierarchyRepository.GetCodePath` 的结果，可避免重复实现循环校验逻辑。

### 4.2 错误分类与处理

```typescript
export enum TemporalValidationError {
    PARENT_NOT_AVAILABLE = 'TEMPORAL_PARENT_UNAVAILABLE',
    CIRCULAR_REFERENCE = 'CIRCULAR_REFERENCE',
    DEPTH_EXCEEDED = 'DEPTH_EXCEEDED',
    INVALID_EFFECTIVE_DATE = 'INVALID_EFFECTIVE_DATE'
}

export interface ValidationErrorContext {
    suggestedDate?: string;
    parentName?: string;
    maxDepth?: number;
    conflictingCodes?: string[];
}
```

---

## 5. 预期收益

### 5.1 用户体验提升

- **减少操作错误**：用户无法选择在时态上不合法的组织
- **明确错误指导**：具体的错误信息和操作建议
- **操作效率提升**：避免反复试错，直接定位到可行方案

### 5.2 系统稳定性

- **数据一致性保障**：从UI层面预防时态约束违反
- **错误处理标准化**：统一的错误格式和处理流程
- **调试效率提升**：清晰的错误日志和用户反馈

### 5.3 维护成本降低

- **支持工单减少**：用户能够自主解决常见的操作问题
- **培训成本降低**：直观的界面减少用户学习成本
- **代码质量提升**：统一的错误处理和验证逻辑

---

## 6. 风险评估与缓解

### 6.1 技术风险

**风险**：时态查询性能影响
**缓解措施**：
- 实施前先通过现有 APM 指标基线评估查询时间；若落地后 P95 > 200ms，再按需加索引 (effective_date, end_date, status)
- 仅在真实瓶颈出现时启用缓存（记录触发条件及回滚方案）
- 将关键查询纳入 `reports/performance/` 监控，版本上线后连续两周跟踪指标

**风险**：错误信息国际化复杂性
**缓解措施**：
- 使用结构化的错误代码和模板
- 分离错误逻辑和展示逻辑
- 预留多语言支持的架构

### 6.2 业务风险

**风险**：现有工作流程改变
**缓解措施**：
- 保持向后兼容，渐进式部署
- 提供用户培训和文档更新
- 收集用户反馈并快速响应

---

## 7. 验收与交付

### 7.1 验收场景

1. **时态约束正确处理**：
   - 在2025-01-01修改1000057时，上级组织列表不包含1000006
   - 选择1000000等在该日期有效的组织可以成功保存

2. **错误提示友好性**：
   - 违反时态约束时显示具体的错误原因和操作建议
   - 错误信息清晰易懂，指导用户如何修正

3. **性能表现**：
   - 上级组织查询响应时间 < 500ms
   - 界面操作流畅，无明显卡顿

### 7.2 交付物

1. **代码实现**：前后端优化代码，包含完整的单元测试
2. **API文档**：更新的接口文档和错误码说明
3. **用户指南**：操作手册和常见问题解答
4. **技术文档**：架构说明和维护指南

---

**下一步行动**：前端组和后端组协调排期，启动Phase 1的API优化开发工作。
