// internal/metacontract/compiler.go
package metacontract

import (
	"fmt"
	
	"github.com/gaogu/cube-castle/go-app/internal/codegen"
)

// Compiler implements the meta-contract compilation functionality
type Compiler struct {
	parser       *Parser
	validator    *Validator
	entGenerator *codegen.EntGenerator
	apiGenerator *codegen.APIGenerator
}

// NewCompiler creates a new meta-contract compiler
func NewCompiler() *Compiler {
	return &Compiler{
		parser:       NewParser(),
		validator:    NewValidator(),
		entGenerator: codegen.NewEntGenerator(),
		apiGenerator: codegen.NewAPIGenerator(),
	}
}

// Compile processes a meta-contract and generates all required artifacts
func (c *Compiler) Compile(inputPath, outputPath string) error {
	// 1. Parse meta-contract YAML
	contract, err := c.parser.ParseMetaContract(inputPath)
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}
	
	// 2. Validate meta-contract completeness
	if err := c.validator.Validate(contract); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// 3. Generate Ent Schema
	if err := c.entGenerator.Generate(contract, outputPath+"/schema"); err != nil {
		return fmt.Errorf("ent generation failed: %w", err)
	}
	
	// 4. Generate API routes
	if err := c.apiGenerator.Generate(contract, outputPath+"/api"); err != nil {
		return fmt.Errorf("api generation failed: %w", err)
	}
	
	return nil
}

// ParseMetaContract parses a YAML meta-contract file
func (c *Compiler) ParseMetaContract(yamlPath string) (*MetaContract, error) {
	return c.parser.ParseMetaContract(yamlPath)
}

// GenerateEntSchemas generates Ent schema files
func (c *Compiler) GenerateEntSchemas(contract *MetaContract, outputDir string) error {
	return c.entGenerator.Generate(contract, outputDir)
}

// GenerateBusinessLogic generates business logic skeleton
func (c *Compiler) GenerateBusinessLogic(contract *MetaContract, outputDir string) error {
	// Implementation for business logic generation
	return nil
}

// GenerateAPIRoutes generates API route definitions
func (c *Compiler) GenerateAPIRoutes(contract *MetaContract, outputDir string) error {
	return c.apiGenerator.Generate(contract, outputDir)
}