package authbff

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
)

type jwk struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

func rsaPublicJWK(key *rsa.PublicKey, kid string) jwk {
	n := base64.RawURLEncoding.EncodeToString(key.N.Bytes())
	// Exponent to bytes
	eBytes := []byte{byte(key.E >> 16), byte(key.E >> 8), byte(key.E)}
	// Trim leading zeros
	i := 0
	for i < len(eBytes) && eBytes[i] == 0 {
		i++
	}
	e := base64.RawURLEncoding.EncodeToString(eBytes[i:])
	return jwk{Kty: "RSA", Kid: kid, Alg: "RS256", Use: "sig", N: n, E: e}
}

func (h *BFFHandler) buildJWKS() ([]byte, error) {
	if h.jwtCfg.Alg != "RS256" || h.jwtCfg.PrivateKey == nil {
		return json.Marshal(jwks{Keys: []jwk{}})
	}
	pub := &h.jwtCfg.PrivateKey.PublicKey
	set := jwks{Keys: []jwk{rsaPublicJWK(pub, h.jwtCfg.KeyID)}}
	return json.Marshal(set)
}
