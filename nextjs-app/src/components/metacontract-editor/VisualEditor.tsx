import React, { useState, useCallback, useMemo } from 'react';
import { DndContext, DragEndEvent, DragOverlay, DragStartEvent } from '@dnd-kit/core';
import { arrayMove, SortableContext } from '@dnd-kit/sortable';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Database, 
  Key, 
  Shield, 
  GitBranch, 
  Search, 
  Code, 
  Eye,
  Plus,
  Settings,
  Trash2,
  Edit3,
  Layers,
  BookTemplate,
  Sparkles
} from 'lucide-react';
import { ComponentPalette } from './visual/ComponentPalette';
import { DropZone } from './visual/DropZone';
import { PropertyPanel } from './visual/PropertyPanel';
import { PreviewPanel } from './visual/PreviewPanel';
import { ThemeProvider } from './visual/ThemeProvider';
import { IntelligentAssistant } from './visual/IntelligentAssistant';
import { ERDiagram } from './visual/ERDiagram';
import { EnhancedERDiagram } from './visual/EnhancedERDiagram';
import { MultiPanelPreview } from './visual/MultiPanelPreview';
import { TemplateManager } from './template/TemplateManager';
import { IntelligentTemplateApplicationEngine } from '@/lib/template-application';
import { useHotkeys } from 'react-hotkeys-hook';
import * as yaml from 'js-yaml';
import { toast } from 'sonner';

export interface MetaContractElement {
  id: string;
  type: 'field' | 'relationship' | 'security' | 'validation' | 'index' | 'trigger';
  name: string;
  properties: Record<string, any>;
  position?: { x: number; y: number };
  children?: string[];
}

export interface MetaContractSchema {
  specification_version: string;
  api_id: string;
  namespace: string;
  resource_name: string;
  data_structure: {
    primary_key: string;
    data_classification: string;
    fields: any[];
  };
  relationships?: any[];
  security_model?: any;
  validation_rules?: any[];
  business_logic?: any;
}

interface VisualEditorProps {
  content: string;
  onChange: (content: string) => void;
  readonly?: boolean;
  theme?: 'light' | 'dark';
}

