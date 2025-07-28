package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/localai"
	"github.com/gaogu/cube-castle/go-app/internal/metacontract"
	"github.com/gaogu/cube-castle/go-app/internal/metacontracteditor"
	"github.com/gaogu/cube-castle/go-app/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCompiler is a mock implementation of the CompilerInterface
type MockCompiler struct {
	mock.Mock
}

func (m *MockCompiler) ParseMetaContract(yamlPath string) (*types.MetaContract, error) {
	args := m.Called(yamlPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.MetaContract), args.Error(1)
}

func (m *MockCompiler) GenerateEntSchemas(contract *types.MetaContract, outputDir string) error {
	args := m.Called(contract, outputDir)
	return args.Error(0)
}

func (m *MockCompiler) GenerateBusinessLogic(contract *types.MetaContract, outputDir string) error {
	args := m.Called(contract, outputDir)
	return args.Error(0)
}

func (m *MockCompiler) GenerateAPIRoutes(contract *types.MetaContract, outputDir string) error {
	args := m.Called(contract, outputDir)
	return args.Error(0)
}

// MockLocalAIService is a mock implementation of the LocalAI service
type MockLocalAIService struct {
	mock.Mock
}

func (m *MockLocalAIService) ProcessRequest(ctx context.Context, req *localai.AIRequest) (*localai.AIResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*localai.AIResponse), args.Error(1)
}

// MockWebSocketHub is a mock implementation of the WebSocket hub
type MockWebSocketHub struct {
	mock.Mock
}

func (m *MockWebSocketHub) RegisterClient(userID, sessionID uuid.UUID) *metacontracteditor.Client {
	args := m.Called(userID, sessionID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*metacontracteditor.Client)
}

func (m *MockWebSocketHub) RegisterSession(sessionID uuid.UUID, session *metacontracteditor.EditorSession) {
	m.Called(sessionID, session)
}

func (m *MockWebSocketHub) UnregisterSession(sessionID uuid.UUID) {
	m.Called(sessionID)
}

func (m *MockWebSocketHub) SubscribeToProject(client *metacontracteditor.Client, projectID uuid.UUID) {
	m.Called(client, projectID)
}

func (m *MockWebSocketHub) UnsubscribeFromProject(client *metacontracteditor.Client, projectID uuid.UUID) {
	m.Called(client, projectID)
}

func (m *MockWebSocketHub) BroadcastToProject(projectID uuid.UUID, message *metacontracteditor.WebSocketMessage) {
	m.Called(projectID, message)
}

// MockNLPEngine is a mock implementation of the NLP engine
type MockNLPEngine struct {
	mock.Mock
}

func (m *MockNLPEngine) Process(req *localai.NLPRequest) (*localai.NLPResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*localai.NLPResponse), args.Error(1)
}

// MockCodeAnalyzer is a mock implementation of the code analyzer
type MockCodeAnalyzer struct {
	mock.Mock
}

func (m *MockCodeAnalyzer) Analyze(context, query string) (*localai.AnalysisResult, error) {
	args := m.Called(context, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*localai.AnalysisResult), args.Error(1)
}

// Test Data Builders

// CreateValidMetaContract creates a valid meta-contract for testing
func CreateValidMetaContract() *types.MetaContract {
	return &types.MetaContract{
		SpecificationVersion: "1.0",
		APIID:               uuid.New(),
		Namespace:           "test",
		ResourceName:        "user",
		Version:            "1.0.0",
		DataStructure: types.DataStructure{
			PrimaryKey:         "id",
			DataClassification: "INTERNAL",
			Fields: []types.FieldDefinition{
				{
					Name:               "id",
					Type:               "uuid",
					Required:           true,
					Unique:             true,
					DataClassification: "INTERNAL",
				},
				{
					Name:               "email",
					Type:               "string",
					Required:           true,
					Unique:             true,
					DataClassification: "CONFIDENTIAL",
					ValidationRules:    []string{"email_format"},
				},
				{
					Name:               "name",
					Type:               "string",
					Required:           true,
					DataClassification: "INTERNAL",
				},
			},
		},
		SecurityModel: types.SecurityModel{
			TenantIsolation:    true,
			AccessControl:      "RBAC",
			DataClassification: "CONFIDENTIAL",
			ComplianceTags:     []string{"GDPR", "CCPA"},
		},
		TemporalBehavior: types.TemporalBehaviorModel{
			TemporalityParadigm:  "EVENT_DRIVEN",
			StateTransitionModel: "EVENT_DRIVEN",
			HistoryRetention:     "7 years",
			EventDriven:          true,
		},
		APIBehavior: types.APIBehaviorModel{
			RESTEnabled:    true,
			GraphQLEnabled: true,
			EventsEnabled:  true,
		},
		Relationships: []types.RelationshipDef{
			{
				Name:         "user_profile",
				Type:         "one-to-one",
				TargetEntity: "profile",
				Cardinality:  "1:1",
				IsOptional:   false,
			},
		},
	}
}

