// Package localai provides AI-related types and interfaces
// This is a placeholder implementation for testing purposes
package localai

import "context"

// AI Request types
type AIRequest struct {
	Query    string         `json:"query"`
	Context  string         `json:"context"`
	Position CursorPosition `json:"position"`
}

type AIResponse struct {
	Suggestions []Suggestion `json:"suggestions"`
	Error       string       `json:"error,omitempty"`
}

type CursorPosition struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Suggestion struct {
	Text        string  `json:"text"`
	Type        string  `json:"type"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// NLP types
type NLPRequest struct {
	Text     string            `json:"text"`
	Context  string            `json:"context"`
	Metadata map[string]string `json:"metadata"`
}

type NLPResponse struct {
	Entities []NLPEntity `json:"entities"`
	Intent   string      `json:"intent"`
	Error    string      `json:"error,omitempty"`
}

type NLPEntity struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
}

// Analysis types
type AnalysisResult struct {
	Issues        []Issue              `json:"issues"`
	Performance   *PerformanceAnalysis `json:"performance"`
	Security      *SecurityAnalysis    `json:"security"`
	Relationships []Relationship       `json:"relationships"`
}

type Issue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Suggestion string `json:"suggestion"`
}

type PerformanceAnalysis struct {
	Complexity int     `json:"complexity"`
	Score      float64 `json:"score"`
	Issues     []Issue `json:"issues"`
}

type SecurityAnalysis struct {
	Vulnerabilities []Issue `json:"vulnerabilities"`
	Score           float64 `json:"score"`
}

type Relationship struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
}

// Configuration
type AIConfiguration struct {
	Endpoint string            `json:"endpoint"`
	APIKey   string            `json:"api_key"`
	Model    string            `json:"model"`
	Settings map[string]string `json:"settings"`
}

// Service interface
type LocalAIService interface {
	ProcessRequest(ctx context.Context, req *AIRequest) (*AIResponse, error)
}

// NLP Engine interface
type NLPEngine interface {
	Process(req *NLPRequest) (*NLPResponse, error)
}

// Code Analyzer interface
type CodeAnalyzer interface {
	Analyze(context, query string) (*AnalysisResult, error)
}
