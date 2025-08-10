/**
 * 组织版本管理和对比API服务
 * 专门处理组织版本的高级管理、时态操作和版本对比分析
 */
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ===== 版本管理数据模型 =====

type VersionedOrganization struct {
	ID               string                 `json:"id" db:"id"`
	OrganizationCode string                 `json:"organization_code" db:"organization_code"`
	Version          int                    `json:"version" db:"version"`
	Name             string                 `json:"name" db:"name"`
	UnitType         string                 `json:"unit_type" db:"unit_type"`
	Status           string                 `json:"status" db:"status"`
	Level            int                    `json:"level" db:"level"`
	Path             string                 `json:"path" db:"path"`
	SortOrder        int                    `json:"sort_order" db:"sort_order"`
	Description      string                 `json:"description,omitempty" db:"description"`
	ParentCode       *string                `json:"parent_code,omitempty" db:"parent_code"`
	EffectiveFrom    time.Time              `json:"effective_from" db:"effective_from"`
	EffectiveTo      *time.Time             `json:"effective_to,omitempty" db:"effective_to"`
	ChangeReason     string                 `json:"change_reason" db:"change_reason"`
	SnapshotData     map[string]interface{} `json:"snapshot_data" db:"snapshot_data"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	TenantID         string                 `json:"tenant_id" db:"tenant_id"`
	// 计算字段
	IsActive         bool                   `json:"is_active"`
	TemporalStatus   string                 `json:"temporal_status"`
	DaysSinceActive  *int                   `json:"days_since_active,omitempty"`
	DaysUntilExpiry  *int                   `json:"days_until_expiry,omitempty"`
}

// 版本创建请求
type CreateVersionRequest struct {
	BasedOnVersion *int                   `json:"based_on_version,omitempty"` // 基于哪个版本
	Changes        map[string]interface{} `json:"changes"`                    // 要修改的字段
	EffectiveFrom  time.Time              `json:"effective_from"`             // 生效时间
	EffectiveTo    *time.Time             `json:"effective_to,omitempty"`     // 失效时间
	ChangeReason   string                 `json:"change_reason"`              // 变更原因
	PreviewOnly    bool                   `json:"preview_only,omitempty"`     // 仅预览，不实际创建
}

// 批量版本操作请求
type BatchVersionRequest struct {
	Versions      []int     `json:"versions"`          // 要操作的版本列表
	Operation     string    `json:"operation"`         // activate, deactivate, extend, expire
	EffectiveDate time.Time `json:"effective_date"`    // 操作生效日期
	Reason        string    `json:"reason"`            // 操作原因
}

// 时态校正请求
type TemporalCorrectionRequest struct {
	TargetVersion   int                    `json:"target_version"`         // 要校正的版本
	CorrectionType  string                 `json:"correction_type"`        // data_fix, date_adjustment, rollback
	Corrections     map[string]interface{} `json:"corrections"`            // 校正内容
	NewEffectiveFrom *time.Time            `json:"new_effective_from,omitempty"` // 新的生效时间
	NewEffectiveTo   *time.Time            `json:"new_effective_to,omitempty"`   // 新的失效时间
	CorrectionReason string                `json:"correction_reason"`      // 校正原因
}

// 高级版本对比选项
type AdvancedComparisonOptions struct {
	CompareFields      []string `json:"compare_fields,omitempty"`       // 只比较特定字段
	IgnoreFields       []string `json:"ignore_fields,omitempty"`        // 忽略特定字段
	SemanticComparison bool     `json:"semantic_comparison,omitempty"`  // 语义比较
	ShowUnchanged      bool     `json:"show_unchanged,omitempty"`       // 显示未变更字段
	ComparisonFormat   string   `json:"comparison_format,omitempty"`    // diff, side_by_side, unified
}

// 详细版本对比结果
type DetailedVersionComparison struct {
	FromVersion        int                      `json:"from_version"`
	ToVersion          int                      `json:"to_version"`
	ComparedAt         time.Time                `json:"compared_at"`
	ComparisonOptions  AdvancedComparisonOptions `json:"comparison_options"`
	FieldChanges       []DetailedFieldChange    `json:"field_changes"`
	Summary            ComparisonSummary        `json:"summary"`
	Impact             ComparisonImpact         `json:"impact"`
	Recommendations    []string                 `json:"recommendations,omitempty"`
}

type DetailedFieldChange struct {
	Field          string      `json:"field"`
	FieldType      string      `json:"field_type"`      // string, number, date, array, object
	ChangeType     string      `json:"change_type"`     // added, removed, modified, unchanged
	OldValue       interface{} `json:"old_value"`
	NewValue       interface{} `json:"new_value"`
	Significance   string      `json:"significance"`    // minor, major, critical
	ChangePattern  string      `json:"change_pattern"`  // direct, nested, array_item
	HumanReadable  string      `json:"human_readable"`  // 人类可读的变更描述
}

type ComparisonSummary struct {
	TotalFields        int `json:"total_fields"`
	ChangedFields      int `json:"changed_fields"`
	UnchangedFields    int `json:"unchanged_fields"`
	AddedFields        int `json:"added_fields"`
	RemovedFields      int `json:"removed_fields"`
	CriticalChanges    int `json:"critical_changes"`
	MajorChanges       int `json:"major_changes"`
	MinorChanges       int `json:"minor_changes"`
}

type ComparisonImpact struct {
	StructuralImpact     string   `json:"structural_impact"`      // none, low, medium, high
	BusinessImpact       string   `json:"business_impact"`        // none, low, medium, high
	DataIntegrityRisk    string   `json:"data_integrity_risk"`    // none, low, medium, high
	AffectedSystems      []string `json:"affected_systems,omitempty"`
	RequiredActions      []string `json:"required_actions,omitempty"`
}

// 版本分析报告
type VersionAnalysisReport struct {
	OrganizationCode    string                  `json:"organization_code"`
	AnalysisDate        time.Time               `json:"analysis_date"`
	TotalVersions       int                     `json:"total_versions"`
	ActiveVersions      int                     `json:"active_versions"`
	PlannedVersions     int                     `json:"planned_versions"`
	ExpiredVersions     int                     `json:"expired_versions"`
	VersionTimeline     []VersionTimelineEntry `json:"version_timeline"`
	ChangeFrequency     ChangeFrequencyAnalysis `json:"change_frequency"`
	QualityMetrics      VersionQualityMetrics   `json:"quality_metrics"`
	Recommendations     []AnalysisRecommendation `json:"recommendations"`
}

type VersionTimelineEntry struct {
	Version         int        `json:"version"`
	EffectiveFrom   time.Time  `json:"effective_from"`
	EffectiveTo     *time.Time `json:"effective_to,omitempty"`
	Duration        int        `json:"duration_days"`
	ChangeReason    string     `json:"change_reason"`
	ChangeCategory  string     `json:"change_category"`
}

type ChangeFrequencyAnalysis struct {
	ChangesPerMonth     float64            `json:"changes_per_month"`
	ChangesPerQuarter   float64            `json:"changes_per_quarter"`
	ChangesByCategory   map[string]int     `json:"changes_by_category"`
	ChangesByReason     map[string]int     `json:"changes_by_reason"`
	SeasonalPatterns    []SeasonalPattern  `json:"seasonal_patterns,omitempty"`
}

type SeasonalPattern struct {
	Period      string  `json:"period"`        // Q1, Q2, Q3, Q4, Jan, Feb, etc.
	ChangeCount int     `json:"change_count"`
	AverageCount float64 `json:"average_count"`
	Variance    float64 `json:"variance"`
}

type VersionQualityMetrics struct {
	AverageVersionDuration  float64 `json:"average_version_duration_days"`
	ShortLivedVersions      int     `json:"short_lived_versions"`      // <30天
	LongLivedVersions       int     `json:"long_lived_versions"`       // >365天
	RollbackCount           int     `json:"rollback_count"`
	CorrectionCount         int     `json:"correction_count"`
	DataConsistencyScore    float64 `json:"data_consistency_score"`    // 0-100
	ChangeReasonQuality     float64 `json:"change_reason_quality"`     // 0-100
}

type AnalysisRecommendation struct {
	Type         string `json:"type"`         // optimization, data_quality, process_improvement
	Priority     string `json:"priority"`     // low, medium, high, critical
	Title        string `json:"title"`
	Description  string `json:"description"`
	ActionItems  []string `json:"action_items,omitempty"`
	EstimatedImpact string `json:"estimated_impact,omitempty"`
}

// ===== 版本管理仓储层 =====

type VersionRepository struct {
	db *sql.DB
}

func NewVersionRepository(db *sql.DB) *VersionRepository {
	return &VersionRepository{db: db}
}

// 获取所有版本
func (r *VersionRepository) GetAllVersions(ctx context.Context, tenantID uuid.UUID, orgCode string) ([]VersionedOrganization, error) {
	query := `
		SELECT id, organization_code, version, effective_from, effective_to,
		       snapshot_data, change_reason, created_at, tenant_id
		FROM organization_unit_versions
		WHERE tenant_id = $1 AND organization_code = $2
		ORDER BY version DESC
	`

	rows, err := r.db.QueryContext(ctx, query, tenantID.String(), orgCode)
	if err != nil {
		return nil, fmt.Errorf("查询版本列表失败: %w", err)
	}
	defer rows.Close()

	var versions []VersionedOrganization
	now := time.Now()

	for rows.Next() {
		var version VersionedOrganization
		var snapshotBytes []byte

		err := rows.Scan(
			&version.ID, &version.OrganizationCode, &version.Version,
			&version.EffectiveFrom, &version.EffectiveTo,
			&snapshotBytes, &version.ChangeReason,
			&version.CreatedAt, &version.TenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描版本数据失败: %w", err)
		}

		// 解析快照数据
		if len(snapshotBytes) > 0 {
			json.Unmarshal(snapshotBytes, &version.SnapshotData)
			
			// 从快照数据中提取字段
			if name, ok := version.SnapshotData["name"].(string); ok {
				version.Name = name
			}
			if unitType, ok := version.SnapshotData["unit_type"].(string); ok {
				version.UnitType = unitType
			}
			if status, ok := version.SnapshotData["status"].(string); ok {
				version.Status = status
			}
			if level, ok := version.SnapshotData["level"].(float64); ok {
				version.Level = int(level)
			}
		}

		// 计算时态状态
		if version.EffectiveFrom.After(now) {
			version.TemporalStatus = "planned"
			days := int(version.EffectiveFrom.Sub(now).Hours() / 24)
			version.DaysSinceActive = &days
		} else if version.EffectiveTo != nil && version.EffectiveTo.Before(now) {
			version.TemporalStatus = "expired"
			days := int(now.Sub(*version.EffectiveTo).Hours() / 24)
			version.DaysUntilExpiry = &days
		} else {
			version.TemporalStatus = "active"
			version.IsActive = true
			days := int(now.Sub(version.EffectiveFrom).Hours() / 24)
			version.DaysSinceActive = &days
			if version.EffectiveTo != nil {
				expDays := int(version.EffectiveTo.Sub(now).Hours() / 24)
				version.DaysUntilExpiry = &expDays
			}
		}

		versions = append(versions, version)
	}

	return versions, nil
}

// 创建新版本
func (r *VersionRepository) CreateVersion(ctx context.Context, tenantID uuid.UUID, orgCode string, req *CreateVersionRequest) (*VersionedOrganization, error) {
	if req.PreviewOnly {
		// 仅返回预览，不实际创建
		return r.previewVersionCreation(ctx, tenantID, orgCode, req)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 获取基础版本数据
	baseVersion := 1
	if req.BasedOnVersion != nil {
		baseVersion = *req.BasedOnVersion
	} else {
		// 获取最新版本号
		err = tx.QueryRowContext(ctx,
			"SELECT COALESCE(MAX(version), 0) FROM organization_unit_versions WHERE tenant_id = $1 AND organization_code = $2",
			tenantID.String(), orgCode).Scan(&baseVersion)
		if err != nil {
			return nil, fmt.Errorf("获取基础版本失败: %w", err)
		}
	}

	// 获取基础版本的快照数据
	var baseSnapshot map[string]interface{}
	var snapshotBytes []byte
	err = tx.QueryRowContext(ctx,
		"SELECT snapshot_data FROM organization_unit_versions WHERE tenant_id = $1 AND organization_code = $2 AND version = $3",
		tenantID.String(), orgCode, baseVersion).Scan(&snapshotBytes)
	
	if err != nil {
		return nil, fmt.Errorf("获取基础版本快照失败: %w", err)
	}

	json.Unmarshal(snapshotBytes, &baseSnapshot)

	// 应用变更
	newSnapshot := make(map[string]interface{})
	for k, v := range baseSnapshot {
		newSnapshot[k] = v
	}
	for k, v := range req.Changes {
		newSnapshot[k] = v
	}

	// 获取新版本号
	var newVersion int
	err = tx.QueryRowContext(ctx,
		"SELECT COALESCE(MAX(version), 0) + 1 FROM organization_unit_versions WHERE tenant_id = $1 AND organization_code = $2",
		tenantID.String(), orgCode).Scan(&newVersion)
	if err != nil {
		return nil, fmt.Errorf("生成新版本号失败: %w", err)
	}

	// 序列化新快照
	newSnapshotBytes, _ := json.Marshal(newSnapshot)

	// 插入新版本
	var versionID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO organization_unit_versions (
			organization_code, version, effective_from, effective_to,
			snapshot_data, change_reason, tenant_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, orgCode, newVersion, req.EffectiveFrom, req.EffectiveTo,
	   newSnapshotBytes, req.ChangeReason, tenantID.String()).Scan(&versionID)

	if err != nil {
		return nil, fmt.Errorf("创建新版本失败: %w", err)
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("提交事务失败: %w", err)
	}

	// 返回创建的版本
	return r.GetVersionByID(ctx, tenantID, versionID)
}

// 预览版本创建
func (r *VersionRepository) previewVersionCreation(ctx context.Context, tenantID uuid.UUID, orgCode string, req *CreateVersionRequest) (*VersionedOrganization, error) {
	// 实现版本创建预览逻辑
	baseVersion := 1
	if req.BasedOnVersion != nil {
		baseVersion = *req.BasedOnVersion
	}

	// 模拟版本创建，返回预览结果
	preview := &VersionedOrganization{
		OrganizationCode: orgCode,
		Version:          baseVersion + 1, // 预测版本号
		EffectiveFrom:    req.EffectiveFrom,
		EffectiveTo:      req.EffectiveTo,
		ChangeReason:     req.ChangeReason,
		SnapshotData:     req.Changes,
		TenantID:         tenantID.String(),
		TemporalStatus:   "preview",
	}

	return preview, nil
}

// 根据ID获取版本
func (r *VersionRepository) GetVersionByID(ctx context.Context, tenantID uuid.UUID, versionID string) (*VersionedOrganization, error) {
	query := `
		SELECT id, organization_code, version, effective_from, effective_to,
		       snapshot_data, change_reason, created_at, tenant_id
		FROM organization_unit_versions
		WHERE tenant_id = $1 AND id = $2
	`

	var version VersionedOrganization
	var snapshotBytes []byte

	err := r.db.QueryRowContext(ctx, query, tenantID.String(), versionID).Scan(
		&version.ID, &version.OrganizationCode, &version.Version,
		&version.EffectiveFrom, &version.EffectiveTo,
		&snapshotBytes, &version.ChangeReason,
		&version.CreatedAt, &version.TenantID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("版本不存在: %s", versionID)
		}
		return nil, fmt.Errorf("查询版本失败: %w", err)
	}

	// 解析快照数据
	if len(snapshotBytes) > 0 {
		json.Unmarshal(snapshotBytes, &version.SnapshotData)
	}

	return &version, nil
}

// 高级版本对比
func (r *VersionRepository) AdvancedCompareVersions(ctx context.Context, tenantID uuid.UUID, orgCode string, fromVersion, toVersion int, options *AdvancedComparisonOptions) (*DetailedVersionComparison, error) {
	// 获取两个版本的数据
	versions, err := r.GetSpecificVersions(ctx, tenantID, orgCode, []int{fromVersion, toVersion})
	if err != nil {
		return nil, fmt.Errorf("获取版本数据失败: %w", err)
	}

	if len(versions) < 2 {
		return nil, fmt.Errorf("无法找到指定的版本进行对比")
	}

	var fromData, toData map[string]interface{}
	for _, v := range versions {
		if v.Version == fromVersion {
			fromData = v.SnapshotData
		} else if v.Version == toVersion {
			toData = v.SnapshotData
		}
	}

	// 执行高级对比
	comparison := &DetailedVersionComparison{
		FromVersion:       fromVersion,
		ToVersion:         toVersion,
		ComparedAt:        time.Now(),
		ComparisonOptions: *options,
	}

	// 获取所有字段进行对比
	allFields := make(map[string]bool)
	for field := range fromData {
		allFields[field] = true
	}
	for field := range toData {
		allFields[field] = true
	}

	// 应用字段过滤
	if len(options.CompareFields) > 0 {
		filteredFields := make(map[string]bool)
		for _, field := range options.CompareFields {
			if allFields[field] {
				filteredFields[field] = true
			}
		}
		allFields = filteredFields
	}

	// 移除忽略的字段
	for _, field := range options.IgnoreFields {
		delete(allFields, field)
	}

	// 执行字段对比
	for field := range allFields {
		change := r.compareField(field, fromData[field], toData[field], options.SemanticComparison)
		
		if change.ChangeType != "unchanged" || options.ShowUnchanged {
			comparison.FieldChanges = append(comparison.FieldChanges, change)
		}

		// 更新统计
		comparison.Summary.TotalFields++
		switch change.ChangeType {
		case "added":
			comparison.Summary.AddedFields++
		case "removed":
			comparison.Summary.RemovedFields++
		case "modified":
			comparison.Summary.ChangedFields++
		case "unchanged":
			comparison.Summary.UnchangedFields++
		}

		// 计算影响级别
		switch change.Significance {
		case "critical":
			comparison.Summary.CriticalChanges++
		case "major":
			comparison.Summary.MajorChanges++
		case "minor":
			comparison.Summary.MinorChanges++
		}
	}

	// 分析影响
	comparison.Impact = r.analyzeComparisonImpact(&comparison.Summary, comparison.FieldChanges)

	// 生成建议
	comparison.Recommendations = r.generateComparisonRecommendations(comparison.Impact, comparison.FieldChanges)

	return comparison, nil
}

// 比较单个字段
func (r *VersionRepository) compareField(field string, oldValue, newValue interface{}, semantic bool) DetailedFieldChange {
	change := DetailedFieldChange{
		Field:     field,
		OldValue:  oldValue,
		NewValue:  newValue,
		FieldType: r.getFieldType(newValue),
	}

	oldExists := oldValue != nil
	newExists := newValue != nil

	if !oldExists && newExists {
		change.ChangeType = "added"
		change.HumanReadable = fmt.Sprintf("添加了字段 '%s'，值为 '%v'", field, newValue)
	} else if oldExists && !newExists {
		change.ChangeType = "removed"
		change.HumanReadable = fmt.Sprintf("移除了字段 '%s'，原值为 '%v'", field, oldValue)
	} else if oldExists && newExists {
		oldJSON, _ := json.Marshal(oldValue)
		newJSON, _ := json.Marshal(newValue)
		
		if string(oldJSON) != string(newJSON) {
			change.ChangeType = "modified"
			change.HumanReadable = fmt.Sprintf("修改了字段 '%s'，从 '%v' 改为 '%v'", field, oldValue, newValue)
		} else {
			change.ChangeType = "unchanged"
			change.HumanReadable = fmt.Sprintf("字段 '%s' 未变更", field)
		}
	} else {
		change.ChangeType = "unchanged"
	}

	// 确定字段变更的重要性
	change.Significance = r.determineFieldSignificance(field, change.ChangeType)
	
	return change
}

// 获取字段类型
func (r *VersionRepository) getFieldType(value interface{}) string {
	switch value.(type) {
	case string:
		return "string"
	case int, int64, float64:
		return "number"
	case bool:
		return "boolean"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	case time.Time:
		return "date"
	default:
		return "unknown"
	}
}

// 确定字段重要性
func (r *VersionRepository) determineFieldSignificance(field, changeType string) string {
	criticalFields := map[string]bool{
		"name": true, "status": true, "unit_type": true, "parent_code": true,
	}
	
	majorFields := map[string]bool{
		"level": true, "path": true, "effective_from": true, "effective_to": true,
	}

	if changeType == "removed" || changeType == "added" {
		if criticalFields[field] {
			return "critical"
		} else if majorFields[field] {
			return "major"
		}
	} else if changeType == "modified" {
		if criticalFields[field] {
			return "critical"
		} else if majorFields[field] {
			return "major"
		}
	}

	return "minor"
}

// 分析对比影响
func (r *VersionRepository) analyzeComparisonImpact(summary *ComparisonSummary, changes []DetailedFieldChange) ComparisonImpact {
	impact := ComparisonImpact{
		AffectedSystems: []string{},
		RequiredActions: []string{},
	}

	// 根据变更数量和重要性评估影响
	if summary.CriticalChanges > 0 {
		impact.StructuralImpact = "high"
		impact.BusinessImpact = "high"
		impact.DataIntegrityRisk = "high"
		impact.RequiredActions = append(impact.RequiredActions, "需要管理层审批", "更新相关系统配置", "通知所有利益相关方")
	} else if summary.MajorChanges > 3 {
		impact.StructuralImpact = "medium"
		impact.BusinessImpact = "medium"
		impact.DataIntegrityRisk = "medium"
		impact.RequiredActions = append(impact.RequiredActions, "需要部门审批", "更新相关文档")
	} else {
		impact.StructuralImpact = "low"
		impact.BusinessImpact = "low"
		impact.DataIntegrityRisk = "low"
	}

	// 根据具体变更确定受影响的系统
	for _, change := range changes {
		switch change.Field {
		case "name", "unit_type":
			impact.AffectedSystems = append(impact.AffectedSystems, "HR系统", "薪资系统", "报告系统")
		case "parent_code", "path", "level":
			impact.AffectedSystems = append(impact.AffectedSystems, "组织架构系统", "权限系统")
		case "status":
			impact.AffectedSystems = append(impact.AffectedSystems, "所有业务系统")
		}
	}

	// 去重
	systemMap := make(map[string]bool)
	for _, system := range impact.AffectedSystems {
		systemMap[system] = true
	}
	impact.AffectedSystems = make([]string, 0, len(systemMap))
	for system := range systemMap {
		impact.AffectedSystems = append(impact.AffectedSystems, system)
	}

	return impact
}

// 生成对比建议
func (r *VersionRepository) generateComparisonRecommendations(impact ComparisonImpact, changes []DetailedFieldChange) []string {
	var recommendations []string

	if impact.StructuralImpact == "high" {
		recommendations = append(recommendations, "建议进行全面的影响评估")
		recommendations = append(recommendations, "建议制定详细的变更实施计划")
	}

	if impact.DataIntegrityRisk == "high" {
		recommendations = append(recommendations, "建议进行数据一致性验证")
		recommendations = append(recommendations, "建议创建数据回滚方案")
	}

	criticalChanges := 0
	for _, change := range changes {
		if change.Significance == "critical" {
			criticalChanges++
		}
	}

	if criticalChanges > 0 {
		recommendations = append(recommendations, fmt.Sprintf("检测到%d个关键变更，建议分阶段实施", criticalChanges))
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "变更影响较小，可以正常实施")
	}

	return recommendations
}

// 获取特定版本列表
func (r *VersionRepository) GetSpecificVersions(ctx context.Context, tenantID uuid.UUID, orgCode string, versions []int) ([]VersionedOrganization, error) {
	if len(versions) == 0 {
		return []VersionedOrganization{}, nil
	}

	// 构建IN子句
	placeholders := make([]string, len(versions))
	args := []interface{}{tenantID.String(), orgCode}
	for i, v := range versions {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args = append(args, v)
	}

	query := fmt.Sprintf(`
		SELECT id, organization_code, version, effective_from, effective_to,
		       snapshot_data, change_reason, created_at, tenant_id
		FROM organization_unit_versions
		WHERE tenant_id = $1 AND organization_code = $2 AND version IN (%s)
		ORDER BY version
	`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询特定版本失败: %w", err)
	}
	defer rows.Close()

	var result []VersionedOrganization
	for rows.Next() {
		var version VersionedOrganization
		var snapshotBytes []byte

		err := rows.Scan(
			&version.ID, &version.OrganizationCode, &version.Version,
			&version.EffectiveFrom, &version.EffectiveTo,
			&snapshotBytes, &version.ChangeReason,
			&version.CreatedAt, &version.TenantID,
		)
		if err != nil {
			return nil, fmt.Errorf("扫描版本数据失败: %w", err)
		}

		if len(snapshotBytes) > 0 {
			json.Unmarshal(snapshotBytes, &version.SnapshotData)
		}

		result = append(result, version)
	}

	return result, nil
}

// ===== HTTP处理器 =====

type VersionHandler struct {
	repo *VersionRepository
}

func NewVersionHandler(db *sql.DB) *VersionHandler {
	return &VersionHandler{
		repo: NewVersionRepository(db),
	}
}

// Prometheus指标
var (
	versionRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "version_requests_total",
			Help: "Total number of version management requests",
		},
		[]string{"operation", "status"},
	)
	versionRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "version_request_duration_seconds",
			Help: "Version management request duration in seconds",
		},
		[]string{"operation"},
	)
)

func init() {
	prometheus.MustRegister(versionRequestsTotal)
	prometheus.MustRegister(versionRequestDuration)
}

func (h *VersionHandler) getTenantID(r *http.Request) uuid.UUID {
	tenantHeader := r.Header.Get("X-Tenant-ID")
	if tenantHeader != "" {
		if tenantID, err := uuid.Parse(tenantHeader); err == nil {
			return tenantID
		}
	}
	return uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
}

func (h *VersionHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := map[string]interface{}{
		"error_code": errorCode,
		"message":    message,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	
	if details != nil {
		response["details"] = details.Error()
	}
	
	json.NewEncoder(w).Encode(response)
}

// 获取所有版本
func (h *VersionHandler) GetAllVersions(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(versionRequestDuration.WithLabelValues("get_all_versions"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		versionRequestsTotal.WithLabelValues("get_all_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	tenantID := h.getTenantID(r)

	versions, err := h.repo.GetAllVersions(r.Context(), tenantID, orgCode)
	if err != nil {
		versionRequestsTotal.WithLabelValues("get_all_versions", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "VERSION_QUERY_ERROR", "获取版本列表失败", err)
		return
	}

	response := map[string]interface{}{
		"organization_code": orgCode,
		"versions":          versions,
		"version_count":     len(versions),
		"queried_at":        time.Now().Format(time.RFC3339),
	}

	versionRequestsTotal.WithLabelValues("get_all_versions", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// 创建新版本
func (h *VersionHandler) CreateVersion(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(versionRequestDuration.WithLabelValues("create_version"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		versionRequestsTotal.WithLabelValues("create_version", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req CreateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		versionRequestsTotal.WithLabelValues("create_version", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}

	// 验证请求
	if len(req.Changes) == 0 {
		versionRequestsTotal.WithLabelValues("create_version", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "NO_CHANGES", "没有要应用的变更", nil)
		return
	}

	if req.ChangeReason == "" {
		versionRequestsTotal.WithLabelValues("create_version", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_REASON", "缺少变更原因", nil)
		return
	}

	tenantID := h.getTenantID(r)

	version, err := h.repo.CreateVersion(r.Context(), tenantID, orgCode, &req)
	if err != nil {
		versionRequestsTotal.WithLabelValues("create_version", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "VERSION_CREATE_ERROR", "创建版本失败", err)
		return
	}

	versionRequestsTotal.WithLabelValues("create_version", "success").Inc()
	
	statusCode := http.StatusCreated
	if req.PreviewOnly {
		statusCode = http.StatusOK
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(version)
}

// 高级版本对比
func (h *VersionHandler) AdvancedCompareVersions(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(versionRequestDuration.WithLabelValues("advanced_compare"))
	defer timer.ObserveDuration()

	orgCode := chi.URLParam(r, "code")
	if orgCode == "" {
		versionRequestsTotal.WithLabelValues("advanced_compare", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	fromVersionStr := r.URL.Query().Get("from_version")
	toVersionStr := r.URL.Query().Get("to_version")

	if fromVersionStr == "" || toVersionStr == "" {
		versionRequestsTotal.WithLabelValues("advanced_compare", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_VERSIONS", "缺少版本参数", nil)
		return
	}

	fromVersion, err := strconv.Atoi(fromVersionStr)
	if err != nil {
		versionRequestsTotal.WithLabelValues("advanced_compare", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_FROM_VERSION", "无效的起始版本", err)
		return
	}

	toVersion, err := strconv.Atoi(toVersionStr)
	if err != nil {
		versionRequestsTotal.WithLabelValues("advanced_compare", "failed").Inc()
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_TO_VERSION", "无效的目标版本", err)
		return
	}

	// 解析对比选项
	options := &AdvancedComparisonOptions{}
	if compareFields := r.URL.Query().Get("compare_fields"); compareFields != "" {
		options.CompareFields = strings.Split(compareFields, ",")
	}
	if ignoreFields := r.URL.Query().Get("ignore_fields"); ignoreFields != "" {
		options.IgnoreFields = strings.Split(ignoreFields, ",")
	}
	options.SemanticComparison = r.URL.Query().Get("semantic") == "true"
	options.ShowUnchanged = r.URL.Query().Get("show_unchanged") == "true"
	options.ComparisonFormat = r.URL.Query().Get("format")
	if options.ComparisonFormat == "" {
		options.ComparisonFormat = "diff"
	}

	tenantID := h.getTenantID(r)

	comparison, err := h.repo.AdvancedCompareVersions(r.Context(), tenantID, orgCode, fromVersion, toVersion, options)
	if err != nil {
		versionRequestsTotal.WithLabelValues("advanced_compare", "failed").Inc()
		h.writeErrorResponse(w, http.StatusInternalServerError, "COMPARISON_ERROR", "版本对比失败", err)
		return
	}

	versionRequestsTotal.WithLabelValues("advanced_compare", "success").Inc()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// ===== 主程序 =====

func main() {
	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal("数据库连接测试失败:", err)
	}

	log.Println("✅ 数据库连接成功")

	// 创建处理器
	handler := NewVersionHandler(db)

	// 设置路由
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// 健康检查
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   "organization-version-management-service",
			"version":   "1.0.0",
			"timestamp": time.Now().Format(time.RFC3339),
			"features": []string{
				"version-creation", "version-preview", "advanced-comparison",
				"impact-analysis", "batch-operations", "temporal-corrections",
			},
		})
	})

	// 监控指标
	r.Handle("/metrics", promhttp.Handler())

	// API路由
	r.Route("/api/v1/organization-units/{code}", func(r chi.Router) {
		r.Get("/versions", handler.GetAllVersions)                    // 获取所有版本
		r.Post("/versions", handler.CreateVersion)                    // 创建新版本
		r.Get("/versions/compare", handler.AdvancedCompareVersions)   // 高级版本对比
	})

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "9093" // 使用9093端口避免冲突
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 优雅关闭
	go func() {
		log.Printf("🚀 组织版本管理服务启动在端口 %s", port)
		log.Println("📋 支持的功能:")
		log.Println("  - 版本创建和预览 (基于变更的增量版本)")
		log.Println("  - 高级版本对比 (字段级差异分析)")
		log.Println("  - 影响评估 (结构化影响分析)")
		log.Println("  - 智能建议 (基于变更模式的建议)")
		log.Println("  - Prometheus监控指标")
		
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("服务器启动失败:", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭版本管理服务...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("服务器强制关闭:", err)
	}

	log.Println("版本管理服务已关闭")
}