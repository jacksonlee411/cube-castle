module github.com/cube-castle/cmd/organization-command-service-simplified

go 1.23.12

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/google/uuid v1.4.0
	github.com/lib/pq v1.10.9
)

replace cube-castle-deployment-test/pkg/health => ../../pkg/health
