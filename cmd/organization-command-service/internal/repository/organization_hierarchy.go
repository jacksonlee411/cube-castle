package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"organization-command-service/internal/utils"
)

type hierarchyFields struct {
	Path     string
	CodePath string
	NamePath string
	Level    int
	oldLevel int
}

func ensureJoinedPath(base, segment string) string {
	base = strings.TrimSpace(base)
	segment = strings.TrimSpace(segment)
	base = strings.TrimRight(base, "/")
	segment = strings.TrimLeft(segment, "/")
	if base == "" {
		return "/" + segment
	}
	return base + "/" + segment
}

func (r *OrganizationRepository) recalculateSelfHierarchy(ctx context.Context, tenantID uuid.UUID, code string, recordID *string, parentCode *string, overrideName *string) (*hierarchyFields, error) {
	var (
		resolvedCode string
		currentName  string
		currentLevel int
	)

	if recordID != nil {
		err := r.db.QueryRowContext(ctx, `
			SELECT code, name, level
			FROM organization_units
			WHERE tenant_id = $1 AND record_id = $2 AND status <> 'DELETED'
			LIMIT 1
		`, tenantID.String(), *recordID).Scan(&resolvedCode, &currentName, &currentLevel)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("记录不存在: %s", *recordID)
			}
			return nil, fmt.Errorf("查询组织记录失败: %w", err)
		}
	} else {
		resolvedCode = code
		err := r.db.QueryRowContext(ctx, `
			SELECT name, level
			FROM organization_units
			WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
			LIMIT 1
		`, tenantID.String(), code).Scan(&currentName, &currentLevel)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("组织不存在或已删除不可修改: %s", code)
			}
			return nil, fmt.Errorf("查询组织失败: %w", err)
		}
	}

	finalName := currentName
	if overrideName != nil {
		finalName = strings.TrimSpace(*overrideName)
	}

	if resolvedCode == "" {
		resolvedCode = code
	}

	fields, err := r.calculateHierarchyFields(ctx, tenantID, resolvedCode, parentCode, finalName)
	if err != nil {
		return nil, err
	}
	fields.oldLevel = currentLevel

	r.logger.Printf("recalculateSelfHierarchy: code=%s oldLevel=%d newLevel=%d path=%s", resolvedCode, fields.oldLevel, fields.Level, fields.Path)
	return fields, nil
}

func (r *OrganizationRepository) calculateHierarchyFields(ctx context.Context, tenantID uuid.UUID, code string, parentCode *string, finalName string) (*hierarchyFields, error) {
	finalName = strings.TrimSpace(finalName)
	if finalName == "" {
		return nil, fmt.Errorf("组织名称不能为空")
	}

	fields := &hierarchyFields{}

	if parentCode == nil {
		fields.Level = 1
		fields.Path = ensureJoinedPath("", code)
		fields.CodePath = fields.Path
		fields.NamePath = ensureJoinedPath("", finalName)
		return fields, nil
	}

	trimmedParent := strings.TrimSpace(*parentCode)
	if trimmedParent == "" || trimmedParent == utils.RootParentCode {
		fields.Level = 1
		fields.Path = ensureJoinedPath("", code)
		fields.CodePath = fields.Path
		fields.NamePath = ensureJoinedPath("", finalName)
		return fields, nil
	}

	var parentCodePath, parentNamePath string
	var parentLevel int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(NULLIF(code_path, ''), '/' || code),
		       COALESCE(NULLIF(name_path, ''), '/' || name),
		       level
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2 AND is_current = true AND status <> 'DELETED'
		LIMIT 1
	`, tenantID.String(), trimmedParent).Scan(&parentCodePath, &parentNamePath, &parentLevel)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("父组织不存在: %s", trimmedParent)
		}
		return nil, fmt.Errorf("查询父组织失败: %w", err)
	}

	fields.Level = parentLevel + 1
	fields.Path = ensureJoinedPath(parentCodePath, code)
	fields.CodePath = fields.Path
	fields.NamePath = ensureJoinedPath(parentNamePath, finalName)

	return fields, nil
}

// ComputeHierarchyForNew 计算新建或新版本的层级字段（path/codePath/namePath/level）
func (r *OrganizationRepository) ComputeHierarchyForNew(ctx context.Context, tenantID uuid.UUID, code string, parentCode *string, name string) (*hierarchyFields, error) {
	return r.calculateHierarchyFields(ctx, tenantID, strings.TrimSpace(code), parentCode, name)
}
