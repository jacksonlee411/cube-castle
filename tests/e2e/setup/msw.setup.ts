/**
 * Mock Service Worker (MSW) setup for API mocking in tests
 */
import { setupServer } from 'msw/node';
import { rest } from 'msw';
import { IntelligentTemplate, TemplateCategory, TemplateComplexity } from '@/types/template';

// Mock data for templates
const mockTemplates: IntelligentTemplate[] = [
  {
    id: 'test-template-1',
    name: 'Test Employee Template',
    description: 'Test template for employee management',
    category: TemplateCategory.HR_MANAGEMENT,
    complexity: TemplateComplexity.BASIC,
    version: '1.0.0',
    author: {
      id: 'test-author',
      name: 'Test Author',
      organization: 'Test Org',
      verified: true
    },
    createdAt: new Date('2024-01-01'),
    updatedAt: new Date('2024-01-02'),
    schema: {
      specification_version: '1.0',
      api_id: 'test-api',
      namespace: 'test',
      resource_name: 'employees',
      data_structure: {
        primary_key: 'id',
        data_classification: 'internal',
        fields: []
      }
    },
    elements: [],
    tags: ['test', 'employee'],
    keywords: ['test', 'employee'],
    compatibility: {
      minSpecVersion: '1.0',
      supportedDatabases: ['postgresql'],
      supportedFrameworks: ['rest']
    },
    quality: {
      performanceScore: 85,
      securityScore: 90,
      maintainabilityScore: 88,
      bestPracticesScore: 92,
      communityRating: 4.5,
      usageCount: 100,
      lastValidated: new Date()
    },
    configurable: true
  }
];

// Mock project data
const mockProject = {
  id: 'test-project-1',
  name: 'Test Meta-Contract Project',
  description: 'Test project for unit testing',
  content: 'specification_version: "1.0"\napi_id: "test-api"',
  version: '1.0.0',
  status: 'draft' as const,
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-02T00:00:00Z'
};

// Mock compilation results
const mockCompileResults = {
  success: true,
  errors: [],
  warnings: [],
  generated_files: {
    'schema.sql': 'CREATE TABLE employees (id UUID PRIMARY KEY);',
    'api.yaml': 'openapi: 3.0.0\ninfo:\n  title: Test API'
  },
  schema: mockProject,
  compile_time: '2024-01-02T00:00:00Z'
};

// API request handlers
export const handlers = [
  // Meta-contract editor API endpoints
  rest.get('/api/v1/metacontract-editor/projects/:id', (req, res, ctx) => {
    return res(ctx.json(mockProject));
  }),

  rest.get('/api/v1/metacontract-editor/projects', (req, res, ctx) => {
    return res(ctx.json({ 
      projects: [mockProject], 
      total: 1, 
      limit: 20, 
      offset: 0 
    }));
  }),

  rest.post('/api/v1/metacontract-editor/projects', (req, res, ctx) => {
    return res(ctx.json({ ...mockProject, id: 'new-project-id' }));
  }),

  rest.put('/api/v1/metacontract-editor/projects/:id', (req, res, ctx) => {
    return res(ctx.json({ ...mockProject, updated_at: new Date().toISOString() }));
  }),

  rest.delete('/api/v1/metacontract-editor/projects/:id', (req, res, ctx) => {
    return res(ctx.status(204));
  }),

  rest.post('/api/v1/metacontract-editor/projects/:id/compile', (req, res, ctx) => {
    return res(ctx.json(mockCompileResults));
  }),

  rest.post('/api/v1/metacontract-editor/compile', (req, res, ctx) => {
    return res(ctx.json(mockCompileResults));
  }),

  rest.post('/api/v1/metacontract-editor/projects/:id/sessions', (req, res, ctx) => {
    return res(ctx.json({ 
      session_id: 'test-session-id',
      project_id: req.params.id,
      created_at: new Date().toISOString()
    }));
  }),

  rest.delete('/api/v1/metacontract-editor/sessions/:id', (req, res, ctx) => {
    return res(ctx.status(204));
  }),

  rest.get('/api/v1/metacontract-editor/templates', (req, res, ctx) => {
    const category = req.url.searchParams.get('category');
    const filteredTemplates = category 
      ? mockTemplates.filter(t => t.category === category)
      : mockTemplates;
    
    return res(ctx.json({ templates: filteredTemplates }));
  }),

  // Template recommendation API
  rest.post('/api/v1/templates/recommend', (req, res, ctx) => {
    return res(ctx.json({
      recommendations: mockTemplates.slice(0, 3),
      confidence_score: 0.85,
      reasoning: 'Based on project context and user preferences'
    }));
  }),

  // WebSocket mock (will be handled differently)
  rest.get('/ws/metacontract-editor/:projectId', (req, res, ctx) => {
    return res(ctx.status(101)); // Switching Protocols
  }),

  // Error scenarios for testing
  rest.get('/api/v1/metacontract-editor/projects/error-project', (req, res, ctx) => {
    return res(ctx.status(404), ctx.json({ error: 'Project not found' }));
  }),

  rest.post('/api/v1/metacontract-editor/projects/compile-error/compile', (req, res, ctx) => {
    return res(ctx.json({
      success: false,
      errors: [
        {
          line: 5,
          column: 12,
          message: 'Invalid YAML syntax',
          severity: 'error'
        }
      ],
      warnings: []
    }));
  })
];

