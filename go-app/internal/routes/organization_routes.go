package routes

import (
	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/internal/handler"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/go-chi/chi/v5"
)

// SetupOrganizationRoutes configures organization API routes
// Maps frontend /api/v1/corehr/organizations/* to backend OrganizationUnit handlers
func SetupOrganizationRoutes(r chi.Router, client *ent.Client, logger *logging.StructuredLogger) {
	// Create organization adapter
	orgAdapter := handler.NewOrganizationAdapter(client, logger)

	// CoreHR Organization API routes (frontend compatibility)
	r.Route("/corehr/organizations", func(r chi.Router) {
		r.Get("/", orgAdapter.GetOrganizations())       // GET /api/v1/corehr/organizations
		r.Post("/", orgAdapter.CreateOrganization())    // POST /api/v1/corehr/organizations
		r.Get("/stats", orgAdapter.GetOrganizationStats()) // GET /api/v1/corehr/organizations/stats
		
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", orgAdapter.GetOrganization())    // GET /api/v1/corehr/organizations/{id}
			r.Put("/", orgAdapter.UpdateOrganization()) // PUT /api/v1/corehr/organizations/{id}
			r.Delete("/", orgAdapter.DeleteOrganization()) // DELETE /api/v1/corehr/organizations/{id}
		})
	})

	// Keep existing OrganizationUnit API routes for backward compatibility
	orgUnitHandler := handler.NewOrganizationUnitHandler(client, logger)
	r.Route("/organization-units", func(r chi.Router) {
		r.Get("/", orgUnitHandler.ListOrganizationUnits())
		r.Post("/", orgUnitHandler.CreateOrganizationUnit())
		
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", orgUnitHandler.GetOrganizationUnit())
			r.Put("/", orgUnitHandler.UpdateOrganizationUnit())
			r.Delete("/", orgUnitHandler.DeleteOrganizationUnit())
		})
	})

	logger.Info("Organization routes configured successfully",
		"corehr_prefix", "/api/v1/corehr/organizations",
		"backend_prefix", "/api/v1/organization-units",
	)
}