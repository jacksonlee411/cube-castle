import React, { forwardRef, useImperativeHandle, useRef } from 'react';

interface SimpleTextEditorProps {
  value: string;
  onChange: (value: string) => void;
  language?: string;
  theme?: string;
  options?: any;
  height?: string;
  width?: string;
}

export interface SimpleTextEditorRef {
  getValue: () => string;
  setValue: (value: string) => void;
  focus: () => void;
  setPosition: (position: any) => void;
  revealLineInCenter: (lineNumber: number) => void;
  getEditor: () => HTMLTextAreaElement | null;
}

export const SimpleTextEditor = forwardRef<SimpleTextEditorRef, SimpleTextEditorProps>(({
  value,
  onChange,
  language = 'yaml',
  theme = 'vs-dark',
  options = {},
  height = '100%',
  width = '100%'
}, ref) => {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useImperativeHandle(ref, () => ({
    getValue: () => textareaRef.current?.value || '',
    setValue: (newValue: string) => {
      if (textareaRef.current) {
        textareaRef.current.value = newValue;
      }
    },
    focus: () => textareaRef.current?.focus(),
    setPosition: (position: any) => {
      // Simple implementation - just focus
      textareaRef.current?.focus();
    },
    revealLineInCenter: (lineNumber: number) => {
      // Simple implementation - just focus
      textareaRef.current?.focus();
    },
    getEditor: () => textareaRef.current
  }));

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    onChange(e.target.value);
  };

  return (
    <div style={{ height, width, display: 'flex', flexDirection: 'column' }}>
      <div style={{ 
        backgroundColor: theme === 'vs-dark' ? '#1e1e1e' : '#ffffff',
        color: theme === 'vs-dark' ? '#d4d4d4' : '#000000',
        padding: '8px 12px',
        fontSize: '12px',
        borderBottom: '1px solid #333'
      }}>
        YAML Editor ({language}) - Simple Mode
      </div>
      <textarea
        ref={textareaRef}
        value={value}
        onChange={handleChange}
        style={{
          flex: 1,
          width: '100%',
          padding: '12px',
          border: 'none',
          outline: 'none',
          resize: 'none',
          fontSize: '14px',
          fontFamily: 'Monaco, Consolas, "Courier New", monospace',
          lineHeight: '1.5',
          backgroundColor: theme === 'vs-dark' ? '#1e1e1e' : '#ffffff',
          color: theme === 'vs-dark' ? '#d4d4d4' : '#000000',
          ...options.style
        }}
        placeholder="Enter your YAML content here..."
        spellCheck={false}
      />
    </div>
  );
});

SimpleTextEditor.displayName = 'SimpleTextEditor';