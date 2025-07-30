// Package intelligencegateway provides intelligence gateway types and interfaces
// This is a placeholder implementation for testing purposes
package intelligencegateway

import "context"

// Gateway request types
type GatewayRequest struct {
	Query     string            `json:"query"`
	Context   string            `json:"context"`
	Metadata  map[string]string `json:"metadata"`
	SessionID string            `json:"session_id"`
}

type GatewayResponse struct {
	Response   string            `json:"response"`
	Confidence float64           `json:"confidence"`
	Sources    []string          `json:"sources"`
	Metadata   map[string]string `json:"metadata"`
	Error      string            `json:"error,omitempty"`
}

// Intelligence types
type IntelligenceConfig struct {
	Providers []ProviderConfig `json:"providers"`
	Strategy  string           `json:"strategy"`
	Timeout   int              `json:"timeout"`
}

type ProviderConfig struct {
	Name     string            `json:"name"`
	Type     string            `json:"type"`
	Endpoint string            `json:"endpoint"`
	Config   map[string]string `json:"config"`
}

// Gateway interface
type IntelligenceGateway interface {
	Process(ctx context.Context, req *GatewayRequest) (*GatewayResponse, error)
	GetHealth() error
}
