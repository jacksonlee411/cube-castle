package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	// 使用与后端相同的密钥和设置
	secret := "cube-castle-development-secret-key-2025"
	issuer := "cube-castle"
	audience := "cube-castle-api"

	// 创建claims
	claims := jwt.MapClaims{
		"sub":       "dev-user-001",
		"tenant_id": "550e8400-e29b-41d4-a716-446655440000",
		"roles":     []string{"ADMIN", "MANAGER"},
		"iss":       issuer,
		"aud":       audience,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
		"iat":       time.Now().Unix(),
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}

	fmt.Printf("Valid JWT Token:\n%s\n", tokenString)
}