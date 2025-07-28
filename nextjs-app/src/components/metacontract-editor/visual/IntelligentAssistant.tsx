import React, { useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Brain, Lightbulb, RefreshCw } from 'lucide-react';
import { MetaContractElement } from '../VisualEditor';

interface IntelligentAssistantProps {
  content: string;
  elements: MetaContractElement[];
  onAddElement: (type: MetaContractElement['type'], properties: any) => void;
  onUpdateElement: (elementId: string, updates: Partial<MetaContractElement>) => void;
}

export const IntelligentAssistant: React.FC<IntelligentAssistantProps> = ({
  content,
  elements,
  onAddElement,
  onUpdateElement
}) => {
  const [isAnalyzing, setIsAnalyzing] = useState(false);

  const generateSuggestions = useCallback(async () => {
    setIsAnalyzing(true);
    // Simulate AI analysis
    setTimeout(() => {
      setIsAnalyzing(false);
    }, 2000);
  }, []);

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="pb-2">
        <div className="flex items-center justify-between">
          <CardTitle className="text-base flex items-center">
            <Brain className="w-5 h-5 mr-2" />
            AI Assistant
          </CardTitle>
          <Button
            size="sm"
            variant="outline"
            onClick={generateSuggestions}
            disabled={isAnalyzing}
          >
            {isAnalyzing ? (
              <RefreshCw className="w-4 h-4 mr-1 animate-spin" />
            ) : (
              <Lightbulb className="w-4 h-4 mr-1" />
            )}
            Analyze
          </Button>
        </div>
      </CardHeader>

      <CardContent className="flex-1 overflow-auto space-y-2">
        <div className="text-center py-8">
          {isAnalyzing ? (
            <div className="space-y-2">
              <RefreshCw className="w-8 h-8 mx-auto animate-spin text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Analyzing schema...</p>
            </div>
          ) : (
            <div className="space-y-2">
              <Lightbulb className="w-8 h-8 mx-auto text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Click "Analyze" to get AI suggestions</p>
            </div>
          )}
        </div>
        
        <div className="space-y-2">
          <Badge variant="outline" className="text-xs">
            {elements.length} elements analyzed
          </Badge>
        </div>
      </CardContent>
    </Card>
  );
};