// CreateValidAIRequest creates a valid AI request for testing
func CreateValidAIRequest() *localai.AIRequest {
	return &localai.AIRequest{
		Type:    "completion",
		Context: "entities:\n  - name: user\n    fields:",
		Query:   "id",
		Position: localai.CursorPosition{
			Line:   3,
			Column: 10,
		},
		Metadata:  map[string]string{"test": "true"},
		SessionID: "test-session",
	}
}

// CreateValidAIResponse creates a valid AI response for testing
func CreateValidAIResponse() *localai.AIResponse {
	return &localai.AIResponse{
		Type: "completion",
		Suggestions: []localai.Suggestion{
			{
				Label:       "id",
				InsertText:  "id:",
				Detail:      "Primary key field",
				Kind:        "field",
				Priority:    90,
				Description: "Unique identifier field",
			},
			{
				Label:       "name",
				InsertText:  "name:",
				Detail:      "String field",
				Kind:        "field",
				Priority:    80,
				Description: "Name field",
			},
		},
		ProcessTime: 50 * time.Millisecond,
	}
}

// CreateValidWebSocketMessage creates a valid WebSocket message for testing
func CreateValidWebSocketMessage() *metacontracteditor.WebSocketMessage {
	return &metacontracteditor.WebSocketMessage{
		Type:      metacontracteditor.MessageTypeContentChange,
		ProjectID: uuid.New(),
		UserID:    uuid.New(),
		Data: map[string]interface{}{
			"content": "test content change",
			"line":    10,
			"column":  5,
		},
		Timestamp: time.Now(),
	}
}

// CreateValidEditorSession creates a valid editor session for testing
func CreateValidEditorSession() *metacontracteditor.EditorSession {
	return &metacontracteditor.EditorSession{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		ProjectID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Active:    true,
	}
}

// CreateValidNLPRequest creates a valid NLP request for testing
func CreateValidNLPRequest() *localai.NLPRequest {
	return &localai.NLPRequest{
		Text:      "Create a user entity with id, name, and email fields",
		Context:   "",
		SessionID: "test-session",
		Language:  "en",
	}
}

// CreateValidNLPResponse creates a valid NLP response for testing
func CreateValidNLPResponse() *localai.NLPResponse {
	return &localai.NLPResponse{
		Intent: "create_entity",
		Entities: []localai.NLPEntity{
			{
				Type:       "entity",
				Value:      "user",
				Confidence: 0.95,
				Context:    "user entity",
			},
			{
				Type:       "field",
				Value:      "id",
				Confidence: 0.90,
				Context:    "id field",
			},
			{
				Type:       "field",
				Value:      "name",
				Confidence: 0.90,
				Context:    "name field",
			},
			{
				Type:       "field",
				Value:      "email",
				Confidence: 0.90,
				Context:    "email field",
			},
		},
		GeneratedYAML: `name: user
fields:
  - name: id
    type: uuid
    required: true
  - name: name
    type: string
    required: true
  - name: email
    type: string
    required: true
    unique: true`,
		Explanation:  "Created a user entity with the specified fields: id, name, and email",
		Confidence:   0.92,
		Alternatives: []string{"user table", "user model"},
		RequiredInfo: []string{},
	}
}