export const VisualEditor: React.FC<VisualEditorProps> = ({
  content,
  onChange,
  readonly = false,
  theme = 'light'
}) => {
  const [elements, setElements] = useState<MetaContractElement[]>([]);
  const [selectedElement, setSelectedElement] = useState<MetaContractElement | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [viewMode, setViewMode] = useState<'design' | 'code' | 'preview' | 'diagram' | 'enhanced-diagram' | 'multi-preview'>('design');
  const [draggedElement, setDraggedElement] = useState<MetaContractElement | null>(null);
  const [undoStack, setUndoStack] = useState<MetaContractElement[][]>([]);
  const [redoStack, setRedoStack] = useState<MetaContractElement[][]>([]);
  
  // 模板系统状态
  const [templateManagerOpen, setTemplateManagerOpen] = useState(false);
  const [templateApplicationEngine] = useState(() => new IntelligentTemplateApplicationEngine());

  // Parse YAML content to elements
  const parseContent = useCallback((yamlContent: string) => {
    try {
      const schema = yaml.load(yamlContent) as MetaContractSchema;
      const newElements: MetaContractElement[] = [];

      // Parse fields
      if (schema.data_structure?.fields) {
        schema.data_structure.fields.forEach((field: any, index: number) => {
          newElements.push({
            id: `field-${field.name}`,
            type: 'field',
            name: field.name,
            properties: field,
            position: { x: 50, y: 50 + index * 80 }
          });
        });
      }

      // Parse relationships
      if (schema.relationships) {
        schema.relationships.forEach((rel: any, index: number) => {
          newElements.push({
            id: `relationship-${rel.name || index}`,
            type: 'relationship',
            name: rel.name || `Relationship ${index + 1}`,
            properties: rel,
            position: { x: 300, y: 50 + index * 80 }
          });
        });
      }

      // Parse security model
      if (schema.security_model) {
        newElements.push({
          id: 'security-model',
          type: 'security',
          name: 'Security Model',
          properties: schema.security_model,
          position: { x: 550, y: 50 }
        });
      }

      setElements(newElements);
    } catch (error) {
      console.error('Failed to parse YAML content:', error);
    }
  }, []);

  // Generate YAML from elements
  const generateContent = useCallback((elements: MetaContractElement[]) => {
    try {
      const schema: MetaContractSchema = {
        specification_version: "1.0",
        api_id: "generated-api",
        namespace: "default",
        resource_name: "resource",
        data_structure: {
          primary_key: "id",
          data_classification: "internal",
          fields: elements
            .filter(el => el.type === 'field')
            .map(el => el.properties)
        }
      };

      const relationships = elements.filter(el => el.type === 'relationship');
      if (relationships.length > 0) {
        schema.relationships = relationships.map(el => el.properties);
      }

      const securityElements = elements.filter(el => el.type === 'security');
      if (securityElements.length > 0) {
        schema.security_model = securityElements[0].properties;
      }

      return yaml.dump(schema, { indent: 2 });
    } catch (error) {
      console.error('Failed to generate YAML:', error);
      return content;
    }
  }, [content]);

  // Sync with parent content
  React.useEffect(() => {
    parseContent(content);
  }, [content, parseContent]);

  // Undo/Redo functionality
  const pushToUndoStack = useCallback(() => {
    setUndoStack(prev => [...prev.slice(-9), [...elements]]);
    setRedoStack([]);
  }, [elements]);

  const undo = useCallback(() => {
    if (undoStack.length > 0) {
      const previousState = undoStack[undoStack.length - 1];
      setRedoStack(prev => [elements, ...prev.slice(0, 9)]);
      setElements(previousState);
      setUndoStack(prev => prev.slice(0, -1));
    }
  }, [undoStack, elements]);

  const redo = useCallback(() => {
    if (redoStack.length > 0) {
      const nextState = redoStack[0];
      setUndoStack(prev => [...prev.slice(-9), elements]);
      setElements(nextState);
      setRedoStack(prev => prev.slice(1));
    }
  }, [redoStack, elements]);

  // Hotkeys
  useHotkeys('ctrl+z', undo, { enabled: !readonly });
  useHotkeys('ctrl+y', redo, { enabled: !readonly });
  useHotkeys('delete', () => {
    if (selectedElement && !readonly) {
      handleDeleteElement(selectedElement.id);
    }
  }, { enabled: !readonly });

  // Element operations
  const handleAddElement = useCallback((elementType: MetaContractElement['type'], properties: any) => {
    if (readonly) return;

    pushToUndoStack();
    const newElement: MetaContractElement = {
      id: `${elementType}-${Date.now()}`,
      type: elementType,
      name: properties.name || `New ${elementType}`,
      properties,
      position: { x: 100, y: 100 }
    };

    setElements(prev => [...prev, newElement]);
    setSelectedElement(newElement);
    
    const newContent = generateContent([...elements, newElement]);
    onChange(newContent);
  }, [readonly, pushToUndoStack, elements, generateContent, onChange]);

  const handleUpdateElement = useCallback((elementId: string, updates: Partial<MetaContractElement>) => {
    if (readonly) return;

    pushToUndoStack();
    setElements(prev => prev.map(el => 
      el.id === elementId ? { ...el, ...updates } : el
    ));

    const updatedElements = elements.map(el => 
      el.id === elementId ? { ...el, ...updates } : el
    );
    const newContent = generateContent(updatedElements);
    onChange(newContent);
  }, [readonly, pushToUndoStack, elements, generateContent, onChange]);

  const handleDeleteElement = useCallback((elementId: string) => {
    if (readonly) return;

    pushToUndoStack();
    const updatedElements = elements.filter(el => el.id !== elementId);
    setElements(updatedElements);
    
    if (selectedElement?.id === elementId) {
      setSelectedElement(null);
    }

    const newContent = generateContent(updatedElements);
    onChange(newContent);
  }, [readonly, pushToUndoStack, elements, selectedElement, generateContent, onChange]);

  // Drag and drop
  const handleDragStart = useCallback((event: DragStartEvent) => {
    const element = elements.find(el => el.id === event.active.id);
    setDraggedElement(element || null);
  }, [elements]);

  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event;
    setDraggedElement(null);

    if (!over || active.id === over.id) return;

    const activeElement = elements.find(el => el.id === active.id);
    const overElement = elements.find(el => el.id === over.id);

    if (activeElement && overElement) {
      const activeIndex = elements.indexOf(activeElement);
      const overIndex = elements.indexOf(overElement);
      
      pushToUndoStack();
      const newElements = arrayMove(elements, activeIndex, overIndex);
      setElements(newElements);
      
      const newContent = generateContent(newElements);
      onChange(newContent);
    }
  }, [elements, pushToUndoStack, generateContent, onChange]);

  // Filter elements based on search
  const filteredElements = useMemo(() => {
    if (!searchTerm) return elements;
    return elements.filter(el => 
      el.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      el.type.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [elements, searchTerm]);

  // 处理模板应用
  const handleApplyTemplate = useCallback(async (template: any) => {
    if (readonly) return;

    try {
      pushToUndoStack();
      
      // 应用模板
      const result = await templateApplicationEngine.applyTemplate(
        template,
        elements,
        {
          generateBackup: true,
          validateAfterApply: true,
          mergeStrategy: 'additive',
          preserveExisting: true
        }
      );

      if (result.success) {
        // 合并现有元素和模板元素
        const mergedElements = [...elements];
        
        // 添加新元素
        result.appliedElements.forEach(templateElement => {
          const existingIndex = mergedElements.findIndex(e => e.id === templateElement.id);
          if (existingIndex >= 0) {
            // 更新现有元素
            mergedElements[existingIndex] = templateElement;
          } else {
            // 添加新元素
            mergedElements.push(templateElement);
          }
        });

        setElements(mergedElements);
        const newContent = generateContent(mergedElements);
        onChange(newContent);

        // 显示成功消息
        let message = `Applied template "${template.name}" successfully`;
        if (result.warnings && result.warnings.length > 0) {
          message += ` with ${result.warnings.length} warnings`;
        }
        toast.success(message);

        // 显示警告
        if (result.warnings && result.warnings.length > 0) {
          result.warnings.forEach(warning => {
            toast.warning(warning, { duration: 5000 });
          });
        }

        // 显示冲突信息
        if (result.conflicts && result.conflicts.length > 0) {
          toast.info(`Resolved ${result.conflicts.length} conflicts automatically`, {
            duration: 5000
          });
        }

      } else {
        toast.error(`Failed to apply template: ${result.error || 'Unknown error'}`);
      }

    } catch (error) {
      console.error('Template application error:', error);
      toast.error('Failed to apply template. Please try again.');
    }
  }, [readonly, pushToUndoStack, templateApplicationEngine, elements, generateContent, onChange]);

  return (
    <ThemeProvider theme={theme}>
      <DndContext onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
        <div className="h-screen flex flex-col bg-background">
          {/* Toolbar */}
          <div className="bg-card border-b px-4 py-2 flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <Tabs value={viewMode} onValueChange={(value: any) => setViewMode(value)}>
                <TabsList className="grid w-full grid-cols-6">
                  <TabsTrigger value="design" className="flex items-center space-x-1">
                    <Edit3 className="w-4 h-4" />
                    <span>Design</span>
                  </TabsTrigger>
                  <TabsTrigger value="code" className="flex items-center space-x-1">
                    <Code className="w-4 h-4" />
                    <span>Code</span>
                  </TabsTrigger>
                  <TabsTrigger value="preview" className="flex items-center space-x-1">
                    <Eye className="w-4 h-4" />
                    <span>Preview</span>
                  </TabsTrigger>
                  <TabsTrigger value="diagram" className="flex items-center space-x-1">
                    <Database className="w-4 h-4" />
                    <span>ER Diagram</span>
                  </TabsTrigger>
                  <TabsTrigger value="enhanced-diagram" className="flex items-center space-x-1">
                    <Database className="w-4 h-4" />
                    <span>Enhanced ER</span>
                  </TabsTrigger>
                  <TabsTrigger value="multi-preview" className="flex items-center space-x-1">
                    <Layers className="w-4 h-4" />
                    <span>Multi Preview</span>
                  </TabsTrigger>
                </TabsList>
              </Tabs>
            </div>

            <div className="flex items-center space-x-2">
              <div className="relative">
                <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder="Search elements..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-8 w-64"
                />
              </div>
              
              {!readonly && (
                <>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setTemplateManagerOpen(true)}
                    className="flex items-center space-x-1"
                  >
                    <BookTemplate className="w-4 h-4" />
                    <span>Templates</span>
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={undo}
                    disabled={undoStack.length === 0}
                  >
                    Undo
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={redo}
                    disabled={redoStack.length === 0}
                  >
                    Redo
                  </Button>
                </>
              )}
            </div>
          </div>

          {/* Main Content */}
          <div className="flex-1 flex">
            {/* Component Palette */}
            {(viewMode === 'design' || viewMode === 'diagram' || viewMode === 'enhanced-diagram') && !readonly && (
              <div className="w-64 border-r bg-card">
                {viewMode === 'design' ? (
                  <ComponentPalette onAddElement={handleAddElement} />
                ) : (
                  <IntelligentAssistant
                    content={content}
                    elements={filteredElements}
                    onAddElement={handleAddElement}
                    onUpdateElement={handleUpdateElement}
                  />
                )}
              </div>
            )}

            {/* Central Canvas */}
            <div className="flex-1 flex flex-col">
              <Tabs value={viewMode} className="flex-1">
                <TabsContent value="design" className="flex-1 m-0">
                  <DropZone
                    elements={filteredElements}
                    selectedElement={selectedElement}
                    onSelectElement={setSelectedElement}
                    onUpdateElement={handleUpdateElement}
                    onDeleteElement={handleDeleteElement}
                    readonly={readonly}
                  />
                </TabsContent>
                
                <TabsContent value="code" className="flex-1 m-0">
                  <div className="h-full p-4">
                    <pre className="bg-muted p-4 rounded-lg overflow-auto h-full">
                      <code>{content}</code>
                    </pre>
                  </div>
                </TabsContent>
                
                <TabsContent value="diagram" className="flex-1 m-0">
                  <ERDiagram
                    elements={filteredElements}
                    onElementSelect={(elementId) => {
                      const element = elements.find(el => el.id === elementId);
                      if (element) {
                        setSelectedElement(element);
                      }
                    }}
                  />
                </TabsContent>
                
                <TabsContent value="enhanced-diagram" className="flex-1 m-0">
                  <EnhancedERDiagram
                    elements={filteredElements}
                    onElementSelect={(elementId) => {
                      const element = elements.find(el => el.id === elementId);
                      if (element) {
                        setSelectedElement(element);
                      }
                    }}
                    onElementUpdate={handleUpdateElement}
                  />
                </TabsContent>
                
                <TabsContent value="multi-preview" className="flex-1 m-0">
                  <MultiPanelPreview
                    content={content}
                    elements={elements}
                    onElementSelect={(elementId) => {
                      const element = elements.find(el => el.id === elementId);
                      if (element) {
                        setSelectedElement(element);
                      }
                    }}
                  />
                </TabsContent>
                
                <TabsContent value="preview" className="flex-1 m-0">
                  <PreviewPanel content={content} elements={elements} />
                </TabsContent>
              </Tabs>
            </div>

            {/* Property Panel */}
            {(viewMode === 'design' || viewMode === 'diagram' || viewMode === 'enhanced-diagram') && selectedElement && (
              <div className="w-80 border-l bg-card">
                <PropertyPanel
                  element={selectedElement}
                  onUpdateElement={(updates) => 
                    handleUpdateElement(selectedElement.id, updates)
                  }
                  onDeleteElement={() => handleDeleteElement(selectedElement.id)}
                  readonly={readonly}
                />
              </div>
            )}
          </div>

          {/* Status Bar */}
          <div className="bg-card border-t px-4 py-2 flex items-center justify-between text-sm">
            <div className="flex items-center space-x-4">
              <Badge variant="outline">
                {elements.length} elements
              </Badge>
              {selectedElement && (
                <Badge variant="secondary">
                  {selectedElement.name} selected
                </Badge>
              )}
            </div>
            <div className="flex items-center space-x-2">
              <span className="text-muted-foreground">
                Visual Editor v1.0
              </span>
            </div>
          </div>

          {/* Drag Overlay */}
          <DragOverlay>
            {draggedElement ? (
              <Card className="w-48 opacity-50">
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm">{draggedElement.name}</CardTitle>
                </CardHeader>
                <CardContent className="pt-0">
                  <Badge variant="outline">{draggedElement.type}</Badge>
                </CardContent>
              </Card>
            ) : null}
          </DragOverlay>
        </div>
      </DndContext>
      
      {/* Template Manager Dialog */}
      <TemplateManager
        existingElements={elements}
        onApplyTemplate={handleApplyTemplate}
        onClose={() => setTemplateManagerOpen(false)}
        open={templateManagerOpen}
      />
    </ThemeProvider>
  );
};