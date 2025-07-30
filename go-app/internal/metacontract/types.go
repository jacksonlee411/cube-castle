// internal/metacontract/types.go
package metacontract

// This file is deprecated. All types have been moved to internal/types package.
// This file is kept for backward compatibility and will be removed in future versions.

import (
	"github.com/gaogu/cube-castle/go-app/internal/types"
)

// Re-export types for backward compatibility
type MetaContract = types.MetaContract
type DataStructure = types.DataStructure
type FieldDefinition = types.FieldDefinition
type SecurityModel = types.SecurityModel
type TemporalBehaviorModel = types.TemporalBehaviorModel
type APIBehaviorModel = types.APIBehaviorModel
type RelationshipDef = types.RelationshipDef
type CompilerInterface = types.CompilerInterface
