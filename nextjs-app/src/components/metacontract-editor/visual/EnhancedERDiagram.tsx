import React, { useState, useMemo, useCallback, useRef, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Database, 
  ZoomIn, 
  ZoomOut, 
  RotateCcw, 
  Download, 
  Search, 
  Settings, 
  Eye, 
  GitBranch, 
  Key, 
  Shield, 
  AlertTriangle, 
  Info,
  Move,
  Maximize2,
  Minimize2,
  MousePointer,
  Grid,
  Layers,
  Filter,
  RefreshCw
} from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';

interface EnhancedERDiagramProps {
  elements: MetaContractElement[];
  onElementSelect?: (elementId: string) => void;
  onElementUpdate?: (elementId: string, updates: Partial<MetaContractElement>) => void;
  className?: string;
}

interface DiagramNode {
  id: string;
  type: 'entity' | 'relationship';
  name: string;
  x: number;
  y: number;
  width: number;
  height: number;
  properties: any;
  connected: string[];
}

interface DiagramEdge {
  id: string;
  from: string;
  to: string;
  type: 'one-to-one' | 'one-to-many' | 'many-to-many';
  label?: string;
}

interface FieldInspectorData {
  element: MetaContractElement;
  relationships: MetaContractElement[];
  usageCount: number;
  dependencies: string[];
  suggestions: string[];
}

