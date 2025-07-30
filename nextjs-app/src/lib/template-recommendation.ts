/**
 * 智能模板推荐系统
 * 基于项目上下文、现有元素、用户偏好和兼容性分析提供智能推荐
 */

import {
  IntelligentTemplate,
  TemplateRecommendationContext,
  TemplateRecommendation,
  TemplateCategory,
  TemplateComplexity,
  MetaContractElement,
  TemplateSearchFilter,
  TemplateSearchResult
} from '@/types/template';
import { EnterpriseTemplateLibrary } from './template-library';

// 推荐策略接口
interface RecommendationStrategy {
  name: string;
  weight: number;
  calculate(template: IntelligentTemplate, context: TemplateRecommendationContext): number;
}

/**
 * 基于内容的推荐策略
 */
class ContentBasedStrategy implements RecommendationStrategy {
  name = 'content-based';
  weight = 0.4;

  calculate(template: IntelligentTemplate, context: TemplateRecommendationContext): number {
    let score = 0;
    let factors = 0;

    // 1. 类别匹配度
    if (context.existingCategories.length > 0) {
      const categoryMatch = context.existingCategories.some(cat => 
        template.category === cat || this.isRelatedCategory(template.category, cat)
      );
      score += categoryMatch ? 30 : 0;
      factors += 30;
    }

    // 2. 字段相似性分析
    const fieldSimilarity = this.calculateFieldSimilarity(template.elements, context.existingElements);
    score += fieldSimilarity * 25;
    factors += 25;

    // 3. 复杂度匹配
    if (context.userPreferences?.complexity) {
      const complexityMatch = context.userPreferences.complexity.includes(template.complexity);
      score += complexityMatch ? 20 : 0;
      factors += 20;
    }

    // 4. 关键词匹配
    const keywordScore = this.calculateKeywordMatch(template, context);
    score += keywordScore * 15;
    factors += 15;

    // 5. 技术兼容性
    const compatibilityScore = this.calculateTechnicalCompatibility(template, context);
    score += compatibilityScore * 10;
    factors += 10;

    return factors > 0 ? (score / factors) * 100 : 0;
  }

  private isRelatedCategory(cat1: TemplateCategory, cat2: TemplateCategory): boolean {
    // 定义相关类别映射
    const relatedCategories: Record<TemplateCategory, TemplateCategory[]> = {
      [TemplateCategory.HR_MANAGEMENT]: [TemplateCategory.EMPLOYEE_PROFILE, TemplateCategory.ORGANIZATION],
      [TemplateCategory.FINANCIAL_SERVICES]: [TemplateCategory.AUDIT_TRAIL, TemplateCategory.DATA_ENCRYPTION],
      [TemplateCategory.ECOMMERCE]: [TemplateCategory.PRODUCT_CATALOG, TemplateCategory.ORDER_MANAGEMENT],
      [TemplateCategory.RBAC]: [TemplateCategory.PERMISSION_MANAGEMENT, TemplateCategory.USER_MANAGEMENT],
      [TemplateCategory.AUDIT_TRAIL]: [TemplateCategory.VERSION_CONTROL, TemplateCategory.SOFT_DELETE],
      [TemplateCategory.DATA_ENCRYPTION]: [TemplateCategory.PII_PROTECTION, TemplateCategory.GDPR_COMPLIANCE],
      // 其他映射...
    } as any;

    return relatedCategories[cat1]?.includes(cat2) || relatedCategories[cat2]?.includes(cat1) || false;
  }

  private calculateFieldSimilarity(templateElements: MetaContractElement[], existingElements: MetaContractElement[]): number {
    if (existingElements.length === 0) return 0;

    const existingFieldTypes = new Set(existingElements.filter(e => e.type === 'field').map(e => e.properties.type));
    const templateFieldTypes = new Set(templateElements.filter(e => e.type === 'field').map(e => e.properties.type));
    
    const intersection = new Set([...existingFieldTypes].filter(type => templateFieldTypes.has(type)));
    const union = new Set([...existingFieldTypes, ...templateFieldTypes]);
    
    return union.size > 0 ? intersection.size / union.size : 0;
  }

  private calculateKeywordMatch(template: IntelligentTemplate, context: TemplateRecommendationContext): number {
    // 从现有元素中提取关键词
    const contextKeywords = context.existingElements
      .flatMap(e => [e.name, e.properties.name, e.properties.description])
      .filter(Boolean)
      .map(k => k.toString().toLowerCase())
      .join(' ');

    const templateKeywords = template.keywords.concat(template.tags).join(' ').toLowerCase();
    
    // 简单的关键词匹配计算
    const matches = template.keywords.filter(keyword => 
      contextKeywords.includes(keyword.toLowerCase())
    ).length;

    return template.keywords.length > 0 ? matches / template.keywords.length : 0;
  }

