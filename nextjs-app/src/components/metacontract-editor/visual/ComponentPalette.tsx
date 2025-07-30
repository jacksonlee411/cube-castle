import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  Database, 
  Key, 
  Shield, 
  GitBranch, 
  Search, 
  Hash,
  Calendar,
  FileText,
  Mail,
  Phone,
  MapPin,
  User,
  Building,
  DollarSign,
  Clock,
  CheckSquare,
  Link,
  Filter,
  Zap,
  Eye,
  Lock,
  AlertTriangle,
  Plus
} from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';

interface ComponentTemplate {
  id: string;
  name: string;
  description: string;
  type: MetaContractElement['type'];
  icon: React.ComponentType<any>;
  category: string;
  properties: Record<string, any>;
  tags: string[];
}

const COMPONENT_TEMPLATES: ComponentTemplate[] = [
  // Field Components
  {
    id: 'field-string',
    name: 'Text Field',
    description: 'Basic text field with optional validation',
    type: 'field',
    icon: FileText,
    category: 'fields',
    properties: {
      name: 'text_field',
      type: 'string',
      required: false,
      max_length: 255,
      validation: []
    },
    tags: ['basic', 'text', 'string']
  },
  {
    id: 'field-email',
    name: 'Email Field',
    description: 'Email field with built-in validation',
    type: 'field',
    icon: Mail,
    category: 'fields',
    properties: {
      name: 'email',
      type: 'string',
      required: true,
      format: 'email',
      validation: ['email']
    },
    tags: ['contact', 'email', 'validation']
  },
  {
    id: 'field-phone',
    name: 'Phone Field',
    description: 'Phone number with international format support',
    type: 'field',
    icon: Phone,
    category: 'fields',
    properties: {
      name: 'phone',
      type: 'string',
      required: false,
      format: 'phone',
      validation: ['phone']
    },
    tags: ['contact', 'phone', 'validation']
  },
  {
    id: 'field-number',
    name: 'Number Field',
    description: 'Numeric field with range validation',
    type: 'field',
    icon: Hash,
    category: 'fields',
    properties: {
      name: 'number_field',
      type: 'integer',
      required: false,
      min: 0,
      max: 999999
    },
    tags: ['numeric', 'integer', 'validation']
  },
  {
    id: 'field-decimal',
    name: 'Decimal Field',
    description: 'Decimal number with precision control',
    type: 'field',
    icon: DollarSign,
    category: 'fields',
    properties: {
      name: 'decimal_field',
      type: 'decimal',
      required: false,
      precision: 10,
      scale: 2
    },
    tags: ['numeric', 'decimal', 'money']
  },
  {
    id: 'field-date',
    name: 'Date Field',
    description: 'Date field with format validation',
    type: 'field',
    icon: Calendar,
    category: 'fields',
    properties: {
      name: 'date_field',
      type: 'date',
      required: false,
      format: 'YYYY-MM-DD'
    },
    tags: ['time', 'date', 'calendar']
  },
  {
    id: 'field-datetime',
    name: 'DateTime Field',
    description: 'Date and time with timezone support',
    type: 'field',
    icon: Clock,
    category: 'fields',
    properties: {
      name: 'datetime_field',
      type: 'datetime',
      required: false,
      timezone: 'UTC'
    },
    tags: ['time', 'datetime', 'timezone']
  },
  {
    id: 'field-boolean',
    name: 'Boolean Field',
    description: 'True/false checkbox field',
    type: 'field',
    icon: CheckSquare,
    category: 'fields',
    properties: {
      name: 'boolean_field',
      type: 'boolean',
      required: false,
      default: false
    },
    tags: ['boolean', 'checkbox', 'toggle']
  },
  {
    id: 'field-uuid',
    name: 'UUID Field',
    description: 'Unique identifier field',
    type: 'field',
    icon: Key,
    category: 'fields',
    properties: {
      name: 'id',
      type: 'uuid',
      required: true,
      unique: true,
      primary_key: true
    },
    tags: ['id', 'uuid', 'primary', 'unique']
  },
  {
    id: 'field-json',
    name: 'JSON Field',
    description: 'Structured JSON data field',
    type: 'field',
    icon: Database,
    category: 'fields',
    properties: {
      name: 'metadata',
      type: 'json',
      required: false,
      schema: {}
    },
    tags: ['json', 'object', 'metadata']
  },

  // Relationship Components
  {
    id: 'relationship-one-to-many',
    name: 'One-to-Many',
    description: 'One-to-many relationship',
    type: 'relationship',
    icon: GitBranch,
    category: 'relationships',
    properties: {
      name: 'has_many',
      type: 'one_to_many',
      target_resource: 'related_resource',
      foreign_key: 'parent_id',
      cascade_delete: false
    },
    tags: ['relation', 'one-to-many', 'foreign']
  },
  {
    id: 'relationship-many-to-one',
    name: 'Many-to-One',
    description: 'Many-to-one relationship',
    type: 'relationship',
    icon: Link,
    category: 'relationships',
    properties: {
      name: 'belongs_to',
      type: 'many_to_one',
      target_resource: 'parent_resource',
      foreign_key: 'parent_id',
      required: false
    },
    tags: ['relation', 'many-to-one', 'belongs']
  },
  {
    id: 'relationship-many-to-many',
    name: 'Many-to-Many',
    description: 'Many-to-many relationship with join table',
    type: 'relationship',
    icon: GitBranch,
    category: 'relationships',
    properties: {
      name: 'has_and_belongs_to_many',
      type: 'many_to_many',
      target_resource: 'related_resource',
      join_table: 'junction_table',
      through: {}
    },
    tags: ['relation', 'many-to-many', 'junction']
  },

  // Security Components
  {
    id: 'security-rbac',
    name: 'Role-Based Access',
    description: 'Role-based access control model',
    type: 'security',
    icon: Shield,
    category: 'security',
    properties: {
      type: 'rbac',
      roles: ['admin', 'user', 'viewer'],
      permissions: {
        create: ['admin'],
        read: ['admin', 'user', 'viewer'],
        update: ['admin', 'user'],
        delete: ['admin']
      }
    },
    tags: ['security', 'rbac', 'roles']
  },
  {
    id: 'security-rls',
    name: 'Row Level Security',
    description: 'Database row-level security policies',
    type: 'security',
    icon: Lock,
    category: 'security',
    properties: {
      type: 'row_level_security',
      policies: [
        {
          name: 'owner_policy',
          condition: 'user_id = current_user_id()',
          operations: ['select', 'update', 'delete']
        }
      ]
    },
    tags: ['security', 'rls', 'database']
  },

  // Validation Components
  {
    id: 'validation-required',
    name: 'Required Validation',
    description: 'Mark field as required',
    type: 'validation',
    icon: AlertTriangle,
    category: 'validation',
    properties: {
      type: 'required',
      message: 'This field is required',
      fields: []
    },
    tags: ['validation', 'required', 'mandatory']
  },
  {
    id: 'validation-range',
    name: 'Range Validation',
    description: 'Validate numeric ranges',
    type: 'validation',
    icon: Filter,
    category: 'validation',
    properties: {
      type: 'range',
      min: 0,
      max: 100,
      message: 'Value must be between {min} and {max}',
      fields: []
    },
    tags: ['validation', 'range', 'numeric']
  },
  {
    id: 'validation-pattern',
    name: 'Pattern Validation',
    description: 'Regular expression validation',
    type: 'validation',
    icon: Search,
    category: 'validation',
    properties: {
      type: 'pattern',
      pattern: '^[A-Za-z0-9]+$',
      message: 'Invalid format',
      fields: []
    },
    tags: ['validation', 'regex', 'pattern']
  },

  // Index Components
  {
    id: 'index-simple',
    name: 'Simple Index',
    description: 'Single column database index',
    type: 'index',
    icon: Zap,
    category: 'performance',
    properties: {
      name: 'idx_field',
      type: 'btree',
      fields: ['field_name'],
      unique: false
    },
    tags: ['index', 'performance', 'database']
  },
  {
    id: 'index-composite',
    name: 'Composite Index',
    description: 'Multi-column database index',
    type: 'index',
    icon: Database,
    category: 'performance',
    properties: {
      name: 'idx_composite',
      type: 'btree',
      fields: ['field1', 'field2'],
      unique: false
    },
    tags: ['index', 'composite', 'performance']
  }
];

