import React, { useCallback, useState } from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { useDroppable } from '@dnd-kit/core';
import { CSS } from '@dnd-kit/utilities';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { 
  Database, 
  Key, 
  Shield, 
  GitBranch, 
  Hash,
  Calendar,
  FileText,
  Mail,
  Phone,
  CheckSquare,
  Link,
  Filter,
  Zap,
  AlertTriangle,
  MoreHorizontal,
  Edit,
  Trash2,
  Copy,
  Eye,
  EyeOff
} from 'lucide-react';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { MetaContractElement } from '../VisualEditor';

const ELEMENT_ICONS: Record<string, React.ComponentType<any>> = {
  // Field types
  string: FileText,
  email: Mail,
  phone: Phone,
  integer: Hash,
  decimal: Hash,
  number: Hash,
  date: Calendar,
  datetime: Calendar,
  boolean: CheckSquare,
  uuid: Key,
  json: Database,
  
  // Element types
  field: Database,
  relationship: GitBranch,
  security: Shield,
  validation: AlertTriangle,
  index: Zap,
  trigger: Zap
};

const ELEMENT_COLORS: Record<string, string> = {
  field: 'bg-blue-50 border-blue-200 hover:border-blue-300',
  relationship: 'bg-green-50 border-green-200 hover:border-green-300',
  security: 'bg-red-50 border-red-200 hover:border-red-300',
  validation: 'bg-yellow-50 border-yellow-200 hover:border-yellow-300',
  index: 'bg-purple-50 border-purple-200 hover:border-purple-300',
  trigger: 'bg-orange-50 border-orange-200 hover:border-orange-300'
};

interface ElementCardProps {
  element: MetaContractElement;
  isSelected: boolean;
  onSelect: () => void;
  onUpdate: (updates: Partial<MetaContractElement>) => void;
  onDelete: () => void;
  readonly?: boolean;
}