  private calculateTechnicalCompatibility(template: IntelligentTemplate, context: TemplateRecommendationContext): number {
    let score = 1; // 默认兼容

    if (context.technicalConstraints) {
      const { database, framework, specVersion } = context.technicalConstraints;
      
      if (database && template.compatibility.supportedDatabases && 
          !template.compatibility.supportedDatabases.includes(database)) {
        score *= 0.5;
      }
      
      if (framework && template.compatibility.supportedFrameworks && 
          !template.compatibility.supportedFrameworks.includes(framework)) {
        score *= 0.5;
      }
      
      if (specVersion && !this.isVersionCompatible(specVersion, template.compatibility.minSpecVersion)) {
        score *= 0.3;
      }
    }

    return score;
  }

  private isVersionCompatible(required: string, minimum: string): boolean {
    // 简单的版本比较
    const reqParts = required.split('.').map(Number);
    const minParts = minimum.split('.').map(Number);
    
    for (let i = 0; i < Math.max(reqParts.length, minParts.length); i++) {
      const req = reqParts[i] || 0;
      const min = minParts[i] || 0;
      
      if (req > min) return true;
      if (req < min) return false;
    }
    
    return true;
  }
}

/**
 * 协同过滤推荐策略
 */
class CollaborativeStrategy implements RecommendationStrategy {
  name = 'collaborative';
  weight = 0.3;

  calculate(template: IntelligentTemplate, context: TemplateRecommendationContext): number {
    // 基于社区评分和使用统计
    let score = 0;
    
    // 1. 社区评分权重 (40%)
    score += (template.quality.communityRating / 5) * 40;
    
    // 2. 使用频次权重 (30%)
    const maxUsage = 10000; // 假设最大使用次数
    score += Math.min(template.quality.usageCount / maxUsage, 1) * 30;
    
    // 3. 质量评分权重 (30%)
    const avgQuality = (
      template.quality.performanceScore +
      template.quality.securityScore +
      template.quality.maintainabilityScore +
      template.quality.bestPracticesScore
    ) / 4;
    score += (avgQuality / 100) * 30;

    return score;
  }
}

/**
 * 基于规则的推荐策略
 */
class RuleBasedStrategy implements RecommendationStrategy {
  name = 'rule-based';
  weight = 0.3;

  calculate(template: IntelligentTemplate, context: TemplateRecommendationContext): number {
    let score = 0;
    
    // 1. 项目类型匹配
    if (context.projectType) {
      score += this.getProjectTypeScore(template.category, context.projectType);
    }
    
    // 2. 行业匹配
    if (context.industry) {
      score += this.getIndustryScore(template.category, context.industry);
    }
    
    // 3. 团队规模适配
    if (context.teamSize) {
      score += this.getTeamSizeScore(template.complexity, context.teamSize);
    }
    
    // 4. 用户偏好匹配
    if (context.userPreferences) {
      score += this.getUserPreferenceScore(template, context.userPreferences);
    }

    return Math.min(score, 100);
  }

  private getProjectTypeScore(category: TemplateCategory, projectType: string): number {
    const mapping: Record<string, TemplateCategory[]> = {
      'hr': [TemplateCategory.HR_MANAGEMENT, TemplateCategory.EMPLOYEE_PROFILE, TemplateCategory.ORGANIZATION],
      'finance': [TemplateCategory.FINANCIAL_SERVICES, TemplateCategory.AUDIT_TRAIL],
      'ecommerce': [TemplateCategory.ECOMMERCE, TemplateCategory.PRODUCT_CATALOG, TemplateCategory.ORDER_MANAGEMENT],
      'security': [TemplateCategory.RBAC, TemplateCategory.DATA_ENCRYPTION, TemplateCategory.PII_PROTECTION]
    };

    return mapping[projectType]?.includes(category) ? 30 : 0;
  }

  private getIndustryScore(category: TemplateCategory, industry: string): number {
    const mapping: Record<string, TemplateCategory[]> = {
      'healthcare': [TemplateCategory.HEALTHCARE, TemplateCategory.PII_PROTECTION, TemplateCategory.AUDIT_TRAIL],
      'finance': [TemplateCategory.FINANCIAL_SERVICES, TemplateCategory.DATA_ENCRYPTION, TemplateCategory.GDPR_COMPLIANCE],
      'education': [TemplateCategory.EDUCATION, TemplateCategory.USER_MANAGEMENT],
      'retail': [TemplateCategory.ECOMMERCE, TemplateCategory.PRODUCT_CATALOG]
    };

    return mapping[industry]?.includes(category) ? 25 : 0;
  }

