package main

import (
	"context"
	"log"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	logger := log.New(os.Stdout, "[NEO4J-CLEANUP] ", log.LstdFlags)

	neo4jURI := os.Getenv("NEO4J_URI")
	if neo4jURI == "" {
		neo4jURI = "bolt://localhost:7687"
	}

	neo4jUser := os.Getenv("NEO4J_USER")
	if neo4jUser == "" {
		neo4jUser = "neo4j"
	}

	neo4jPassword := os.Getenv("NEO4J_PASSWORD")
	if neo4jPassword == "" {
		neo4jPassword = "password"
	}

	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		log.Fatalf("Neo4j驱动创建失败: %v", err)
	}
	defer driver.Close(context.Background())

	// 测试连接
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		log.Fatalf("Neo4j连接失败: %v", err)
	}
	logger.Println("Neo4j连接成功")

	session := driver.NewSession(context.Background(), neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(context.Background())

	ctx := context.Background()

	// 1. 检查当前数据
	logger.Println("检查当前数据...")
	countResult, err := session.Run(ctx, "MATCH (n) RETURN count(n) as count", map[string]interface{}{})
	if err != nil {
		log.Fatalf("查询数据失败: %v", err)
	}
	
	count := 0
	if countResult.Next(ctx) {
		record := countResult.Record()
		if c, ok := record.Get("count"); ok {
			if countInt, ok := c.(int64); ok {
				count = int(countInt)
			}
		}
	}
	logger.Printf("找到 %d 个节点", count)

	// 2. 检查Organization节点
	orgResult, err := session.Run(ctx, "MATCH (n:Organization) RETURN count(n) as count", map[string]interface{}{})
	if err != nil {
		log.Fatalf("查询Organization节点失败: %v", err)
	}
	
	orgCount := 0
	if orgResult.Next(ctx) {
		record := orgResult.Record()
		if c, ok := record.Get("count"); ok {
			if countInt, ok := c.(int64); ok {
				orgCount = int(countInt)
			}
		}
	}
	logger.Printf("找到 %d 个Organization节点", orgCount)

	if count == 0 {
		logger.Println("数据库已为空")
		return
	}

	// 3. 清空所有数据
	logger.Println("开始清空所有数据...")
	
	result, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, "MATCH (n) DETACH DELETE n RETURN count(n) as deleted_count", map[string]interface{}{})
		if err != nil {
			return nil, err
		}
		
		if result.Next(ctx) {
			record := result.Record()
			if c, ok := record.Get("deleted_count"); ok {
				return c, nil
			}
		}
		return 0, nil
	})
	
	if err != nil {
		log.Fatalf("清空数据失败: %v", err)
	}
	
	deletedCount := 0
	if c, ok := result.(int64); ok {
		deletedCount = int(c)
	}
	
	logger.Printf("✅ 成功删除 %d 个节点", deletedCount)

	// 4. 验证清空结果
	logger.Println("验证清空结果...")
	finalResult, err := session.Run(ctx, "MATCH (n) RETURN count(n) as count", map[string]interface{}{})
	if err != nil {
		log.Fatalf("验证清空结果失败: %v", err)
	}
	
	finalCount := 0
	if finalResult.Next(ctx) {
		record := finalResult.Record()
		if c, ok := record.Get("count"); ok {
			if countInt, ok := c.(int64); ok {
				finalCount = int(countInt)
			}
		}
	}
	
	if finalCount == 0 {
		logger.Println("✅ Neo4j数据库已完全清空")
	} else {
		logger.Printf("⚠️ 仍有 %d 个节点未删除", finalCount)
	}

	logger.Println("数据清空完成")
}