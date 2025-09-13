module cube-castle-deployment-test/cmd/organization-query-service

go 1.23.0

toolchain go1.23.12

require (
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.3.1
	github.com/graph-gophers/graphql-go v1.5.0
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.2.1
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
)

replace cube-castle-deployment-test/pkg/health => ../../pkg/health