  private getTeamSizeScore(complexity: TemplateComplexity, teamSize: number): number {
    if (teamSize <= 3) {
      return complexity === TemplateComplexity.BASIC ? 20 : 0;
    } else if (teamSize <= 10) {
      return complexity === TemplateComplexity.INTERMEDIATE ? 20 : 10;
    } else if (teamSize <= 50) {
      return complexity === TemplateComplexity.ADVANCED ? 20 : 15;
    } else {
      return complexity === TemplateComplexity.ENTERPRISE ? 20 : 10;
    }
  }

  private getUserPreferenceScore(template: IntelligentTemplate, preferences: any): number {
    let score = 0;
    
    if (preferences.categories && preferences.categories.includes(template.category)) {
      score += 15;
    }
    
    if (preferences.complexity && preferences.complexity.includes(template.complexity)) {
      score += 15;
    }
    
    if (preferences.author && template.author.id === preferences.author) {
      score += 10;
    }
    
    if (preferences.minRating && template.quality.communityRating >= preferences.minRating) {
      score += 10;
    }

    return score;
  }
}

/**
 * 智能模板推荐引擎
 */
export class IntelligentTemplateRecommendationEngine {
  private strategies: RecommendationStrategy[] = [
    new ContentBasedStrategy(),
    new CollaborativeStrategy(),
    new RuleBasedStrategy()
  ];

  /**
   * 获取模板推荐
   */
  async getRecommendations(
    context: TemplateRecommendationContext,
    limit: number = 10
  ): Promise<TemplateRecommendation[]> {
    const allTemplates = EnterpriseTemplateLibrary.getAllTemplates();
    const recommendations: TemplateRecommendation[] = [];

    for (const template of allTemplates) {
      // 检查基本兼容性
      const compatibility = this.checkCompatibility(template, context);
      if (compatibility === 'poor') continue;

      // 计算综合推荐得分
      const score = this.calculateCompositeScore(template, context);
      
      // 生成推荐理由
      const reasons = this.generateReasons(template, context, score);
      
      // 评估冲突风险
      const conflictRisk = this.assessConflictRisk(template, context);

      recommendations.push({
        template,
        score,
        reasons,
        compatibility,
        conflictRisk
      });
    }

    // 按得分排序并返回前N个
    return recommendations
      .sort((a, b) => b.score - a.score)
      .slice(0, limit);
  }

  /**
   * 搜索模板
   */
  async searchTemplates(filter: TemplateSearchFilter): Promise<TemplateSearchResult> {
    let templates = EnterpriseTemplateLibrary.getAllTemplates();

    // 应用过滤器
    if (filter.query) {
      templates = EnterpriseTemplateLibrary.searchTemplates(filter.query);
    }

    if (filter.categories && filter.categories.length > 0) {
      templates = templates.filter(t => filter.categories!.includes(t.category));
    }

    if (filter.complexity && filter.complexity.length > 0) {
      templates = templates.filter(t => filter.complexity!.includes(t.complexity));
    }

    if (filter.tags && filter.tags.length > 0) {
      templates = templates.filter(t => 
        filter.tags!.some(tag => t.tags.includes(tag))
      );
    }

    if (filter.author) {
      templates = templates.filter(t => t.author.id === filter.author);
    }

    if (filter.minRating) {
      templates = templates.filter(t => t.quality.communityRating >= filter.minRating!);
    }

    if (filter.maxRating) {
      templates = templates.filter(t => t.quality.communityRating <= filter.maxRating!);
    }

    // 排序
    if (filter.sortBy) {
      templates = this.sortTemplates(templates, filter.sortBy, filter.sortOrder || 'desc');
    }

    // 分页
    const total = templates.length;
    const offset = filter.offset || 0;
    const limit = filter.limit || 20;
    const paginatedTemplates = templates.slice(offset, offset + limit);

    // 生成聚合信息
    const facets = this.generateFacets(templates);

    return {
      templates: paginatedTemplates,
      total,
      facets
    };
  }

  /**
   * 计算综合推荐得分
   */
  private calculateCompositeScore(template: IntelligentTemplate, context: TemplateRecommendationContext): number {
    let weightedScore = 0;
    let totalWeight = 0;

    for (const strategy of this.strategies) {
      const strategyScore = strategy.calculate(template, context);
      weightedScore += strategyScore * strategy.weight;
      totalWeight += strategy.weight;
    }

    return totalWeight > 0 ? weightedScore / totalWeight : 0;
  }

