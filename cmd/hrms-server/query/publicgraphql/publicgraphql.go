package publicgraphql

// Public adapter that exposes a GraphQL http.Handler for use by the unified
// monolith server. This package lives under cmd/hrms-server/query so it can
// legally import the internal gqlgen runtime while providing a non-internal
// import path to the rest of the repository.

import (
	"net/http"
	"os"

	graphqlruntime "cube-castle/cmd/hrms-server/query/internal/graphql"
	graphqlresolver "cube-castle/cmd/hrms-server/query/internal/graphql/resolver"
	"cube-castle/internal/auth"
	"cube-castle/internal/config"
	// schemaLoader "cube-castle/internal/graphql"
	"cube-castle/internal/middleware"
	organization "cube-castle/internal/organization"
	pkglogger "cube-castle/pkg/logger"
	"database/sql"

	"github.com/99designs/gqlgen/graphql/handler"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

// BuildHandlers builds the http.Handlers for /graphql and /graphiql (optional).
// Health/metrics should由上层统一暴露，避免重复端点。
func BuildHandlers(sqlDB *sql.DB, repo organization.QueryRepositoryInterface, assignments organization.AssignmentFacade, logger pkglogger.Logger, devMode bool) (graphql http.Handler, graphiql http.Handler, err error) {
	jwtCfg := config.GetJWTConfig()
	var pubPEM []byte
	if jwtCfg.HasPublicKey() {
		b, err := os.ReadFile(jwtCfg.PublicKeyPath)
		if err != nil {
			return nil, nil, err
		}
		pubPEM = b
	}
	jwtMiddleware := auth.NewJWTMiddlewareWithOptions(jwtCfg.Secret, jwtCfg.Issuer, jwtCfg.Audience, auth.Options{
		Alg:          jwtCfg.Algorithm,
		JWKSURL:      jwtCfg.JWKSUrl,
		PublicKeyPEM: pubPEM,
		ClockSkew:    jwtCfg.AllowedClockSkew,
	})
	authLogger := logger.WithFields(pkglogger.Fields{"component": "graphql-auth"})
	permissionChecker := auth.NewPBACPermissionChecker(sqlDB, authLogger)
	graphqlPerm := auth.NewGraphQLPermissionMiddleware(jwtMiddleware, permissionChecker, authLogger, devMode)

	// Resolver wiring
	qr := organization.NewQueryResolver(repo, assignments, logger, graphqlPerm)
	gqlResolver := graphqlresolver.New(qr)
	executableSchema := graphqlruntime.NewExecutableSchema(graphqlruntime.Config{
		Resolvers: gqlResolver,
	})
	graphqlServer := handler.NewDefaultServer(executableSchema)

	envelope := middleware.NewGraphQLEnvelopeMiddleware()
	baseGraphQLHandler := envelope.Middleware()(graphqlPerm.Middleware()(graphqlServer))
	graphql = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ww := chiMiddleware.NewWrapResponseWriter(w, req.ProtoMajor)
		baseGraphQLHandler.ServeHTTP(ww, req)
		status := ww.Status()
		if status == 0 {
			status = http.StatusOK
		}
		organization.RecordHTTPRequest(req.Method, "/graphql", status)
	})

	if devMode {
		gi := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`<html><body><div>Open GraphiQL with your tooling</div></body></html>`))
		})
		graphiql = gi
	}

	return graphql, graphiql, nil
}

// Adapter to expose minimal DB handle needed by PBAC checker when db is a facade.
type RedisFacade interface {
	Underlying() *redis.Client
}
