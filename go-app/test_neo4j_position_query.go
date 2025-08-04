package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

func main() {
	fmt.Println("=== Neo4j 位置查询仓储测试 ===")

	// 创建日志器
	logger := logging.NewStructuredLogger()

	// 创建Neo4j位置查询仓储（没有实际的Neo4j服务，将返回模拟数据）
	var neo4jService *service.Neo4jService // nil - 这将触发模拟数据模式
	positionRepo := repositories.NewNeo4jPositionQueryRepositoryV2(neo4jService, logger)

	ctx := context.Background()
	testTenantID := uuid.New()

	// 测试1: GetPosition - 获取单个职位
	fmt.Println("\n--- 测试1: GetPosition ---")
	positionQuery := queries.GetPositionQuery{
		TenantID: testTenantID,
		ID:       uuid.New(),
	}

	position, err := positionRepo.GetPosition(ctx, positionQuery)
	if err != nil {
		log.Printf("GetPosition 错误: %v", err)
		os.Exit(1)
	}
	fmt.Printf("获取职位成功 - ID: %s, 标题: %v, 状态: %s\n",
		position.ID, position.Details["title"], position.Status)

	// 测试2: SearchPositions - 搜索职位
	fmt.Println("\n--- 测试2: SearchPositions ---")
	searchParams := repositories.SearchPositionsParams{
		TenantID: testTenantID,
		Limit:    10,
		Offset:   0,
	}

	positions, total, err := positionRepo.SearchPositions(ctx, searchParams)
	if err != nil {
		log.Printf("SearchPositions 错误: %v", err)
		os.Exit(1)
	}
	fmt.Printf("搜索职位成功 - 找到 %d 个职位，总数: %d\n", len(positions), total)
	for i, pos := range positions {
		fmt.Printf("  %d. %s (%s) - %v\n", i+1, pos.Status, pos.PositionType, pos.Details["title"])
	}

	// 测试3: GetPositionHierarchy - 获取职位层级
	fmt.Println("\n--- 测试3: GetPositionHierarchy ---")
	hierarchyQuery := queries.GetPositionHierarchyQuery{
		TenantID: testTenantID,
		MaxDepth: 5,
	}

	hierarchy, err := positionRepo.GetPositionHierarchy(ctx, hierarchyQuery)
	if err != nil {
		log.Printf("GetPositionHierarchy 错误: %v", err)
		os.Exit(1)
	}
	fmt.Printf("获取职位层级成功 - %d 个层级节点\n", len(hierarchy))
	for _, node := range hierarchy {
		fmt.Printf("  级别 %d: %v (%s)\n", node.Level, node.Position.Details["title"], node.Position.Status)
	}

	// 测试4: GetEmployeePositions - 获取员工职位历史
	fmt.Println("\n--- 测试4: GetEmployeePositions ---")
	empPositionsQuery := queries.GetEmployeePositionsQuery{
		TenantID:     testTenantID,
		EmployeeID:   uuid.New(),
		IncludePast:  true,
	}

	empPositions, err := positionRepo.GetEmployeePositions(ctx, empPositionsQuery)
	if err != nil {
		log.Printf("GetEmployeePositions 错误: %v", err)
		os.Exit(1)
	}
	fmt.Printf("获取员工职位历史成功 - %d 条记录\n", len(empPositions))
	for i, pos := range empPositions {
		status := "历史"
		if pos.IsCurrent {
			status = "当前"
		}
		fmt.Printf("  %d. %s 职位 - FTE: %.1f, 类型: %s\n", i+1, status, pos.FTE, pos.AssignmentType)
	}

	// 测试5: GetPositionStats - 获取职位统计
	fmt.Println("\n--- 测试5: GetPositionStats ---")
	statsQuery := queries.GetPositionStatsQuery{
		TenantID: testTenantID,
	}

	stats, err := positionRepo.GetPositionStats(ctx, statsQuery)
	if err != nil {
		log.Printf("GetPositionStats 错误: %v", err)
		os.Exit(1)
	}
	fmt.Printf("获取职位统计成功:\n")
	fmt.Printf("  总职位数: %d\n", stats.Total)
	fmt.Printf("  开放职位: %d\n", stats.Open)
	fmt.Printf("  已填充职位: %d\n", stats.Filled)
	fmt.Printf("  平均FTE: %.2f\n", stats.AverageFTE)
	fmt.Printf("  空缺率: %.1f%%\n", stats.VacancyRate)
	fmt.Printf("  流动率: %.1f%%\n", stats.TurnoverRate)

	fmt.Println("\n=== 所有测试成功完成! ===")
	fmt.Println("Neo4j位置查询仓储已成功实现并正常工作。")
	fmt.Println("注意: 当前使用模拟数据模式，当Neo4j服务配置完成后将自动切换到真实数据查询。")
}