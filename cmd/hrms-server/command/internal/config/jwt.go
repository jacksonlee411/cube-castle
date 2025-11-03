package config

import (
	"os"
	"strings"
	"time"
)

// JWTConfig 统一JWT配置结构
type JWTConfig struct {
	Secret           string
	Issuer           string
	Audience         string
	Algorithm        string
	PublicKeyPath    string
	PrivateKeyPath   string
	JWKSUrl          string
	KeyID            string
	AllowedClockSkew time.Duration
}

// GetJWTConfig 获取统一JWT配置
// 消除6个文件中的重复JWT配置逻辑
func GetJWTConfig() *JWTConfig {
	config := &JWTConfig{}

	// JWT密钥配置
	config.Secret = os.Getenv("JWT_SECRET")
	if config.Secret == "" {
		config.Secret = "cube-castle-development-secret-key-2025"
	}

	// JWT发行者配置
	config.Issuer = os.Getenv("JWT_ISSUER")
	if config.Issuer == "" {
		config.Issuer = "cube-castle"
	}

	// JWT受众配置
	config.Audience = os.Getenv("JWT_AUDIENCE")
	if config.Audience == "" {
		config.Audience = "cube-castle-users"
	}

	// JWT算法配置（强制使用 RS256，保持与查询服务一致）
	config.Algorithm = strings.ToUpper(strings.TrimSpace(os.Getenv("JWT_ALG")))
	if config.Algorithm == "" {
		config.Algorithm = "RS256"
	}
	if config.Algorithm != "RS256" {
		panic("JWT_ALG 只能配置为 RS256，已禁止 HS256 混用。请更新环境配置。")
	}

	// RS256公钥路径配置
	config.PublicKeyPath = os.Getenv("JWT_PUBLIC_KEY_PATH")

	// 私钥与JWKS配置
	config.PrivateKeyPath = os.Getenv("JWT_PRIVATE_KEY_PATH")
	config.JWKSUrl = os.Getenv("JWT_JWKS_URL")
	config.KeyID = os.Getenv("JWT_KEY_ID")

	// 时钟偏差容忍配置
	clockSkewStr := os.Getenv("JWT_ALLOWED_CLOCK_SKEW")
	if clockSkewStr != "" {
		if duration, err := time.ParseDuration(clockSkewStr); err == nil {
			config.AllowedClockSkew = duration
		}
	}
	// 默认5分钟时钟偏差容忍
	if config.AllowedClockSkew == 0 {
		config.AllowedClockSkew = 5 * time.Minute
	}

	if config.KeyID == "" && strings.EqualFold(config.Algorithm, "RS256") {
		config.KeyID = "bff-key-1"
	}

	return config
}

// IsRS256 检查是否使用RS256算法
func (c *JWTConfig) IsRS256() bool {
	return c.Algorithm == "RS256"
}

// IsHS256 检查是否使用HS256算法
func (c *JWTConfig) IsHS256() bool {
	return c.Algorithm == "HS256"
}

// HasJWKS 检查是否配置了JWKS
func (c *JWTConfig) HasJWKS() bool {
	return c.JWKSUrl != ""
}

// HasPublicKey 检查是否配置了公钥文件
func (c *JWTConfig) HasPublicKey() bool {
	return c.PublicKeyPath != ""
}

// HasPrivateKey 检查是否配置了私钥文件
func (c *JWTConfig) HasPrivateKey() bool {
	return c.PrivateKeyPath != ""
}
