package auth

import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "encoding/pem"
    "fmt"
    "io"
    "math/big"
    "net/http"
    "sync"
    "time"
)

// JWKS 结构
type jwkKey struct {
    Kty string `json:"kty"`
    Kid string `json:"kid"`
    N   string `json:"n"`
    E   string `json:"e"`
}
type jwkSet struct {
    Keys []jwkKey `json:"keys"`
}

// JWKSManager 负责缓存和刷新JWKS
type JWKSManager struct {
    url      string
    ttl      time.Duration
    mu       sync.RWMutex
    cache    map[string]*rsa.PublicKey // kid -> key
    lastFetch time.Time
}

func NewJWKSManager(url string, ttl time.Duration) *JWKSManager {
    return &JWKSManager{
        url:   url,
        ttl:   ttl,
        cache: make(map[string]*rsa.PublicKey),
    }
}

// GetKey 返回kid对应的公钥（如果缓存有效）
func (m *JWKSManager) GetKey(kid string) *rsa.PublicKey {
    m.mu.RLock()
    defer m.mu.RUnlock()
    if k, ok := m.cache[kid]; ok {
        return k
    }
    return nil
}

// Refresh 从远端拉取JWKS并更新缓存
func (m *JWKSManager) Refresh() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.url == "" {
        return fmt.Errorf("jwks url empty")
    }
    // 若未过期也允许强刷（简单实现）
    resp, err := http.Get(m.url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        b, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("jwks http %d: %s", resp.StatusCode, string(b))
    }
    var set jwkSet
    if err := json.NewDecoder(resp.Body).Decode(&set); err != nil {
        return err
    }
    // 解析RSA公钥
    newCache := make(map[string]*rsa.PublicKey)
    for _, k := range set.Keys {
        if k.Kty != "RSA" || k.Kid == "" || k.N == "" || k.E == "" {
            continue
        }
        if pk, err := RSAFromModExp(k.N, k.E); err == nil {
            newCache[k.Kid] = pk
        }
    }
    if len(newCache) > 0 {
        m.cache = newCache
        m.lastFetch = time.Now()
    }
    return nil
}

// ParseRSAPublicKeyFromPEM 从PEM解析RSA公钥
func ParseRSAPublicKeyFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
    block, _ := pem.Decode(pemBytes)
    if block == nil {
        return nil, fmt.Errorf("invalid pem")
    }
    pub, err := x509.ParsePKIXPublicKey(block.Bytes)
    if err != nil {
        return nil, err
    }
    rsaPub, ok := pub.(*rsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("not rsa public key")
    }
    return rsaPub, nil
}

// RSAFromModExp 将base64url编码的模数与指数转换为公钥
func RSAFromModExp(nB64URL, eB64URL string) (*rsa.PublicKey, error) {
    nBytes, err := base64.RawURLEncoding.DecodeString(nB64URL)
    if err != nil {
        return nil, err
    }
    eBytes, err := base64.RawURLEncoding.DecodeString(eB64URL)
    if err != nil {
        return nil, err
    }
    var eInt int
    for _, b := range eBytes {
        eInt = eInt<<8 | int(b)
    }
    key := &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: eInt}
    return key, nil
}