// CreateValidAnalysisResult creates a valid analysis result for testing
func CreateValidAnalysisResult() *localai.AnalysisResult {
	return &localai.AnalysisResult{
		Issues: []localai.Issue{
			{
				Type:       "warning",
				Message:    "Missing primary key definition",
				Line:       5,
				Column:     10,
				Severity:   "medium",
				Suggestion: "Add a primary key field",
				Category:   "semantic",
			},
		},
		Suggestions: []string{
			"Consider adding indexes for better performance",
			"Add validation rules for email field",
		},
		Complexity: 3,
		Performance: &localai.PerformanceAnalysis{
			Score:            75,
			Bottlenecks:      []string{"Missing indexes on frequently queried fields"},
			Recommendations:  []string{"Add index on email field", "Consider partitioning for large datasets"},
			QueryComplexity:  "Medium",
			IndexSuggestions: []string{"CREATE INDEX idx_user_email ON user(email)"},
		},
		Security: &localai.SecurityAnalysis{
			Score:           80,
			Vulnerabilities: []string{"Email field not encrypted"},
			Recommendations: []string{"Consider encrypting sensitive fields", "Add access control policies"},
			Compliance:      []string{"GDPR compliance needed for email field"},
			DataSensitivity: "CONFIDENTIAL",
		},
		Dependencies: []string{"uuid", "string", "time"},
		Relationships: []localai.Relationship{
			{
				From:        "user",
				To:          "profile",
				Type:        "one-to-one",
				Cardinality: "1:1",
				Description: "User has one profile",
			},
		},
	}
}

// Test Helpers

// AssertNoError is a helper function to assert no error and fail fast
func AssertNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}
}

// AssertError is a helper function to assert an error exists
func AssertError(t *testing.T, err error) {
	if err == nil {
		t.Fatal("Expected an error but got nil")
	}
}

// AssertContains is a helper function to assert a string contains a substring
func AssertContains(t *testing.T, haystack, needle string) {
	if !strings.Contains(haystack, needle) {
		t.Fatalf("Expected '%s' to contain '%s'", haystack, needle)
	}
}

// AssertNotContains is a helper function to assert a string does not contain a substring
func AssertNotContains(t *testing.T, haystack, needle string) {
	if strings.Contains(haystack, needle) {
		t.Fatalf("Expected '%s' to not contain '%s'", haystack, needle)
	}
}

// CreateTestConfig creates a test configuration for AI service
func CreateTestConfig() *localai.AIConfiguration {
	return &localai.AIConfiguration{
		CacheSize:          100,
		CacheTTL:           5 * time.Minute,
		MaxTokens:          2048,
		Temperature:        0.7,
		TopP:               0.9,
		ContextWindow:      4096,
		EnableCodeCompl:    true,
		EnableNLP:          true,
		EnableAnalysis:     true,
		EnableOptimization: true,
	}
}

// ValidateMetaContract validates a meta-contract for testing purposes
func ValidateMetaContract(t *testing.T, contract *types.MetaContract) {
	if contract == nil {
		t.Fatal("Contract is nil")
	}
	
	if contract.ResourceName == "" {
		t.Error("ResourceName is empty")
	}
	
	if contract.Namespace == "" {
		t.Error("Namespace is empty")
	}
	
	if len(contract.DataStructure.Fields) == 0 {
		t.Error("No fields defined")
	}
	
	// Validate primary key exists if specified
	if contract.DataStructure.PrimaryKey != "" {
		found := false
		for _, field := range contract.DataStructure.Fields {
			if field.Name == contract.DataStructure.PrimaryKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Primary key field '%s' not found in fields", contract.DataStructure.PrimaryKey)
		}
	}
}

// ValidateAIResponse validates an AI response for testing purposes
func ValidateAIResponse(t *testing.T, response *localai.AIResponse) {
	if response == nil {
		t.Fatal("Response is nil")
	}
	
	if response.Type == "" {
		t.Error("Response type is empty")
	}
	
	if response.ProcessTime <= 0 {
		t.Error("Process time should be greater than 0")
	}
	
	// Validate suggestions if present
	for i, suggestion := range response.Suggestions {
		if suggestion.Label == "" {
			t.Errorf("Suggestion %d has empty label", i)
		}
		if suggestion.Priority <= 0 {
			t.Errorf("Suggestion %d has invalid priority", i)
		}
	}
}

// SetupTestEnvironment sets up a test environment with temporary directories
func SetupTestEnvironment(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}
	
	return tmpDir, cleanup
}