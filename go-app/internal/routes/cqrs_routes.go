package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/handlers"
)

// SetupCQRSRoutes 设置CQRS路由
func SetupCQRSRoutes(r chi.Router, cmdHandler *handlers.CommandHandler, queryHandler *handlers.QueryHandler) {
	// 命令端点 - 所有写操作
	r.Route("/commands", func(r chi.Router) {
		// 员工管理命令
		r.Post("/hire-employee", cmdHandler.HireEmployee)
		r.Put("/update-employee", cmdHandler.UpdateEmployee)
		r.Post("/terminate-employee", cmdHandler.TerminateEmployee)
		
		// 组织管理命令 (新实现)
		r.Post("/organizations", cmdHandler.CreateOrganization)
		r.Put("/organizations/{id}", cmdHandler.UpdateOrganization)
		r.Delete("/organizations/{id}", cmdHandler.DeleteOrganization)
		
		// 组织单元管理命令 (向后兼容)
		r.Post("/create-organization-unit", cmdHandler.CreateOrganizationUnit)
		r.Put("/update-organization-unit", cmdHandler.UpdateOrganizationUnit)
		r.Delete("/delete-organization-unit", cmdHandler.DeleteOrganizationUnit)
		
		// 职位管理命令
		r.Post("/assign-employee-position", cmdHandler.AssignEmployeePosition)
		r.Post("/create-position", cmdHandler.CreatePosition)
	})
	
	// 查询端点 - 所有读操作  
	r.Route("/queries", func(r chi.Router) {
		// 员工查询
		r.Get("/employees/{id}", queryHandler.GetEmployee)
		r.Get("/employees", queryHandler.SearchEmployees)
		
		// 组织查询 (新实现)
		r.Get("/organizations", queryHandler.ListOrganizations)
		r.Get("/organizations/{id}", queryHandler.GetOrganization)
		r.Get("/organization-tree", queryHandler.GetOrganizationTree)
		r.Get("/organization-stats", queryHandler.GetOrganizationStats)
		
		// 组织结构查询 (向后兼容)
		r.Get("/organization-chart", queryHandler.GetOrgChart)
		r.Get("/organization-units/{id}", queryHandler.GetOrganizationUnit)
		r.Get("/organization-units", queryHandler.ListOrganizationUnits)
		
		// 层级关系查询
		r.Get("/reporting-hierarchy/{manager_id}", queryHandler.GetReportingHierarchy)
		r.Get("/employee-path/{from_id}/{to_id}", queryHandler.FindEmployeePath)
		
		// 高级查询
		r.Get("/department-structure/{dept_id}", queryHandler.GetDepartmentStructure)
		r.Get("/common-manager", queryHandler.FindCommonManager)
	})
}