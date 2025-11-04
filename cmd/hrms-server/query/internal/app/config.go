package app

import (
	"os"
	"strconv"
	"strings"

	"cube-castle/internal/organization/repository"
)

func loadAuditHistoryConfig() repository.AuditHistoryConfig {
	threshold := getEnvAsInt("AUDIT_HISTORY_CIRCUIT_BREAKER_THRESHOLD", 25)
	if threshold < 0 {
		threshold = 0
	}
	return repository.AuditHistoryConfig{
		StrictValidation:        getEnvAsBool("AUDIT_HISTORY_STRICT_VALIDATION", true),
		AllowFallback:           getEnvAsBool("AUDIT_HISTORY_ALLOW_FALLBACK", true),
		CircuitBreakerThreshold: int32(threshold),
		LegacyMode:              getEnvAsBool("AUDIT_HISTORY_LEGACY_MODE", false),
	}
}

func getEnvAsBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return defaultValue
	}
}

func getEnvAsInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	if v, err := strconv.Atoi(value); err == nil {
		return v
	}
	return defaultValue
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
