package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cube-castle/internal/organization/middleware"
	"cube-castle/internal/organization/repository"
	"cube-castle/internal/organization/utils"
	pkglogger "cube-castle/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *OrganizationHandler) CreateOrganizationEvent(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if code == "" {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
		return
	}

	var req struct {
		EventType     string `json:"eventType"`
		RecordID      string `json:"recordId"`
		EffectiveDate string `json:"effectiveDate"`
		ChangeReason  string `json:"changeReason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
		return
	}
	eventType := strings.TrimSpace(req.EventType)
	logger := h.requestLogger(r, "CreateOrganizationEvent", pkglogger.Fields{
		"code":      code,
		"eventType": eventType,
	})

	tenantID := h.getTenantID(r)
	operationReason := strings.TrimSpace(req.ChangeReason)

	switch eventType {
	case "DEACTIVATE":
		if strings.TrimSpace(req.RecordID) == "" {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "MISSING_RECORD_ID", "缺少记录ID", nil)
			return
		}

		// 处理版本作废事件
		actorID := h.getActorID(r)
		requestID := middleware.GetRequestID(r.Context())

		err := h.handleDeactivateEvent(r.Context(), tenantID, code, req.RecordID, operationReason, actorID, requestID)
		if err != nil {
			if errors.Is(err, repository.ErrOrganizationHasChildren) {
				details := map[string]interface{}{
					"resolution": "Delete or reassign child units first",
				}
				var childErr *repository.OrganizationHasChildrenError
				if errors.As(err, &childErr) && childErr.Count > 0 {
					details["affectedCount"] = childErr.Count
				}
				h.writeErrorResponse(w, r, http.StatusConflict, "HAS_CHILD_UNITS", "Cannot delete organization unit with child units", details)
				return
			}
			logger.WithFields(pkglogger.Fields{"error": err}).Error("deactivate organization version failed")
			h.writeErrorResponse(w, r, http.StatusInternalServerError, "DEACTIVATE_ERROR", "作废版本失败", err)
			return
		}

		// 获取最新时间线（非删除记录），用于前端立即刷新，避免读缓存延迟
		versions, listErr := h.repo.ListVersionsByCode(r.Context(), tenantID, code)
		if listErr != nil {
			logger.WithFields(pkglogger.Fields{"error": listErr}).Warn("fetch latest timeline failed (non-blocking)")
		}

		// 构建轻量时间线返回
		timeline := make([]map[string]interface{}, 0, len(versions))
		for _, v := range versions {
			timeline = append(timeline, map[string]interface{}{
				"recordId":  v.RecordID,
				"code":      v.Code,
				"name":      v.Name,
				"unitType":  v.UnitType,
				"status":    v.Status,
				"level":     v.Level,
				"codePath":  v.CodePath,
				"namePath":  v.NamePath,
				"sortOrder": v.SortOrder,
				"effectiveDate": func() string {
					if v.EffectiveDate != nil {
						return v.EffectiveDate.String()
					}
					return ""
				}(),
				"endDate": func() *string {
					if v.EndDate != nil {
						s := v.EndDate.String()
						return &s
					}
					return nil
				}(),
				"isCurrent":   v.IsCurrent,
				"createdAt":   v.CreatedAt,
				"updatedAt":   v.UpdatedAt,
				"parentCode":  v.ParentCode,
				"description": v.Description,
			})
		}

		logger.WithFields(pkglogger.Fields{
			"timelineCount": len(timeline),
		}).Info("organization version deactivated")
		if err := utils.WriteSuccess(w, map[string]interface{}{
			"code":      code,
			"record_id": req.RecordID,
			"timeline":  timeline,
		}, "版本作废成功", requestID); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("write deactivate response failed")
		}

	case "DELETE_ORGANIZATION":
		actorID := h.getActorID(r)
		requestID := middleware.GetRequestID(r.Context())

		if strings.TrimSpace(req.EffectiveDate) == "" {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "缺少生效日期", nil)
			return
		}

		effectiveDate, err := time.Parse("2006-01-02", strings.TrimSpace(req.EffectiveDate))
		if err != nil {
			h.writeErrorResponse(w, r, http.StatusBadRequest, "INVALID_EFFECTIVE_DATE", "生效日期格式无效，应为YYYY-MM-DD", err)
			return
		}

		ifMatch, err := h.getIfMatchValue(r)
		if err != nil {
			h.writeErrorResponse(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "缺少或无效的 If-Match 标头", err)
			return
		}

		currentOrg, err := h.repo.GetByCode(r.Context(), tenantID, code)
		if err != nil {
			h.handleRepositoryError(w, r, "GET_FOR_DELETE", err)
			return
		}

		expectedETag := strings.TrimSpace(currentOrg.RecordID)
		if expectedETag == "" {
			expectedETag = currentOrg.UpdatedAt.Format(time.RFC3339Nano)
		}

		if ifMatch != expectedETag {
			h.writeErrorResponse(w, r, http.StatusPreconditionFailed, "PRECONDITION_FAILED", "资源已发生变更，请刷新后重试", map[string]interface{}{
				"expected": expectedETag,
				"provided": ifMatch,
			})
			return
		}

		childCount, err := h.repo.CountNonDeletedChildren(r.Context(), tenantID, code)
		if err != nil {
			h.handleRepositoryError(w, r, "COUNT_CHILDREN", err)
			return
		}

		if childCount > 0 {
			h.writeErrorResponse(w, r, http.StatusConflict, "HAS_CHILD_UNITS", "Cannot delete organization unit with child units", map[string]interface{}{
				"affectedCount": childCount,
				"resolution":    "Delete or reassign child units first",
			})
			return
		}

		deletionMoment := time.Date(effectiveDate.Year(), effectiveDate.Month(), effectiveDate.Day(), 0, 0, 0, 0, time.UTC)

		if err := h.repo.SoftDeleteOrganization(r.Context(), tenantID, code, deletionMoment, actorID, operationReason); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				h.writeErrorResponse(w, r, http.StatusNotFound, "ORGANIZATION_NOT_FOUND", "组织单元不存在或已删除", err)
				return
			}
			logger.WithFields(pkglogger.Fields{"error": err}).Error("soft delete organization failed")
			h.handleRepositoryError(w, r, "DELETE", err)
			return
		}

		if err := h.auditLogger.LogOrganizationDelete(r.Context(), tenantID, code, currentOrg, actorID, requestID, operationReason); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Warn("record organization delete audit log failed")
		}

		responseData := map[string]interface{}{
			"code":            code,
			"status":          "DELETED",
			"operationType":   "DELETE_ORGANIZATION",
			"record_id":       nil,
			"effectiveDate":   effectiveDate.Format("2006-01-02"),
			"operationReason": operationReason,
			"timeline":        []map[string]interface{}{},
		}

		if err := utils.WriteSuccess(w, responseData, "组织删除成功", requestID); err != nil {
			logger.WithFields(pkglogger.Fields{"error": err}).Error("write organization delete response failed")
		}

		logger.WithFields(pkglogger.Fields{"tenantId": tenantID}).Info("organization deleted")

	default:
		h.writeErrorResponse(w, r, http.StatusBadRequest, "UNSUPPORTED_EVENT", fmt.Sprintf("不支持的事件类型: %s", req.EventType), nil)
	}
}

func (h *OrganizationHandler) handleDeactivateEvent(ctx context.Context, tenantID uuid.UUID, code string, recordID string, changeReason string, actorID string, requestID string) error {
	// 验证UUID格式
	if _, err := uuid.Parse(recordID); err != nil {
		return fmt.Errorf("无效的记录ID格式: %w", err)
	}

	// 获取删除前的组织数据用于审计日志
	oldOrg, err := h.repo.GetByRecordId(ctx, tenantID, recordID)
	if err != nil {
		return fmt.Errorf("获取记录失败: %w", err)
	}

	if oldOrg != nil {
		hasOtherVersions, err := h.repo.HasOtherNonDeletedVersions(ctx, tenantID, oldOrg.Code, recordID)
		if err != nil {
			return fmt.Errorf("检查版本数量失败: %w", err)
		}
		if !hasOtherVersions {
			childCount, err := h.repo.CountNonDeletedChildren(ctx, tenantID, oldOrg.Code)
			if err != nil {
				return fmt.Errorf("检查子组织失败: %w", err)
			}
			if childCount > 0 {
				return repository.NewOrganizationHasChildrenError(childCount)
			}
		}
	}

	// 使用时间线管理器执行“单事务 软删 + 全链重算”
	rid, _ := uuid.Parse(recordID)
	if _, err := h.timelineManager.DeleteVersion(ctx, tenantID, rid); err != nil {
		return fmt.Errorf("作废记录失败: %w", err)
	}

	// 记录审计日志 - 使用删除日志方法
	err = h.auditLogger.LogOrganizationDelete(ctx, tenantID, code, oldOrg, actorID, requestID, changeReason)
	if err != nil {
		h.logger.WithFields(pkglogger.Fields{"error": err, "recordId": recordID}).Warn("audit log for organization version delete failed")
		// 审计日志失败不应该导致业务操作失败，只记录警告
	} else {
		h.logger.WithFields(pkglogger.Fields{"recordId": recordID}).Info("audit log recorded for organization version delete")
	}

	return nil
}
