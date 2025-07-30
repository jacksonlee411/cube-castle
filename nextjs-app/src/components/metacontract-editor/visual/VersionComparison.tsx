import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { GitBranch, Clock, Plus, Minus, Edit } from 'lucide-react';

interface Version {
  id: string;
  name: string;
  description: string;
  content: string;
  timestamp: string;
  author: string;
  changes: {
    added: number;
    modified: number;
    removed: number;
  };
}

interface VersionComparisonProps {
  versions: Version[];
  currentVersion: string;
  onVersionSelect?: (versionId: string) => void;
  className?: string;
}

export const VersionComparison: React.FC<VersionComparisonProps> = ({
  versions,
  currentVersion,
  onVersionSelect,
  className = ''
}) => {
  const [fromVersion, setFromVersion] = useState<string>(versions[1]?.id || '');
  const [toVersion, setToVersion] = useState<string>(currentVersion);

  return (
    <Card className={`h-full flex flex-col ${className}`}>
      <CardHeader className="pb-2">
        <CardTitle className="text-base flex items-center">
          <GitBranch className="w-5 h-5 mr-2" />
          Version Comparison
        </CardTitle>
        
        <div className="flex items-center space-x-4">
          <Badge variant="outline" className="text-green-600">
            <Plus className="w-3 h-3 mr-1" />
            5 added
          </Badge>
          <Badge variant="outline" className="text-blue-600">
            <Edit className="w-3 h-3 mr-1" />
            2 modified
          </Badge>
          <Badge variant="outline" className="text-red-600">
            <Minus className="w-3 h-3 mr-1" />
            0 removed
          </Badge>
        </div>
      </CardHeader>

      <CardContent className="flex-1">
        <div className="space-y-4">
          {versions.map(version => (
            <div key={version.id} className="p-3 border rounded-lg">
              <div className="flex items-center justify-between mb-2">
                <h4 className="font-medium">{version.name}</h4>
                <Badge variant="outline">
                  {new Date(version.timestamp).toLocaleDateString()}
                </Badge>
              </div>
              <p className="text-sm text-muted-foreground mb-2">
                {version.description}
              </p>
              <div className="flex items-center space-x-2 text-xs">
                <span>by {version.author}</span>
                <Clock className="w-3 h-3" />
                <span>{new Date(version.timestamp).toLocaleTimeString()}</span>
              </div>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
};