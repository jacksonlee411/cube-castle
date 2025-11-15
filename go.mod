module cube-castle

go 1.24.0

toolchain go1.24.9

require (
	github.com/99designs/gqlgen v0.17.45
	github.com/DATA-DOG/go-sqlmock v1.5.0
	github.com/alicebob/miniredis/v2 v2.30.5
	github.com/gin-gonic/gin v1.9.1
	github.com/go-chi/chi/v5 v5.0.10
	github.com/go-chi/cors v1.2.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.7.5
	github.com/lib/pq v1.10.9
	github.com/pressly/goose/v3 v3.26.0
	github.com/prometheus/client_golang v1.19.1
	github.com/redis/go-redis/v9 v9.6.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.11.1
	github.com/vektah/gqlparser/v2 v2.5.11
	gopkg.in/yaml.v3 v3.0.1
)

replace (
	github.com/99designs/gqlgen => ./third_party/github.com/99designs/gqlgen
	github.com/prometheus/client_model => github.com/prometheus/client_model v0.5.0
	github.com/prometheus/common => github.com/prometheus/common v0.48.0
	github.com/prometheus/procfs => github.com/prometheus/procfs v0.12.0
	golang.org/x/net => golang.org/x/net v0.20.0
	golang.org/x/sys => golang.org/x/sys v0.17.0
)

require (
	github.com/agnivade/levenshtein v1.1.1 // indirect
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.14.0 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/goccy/go-json v0.10.2 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	github.com/sosodev/duration v1.2.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/yuin/gopher-lua v1.1.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/protobuf v1.36.9 // indirect
)
