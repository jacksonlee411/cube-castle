package main

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	config "cube-castle-deployment-test/internal/config"
	"github.com/golang-jwt/jwt/v5"
	sharedconfig "shared/config"
)

func main() {
	keyPathFlag := flag.String("key", "", "Path to RS256 private key")
	subjectFlag := flag.String("sub", "dev-user-001", "Subject claim value")
	issuerFlag := flag.String("issuer", "cube-castle", "JWT issuer")
	audienceFlag := flag.String("audience", "cube-castle-api", "JWT audience")
	rolesFlag := flag.String("roles", "ADMIN,MANAGER", "Comma separated role list")
	durationFlag := flag.Duration("ttl", 24*time.Hour, "Token validity duration")
	outputFlag := flag.String("out", ".cache/dev.jwt", "Path to write token (empty to skip)")
	jwksFlag := flag.String("jwks", ".well-known/jwks.json", "Path to write JWKS JSON (empty to skip)")
	flag.Parse()

	provided := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		provided[f.Name] = true
	})

	cfg := config.GetJWTConfig()
	if !cfg.IsRS256() {
		log.Fatalf("当前 JWT_ALG=%s，开发令牌工具仅支持 RS256，请纠正环境配置。", cfg.Algorithm)
	}

	issuer := strings.TrimSpace(*issuerFlag)
	if !provided["issuer"] && strings.TrimSpace(cfg.Issuer) != "" {
		issuer = cfg.Issuer
	}
	if issuer == "" {
		log.Fatalf("JWT issuer 不能为空")
	}

	audience := strings.TrimSpace(*audienceFlag)
	if !provided["audience"] && strings.TrimSpace(cfg.Audience) != "" {
		audience = cfg.Audience
	}
	if audience == "" {
		log.Fatalf("JWT audience 不能为空")
	}

	keyPath, err := resolveKeyPath(*keyPathFlag, cfg.PrivateKeyPath)
	if err != nil {
		log.Fatalf("解析私钥路径失败: %v", err)
	}
	privateKey, err := loadPrivateKey(keyPath)
	if err != nil {
		log.Fatalf("加载RS256私钥失败(%s): %v", keyPath, err)
	}

	tokenPath, err := resolveWorkspacePath(*outputFlag, ".cache/dev.jwt", true)
	if err != nil {
		log.Fatalf("解析令牌输出路径失败: %v", err)
	}

	jwksPath, err := resolveWorkspacePath(*jwksFlag, ".well-known/jwks.json", true)
	if err != nil {
		log.Fatalf("解析JWKS输出路径失败: %v", err)
	}

	roles := parseRoles(*rolesFlag)
	expiresAt := time.Now().Add(*durationFlag)

	claims := jwt.MapClaims{
		"sub":       *subjectFlag,
		"tenant_id": sharedconfig.GetDefaultTenantIDString(),
		"roles":     roles,
		"iss":       issuer,
		"aud":       audience,
		"exp":       expiresAt.Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if cfg.KeyID != "" {
		token.Header["kid"] = cfg.KeyID
	}

	signed, err := token.SignedString(privateKey)
	if err != nil {
		log.Fatalf("签名令牌失败: %v", err)
	}

	if err := ensureTokenAlgorithm(signed, jwt.SigningMethodRS256.Alg()); err != nil {
		log.Fatalf("生成的令牌签名算法校验失败: %v", err)
	}

	fmt.Printf("Valid JWT Token (alg=RS256, aud=%s)\n%s\n", audience, signed)

	if err := maybeWriteFile(tokenPath, []byte(signed)); err != nil {
		log.Fatalf("写入开发令牌失败: %v", err)
	}

	if err := maybeWriteJWKS(jwksPath, &privateKey.PublicKey, resolveKeyID(cfg.KeyID)); err != nil {
		log.Fatalf("写入JWKS失败: %v", err)
	}
}

func resolveKeyPath(flagValue, configValue string) (string, error) {
	if strings.TrimSpace(flagValue) != "" {
		return resolveWorkspacePath(flagValue, "", false)
	}
	if strings.TrimSpace(configValue) != "" {
		return resolveWorkspacePath(configValue, "", false)
	}
	return resolveWorkspacePath("secrets/dev-jwt-private.pem", "", false)
}

func parseRoles(raw string) []string {
	if raw == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	roles := make([]string, 0, len(parts))
	for _, part := range parts {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		roles = append(roles, token)
	}
	return roles
}

func loadPrivateKey(keyPath string) (*rsa.PrivateKey, error) {
	// #nosec G304 -- keyPath 已通过 resolveKeyPath 校验并限定在工作目录内
	pemBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("读取私钥失败: %w", err)
	}

	key, parseErr := jwt.ParseRSAPrivateKeyFromPEM(pemBytes)
	if parseErr != nil {
		return nil, fmt.Errorf("解析RS256私钥失败: %w", parseErr)
	}
	return key, nil
}

func maybeWriteFile(path string, data []byte) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}
	return nil
}

func maybeWriteJWKS(path string, pub *rsa.PublicKey, keyID string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	jwks, err := marshalJWKS(pub, keyID)
	if err != nil {
		return err
	}
	return maybeWriteFile(path, jwks)
}

func marshalJWKS(pub *rsa.PublicKey, keyID string) ([]byte, error) {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	set := map[string]interface{}{
		"keys": []map[string]string{{
			"kty": "RSA",
			"kid": keyID,
			"n":   n,
			"e":   e,
		}},
	}

	return json.MarshalIndent(set, "", "  ")
}

func resolveKeyID(candidate string) string {
	if strings.TrimSpace(candidate) != "" {
		return candidate
	}
	return "bff-key-1"
}

func resolveWorkspacePath(raw string, fallback string, allowEmpty bool) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		value = strings.TrimSpace(fallback)
		if value == "" {
			if allowEmpty {
				return "", nil
			}
			return "", fmt.Errorf("路径不能为空")
		}
	}

	cleaned := filepath.Clean(value)
	if cleaned == "" {
		return "", fmt.Errorf("路径不能为空")
	}

	var err error
	if !filepath.IsAbs(cleaned) {
		cleaned, err = filepath.Abs(cleaned)
		if err != nil {
			return "", fmt.Errorf("解析绝对路径失败: %w", err)
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("获取工作目录失败: %w", err)
	}
	wd = filepath.Clean(wd)
	if cleaned != wd && !strings.HasPrefix(cleaned, wd+string(os.PathSeparator)) {
		return "", fmt.Errorf("拒绝访问工作目录外的路径: %s", cleaned)
	}

	return cleaned, nil
}

func ensureTokenAlgorithm(tokenString, expectedAlg string) error {
	segments := strings.Split(tokenString, ".")
	if len(segments) != 3 {
		return fmt.Errorf("JWT 格式错误: 期望 3 个段，实际 %d", len(segments))
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(segments[0])
	if err != nil {
		return fmt.Errorf("解析 JWT header 失败: %w", err)
	}

	var header map[string]interface{}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return fmt.Errorf("解析 JWT header JSON 失败: %w", err)
	}

	rawAlg, ok := header["alg"].(string)
	if !ok {
		return fmt.Errorf("JWT header 缺少 alg 字段")
	}
	if !strings.EqualFold(rawAlg, expectedAlg) {
		return fmt.Errorf("签名算法不匹配: 期望 %s，实际 %s", expectedAlg, rawAlg)
	}

	return nil
}
