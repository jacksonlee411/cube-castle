import React, { useState, useMemo, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Input } from '@/components/ui/input';
import { 
  Code2, 
  Database, 
  Server, 
  Globe, 
  FileText, 
  Eye, 
  Search, 
  Copy, 
  Download, 
  ChevronDown, 
  ChevronRight, 
  AlertCircle, 
  CheckCircle,
  Settings,
  Layers,
  Zap,
  ShieldCheck
} from 'lucide-react';
import { MetaContractElement, MetaContractSchema } from '../VisualEditor';
import * as yaml from 'js-yaml';

interface MultiPanelPreviewProps {
  content: string;
  elements: MetaContractElement[];
  onElementSelect?: (elementId: string) => void;
  className?: string;
}

interface ValidationIssue {
  id: string;
  type: 'error' | 'warning' | 'info';
  message: string;
  line?: number;
  column?: number;
  elementId?: string;
}

interface QualityMetrics {
  score: number;
  complexity: number;
  maintainability: number;
  performance: number;
  security: number;
}

export const MultiPanelPreview: React.FC<MultiPanelPreviewProps> = ({
  content,
  elements,
  onElementSelect,
  className = ''
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [activeFormat, setActiveFormat] = useState<'yaml' | 'json' | 'sql' | 'go' | 'graphql'>('yaml');
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set());
  const [selectedPreviewTab, setSelectedPreviewTab] = useState('code');

  // Parse schema
  const parsedSchema = useMemo(() => {
    try {
      return yaml.load(content) as MetaContractSchema;
    } catch {
      return null;
    }
  }, [content]);

  // Validation logic
  const validationIssues = useMemo((): ValidationIssue[] => {
    const issues: ValidationIssue[] = [];

    if (!parsedSchema) {
      issues.push({
        id: 'parse-error',
        type: 'error',
        message: 'Invalid YAML syntax',
        line: 1,
        column: 1
      });
      return issues;
    }

    // Check required fields
    if (!parsedSchema.api_id) {
      issues.push({
        id: 'missing-api-id',
        type: 'error',
        message: 'Missing required field: api_id'
      });
    }

    if (!parsedSchema.data_structure?.primary_key) {
      issues.push({
        id: 'missing-primary-key',
        type: 'warning',
        message: 'No primary key defined'
      });
    }

    // Check field validations
    elements.filter(el => el.type === 'field').forEach(field => {
      if (!field.properties.type) {
        issues.push({
          id: `field-no-type-${field.id}`,
          type: 'error',
          message: `Field '${field.name}' has no type specified`,
          elementId: field.id
        });
      }

      if (field.properties.type === 'string' && !field.properties.max_length) {
        issues.push({
          id: `field-no-length-${field.id}`,
          type: 'warning',
          message: `String field '${field.name}' should have max_length`,
          elementId: field.id
        });
      }
    });

    // Check relationships
    elements.filter(el => el.type === 'relationship').forEach(rel => {
      if (!rel.properties.target_resource) {
        issues.push({
          id: `rel-no-target-${rel.id}`,
          type: 'error',
          message: `Relationship '${rel.name}' has no target resource`,
          elementId: rel.id
        });
      }
    });

    return issues;
  }, [parsedSchema, elements]);

  // Quality metrics calculation
  const qualityMetrics = useMemo((): QualityMetrics => {
    const fieldCount = elements.filter(el => el.type === 'field').length;
    const relationshipCount = elements.filter(el => el.type === 'relationship').length;
    const securityCount = elements.filter(el => el.type === 'security').length;
    const validationCount = elements.filter(el => el.type === 'validation').length;
    
    const errorCount = validationIssues.filter(i => i.type === 'error').length;
    const warningCount = validationIssues.filter(i => i.type === 'warning').length;

    // Calculate metrics (0-100 scale)
    const complexity = Math.min(100, (fieldCount * 5) + (relationshipCount * 10) + (validationCount * 3));
    const maintainability = Math.max(0, 100 - (errorCount * 15) - (warningCount * 5));
    const performance = Math.max(20, 100 - (fieldCount * 2) - (relationshipCount * 5));
    const security = securityCount > 0 ? Math.min(100, 60 + (securityCount * 20)) : 30;
    
    const score = Math.round((maintainability + performance + security) / 3);

    return {
      score,
      complexity,
      maintainability,
      performance,
      security
    };
  }, [elements, validationIssues]);

  // Generate code in different formats
  const generateCode = useCallback((format: string): string => {
    if (!parsedSchema) return '';

    switch (format) {
      case 'json':
        return JSON.stringify(parsedSchema, null, 2);
      
      case 'sql':
        return generateSQLSchema(parsedSchema, elements);
      
      case 'go':
        return generateGoEntSchema(parsedSchema, elements);
      
      case 'graphql':
        return generateGraphQLSchema(parsedSchema, elements);
      
      default:
        return content; // YAML
    }
  }, [parsedSchema, elements, content]);

  // Toggle section collapse
  const toggleSection = useCallback((sectionId: string) => {
    setCollapsedSections(prev => {
      const newSet = new Set(prev);
      if (newSet.has(sectionId)) {
        newSet.delete(sectionId);
      } else {
        newSet.add(sectionId);
      }
      return newSet;
    });
  }, []);

  // Copy to clipboard
  const handleCopy = useCallback(async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      // Could add toast notification here
    } catch (err) {
      console.error('Failed to copy:', err);
    }
  }, []);

  // Download file
  const handleDownload = useCallback((content: string, filename: string) => {
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  }, []);

  // Filter elements based on search
  const filteredElements = useMemo(() => {
    if (!searchTerm) return elements;
    return elements.filter(el => 
      el.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      el.type.toLowerCase().includes(searchTerm.toLowerCase()) ||
      JSON.stringify(el.properties).toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [elements, searchTerm]);

  const renderCodePreview = () => {
    const code = generateCode(activeFormat);
    
    return (
      <div className="h-full flex flex-col">
        <div className="flex items-center justify-between p-3 border-b">
          <div className="flex items-center space-x-2">
            <Code2 className="w-4 h-4" />
            <span className="text-sm font-medium">Generated Code</span>
          </div>
          <div className="flex items-center space-x-2">
            <Tabs value={activeFormat} onValueChange={(value: any) => setActiveFormat(value)}>
              <TabsList className="h-8">
                <TabsTrigger value="yaml" className="text-xs px-2">YAML</TabsTrigger>
                <TabsTrigger value="json" className="text-xs px-2">JSON</TabsTrigger>
                <TabsTrigger value="sql" className="text-xs px-2">SQL</TabsTrigger>
                <TabsTrigger value="go" className="text-xs px-2">Go</TabsTrigger>
                <TabsTrigger value="graphql" className="text-xs px-2">GraphQL</TabsTrigger>
              </TabsList>
            </Tabs>
            <Button size="sm" variant="outline" onClick={() => handleCopy(code)}>
              <Copy className="w-3 h-3" />
            </Button>
            <Button 
              size="sm" 
              variant="outline" 
              onClick={() => handleDownload(code, `schema.${activeFormat}`)}
            >
              <Download className="w-3 h-3" />
            </Button>
          </div>
        </div>
        
        <ScrollArea className="flex-1">
          <pre className="p-4 text-xs font-mono bg-muted/50">
            <code className={`language-${activeFormat}`}>{code}</code>
          </pre>
        </ScrollArea>
      </div>
    );
  };

  const renderValidationPanel = () => (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between p-3 border-b">
        <div className="flex items-center space-x-2">
          <ShieldCheck className="w-4 h-4" />
          <span className="text-sm font-medium">Validation</span>
          <Badge variant={validationIssues.some(i => i.type === 'error') ? 'destructive' : 'secondary'}>
            {validationIssues.length} issues
          </Badge>
        </div>
        <div className="text-xs text-muted-foreground">
          Quality Score: {qualityMetrics.score}/100
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-3 space-y-3">
          {/* Quality Metrics */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Quality Metrics</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="grid grid-cols-2 gap-3">
                <div className="flex items-center justify-between">
                  <span className="text-xs">Overall Score</span>
                  <Badge variant={qualityMetrics.score >= 80 ? 'default' : qualityMetrics.score >= 60 ? 'secondary' : 'destructive'}>
                    {qualityMetrics.score}
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs">Complexity</span>
                  <Badge variant="outline">{qualityMetrics.complexity}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs">Maintainability</span>
                  <Badge variant="outline">{qualityMetrics.maintainability}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs">Performance</span>
                  <Badge variant="outline">{qualityMetrics.performance}</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-xs">Security</span>
                  <Badge variant="outline">{qualityMetrics.security}</Badge>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Validation Issues */}
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm">Issues</CardTitle>
            </CardHeader>
            <CardContent>
              {validationIssues.length === 0 ? (
                <div className="flex items-center space-x-2 text-green-600">
                  <CheckCircle className="w-4 h-4" />
                  <span className="text-sm">No issues found</span>
                </div>
              ) : (
                <div className="space-y-2">
                  {validationIssues.map(issue => (
                    <div 
                      key={issue.id} 
                      className={`p-2 rounded border cursor-pointer hover:bg-muted/50 ${
                        issue.elementId ? 'hover:bg-blue-50' : ''
                      }`}
                      onClick={() => issue.elementId && onElementSelect?.(issue.elementId)}
                    >
                      <div className="flex items-start space-x-2">
                        <AlertCircle 
                          className={`w-4 h-4 mt-0.5 ${
                            issue.type === 'error' 
                              ? 'text-red-500' 
                              : issue.type === 'warning'
                              ? 'text-yellow-500'
                              : 'text-blue-500'
                          }`} 
                        />
                        <div className="flex-1">
                          <p className="text-sm">{issue.message}</p>
                          {(issue.line || issue.column) && (
                            <p className="text-xs text-muted-foreground">
                              Line {issue.line}:{issue.column}
                            </p>
                          )}
                        </div>
                        <Badge 
                          variant={
                            issue.type === 'error' 
                              ? 'destructive' 
                              : issue.type === 'warning'
                              ? 'secondary'
                              : 'outline'
                          }
                          className="text-xs"
                        >
                          {issue.type}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </ScrollArea>
    </div>
  );

  const renderStructurePanel = () => (
    <div className="h-full flex flex-col">
      <div className="flex items-center justify-between p-3 border-b">
        <div className="flex items-center space-x-2">
          <Layers className="w-4 h-4" />
          <span className="text-sm font-medium">Data Structure</span>
        </div>
        <div className="relative">
          <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-3 h-3 text-muted-foreground" />
          <Input
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-7 h-7 w-32 text-xs"
          />
        </div>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-3 space-y-3">
          {/* Fields Section */}
          <div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => toggleSection('fields')}
              className="w-full justify-start p-2 h-auto"
            >
              {collapsedSections.has('fields') ? 
                <ChevronRight className="w-4 h-4 mr-1" /> : 
                <ChevronDown className="w-4 h-4 mr-1" />
              }
              <Database className="w-4 h-4 mr-2" />
              <span className="text-sm">Fields ({filteredElements.filter(el => el.type === 'field').length})</span>
            </Button>
            
            {!collapsedSections.has('fields') && (
              <div className="ml-6 mt-2 space-y-2">
                {filteredElements
                  .filter(el => el.type === 'field')
                  .map(element => (
                    <Card 
                      key={element.id} 
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => onElementSelect?.(element.id)}
                    >
                      <CardContent className="p-3">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="font-medium text-sm">{element.properties.name}</p>
                            <p className="text-xs text-muted-foreground">
                              {element.properties.type}
                              {element.properties.max_length && ` (${element.properties.max_length})`}
                            </p>
                          </div>
                          <div className="flex space-x-1">
                            {element.properties.primary_key && (
                              <Badge variant="default" className="text-xs">PK</Badge>
                            )}
                            {element.properties.required && (
                              <Badge variant="destructive" className="text-xs">REQ</Badge>
                            )}
                            {element.properties.unique && (
                              <Badge variant="secondary" className="text-xs">UNQ</Badge>
                            )}
                          </div>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
              </div>
            )}
          </div>

          {/* Relationships Section */}
          <div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => toggleSection('relationships')}
              className="w-full justify-start p-2 h-auto"
            >
              {collapsedSections.has('relationships') ? 
                <ChevronRight className="w-4 h-4 mr-1" /> : 
                <ChevronDown className="w-4 h-4 mr-1" />
              }
              <Zap className="w-4 h-4 mr-2" />
              <span className="text-sm">Relationships ({filteredElements.filter(el => el.type === 'relationship').length})</span>
            </Button>
            
            {!collapsedSections.has('relationships') && (
              <div className="ml-6 mt-2 space-y-2">
                {filteredElements
                  .filter(el => el.type === 'relationship')
                  .map(element => (
                    <Card 
                      key={element.id} 
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => onElementSelect?.(element.id)}
                    >
                      <CardContent className="p-3">
                        <div className="flex items-center justify-between">
                          <div>
                            <p className="font-medium text-sm">{element.properties.name}</p>
                            <p className="text-xs text-muted-foreground">
                              {element.properties.type} → {element.properties.target_resource}
                            </p>
                          </div>
                          <Badge variant="outline" className="text-xs">
                            {element.properties.type}
                          </Badge>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
              </div>
            )}
          </div>
        </div>
      </ScrollArea>
    </div>
  );

  return (
    <Card className={`h-full flex flex-col ${className}`}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center">
            <Eye className="w-5 h-5 mr-2" />
            Multi-Panel Preview
          </CardTitle>
          <div className="flex items-center space-x-2">
            <Badge variant="outline" className="text-xs">
              {elements.length} elements
            </Badge>
            <Badge 
              variant={qualityMetrics.score >= 80 ? 'default' : qualityMetrics.score >= 60 ? 'secondary' : 'destructive'}
              className="text-xs"
            >
              {qualityMetrics.score}% quality
            </Badge>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex-1 p-0 overflow-hidden">
        <Tabs value={selectedPreviewTab} onValueChange={setSelectedPreviewTab} className="h-full flex flex-col">
          <div className="px-3 pt-2">
            <TabsList className="grid w-full grid-cols-3 text-xs">
              <TabsTrigger value="code">Code Preview</TabsTrigger>
              <TabsTrigger value="validation">Validation</TabsTrigger>
              <TabsTrigger value="structure">Structure</TabsTrigger>
            </TabsList>
          </div>

          <div className="flex-1 overflow-hidden">
            <TabsContent value="code" className="h-full m-0">
              {renderCodePreview()}
            </TabsContent>

            <TabsContent value="validation" className="h-full m-0">
              {renderValidationPanel()}
            </TabsContent>

            <TabsContent value="structure" className="h-full m-0">
              {renderStructurePanel()}
            </TabsContent>
          </div>
        </Tabs>
      </CardContent>
    </Card>
  );
};

// Helper functions for code generation
function generateSQLSchema(schema: MetaContractSchema, elements: MetaContractElement[]): string {
  const fields = elements.filter(el => el.type === 'field');
  const tableName = schema.resource_name || 'resource';
  
  let sql = `CREATE TABLE ${tableName} (\n`;
  
  sql += fields.map(field => {
    const prop = field.properties;
    let line = `  ${prop.name} `;
    
    switch (prop.type) {
      case 'string':
        line += `VARCHAR(${prop.max_length || 255})`;
        break;
      case 'integer':
        line += 'INTEGER';
        break;
      case 'decimal':
        line += `DECIMAL(${prop.precision || 10},${prop.scale || 2})`;
        break;
      case 'boolean':
        line += 'BOOLEAN';
        break;
      case 'date':
        line += 'DATE';
        break;
      case 'datetime':
        line += 'TIMESTAMP';
        break;
      case 'uuid':
        line += 'UUID';
        break;
      default:
        line += 'TEXT';
    }
    
    if (prop.required) line += ' NOT NULL';
    if (prop.unique) line += ' UNIQUE';
    if (prop.primary_key) line += ' PRIMARY KEY';
    if (prop.default_value) line += ` DEFAULT '${prop.default_value}'`;
    
    return line;
  }).join(',\n');
  
  sql += '\n);';
  
  // Add indexes
  const indexes = fields.filter(f => f.properties.indexed);
  if (indexes.length > 0) {
    sql += '\n\n-- Indexes\n';
    indexes.forEach(field => {
      sql += `CREATE INDEX idx_${tableName}_${field.properties.name} ON ${tableName}(${field.properties.name});\n`;
    });
  }
  
  return sql;
}

function generateGoEntSchema(schema: MetaContractSchema, elements: MetaContractElement[]): string {
  const fields = elements.filter(el => el.type === 'field');
  const entityName = (schema.resource_name || 'Resource').charAt(0).toUpperCase() + 
                     (schema.resource_name || 'Resource').slice(1);
  
  let goCode = `package schema

import (
  "entgo.io/ent"
  "entgo.io/ent/schema/field"
  "entgo.io/ent/schema/edge"
)

// ${entityName} holds the schema definition for the ${entityName} entity.
type ${entityName} struct {
  ent.Schema
}

// Fields of the ${entityName}.
func (${entityName}) Fields() []ent.Field {
  return []ent.Field{
`;

  goCode += fields.map(field => {
    const prop = field.properties;
    let fieldDef = `    field.`;
    
    switch (prop.type) {
      case 'string':
        fieldDef += `String("${prop.name}")`;
        if (prop.max_length) fieldDef += `.MaxLen(${prop.max_length})`;
        break;
      case 'integer':
        fieldDef += `Int("${prop.name}")`;
        break;
      case 'decimal':
        fieldDef += `Float("${prop.name}")`;
        break;
      case 'boolean':
        fieldDef += `Bool("${prop.name}")`;
        break;
      case 'date':
      case 'datetime':
        fieldDef += `Time("${prop.name}")`;
        break;
      case 'uuid':
        fieldDef += `UUID("${prop.name}", uuid.UUID{})`;
        break;
      default:
        fieldDef += `String("${prop.name}")`;
    }
    
    if (prop.required) fieldDef += '.Required()';
    if (prop.unique) fieldDef += '.Unique()';
    if (prop.default_value) fieldDef += `.Default("${prop.default_value}")`;
    
    return fieldDef + ',';
  }).join('\n');

  goCode += `
  }
}

// Edges of the ${entityName}.
func (${entityName}) Edges() []ent.Edge {
  return []ent.Edge{
    // Define relationships here
  }
}
`;

  return goCode;
}

function generateGraphQLSchema(schema: MetaContractSchema, elements: MetaContractElement[]): string {
  const fields = elements.filter(el => el.type === 'field');
  const typeName = (schema.resource_name || 'Resource').charAt(0).toUpperCase() + 
                   (schema.resource_name || 'Resource').slice(1);
  
  let graphql = `type ${typeName} {\n`;
  
  graphql += fields.map(field => {
    const prop = field.properties;
    let fieldType = '';
    
    switch (prop.type) {
      case 'string':
        fieldType = 'String';
        break;
      case 'integer':
        fieldType = 'Int';
        break;
      case 'decimal':
        fieldType = 'Float';
        break;
      case 'boolean':
        fieldType = 'Boolean';
        break;
      case 'date':
      case 'datetime':
        fieldType = 'DateTime';
        break;
      case 'uuid':
        fieldType = 'ID';
        break;
      default:
        fieldType = 'String';
    }
    
    if (prop.required) fieldType += '!';
    
    return `  ${prop.name}: ${fieldType}`;
  }).join('\n');
  
  graphql += '\n}';
  
  // Add input types
  const inputFields = fields.filter(f => !f.properties.primary_key);
  if (inputFields.length > 0) {
    graphql += `\n\ninput ${typeName}Input {\n`;
    graphql += inputFields.map(field => {
      const prop = field.properties;
      let fieldType = '';
      
      switch (prop.type) {
        case 'string':
          fieldType = 'String';
          break;
        case 'integer':
          fieldType = 'Int';
          break;
        case 'decimal':
          fieldType = 'Float';
          break;
        case 'boolean':
          fieldType = 'Boolean';
          break;
        case 'date':
        case 'datetime':
          fieldType = 'DateTime';
          break;
        case 'uuid':
          fieldType = 'ID';
          break;
        default:
          fieldType = 'String';
      }
      
      if (prop.required) fieldType += '!';
      
      return `  ${prop.name}: ${fieldType}`;
    }).join('\n');
    graphql += '\n}';
  }
  
  return graphql;
}