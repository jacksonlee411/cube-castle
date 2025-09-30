package main

import (
	"log"

	"cube-castle-deployment-test/cmd/organization-query-service/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("❌ 服务启动失败: %v", err)
	}
}
