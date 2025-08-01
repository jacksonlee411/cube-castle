package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gaogu/cube-castle/go-app/ent"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	PostgreSQLURL string
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
}

// Database 数据库连接管理器
type Database struct {
	PostgreSQL *pgxpool.Pool
	Neo4j      neo4j.DriverWithContext
}

// NewDatabaseConfig 从环境变量创建数据库配置
func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		PostgreSQLURL: os.Getenv("DATABASE_URL"),
		Neo4jURI:      os.Getenv("NEO4J_URI"),
		Neo4jUser:     os.Getenv("NEO4J_USER"),
		Neo4jPassword: os.Getenv("NEO4J_PASSWORD"),
	}
}

// Connect 连接数据库
func Connect(config *DatabaseConfig) (*Database, error) {
	// 连接 PostgreSQL
	pgPool, err := pgxpool.New(context.Background(), config.PostgreSQLURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// 测试 PostgreSQL 连接
	if err := pgPool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// 创建Database对象，Neo4j可以为nil
	db := &Database{
		PostgreSQL: pgPool,
		Neo4j:      nil, // 暂时不使用Neo4j
	}

	// 如果Neo4j配置可用，尝试连接
	if config.Neo4jURI != "" {
		neo4jDriver, err := neo4j.NewDriverWithContext(
			config.Neo4jURI,
			neo4j.BasicAuth(config.Neo4jUser, config.Neo4jPassword, ""),
		)
		if err == nil {
			// 测试 Neo4j 连接
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := neo4jDriver.VerifyConnectivity(ctx); err == nil {
				db.Neo4j = neo4jDriver
			}
		}
	}

	return db, nil
}

// Close 关闭数据库连接
func (db *Database) Close() {
	if db.PostgreSQL != nil {
		db.PostgreSQL.Close()
	}
	if db.Neo4j != nil {
		db.Neo4j.Close(context.Background())
	}
}

// HealthCheck 健康检查
func (db *Database) HealthCheck(ctx context.Context) error {
	// 检查 PostgreSQL
	if err := db.PostgreSQL.Ping(ctx); err != nil {
		return fmt.Errorf("PostgreSQL health check failed: %w", err)
	}

	// 检查 Neo4j
	if err := db.Neo4j.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("Neo4j health check failed: %w", err)
	}

	return nil
}

// InitDatabase 初始化数据库
func InitDatabase(db *Database) error {
	ctx := context.Background()

	// 读取并执行初始化脚本
	initScript, err := os.ReadFile("scripts/init-db.sql")
	if err != nil {
		log.Printf("Warning: Could not read init-db.sql: %v", err)
		return nil
	}

	// 执行 PostgreSQL 初始化脚本
	if _, err := db.PostgreSQL.Exec(ctx, string(initScript)); err != nil {
		return fmt.Errorf("failed to execute PostgreSQL init script: %w", err)
	}

	log.Println("✅ Database initialization completed")
	return nil
}

// SeedData 种子数据
func SeedData(db *Database) error {
	ctx := context.Background()

	// Neo4j 种子数据
	neo4jSession := db.Neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer neo4jSession.Close(ctx)

	_, err := neo4jSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 清理现有数据
		_, err := tx.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
		if err != nil {
			return nil, err
		}

		// 创建示例组织架构
		_, err = tx.Run(ctx, `
			CREATE (company:Organization {id: 'company-001', name: 'Cube Castle Inc.', type: 'company'}),
			       (hr:Organization {id: 'hr-dept', name: 'Human Resources', type: 'department'}),
			       (it:Organization {id: 'it-dept', name: 'Information Technology', type: 'department'}),
			       (hr)-[:BELONGS_TO]->(company),
			       (it)-[:BELONGS_TO]->(company)
		`, nil)
		return nil, err
	})

	if err != nil {
		return fmt.Errorf("failed to seed Neo4j data: %w", err)
	}

	log.Println("🌱 Neo4j database has been seeded")

	// PostgreSQL 种子数据
	seedQueries := []string{
		`INSERT INTO tenancy.tenants (id, name, domain, status) VALUES 
		 ('00000000-0000-0000-0000-000000000000', 'Default Tenant', 'default.cubecastle.com', 'active')
		 ON CONFLICT (id) DO NOTHING`,
		`INSERT INTO corehr.organizations (id, tenant_id, name, code, level) VALUES 
		 ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000000', 'Cube Castle Inc.', 'CC001', 1),
		 ('22222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000000', 'Human Resources', 'HR001', 2),
		 ('33333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000000', 'Information Technology', 'IT001', 2)
		 ON CONFLICT (id) DO NOTHING`,
	}

	for _, query := range seedQueries {
		if _, err := db.PostgreSQL.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute seed query: %w", err)
		}
	}

	log.Println("🌱 PostgreSQL database has been seeded")
	return nil
}

// InitDatabaseConnection 初始化数据库连接
func InitDatabaseConnection() *Database {
	config := NewDatabaseConfig()

	// 如果没有配置数据库URL，返回nil（使用Mock模式）
	if config.PostgreSQLURL == "" {
		return nil
	}

	db, err := Connect(config)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil
	}

	// 初始化数据库
	if err := InitDatabase(db); err != nil {
		log.Printf("Failed to initialize database: %v", err)
		db.Close()
		return nil
	}

	return db
}

// GetEntClient 获取Ent客户端
func GetEntClient() *ent.Client {
	// 使用与迁移工具相同的连接字符串
	connectionString := "postgresql://user:password@localhost:5432/cubecastle?sslmode=disable"

	// 从环境变量获取连接字符串（如果存在）
	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		connectionString = envDB
	}

	client, err := ent.Open("postgres", connectionString)
	if err != nil {
		log.Printf("Failed to open Ent client: %v", err)
		// 返回nil，调用方需要处理这种情况
		return nil
	}

	return client
}
