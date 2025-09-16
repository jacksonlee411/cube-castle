package graphql

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// SchemaLoader 统一GraphQL Schema加载器
// 消除双源维护漂移风险，确保Schema单一真源
type SchemaLoader struct {
	schemaPath string
	cache      string
}

// NewSchemaLoader 创建Schema加载器
func NewSchemaLoader(schemaPath string) *SchemaLoader {
	return &SchemaLoader{
		schemaPath: schemaPath,
	}
}

// LoadSchema 从docs/api/schema.graphql加载Schema
// 单一真源：以文档中的Schema为准
func (sl *SchemaLoader) LoadSchema() (string, error) {
	// 如果已有缓存，直接返回
	if sl.cache != "" {
		return sl.cache, nil
	}

	// 确保路径存在
	if _, err := os.Stat(sl.schemaPath); os.IsNotExist(err) {
		return "", fmt.Errorf("GraphQL schema file not found: %s", sl.schemaPath)
	}

	// 读取Schema文件
	content, err := ioutil.ReadFile(sl.schemaPath)
	if err != nil {
		return "", fmt.Errorf("failed to read GraphQL schema: %w", err)
	}

	// 缓存Schema内容
	sl.cache = string(content)

	return sl.cache, nil
}

// GetDefaultSchemaPath 获取默认Schema路径
func GetDefaultSchemaPath() string {
	// 相对于项目根目录的Schema文件路径
	return filepath.Join("docs", "api", "schema.graphql")
}

// MustLoadSchema 加载Schema，失败时panic
// 用于服务启动时的关键初始化
func MustLoadSchema(schemaPath string) string {
	loader := NewSchemaLoader(schemaPath)
	schema, err := loader.LoadSchema()
	if err != nil {
		panic(fmt.Sprintf("Failed to load GraphQL schema from %s: %v", schemaPath, err))
	}
	return schema
}

// ValidateSchemaConsistency 验证Schema一致性
// 用于CI/CD验证文档Schema与运行时Schema的一致性
func ValidateSchemaConsistency(docSchemaPath, runtimeSchema string) error {
	loader := NewSchemaLoader(docSchemaPath)
	docSchema, err := loader.LoadSchema()
	if err != nil {
		return fmt.Errorf("failed to load documentation schema: %w", err)
	}

	// 简单的字符串比较验证
	// TODO: 可以扩展为更复杂的AST比较
	if docSchema != runtimeSchema {
		return fmt.Errorf("schema inconsistency detected between documentation and runtime")
	}

	return nil
}
