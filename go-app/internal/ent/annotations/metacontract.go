// internal/ent/annotations/metacontract.go
package annotations

// MetaContractAnnotation holds governance metadata from the meta-contract
type MetaContractAnnotation struct {
	DataClassification string              `json:"data_classification,omitempty"`
	ComplianceTags     []string            `json:"compliance_tags,omitempty"`
	PersistenceProfile *PersistenceProfile `json:"persistence_profile,omitempty"`
}

// PersistenceProfile defines persistence configuration
type PersistenceProfile struct {
	PrimaryStore         string   `json:"primary_store"`
	IndexedIn            []string `json:"indexed_in,omitempty"`
	GraphNodeLabel       string   `json:"graph_node_label,omitempty"`
	GraphEdgeDefinitions []string `json:"graph_edge_definitions,omitempty"`
}

// Name implements the ent.Annotation interface
func (MetaContractAnnotation) Name() string {
	return "MetaContract"
}
