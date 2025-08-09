package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/cube-castle/cmd/organization-command-server/internal/presentation/http/handlers"
	custommiddleware "github.com/cube-castle/cmd/organization-command-server/internal/presentation/http/middleware"
	"cube-castle-deployment-test/pkg/monitoring"
)

// RouterConfig contains the dependencies needed for setting up routes
type RouterConfig struct {
	OrganizationHandler *handlers.OrganizationHTTPHandler
	HealthHandler       *handlers.HealthHandler
	RequestLogger       *custommiddleware.RequestLogger
	ErrorHandler        *custommiddleware.ErrorHandler
}

// SetupRoutes configures and returns the main HTTP router
func SetupRoutes(config RouterConfig) chi.Router {
	r := chi.NewRouter()

	// Basic middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(monitoring.MetricsMiddleware("command-server")) // 添加指标收集中间件
	r.Use(config.RequestLogger.Handle)
	r.Use(config.ErrorHandler.Handle)
	r.Use(middleware.Recoverer)

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, specify actual origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Health check routes (no authentication required)
	r.Get("/health", config.HealthHandler.CheckHealth)
	r.Get("/ready", config.HealthHandler.CheckReadiness)
	r.Get("/live", config.HealthHandler.CheckLiveness)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Organization command routes
		r.Route("/organization-units", func(r chi.Router) {
			// Add authentication middleware here in production
			// r.Use(authMiddleware)

			r.Post("/", config.OrganizationHandler.CreateOrganization)
			r.Put("/{code}", config.OrganizationHandler.UpdateOrganization)
			r.Delete("/{code}", config.OrganizationHandler.DeleteOrganization)
		})
	})

	// Catch-all for undefined routes
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		config.ErrorHandler.WriteErrorResponse(w, r,
			custommiddleware.DomainError{
				Code:    "NOT_FOUND",
				Message: "endpoint not found",
				Details: "The requested endpoint does not exist",
			}, 404)
	})

	// Method not allowed handler
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		config.ErrorHandler.WriteErrorResponse(w, r,
			custommiddleware.DomainError{
				Code:    "METHOD_NOT_ALLOWED",
				Message: "method not allowed",
				Details: "The HTTP method is not supported for this endpoint",
			}, 405)
	})

	return r
}

// SetupAPIRoutes sets up just the API routes (useful for testing)
func SetupAPIRoutes(config RouterConfig) chi.Router {
	r := chi.NewRouter()

	// Minimal middleware for API-only setup
	r.Use(middleware.RequestID)
	r.Use(config.ErrorHandler.Handle)

	// Organization command routes
	r.Route("/organization-units", func(r chi.Router) {
		r.Post("/", config.OrganizationHandler.CreateOrganization)
		r.Put("/{code}", config.OrganizationHandler.UpdateOrganization) 
		r.Delete("/{code}", config.OrganizationHandler.DeleteOrganization)
	})

	return r
}