const CATEGORIES = [
  { id: 'all', name: 'All', icon: Database },
  { id: 'fields', name: 'Fields', icon: FileText },
  { id: 'relationships', name: 'Relations', icon: GitBranch },
  { id: 'security', name: 'Security', icon: Shield },
  { id: 'validation', name: 'Validation', icon: AlertTriangle },
  { id: 'performance', name: 'Performance', icon: Zap }
];

interface ComponentPaletteProps {
  onAddElement: (type: MetaContractElement['type'], properties: any) => void;
}

export const ComponentPalette: React.FC<ComponentPaletteProps> = ({
  onAddElement
}) => {
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [searchTerm, setSearchTerm] = useState('');

  const filteredTemplates = React.useMemo(() => {
    let filtered = COMPONENT_TEMPLATES;
    
    if (selectedCategory !== 'all') {
      filtered = filtered.filter(template => template.category === selectedCategory);
    }
    
    if (searchTerm) {
      filtered = filtered.filter(template =>
        template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.tags.some(tag => tag.toLowerCase().includes(searchTerm.toLowerCase()))
      );
    }
    
    return filtered;
  }, [selectedCategory, searchTerm]);

  const handleAddComponent = (template: ComponentTemplate) => {
    const properties = { ...template.properties };
    // Generate unique name if needed
    if (properties.name && properties.name.includes('field')) {
      properties.name = `${properties.name}_${Date.now()}`;
    }
    onAddElement(template.type, properties);
  };

  return (
    <div className="h-full flex flex-col">
      <CardHeader className="pb-2">
        <CardTitle className="text-base">Components</CardTitle>
        <div className="relative">
          <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search components..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8 h-8"
          />
        </div>
      </CardHeader>

      <div className="px-4 pb-2">
        <Tabs value={selectedCategory} onValueChange={setSelectedCategory}>
          <TabsList className="grid grid-cols-2 gap-1 h-auto p-1">
            {CATEGORIES.slice(0, 6).map((category) => {
              const Icon = category.icon;
              return (
                <TabsTrigger
                  key={category.id}
                  value={category.id}
                  className="flex flex-col items-center justify-center h-12 text-xs"
                >
                  <Icon className="w-4 h-4 mb-1" />
                  <span>{category.name}</span>
                </TabsTrigger>
              );
            })}
          </TabsList>
        </Tabs>
      </div>

      <ScrollArea className="flex-1 px-4">
        <div className="space-y-2 pb-4">
          {filteredTemplates.map((template) => {
            const Icon = template.icon;
            return (
              <Card 
                key={template.id} 
                className="cursor-pointer hover:shadow-md transition-shadow duration-200 border-2 border-transparent hover:border-primary/20"
                onClick={() => handleAddComponent(template)}
              >
                <CardContent className="p-3">
                  <div className="flex items-start space-x-3">
                    <div className="flex-shrink-0">
                      <Icon className="w-5 h-5 text-primary" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="text-sm font-medium text-foreground mb-1">
                        {template.name}
                      </h4>
                      <p className="text-xs text-muted-foreground mb-2 line-clamp-2">
                        {template.description}
                      </p>
                      <div className="flex flex-wrap gap-1">
                        <Badge variant="outline" className="text-xs px-1.5 py-0.5">
                          {template.type}
                        </Badge>
                        {template.tags.slice(0, 2).map(tag => (
                          <Badge 
                            key={tag} 
                            variant="secondary" 
                            className="text-xs px-1.5 py-0.5"
                          >
                            {tag}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  </div>
                  <Button
                    size="sm"
                    variant="ghost"
                    className="w-full mt-2 h-6 text-xs"
                  >
                    <Plus className="w-3 h-3 mr-1" />
                    Add
                  </Button>
                </CardContent>
              </Card>
            );
          })}
          
          {filteredTemplates.length === 0 && (
            <div className="text-center py-8 text-muted-foreground">
              <Search className="w-8 h-8 mx-auto mb-2" />
              <p className="text-sm">No components found</p>
              <p className="text-xs">Try adjusting your search or category filter</p>
            </div>
          )}
        </div>
      </ScrollArea>
    </div>
  );
};