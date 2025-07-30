import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Database, ZoomIn, ZoomOut, RotateCcw, Download } from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';

interface ERDiagramProps {
  elements: MetaContractElement[];
  onElementSelect?: (elementId: string) => void;
  className?: string;
}

export const ERDiagram: React.FC<ERDiagramProps> = ({
  elements,
  onElementSelect,
  className = ''
}) => {
  const handleZoom = (direction: 'in' | 'out') => {
    console.log('Zoom', direction);
  };

  const handleReset = () => {
    console.log('Reset view');
  };

  const handleDownload = () => {
    console.log('Download diagram');
  };

  return (
    <Card className={`h-full flex flex-col ${className}`}>
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center">
            <Database className="w-5 h-5 mr-2" />
            Entity Relationship Diagram
          </CardTitle>
          <div className="flex items-center space-x-1">
            <Button size="sm" variant="outline" onClick={() => handleZoom('in')}>
              <ZoomIn className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={() => handleZoom('out')}>
              <ZoomOut className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={handleReset}>
              <RotateCcw className="w-4 h-4" />
            </Button>
            <Button size="sm" variant="outline" onClick={handleDownload}>
              <Download className="w-4 h-4" />
            </Button>
          </div>
        </div>
        
        <div className="flex items-center space-x-2 text-xs text-muted-foreground">
          <Badge variant="outline" className="text-xs">
            {elements.filter(el => el.type === 'field').length} entities
          </Badge>
          <Badge variant="outline" className="text-xs">
            {elements.filter(el => el.type === 'relationship').length} relationships
          </Badge>
          <span>Zoom: 100%</span>
        </div>
      </CardHeader>

      <CardContent className="flex-1 p-0 overflow-hidden">
        {elements.length === 0 ? (
          <div className="flex items-center justify-center h-full text-center">
            <div className="text-muted-foreground">
              <Database className="w-12 h-12 mx-auto mb-4" />
              <p className="text-sm">No entities to display</p>
              <p className="text-xs">Add fields and relationships to see the ER diagram</p>
            </div>
          </div>
        ) : (
          <div className="w-full h-full relative overflow-auto bg-gray-50">
            <div className="flex items-center justify-center h-full">
              <div className="text-center">
                <Database className="w-16 h-16 mx-auto mb-4 text-blue-500" />
                <h3 className="text-lg font-medium mb-2">ER Diagram Placeholder</h3>
                <p className="text-sm text-muted-foreground">
                  Interactive entity relationship diagram would be rendered here
                </p>
                <div className="mt-4 space-y-2">
                  {elements.filter(el => el.type === 'field').slice(0, 3).map((element, index) => (
                    <div key={element.id} className="text-xs bg-white px-2 py-1 rounded border">
                      {element.name} ({element.properties?.type || 'unknown'})
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};