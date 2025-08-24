module postgresql-graphql-service

go 1.23.0

toolchain go1.23.12

require (
	cube-castle-deployment-test/pkg/health v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/google/uuid v1.3.1
	github.com/graph-gophers/graphql-go v1.5.0
	github.com/lib/pq v1.10.9
	github.com/prometheus/client_golang v1.17.0
	github.com/redis/go-redis/v9 v9.2.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/neo4j/neo4j-go-driver/v5 v5.28.1 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	golang.org/x/sys v0.11.0 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace cube-castle-deployment-test/pkg/health => ../../pkg/health
