// src/components/metacontract-editor/AIEnhancedMetaContractEditor.tsx
import React, { useState, useEffect, useRef } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { 
  Play, 
  Save, 
  Share, 
  Download, 
  Settings, 
  FileText, 
  AlertCircle, 
  CheckCircle,
  Clock,
  Users,
  Edit3,
  Code,
  Sun,
  Moon,
  Brain,
  Sparkles
} from 'lucide-react';
import { AIEnabledMonacoEditor } from './ai/AIEnabledMonacoEditor';
import { IntelligentAssistant } from './ai/IntelligentAssistant';
import { CompilationResults } from './CompilationResults';
import { VisualEditor } from './VisualEditor';
import { useMetaContractEditor } from '../hooks/useMetaContractEditor';
import { useWebSocket } from '../hooks/useWebSocket';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { toast } from 'sonner';

interface EditorProps {
  projectId?: string;
  initialContent?: string;
  readonly?: boolean;
  aiEnabled?: boolean;
}

export const AIEnhancedMetaContractEditor: React.FC<EditorProps> = ({
  projectId,
  initialContent = '',
  readonly = false,
  aiEnabled = true
}) => {
  const [content, setContent] = useState(initialContent);
  const [isCompiling, setIsCompiling] = useState(false);
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [collaborators, setCollaborators] = useState<string[]>([]);
  const [editorMode, setEditorMode] = useState<'visual' | 'code'>('code');
  const [theme, setTheme] = useState<'light' | 'dark'>('light');
  const [cursorPosition, setCursorPosition] = useState({ line: 1, column: 1 });
  const [showAIAssistant, setShowAIAssistant] = useState(aiEnabled);
  const [aiProcessing, setAiProcessing] = useState(false);
  
  const editorRef = useRef<any>(null);

  const {
    project,
    compileProject,
    saveProject,
    isLoading,
    error
  } = useMetaContractEditor(projectId);

  const { 
    isConnected, 
    sendMessage, 
    lastMessage,
    connectionStatus 
  } = useWebSocket(projectId);

  // Handle real-time compilation results
  useEffect(() => {
    if (lastMessage?.type === 'compile_response') {
      setIsCompiling(false);
      if (lastMessage.data.success) {
        toast.success('Compilation successful!');
      } else {
        toast.error('Compilation failed. Check the results below.');
      }
    }
  }, [lastMessage]);

  // Auto-save functionality
  useEffect(() => {
    if (!readonly && content !== initialContent) {
      const timer = setTimeout(() => {
        handleSave();
      }, 2000); // Auto-save after 2 seconds of no changes

      return () => clearTimeout(timer);
    }
  }, [content, readonly, initialContent]);

  const handleContentChange = (newContent: string) => {
    setContent(newContent);
    
    // Send content change to other collaborators
    if (isConnected && projectId) {
      sendMessage({
        type: 'content_change',
        projectId,
        data: {
          content: newContent,
          timestamp: new Date()
        }
      });
    }
  };

  const handleCursorPositionChange = (position: {line: number; column: number}) => {
    setCursorPosition(position);
  };

  const handleCompile = async () => {
    if (isCompiling) return;
    
    setIsCompiling(true);
    
    try {
      if (projectId) {
        // Compile saved project
        await compileProject(projectId);
      } else {
        // Compile current content (preview mode)
        sendMessage({
          type: 'compile_request',
          projectId: projectId || 'preview',
          data: {
            content,
            preview: !projectId
          }
        });
      }
    } catch (error) {
      setIsCompiling(false);
      toast.error('Failed to start compilation');
    }
  };

  const handleSave = async () => {
    if (readonly || !projectId) return;
    
    try {
      await saveProject(projectId, { content });
      setLastSaved(new Date());
      toast.success('Project saved successfully');
    } catch (error) {
      toast.error('Failed to save project');
    }
  };

  const handleExport = () => {
    const blob = new Blob([content], { type: 'text/yaml' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${project?.name || 'metacontract'}.yaml`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const handleInsertCode = (code: string) => {
    if (editorRef.current && editorRef.current.insertText) {
      editorRef.current.insertText(code);
    } else {
      // Fallback: insert at cursor position
      const lines = content.split('\n');
      const beforeCursor = lines.slice(0, cursorPosition.line - 1);
      const currentLine = lines[cursorPosition.line - 1] || '';
      const afterCursor = lines.slice(cursorPosition.line);
      
      const newCurrentLine = 
        currentLine.slice(0, cursorPosition.column - 1) + 
        code + 
        currentLine.slice(cursorPosition.column - 1);
      
      const newContent = [...beforeCursor, newCurrentLine, ...afterCursor].join('\n');
      handleContentChange(newContent);
    }
  };

  const handleReplaceCode = (oldCode: string, newCode: string) => {
    const newContent = content.replace(oldCode, newCode);
    handleContentChange(newContent);
  };

  const getStatusBadge = () => {
    if (isCompiling) {
      return <Badge variant="outline"><Clock className="w-3 h-3 mr-1" />Compiling</Badge>;
    }
    
    if (project?.status === 'valid') {
      return <Badge variant="default"><CheckCircle className="w-3 h-3 mr-1" />Valid</Badge>;
    }
    
    if (project?.status === 'error') {
      return <Badge variant="destructive"><AlertCircle className="w-3 h-3 mr-1" />Error</Badge>;
    }
    
    return <Badge variant="secondary">Draft</Badge>;
  };

  const toggleAIAssistant = () => {
    setShowAIAssistant(!showAIAssistant);
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <div className="bg-white border-b px-6 py-4 flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-xl font-semibold text-gray-900 flex items-center gap-2">
            {project?.name || 'Meta-Contract Editor'}
            {aiEnabled && <Brain className="w-5 h-5 text-blue-500" />}
          </h1>
          {getStatusBadge()}
          {collaborators.length > 0 && (
            <div className="flex items-center space-x-1">
              <Users className="w-4 h-4 text-gray-500" />
              <span className="text-sm text-gray-500">{collaborators.length} online</span>
            </div>
          )}
          {aiProcessing && (
            <div className="flex items-center space-x-1 text-amber-600">
              <Sparkles className="w-4 h-4 animate-spin" />
              <span className="text-sm">AI Processing...</span>
            </div>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          <Button
            variant="outline"
            size="sm"
            onClick={handleCompile}
            disabled={isCompiling || readonly}
          >
            <Play className="w-4 h-4 mr-1" />
            {isCompiling ? 'Compiling...' : 'Compile'}
          </Button>
          
          {!readonly && (
            <Button
              variant="outline"
              size="sm"
              onClick={handleSave}
              disabled={isLoading}
            >
              <Save className="w-4 h-4 mr-1" />
              Save
            </Button>
          )}
          
          <Button
            variant="outline"
            size="sm"
            onClick={handleExport}
          >
            <Download className="w-4 h-4 mr-1" />
            Export
          </Button>

          {aiEnabled && (
            <Button
              variant={showAIAssistant ? "default" : "outline"}
              size="sm"
              onClick={toggleAIAssistant}
            >
              <Brain className="w-4 h-4 mr-1" />
              AI Assistant
            </Button>
          )}
          
          <Button 
            variant="outline" 
            size="sm"
            onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
          >
            {theme === 'light' ? <Moon className="w-4 h-4 mr-1" /> : <Sun className="w-4 h-4 mr-1" />}
            {theme === 'light' ? 'Dark' : 'Light'}
          </Button>
          
          <Button variant="outline" size="sm">
            <Settings className="w-4 h-4 mr-1" />
            Settings
          </Button>
        </div>
      </div>

      {/* Status Bar */}
      <div className="bg-gray-100 px-6 py-2 flex items-center justify-between text-sm text-gray-600">
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-1">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            <span>{isConnected ? 'Connected' : 'Disconnected'}</span>
          </div>
          {lastSaved && (
            <span>Last saved: {lastSaved.toLocaleTimeString()}</span>
          )}
          {aiEnabled && (
            <div className="flex items-center space-x-1">
              <Brain className="w-3 h-3 text-blue-500" />
              <span>AI Enhanced</span>
            </div>
          )}
        </div>
        
        <div className="flex items-center space-x-4">
          <span>YAML</span>
          <span>Line {cursorPosition.line}, Column {cursorPosition.column}</span>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col">
        {/* Editor Mode Tabs */}
        <div className="bg-white border-b px-6 py-2">
          <Tabs value={editorMode} onValueChange={(value: any) => setEditorMode(value)}>
            <TabsList className="grid w-full max-w-md grid-cols-2">
              <TabsTrigger value="visual" className="flex items-center space-x-2">
                <Edit3 className="w-4 h-4" />
                <span>Visual Editor</span>
              </TabsTrigger>
              <TabsTrigger value="code" className="flex items-center space-x-2">
                <Code className="w-4 h-4" />
                <span>Code Editor</span>
                {aiEnabled && <Brain className="w-3 h-3 text-blue-500" />}
              </TabsTrigger>
            </TabsList>
          </Tabs>
        </div>

        <div className="flex-1 flex">
          {editorMode === 'visual' ? (
            <VisualEditor
              content={content}
              onChange={handleContentChange}
              readonly={readonly}
              theme={theme}
            />
          ) : (
            <>
              {/* Code Editor Panel */}
              <div className={`flex-1 flex flex-col ${showAIAssistant ? 'w-2/3' : 'w-full'}`}>
                <div className="flex-1">
                  <AIEnabledMonacoEditor
                    value={content}
                    onChange={handleContentChange}
                    language="yaml"
                    theme={theme === 'dark' ? 'vs-dark' : 'vs-light'}
                    onCursorPositionChange={handleCursorPositionChange}
                    aiEnabled={aiEnabled}
                    options={{
                      readOnly: readonly,
                      minimap: { enabled: true },
                      lineNumbers: 'on',
                      wordWrap: 'on',
                      automaticLayout: true,
                      scrollBeyondLastLine: false,
                      fontSize: 14,
                      tabSize: 2,
                      insertSpaces: true,
                      detectIndentation: false
                    }}
                  />
                </div>
              </div>

              {/* AI Assistant Panel */}
              {showAIAssistant && aiEnabled && (
                <div className="w-1/3 border-l bg-white">
                  <IntelligentAssistant
                    content={content}
                    cursorPosition={cursorPosition}
                    onInsertCode={handleInsertCode}
                    onReplaceCode={handleReplaceCode}
                    projectId={projectId}
                    className="h-full"
                  />
                </div>
              )}

              {/* Results Panel - shown when AI assistant is hidden */}
              {!showAIAssistant && (
                <div className="w-1/3 border-l bg-white">
                  <CompilationResults 
                    isCompiling={isCompiling}
                    results={lastMessage?.data}
                    onErrorClick={(line, column) => {
                      if (editorRef.current) {
                        editorRef.current.revealLineInCenter(line);
                        editorRef.current.setPosition({ lineNumber: line, column });
                        editorRef.current.focus();
                      }
                    }}
                  />
                </div>
              )}
            </>
          )}
        </div>
      </div>

      {/* AI Status Bar */}
      {aiEnabled && (
        <div className="bg-blue-50 border-t px-6 py-2 flex items-center justify-between text-sm">
          <div className="flex items-center space-x-4 text-blue-700">
            <div className="flex items-center space-x-1">
              <Brain className="w-4 h-4" />
              <span>Local AI Assistant Active</span>
            </div>
            <span className="text-blue-500">•</span>
            <span>Code completion, analysis, and optimization enabled</span>
          </div>
          <div className="text-blue-600 text-xs">
            Press Ctrl+Alt+A for analysis, Ctrl+Alt+G for generation
          </div>
        </div>
      )}
    </div>
  );
};