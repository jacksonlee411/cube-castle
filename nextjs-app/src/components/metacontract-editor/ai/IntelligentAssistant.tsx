// src/components/metacontract-editor/ai/IntelligentAssistant.tsx
import React, { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Brain, 
  MessageSquare, 
  Code2, 
  Zap, 
  AlertTriangle, 
  CheckCircle, 
  Lightbulb, 
  Send, 
  Copy, 
  Sparkles,
  Settings,
  Mic,
  MicOff
} from 'lucide-react';

interface IntelligentAssistantProps {
  content: string;
  cursorPosition: { line: number; column: number };
  onInsertCode: (code: string) => void;
  onReplaceCode: (oldCode: string, newCode: string) => void;
  projectId?: string;
  className?: string;
}

interface AIResponse {
  type: string;
  suggestions?: Suggestion[];
  analysis?: AnalysisResult;
  generation?: GenerationResult;
  optimization?: OptimizationResult;
  error?: string;
  processTime?: number;
}

interface Suggestion {
  label: string;
  insertText: string;
  detail: string;
  kind: string;
  priority: number;
  description: string;
}

interface AnalysisResult {
  issues: Issue[];
  suggestions: string[];
  complexity: number;
  performance: PerformanceAnalysis;
  security: SecurityAnalysis;
  dependencies: string[];
  relationships: Relationship[];
}

interface Issue {
  type: string;
  message: string;
  line: number;
  column: number;
  severity: string;
  category: string;
  suggestion: string;
}

interface GenerationResult {
  generatedYAML: string;
  explanation: string;
  confidence: number;
  alternatives: string[];
  metadata: Record<string, string>;
}

interface OptimizationResult {
  optimizations: Optimization[];
  impact: string;
  effort: string;
  priority: string;
}

interface Optimization {
  type: string;
  description: string;
  before: string;
  after: string;
  benefit: string;
  risk: string;
}

interface PerformanceAnalysis {
  score: number;
  bottlenecks: string[];
  recommendations: string[];
  queryComplexity: string;
  indexSuggestions: string[];
}

interface SecurityAnalysis {
  score: number;
  vulnerabilities: string[];
  recommendations: string[];
  compliance: string[];
  dataSensitivity: string;
}

interface Relationship {
  from: string;
  to: string;
  type: string;
  cardinality: string;
  description: string;
}