export const EnhancedERDiagram: React.FC<EnhancedERDiagramProps> = ({
  elements,
  onElementSelect,
  onElementUpdate,
  className = ''
}) => {
  const svgRef = useRef<SVGSVGElement>(null);
  const [zoom, setZoom] = useState(1);
  const [pan, setPan] = useState({ x: 0, y: 0 });
  const [isDragging, setIsDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [showRelationships, setShowRelationships] = useState(true);
  const [showFieldDetails, setShowFieldDetails] = useState(true);
  const [layoutMode, setLayoutMode] = useState<'auto' | 'manual'>('auto');
  const [inspectorData, setInspectorData] = useState<FieldInspectorData | null>(null);

  // Convert elements to diagram nodes
  const diagramNodes = useMemo((): DiagramNode[] => {
    const entities = elements.filter(el => el.type === 'field');
    const groupedByEntity = entities.reduce((acc, field) => {
      const entityName = field.properties.entity || 'DefaultEntity';
      if (!acc[entityName]) {
        acc[entityName] = [];
      }
      acc[entityName].push(field);
      return acc;
    }, {} as Record<string, MetaContractElement[]>);

    return Object.entries(groupedByEntity).map(([entityName, fields], index) => {
      const nodeWidth = 200;
      const nodeHeight = Math.max(120, 40 + fields.length * 25);
      const x = (index % 3) * 300 + 50;
      const y = Math.floor(index / 3) * 250 + 50;

      return {
        id: `entity-${entityName}`,
        type: 'entity' as const,
        name: entityName,
        x,
        y,
        width: nodeWidth,
        height: nodeHeight,
        properties: { fields },
        connected: []
      };
    });
  }, [elements]);

  // Generate diagram edges from relationships
  const diagramEdges = useMemo((): DiagramEdge[] => {
    const relationships = elements.filter(el => el.type === 'relationship');
    return relationships.map(rel => ({
      id: rel.id,
      from: `entity-${rel.properties.source_entity || 'DefaultEntity'}`,
      to: `entity-${rel.properties.target_entity || rel.properties.target_resource}`,
      type: rel.properties.cardinality || 'one-to-many',
      label: rel.name
    }));
  }, [elements]);

  // Field inspector logic
  const analyzeField = useCallback((element: MetaContractElement): FieldInspectorData => {
    const relationships = elements.filter(el => 
      el.type === 'relationship' && 
      (el.properties.source_field === element.properties.name || 
       el.properties.target_field === element.properties.name)
    );

    const usageCount = relationships.length;
    const dependencies = relationships.map(rel => rel.name);
    
    const suggestions = [];
    
    // Generate suggestions based on field properties
    if (element.properties.type === 'string' && !element.properties.max_length) {
      suggestions.push('Consider adding max_length constraint');
    }
    
    if (!element.properties.indexed && relationships.length > 0) {
      suggestions.push('Consider adding index for better query performance');
    }
    
    if (!element.properties.validation && element.properties.type === 'string') {
      suggestions.push('Consider adding validation rules');
    }
    
    if (element.properties.nullable && !element.properties.default_value) {
      suggestions.push('Consider adding default value for nullable field');
    }

    return {
      element,
      relationships,
      usageCount,
      dependencies,
      suggestions
    };
  }, [elements]);

  // Handle zoom
  const handleZoom = useCallback((direction: 'in' | 'out' | 'reset') => {
    setZoom(prev => {
      if (direction === 'reset') return 1;
      if (direction === 'in') return Math.min(prev * 1.2, 3);
      return Math.max(prev / 1.2, 0.1);
    });
  }, []);

  // Handle mouse events for panning
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button === 0) { // Left click
      setIsDragging(true);
      setDragStart({ x: e.clientX - pan.x, y: e.clientY - pan.y });
    }
  }, [pan]);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (isDragging) {
      setPan({
        x: e.clientX - dragStart.x,
        y: e.clientY - dragStart.y,
      });
    }
  }, [isDragging, dragStart]);

  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
  }, []);

  // Handle node selection
  const handleNodeClick = useCallback((nodeId: string, element?: MetaContractElement) => {
    setSelectedNodeId(nodeId);
    if (element) {
      setInspectorData(analyzeField(element));
      onElementSelect?.(element.id);
    }
  }, [analyzeField, onElementSelect]);

  // Auto-layout algorithm
  const applyAutoLayout = useCallback(() => {
    // Simple force-directed layout
    const updatedNodes = [...diagramNodes];
    const iterations = 50;
    const k = Math.sqrt((800 * 600) / updatedNodes.length);
    
    for (let i = 0; i < iterations; i++) {
      // Calculate forces
      updatedNodes.forEach(node => {
        let fx = 0, fy = 0;
        
        // Repulsive forces
        updatedNodes.forEach(other => {
          if (node.id !== other.id) {
            const dx = node.x - other.x;
            const dy = node.y - other.y;
            const distance = Math.sqrt(dx * dx + dy * dy);
            if (distance > 0) {
              const force = (k * k) / distance;
              fx += (dx / distance) * force;
              fy += (dy / distance) * force;
            }
          }
        });
        
        // Attractive forces from edges
        diagramEdges.forEach(edge => {
          if (edge.from === node.id || edge.to === node.id) {
            const other = updatedNodes.find(n => 
              n.id === (edge.from === node.id ? edge.to : edge.from)
            );
            if (other) {
              const dx = other.x - node.x;
              const dy = other.y - node.y;
              const distance = Math.sqrt(dx * dx + dy * dy);
              if (distance > 0) {
                const force = (distance * distance) / k;
                fx += (dx / distance) * force * 0.1;
                fy += (dy / distance) * force * 0.1;
              }
            }
          }
        });
        
        // Apply forces with cooling
        const cooling = 1 - (i / iterations);
        node.x += fx * cooling * 0.1;
        node.y += fy * cooling * 0.1;
        
        // Keep nodes within bounds
        node.x = Math.max(50, Math.min(750, node.x));
        node.y = Math.max(50, Math.min(450, node.y));
      });
    }
    
    // This would trigger a re-render in a real implementation
    console.log('Auto-layout applied');
  }, [diagramNodes, diagramEdges]);

  // Download diagram as SVG
  const handleDownload = useCallback(() => {
    if (svgRef.current) {
      const svgData = new XMLSerializer().serializeToString(svgRef.current);
      const blob = new Blob([svgData], { type: 'image/svg+xml' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'er-diagram.svg';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  }, []);

  // Filter nodes based on search
  const filteredNodes = useMemo(() => {
    if (!searchTerm) return diagramNodes;
    return diagramNodes.filter(node => 
      node.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      node.properties.fields?.some((field: any) => 
        field.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        field.properties.type.toLowerCase().includes(searchTerm.toLowerCase())
      )
    );
  }, [diagramNodes, searchTerm]);

  const renderEntityNode = (node: DiagramNode) => {
    const isSelected = selectedNodeId === node.id;
    const fields = node.properties.fields || [];
    
    return (
      <g key={node.id} transform={`translate(${node.x}, ${node.y})`}>
        {/* Entity box */}
        <rect
          width={node.width}
          height={node.height}
          fill={isSelected ? '#e3f2fd' : '#ffffff'}
          stroke={isSelected ? '#2196f3' : '#e0e0e0'}
          strokeWidth={isSelected ? 2 : 1}
          rx={8}
          className="cursor-pointer hover:fill-gray-50"
          onClick={() => handleNodeClick(node.id)}
        />
        
        {/* Entity title */}
        <rect
          width={node.width}
          height={30}
          fill={isSelected ? '#2196f3' : '#f5f5f5'}
          stroke="none"
          rx={8}
        />
        <rect
          width={node.width}
          height={15}
          y={15}
          fill={isSelected ? '#2196f3' : '#f5f5f5'}
          stroke="none"
        />
        
        <text
          x={node.width / 2}
          y={20}
          textAnchor="middle"
          className={`text-sm font-semibold ${isSelected ? 'fill-white' : 'fill-gray-800'}`}
        >
          {node.name}
        </text>
        
        {/* Fields */}
        {showFieldDetails && fields.map((field: any, index: number) => {
          const y = 45 + index * 25;
          const isPrimaryKey = field.properties.primary_key;
          const isRequired = field.properties.required;
          const isUnique = field.properties.unique;
          
          return (
            <g key={field.id}>
              {/* Field background */}
              <rect
                x={2}
                y={y - 12}
                width={node.width - 4}
                height={24}
                fill="transparent"
                className="cursor-pointer hover:fill-blue-50"
                onClick={(e) => {
                  e.stopPropagation();
                  handleNodeClick(node.id, field);
                }}
              />
              
              {/* Primary key icon */}
              {isPrimaryKey && (
                <g transform={`translate(8, ${y - 6})`}>
                  <Key className="w-3 h-3 text-yellow-600" />
                </g>
              )}
              
              {/* Field name */}
              <text
                x={isPrimaryKey ? 25 : 8}
                y={y}
                className={`text-xs ${isPrimaryKey ? 'font-semibold' : 'font-medium'} fill-gray-800`}
              >
                {field.properties.name}
              </text>
              
              {/* Field type */}
              <text
                x={node.width - 8}
                y={y}
                textAnchor="end"
                className="text-xs fill-gray-600"
              >
                {field.properties.type}
                {field.properties.max_length && `(${field.properties.max_length})`}
              </text>
              
              {/* Field constraints indicators */}
              <g transform={`translate(${node.width - 60}, ${y - 8})`}>
                {isRequired && (
                  <rect width={12} height={12} fill="#f44336" rx={2} className="text-xs">
                    <title>Required</title>
                  </rect>
                )}
                {isUnique && (
                  <rect x={15} width={12} height={12} fill="#ff9800" rx={2} className="text-xs">
                    <title>Unique</title>
                  </rect>
                )}
              </g>
            </g>
          );
        })}
      </g>
    );
  };

  const renderRelationshipEdge = (edge: DiagramEdge) => {
    const fromNode = filteredNodes.find(n => n.id === edge.from);
    const toNode = filteredNodes.find(n => n.id === edge.to);
    
    if (!fromNode || !toNode) return null;
    
    const fromX = fromNode.x + fromNode.width / 2;
    const fromY = fromNode.y + fromNode.height / 2;
    const toX = toNode.x + toNode.width / 2;
    const toY = toNode.y + toNode.height / 2;
    
    // Calculate connection points on node edges
    const dx = toX - fromX;
    const dy = toY - fromY;
    const angle = Math.atan2(dy, dx);
    
    const fromConnX = fromX + Math.cos(angle) * (fromNode.width / 2 + 10);
    const fromConnY = fromY + Math.sin(angle) * (fromNode.height / 2 + 10);
    const toConnX = toX - Math.cos(angle) * (toNode.width / 2 + 10);
    const toConnY = toY - Math.sin(angle) * (toNode.height / 2 + 10);
    
    return (
      <g key={edge.id}>
        {/* Connection line */}
        <line
          x1={fromConnX}
          y1={fromConnY}
          x2={toConnX}
          y2={toConnY}
          stroke="#666"
          strokeWidth={2}
          markerEnd="url(#arrowhead)"
        />
        
        {/* Relationship label */}
        {edge.label && (
          <text
            x={(fromConnX + toConnX) / 2}
            y={(fromConnY + toConnY) / 2 - 5}
            textAnchor="middle"
            className="text-xs fill-gray-600 bg-white"
          >
            {edge.label}
          </text>
        )}
        
        {/* Cardinality indicators */}
        <text
          x={fromConnX + 10}
          y={fromConnY - 5}
          className="text-xs fill-gray-500"
        >
          1
        </text>
        <text
          x={toConnX - 15}
          y={toConnY - 5}
          className="text-xs fill-gray-500"
        >
          {edge.type.includes('many') ? 'N' : '1'}
        </text>
      </g>
    );
  };

  const renderFieldInspector = () => {
    if (!inspectorData) return null;
    
    const { element, relationships, usageCount, dependencies, suggestions } = inspectorData;
    
    return (
      <Card className="mt-4">
        <CardHeader className="pb-2">
          <CardTitle className="text-sm flex items-center">
            <Eye className="w-4 h-4 mr-2" />
            Field Inspector: {element.properties.name}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {/* Basic Properties */}
          <div className="grid grid-cols-2 gap-2 text-xs">
            <div>
              <span className="text-muted-foreground">Type:</span>
              <Badge variant="outline" className="ml-2">{element.properties.type}</Badge>
            </div>
            <div>
              <span className="text-muted-foreground">Usage:</span>
              <Badge variant="secondary" className="ml-2">{usageCount} refs</Badge>
            </div>
          </div>
          
          {/* Constraints */}
          <div>
            <p className="text-xs text-muted-foreground mb-1">Constraints:</p>
            <div className="flex flex-wrap gap-1">
              {element.properties.primary_key && (
                <Badge variant="default" className="text-xs">Primary Key</Badge>
              )}
              {element.properties.required && (
                <Badge variant="destructive" className="text-xs">Required</Badge>
              )}
              {element.properties.unique && (
                <Badge variant="secondary" className="text-xs">Unique</Badge>
              )}
              {element.properties.indexed && (
                <Badge variant="outline" className="text-xs">Indexed</Badge>
              )}
            </div>
          </div>
          
          {/* Dependencies */}
          {dependencies.length > 0 && (
            <div>
              <p className="text-xs text-muted-foreground mb-1">Dependencies:</p>
              <div className="space-y-1">
                {dependencies.map((dep, index) => (
                  <div key={index} className="text-xs bg-muted p-1 rounded">
                    {dep}
                  </div>
                ))}
              </div>
            </div>
          )}
          
          {/* Suggestions */}
          {suggestions.length > 0 && (
            <div>
              <p className="text-xs text-muted-foreground mb-1">Suggestions:</p>
              <div className="space-y-1">
                {suggestions.map((suggestion, index) => (
                  <div key={index} className="flex items-start space-x-2 text-xs">
                    <Info className="w-3 h-3 text-blue-500 mt-0.5" />
                    <span>{suggestion}</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    );
  };

  return (
    <Card className={`h-full flex flex-col ${className}`}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center">
            <Database className="w-5 h-5 mr-2" />
            Enhanced ER Diagram
          </CardTitle>
          <div className="flex items-center space-x-1">
            <Button size="sm" variant="outline" onClick={() => handleZoom('in')}>
              <ZoomIn className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={() => handleZoom('out')}>
              <ZoomOut className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={() => handleZoom('reset')}>
              <RotateCcw className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={applyAutoLayout}>
              <RefreshCw className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={handleDownload}>
              <Download className="w-4 h-4" />
            </Button>
          </div>
        </div>
        
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2 text-xs text-muted-foreground">
            <Badge variant="outline" className="text-xs">
              {filteredNodes.length} entities
            </Badge>
            <Badge variant="outline" className="text-xs">
              {diagramEdges.length} relationships
            </Badge>
            <span>Zoom: {Math.round(zoom * 100)}%</span>
          </div>
          
          <div className="flex items-center space-x-2">
            <div className="relative">
              <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-3 h-3 text-muted-foreground" />
              <Input
                placeholder="Search entities..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-7 h-7 w-32 text-xs"
              />
            </div>
            <Button
              size="sm"
              variant={showRelationships ? "default" : "outline"}
              onClick={() => setShowRelationships(!showRelationships)}
            >
              <GitBranch className="w-3 h-3" />
            </Button>
            <Button
              size="sm"
              variant={showFieldDetails ? "default" : "outline"}
              onClick={() => setShowFieldDetails(!showFieldDetails)}
            >
              <Grid className="w-3 h-3" />
            </Button>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex-1 p-0 overflow-hidden">
        <div className="h-full flex">
          {/* Main diagram area */}
          <div className="flex-1 overflow-hidden relative">
            {filteredNodes.length === 0 ? (
              <div className="flex items-center justify-center h-full text-center">
                <div className="text-muted-foreground">
                  <Database className="w-12 h-12 mx-auto mb-4" />
                  <p className="text-sm">No entities to display</p>
                  <p className="text-xs">Add fields to see the ER diagram</p>
                </div>
              </div>
            ) : (
              <div 
                className="w-full h-full overflow-auto cursor-grab active:cursor-grabbing"
                onMouseDown={handleMouseDown}
                onMouseMove={handleMouseMove}
                onMouseUp={handleMouseUp}
              >
                <svg
                  ref={svgRef}
                  width="100%"
                  height="100%"
                  viewBox="0 0 1000 600"
                  className="bg-gray-50"
                  style={{
                    transform: `scale(${zoom}) translate(${pan.x}px, ${pan.y}px)`,
                    transformOrigin: '0 0'
                  }}
                >
                  {/* Arrowhead marker */}
                  <defs>
                    <marker
                      id="arrowhead"
                      markerWidth="10"
                      markerHeight="7"
                      refX="9"
                      refY="3.5"
                      orient="auto"
                    >
                      <polygon
                        points="0 0, 10 3.5, 0 7"
                        fill="#666"
                      />
                    </marker>
                  </defs>
                  
                  {/* Grid pattern */}
                  <defs>
                    <pattern
                      id="grid"
                      width="20"
                      height="20"
                      patternUnits="userSpaceOnUse"
                    >
                      <path
                        d="M 20 0 L 0 0 0 20"
                        fill="none"
                        stroke="#e0e0e0"
                        strokeWidth="1"
                      />
                    </pattern>
                  </defs>
                  <rect width="100%" height="100%" fill="url(#grid)" />
                  
                  {/* Relationship edges */}
                  {showRelationships && diagramEdges.map(renderRelationshipEdge)}
                  
                  {/* Entity nodes */}
                  {filteredNodes.map(renderEntityNode)}
                </svg>
              </div>
            )}
          </div>
          
          {/* Side panel for field inspector */}
          {inspectorData && (
            <div className="w-80 border-l p-4 overflow-auto">
              {renderFieldInspector()}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
};