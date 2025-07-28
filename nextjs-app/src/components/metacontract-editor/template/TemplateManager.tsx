/**
 * 智能模板管理界面
 * 提供模板浏览、搜索、预览、推荐和应用功能
 */

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Separator } from '@/components/ui/separator';
import { 
  Search, 
  Filter, 
  Star, 
  Download, 
  Eye, 
  Plus,
  TrendingUp,
  Shield,
  Zap,
  Users,
  AlertTriangle,
  CheckCircle,
  Info,
  Sparkles,
  BookOpen,
  Settings,
  Layers,
  ChevronDown,
  ChevronRight
} from 'lucide-react';

import {
  IntelligentTemplate,
  TemplateCategory,
  TemplateComplexity,
  TemplateRecommendation,
  TemplateSearchFilter,
  TemplateSearchResult,
  MetaContractElement
} from '@/types/template';
import { IntelligentTemplateRecommendationEngine } from '@/lib/template-recommendation';
import { EnterpriseTemplateLibrary } from '@/lib/template-library';

interface TemplateManagerProps {
  existingElements: MetaContractElement[];
  onApplyTemplate: (template: IntelligentTemplate) => void;
  onClose: () => void;
  open: boolean;
}

// 分类映射显示名称
const CATEGORY_LABELS: Record<TemplateCategory, string> = {
  [TemplateCategory.HR_MANAGEMENT]: 'HR Management',
  [TemplateCategory.FINANCIAL_SERVICES]: 'Financial Services',
  [TemplateCategory.HEALTHCARE]: 'Healthcare',
  [TemplateCategory.ECOMMERCE]: 'E-commerce',
  [TemplateCategory.EDUCATION]: 'Education',
  [TemplateCategory.MANUFACTURING]: 'Manufacturing',
  [TemplateCategory.AUDIT_TRAIL]: 'Audit Trail',
  [TemplateCategory.SOFT_DELETE]: 'Soft Delete',
  [TemplateCategory.MULTI_TENANT]: 'Multi-Tenant',
  [TemplateCategory.VERSION_CONTROL]: 'Version Control',
  [TemplateCategory.CACHING_STRATEGY]: 'Caching Strategy',
  [TemplateCategory.EVENT_SOURCING]: 'Event Sourcing',
  [TemplateCategory.RBAC]: 'RBAC',
  [TemplateCategory.DATA_ENCRYPTION]: 'Data Encryption',
  [TemplateCategory.PII_PROTECTION]: 'PII Protection',
  [TemplateCategory.GDPR_COMPLIANCE]: 'GDPR Compliance',
  [TemplateCategory.OAUTH_INTEGRATION]: 'OAuth Integration',
  [TemplateCategory.API_SECURITY]: 'API Security',
  [TemplateCategory.EMPLOYEE_PROFILE]: 'Employee Profile',
  [TemplateCategory.ORGANIZATION]: 'Organization',
  [TemplateCategory.PERMISSION_MANAGEMENT]: 'Permission Management',
  [TemplateCategory.USER_MANAGEMENT]: 'User Management',
  [TemplateCategory.PRODUCT_CATALOG]: 'Product Catalog',
  [TemplateCategory.ORDER_MANAGEMENT]: 'Order Management'
};

// 复杂度标签
const COMPLEXITY_LABELS: Record<TemplateComplexity, string> = {
  [TemplateComplexity.BASIC]: 'Basic',
  [TemplateComplexity.INTERMEDIATE]: 'Intermediate',
  [TemplateComplexity.ADVANCED]: 'Advanced',
  [TemplateComplexity.ENTERPRISE]: 'Enterprise'
};

// 复杂度颜色
const COMPLEXITY_COLORS: Record<TemplateComplexity, string> = {
  [TemplateComplexity.BASIC]: 'bg-green-100 text-green-800',
  [TemplateComplexity.INTERMEDIATE]: 'bg-blue-100 text-blue-800',
  [TemplateComplexity.ADVANCED]: 'bg-orange-100 text-orange-800',
  [TemplateComplexity.ENTERPRISE]: 'bg-purple-100 text-purple-800'
};