export const IntelligentAssistant: React.FC<IntelligentAssistantProps> = ({
  content,
  cursorPosition,
  onInsertCode,
  onReplaceCode,
  projectId,
  className
}) => {
  const [activeTab, setActiveTab] = useState('chat');
  const [chatInput, setChatInput] = useState('');
  const [chatHistory, setChatHistory] = useState<Array<{role: string; content: string; timestamp: Date}>>([]);
  const [isProcessing, setIsProcessing] = useState(false);
  const [analysisResult, setAnalysisResult] = useState<AnalysisResult | null>(null);
  const [optimizationResult, setOptimizationResult] = useState<OptimizationResult | null>(null);
  const [suggestions, setSuggestions] = useState<Suggestion[]>([]);
  const [isListening, setIsListening] = useState(false);
  const [aiConfig, setAiConfig] = useState({
    enableCompletion: true,
    enableAnalysis: true,
    enableOptimization: true,
    enableNLP: true
  });

  const wsRef = useRef<WebSocket | null>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);
  const recognitionRef = useRef<any>(null);

  // Initialize WebSocket connection for real-time AI assistance
  useEffect(() => {
    const connectWebSocket = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/api/ai/ws`;
      
      wsRef.current = new WebSocket(wsUrl);
      
      wsRef.current.onopen = () => {
        console.log('AI WebSocket connected');
        addChatMessage('system', 'AI Assistant connected. How can I help you with your metacontract?');
      };
      
      wsRef.current.onmessage = (event) => {
        try {
          const response = JSON.parse(event.data);
          handleAIResponse(response);
        } catch (error) {
          console.error('Failed to parse AI response:', error);
        }
      };
      
      wsRef.current.onclose = () => {
        console.log('AI WebSocket disconnected');
        // Attempt to reconnect after 3 seconds
        setTimeout(connectWebSocket, 3000);
      };
      
      wsRef.current.onerror = (error) => {
        console.error('AI WebSocket error:', error);
      };
    };

    connectWebSocket();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, []);

  // Initialize speech recognition
  useEffect(() => {
    if ('webkitSpeechRecognition' in window || 'SpeechRecognition' in window) {
      const SpeechRecognition = (window as any).webkitSpeechRecognition || (window as any).SpeechRecognition;
      recognitionRef.current = new SpeechRecognition();
      
      recognitionRef.current.continuous = false;
      recognitionRef.current.interimResults = false;
      recognitionRef.current.lang = 'en-US';
      
      recognitionRef.current.onresult = (event: any) => {
        const transcript = event.results[0][0].transcript;
        setChatInput(transcript);
        setIsListening(false);
      };
      
      recognitionRef.current.onerror = () => {
        setIsListening(false);
      };
      
      recognitionRef.current.onend = () => {
        setIsListening(false);
      };
    }
  }, []);

  // Auto-scroll chat to bottom
  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatHistory]);

  // Auto-analyze code when content changes
  useEffect(() => {
    if (content && aiConfig.enableAnalysis) {
      const timer = setTimeout(() => {
        analyzeCode();
      }, 1000); // Debounce analysis

      return () => clearTimeout(timer);
    }
  }, [content, aiConfig.enableAnalysis]);

  const addChatMessage = (role: string, content: string) => {
    setChatHistory(prev => [...prev, { role, content, timestamp: new Date() }]);
  };

  const sendAIRequest = (type: string, query: string = '', additionalData: any = {}) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      addChatMessage('system', 'AI Assistant is not connected. Please wait...');
      return;
    }

    const request = {
      type,
      context: content,
      query,
      position: cursorPosition,
      sessionId: projectId || 'default',
      metadata: { timestamp: new Date().toISOString(), ...additionalData }
    };

    wsRef.current.send(JSON.stringify(request));
    setIsProcessing(true);
  };

  const handleAIResponse = (response: AIResponse) => {
    setIsProcessing(false);

    switch (response.type) {
      case 'completion':
        if (response.suggestions) {
          setSuggestions(response.suggestions);
        }
        break;
      
      case 'analysis':
        if (response.analysis) {
          setAnalysisResult(response.analysis);
        }
        break;
      
      case 'generation':
        if (response.generation) {
          const { generatedYAML, explanation } = response.generation;
          addChatMessage('assistant', explanation);
          if (generatedYAML) {
            addChatMessage('assistant', `Generated YAML:\n\`\`\`yaml\n${generatedYAML}\n\`\`\``);
          }
        }
        break;
      
      case 'optimization':
        if (response.optimization) {
          setOptimizationResult(response.optimization);
        }
        break;
      
      case 'nlp':
        if (response.generation) {
          const { generatedYAML, explanation } = response.generation;
          addChatMessage('assistant', explanation);
          if (generatedYAML) {
            addChatMessage('assistant', `Here's the generated YAML:\n\`\`\`yaml\n${generatedYAML}\n\`\`\``);
          }
        }
        break;
      
      default:
        if (response.error) {
          addChatMessage('system', `Error: ${response.error}`);
        }
    }
  };

  const handleChatSubmit = () => {
    if (!chatInput.trim()) return;

    addChatMessage('user', chatInput);
    sendAIRequest('nlp', chatInput);
    setChatInput('');
  };

  const analyzeCode = () => {
    sendAIRequest('analysis', 'Analyze the current metacontract code');
  };

  const optimizeCode = () => {
    sendAIRequest('optimization', 'Suggest optimizations for the current code');
  };

  const getCompletions = () => {
    sendAIRequest('completion', 'Get code completions for current position');
  };

  const handleSuggestionApply = (suggestion: Suggestion) => {
    onInsertCode(suggestion.insertText);
  };

  const handleOptimizationApply = (optimization: Optimization) => {
    onReplaceCode(optimization.before, optimization.after);
    addChatMessage('system', `Applied optimization: ${optimization.description}`);
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    addChatMessage('system', 'Copied to clipboard!');
  };

  const startVoiceInput = () => {
    if (recognitionRef.current && !isListening) {
      setIsListening(true);
      recognitionRef.current.start();
    }
  };

  const stopVoiceInput = () => {
    if (recognitionRef.current && isListening) {
      recognitionRef.current.stop();
      setIsListening(false);
    }
  };

  const renderChatMessage = (message: {role: string; content: string; timestamp: Date}, index: number) => {
    const isUser = message.role === 'user';
    const isSystem = message.role === 'system';
    
    return (
      <div key={index} className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-3`}>
        <div className={`max-w-[80%] p-3 rounded-lg ${
          isUser 
            ? 'bg-blue-500 text-white' 
            : isSystem 
              ? 'bg-gray-100 text-gray-700 text-sm'
              : 'bg-gray-50 text-gray-900'
        }`}>
          <div className="whitespace-pre-wrap">{message.content}</div>
          <div className="text-xs mt-1 opacity-70">
            {message.timestamp.toLocaleTimeString()}
          </div>
        </div>
      </div>
    );
  };

  const renderAnalysisIssue = (issue: Issue, index: number) => {
    const severityColors = {
      error: 'destructive',
      warning: 'secondary',
      info: 'outline'
    };

    return (
      <div key={index} className="p-3 border rounded-lg mb-2">
        <div className="flex items-center justify-between mb-2">
          <Badge variant={severityColors[issue.severity as keyof typeof severityColors] as any}>
            {issue.severity}
          </Badge>
          <span className="text-sm text-gray-500">Line {issue.line}:{issue.column}</span>
        </div>
        <p className="text-sm font-medium mb-1">{issue.message}</p>
        <p className="text-xs text-gray-600">{issue.suggestion}</p>
      </div>
    );
  };

  const renderOptimization = (optimization: Optimization, index: number) => {
    return (
      <div key={index} className="p-3 border rounded-lg mb-3">
        <div className="flex items-center justify-between mb-2">
          <Badge variant="outline">{optimization.type}</Badge>
          <Button
            size="sm"
            onClick={() => handleOptimizationApply(optimization)}
            className="text-xs"
          >
            Apply
          </Button>
        </div>
        <p className="text-sm font-medium mb-2">{optimization.description}</p>
        <p className="text-xs text-green-600 mb-1">✓ {optimization.benefit}</p>
        <p className="text-xs text-amber-600">⚠ Risk: {optimization.risk}</p>
        
        <div className="mt-2">
          <details className="text-xs">
            <summary className="cursor-pointer text-gray-600 hover:text-gray-800">
              View changes
            </summary>
            <div className="mt-2 p-2 bg-gray-50 rounded">
              <div className="mb-2">
                <span className="text-red-600">- Before:</span>
                <pre className="text-xs mt-1 p-1 bg-red-50 rounded">{optimization.before}</pre>
              </div>
              <div>
                <span className="text-green-600">+ After:</span>
                <pre className="text-xs mt-1 p-1 bg-green-50 rounded">{optimization.after}</pre>
              </div>
            </div>
          </details>
        </div>
      </div>
    );
  };

  return (
    <Card className={`w-full h-full flex flex-col ${className}`}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2">
            <Brain className="w-5 h-5 text-blue-500" />
            AI Assistant
            {isProcessing && <Sparkles className="w-4 h-4 animate-spin text-amber-500" />}
          </CardTitle>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setActiveTab('settings')}
          >
            <Settings className="w-4 h-4" />
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 flex flex-col p-0">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
          <TabsList className="grid w-full grid-cols-4 mx-4">
            <TabsTrigger value="chat" className="flex items-center gap-1">
              <MessageSquare className="w-4 h-4" />
              Chat
            </TabsTrigger>
            <TabsTrigger value="analysis" className="flex items-center gap-1">
              <AlertTriangle className="w-4 h-4" />
              Analysis
              {analysisResult?.issues.length && (
                <Badge variant="destructive" className="ml-1 text-xs">
                  {analysisResult.issues.length}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="optimize" className="flex items-center gap-1">
              <Zap className="w-4 h-4" />
              Optimize
              {optimizationResult?.optimizations.length && (
                <Badge variant="secondary" className="ml-1 text-xs">
                  {optimizationResult.optimizations.length}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="suggestions" className="flex items-center gap-1">
              <Lightbulb className="w-4 h-4" />
              Suggest
              {suggestions.length > 0 && (
                <Badge variant="outline" className="ml-1 text-xs">
                  {suggestions.length}
                </Badge>
              )}
            </TabsTrigger>
          </TabsList>

          <TabsContent value="chat" className="flex-1 flex flex-col p-4">
            <div className="flex-1 overflow-y-auto mb-4 min-h-0">
              {chatHistory.map(renderChatMessage)}
              <div ref={chatEndRef} />
            </div>
            
            <div className="flex gap-2">
              <Textarea
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                placeholder="Ask me anything about your metacontract..."
                className="flex-1 resize-none"
                rows={2}
                onKeyDown={(e) => {
                  if (e.key === 'Enter' && !e.shiftKey) {
                    e.preventDefault();
                    handleChatSubmit();
                  }
                }}
              />
              <div className="flex flex-col gap-1">
                <Button
                  onClick={handleChatSubmit}
                  disabled={!chatInput.trim() || isProcessing}
                  size="sm"
                >
                  <Send className="w-4 h-4" />
                </Button>
                <Button
                  onClick={isListening ? stopVoiceInput : startVoiceInput}
                  variant="outline"
                  size="sm"
                  className={isListening ? 'bg-red-50 border-red-200' : ''}
                >
                  {isListening ? <MicOff className="w-4 h-4" /> : <Mic className="w-4 h-4" />}
                </Button>
              </div>
            </div>
          </TabsContent>

          <TabsContent value="analysis" className="flex-1 overflow-y-auto p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Code Analysis</h3>
              <Button onClick={analyzeCode} size="sm" disabled={isProcessing}>
                {isProcessing ? 'Analyzing...' : 'Re-analyze'}
              </Button>
            </div>

            {analysisResult ? (
              <div className="space-y-4">
                {/* Summary Cards */}
                <div className="grid grid-cols-2 gap-4">
                  <Card className="p-3">
                    <div className="text-xs text-gray-600">Complexity</div>
                    <div className="text-lg font-bold">{analysisResult.complexity}</div>
                  </Card>
                  <Card className="p-3">
                    <div className="text-xs text-gray-600">Issues</div>
                    <div className="text-lg font-bold">{analysisResult.issues.length}</div>
                  </Card>
                </div>

                {/* Performance Score */}
                {analysisResult.performance && (
                  <Card className="p-3">
                    <div className="text-xs text-gray-600 mb-1">Performance Score</div>
                    <div className="flex items-center gap-2">
                      <div className="text-lg font-bold">{analysisResult.performance.score}/100</div>
                      <div className={`text-xs px-2 py-1 rounded ${
                        analysisResult.performance.score >= 80 ? 'bg-green-100 text-green-800' :
                        analysisResult.performance.score >= 60 ? 'bg-yellow-100 text-yellow-800' :
                        'bg-red-100 text-red-800'
                      }`}>
                        {analysisResult.performance.score >= 80 ? 'Good' :
                         analysisResult.performance.score >= 60 ? 'Fair' : 'Poor'}
                      </div>
                    </div>
                  </Card>
                )}

                {/* Security Score */}
                {analysisResult.security && (
                  <Card className="p-3">
                    <div className="text-xs text-gray-600 mb-1">Security Score</div>
                    <div className="flex items-center gap-2">
                      <div className="text-lg font-bold">{analysisResult.security.score}/100</div>
                      <div className={`text-xs px-2 py-1 rounded ${
                        analysisResult.security.score >= 80 ? 'bg-green-100 text-green-800' :
                        analysisResult.security.score >= 60 ? 'bg-yellow-100 text-yellow-800' :
                        'bg-red-100 text-red-800'
                      }`}>
                        {analysisResult.security.score >= 80 ? 'Secure' :
                         analysisResult.security.score >= 60 ? 'Moderate' : 'At Risk'}
                      </div>
                    </div>
                  </Card>
                )}

                {/* Issues */}
                {analysisResult.issues.length > 0 && (
                  <div>
                    <h4 className="font-medium mb-2">Issues Found</h4>
                    {analysisResult.issues.map(renderAnalysisIssue)}
                  </div>
                )}

                {/* Suggestions */}
                {analysisResult.suggestions.length > 0 && (
                  <div>
                    <h4 className="font-medium mb-2">Suggestions</h4>
                    <ul className="space-y-1">
                      {analysisResult.suggestions.map((suggestion, index) => (
                        <li key={index} className="text-sm text-gray-700 flex items-start gap-2">
                          <Lightbulb className="w-4 h-4 text-amber-500 flex-shrink-0 mt-0.5" />
                          {suggestion}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center text-gray-500 py-8">
                <AlertTriangle className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p>No analysis available</p>
                <p className="text-xs">Click "Analyze" to check your code</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="optimize" className="flex-1 overflow-y-auto p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Code Optimization</h3>
              <Button onClick={optimizeCode} size="sm" disabled={isProcessing}>
                {isProcessing ? 'Optimizing...' : 'Find Optimizations'}
              </Button>
            </div>

            {optimizationResult ? (
              <div className="space-y-4">
                {/* Summary */}
                <Card className="p-3">
                  <div className="text-xs text-gray-600 mb-1">Optimization Summary</div>
                  <div className="text-sm font-medium">{optimizationResult.impact}</div>
                  <div className="text-xs text-gray-600">Effort: {optimizationResult.effort}</div>
                  <div className="text-xs text-gray-600">Priority: {optimizationResult.priority}</div>
                </Card>

                {/* Optimizations */}
                {optimizationResult.optimizations.length > 0 ? (
                  <div>
                    <h4 className="font-medium mb-2">Available Optimizations</h4>
                    {optimizationResult.optimizations.map(renderOptimization)}
                  </div>
                ) : (
                  <div className="text-center text-gray-500 py-4">
                    <CheckCircle className="w-8 h-8 mx-auto mb-2 text-green-500" />
                    <p>No optimizations needed</p>
                    <p className="text-xs">Your code looks great!</p>
                  </div>
                )}
              </div>
            ) : (
              <div className="text-center text-gray-500 py-8">
                <Zap className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p>No optimization analysis available</p>
                <p className="text-xs">Click "Find Optimizations" to get suggestions</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="suggestions" className="flex-1 overflow-y-auto p-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="font-medium">Code Suggestions</h3>
              <Button onClick={getCompletions} size="sm" disabled={isProcessing}>
                {isProcessing ? 'Loading...' : 'Get Suggestions'}
              </Button>
            </div>

            {suggestions.length > 0 ? (
              <div className="space-y-2">
                {suggestions.map((suggestion, index) => (
                  <div key={index} className="p-3 border rounded-lg hover:bg-gray-50 cursor-pointer"
                       onClick={() => handleSuggestionApply(suggestion)}>
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-sm">{suggestion.label}</span>
                      <Badge variant="outline">{suggestion.kind}</Badge>
                    </div>
                    <p className="text-xs text-gray-600 mb-2">{suggestion.description}</p>
                    <code className="text-xs bg-gray-100 p-1 rounded">{suggestion.insertText}</code>
                  </div>
                ))}
              </div>
            ) : (
              <div className="text-center text-gray-500 py-8">
                <Code2 className="w-12 h-12 mx-auto mb-3 text-gray-300" />
                <p>No suggestions available</p>
                <p className="text-xs">Position your cursor and click "Get Suggestions"</p>
              </div>
            )}
          </TabsContent>

          <TabsContent value="settings" className="p-4">
            <h3 className="font-medium mb-4">AI Assistant Settings</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <label className="text-sm">Code Completion</label>
                <input
                  type="checkbox"
                  checked={aiConfig.enableCompletion}
                  onChange={(e) => setAiConfig({...aiConfig, enableCompletion: e.target.checked})}
                />
              </div>
              <div className="flex items-center justify-between">
                <label className="text-sm">Code Analysis</label>
                <input
                  type="checkbox"
                  checked={aiConfig.enableAnalysis}
                  onChange={(e) => setAiConfig({...aiConfig, enableAnalysis: e.target.checked})}
                />
              </div>
              <div className="flex items-center justify-between">
                <label className="text-sm">Code Optimization</label>
                <input
                  type="checkbox"
                  checked={aiConfig.enableOptimization}
                  onChange={(e) => setAiConfig({...aiConfig, enableOptimization: e.target.checked})}
                />
              </div>
              <div className="flex items-center justify-between">
                <label className="text-sm">Natural Language Processing</label>
                <input
                  type="checkbox"
                  checked={aiConfig.enableNLP}
                  onChange={(e) => setAiConfig({...aiConfig, enableNLP: e.target.checked})}
                />
              </div>
            </div>
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
};