import React, { useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  X, 
  Save, 
  RotateCcw, 
  Plus, 
  Trash2, 
  Copy,
  Database,
  Shield,
  AlertTriangle,
  Zap,
  Settings,
  Info
} from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';

const FIELD_TYPES = [
  'string', 'integer', 'decimal', 'boolean', 'date', 'datetime', 
  'uuid', 'json', 'text', 'email', 'phone', 'url'
];

const VALIDATION_TYPES = [
  'required', 'unique', 'email', 'phone', 'url', 'pattern', 
  'range', 'length', 'custom'
];

const RELATIONSHIP_TYPES = [
  'one_to_one', 'one_to_many', 'many_to_one', 'many_to_many'
];

const INDEX_TYPES = [
  'btree', 'hash', 'gin', 'gist', 'spgist', 'brin'
];

interface PropertyPanelProps {
  element: MetaContractElement;
  onUpdateElement: (updates: Partial<MetaContractElement>) => void;
  onDeleteElement: () => void;
  readonly?: boolean;
}

export const PropertyPanel: React.FC<PropertyPanelProps> = ({
  element,
  onUpdateElement,
  onDeleteElement,
  readonly = false
}) => {
  const [localProperties, setLocalProperties] = useState(element.properties);
  const [hasChanges, setHasChanges] = useState(false);

  const handlePropertyChange = useCallback((key: string, value: any) => {
    setLocalProperties(prev => ({ ...prev, [key]: value }));
    setHasChanges(true);
  }, []);

  const handleNestedPropertyChange = useCallback((parentKey: string, childKey: string, value: any) => {
    setLocalProperties(prev => ({
      ...prev,
      [parentKey]: {
        ...prev[parentKey],
        [childKey]: value
      }
    }));
    setHasChanges(true);
  }, []);

  const handleArrayPropertyChange = useCallback((key: string, index: number, value: any) => {
    setLocalProperties(prev => ({
      ...prev,
      [key]: prev[key]?.map((item: any, i: number) => i === index ? value : item) || []
    }));
    setHasChanges(true);
  }, []);

  const handleAddArrayItem = useCallback((key: string, defaultValue: any) => {
    setLocalProperties(prev => ({
      ...prev,
      [key]: [...(prev[key] || []), defaultValue]
    }));
    setHasChanges(true);
  }, []);

  const handleRemoveArrayItem = useCallback((key: string, index: number) => {
    setLocalProperties(prev => ({
      ...prev,
      [key]: prev[key]?.filter((_: any, i: number) => i !== index) || []
    }));
    setHasChanges(true);
  }, []);

  const handleSave = useCallback(() => {
    onUpdateElement({
      name: localProperties.name || element.name,
      properties: localProperties
    });
    setHasChanges(false);
  }, [localProperties, element.name, onUpdateElement]);

  const handleReset = useCallback(() => {
    setLocalProperties(element.properties);
    setHasChanges(false);
  }, [element.properties]);

  const renderFieldProperties = () => (
    <div className="space-y-4">
      <div>
        <Label htmlFor="field-name">Field Name</Label>
        <Input
          id="field-name"
          value={localProperties.name || ''}
          onChange={(e) => handlePropertyChange('name', e.target.value)}
          placeholder="field_name"
          disabled={readonly}
        />
      </div>

      <div>
        <Label htmlFor="field-type">Data Type</Label>
        <Select
          value={localProperties.type || 'string'}
          onValueChange={(value) => handlePropertyChange('type', value)}
          disabled={readonly}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {FIELD_TYPES.map(type => (
              <SelectItem key={type} value={type}>
                {type}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          id="field-required"
          checked={localProperties.required || false}
          onCheckedChange={(checked) => handlePropertyChange('required', checked)}
          disabled={readonly}
        />
        <Label htmlFor="field-required">Required</Label>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          id="field-unique"
          checked={localProperties.unique || false}
          onCheckedChange={(checked) => handlePropertyChange('unique', checked)}
          disabled={readonly}
        />
        <Label htmlFor="field-unique">Unique</Label>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          id="field-primary"
          checked={localProperties.primary_key || false}
          onCheckedChange={(checked) => handlePropertyChange('primary_key', checked)}
          disabled={readonly}
        />
        <Label htmlFor="field-primary">Primary Key</Label>
      </div>

      {(localProperties.type === 'string' || localProperties.type === 'text') && (
        <div>
          <Label htmlFor="max-length">Max Length</Label>
          <Input
            id="max-length"
            type="number"
            value={localProperties.max_length || ''}
            onChange={(e) => handlePropertyChange('max_length', parseInt(e.target.value) || null)}
            placeholder="255"
            disabled={readonly}
          />
        </div>
      )}

      {(localProperties.type === 'integer' || localProperties.type === 'decimal') && (
        <div className="grid grid-cols-2 gap-2">
          <div>
            <Label htmlFor="min-value">Min Value</Label>
            <Input
              id="min-value"
              type="number"
              value={localProperties.min || ''}
              onChange={(e) => handlePropertyChange('min', parseInt(e.target.value) || null)}
              disabled={readonly}
            />
          </div>
          <div>
            <Label htmlFor="max-value">Max Value</Label>
            <Input
              id="max-value"
              type="number"
              value={localProperties.max || ''}
              onChange={(e) => handlePropertyChange('max', parseInt(e.target.value) || null)}
              disabled={readonly}
            />
          </div>
        </div>
      )}

      {localProperties.type === 'decimal' && (
        <div className="grid grid-cols-2 gap-2">
          <div>
            <Label htmlFor="precision">Precision</Label>
            <Input
              id="precision"
              type="number"
              value={localProperties.precision || ''}
              onChange={(e) => handlePropertyChange('precision', parseInt(e.target.value) || null)}
              placeholder="10"
              disabled={readonly}
            />
          </div>
          <div>
            <Label htmlFor="scale">Scale</Label>
            <Input
              id="scale"
              type="number"
              value={localProperties.scale || ''}
              onChange={(e) => handlePropertyChange('scale', parseInt(e.target.value) || null)}
              placeholder="2"
              disabled={readonly}
            />
          </div>
        </div>
      )}

      <div>
        <Label htmlFor="default-value">Default Value</Label>
        <Input
          id="default-value"
          value={localProperties.default || ''}
          onChange={(e) => handlePropertyChange('default', e.target.value)}
          placeholder="null"
          disabled={readonly}
        />
      </div>

      <div>
        <Label htmlFor="description">Description</Label>
        <Textarea
          id="description"
          value={localProperties.description || ''}
          onChange={(e) => handlePropertyChange('description', e.target.value)}
          placeholder="Field description..."
          disabled={readonly}
        />
      </div>
    </div>
  );

  const renderRelationshipProperties = () => (
    <div className="space-y-4">
      <div>
        <Label htmlFor="relation-name">Relationship Name</Label>
        <Input
          id="relation-name"
          value={localProperties.name || ''}
          onChange={(e) => handlePropertyChange('name', e.target.value)}
          placeholder="has_many_items"
          disabled={readonly}
        />
      </div>

      <div>
        <Label htmlFor="relation-type">Relationship Type</Label>
        <Select
          value={localProperties.type || 'one_to_many'}
          onValueChange={(value) => handlePropertyChange('type', value)}
          disabled={readonly}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {RELATIONSHIP_TYPES.map(type => (
              <SelectItem key={type} value={type}>
                {type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label htmlFor="target-resource">Target Resource</Label>
        <Input
          id="target-resource"
          value={localProperties.target_resource || ''}
          onChange={(e) => handlePropertyChange('target_resource', e.target.value)}
          placeholder="related_table"
          disabled={readonly}
        />
      </div>

      <div>
        <Label htmlFor="foreign-key">Foreign Key</Label>
        <Input
          id="foreign-key"
          value={localProperties.foreign_key || ''}
          onChange={(e) => handlePropertyChange('foreign_key', e.target.value)}
          placeholder="parent_id"
          disabled={readonly}
        />
      </div>

      {localProperties.type === 'many_to_many' && (
        <div>
          <Label htmlFor="join-table">Join Table</Label>
          <Input
            id="join-table"
            value={localProperties.join_table || ''}
            onChange={(e) => handlePropertyChange('join_table', e.target.value)}
            placeholder="junction_table"
            disabled={readonly}
          />
        </div>
      )}

      <div className="flex items-center space-x-2">
        <Switch
          id="cascade-delete"
          checked={localProperties.cascade_delete || false}
          onCheckedChange={(checked) => handlePropertyChange('cascade_delete', checked)}
          disabled={readonly}
        />
        <Label htmlFor="cascade-delete">Cascade Delete</Label>
      </div>
    </div>
  );

  const renderSecurityProperties = () => (
    <div className="space-y-4">
      <div>
        <Label htmlFor="security-type">Security Type</Label>
        <Select
          value={localProperties.type || 'rbac'}
          onValueChange={(value) => handlePropertyChange('type', value)}
          disabled={readonly}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="rbac">Role-Based Access Control</SelectItem>
            <SelectItem value="row_level_security">Row Level Security</SelectItem>
            <SelectItem value="attribute_based">Attribute-Based Access</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {localProperties.type === 'rbac' && (
        <>
          <div>
            <Label>Roles</Label>
            <div className="space-y-2">
              {(localProperties.roles || []).map((role: string, index: number) => (
                <div key={index} className="flex items-center space-x-2">
                  <Input
                    value={role}
                    onChange={(e) => handleArrayPropertyChange('roles', index, e.target.value)}
                    disabled={readonly}
                  />
                  {!readonly && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleRemoveArrayItem('roles', index)}
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  )}
                </div>
              ))}
              {!readonly && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleAddArrayItem('roles', 'new_role')}
                >
                  <Plus className="w-4 h-4 mr-1" />
                  Add Role
                </Button>
              )}
            </div>
          </div>

          <div>
            <Label>Permissions</Label>
            <div className="space-y-2">
              {Object.entries(localProperties.permissions || {}).map(([operation, roles]) => (
                <div key={operation} className="p-2 border rounded">
                  <Label className="text-sm font-medium">{operation}</Label>
                  <div className="flex flex-wrap gap-1 mt-1">
                    {(roles as string[]).map((role, index) => (
                      <Badge key={index} variant="secondary" className="text-xs">
                        {role}
                      </Badge>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );

  const renderValidationProperties = () => (
    <div className="space-y-4">
      <div>
        <Label htmlFor="validation-type">Validation Type</Label>
        <Select
          value={localProperties.type || 'required'}
          onValueChange={(value) => handlePropertyChange('type', value)}
          disabled={readonly}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {VALIDATION_TYPES.map(type => (
              <SelectItem key={type} value={type}>
                {type.charAt(0).toUpperCase() + type.slice(1)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label htmlFor="validation-message">Error Message</Label>
        <Input
          id="validation-message"
          value={localProperties.message || ''}
          onChange={(e) => handlePropertyChange('message', e.target.value)}
          placeholder="Validation error message"
          disabled={readonly}
        />
      </div>

      {localProperties.type === 'pattern' && (
        <div>
          <Label htmlFor="validation-pattern">Pattern (Regex)</Label>
          <Input
            id="validation-pattern"
            value={localProperties.pattern || ''}
            onChange={(e) => handlePropertyChange('pattern', e.target.value)}
            placeholder="^[A-Za-z0-9]+$"
            disabled={readonly}
          />
        </div>
      )}

      {localProperties.type === 'range' && (
        <div className="grid grid-cols-2 gap-2">
          <div>
            <Label htmlFor="range-min">Min Value</Label>
            <Input
              id="range-min"
              type="number"
              value={localProperties.min || ''}
              onChange={(e) => handlePropertyChange('min', parseInt(e.target.value) || null)}
              disabled={readonly}
            />
          </div>
          <div>
            <Label htmlFor="range-max">Max Value</Label>
            <Input
              id="range-max"
              type="number"
              value={localProperties.max || ''}
              onChange={(e) => handlePropertyChange('max', parseInt(e.target.value) || null)}
              disabled={readonly}
            />
          </div>
        </div>
      )}

      <div>
        <Label>Target Fields</Label>
        <div className="space-y-2">
          {(localProperties.fields || []).map((field: string, index: number) => (
            <div key={index} className="flex items-center space-x-2">
              <Input
                value={field}
                onChange={(e) => handleArrayPropertyChange('fields', index, e.target.value)}
                disabled={readonly}
              />
              {!readonly && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleRemoveArrayItem('fields', index)}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              )}
            </div>
          ))}
          {!readonly && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleAddArrayItem('fields', 'field_name')}
            >
              <Plus className="w-4 h-4 mr-1" />
              Add Field
            </Button>
          )}
        </div>
      </div>
    </div>
  );

  const renderIndexProperties = () => (
    <div className="space-y-4">
      <div>
        <Label htmlFor="index-name">Index Name</Label>
        <Input
          id="index-name"
          value={localProperties.name || ''}
          onChange={(e) => handlePropertyChange('name', e.target.value)}
          placeholder="idx_field_name"
          disabled={readonly}
        />
      </div>

      <div>
        <Label htmlFor="index-type">Index Type</Label>
        <Select
          value={localProperties.type || 'btree'}
          onValueChange={(value) => handlePropertyChange('type', value)}
          disabled={readonly}
        >
          <SelectTrigger>
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {INDEX_TYPES.map(type => (
              <SelectItem key={type} value={type}>
                {type.toUpperCase()}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      <div>
        <Label>Fields</Label>
        <div className="space-y-2">
          {(localProperties.fields || []).map((field: string, index: number) => (
            <div key={index} className="flex items-center space-x-2">
              <Input
                value={field}
                onChange={(e) => handleArrayPropertyChange('fields', index, e.target.value)}
                disabled={readonly}
              />
              {!readonly && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => handleRemoveArrayItem('fields', index)}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              )}
            </div>
          ))}
          {!readonly && (
            <Button
              variant="outline"
              size="sm"
              onClick={() => handleAddArrayItem('fields', 'field_name')}
            >
              <Plus className="w-4 h-4 mr-1" />
              Add Field
            </Button>
          )}
        </div>
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          id="index-unique"
          checked={localProperties.unique || false}
          onCheckedChange={(checked) => handlePropertyChange('unique', checked)}
          disabled={readonly}
        />
        <Label htmlFor="index-unique">Unique Index</Label>
      </div>
    </div>
  );

  const renderProperties = () => {
    switch (element.type) {
      case 'field':
        return renderFieldProperties();
      case 'relationship':
        return renderRelationshipProperties();
      case 'security':
        return renderSecurityProperties();
      case 'validation':
        return renderValidationProperties();
      case 'index':
        return renderIndexProperties();
      default:
        return <div>No properties available for this element type.</div>;
    }
  };

  return (
    <div className="h-full flex flex-col">
      <CardHeader className="flex-shrink-0 pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base">Properties</CardTitle>
          <div className="flex items-center space-x-1">
            {hasChanges && (
              <div className="w-2 h-2 bg-orange-500 rounded-full" title="Unsaved changes" />
            )}
            <Badge variant="outline" className="text-xs">
              {element.type}
            </Badge>
          </div>
        </div>
      </CardHeader>

      <ScrollArea className="flex-1 px-4">
        <div className="space-y-4 pb-4">
          {renderProperties()}
        </div>
      </ScrollArea>

      <div className="flex-shrink-0 p-4 border-t bg-muted/20">
        <div className="flex items-center justify-between space-x-2">
          <div className="flex space-x-2">
            {!readonly && hasChanges && (
              <>
                <Button size="sm" onClick={handleSave}>
                  <Save className="w-4 h-4 mr-1" />
                  Save
                </Button>
                <Button variant="outline" size="sm" onClick={handleReset}>
                  <RotateCcw className="w-4 h-4 mr-1" />
                  Reset
                </Button>
              </>
            )}
          </div>
          
          {!readonly && (
            <Button 
              variant="destructive" 
              size="sm" 
              onClick={onDeleteElement}
            >
              <Trash2 className="w-4 h-4 mr-1" />
              Delete
            </Button>
          )}
        </div>
      </div>
    </div>
  );
};