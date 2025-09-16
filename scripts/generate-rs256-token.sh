#!/bin/bash
# 生成RS256签名的JWT令牌（与JWKS验签兼容）

set -euo pipefail

# 确保密钥文件存在
if [[ ! -f secrets/dev-jwt-private.pem ]]; then
    echo "Error: RS256 private key not found at secrets/dev-jwt-private.pem"
    echo "Run 'make jwt-dev-setup' first"
    exit 1
fi

# 使用Go脚本生成RS256令牌
cat > /tmp/gen-rs256-jwt.go << 'EOF'
package main

import (
    "crypto/rsa"
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

func main() {
    // 读取私钥
    keyData, err := ioutil.ReadFile("secrets/dev-jwt-private.pem")
    if err != nil {
        log.Fatalf("Failed to read private key: %v", err)
    }

    block, _ := pem.Decode(keyData)
    if block == nil {
        log.Fatal("Failed to decode PEM block")
    }

    privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
        // 尝试PKCS8格式
        key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
        if err2 != nil {
            log.Fatalf("Failed to parse private key: %v", err)
        }
        var ok bool
        privateKey, ok = key.(*rsa.PrivateKey)
        if !ok {
            log.Fatal("Not an RSA private key")
        }
    }

    // 创建claims
    claims := jwt.MapClaims{
        "sub":       "dev-user",
        "userId":    "dev-user",
        "tenantId":  "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
        "roles":     []string{"ADMIN", "USER"},
        "scopes":    []string{"org:read", "org:write", "org:delete"},
        "iss":       "cube-castle-dev",
        "aud":       []string{"organization-query-service", "organization-command-service"},
        "exp":       time.Now().Add(8 * time.Hour).Unix(),
        "iat":       time.Now().Unix(),
        "nbf":       time.Now().Unix(),
    }

    // 创建token
    token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
    token.Header["kid"] = "bff-key-1"

    // 签名
    tokenString, err := token.SignedString(privateKey)
    if err != nil {
        log.Fatalf("Failed to sign token: %v", err)
    }

    fmt.Print(tokenString)
}
EOF

# 编译并运行
cd /home/shangmeilin/cube-castle
go run /tmp/gen-rs256-jwt.go > .cache/dev.jwt

echo "✅ Generated RS256 JWT token"
echo "Token saved to: .cache/dev.jwt"
echo "First 50 chars: $(head -c 50 .cache/dev.jwt)..."