const ElementCard: React.FC<ElementCardProps> = ({
  element,
  isSelected,
  onSelect,
  onUpdate,
  onDelete,
  readonly = false
}) => {
  const [isVisible, setIsVisible] = useState(true);
  
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({ 
    id: element.id,
    disabled: readonly
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  const Icon = ELEMENT_ICONS[element.properties?.type] || ELEMENT_ICONS[element.type];
  const colorClass = ELEMENT_COLORS[element.type] || ELEMENT_COLORS.field;

  const handleDuplicate = useCallback(() => {
    const duplicated = {
      ...element,
      id: `${element.type}-${Date.now()}`,
      name: `${element.name} Copy`,
      properties: { ...element.properties, name: `${element.properties.name}_copy` }
    };
    // This would need to be passed down as a prop
    console.log('Duplicate element:', duplicated);
  }, [element]);

  const handleToggleVisibility = useCallback(() => {
    setIsVisible(!isVisible);
    onUpdate({ 
      properties: { 
        ...element.properties, 
        hidden: !isVisible 
      } 
    });
  }, [isVisible, onUpdate, element.properties]);

  return (
    <Card
      ref={setNodeRef}
      style={style}
      className={`
        cursor-pointer transition-all duration-200 select-none
        ${colorClass}
        ${isSelected ? 'ring-2 ring-primary ring-offset-2' : ''}
        ${isDragging ? 'shadow-lg rotate-2' : 'hover:shadow-md'}
        ${!isVisible ? 'opacity-50' : ''}
      `}
      onClick={onSelect}
      {...attributes}
      {...listeners}
    >
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <Icon className="w-4 h-4 text-foreground" />
            <CardTitle className="text-sm font-medium">
              {element.name}
            </CardTitle>
          </div>
          
          {!readonly && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button 
                  variant="ghost" 
                  size="sm" 
                  className="h-6 w-6 p-0 hover:bg-background/50"
                  onClick={(e) => e.stopPropagation()}
                >
                  <MoreHorizontal className="w-3 h-3" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end" className="w-48">
                <DropdownMenuItem onClick={(e) => {
                  e.stopPropagation();
                  onSelect();
                }}>
                  <Edit className="w-4 h-4 mr-2" />
                  Edit Properties
                </DropdownMenuItem>
                <DropdownMenuItem onClick={(e) => {
                  e.stopPropagation();
                  handleDuplicate();
                }}>
                  <Copy className="w-4 h-4 mr-2" />
                  Duplicate
                </DropdownMenuItem>
                <DropdownMenuItem onClick={(e) => {
                  e.stopPropagation();
                  handleToggleVisibility();
                }}>
                  {isVisible ? (
                    <>
                      <EyeOff className="w-4 h-4 mr-2" />
                      Hide
                    </>
                  ) : (
                    <>
                      <Eye className="w-4 h-4 mr-2" />
                      Show
                    </>
                  )}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem 
                  onClick={(e) => {
                    e.stopPropagation();
                    onDelete();
                  }}
                  className="text-destructive focus:text-destructive"
                >
                  <Trash2 className="w-4 h-4 mr-2" />
                  Delete
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </CardHeader>

      <CardContent className="pt-0 pb-3">
        <div className="space-y-2">
          <Badge variant="outline" className="text-xs">
            {element.type}
          </Badge>
          
          {/* Display key properties */}
          <div className="grid grid-cols-2 gap-2 text-xs">
            {element.properties?.type && (
              <div>
                <span className="text-muted-foreground">Type:</span>
                <span className="ml-1 font-mono">{element.properties.type}</span>
              </div>
            )}
            {element.properties?.required !== undefined && (
              <div>
                <span className="text-muted-foreground">Required:</span>
                <span className={`ml-1 ${element.properties.required ? 'text-red-600' : 'text-green-600'}`}>
                  {element.properties.required ? 'Yes' : 'No'}
                </span>
              </div>
            )}
            {element.properties?.unique && (
              <div className="col-span-2">
                <Badge variant="secondary" className="text-xs">
                  Unique
                </Badge>
              </div>
            )}
            {element.properties?.primary_key && (
              <div className="col-span-2">
                <Badge variant="default" className="text-xs">
                  Primary Key
                </Badge>
              </div>
            )}
          </div>

          {/* Validation indicators */}
          {element.properties?.validation && element.properties.validation.length > 0 && (
            <div className="flex flex-wrap gap-1">
              {element.properties.validation.map((validation: string, index: number) => (
                <Badge key={index} variant="outline" className="text-xs">
                  {validation}
                </Badge>
              ))}
            </div>
          )}

          {/* Relationship details */}
          {element.type === 'relationship' && element.properties?.target_resource && (
            <div className="text-xs">
              <span className="text-muted-foreground">Target:</span>
              <span className="ml-1 font-mono">{element.properties.target_resource}</span>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};

interface DropZoneProps {
  elements: MetaContractElement[];
  selectedElement: MetaContractElement | null;
  onSelectElement: (element: MetaContractElement | null) => void;
  onUpdateElement: (elementId: string, updates: Partial<MetaContractElement>) => void;
  onDeleteElement: (elementId: string) => void;
  readonly?: boolean;
}

export const DropZone: React.FC<DropZoneProps> = ({
  elements,
  selectedElement,
  onSelectElement,
  onUpdateElement,
  onDeleteElement,
  readonly = false
}) => {
  const { setNodeRef } = useDroppable({
    id: 'drop-zone',
  });

  const handleElementSelect = useCallback((element: MetaContractElement) => {
    onSelectElement(selectedElement?.id === element.id ? null : element);
  }, [selectedElement, onSelectElement]);

  const handleElementUpdate = useCallback((elementId: string, updates: Partial<MetaContractElement>) => {
    onUpdateElement(elementId, updates);
  }, [onUpdateElement]);

  return (
    <div 
      ref={setNodeRef}
      className="h-full p-6 bg-gradient-to-br from-background to-muted/20 overflow-auto"
    >
      <div className="max-w-7xl mx-auto">
        {elements.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-96 text-center">
            <div className="w-24 h-24 bg-muted rounded-full flex items-center justify-center mb-4">
              <Database className="w-12 h-12 text-muted-foreground" />
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">
              No Elements Yet
            </h3>
            <p className="text-muted-foreground mb-4 max-w-md">
              Start building your meta-contract by dragging components from the palette on the left.
            </p>
            {!readonly && (
              <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-blue-200 rounded border-2 border-dashed border-blue-400"></div>
                  <span>Fields</span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-green-200 rounded border-2 border-dashed border-green-400"></div>
                  <span>Relations</span>
                </div>
                <div className="flex items-center space-x-1">
                  <div className="w-4 h-4 bg-red-200 rounded border-2 border-dashed border-red-400"></div>
                  <span>Security</span>
                </div>
              </div>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
            {elements.map((element) => (
              <ElementCard
                key={element.id}
                element={element}
                isSelected={selectedElement?.id === element.id}
                onSelect={() => handleElementSelect(element)}
                onUpdate={(updates) => handleElementUpdate(element.id, updates)}
                onDelete={() => onDeleteElement(element.id)}
                readonly={readonly}
              />
            ))}
          </div>
        )}

        {/* Drop Zone Overlay */}
        {!readonly && (
          <div className="fixed inset-0 pointer-events-none">
            <div className="absolute inset-4 border-2 border-dashed border-primary/20 rounded-lg bg-primary/5 opacity-0 transition-opacity duration-200 peer-[.drop-target]:opacity-100">
              <div className="flex items-center justify-center h-full">
                <div className="text-center">
                  <Database className="w-12 h-12 text-primary mx-auto mb-2" />
                  <p className="text-primary font-medium">Drop component here</p>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};