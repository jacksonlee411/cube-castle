package config

import (
	"fmt"
	"os"
	"strings"
)

// ResolveAllowedOrigins 依据优先级解析 CORS 允许的来源：
// 1) primaryEnv（逗号分隔）
// 2) fallbackEnv（可选，逗号分隔）
// 3) defaults（非空即返回）
// 4) 最终回落到 "*"
func ResolveAllowedOrigins(primaryEnv, fallbackEnv string, defaults []string) []string {
	if origins := parseOrigins(os.Getenv(primaryEnv)); len(origins) > 0 {
		return origins
	}
	if fallbackEnv != "" {
		if origins := parseOrigins(os.Getenv(fallbackEnv)); len(origins) > 0 {
			return origins
		}
	}
	if len(defaults) > 0 {
		return defaults
	}
	return []string{"*"}
}

// BuildOrigin 根据 scheme/host/port 生成单个 origin（用于默认值）
func BuildOrigin(scheme, host, port string) string {
	cleanScheme := strings.TrimSpace(scheme)
	if cleanScheme == "" {
		cleanScheme = "http"
	}
	cleanHost := strings.TrimSpace(host)
	if cleanHost == "" {
		cleanHost = "127.0.0.1"
	}
	cleanPort := strings.TrimPrefix(strings.TrimSpace(port), ":")
	if cleanPort == "" {
		return fmt.Sprintf("%s://%s", cleanScheme, cleanHost)
	}
	return fmt.Sprintf("%s://%s:%s", cleanScheme, cleanHost, cleanPort)
}

func parseOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	var origins []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}
