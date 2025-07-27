// internal/metacontract/parser.go
package metacontract

import (
	"fmt"
	"os"
	
	"gopkg.in/yaml.v3"
)

// Parser implements the meta-contract parsing functionality
type Parser struct{}

// NewParser creates a new meta-contract parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseMetaContract reads and parses a meta-contract YAML file
func (p *Parser) ParseMetaContract(yamlPath string) (*MetaContract, error) {
	// Read the YAML file
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta-contract file %s: %w", yamlPath, err)
	}
	
	// Parse the YAML content
	var contract MetaContract
	if err := yaml.Unmarshal(data, &contract); err != nil {
		return nil, fmt.Errorf("failed to parse meta-contract YAML: %w", err)
	}
	
	// Validate the parsed contract
	if err := p.validateContract(&contract); err != nil {
		return nil, fmt.Errorf("meta-contract validation failed: %w", err)
	}
	
	return &contract, nil
}

// validateContract performs basic validation on the parsed contract
func (p *Parser) validateContract(contract *MetaContract) error {
	if contract.ResourceName == "" {
		return fmt.Errorf("resource_name is required")
	}
	
	if contract.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	
	if len(contract.DataStructure.Fields) == 0 {
		return fmt.Errorf("at least one field must be defined in data_structure")
	}
	
	// Validate primary key field exists
	if contract.DataStructure.PrimaryKey != "" {
		found := false
		for _, field := range contract.DataStructure.Fields {
			if field.Name == contract.DataStructure.PrimaryKey {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("primary_key field '%s' not found in fields definition", contract.DataStructure.PrimaryKey)
		}
	}
	
	// Validate field types
	validTypes := map[string]bool{
		"string":    true,
		"int":       true,
		"int64":     true,
		"float64":   true,
		"bool":      true,
		"time":      true,
		"uuid":      true,
		"enum":      true,
		"json":      true,
	}
	
	for _, field := range contract.DataStructure.Fields {
		if !validTypes[field.Type] {
			return fmt.Errorf("invalid field type '%s' for field '%s'", field.Type, field.Name)
		}
	}
	
	return nil
}