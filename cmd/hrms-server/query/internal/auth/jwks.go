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

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}
type jwkSet struct {
	Keys []jwkKey `json:"keys"`
}

type JWKSManager struct {
	url       string
	ttl       time.Duration
	mu        sync.RWMutex
	cache     map[string]*rsa.PublicKey
	lastFetch time.Time
}

func NewJWKSManager(url string, ttl time.Duration) *JWKSManager {
	return &JWKSManager{url: url, ttl: ttl, cache: make(map[string]*rsa.PublicKey)}
}

func (m *JWKSManager) GetKey(kid string) *rsa.PublicKey {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cache[kid]
}

func (m *JWKSManager) Refresh() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.url == "" {
		return fmt.Errorf("jwks url empty")
	}
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
	nc := make(map[string]*rsa.PublicKey)
	for _, k := range set.Keys {
		if k.Kty != "RSA" || k.Kid == "" || k.N == "" || k.E == "" {
			continue
		}
		if pk, err := RSAFromModExp(k.N, k.E); err == nil {
			nc[k.Kid] = pk
		}
	}
	if len(nc) > 0 {
		m.cache = nc
		m.lastFetch = time.Now()
	}
	return nil
}

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
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: eInt}, nil
}
