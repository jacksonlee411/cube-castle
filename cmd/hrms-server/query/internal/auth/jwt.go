package auth

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

type JWTMiddleware struct {
    secretKey []byte
    issuer    string
    audience  string
    alg       string
    jwks      *JWKSManager
    publicKey interface{}
    clockSkew time.Duration
}

func NewJWTMiddleware(secretKey, issuer, audience string) *JWTMiddleware {
    return &JWTMiddleware{secretKey: []byte(secretKey), issuer: issuer, audience: audience, alg: "HS256"}
}

type Options struct {
    Alg          string
    JWKSURL      string
    PublicKeyPEM []byte
    ClockSkew    time.Duration
}

func NewJWTMiddlewareWithOptions(secretKey, issuer, audience string, opt Options) *JWTMiddleware {
    mw := &JWTMiddleware{secretKey: []byte(secretKey), issuer: issuer, audience: audience, alg: strings.ToUpper(strings.TrimSpace(opt.Alg)), clockSkew: opt.ClockSkew}
    if mw.alg == "RS256" {
        if opt.JWKSURL != "" {
            mw.jwks = NewJWKSManager(opt.JWKSURL, 5*time.Minute)
        } else if len(opt.PublicKeyPEM) > 0 {
            if pk, err := ParseRSAPublicKeyFromPEM(opt.PublicKeyPEM); err == nil {
                mw.publicKey = pk
            }
        }
    }
    if mw.alg == "" { mw.alg = "HS256" }
    return mw
}

// Claims JWT声明结构
type Claims struct {
    UserID    string   `json:"sub"`
    TenantID  string   `json:"tenant_id"`
    Roles     []string `json:"roles"`
    ExpiresAt int64    `json:"exp"`
    // PBAC scopes
    Scope       string   `json:"scope"`
    Permissions []string `json:"permissions"`
}

// ValidateToken 验证JWT令牌
func (j *JWTMiddleware) ValidateToken(tokenString string) (*Claims, error) {
    // 移除Bearer前缀
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    keyFunc := func(token *jwt.Token) (interface{}, error) {
        alg := token.Header["alg"]
        if j.alg == "HS256" {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { return nil, fmt.Errorf("invalid signing method: %v", alg) }
            return j.secretKey, nil
        }
        if j.alg == "RS256" {
            if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok { return nil, fmt.Errorf("invalid signing method: %v", alg) }
            if j.jwks != nil {
                if kid, _ := token.Header["kid"].(string); kid != "" {
                    if key := j.jwks.GetKey(kid); key != nil { return key, nil }
                    if err := j.jwks.Refresh(); err == nil { if key := j.jwks.GetKey(kid); key != nil { return key, nil } }
                    return nil, fmt.Errorf("unknown kid: %s", kid)
                }
            }
            if j.publicKey != nil { return j.publicKey, nil }
            return nil, fmt.Errorf("no public key available for RS256")
        }
        return nil, fmt.Errorf("unsupported alg: %s", j.alg)
    }

    token, err := jwt.Parse(tokenString, keyFunc)
    if err != nil { return nil, fmt.Errorf("token parsing failed: %w", err) }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        // 验证issuer和audience
        if iss, ok := claims["iss"].(string); !ok || iss != j.issuer { return nil, fmt.Errorf("invalid issuer") }
        if aud, ok := claims["aud"].(string); !ok || aud != j.audience { return nil, fmt.Errorf("invalid audience") }

        userClaims := &Claims{
            UserID:   getClaimString(claims, "sub"),
            TenantID: getTenantIDClaim(claims),
        }

        if rolesClaim, ok := claims["roles"].([]interface{}); ok {
            for _, role := range rolesClaim {
                if roleStr, ok := role.(string); ok { userClaims.Roles = append(userClaims.Roles, roleStr) }
            }
        }
        if scopeStr, ok := claims["scope"].(string); ok { userClaims.Scope = scopeStr }
        if perms, ok := claims["permissions"].([]interface{}); ok {
            for _, p := range perms { if ps, ok := p.(string); ok { userClaims.Permissions = append(userClaims.Permissions, ps) } }
        }

        now := time.Now()
        if exp, ok := claims["exp"].(float64); ok {
            userClaims.ExpiresAt = int64(exp)
            if now.After(time.Unix(userClaims.ExpiresAt, 0).Add(j.clockSkew)) { return nil, fmt.Errorf("token expired") }
        }
        if nbf, ok := claims["nbf"].(float64); ok {
            if now.Add(j.clockSkew).Before(time.Unix(int64(nbf), 0)) { return nil, fmt.Errorf("token not valid yet") }
        }

        return userClaims, nil
    }
    return nil, fmt.Errorf("invalid token claims")
}

// getClaimString 安全地获取字符串类型的claim
func getClaimString(claims jwt.MapClaims, key string) string {
    if value, ok := claims[key].(string); ok { return value }
    return ""
}

// getTenantIDClaim 兼容 camelCase 与 snake_case
func getTenantIDClaim(claims jwt.MapClaims) string {
    if v, ok := claims["tenantId"].(string); ok && v != "" { return v }
    if v, ok := claims["tenant_id"].(string); ok && v != "" { return v }
    return ""
}

// SetUserContext 将用户信息设置到上下文中
func SetUserContext(ctx context.Context, claims *Claims) context.Context {
    ctx = context.WithValue(ctx, "user_id", claims.UserID)
    ctx = context.WithValue(ctx, "tenant_id", claims.TenantID)
    ctx = context.WithValue(ctx, "user_roles", claims.Roles)
    scopeSet := map[string]struct{}{}
    for _, s := range strings.Fields(claims.Scope) { scopeSet[s] = struct{}{} }
    for _, s := range claims.Permissions { scopeSet[s] = struct{}{} }
    var scopes []string
    for s := range scopeSet { scopes = append(scopes, s) }
    ctx = context.WithValue(ctx, "user_scopes", scopes)
    return ctx
}

func GetUserID(ctx context.Context) string { if v, ok := ctx.Value("user_id").(string); ok { return v }; return "" }
func GetTenantID(ctx context.Context) string { if v, ok := ctx.Value("tenant_id").(string); ok { return v }; return "" }
func GetUserRoles(ctx context.Context) []string { if v, ok := ctx.Value("user_roles").([]string); ok { return v }; return []string{} }
func GetUserScopes(ctx context.Context) []string { if v, ok := ctx.Value("user_scopes").([]string); ok { return v }; return []string{} }