  /**
   * 检查模板兼容性
   */
  private checkCompatibility(template: IntelligentTemplate, context: TemplateRecommendationContext): 'perfect' | 'good' | 'partial' | 'poor' {
    let compatibilityScore = 100;

    // 检查依赖冲突
    if (template.compatibility.conflicts) {
      const hasConflicts = context.existingCategories.some(cat => 
        template.compatibility.conflicts!.some(conflict => conflict.includes(cat))
      );
      if (hasConflicts) compatibilityScore -= 50;
    }

    // 检查技术约束
    if (context.technicalConstraints) {
      const { database, framework, specVersion } = context.technicalConstraints;
      
      if (database && template.compatibility.supportedDatabases && 
          !template.compatibility.supportedDatabases.includes(database)) {
        compatibilityScore -= 30;
      }
      
      if (framework && template.compatibility.supportedFrameworks && 
          !template.compatibility.supportedFrameworks.includes(framework)) {
        compatibilityScore -= 20;
      }
    }

    if (compatibilityScore >= 90) return 'perfect';
    if (compatibilityScore >= 70) return 'good';
    if (compatibilityScore >= 50) return 'partial';
    return 'poor';
  }

  /**
   * 生成推荐理由
   */
  private generateReasons(template: IntelligentTemplate, context: TemplateRecommendationContext, score: number): string[] {
    const reasons: string[] = [];

    // 高质量评分
    if (template.quality.communityRating >= 4.0) {
      reasons.push(`High community rating (${template.quality.communityRating}/5.0)`);
    }

    // 广泛使用
    if (template.quality.usageCount > 1000) {
      reasons.push(`Widely adopted (${template.quality.usageCount.toLocaleString()} projects)`);
    }

    // 类别匹配
    if (context.existingCategories.includes(template.category)) {
      reasons.push(`Matches your project category (${template.category})`);
    }

    // 复杂度适配
    if (context.userPreferences?.complexity?.includes(template.complexity)) {
      reasons.push(`Matches preferred complexity level (${template.complexity})`);
    }

    // 安全性优秀
    if (template.quality.securityScore >= 95) {
      reasons.push('Excellent security practices');
    }

    // 性能优秀
    if (template.quality.performanceScore >= 90) {
      reasons.push('Optimized for performance');
    }

    // 最佳实践
    if (template.quality.bestPracticesScore >= 90) {
      reasons.push('Follows industry best practices');
    }

    return reasons.slice(0, 3); // 最多返回3个理由
  }

  /**
   * 评估冲突风险
   */
  private assessConflictRisk(template: IntelligentTemplate, context: TemplateRecommendationContext): 'none' | 'low' | 'medium' | 'high' {
    // 检查字段名称冲突
    const existingFieldNames = new Set(
      context.existingElements
        .filter(e => e.type === 'field')
        .map(e => e.properties.name)
    );
    
    const templateFieldNames = new Set(
      template.elements
        .filter(e => e.type === 'field')
        .map(e => e.properties.name)
    );

    const nameConflicts = [...existingFieldNames].filter(name => templateFieldNames.has(name));
    
    if (nameConflicts.length === 0) return 'none';
    if (nameConflicts.length <= 2) return 'low';
    if (nameConflicts.length <= 5) return 'medium';
    return 'high';
  }

  /**
   * 排序模板
   */
  private sortTemplates(templates: IntelligentTemplate[], sortBy: string, sortOrder: 'asc' | 'desc'): IntelligentTemplate[] {
    const sorted = [...templates].sort((a, b) => {
      let comparison = 0;

      switch (sortBy) {
        case 'rating':
          comparison = a.quality.communityRating - b.quality.communityRating;
          break;
        case 'usage':
          comparison = a.quality.usageCount - b.quality.usageCount;
          break;
        case 'recent':
          comparison = a.updatedAt.getTime() - b.updatedAt.getTime();
          break;
        case 'name':
          comparison = a.name.localeCompare(b.name);
          break;
        default:
          comparison = 0;
      }

      return sortOrder === 'desc' ? -comparison : comparison;
    });

    return sorted;
  }

  /**
   * 生成聚合信息
   */
  private generateFacets(templates: IntelligentTemplate[]): TemplateSearchResult['facets'] {
    const categories = new Map<TemplateCategory, number>();
    const complexity = new Map<TemplateComplexity, number>();
    const tags = new Map<string, number>();
    const authors = new Map<string, number>();

    templates.forEach(template => {
      // 统计类别
      categories.set(template.category, (categories.get(template.category) || 0) + 1);
      
      // 统计复杂度
      complexity.set(template.complexity, (complexity.get(template.complexity) || 0) + 1);
      
      // 统计标签
      template.tags.forEach(tag => {
        tags.set(tag, (tags.get(tag) || 0) + 1);
      });
      
      // 统计作者
      authors.set(template.author.name, (authors.get(template.author.name) || 0) + 1);
    });

    return {
      categories: Array.from(categories.entries()).map(([category, count]) => ({ category, count })),
      complexity: Array.from(complexity.entries()).map(([complexity, count]) => ({ complexity, count })),
      tags: Array.from(tags.entries()).map(([tag, count]) => ({ tag, count })).sort((a, b) => b.count - a.count).slice(0, 20),
      authors: Array.from(authors.entries()).map(([author, count]) => ({ author, count }))
    };
  }
}