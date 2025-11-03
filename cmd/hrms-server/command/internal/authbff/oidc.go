package authbff

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	auth "cube-castle/internal/auth"
	jwt "github.com/golang-jwt/jwt/v5"
)

// OIDCConfig 基础配置（从环境变量读取）
type OIDCConfig struct {
	Issuer        string
	ClientID      string
	RedirectURI   string
	PostLogoutURI string
	Scopes        []string
}

// DiscoveryDoc OIDC 发现文档关键字段
type DiscoveryDoc struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
	JWKSURI               string `json:"jwks_uri"`
}

// OIDCClient 简化OIDC客户端：缓存发现文档、校验ID Token（RS256+JWKS）
type OIDCClient struct {
	cfg       OIDCConfig
	httpCli   *http.Client
	mu        sync.RWMutex
	cachedDoc *DiscoveryDoc
	cachedAt  time.Time
	cacheTTL  time.Duration
	jwks      *auth.JWKSManager
}

func NewOIDCClientFromEnv() *OIDCClient {
	scopes := os.Getenv("OIDC_SCOPES")
	if strings.TrimSpace(scopes) == "" {
		scopes = "openid profile email"
	}
	return &OIDCClient{
		cfg: OIDCConfig{
			Issuer:        strings.TrimSpace(os.Getenv("OIDC_ISSUER")),
			ClientID:      os.Getenv("OIDC_CLIENT_ID"),
			RedirectURI:   os.Getenv("OIDC_REDIRECT_URI"),
			PostLogoutURI: os.Getenv("OIDC_POST_LOGOUT_REDIRECT_URI"),
			Scopes:        strings.Fields(scopes),
		},
		httpCli:  &http.Client{Timeout: 10 * time.Second},
		cacheTTL: 10 * time.Minute,
	}
}

func (c *OIDCClient) IsConfigured() bool {
	return c != nil && c.cfg.Issuer != "" && c.cfg.ClientID != "" && c.cfg.RedirectURI != ""
}

// Discover 拉取并缓存发现文档
func (c *OIDCClient) Discover() (*DiscoveryDoc, error) {
	c.mu.RLock()
	if c.cachedDoc != nil && time.Since(c.cachedAt) < c.cacheTTL {
		defer c.mu.RUnlock()
		return c.cachedDoc, nil
	}
	c.mu.RUnlock()

	wellKnown := strings.TrimRight(c.cfg.Issuer, "/") + "/.well-known/openid-configuration"
	resp, err := c.httpCli.Get(wellKnown)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("oidc discovery http %d: %s", resp.StatusCode, string(b))
	}
	var doc DiscoveryDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}
	if doc.JWKSURI != "" {
		c.jwks = auth.NewJWKSManager(doc.JWKSURI, 5*time.Minute)
		_ = c.jwks.Refresh() // 预热
	}
	c.mu.Lock()
	c.cachedDoc = &doc
	c.cachedAt = time.Now()
	c.mu.Unlock()
	return &doc, nil
}

// BuildAuthURL 构建授权码+PKCE跳转地址
func (c *OIDCClient) BuildAuthURL(authzEndpoint, state, nonce, codeChallenge string, redirect string) (string, error) {
	u, err := url.Parse(authzEndpoint)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", c.cfg.ClientID)
	q.Set("redirect_uri", c.cfg.RedirectURI)
	scopes := strings.Join(c.cfg.Scopes, " ")
	q.Set("scope", scopes)
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	// 前端回跳路径透传（非标准参数，部分IdP允许透传，或放state中，简化起见保留在flow中）
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ExchangeCode 换票：使用授权码+PKCE换取token集
func (c *OIDCClient) ExchangeCode(tokenEndpoint, code, codeVerifier string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("client_id", c.cfg.ClientID)
	form.Set("redirect_uri", c.cfg.RedirectURI)
	form.Set("code_verifier", codeVerifier)
	req, _ := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token http %d: %s", resp.StatusCode, string(b))
	}
	var tr TokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

// ValidateIDToken 通过JWKS校验ID Token并返回claims（最小字段）
func (c *OIDCClient) ValidateIDToken(idToken string, expectedNonce string) (map[string]any, error) {
	if c.jwks == nil {
		return nil, fmt.Errorf("OIDC_JWKS_NOT_CONFIGURED")
	}
	// 解析并验证签名（RS256，使用kid从JWKS取key）
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, fmt.Errorf("missing kid")
		}
		if k := c.jwks.GetKey(kid); k != nil {
			return k, nil
		}
		if err := c.jwks.Refresh(); err == nil {
			if k := c.jwks.GetKey(kid); k != nil {
				return k, nil
			}
		}
		return nil, fmt.Errorf("unknown kid: %s", kid)
	}
	tok, err := jwt.Parse(idToken, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("ID_TOKEN_INVALID: %w", err)
	}
	mc, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return nil, fmt.Errorf("ID_TOKEN_INVALID_CLAIMS")
	}

	// iss 校验
	if iss, _ := mc["iss"].(string); iss == "" || (c.cfg.Issuer != "" && iss != c.cfg.Issuer) {
		return nil, fmt.Errorf("ID_TOKEN_ISSUER_MISMATCH")
	}
	// aud 校验（字符串或数组）
	audOK := false
	if audStr, ok := mc["aud"].(string); ok {
		audOK = (audStr == c.cfg.ClientID)
	} else if audArr, ok := mc["aud"].([]interface{}); ok {
		for _, v := range audArr {
			if s, ok := v.(string); ok && s == c.cfg.ClientID {
				audOK = true
				break
			}
		}
	}
	if !audOK {
		return nil, fmt.Errorf("ID_TOKEN_AUDIENCE_MISMATCH")
	}
	// exp/nbf 时间校验（含5分钟容忍）
	now := time.Now()
	if exp, ok := mc["exp"].(float64); ok {
		if now.After(time.Unix(int64(exp), 0).Add(5 * time.Minute)) {
			return nil, fmt.Errorf("ID_TOKEN_EXPIRED")
		}
	}
	if nbf, ok := mc["nbf"].(float64); ok {
		if now.Add(5 * time.Minute).Before(time.Unix(int64(nbf), 0)) {
			return nil, fmt.Errorf("ID_TOKEN_NOT_YET_VALID")
		}
	}
	// nonce 校验（如存在）
	if expectedNonce != "" {
		if nonce, _ := mc["nonce"].(string); nonce != "" && nonce != expectedNonce {
			return nil, fmt.Errorf("ID_TOKEN_NONCE_MISMATCH")
		}
	}
	// 返回 claims map
	claims := map[string]any{}
	for k, v := range mc {
		claims[k] = v
	}
	return claims, nil
}

// BuildCodeChallenge 计算S256挑战
func BuildCodeChallenge(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// TokenResponse OIDC Token 响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
}

// parseUnverifiedClaims 在无法使用JWKS时，解析未验证签名的claims（仅用于开发/模拟）
func parseUnverifiedClaims(idToken string) (map[string]any, error) {
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid id_token format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	return claims, nil
}

// RefreshWithToken 使用服务端保存的 refresh token 调用IdP刷新
func (c *OIDCClient) RefreshWithToken(tokenEndpoint, refreshToken string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	form.Set("client_id", c.cfg.ClientID)
	req, _ := http.NewRequest("POST", tokenEndpoint, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpCli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh http %d: %s", resp.StatusCode, string(b))
	}
	var tr TokenResponse
	if err := json.Unmarshal(b, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}
