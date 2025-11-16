//go:build legacy

package main

import (
	"log"

	"cube-castle/cmd/hrms-server/query/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}
