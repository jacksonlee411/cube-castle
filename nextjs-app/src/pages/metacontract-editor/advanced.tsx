import React, { useState } from 'react';
import Head from 'next/head';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { 
  Palette, 
  Code, 
  Eye, 
  Database,
  Search,
  GitBranch,
  Save,
  Download,
  Share,
  Play,
  ArrowLeft
} from 'lucide-react';
import Link from 'next/link';
import { MetaContractEditor } from '@/components/metacontract-editor/MetaContractEditor';
import { VisualEditor } from '@/components/metacontract-editor/VisualEditor';
import { AdvancedSearch } from '@/components/metacontract-editor/visual/AdvancedSearch';
import { VersionComparison } from '@/components/metacontract-editor/visual/VersionComparison';

const SAMPLE_CONTENT = `specification_version: "1.0"
api_id: "employee-management-api"
namespace: "hr.company.com"
resource_name: "employee"

data_structure:
  primary_key: "id"
  data_classification: "internal"
  fields:
    - name: "id"
      type: "uuid"
      required: true
      unique: true
      primary_key: true
    - name: "email"
      type: "string"
      required: true
      unique: true
      max_length: 255
      format: "email"
    - name: "first_name"
      type: "string"
      required: true
      max_length: 100
    - name: "last_name"
      type: "string"
      required: true
      max_length: 100`;

const SAMPLE_VERSIONS = [
  {
    id: 'v1.0.0',
    name: 'v1.0.0',
    description: 'Initial employee model',
    content: SAMPLE_CONTENT,
    timestamp: '2024-01-15T10:00:00Z',
    author: 'John Doe',
    changes: { added: 12, modified: 0, removed: 0 }
  },
  {
    id: 'v1.1.0',
    name: 'v1.1.0', 
    description: 'Added security model and validation',
    content: SAMPLE_CONTENT,
    timestamp: '2024-01-20T14:30:00Z',
    author: 'Jane Smith',
    changes: { added: 8, modified: 3, removed: 1 }
  }
];

const AdvancedMetaContractEditor: React.FC = () => {
  const [currentTab, setCurrentTab] = useState('visual-editor');
  const [content, setContent] = useState(SAMPLE_CONTENT);
  const [theme, setTheme] = useState<'light' | 'dark'>('light');

  return (
    <>
      <Head>
        <title>Advanced Meta-Contract Visual Editor | Cube Castle</title>
        <meta name="description" content="Professional visual editor for meta-contracts" />
      </Head>

      <div className="min-h-screen bg-gradient-to-br from-background to-muted/20">
        <header className="bg-white border-b shadow-sm">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center justify-between h-16">
              <div className="flex items-center space-x-4">
                <Link href="/" className="flex items-center space-x-2 text-primary hover:text-primary/80">
                  <ArrowLeft className="w-5 h-5" />
                  <span>Back to Dashboard</span>
                </Link>
                <div className="h-6 w-px bg-border" />
                <h1 className="text-xl font-bold text-foreground">
                  Advanced Meta-Contract Editor
                </h1>
                <Badge variant="secondary" className="text-xs">
                  Beta
                </Badge>
              </div>
              
              <div className="flex items-center space-x-2">
                <Button size="sm" variant="outline">
                  <Save className="w-4 h-4 mr-1" />
                  Save
                </Button>
                <Button size="sm" variant="outline">
                  <Share className="w-4 h-4 mr-1" />
                  Share
                </Button>
                <Button size="sm" variant="outline">
                  <Download className="w-4 h-4 mr-1" />
                  Export
                </Button>
                <Button size="sm">
                  <Play className="w-4 h-4 mr-1" />
                  Compile
                </Button>
              </div>
            </div>
          </div>
        </header>

        <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <Tabs value={currentTab} onValueChange={setCurrentTab} className="space-y-6">
            <div className="bg-white rounded-lg shadow-sm border p-1">
              <TabsList className="grid w-full grid-cols-4">
                <TabsTrigger value="visual-editor" className="flex items-center space-x-2">
                  <Palette className="w-4 h-4" />
                  <span>Visual Editor</span>
                </TabsTrigger>
                <TabsTrigger value="code-editor" className="flex items-center space-x-2">
                  <Code className="w-4 h-4" />
                  <span>Code Editor</span>
                </TabsTrigger>
                <TabsTrigger value="search-tools" className="flex items-center space-x-2">
                  <Search className="w-4 h-4" />
                  <span>Search & Filter</span>
                </TabsTrigger>
                <TabsTrigger value="version-control" className="flex items-center space-x-2">
                  <GitBranch className="w-4 h-4" />
                  <span>Version Control</span>
                </TabsTrigger>
              </TabsList>
            </div>

            <TabsContent value="visual-editor">
              <Card className="overflow-hidden">
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between">
                    <CardTitle className="flex items-center space-x-2">
                      <Palette className="w-5 h-5" />
                      <span>Visual Meta-Contract Editor</span>
                    </CardTitle>
                    <Badge variant="outline" className="text-xs">
                      Employee Management API v1.2.0
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="p-0">
                  <div className="h-[800px]">
                    <VisualEditor
                      content={content}
                      onChange={setContent}
                      theme={theme}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="code-editor">
              <Card className="overflow-hidden">
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center space-x-2">
                    <Code className="w-5 h-5" />
                    <span>Enhanced Code Editor</span>
                  </CardTitle>
                </CardHeader>
                <CardContent className="p-0">
                  <div className="h-[800px]">
                    <MetaContractEditor
                      initialContent={content}
                      readonly={false}
                    />
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="search-tools">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <Search className="w-5 h-5" />
                    <span>Advanced Search & Filter</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <AdvancedSearch
                    elements={[]}
                    onFilter={(filtered) => console.log('Filtered:', filtered)}
                  />
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="version-control">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center space-x-2">
                    <GitBranch className="w-5 h-5" />
                    <span>Version Control & History</span>
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <VersionComparison
                    versions={SAMPLE_VERSIONS}
                    currentVersion="v1.1.0"
                    onVersionSelect={(versionId) => console.log('Selected version:', versionId)}
                  />
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </main>
      </div>
    </>
  );
};

export default AdvancedMetaContractEditor;