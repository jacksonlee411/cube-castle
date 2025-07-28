import React, { forwardRef, useEffect, useRef, useImperativeHandle } from 'react';
import dynamic from 'next/dynamic';

// Dynamically import Monaco Editor to avoid SSR issues
const monaco = typeof window !== 'undefined' ? require('monaco-editor') : null;

interface MonacoEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  theme?: string;
  options?: monaco.editor.IStandaloneEditorConstructionOptions;
  height?: string;
  width?: string;
}

export interface MonacoEditorRef {
  getValue: () => string;
  setValue: (value: string) => void;
  focus: () => void;
  setPosition: (position: monaco.IPosition) => void;
  revealLineInCenter: (lineNumber: number) => void;
  getEditor: () => monaco.editor.IStandaloneCodeEditor | null;
}

export const MonacoEditor = forwardRef<MonacoEditorRef, MonacoEditorProps>(({
  value,
  onChange,
  language = 'yaml',
  theme = 'vs-dark',
  options = {},
  height = '100%',
  width = '100%'
}, ref) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const editorRef = useRef<any>(null);
  const subscriptionRef = useRef<any>(null);
  const [isMonacoLoaded, setIsMonacoLoaded] = React.useState(false);

  useImperativeHandle(ref, () => ({
    getValue: () => editorRef.current?.getValue() || '',
    setValue: (newValue: string) => editorRef.current?.setValue(newValue),
    focus: () => editorRef.current?.focus(),
    setPosition: (position: any) => editorRef.current?.setPosition(position),
    revealLineInCenter: (lineNumber: number) => editorRef.current?.revealLineInCenter(lineNumber),
    getEditor: () => editorRef.current
  }));

  useEffect(() => {
    if (typeof window === 'undefined' || !containerRef.current) return;

    // Dynamically load Monaco Editor
    const loadMonaco = async () => {
      try {
        const monacoModule = await import('monaco-editor');
        
        // Configure YAML language features
        monacoModule.languages.setLanguageConfiguration('yaml', {
          comments: {
            lineComment: '#'
          },
          brackets: [
            ['{', '}'],
            ['[', ']'],
            ['(', ')']
          ],
          autoClosingPairs: [
            { open: '{', close: '}' },
            { open: '[', close: ']' },
            { open: '(', close: ')' },
            { open: '"', close: '"' },
            { open: "'", close: "'" }
          ],
          surroundingPairs: [
            { open: '{', close: '}' },
            { open: '[', close: ']' },
            { open: '(', close: ')' },
            { open: '"', close: '"' },
            { open: "'", close: "'" }
          ],
          folding: {
            offSide: true
          }
        });

        // Configure Meta-Contract specific completion items
        monacoModule.languages.registerCompletionItemProvider('yaml', {
          provideCompletionItems: (model, position) => {
            const textUntilPosition = model.getValueInRange({
              startLineNumber: 1,
              startColumn: 1,
              endLineNumber: position.lineNumber,
              endColumn: position.column,
            });

            const suggestions: any[] = [];

            // Top-level properties
            if (textUntilPosition.trim() === '' || /^[a-zA-Z_]*$/.test(textUntilPosition.split('\n').pop() || '')) {
              suggestions.push(
                {
                  label: 'specification_version',
                  kind: monacoModule.languages.CompletionItemKind.Property,
                  insertText: 'specification_version: "1.0"',
                  documentation: 'Version of the meta-contract specification'
                },
                {
                  label: 'api_id',
                  kind: monacoModule.languages.CompletionItemKind.Property,
                  insertText: 'api_id: "${1:uuid}"',
                  insertTextRules: monacoModule.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                  documentation: 'Unique identifier for this API'
                },
                {
                  label: 'namespace',
                  kind: monacoModule.languages.CompletionItemKind.Property,
                  insertText: 'namespace: "${1:namespace}"',
                  insertTextRules: monacoModule.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                  documentation: 'Namespace for the resource'
                },
                {
                  label: 'resource_name',
                  kind: monacoModule.languages.CompletionItemKind.Property,
                  insertText: 'resource_name: "${1:resource}"',
                  insertTextRules: monacoModule.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                  documentation: 'Name of the resource'
                },
                {
                  label: 'data_structure',
                  kind: monacoModule.languages.CompletionItemKind.Struct,
                  insertText: [
                    'data_structure:',
                    '  primary_key: "${1:id}"',
                    '  data_classification: "${2:internal}"',
                    '  fields:',
                    '    - name: "${3:id}"',
                    '      type: "${4:uuid}"',
                    '      required: true',
                    '      unique: true'
                  ].join('\n'),
                  insertTextRules: monacoModule.languages.CompletionItemInsertTextRule.InsertAsSnippet,
                  documentation: 'Data structure definition'
                }
              );
            }

            return { suggestions };
          }
        });

        // Create editor
        const editor = monacoModule.editor.create(containerRef.current!, {
          value,
          language,
          theme,
          automaticLayout: true,
          ...options
        });

        editorRef.current = editor;
        setIsMonacoLoaded(true);

        // Listen for content changes
        subscriptionRef.current = editor.onDidChangeModelContent(() => {
          const currentValue = editor.getValue();
          if (currentValue !== value) {
            onChange(currentValue);
          }
        });

        return () => {
          subscriptionRef.current?.dispose();
          editor.dispose();
        };
      } catch (error) {
        console.error('Failed to load Monaco Editor:', error);
      }
    };

    loadMonaco();
  }, []); // Only run once

  // Update editor value when prop changes
  useEffect(() => {
    if (isMonacoLoaded && editorRef.current && editorRef.current.getValue() !== value) {
      editorRef.current.setValue(value);
    }
  }, [value, isMonacoLoaded]);

  // Update editor theme
  useEffect(() => {
    if (isMonacoLoaded && editorRef.current) {
      const monacoModule = require('monaco-editor');
      monacoModule.editor.setTheme(theme);
    }
  }, [theme, isMonacoLoaded]);

  if (!isMonacoLoaded) {
    return (
      <div 
        style={{ height, width, display: 'flex', alignItems: 'center', justifyContent: 'center' }}
        className="monaco-editor-container bg-gray-100 text-gray-500"
      >
        Loading editor...
      </div>
    );
  }

  return (
    <div 
      ref={containerRef} 
      style={{ height, width }}
      className="monaco-editor-container"
    />
  );
});