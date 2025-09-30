# 代码异味基线报告

**生成日期**: 2025-09-29  
**报告版本**: v1.0  
**扫描范围**: Go后端 (`cmd/`) + 前端TypeScript (`frontend/src`)  
**目的**: 为16号计划（代码异味分析与改进）提供基线数据

---

## 1. Go后端文件统计

### 总体情况
- **文件总数**: 54个
- **代码总行数**: 16,888行
- **平均行数**: 312.7行/文件

### 文件大小分级（基于Go工程实践标准）

#### 🔴 红灯区域（>800行，强制重构）
| 文件路径 | 行数 | 超标程度 |
|---------|------|---------|
| `cmd/organization-query-service/main.go` | 2,264 | 严重超标 (283%) |
| `cmd/organization-command-service/internal/handlers/organization.go` | 1,399 | 严重超标 (175%) |
| `cmd/organization-command-service/internal/repository/organization.go` | 817 | 超标 (102%) |

**红灯文件数**: 3个  
**红灯代码占比**: 27.5% (4,480行/16,888行)

#### 🟠 橙灯区域（600-800行，需架构师评估）
| 文件路径 | 行数 |
|---------|------|
| `cmd/organization-command-service/internal/services/temporal.go` | 773 |
| `cmd/organization-command-service/internal/repository/temporal_timeline.go` | 685 |
| `cmd/organization-command-service/internal/validators/business.go` | 596 |
| `cmd/organization-command-service/internal/audit/logger.go` | 595 |
| `cmd/organization-command-service/internal/authbff/handler.go` | 589 |

**橙灯文件数**: 5个  
**橙灯代码占比**: 22.1% (3,238行/16,888行)

#### 🟡 黄灯区域（400-600行，关注结构优化）
| 文件路径 | 行数 |
|---------|------|
| `cmd/organization-command-service/internal/services/organization_temporal_service.go` | 507 |
| `cmd/organization-command-service/internal/services/temporal_monitor.go` | 506 |
| `cmd/organization-command-service/internal/repository/hierarchy.go` | 469 |
| `cmd/organization-command-service/internal/handlers/devtools.go` | 421 |

**黄灯文件数**: 4个  
**黄灯代码占比**: 13.9% (2,403行/16,888行)

#### 🟢 绿灯区域（<400行，符合标准）
**绿灯文件数**: 42个  
**绿灯代码占比**: 36.5% (6,767行/16,888行)

---

## 2. 前端TypeScript文件统计

### 总体情况
- **文件总数**: 112个
- **代码总行数**: 18,254行
- **平均行数**: 163.0行/文件

### 文件大小分级

#### 🔴 红灯区域（>800行，需要重构）
| 文件路径 | 行数 | 超标程度 |
|---------|------|---------|
| `frontend/src/features/temporal/components/TemporalMasterDetailView.tsx` | 1,157 | 严重超标 (145%) |
| `frontend/src/features/temporal/components/InlineNewVersionForm.tsx` | 1,067 | 严重超标 (133%) |

**红灯文件数**: 2个  
**红灯代码占比**: 12.2% (2,224行/18,254行)

#### 🟠 橙灯区域（400-800行，评估拆分价值）
| 文件路径 | 行数 |
|---------|------|
| `frontend/src/features/organizations/components/OrganizationTree.tsx` | 586 |
| `frontend/src/shared/hooks/useEnterpriseOrganizations.ts` | 491 |
| `frontend/src/shared/api/unified-client.ts` | 486 |
| `frontend/src/features/temporal/components/TemporalEditForm.tsx` | 483 |
| `frontend/src/features/temporal/components/OrganizationDetailForm.tsx` | 480 |
| `frontend/src/shared/hooks/useOrganizationMutations.ts` | 454 |
| `frontend/src/features/temporal/components/TimelineComponent.tsx` | 439 |
| `frontend/src/shared/api/auth.ts` | 430 |
| `frontend/src/features/organizations/OrganizationFilters.tsx` | 402 |

**橙灯文件数**: 9个  
**橙灯代码占比**: 26.2% (4,751行/18,254行)

#### 🟢 绿灯区域（<400行，符合标准）
**绿灯文件数**: 101个  
**绿灯代码占比**: 61.6% (11,279行/18,254行)

#### 弱类型使用统计
- **any/unknown 使用合计**: 171处（覆盖36个文件）
- **数据来源**: `rg "\b(any|unknown)\b" frontend/src --stats`（执行时间：2025-09-30）

---

## 3. 问题汇总

### Go后端核心问题
1. **超大文件集中**：3个红灯文件占总代码量27.5%
2. **main.go畸形**：查询服务main.go达2,264行，严重违反单一职责原则
3. **handlers/repository混杂**：handler和repository层都存在超大文件

### 前端核心问题
1. **组件巨石化**：时态管理组件普遍超大（TemporalMasterDetailView、InlineNewVersionForm）
2. **逻辑未分离**：UI组件与业务逻辑耦合严重
3. **Hook过重**：useEnterpriseOrganizations、useOrganizationMutations超400行

---

## 4. 改进目标（16号计划Phase 1-3）

### 量化目标
- **Go文件平均行数**: 312.7 → ≤350行
- **前端文件平均行数**: 163.0 → ≤150行（拆分后按周监测）
- **红灯文件清零**: Go 3个→0个, TS 2个→0个
- **橙灯文件控制**: Go 5个→≤3个, TS 9个→≤5个

### Phase 1 重点（红灯清零）
- `main.go` (2,264行) → 6-8个文件 (<400行/文件)
- `organization.go` handler (1,399行) → 4个文件 (<400行/文件)
- `organization.go` repository (817行) → 3个文件 (<300行/文件)
- `TemporalMasterDetailView.tsx` (1,157行) → 3个文件
- `InlineNewVersionForm.tsx` (1,067行) → 3个文件

---

## 5. 验证方式

### 基线复现命令
```bash
# Go后端统计
find cmd -name '*.go' -type f -print0 | xargs -0 wc -l | sort -rn

# 前端统计
find frontend/src -type f \( -name '*.ts' -o -name '*.tsx' \) -print0 | xargs -0 wc -l | sort -rn

# 平均值计算
echo "Go平均: $(find cmd -name '*.go' -type f -print0 | xargs -0 wc -l | tail -1 | awk '{total=$1; count=NF-2; print total/(count==0?54:count)}')"
```

### 下次检查时间
- **Phase 1 中期检查**: 2025-10-07（预期红灯文件≤1个）
- **Phase 1 完成检查**: 2025-10-14（预期红灯文件=0个）
- **Phase 2 完成检查**: 2025-10-21（类型治理验收）
- **Phase 3 完成检查**: 2025-10-28（监控系统上线）

---

**报告生成者**: 架构组  
**审阅者**: [待补充]  
**批准日期**: [待补充]

---

*本报告为16号计划的唯一事实来源基线，后续所有进展均以此为对比依据。*
