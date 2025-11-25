package noop

import "cube-castle/internal/standardobject"

// Provide wires the phase-A adapter to the dependency injection graph.
func Provide() standardobject.ObjectService {
	return standardobject.NewNoopService()
}
