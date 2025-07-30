import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { ScrollArea } from '@/components/ui/scroll-area';
import { 
  AlertCircle, 
  CheckCircle, 
  AlertTriangle, 
  FileText, 
  Code, 
  Database,
  Clock,
  Loader2,
  ExternalLink
} from 'lucide-react';

interface CompileError {
  line: number;
  column: number;
  message: string;
  type: string;
  severity: string;
}

interface CompileWarning {
  line: number;
  column: number;
  message: string;
  type: string;
}

interface CompileResults {
  success: boolean;
  errors?: CompileError[];
  warnings?: CompileWarning[];
  generated_files?: Record<string, string>;
  schema?: any;
  compile_time?: string;
}

interface CompilationResultsProps {
  isCompiling: boolean;
  results?: CompileResults;
  onErrorClick?: (line: number, column: number) => void;
}

export const CompilationResults: React.FC<CompilationResultsProps> = ({
  isCompiling,
  results,
  onErrorClick
}) => {
  const [activeTab, setActiveTab] = useState('overview');

  const handleErrorClick = (error: CompileError) => {
    onErrorClick?.(error.line, error.column);
  };

  const renderOverview = () => (
    <div className="space-y-4">
      {isCompiling ? (
        <div className="flex items-center justify-center py-8">
          <div className="text-center">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-2 text-blue-500" />
            <p className="text-sm text-gray-600">Compiling meta-contract...</p>
          </div>
        </div>
      ) : results ? (
        <>
          {/* Status Summary */}
          <div className="flex items-center space-x-2">
            {results.success ? (
              <CheckCircle className="w-5 h-5 text-green-500" />
            ) : (
              <AlertCircle className="w-5 h-5 text-red-500" />
            )}
            <span className="font-medium">
              {results.success ? 'Compilation Successful' : 'Compilation Failed'}
            </span>
            {results.compile_time && (
              <Badge variant="outline">
                <Clock className="w-3 h-3 mr-1" />
                {results.compile_time}
              </Badge>
            )}
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-4">
            <div className="text-center p-3 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-gray-900">
                {results.errors?.length || 0}
              </div>
              <div className="text-sm text-gray-600">Errors</div>
            </div>
            <div className="text-center p-3 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-gray-900">
                {results.warnings?.length || 0}
              </div>
              <div className="text-sm text-gray-600">Warnings</div>
            </div>
            <div className="text-center p-3 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-gray-900">
                {Object.keys(results.generated_files || {}).length}
              </div>
              <div className="text-sm text-gray-600">Generated Files</div>
            </div>
          </div>

          {/* Quick Actions */}
          {results.success && (
            <div className="space-y-2">
              <Button variant="outline" size="sm" className="w-full">
                <ExternalLink className="w-4 h-4 mr-1" />
                View Generated Schema
              </Button>
              <Button variant="outline" size="sm" className="w-full">
                <Database className="w-4 h-4 mr-1" />
                Run Database Migration
              </Button>
            </div>
          )}
        </>
      ) : (
        <div className="text-center py-8 text-gray-500">
          <FileText className="w-12 h-12 mx-auto mb-2 opacity-50" />
          <p>No compilation results yet</p>
          <p className="text-sm">Click compile to see results here</p>
        </div>
      )}
    </div>
  );

  const renderErrors = () => (
    <ScrollArea className="h-full">
      <div className="space-y-2">
        {results?.errors?.length ? (
          results.errors.map((error, index) => (
            <div
              key={index}
              className="p-3 border border-red-200 rounded-lg bg-red-50 cursor-pointer hover:bg-red-100 transition-colors"
              onClick={() => handleErrorClick(error)}
            >
              <div className="flex items-start space-x-2">
                <AlertCircle className="w-4 h-4 text-red-500 mt-0.5 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2 mb-1">
                    <Badge variant="destructive" className="text-xs">
                      {error.type}
                    </Badge>
                    <span className="text-xs text-gray-600">
                      Line {error.line}:{error.column}
                    </span>
                  </div>
                  <p className="text-sm text-red-800">{error.message}</p>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="text-center py-8 text-gray-500">
            <CheckCircle className="w-8 h-8 mx-auto mb-2 text-green-500" />
            <p>No errors found</p>
          </div>
        )}
      </div>
    </ScrollArea>
  );

  const renderWarnings = () => (
    <ScrollArea className="h-full">
      <div className="space-y-2">
        {results?.warnings?.length ? (
          results.warnings.map((warning, index) => (
            <div
              key={index}
              className="p-3 border border-yellow-200 rounded-lg bg-yellow-50"
            >
              <div className="flex items-start space-x-2">
                <AlertTriangle className="w-4 h-4 text-yellow-500 mt-0.5 flex-shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-2 mb-1">
                    <Badge variant="outline" className="text-xs">
                      {warning.type}
                    </Badge>
                    <span className="text-xs text-gray-600">
                      Line {warning.line}:{warning.column}
                    </span>
                  </div>
                  <p className="text-sm text-yellow-800">{warning.message}</p>
                </div>
              </div>
            </div>
          ))
        ) : (
          <div className="text-center py-8 text-gray-500">
            <CheckCircle className="w-8 h-8 mx-auto mb-2 text-green-500" />
            <p>No warnings found</p>
          </div>
        )}
      </div>
    </ScrollArea>
  );

  const renderGeneratedFiles = () => (
    <ScrollArea className="h-full">
      <div className="space-y-4">
        {results?.generated_files ? (
          Object.entries(results.generated_files).map(([filename, content]) => (
            <Card key={filename}>
              <CardHeader className="pb-2">
                <CardTitle className="text-sm flex items-center">
                  <Code className="w-4 h-4 mr-1" />
                  {filename}
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="max-h-40 overflow-auto">
                  <pre className="bg-gray-900 text-gray-100 p-2 rounded text-xs font-mono overflow-auto">
                    <code>{content}</code>
                  </pre>
                </div>
              </CardContent>
            </Card>
          ))
        ) : (
          <div className="text-center py-8 text-gray-500">
            <FileText className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>No generated files yet</p>
          </div>
        )}
      </div>
    </ScrollArea>
  );

  const getErrorCount = () => results?.errors?.length || 0;
  const getWarningCount = () => results?.warnings?.length || 0;

  return (
    <div className="h-full flex flex-col">
      <div className="p-4 border-b">
        <h3 className="font-semibold text-gray-900">Compilation Results</h3>
      </div>
      
      <div className="flex-1">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="h-full flex flex-col">
          <TabsList className="mx-4 mt-4 grid w-auto grid-cols-4">
            <TabsTrigger value="overview" className="text-xs">
              Overview
            </TabsTrigger>
            <TabsTrigger value="errors" className="text-xs">
              Errors
              {getErrorCount() > 0 && (
                <Badge variant="destructive" className="ml-1 text-xs px-1">
                  {getErrorCount()}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="warnings" className="text-xs">
              Warnings
              {getWarningCount() > 0 && (
                <Badge variant="outline" className="ml-1 text-xs px-1">
                  {getWarningCount()}
                </Badge>
              )}
            </TabsTrigger>
            <TabsTrigger value="files" className="text-xs">
              Files
            </TabsTrigger>
          </TabsList>

          <div className="flex-1 p-4">
            <TabsContent value="overview" className="h-full m-0">
              {renderOverview()}
            </TabsContent>
            <TabsContent value="errors" className="h-full m-0">
              {renderErrors()}
            </TabsContent>
            <TabsContent value="warnings" className="h-full m-0">
              {renderWarnings()}
            </TabsContent>
            <TabsContent value="files" className="h-full m-0">
              {renderGeneratedFiles()}
            </TabsContent>
          </div>
        </Tabs>
      </div>
    </div>
  );
};

const getLanguageFromFilename = (filename: string): string => {
  const ext = filename.split('.').pop()?.toLowerCase();
  switch (ext) {
    case 'go':
      return 'go';
    case 'sql':
      return 'sql';
    case 'json':
      return 'json';
    case 'yaml':
    case 'yml':
      return 'yaml';
    case 'ts':
    case 'tsx':
      return 'typescript';
    case 'js':
    case 'jsx':
      return 'javascript';
    default:
      return 'text';
  }
};