/**
 * 模板卡片组件
 */
const TemplateCard: React.FC<{
  template: IntelligentTemplate;
  onPreview: (template: IntelligentTemplate) => void;
  onApply: (template: IntelligentTemplate) => void;
  recommendation?: TemplateRecommendation;
}> = ({ template, onPreview, onApply, recommendation }) => {
  const getRatingStars = (rating: number) => {
    return Array.from({ length: 5 }, (_, i) => (
      <Star
        key={i}
        className={`w-4 h-4 ${
          i < Math.floor(rating) 
            ? 'text-yellow-400 fill-current' 
            : 'text-gray-300'
        }`}
      />
    ));
  };

  return (
    <Card className="hover:shadow-lg transition-shadow duration-200 relative">
      {recommendation && (
        <div className="absolute top-2 right-2">
          <Badge variant="secondary" className="bg-blue-100 text-blue-800">
            <Sparkles className="w-3 h-3 mr-1" />
            {recommendation.score.toFixed(0)}% match
          </Badge>
        </div>
      )}
      
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1 pr-8">
            <CardTitle className="text-lg font-semibold mb-2">{template.name}</CardTitle>
            <p className="text-sm text-muted-foreground line-clamp-2">
              {template.description}
            </p>
          </div>
        </div>
        
        <div className="flex items-center justify-between mt-3">
          <div className="flex items-center space-x-2">
            <Badge variant="outline">{CATEGORY_LABELS[template.category]}</Badge>
            <Badge className={COMPLEXITY_COLORS[template.complexity]}>
              {COMPLEXITY_LABELS[template.complexity]}
            </Badge>
          </div>
          
          <div className="flex items-center space-x-1">
            {getRatingStars(template.quality.communityRating)}
            <span className="text-sm text-muted-foreground ml-1">
              ({template.quality.communityRating})
            </span>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="pt-0">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center space-x-4 text-sm text-muted-foreground">
            <div className="flex items-center space-x-1">
              <Users className="w-4 h-4" />
              <span>{template.quality.usageCount.toLocaleString()}</span>
            </div>
            <div className="flex items-center space-x-1">
              <Shield className="w-4 h-4" />
              <span>{template.quality.securityScore}%</span>
            </div>
            <div className="flex items-center space-x-1">
              <Zap className="w-4 h-4" />
              <span>{template.quality.performanceScore}%</span>
            </div>
          </div>
        </div>
        
        {recommendation && recommendation.reasons.length > 0 && (
          <div className="mb-3">
            <p className="text-sm font-medium mb-1">Why recommended:</p>
            <ul className="text-sm text-muted-foreground space-y-1">
              {recommendation.reasons.map((reason, index) => (
                <li key={index} className="flex items-start space-x-1">
                  <CheckCircle className="w-3 h-3 text-green-500 mt-0.5 flex-shrink-0" />
                  <span>{reason}</span>
                </li>
              ))}
            </ul>
          </div>
        )}
        
        <div className="flex flex-wrap gap-1 mb-3">
          {template.tags.slice(0, 4).map((tag) => (
            <Badge key={tag} variant="secondary" className="text-xs">
              {tag}
            </Badge>
          ))}
          {template.tags.length > 4 && (
            <Badge variant="secondary" className="text-xs">
              +{template.tags.length - 4} more
            </Badge>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          <Button variant="outline" size="sm" onClick={() => onPreview(template)}>
            <Eye className="w-4 h-4 mr-1" />
            Preview
          </Button>
          <Button size="sm" onClick={() => onApply(template)}>
            <Plus className="w-4 h-4 mr-1" />
            Apply
          </Button>
        </div>
        
        {recommendation && recommendation.conflictRisk !== 'none' && (
          <div className="mt-2 p-2 bg-yellow-50 border border-yellow-200 rounded-md">
            <div className="flex items-center space-x-1">
              <AlertTriangle className="w-4 h-4 text-yellow-600" />
              <span className="text-sm text-yellow-800">
                {recommendation.conflictRisk} conflict risk detected
              </span>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
};

/**
 * 模板预览对话框
 */
const TemplatePreviewDialog: React.FC<{
  template: IntelligentTemplate | null;
  open: boolean;
  onClose: () => void;
  onApply: (template: IntelligentTemplate) => void;
}> = ({ template, open, onClose, onApply }) => {
  if (!template) return null;

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <span>{template.name}</span>
            <Badge className={COMPLEXITY_COLORS[template.complexity]}>
              {COMPLEXITY_LABELS[template.complexity]}
            </Badge>
          </DialogTitle>
          <DialogDescription>{template.description}</DialogDescription>
        </DialogHeader>
        
        <ScrollArea className="flex-1 mt-4">
          <div className="space-y-6">
            {/* 基本信息 */}
            <div>
              <h4 className="font-semibold mb-2">Basic Information</h4>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span className="text-muted-foreground">Category:</span>
                  <span className="ml-2">{CATEGORY_LABELS[template.category]}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Version:</span>
                  <span className="ml-2">{template.version}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Author:</span>
                  <span className="ml-2">{template.author.name}</span>
                </div>
                <div>
                  <span className="text-muted-foreground">Usage Count:</span>
                  <span className="ml-2">{template.quality.usageCount.toLocaleString()}</span>
                </div>
              </div>
            </div>
            
            <Separator />
            
            {/* 质量指标 */}
            <div>
              <h4 className="font-semibold mb-2">Quality Metrics</h4>
              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Performance</span>
                    <span className="text-sm font-medium">{template.quality.performanceScore}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-600 h-2 rounded-full" 
                      style={{ width: `${template.quality.performanceScore}%` }}
                    />
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Security</span>
                    <span className="text-sm font-medium">{template.quality.securityScore}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-600 h-2 rounded-full" 
                      style={{ width: `${template.quality.securityScore}%` }}
                    />
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Maintainability</span>
                    <span className="text-sm font-medium">{template.quality.maintainabilityScore}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-orange-600 h-2 rounded-full" 
                      style={{ width: `${template.quality.maintainabilityScore}%` }}
                    />
                  </div>
                </div>
                
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm">Best Practices</span>
                    <span className="text-sm font-medium">{template.quality.bestPracticesScore}%</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-purple-600 h-2 rounded-full" 
                      style={{ width: `${template.quality.bestPracticesScore}%` }}
                    />
                  </div>
                </div>
              </div>
            </div>
            
            <Separator />
            
            {/* 字段预览 */}
            <div>
              <h4 className="font-semibold mb-2">Fields ({template.elements.filter(e => e.type === 'field').length})</h4>
              <div className="space-y-2 max-h-48 overflow-y-auto">
                {template.elements.filter(e => e.type === 'field').map((element) => (
                  <div key={element.id} className="flex items-center justify-between p-2 border rounded">
                    <div>
                      <span className="font-medium">{element.properties.name}</span>
                      <span className="text-sm text-muted-foreground ml-2">
                        ({element.properties.type})
                      </span>
                    </div>
                    <div className="flex items-center space-x-1">
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
            </div>
            
            {/* 关系预览 */}
            {template.elements.filter(e => e.type === 'relationship').length > 0 && (
              <>
                <Separator />
                <div>
                  <h4 className="font-semibold mb-2">Relationships</h4>
                  <div className="space-y-2">
                    {template.elements.filter(e => e.type === 'relationship').map((element) => (
                      <div key={element.id} className="p-2 border rounded">
                        <div className="font-medium">{element.properties.name}</div>
                        <div className="text-sm text-muted-foreground">
                          {element.properties.type} → {element.properties.target}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </>
            )}
            
            {/* 标签 */}
            <div>
              <h4 className="font-semibold mb-2">Tags</h4>
              <div className="flex flex-wrap gap-1">
                {template.tags.map((tag) => (
                  <Badge key={tag} variant="secondary">{tag}</Badge>
                ))}
              </div>
            </div>
          </div>
        </ScrollArea>
        
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Close</Button>
          <Button onClick={() => onApply(template)}>
            <Plus className="w-4 h-4 mr-1" />
            Apply Template
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};

/**
 * 主模板管理器组件
 */
export const TemplateManager: React.FC<TemplateManagerProps> = ({
  existingElements,
  onApplyTemplate,
  onClose,
  open
}) => {
  const [activeTab, setActiveTab] = useState('recommended');
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<TemplateCategory | 'all'>('all');
  const [selectedComplexity, setSelectedComplexity] = useState<TemplateComplexity | 'all'>('all');
  const [sortBy, setSortBy] = useState<'relevance' | 'rating' | 'usage' | 'recent'>('relevance');
  
  const [recommendations, setRecommendations] = useState<TemplateRecommendation[]>([]);
  const [searchResults, setSearchResults] = useState<TemplateSearchResult | null>(null);
  const [selectedTemplate, setSelectedTemplate] = useState<IntelligentTemplate | null>(null);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  
  const recommendationEngine = useMemo(() => new IntelligentTemplateRecommendationEngine(), []);

  // 获取推荐模板
  const loadRecommendations = useCallback(async () => {
    setLoading(true);
    try {
      const context = {
        existingElements,
        existingCategories: [...new Set(existingElements.map(e => e.type))] as TemplateCategory[],
        projectType: 'general',
        userPreferences: {
          complexity: [TemplateComplexity.BASIC, TemplateComplexity.INTERMEDIATE],
          minRating: 3.0
        }
      };
      
      const recs = await recommendationEngine.getRecommendations(context, 12);
      setRecommendations(recs);
    } catch (error) {
      console.error('Failed to load recommendations:', error);
    } finally {
      setLoading(false);
    }
  }, [existingElements, recommendationEngine]);

  // 搜索模板
  const searchTemplates = useCallback(async () => {
    setLoading(true);
    try {
      const filter: TemplateSearchFilter = {
        query: searchQuery || undefined,
        categories: selectedCategory !== 'all' ? [selectedCategory] : undefined,
        complexity: selectedComplexity !== 'all' ? [selectedComplexity] : undefined,
        sortBy,
        limit: 20
      };
      
      const results = await recommendationEngine.searchTemplates(filter);
      setSearchResults(results);
    } catch (error) {
      console.error('Failed to search templates:', error);
    } finally {
      setLoading(false);
    }
  }, [searchQuery, selectedCategory, selectedComplexity, sortBy, recommendationEngine]);

  // 初始化加载
  useEffect(() => {
    if (open) {
      loadRecommendations();
      searchTemplates();
    }
  }, [open, loadRecommendations, searchTemplates]);

  // 搜索参数变化时重新搜索
  useEffect(() => {
    if (activeTab === 'browse') {
      searchTemplates();
    }
  }, [activeTab, searchQuery, selectedCategory, selectedComplexity, sortBy, searchTemplates]);

  const handlePreview = (template: IntelligentTemplate) => {
    setSelectedTemplate(template);
    setPreviewOpen(true);
  };

  const handleApply = (template: IntelligentTemplate) => {
    onApplyTemplate(template);
    onClose();
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="max-w-7xl max-h-[90vh] overflow-hidden">
        <DialogHeader>
          <DialogTitle className="flex items-center space-x-2">
            <Layers className="w-5 h-5" />
            <span>Template Library</span>
          </DialogTitle>
          <DialogDescription>
            Discover and apply enterprise-grade templates to accelerate your development
          </DialogDescription>
        </DialogHeader>
        
        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="recommended" className="flex items-center space-x-2">
              <Sparkles className="w-4 h-4" />
              <span>Recommended</span>
            </TabsTrigger>
            <TabsTrigger value="browse" className="flex items-center space-x-2">
              <Search className="w-4 h-4" />
              <span>Browse All</span>
            </TabsTrigger>
            <TabsTrigger value="popular" className="flex items-center space-x-2">
              <TrendingUp className="w-4 h-4" />
              <span>Popular</span>
            </TabsTrigger>
          </TabsList>
          
          <div className="flex-1 flex flex-col mt-4">
            {/* 搜索和过滤栏 */}
            {activeTab === 'browse' && (
              <div className="flex items-center space-x-4 mb-4 p-4 bg-muted/50 rounded-lg">
                <div className="flex-1 relative">
                  <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input
                    placeholder="Search templates..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    className="pl-10"
                  />
                </div>
                
                <Select value={selectedCategory} onValueChange={(value: any) => setSelectedCategory(value)}>
                  <SelectTrigger className="w-48">
                    <SelectValue placeholder="Category" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Categories</SelectItem>
                    {Object.entries(CATEGORY_LABELS).map(([key, label]) => (
                      <SelectItem key={key} value={key}>{label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                
                <Select value={selectedComplexity} onValueChange={(value: any) => setSelectedComplexity(value)}>
                  <SelectTrigger className="w-40">
                    <SelectValue placeholder="Complexity" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="all">All Levels</SelectItem>
                    {Object.entries(COMPLEXITY_LABELS).map(([key, label]) => (
                      <SelectItem key={key} value={key}>{label}</SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                
                <Select value={sortBy} onValueChange={(value: any) => setSortBy(value)}>
                  <SelectTrigger className="w-32">
                    <SelectValue placeholder="Sort by" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="relevance">Relevance</SelectItem>
                    <SelectItem value="rating">Rating</SelectItem>
                    <SelectItem value="usage">Usage</SelectItem>
                    <SelectItem value="recent">Recent</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}
            
            {/* 模板内容区域 */}
            <ScrollArea className="flex-1">
              <TabsContent value="recommended" className="mt-0">
                {loading ? (
                  <div className="flex items-center justify-center h-64">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                ) : (
                  <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-1">
                    {recommendations.map((rec) => (
                      <TemplateCard
                        key={rec.template.id}
                        template={rec.template}
                        recommendation={rec}
                        onPreview={handlePreview}
                        onApply={handleApply}
                      />
                    ))}
                  </div>
                )}
              </TabsContent>
              
              <TabsContent value="browse" className="mt-0">
                {loading ? (
                  <div className="flex items-center justify-center h-64">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
                  </div>
                ) : searchResults ? (
                  <>
                    <div className="mb-4 text-sm text-muted-foreground">
                      Found {searchResults.total} templates
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-1">
                      {searchResults.templates.map((template) => (
                        <TemplateCard
                          key={template.id}
                          template={template}
                          onPreview={handlePreview}
                          onApply={handleApply}
                        />
                      ))}
                    </div>
                  </>
                ) : null}
              </TabsContent>
              
              <TabsContent value="popular" className="mt-0">
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-1">
                  {EnterpriseTemplateLibrary.getAllTemplates()
                    .sort((a, b) => b.quality.usageCount - a.quality.usageCount)
                    .slice(0, 12)
                    .map((template) => (
                      <TemplateCard
                        key={template.id}
                        template={template}
                        onPreview={handlePreview}
                        onApply={handleApply}
                      />
                    ))}
                </div>
              </TabsContent>
            </ScrollArea>
          </div>
        </Tabs>
        
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>Close</Button>
        </DialogFooter>
      </DialogContent>
      
      {/* 模板预览对话框 */}
      <TemplatePreviewDialog
        template={selectedTemplate}
        open={previewOpen}
        onClose={() => setPreviewOpen(false)}
        onApply={handleApply}
      />
    </Dialog>
  );
};