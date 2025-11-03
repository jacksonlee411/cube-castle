package handlers

import "github.com/go-chi/chi/v5"

func (h *OrganizationHandler) SetupRoutes(r chi.Router) {
	r.Route("/api/v1/organization-units", func(r chi.Router) {
		r.Post("/", h.CreateOrganization)
		r.Put("/{code}", h.UpdateOrganization)
		r.Post("/{code}/suspend", h.SuspendOrganization)
		r.Post("/{code}/activate", h.ActivateOrganization)
		// 🚀 时态版本管理端点 - 严格遵循API契约
		r.Post("/{code}/versions", h.CreateOrganizationVersion)
		// 注意: 删除版本请使用 POST /{code}/events (DEACTIVATE)
		// 注意: 修改生效日期请使用 PUT /{code}/history/{record_id}
		// 事件处理和历史记录
		r.Post("/{code}/events", h.CreateOrganizationEvent)
		r.Put("/{code}/history/{record_id}", h.UpdateHistoryRecord)
	})
}
