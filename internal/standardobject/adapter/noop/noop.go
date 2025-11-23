package noop

import (
	"cube-castle/internal/standardobject"
	"cube-castle/internal/standardobject/featureflag"
)

// Provide wires the phase-A adapter to the dependency injection graph.
func Provide(toggle featureflag.Toggle) standardobject.ObjectService {
	return standardobject.NewNoopService(toggle)
}
