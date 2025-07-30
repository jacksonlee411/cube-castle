// src/components/metacontract-editor/ai/AIEnabledMonacoEditor.tsx
import React, { useRef, useEffect, useState } from 'react';
import * as monaco from 'monaco-editor';

interface AIEnabledMonacoEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  theme?: string;
  options?: monaco.editor.IStandaloneEditorConstructionOptions;
  onCursorPositionChange?: (position: {line: number; column: number}) => void;
  aiEnabled?: boolean;
  className?: string;
}

interface AICompletionItem extends monaco.languages.CompletionItem {
  priority?: number;
  aiGenerated?: boolean;
}

export const AIEnabledMonacoEditor: React.FC<AIEnabledMonacoEditorProps> = ({
  value,
  onChange,
  language = 'yaml',
  theme = 'vs-light',
  options = {},
  onCursorPositionChange,
  aiEnabled = true,
  className
}) => {
  const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [isInitialized, setIsInitialized] = useState(false);
  const aiCompletionProviderRef = useRef<monaco.IDisposable | null>(null);

  // Initialize Monaco Editor
  useEffect(() => {
    if (!containerRef.current || isInitialized) return;

    // Configure YAML language features
    configureYAMLLanguage();

    // Create editor
    const editor = monaco.editor.create(containerRef.current, {
      value,
      language,
      theme,
      automaticLayout: true,
      minimap: { enabled: true },
      lineNumbers: 'on',
      wordWrap: 'on',
      scrollBeyondLastLine: false,
      fontSize: 14,
      tabSize: 2,
      insertSpaces: true,
      detectIndentation: false,
      suggestOnTriggerCharacters: true,
      acceptSuggestionOnEnter: 'on',
      acceptSuggestionOnCommitCharacter: true,
      quickSuggestions: {
        other: true,
        comments: false,
        strings: false
      },
      ...options
    });

    editorRef.current = editor;

    // Set up event listeners
    editor.onDidChangeModelContent(() => {
      const currentValue = editor.getValue();
      onChange(currentValue);
    });

    editor.onDidChangeCursorPosition((e) => {
      if (onCursorPositionChange) {
        onCursorPositionChange({
          line: e.position.lineNumber,
          column: e.position.column
        });
      }
    });

    // Register AI completion provider if enabled
    if (aiEnabled) {
      registerAICompletionProvider(editor);
    }

    // Register custom commands
    registerCustomCommands(editor);

    setIsInitialized(true);

    return () => {
      if (aiCompletionProviderRef.current) {
        aiCompletionProviderRef.current.dispose();
      }
      editor.dispose();
    };
  }, []);

  // Update editor value when prop changes
  useEffect(() => {
    if (editorRef.current && value !== editorRef.current.getValue()) {
      editorRef.current.setValue(value);
    }
  }, [value]);

  // Update editor theme
  useEffect(() => {
    if (editorRef.current) {
      monaco.editor.setTheme(theme);
    }
  }, [theme]);

  const configureYAMLLanguage = () => {
    // Register YAML language if not already registered
    if (!monaco.languages.getLanguages().find(lang => lang.id === 'yaml')) {
      monaco.languages.register({ id: 'yaml' });
    }

    // Configure YAML tokenization
    monaco.languages.setMonarchTokensProvider('yaml', {
      tokenizer: {
        root: [
          // Comments
          [/#.*$/, 'comment'],
          
          // Keys
          [/^(\s*)([a-zA-Z_][\w\-]*)\s*:/, ['', 'key']],
          [/(\s+)([a-zA-Z_][\w\-]*)\s*:/, ['', 'key']],
          
          // Values
          [/:\s*(.+)$/, 'value'],
          
          // Arrays
          [/^\s*-\s*/, 'array-marker'],
          
          // Strings
          [/"([^"\\]|\\.)*$/, 'string.invalid'],
          [/'([^'\\]|\\.)*$/, 'string.invalid'],
          [/"/, 'string', '@doubleQuotedString'],
          [/'/, 'string', '@singleQuotedString'],
          
          // Numbers
          [/\b\d+\.?\d*\b/, 'number'],
          
          // Booleans
          [/\b(true|false)\b/, 'boolean'],
          
          // Null
          [/\bnull\b/, 'null'],
          
          // Whitespace
          [/\s+/, '']
        ],
        
        doubleQuotedString: [
          [/[^\\"]+/, 'string'],
          [/\\./, 'string.escape'],
          [/"/, 'string', '@pop']
        ],
        
        singleQuotedString: [
          [/[^\\']+/, 'string'],
          [/\\./, 'string.escape'],
          [/'/, 'string', '@pop']
        ]
      }
    });

    // Configure YAML theme colors
    monaco.editor.defineTheme('yaml-theme', {
      base: 'vs',
      inherit: true,
      rules: [
        { token: 'key', foreground: '0066cc', fontStyle: 'bold' },
        { token: 'value', foreground: '008800' },
        { token: 'comment', foreground: '999999', fontStyle: 'italic' },
        { token: 'array-marker', foreground: 'ff6600', fontStyle: 'bold' },
        { token: 'string', foreground: 'dd0000' },
        { token: 'number', foreground: '0066ff' },
        { token: 'boolean', foreground: 'ff6600' },
        { token: 'null', foreground: '999999' }
      ],
      colors: {}
    });
  };

  const registerAICompletionProvider = (editor: monaco.editor.IStandaloneCodeEditor) => {
    aiCompletionProviderRef.current = monaco.languages.registerCompletionItemProvider('yaml', {
      triggerCharacters: [' ', ':', '-', '\n'],
      
      provideCompletionItems: async (model, position, context, token) => {
        const textUntilPosition = model.getValueInRange({
          startLineNumber: 1,
          startColumn: 1,
          endLineNumber: position.lineNumber,
          endColumn: position.column,
        });

        try {
          // Get AI completions
          const aiSuggestions = await getAICompletions(textUntilPosition, position);
          
          // Get built-in completions
          const builtInSuggestions = getBuiltInCompletions(textUntilPosition, position);
          
          // Combine and prioritize suggestions
          const allSuggestions = [...aiSuggestions, ...builtInSuggestions];
          
          return {
            suggestions: allSuggestions.sort((a, b) => (b.priority || 0) - (a.priority || 0))
          };
        } catch (error) {
          console.error('AI completion error:', error);
          
          // Fallback to built-in completions
          return {
            suggestions: getBuiltInCompletions(textUntilPosition, position)
          };
        }
      }
    });
  };

  const getAICompletions = async (
    textUntilPosition: string, 
    position: monaco.Position
  ): Promise<AICompletionItem[]> => {
    try {
      const response = await fetch('/api/ai/complete', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          type: 'completion',
          context: editorRef.current?.getValue() || '',
          query: textUntilPosition,
          position: {
            line: position.lineNumber,
            column: position.column
          }
        })
      });

      if (!response.ok) {
        throw new Error('AI completion request failed');
      }

      const aiResponse = await response.json();
      
      return (aiResponse.suggestions || []).map((suggestion: any) => ({
        label: suggestion.label,
        kind: getMonacoCompletionItemKind(suggestion.kind),
        insertText: suggestion.insertText,
        detail: suggestion.detail,
        documentation: suggestion.description,
        priority: suggestion.priority,
        aiGenerated: true,
        range: {
          startLineNumber: position.lineNumber,
          endLineNumber: position.lineNumber,
          startColumn: position.column,
          endColumn: position.column,
        }
      }));
    } catch (error) {
      console.error('Failed to get AI completions:', error);
      return [];
    }
  };

  const getBuiltInCompletions = (
    textUntilPosition: string, 
    position: monaco.Position
  ): monaco.languages.CompletionItem[] => {
    const suggestions: monaco.languages.CompletionItem[] = [];
    const currentLine = textUntilPosition.split('\n').pop() || '';
    const trimmedLine = currentLine.trim();
    
    // Get indentation level
    const indent = currentLine.length - currentLine.trimLeft().length;
    const indentStr = ' '.repeat(indent);

    // Top-level keywords
    if (indent === 0) {
      const topLevelKeywords = [
        'entities', 'relations', 'constraints', 'indexes', 
        'templates', 'version', 'metadata'
      ];
      
      topLevelKeywords.forEach(keyword => {
        if (keyword.startsWith(trimmedLine) || trimmedLine === '') {
          suggestions.push({
            label: keyword,
            kind: monaco.languages.CompletionItemKind.Keyword,
            insertText: `${keyword}:`,
            detail: 'Top-level section',
            documentation: `Define ${keyword} section`,
            range: {
              startLineNumber: position.lineNumber,
              endLineNumber: position.lineNumber,
              startColumn: position.column - trimmedLine.length,
              endColumn: position.column,
            }
          });
        }
      });
    }

    // Entity-level keywords
    if (indent === 2 || (indent === 0 && trimmedLine.startsWith('-'))) {
      const entityKeywords = [
        'name', 'fields', 'relations', 'constraints', 'indexes', 'annotations'
      ];
      
      entityKeywords.forEach(keyword => {
        if (keyword.startsWith(trimmedLine.replace(/^-\s*/, ''))) {
          suggestions.push({
            label: keyword,
            kind: monaco.languages.CompletionItemKind.Property,
            insertText: `${keyword}:`,
            detail: 'Entity property',
            documentation: `Define entity ${keyword}`,
            range: {
              startLineNumber: position.lineNumber,
              endLineNumber: position.lineNumber,
              startColumn: position.column - trimmedLine.replace(/^-\s*/, '').length,
              endColumn: position.column,
            }
          });
        }
      });
    }

    // Field properties
    if (indent >= 4) {
      const fieldProperties = [
        'type', 'required', 'unique', 'default', 'validation', 'annotations'
      ];
      
      fieldProperties.forEach(prop => {
        if (prop.startsWith(trimmedLine.replace(/^-\s*/, ''))) {
          suggestions.push({
            label: prop,
            kind: monaco.languages.CompletionItemKind.Property,
            insertText: `${prop}: `,
            detail: 'Field property',
            documentation: `Field ${prop} property`,
            range: {
              startLineNumber: position.lineNumber,
              endLineNumber: position.lineNumber,
              startColumn: position.column - trimmedLine.replace(/^-\s*/, '').length,
              endColumn: position.column,
            }
          });
        }
      });
    }

    // Field types
    if (currentLine.includes('type:')) {
      const fieldTypes = [
        'string', 'int', 'int64', 'float64', 'bool', 'time', 
        'uuid', 'text', 'json', 'enum'
      ];
      
      fieldTypes.forEach(type => {
        suggestions.push({
          label: type,
          kind: monaco.languages.CompletionItemKind.Enum,
          insertText: type,
          detail: 'Field type',
          documentation: `${type} field type`,
          range: {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: position.column,
            endColumn: position.column,
          }
        });
      });
    }

    // Boolean values
    if (currentLine.includes('required:') || currentLine.includes('unique:')) {
      ['true', 'false'].forEach(bool => {
        suggestions.push({
          label: bool,
          kind: monaco.languages.CompletionItemKind.Value,
          insertText: bool,
          detail: 'Boolean value',
          range: {
            startLineNumber: position.lineNumber,
            endLineNumber: position.lineNumber,
            startColumn: position.column,
            endColumn: position.column,
          }
        });
      });
    }

    return suggestions;
  };

  const getMonacoCompletionItemKind = (kind: string): monaco.languages.CompletionItemKind => {
    switch (kind) {
      case 'keyword': return monaco.languages.CompletionItemKind.Keyword;
      case 'property': return monaco.languages.CompletionItemKind.Property;
      case 'field': return monaco.languages.CompletionItemKind.Field;
      case 'type': return monaco.languages.CompletionItemKind.TypeParameter;
      case 'value': return monaco.languages.CompletionItemKind.Value;
      case 'template': return monaco.languages.CompletionItemKind.Snippet;
      case 'enum': return monaco.languages.CompletionItemKind.Enum;
      default: return monaco.languages.CompletionItemKind.Text;
    }
  };

  const registerCustomCommands = (editor: monaco.editor.IStandaloneCodeEditor) => {
    // AI-powered format command
    editor.addCommand(monaco.KeyMod.Shift | monaco.KeyMod.Alt | monaco.KeyCode.KeyF, () => {
      formatWithAI();
    });

    // AI analysis command
    editor.addCommand(monaco.KeyMod.Ctrl | monaco.KeyMod.Alt | monaco.KeyCode.KeyA, () => {
      analyzeWithAI();
    });

    // AI generation command
    editor.addCommand(monaco.KeyMod.Ctrl | monaco.KeyMod.Alt | monaco.KeyCode.KeyG, () => {
      generateWithAI();
    });
  };

  const formatWithAI = async () => {
    if (!editorRef.current) return;

    try {
      const response = await fetch('/api/ai/optimize', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          type: 'optimization',
          context: editorRef.current.getValue(),
          query: 'Format and improve code structure'
        })
      });

      if (response.ok) {
        const result = await response.json();
        // Apply formatting optimizations automatically
        if (result.optimization?.optimizations) {
          result.optimization.optimizations.forEach((opt: any) => {
            if (opt.type === 'maintainability' && opt.risk === 'low') {
              const model = editorRef.current!.getModel();
              if (model) {
                const range = model.findMatches(opt.before, false, false, false, null, false)[0];
                if (range) {
                  editorRef.current!.executeEdits('ai-format', [{
                    range: range.range,
                    text: opt.after
                  }]);
                }
              }
            }
          });
        }
      }
    } catch (error) {
      console.error('AI formatting failed:', error);
    }
  };

  const analyzeWithAI = async () => {
    if (!editorRef.current) return;

    try {
      const response = await fetch('/api/ai/analyze', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          type: 'analysis',
          context: editorRef.current.getValue(),
          query: 'Analyze code for issues and improvements'
        })
      });

      if (response.ok) {
        const result = await response.json();
        if (result.analysis?.issues) {
          // Add markers for issues
          const markers = result.analysis.issues.map((issue: any) => ({
            severity: getMonacoMarkerSeverity(issue.severity),
            startLineNumber: issue.line,
            startColumn: issue.column,
            endLineNumber: issue.line,
            endColumn: issue.column + 10,
            message: issue.message,
            source: 'AI Analysis'
          }));
          
          const model = editorRef.current!.getModel();
          if (model) {
            monaco.editor.setModelMarkers(model, 'ai-analysis', markers);
          }
        }
      }
    } catch (error) {
      console.error('AI analysis failed:', error);
    }
  };

  const generateWithAI = () => {
    // This would open a dialog or panel for natural language generation
    // For now, just focus the editor
    editorRef.current?.focus();
  };

  const getMonacoMarkerSeverity = (severity: string): monaco.MarkerSeverity => {
    switch (severity) {
      case 'error': return monaco.MarkerSeverity.Error;
      case 'warning': return monaco.MarkerSeverity.Warning;
      case 'info': return monaco.MarkerSeverity.Info;
      default: return monaco.MarkerSeverity.Hint;
    }
  };

  // Expose editor instance to parent components
  useEffect(() => {
    if (editorRef.current && onCursorPositionChange) {
      const position = editorRef.current.getPosition();
      if (position) {
        onCursorPositionChange({
          line: position.lineNumber,
          column: position.column
        });
      }
    }
  }, [isInitialized]);

  return (
    <div className={`w-full h-full ${className}`}>
      <div ref={containerRef} className="w-full h-full" />
      {aiEnabled && (
        <div className="absolute bottom-2 right-2 text-xs text-gray-500 bg-white/80 px-2 py-1 rounded shadow">
          AI Enhanced
        </div>
      )}
    </div>
  );
};