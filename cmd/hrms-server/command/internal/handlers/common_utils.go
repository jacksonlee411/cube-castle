package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"cube-castle/internal/types"
)

func getTenantIDFromRequest(r *http.Request) uuid.UUID {
	tenantIDHeader := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
	if tenantIDHeader != "" {
		if tenantID, err := uuid.Parse(tenantIDHeader); err == nil {
			return tenantID
		}
	}
	return types.DefaultTenantID
}

func getOperatorFromRequest(r *http.Request) types.OperatedByInfo {
	operatorID := getActorID(r)
	operatorName := strings.TrimSpace(r.Header.Get("X-Actor-Name"))
	if operatorName == "" {
		operatorName = operatorID
	}
	return types.OperatedByInfo{
		ID:   operatorID,
		Name: operatorName,
	}
}

func getActorID(r *http.Request) string {
	if mock := strings.TrimSpace(r.Header.Get("X-Mock-User")); mock != "" {
		return mock
	}
	if val := r.Context().Value("user_id"); val != nil {
		if uid, ok := val.(string); ok && strings.TrimSpace(uid) != "" {
			return strings.TrimSpace(uid)
		}
	}
	return "system"
}

func getIfMatchHeader(r *http.Request) *string {
	raw := strings.TrimSpace(r.Header.Get("If-Match"))
	if raw == "" {
		return nil
	}
	if strings.HasPrefix(strings.ToLower(raw), "w/") {
		raw = strings.TrimSpace(raw[2:])
	}
	value := strings.Trim(raw, "\"")
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