// Setup MSW server
export const server = setupServer(...handlers);

// Setup and teardown functions
export const setupMSW = () => {
  beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
  afterEach(() => server.resetHandlers());
  afterAll(() => server.close());
};

// Helper functions for tests
export const mockWebSocketConnection = () => {
  const mockWebSocket = {
    readyState: WebSocket.OPEN,
    send: jest.fn(),
    close: jest.fn(),
    addEventListener: jest.fn(),
    removeEventListener: jest.fn()
  };

  // Mock WebSocket constructor
  global.WebSocket = jest.fn(() => mockWebSocket) as any;
  
  return mockWebSocket;
};

export const triggerWebSocketMessage = (mockWS: any, message: any) => {
  const messageEvent = new MessageEvent('message', {
    data: JSON.stringify(message)
  });
  
  // Simulate message reception
  const messageHandlers = mockWS.addEventListener.mock.calls
    .filter(([event]: any) => event === 'message')
    .map(([, handler]: any) => handler);
  
  messageHandlers.forEach(handler => handler(messageEvent));
};

// Mock local storage
export const mockLocalStorage = () => {
  const localStorageMock = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
  };
  
  Object.defineProperty(window, 'localStorage', {
    value: localStorageMock
  });
  
  return localStorageMock;
};

// Mock Monaco Editor
export const mockMonacoEditor = () => {
  const mockEditor = {
    getValue: jest.fn(() => 'test content'),
    setValue: jest.fn(),
    setPosition: jest.fn(),
    revealLineInCenter: jest.fn(),
    focus: jest.fn(),
    dispose: jest.fn(),
    onDidChangeModelContent: jest.fn(),
    getModel: jest.fn(() => ({
      onDidChangeContent: jest.fn(),
      setValue: jest.fn(),
      getValue: jest.fn(() => 'test content')
    }))
  };

  jest.mock('@monaco-editor/react', () => ({
    Editor: jest.fn(({ onChange, value }) => {
      // Simulate editor changes
      if (onChange) {
        setTimeout(() => onChange('updated content'), 0);
      }
      return <div data-testid="monaco-editor">{value}</div>;
    }),
    loader: {
      init: jest.fn(() => Promise.resolve())
    }
  }));

  return mockEditor;
};

// Mock drag and drop
export const mockDragAndDrop = () => {
  const mockDataTransfer = {
    getData: jest.fn(),
    setData: jest.fn(),
    clearData: jest.fn(),
    items: [],
    files: [],
    types: []
  };

  // Mock drag events
  const createDragEvent = (type: string, data: any = {}) => {
    const event = new Event(type, { bubbles: true });
    Object.assign(event, {
      dataTransfer: mockDataTransfer,
      preventDefault: jest.fn(),
      stopPropagation: jest.fn(),
      ...data
    });
    return event;
  };

  return { mockDataTransfer, createDragEvent };
};

export { mockTemplates, mockProject, mockCompileResults };