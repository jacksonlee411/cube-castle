import React, { useState, useCallback, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Label } from '@/components/ui/label';
import { 
  Search, 
  Filter, 
  X, 
  Plus,
  Trash2,
  RotateCcw,
  BookOpen,
  Database,
  Shield,
  AlertTriangle,
  Zap
} from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';

interface SearchFilter {
  id: string;
  field: string;
  operator: 'equals' | 'contains' | 'starts_with' | 'ends_with' | 'exists' | 'not_exists';
  value: string;
}

interface AdvancedSearchProps {
  elements: MetaContractElement[];
  onFilter: (filteredElements: MetaContractElement[]) => void;
  className?: string;
}

const SEARCHABLE_FIELDS = [
  { value: 'name', label: 'Name' },
  { value: 'type', label: 'Type' },
  { value: 'properties.type', label: 'Data Type' },
  { value: 'properties.required', label: 'Required' },
  { value: 'properties.unique', label: 'Unique' },
  { value: 'properties.primary_key', label: 'Primary Key' }
];

const OPERATORS = [
  { value: 'equals', label: 'Equals' },
  { value: 'contains', label: 'Contains' },
  { value: 'starts_with', label: 'Starts with' },
  { value: 'ends_with', label: 'Ends with' },
  { value: 'exists', label: 'Exists' },
  { value: 'not_exists', label: 'Does not exist' }
];

export const AdvancedSearch: React.FC<AdvancedSearchProps> = ({
  elements,
  onFilter,
  className = ''
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState<SearchFilter[]>([]);
  const [selectedTypes, setSelectedTypes] = useState<string[]>([]);
  const [showAdvanced, setShowAdvanced] = useState(false);

  const handleClearAll = useCallback(() => {
    setSearchTerm('');
    setFilters([]);
    setSelectedTypes([]);
  }, []);

  return (
    <Card className={className}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center">
            <Search className="w-5 h-5 mr-2" />
            Advanced Search
          </CardTitle>
          <Button size="sm" variant="outline" onClick={handleClearAll}>
            <RotateCcw className="w-4 h-4 mr-1" />
            Clear
          </Button>
        </div>
      </CardHeader>

      <CardContent className="space-y-4">
        <div className="relative">
          <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="Search elements..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-8"
          />
        </div>
        
        <div className="text-sm text-muted-foreground">
          {elements.length} elements found
        </div>
      </CardContent>
    </Card>
  );
};