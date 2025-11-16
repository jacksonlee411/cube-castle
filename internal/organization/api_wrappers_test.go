package organization

import (
	"database/sql"
	"testing"
	"time"

	auth "cube-castle/internal/auth"
	"cube-castle/internal/organization/repository"
	"cube-castle/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func TestWrapper_Constructors_NoPanic(t *testing.T) {
	// RecordHTTPRequest forwards to utils
	RecordHTTPRequest("GET", "/health", 200)

	// NewDevToolsHandler accepts nil DB, returns handler
	h := NewDevToolsHandler((*sql.DB)(nil), (*auth.JWTMiddleware)(nil), logger.NewNoopLogger(), true)
	if h == nil {
		t.Fatalf("NewDevToolsHandler returned nil")
	}

	// NewQueryRepository allows nil db/redis and returns non-nil repo
	repo := NewQueryRepository(nil, (*redis.Client)(nil), logger.NewNoopLogger(), DefaultAuditHistoryConfig())
	if repo == nil {
		t.Fatalf("NewQueryRepository returned nil")
	}
	// NewQueryResolver with and without assignments
	res := NewQueryResolver(repo, nil, logger.NewNoopLogger(), nil)
	if res == nil {
		t.Fatalf("NewQueryResolver returned nil")
	}
	facade := NewAssignmentFacade(repo, nil, logger.NewNoopLogger(), 5*time.Minute)
	if facade == nil {
		t.Fatalf("NewAssignmentFacade returned nil")
	}
	// ensure AuditHistoryConfig default is sane
	cfg := DefaultAuditHistoryConfig()
	if !cfg.StrictValidation || cfg.CircuitBreakerThreshold <= 0 {
		t.Fatalf("unexpected default audit history config: %#v", cfg)
	}
	_ = repository.AuditHistoryConfig{} // keep import
}

