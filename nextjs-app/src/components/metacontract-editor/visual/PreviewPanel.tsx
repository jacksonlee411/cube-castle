import React, { useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Database, 
  Table, 
  GitBranch, 
  Shield, 
  AlertTriangle,
  Zap,
  Eye,
  Code,
  FileText,
  ExternalLink
} from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';
import * as yaml from 'js-yaml';

interface PreviewPanelProps {
  content: string;
  elements: MetaContractElement[];
}

export const PreviewPanel: React.FC<PreviewPanelProps> = ({
  content,
  elements
}) => {
  const parsedSchema = useMemo(() => {
    try {
      return yaml.load(content) as any;
    } catch {
      return null;
    }
  }, [content]);

  const statistics = useMemo(() => {
    const stats = {
      fields: 0,
      relationships: 0,
      security: 0,
      validations: 0,
      indexes: 0,
      requiredFields: 0,
      uniqueFields: 0
    };

    elements.forEach(element => {
      stats[element.type + 's' as keyof typeof stats]++;
      
      if (element.type === 'field') {
        if (element.properties.required) stats.requiredFields++;
        if (element.properties.unique) stats.uniqueFields++;
      }
    });

    return stats;
  }, [elements]);

  const renderSchemaOverview = () => (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <Database className="w-5 h-5 text-blue-600" />
              <div>
                <p className="text-sm font-medium">Fields</p>
                <p className="text-2xl font-bold">{statistics.fields}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <GitBranch className="w-5 h-5 text-green-600" />
              <div>
                <p className="text-sm font-medium">Relations</p>
                <p className="text-2xl font-bold">{statistics.relationships}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <Shield className="w-5 h-5 text-red-600" />
              <div>
                <p className="text-sm font-medium">Security</p>
                <p className="text-2xl font-bold">{statistics.security}</p>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-4">
            <div className="flex items-center space-x-2">
              <Zap className="w-5 h-5 text-purple-600" />
              <div>
                <p className="text-sm font-medium">Indexes</p>
                <p className="text-2xl font-bold">{statistics.indexes}</p>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {parsedSchema && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Schema Information</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            <div className="grid grid-cols-2 gap-2 text-sm">
              <div>
                <span className="text-muted-foreground">API ID:</span>
                <span className="ml-2 font-mono">{parsedSchema.api_id || 'N/A'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Version:</span>
                <span className="ml-2 font-mono">{parsedSchema.specification_version || 'N/A'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Namespace:</span>
                <span className="ml-2 font-mono">{parsedSchema.namespace || 'N/A'}</span>
              </div>
              <div>
                <span className="text-muted-foreground">Resource:</span>
                <span className="ml-2 font-mono">{parsedSchema.resource_name || 'N/A'}</span>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderDataStructure = () => (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center">
            <Table className="w-4 h-4 mr-2" />
            Data Structure
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-3">
            {elements
              .filter(el => el.type === 'field')
              .map(element => (
                <div key={element.id} className="flex items-center justify-between p-2 border rounded">
                  <div className="flex items-center space-x-2">
                    <Database className="w-4 h-4 text-blue-600" />
                    <div>
                      <p className="font-medium text-sm">{element.properties.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {element.properties.type}
                        {element.properties.max_length && ` (${element.properties.max_length})`}
                      </p>
                    </div>
                  </div>
                  <div className="flex space-x-1">
                    {element.properties.primary_key && (
                      <Badge variant="default" className="text-xs">PK</Badge>
                    )}
                    {element.properties.required && (
                      <Badge variant="destructive" className="text-xs">Required</Badge>
                    )}
                    {element.properties.unique && (
                      <Badge variant="secondary" className="text-xs">Unique</Badge>
                    )}
                  </div>
                </div>
              ))}
          </div>
        </CardContent>
      </Card>

      {elements.some(el => el.type === 'relationship') && (
        <Card>
          <CardHeader>
            <CardTitle className="text-sm flex items-center">
              <GitBranch className="w-4 h-4 mr-2" />
              Relationships
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-3">
              {elements
                .filter(el => el.type === 'relationship')
                .map(element => (
                  <div key={element.id} className="flex items-center justify-between p-2 border rounded">
                    <div className="flex items-center space-x-2">
                      <GitBranch className="w-4 h-4 text-green-600" />
                      <div>
                        <p className="font-medium text-sm">{element.properties.name}</p>
                        <p className="text-xs text-muted-foreground">
                          {element.properties.type} → {element.properties.target_resource}
                        </p>
                      </div>
                    </div>
                    <Badge variant="outline" className="text-xs">
                      {element.properties.type}
                    </Badge>
                  </div>
                ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderSecurityModel = () => (
    <div className="space-y-4">
      {elements
        .filter(el => el.type === 'security')
        .map(element => (
          <Card key={element.id}>
            <CardHeader>
              <CardTitle className="text-sm flex items-center">
                <Shield className="w-4 h-4 mr-2" />
                {element.name}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                <div>
                  <p className="text-xs text-muted-foreground mb-1">Type:</p>
                  <Badge variant="outline">{element.properties.type}</Badge>
                </div>
                
                {element.properties.roles && (
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Roles:</p>
                    <div className="flex flex-wrap gap-1">
                      {element.properties.roles.map((role: string, index: number) => (
                        <Badge key={index} variant="secondary" className="text-xs">
                          {role}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}

                {element.properties.permissions && (
                  <div>
                    <p className="text-xs text-muted-foreground mb-1">Permissions:</p>
                    <div className="space-y-1">
                      {Object.entries(element.properties.permissions).map(([operation, roles]) => (
                        <div key={operation} className="flex items-center justify-between text-xs">
                          <span className="font-medium">{operation}:</span>
                          <div className="flex gap-1">
                            {(roles as string[]).map((role, index) => (
                              <Badge key={index} variant="outline" className="text-xs">
                                {role}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        ))}

      {elements.filter(el => el.type === 'security').length === 0 && (
        <Card>
          <CardContent className="p-6 text-center">
            <Shield className="w-8 h-8 mx-auto mb-2 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">No security model defined</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderValidationRules = () => (
    <div className="space-y-4">
      {elements
        .filter(el => el.type === 'validation')
        .map(element => (
          <Card key={element.id}>
            <CardHeader>
              <CardTitle className="text-sm flex items-center">
                <AlertTriangle className="w-4 h-4 mr-2" />
                {element.name}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium">Type:</span>
                  <Badge variant="outline">{element.properties.type}</Badge>
                </div>
                {element.properties.message && (
                  <div>
                    <span className="text-sm font-medium">Message:</span>
                    <p className="text-sm text-muted-foreground mt-1">
                      {element.properties.message}
                    </p>
                  </div>
                )}
                {element.properties.fields && element.properties.fields.length > 0 && (
                  <div>
                    <span className="text-sm font-medium">Target Fields:</span>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {element.properties.fields.map((field: string, index: number) => (
                        <Badge key={index} variant="secondary" className="text-xs">
                          {field}
                        </Badge>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        ))}

      {elements.filter(el => el.type === 'validation').length === 0 && (
        <Card>
          <CardContent className="p-6 text-center">
            <AlertTriangle className="w-8 h-8 mx-auto mb-2 text-muted-foreground" />
            <p className="text-sm text-muted-foreground">No validation rules defined</p>
          </CardContent>
        </Card>
      )}
    </div>
  );

  const renderGeneratedCode = () => (
    <div className="space-y-4">
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center">
            <Code className="w-4 h-4 mr-2" />
            YAML Output
          </CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="text-xs bg-muted p-3 rounded overflow-x-auto">
            <code>{content}</code>
          </pre>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center">
            <FileText className="w-4 h-4 mr-2" />
            SQL Schema (Preview)
          </CardTitle>
        </CardHeader>
        <CardContent>
          <pre className="text-xs bg-muted p-3 rounded overflow-x-auto">
            <code>
{`CREATE TABLE ${parsedSchema?.resource_name || 'resource'} (
${elements
  .filter(el => el.type === 'field')
  .map(el => {
    const prop = el.properties;
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
    if (prop.default) line += ` DEFAULT ${prop.default}`;
    
    return line;
  })
  .join(',\n')}
);`}
            </code>
          </pre>
        </CardContent>
      </Card>
    </div>
  );

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b">
        <h3 className="text-lg font-semibold flex items-center">
          <Eye className="w-5 h-5 mr-2" />
          Schema Preview
        </h3>
      </div>

      <Tabs defaultValue="overview" className="flex-1 flex flex-col">
        <div className="px-4 pt-2">
          <TabsList className="grid w-full grid-cols-5 text-xs">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="structure">Structure</TabsTrigger>
            <TabsTrigger value="security">Security</TabsTrigger>
            <TabsTrigger value="validation">Validation</TabsTrigger>
            <TabsTrigger value="code">Code</TabsTrigger>
          </TabsList>
        </div>

        <div className="flex-1 overflow-hidden">
          <TabsContent value="overview" className="h-full m-0">
            <ScrollArea className="h-full px-4">
              <div className="py-4">
                {renderSchemaOverview()}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="structure" className="h-full m-0">
            <ScrollArea className="h-full px-4">
              <div className="py-4">
                {renderDataStructure()}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="security" className="h-full m-0">
            <ScrollArea className="h-full px-4">
              <div className="py-4">
                {renderSecurityModel()}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="validation" className="h-full m-0">
            <ScrollArea className="h-full px-4">
              <div className="py-4">
                {renderValidationRules()}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="code" className="h-full m-0">
            <ScrollArea className="h-full px-4">
              <div className="py-4">
                {renderGeneratedCode()}
              </div>
            </ScrollArea>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
};