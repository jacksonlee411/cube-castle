// internal/metacontract/validator.go
package metacontract

import (
	"fmt"
	"strings"
)

// Validator validates meta-contract specifications
type Validator struct{}

// NewValidator creates a new meta-contract validator
func NewValidator() *Validator {
	return &Validator{}
}

// Validate performs comprehensive validation of a meta-contract
func (v *Validator) Validate(contract *MetaContract) error {
	if err := v.validateBasicStructure(contract); err != nil {
		return err
	}
	
	if err := v.validateDataStructure(contract); err != nil {
		return err
	}
	
	if err := v.validateSecurityModel(contract); err != nil {
		return err
	}
	
	if err := v.validateTemporalBehavior(contract); err != nil {
		return err
	}
	
	if err := v.validateRelationships(contract); err != nil {
		return err
	}
	
	return nil
}

// validateBasicStructure validates the basic contract structure
func (v *Validator) validateBasicStructure(contract *MetaContract) error {
	if contract.SpecificationVersion == "" {
		return fmt.Errorf("specification_version is required")
	}
	
	if contract.ResourceName == "" {
		return fmt.Errorf("resource_name is required")
	}
	
	if contract.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	
	if contract.Version == "" {
		return fmt.Errorf("version is required")
	}
	
	return nil
}

// validateDataStructure validates the data structure definition
func (v *Validator) validateDataStructure(contract *MetaContract) error {
	if len(contract.DataStructure.Fields) == 0 {
		return fmt.Errorf("at least one field must be defined in data_structure")
	}
	
	// Validate field names are unique
	fieldNames := make(map[string]bool)
	for _, field := range contract.DataStructure.Fields {
		if fieldNames[field.Name] {
			return fmt.Errorf("duplicate field name: %s", field.Name)
		}
		fieldNames[field.Name] = true
		
		if err := v.validateField(field); err != nil {
			return fmt.Errorf("field '%s' validation failed: %w", field.Name, err)
		}
	}
	
	// Validate primary key field exists
	if contract.DataStructure.PrimaryKey != "" {
		if !fieldNames[contract.DataStructure.PrimaryKey] {
			return fmt.Errorf("primary_key field '%s' not found in fields definition", contract.DataStructure.PrimaryKey)
		}
	}
	
	return nil
}

// validateField validates a single field definition
func (v *Validator) validateField(field FieldDefinition) error {
	if field.Name == "" {
		return fmt.Errorf("field name is required")
	}
	
	if !isValidFieldName(field.Name) {
		return fmt.Errorf("invalid field name format: %s", field.Name)
	}
	
	if !isValidFieldType(field.Type) {
		return fmt.Errorf("invalid field type: %s", field.Type)
	}
	
	if field.DataClassification != "" && !isValidDataClassification(field.DataClassification) {
		return fmt.Errorf("invalid data classification: %s", field.DataClassification)
	}
	
	return nil
}

// validateSecurityModel validates the security model
func (v *Validator) validateSecurityModel(contract *MetaContract) error {
	validAccessControls := map[string]bool{
		"RBAC":  true,
		"ABAC":  true,
		"DAC":   true,
		"MAC":   true,
	}
	
	if contract.SecurityModel.AccessControl != "" {
		if !validAccessControls[contract.SecurityModel.AccessControl] {
			return fmt.Errorf("invalid access control model: %s", contract.SecurityModel.AccessControl)
		}
	}
	
	if contract.SecurityModel.DataClassification != "" {
		if !isValidDataClassification(contract.SecurityModel.DataClassification) {
			return fmt.Errorf("invalid data classification: %s", contract.SecurityModel.DataClassification)
		}
	}
	
	return nil
}

// validateTemporalBehavior validates temporal behavior configuration
func (v *Validator) validateTemporalBehavior(contract *MetaContract) error {
	validParadigms := map[string]bool{
		"EVENT_DRIVEN": true,
		"SNAPSHOT":     true,
		"HYBRID":       true,
	}
	
	if contract.TemporalBehavior.TemporalityParadigm != "" {
		if !validParadigms[contract.TemporalBehavior.TemporalityParadigm] {
			return fmt.Errorf("invalid temporality paradigm: %s", contract.TemporalBehavior.TemporalityParadigm)
		}
	}
	
	validStateModels := map[string]bool{
		"EVENT_DRIVEN":   true,
		"STATE_MACHINE":  true,
		"IMMUTABLE":      true,
	}
	
	if contract.TemporalBehavior.StateTransitionModel != "" {
		if !validStateModels[contract.TemporalBehavior.StateTransitionModel] {
			return fmt.Errorf("invalid state transition model: %s", contract.TemporalBehavior.StateTransitionModel)
		}
	}
	
	return nil
}

// validateRelationships validates relationship definitions
func (v *Validator) validateRelationships(contract *MetaContract) error {
	validRelationshipTypes := map[string]bool{
		"one-to-one":   true,
		"one-to-many":  true,
		"many-to-many": true,
	}
	
	for _, rel := range contract.Relationships {
		if rel.Name == "" {
			return fmt.Errorf("relationship name is required")
		}
		
		if !validRelationshipTypes[rel.Type] {
			return fmt.Errorf("invalid relationship type: %s", rel.Type)
		}
		
		if rel.TargetEntity == "" {
			return fmt.Errorf("target_entity is required for relationship: %s", rel.Name)
		}
	}
	
	return nil
}

// Helper validation functions
func isValidFieldName(name string) bool {
	if name == "" {
		return false
	}
	
	// Check if it's a valid identifier (alphanumeric + underscore, starting with letter or underscore)
	if !((name[0] >= 'a' && name[0] <= 'z') || (name[0] >= 'A' && name[0] <= 'Z') || name[0] == '_') {
		return false
	}
	
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	
	return true
}

func isValidFieldType(fieldType string) bool {
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
	
	return validTypes[fieldType]
}

func isValidDataClassification(classification string) bool {
	validClassifications := map[string]bool{
		"PUBLIC":        true,
		"INTERNAL":      true,
		"CONFIDENTIAL":  true,
		"RESTRICTED":    true,
	}
	
	return validClassifications[strings.ToUpper(classification)]
}