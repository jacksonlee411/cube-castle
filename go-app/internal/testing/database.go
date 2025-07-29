package testing

import (
	"context"
	"os"
	"testing"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/gaogu/cube-castle/go-app/ent/enttest"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// TestDatabaseType 测试数据库类型
type TestDatabaseType string

const (
	SQLiteMemory   TestDatabaseType = "sqlite_memory"
	SQLiteFile     TestDatabaseType = "sqlite_file"
	PostgreSQLTest TestDatabaseType = "postgresql_test"
)

// TestDatabaseConfig 测试数据库配置
type TestDatabaseConfig struct {
	Type           TestDatabaseType
	ConnectionString string
	CleanupOnClose bool
}

// GetTestDatabaseConfig 根据环境变量获取测试数据库配置
func GetTestDatabaseConfig() *TestDatabaseConfig {
	// 从环境变量读取测试数据库类型
	dbType := os.Getenv("TEST_DB_TYPE")
	
	switch dbType {
	case "postgres", "postgresql":
		return &TestDatabaseConfig{
			Type:           PostgreSQLTest,
			ConnectionString: GetPostgreSQLTestConnectionString(),
			CleanupOnClose:   true,
		}
	case "sqlite", "sqlite_file":
		return &TestDatabaseConfig{
			Type:           SQLiteFile,
			ConnectionString: "file:testdb.sqlite?cache=shared&_fk=1",
			CleanupOnClose:   true,
		}
	default:
		// 默认使用SQLite内存数据库（最快）
		return &TestDatabaseConfig{
			Type:           SQLiteMemory,
			ConnectionString: "file:ent?mode=memory&cache=shared&_fk=1",
			CleanupOnClose:   false,
		}
	}
}

// GetPostgreSQLTestConnectionString 获取PostgreSQL测试数据库连接字符串
func GetPostgreSQLTestConnectionString() string {
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL != "" {
		return testDBURL
	}
	
	// 默认测试数据库连接
	return "postgresql://postgres:password@localhost:5432/cubecastle_test?sslmode=disable"
}

// SetupTestDatabase 设置测试数据库
func SetupTestDatabase(t *testing.T) (*ent.Client, func()) {
	config := GetTestDatabaseConfig()
	
	var client *ent.Client
	var cleanup func()
	
	switch config.Type {
	case SQLiteMemory:
		client = enttest.Open(t, "sqlite3", config.ConnectionString)
		cleanup = func() { client.Close() }
		
	case SQLiteFile:
		client = enttest.Open(t, "sqlite3", config.ConnectionString)
		cleanup = func() { 
			client.Close()
			if config.CleanupOnClose {
				os.Remove("testdb.sqlite")
			}
		}
		
	case PostgreSQLTest:
		// 使用PostgreSQL测试数据库
		client = setupPostgreSQLTestDB(t, config.ConnectionString)
		cleanup = func() { 
			cleanupPostgreSQLTestDB(t, client)
			client.Close()
		}
		
	default:
		t.Fatalf("Unsupported test database type: %s", config.Type)
	}
	
	return client, cleanup
}

// setupPostgreSQLTestDB 设置PostgreSQL测试数据库
func setupPostgreSQLTestDB(t *testing.T, connectionString string) *ent.Client {
	client, err := ent.Open("postgres", connectionString)
	if err != nil {
		t.Skipf("无法连接到PostgreSQL测试数据库: %v", err)
	}
	
	// 运行迁移
	if err := client.Schema.Create(context.Background()); err != nil {
		client.Close()
		t.Fatalf("创建数据库schema失败: %v", err)
	}
	
	return client
}

// cleanupPostgreSQLTestDB 清理PostgreSQL测试数据库
func cleanupPostgreSQLTestDB(t *testing.T, client *ent.Client) {
	ctx := context.Background()
	
	// 清理所有表的数据（保留结构）
	tables := []string{
		"position_occupancy_history",
		"position_attribute_history", 
		"positions",
		"organization_units",
	}
	
	for _, table := range tables {
		// 使用Ent的实体删除方法而不是直接的SQL
		switch table {
		case "position_occupancy_history":
			if _, err := client.PositionOccupancyHistory.Delete().Exec(ctx); err != nil {
				t.Logf("清理 position_occupancy_history 失败: %v", err)
			}
		case "position_attribute_history":
			if _, err := client.PositionAttributeHistory.Delete().Exec(ctx); err != nil {
				t.Logf("清理 position_attribute_history 失败: %v", err)
			}
		case "positions":
			if _, err := client.Position.Delete().Exec(ctx); err != nil {
				t.Logf("清理 positions 失败: %v", err)
			}
		case "organization_units":
			if _, err := client.OrganizationUnit.Delete().Exec(ctx); err != nil {
				t.Logf("清理 organization_units 失败: %v", err)
			}
		}
	}
}

// SetupTestHandler 设置测试处理器（通用方法）
func SetupTestHandler(t *testing.T) (*ent.Client, *logging.StructuredLogger, func()) {
	client, cleanup := SetupTestDatabase(t)
	logger := logging.NewStructuredLogger()
	
	return client, logger, cleanup
}

// TestDatabaseInfo 测试数据库信息
type TestDatabaseInfo struct {
	Type             TestDatabaseType
	SupportsFK       bool
	SupportsTx       bool
	PerformanceLevel string
}

// GetTestDatabaseInfo 获取当前测试数据库信息
func GetTestDatabaseInfo() *TestDatabaseInfo {
	config := GetTestDatabaseConfig()
	
	switch config.Type {
	case SQLiteMemory:
		return &TestDatabaseInfo{
			Type:             SQLiteMemory,
			SupportsFK:       true,
			SupportsTx:       true,
			PerformanceLevel: "最快",
		}
	case SQLiteFile:
		return &TestDatabaseInfo{
			Type:             SQLiteFile,
			SupportsFK:       true,
			SupportsTx:       true,
			PerformanceLevel: "快",
		}
	case PostgreSQLTest:
		return &TestDatabaseInfo{
			Type:             PostgreSQLTest,
			SupportsFK:       true,
			SupportsTx:       true,
			PerformanceLevel: "中等（与生产环境一致）",
		}
	default:
		return nil
	}
}