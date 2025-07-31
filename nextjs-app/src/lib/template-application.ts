/**
 * 模板应用机制
 * 实现一键应用、冲突检测解决、增量更新等核心功能
 */

import {
  IntelligentTemplate,
  TemplateApplicationResult,
  MetaContractElement,
  MetaContractSchema
} from '@/types/template';
import * as yaml from 'js-yaml';

// 冲突解决策略
export enum ConflictResolutionStrategy {
  MERGE = 'merge',           // 合并属性
  REPLACE = 'replace',       // 替换现有
  RENAME = 'rename',         // 重命名新的
  SKIP = 'skip',            // 跳过冲突项
  ASK_USER = 'ask_user'     // 询问用户
}

// 字段冲突类型
export enum FieldConflictType {
  NAME_CONFLICT = 'name_conflict',
  TYPE_MISMATCH = 'type_mismatch',
  CONSTRAINT_CONFLICT = 'constraint_conflict',
  RELATIONSHIP_CONFLICT = 'relationship_conflict'
}

// 冲突信息
export interface TemplateConflict {
  id: string;
  type: FieldConflictType;
  existing: MetaContractElement;
  template: MetaContractElement;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  suggestedResolution: ConflictResolutionStrategy;
  resolutionOptions: ConflictResolutionStrategy[];
}

// 应用配置
export interface TemplateApplicationConfig {
  includeElements?: string[];      // 要包含的元素ID
  excludeElements?: string[];      // 要排除的元素ID
  conflictResolution?: Record<string, ConflictResolutionStrategy>;
  mergeStrategy?: 'additive' | 'replace' | 'selective';
  preserveExisting?: boolean;      // 是否保留现有元素
  generateBackup?: boolean;        // 是否生成备份
  validateAfterApply?: boolean;    // 应用后是否验证
}

/**
 * 智能模板应用引擎
 */
export class IntelligentTemplateApplicationEngine {
  /**
   * 应用模板到现有元素
   */
  async applyTemplate(
    template: IntelligentTemplate,
    existingElements: MetaContractElement[],
    config: TemplateApplicationConfig = {}
  ): Promise<TemplateApplicationResult> {
    try {
      // 1. 检测冲突
      const conflicts = this.detectConflicts(template, existingElements);
      
      // 2. 生成备份（如果需要）
      let backup: MetaContractElement[] | undefined;
      if (config.generateBackup !== false) {
        backup = JSON.parse(JSON.stringify(existingElements));
      }
      
      // 3. 准备应用元素
      const elementsToApply = this.prepareElementsForApplication(
        template,
        config.includeElements,
        config.excludeElements
      );
      
      // 4. 解决冲突
      const { resolvedElements, warnings } = this.resolveConflicts(
        elementsToApply,
        existingElements,
        conflicts,
        config.conflictResolution || {}
      );
      
      // 5. 合并元素
      const mergedElements = this.mergeElements(
        existingElements,
        resolvedElements,
        config.mergeStrategy || 'additive',
        config.preserveExisting !== false
      );
      
      // 6. 验证结果（如果需要）
      let validationResults: any = {};
      if (config.validateAfterApply !== false) {
        validationResults = await this.validateAppliedTemplate(mergedElements, template);
      }
      
      // 7. 评估性能影响
      const performanceImpact = this.assessPerformanceImpact(
        existingElements,
        mergedElements,
        template
      );
      
      return {
        success: true,
        template,
        appliedElements: resolvedElements,
        conflicts: conflicts.length > 0 ? conflicts.map(c => ({
          type: this.mapConflictType(c.type),
          existing: c.existing,
          template: c.template,
          resolution: this.mapResolutionStrategy(config.conflictResolution?.[c.id] || c.suggestedResolution)
        })) : undefined,
        warnings,
        performance: performanceImpact
      };
      
    } catch (error) {
      return {
        success: false,
        template,
        appliedElements: [],
        warnings: [error instanceof Error ? error.message : 'Unknown error occurred']
      };
    }
  }

  /**
   * 预览模板应用效果
   */
  async previewTemplateApplication(
    template: IntelligentTemplate,
    existingElements: MetaContractElement[],
    config: TemplateApplicationConfig = {}
  ): Promise<{
    conflicts: TemplateConflict[];
    previewElements: MetaContractElement[];
    impact: {
      added: number;
      modified: number;
      conflicts: number;
      estimatedPerformanceImpact: 'low' | 'medium' | 'high';
    };
  }> {
    // 检测冲突
    const conflicts = this.detectConflicts(template, existingElements);
    
    // 准备应用元素
    const elementsToApply = this.prepareElementsForApplication(
      template,
      config.includeElements,
      config.excludeElements
    );
    
    // 生成预览
    const { resolvedElements } = this.resolveConflicts(
      elementsToApply,
      existingElements,
      conflicts,
      config.conflictResolution || {}
    );
    
    const previewElements = this.mergeElements(
      existingElements,
      resolvedElements,
      config.mergeStrategy || 'additive',
      config.preserveExisting !== false
    );
    
    // 计算影响
    const impact = {
      added: resolvedElements.filter(e => !existingElements.find(ex => ex.id === e.id)).length,
      modified: resolvedElements.filter(e => existingElements.find(ex => ex.id === e.id)).length,
      conflicts: conflicts.length,
      estimatedPerformanceImpact: this.estimatePerformanceImpact(existingElements, previewElements)
    };
    
    return { conflicts, previewElements, impact };
  }

  /**
   * 检测模板与现有元素的冲突
   */
  private detectConflicts(
    template: IntelligentTemplate,
    existingElements: MetaContractElement[]
  ): TemplateConflict[] {
    const conflicts: TemplateConflict[] = [];
    
    // 创建现有元素的映射
    const existingByName = new Map<string, MetaContractElement>();
    const existingById = new Map<string, MetaContractElement>();
    
    existingElements.forEach(element => {
      existingByName.set(element.properties.name || element.name, element);
      existingById.set(element.id, element);
    });
    
    // 检查模板元素的冲突
    template.elements.forEach(templateElement => {
      const existingByNameMatch = existingByName.get(templateElement.properties.name || templateElement.name);
      const existingByIdMatch = existingById.get(templateElement.id);
      
      // 名称冲突
      if (existingByNameMatch && existingByNameMatch.id !== templateElement.id) {
        const conflict = this.createNameConflict(templateElement, existingByNameMatch);
        conflicts.push(conflict);
      }
      
      // ID冲突但内容不同
      if (existingByIdMatch && !this.elementsEqual(templateElement, existingByIdMatch)) {
        const conflict = this.createContentConflict(templateElement, existingByIdMatch);
        conflicts.push(conflict);
      }
      
      // 类型冲突
      if (existingByNameMatch && existingByNameMatch.properties.type !== templateElement.properties.type) {
        const conflict = this.createTypeConflict(templateElement, existingByNameMatch);
        conflicts.push(conflict);
      }
      
      // 约束冲突
      if (existingByNameMatch) {
        const constraintConflicts = this.detectConstraintConflicts(templateElement, existingByNameMatch);
        conflicts.push(...constraintConflicts);
      }
    });
    
    return conflicts;
  }

  /**
   * 创建名称冲突
   */
  private createNameConflict(
    templateElement: MetaContractElement,
    existingElement: MetaContractElement
  ): TemplateConflict {
    return {
      id: `name_conflict_${templateElement.id}`,
      type: FieldConflictType.NAME_CONFLICT,
      existing: existingElement,
      template: templateElement,
      severity: 'medium',
      description: `Field name '${templateElement.properties.name}' already exists`,
      suggestedResolution: ConflictResolutionStrategy.RENAME,
      resolutionOptions: [
        ConflictResolutionStrategy.RENAME,
        ConflictResolutionStrategy.REPLACE,
        ConflictResolutionStrategy.SKIP
      ]
    };
  }

  /**
   * 创建内容冲突
   */
  private createContentConflict(
    templateElement: MetaContractElement,
    existingElement: MetaContractElement
  ): TemplateConflict {
    return {
      id: `content_conflict_${templateElement.id}`,
      type: FieldConflictType.TYPE_MISMATCH,
      existing: existingElement,
      template: templateElement,
      severity: 'high',
      description: `Element '${templateElement.name}' has different properties`,
      suggestedResolution: ConflictResolutionStrategy.MERGE,
      resolutionOptions: [
        ConflictResolutionStrategy.MERGE,
        ConflictResolutionStrategy.REPLACE,
        ConflictResolutionStrategy.ASK_USER
      ]
    };
  }

  /**
   * 创建类型冲突
   */
  private createTypeConflict(
    templateElement: MetaContractElement,
    existingElement: MetaContractElement
  ): TemplateConflict {
    return {
      id: `type_conflict_${templateElement.id}`,
      type: FieldConflictType.TYPE_MISMATCH,
      existing: existingElement,
      template: templateElement,
      severity: 'critical',
      description: `Type mismatch: existing '${existingElement.properties.type}' vs template '${templateElement.properties.type}'`,
      suggestedResolution: ConflictResolutionStrategy.ASK_USER,
      resolutionOptions: [
        ConflictResolutionStrategy.REPLACE,
        ConflictResolutionStrategy.SKIP,
        ConflictResolutionStrategy.ASK_USER
      ]
    };
  }

  /**
   * 检测约束冲突
   */
  private detectConstraintConflicts(
    templateElement: MetaContractElement,
    existingElement: MetaContractElement
  ): TemplateConflict[] {
    const conflicts: TemplateConflict[] = [];
    
    // 检查必填约束冲突
    if (templateElement.properties.required && !existingElement.properties.required) {
      conflicts.push({
        id: `constraint_required_${templateElement.id}`,
        type: FieldConflictType.CONSTRAINT_CONFLICT,
        existing: existingElement,
        template: templateElement,
        severity: 'medium',
        description: `Template requires field to be required, but existing field is optional`,
        suggestedResolution: ConflictResolutionStrategy.MERGE,
        resolutionOptions: [ConflictResolutionStrategy.MERGE, ConflictResolutionStrategy.SKIP]
      });
    }
    
    // 检查唯一约束冲突
    if (templateElement.properties.unique && !existingElement.properties.unique) {
      conflicts.push({
        id: `constraint_unique_${templateElement.id}`,
        type: FieldConflictType.CONSTRAINT_CONFLICT,
        existing: existingElement,
        template: templateElement,
        severity: 'high',
        description: `Template requires field to be unique, but existing field is not`,
        suggestedResolution: ConflictResolutionStrategy.MERGE,
        resolutionOptions: [ConflictResolutionStrategy.MERGE, ConflictResolutionStrategy.SKIP]
      });
    }
    
    return conflicts;
  }

  /**
   * 解决冲突
   */
  private resolveConflicts(
    templateElements: MetaContractElement[],
    existingElements: MetaContractElement[],
    conflicts: TemplateConflict[],
    resolutionConfig: Record<string, ConflictResolutionStrategy>
  ): { resolvedElements: MetaContractElement[]; warnings: string[] } {
    const resolvedElements: MetaContractElement[] = [];
    const warnings: string[] = [];
    const conflictMap = new Map(conflicts.map(c => [c.id, c]));
    
    templateElements.forEach(templateElement => {
      const elementConflicts = conflicts.filter(c => c.template.id === templateElement.id);
      
      if (elementConflicts.length === 0) {
        // 无冲突，直接添加
        resolvedElements.push({ ...templateElement });
      } else {
        // 处理冲突
        let resolvedElement = { ...templateElement };
        
        elementConflicts.forEach(conflict => {
          const resolution = resolutionConfig[conflict.id] || conflict.suggestedResolution;
          
          switch (resolution) {
            case ConflictResolutionStrategy.MERGE:
              resolvedElement = this.mergeElements([conflict.existing], [templateElement], 'additive', true)[0] || resolvedElement;
              break;
              
            case ConflictResolutionStrategy.REPLACE:
              // 保持模板元素不变
              break;
              
            case ConflictResolutionStrategy.RENAME:
              resolvedElement = this.renameElement(resolvedElement, existingElements);
              warnings.push(`Renamed element '${templateElement.name}' to '${resolvedElement.name}' to avoid conflict`);
              break;
              
            case ConflictResolutionStrategy.SKIP:
              resolvedElement = null as any;
              warnings.push(`Skipped element '${templateElement.name}' due to conflict`);
              break;
              
            default:
              warnings.push(`Unresolved conflict for element '${templateElement.name}', using template version`);
          }
        });
        
        if (resolvedElement) {
          resolvedElements.push(resolvedElement);
        }
      }
    });
    
    return { resolvedElements, warnings };
  }

  /**
   * 重命名元素以避免冲突
   */
  private renameElement(
    element: MetaContractElement,
    existingElements: MetaContractElement[]
  ): MetaContractElement {
    const existingNames = new Set(existingElements.map(e => e.properties.name || e.name));
    let newName = element.properties.name || element.name;
    let counter = 1;
    
    while (existingNames.has(newName)) {
      newName = `${element.properties.name || element.name}_${counter}`;
      counter++;
    }
    
    return {
      ...element,
      id: `${element.id}_renamed_${counter - 1}`,
      name: newName,
      properties: {
        ...element.properties,
        name: newName
      }
    };
  }

  /**
   * 合并元素
   */
  private mergeElements(
    existingElements: MetaContractElement[],
    newElements: MetaContractElement[],
    strategy: 'additive' | 'replace' | 'selective',
    preserveExisting: boolean = true
  ): MetaContractElement[] {
    const result: MetaContractElement[] = [];
    const processedIds = new Set<string>();
    
    // 处理现有元素
    if (preserveExisting) {
      existingElements.forEach(existing => {
        const newElement = newElements.find(n => n.id === existing.id);
        
        if (newElement) {
          // 合并属性
          const merged = this.mergeElementProperties(existing, newElement);
          result.push(merged);
          processedIds.add(existing.id);
        } else {
          // 保留现有元素
          result.push({ ...existing });
        }
      });
    }
    
    // 添加新元素
    newElements.forEach(newElement => {
      if (!processedIds.has(newElement.id)) {
        result.push({ ...newElement });
      }
    });
    
    return result;
  }

  /**
   * 合并元素属性
   */
  private mergeElementProperties(
    existing: MetaContractElement,
    template: MetaContractElement
  ): MetaContractElement {
    return {
      ...existing,
      properties: {
        ...existing.properties,
        ...template.properties,
        // 保留现有的某些关键属性
        name: existing.properties.name || template.properties.name,
        // 合并验证规则
        validation: {
          ...existing.properties.validation,
          ...template.properties.validation
        }
      }
    };
  }

  /**
   * 准备要应用的元素
   */
  private prepareElementsForApplication(
    template: IntelligentTemplate,
    includeElements?: string[],
    excludeElements?: string[]
  ): MetaContractElement[] {
    let elements = [...template.elements];
    
    // 应用包含过滤器
    if (includeElements && includeElements.length > 0) {
      elements = elements.filter(e => includeElements.includes(e.id));
    }
    
    // 应用排除过滤器
    if (excludeElements && excludeElements.length > 0) {
      elements = elements.filter(e => !excludeElements.includes(e.id));
    }
    
    return elements;
  }

  /**
   * 验证应用的模板
   */
  private async validateAppliedTemplate(
    elements: MetaContractElement[],
    template: IntelligentTemplate
  ): Promise<any> {
    // 这里可以添加各种验证逻辑
    return {
      valid: true,
      warnings: [],
      errors: []
    };
  }

  /**
   * 评估性能影响
   */
  private assessPerformanceImpact(
    beforeElements: MetaContractElement[],
    afterElements: MetaContractElement[],
    template: IntelligentTemplate
  ): { estimatedImpact: 'low' | 'medium' | 'high'; metrics: Record<string, number> } {
    const metrics = {
      elementCountChange: afterElements.length - beforeElements.length,
      fieldCountChange: afterElements.filter(e => e.type === 'field').length - 
                       beforeElements.filter(e => e.type === 'field').length,
      relationshipCountChange: afterElements.filter(e => e.type === 'relationship').length - 
                              beforeElements.filter(e => e.type === 'relationship').length,
      complexityIncrease: this.calculateComplexityIncrease(beforeElements, afterElements)
    };
    
    const estimatedImpact = this.estimatePerformanceImpact(beforeElements, afterElements);
    
    return { estimatedImpact, metrics };
  }

  /**
   * 估算性能影响
   */
  private estimatePerformanceImpact(
    beforeElements: MetaContractElement[],
    afterElements: MetaContractElement[]
  ): 'low' | 'medium' | 'high' {
    const sizeDifference = afterElements.length - beforeElements.length;
    const percentageIncrease = beforeElements.length > 0 ? sizeDifference / beforeElements.length : 1;
    
    if (percentageIncrease <= 0.2) return 'low';
    if (percentageIncrease <= 0.5) return 'medium';
    return 'high';
  }

  /**
   * 计算复杂度增加
   */
  private calculateComplexityIncrease(
    beforeElements: MetaContractElement[],
    afterElements: MetaContractElement[]
  ): number {
    const beforeComplexity = this.calculateElementsComplexity(beforeElements);
    const afterComplexity = this.calculateElementsComplexity(afterElements);
    return afterComplexity - beforeComplexity;
  }

  /**
   * 计算元素复杂度
   */
  private calculateElementsComplexity(elements: MetaContractElement[]): number {
    return elements.reduce((total, element) => {
      let complexity = 1; // 基础复杂度
      
      // 字段类型复杂度
      if (element.type === 'field') {
        if (element.properties.validation) complexity += 1;
        if (element.properties.unique) complexity += 1;
        if (element.properties.required) complexity += 0.5;
      }
      
      // 关系复杂度
      if (element.type === 'relationship') {
        complexity += 2;
      }
      
      return total + complexity;
    }, 0);
  }

  /**
   * 检查元素是否相等
   */
  private elementsEqual(element1: MetaContractElement, element2: MetaContractElement): boolean {
    return JSON.stringify(element1.properties) === JSON.stringify(element2.properties);
  }

  /**
   * 回滚模板应用
   */
  async rollbackTemplateApplication(backup: MetaContractElement[]): Promise<MetaContractElement[]> {
    if (!backup) {
      throw new Error('No backup available for rollback');
    }
    return JSON.parse(JSON.stringify(backup));
  }

  /**
   * 生成应用报告
   */
  generateApplicationReport(result: TemplateApplicationResult): string {
    const report = [];
    
    report.push(`Template Application Report`);
    report.push(`========================`);
    report.push(`Template: ${result.template.name} v${result.template.version}`);
    report.push(`Status: ${result.success ? 'SUCCESS' : 'FAILED'}`);
    report.push(`Applied Elements: ${result.appliedElements.length}`);
    
    if (result.conflicts && result.conflicts.length > 0) {
      report.push(`\\nConflicts Resolved: ${result.conflicts.length}`);
      result.conflicts.forEach(conflict => {
        report.push(`  - ${conflict.existing.name}: ${conflict.resolution}`);
      });
    }
    
    if (result.warnings && result.warnings.length > 0) {
      report.push(`\\nWarnings: ${result.warnings.length}`);
      result.warnings.forEach(warning => {
        report.push(`  - ${warning}`);
      });
    }
    
    if (result.performance) {
      report.push(`\\nPerformance Impact: ${result.performance.estimatedImpact}`);
      Object.entries(result.performance.metrics).forEach(([key, value]) => {
        report.push(`  ${key}: ${value}`);
      });
    }
    
    return report.join('\\n');
  }

  /**
   * 映射冲突类型到模板应用结果类型
   */
  private mapConflictType(conflictType: FieldConflictType): 'field' | 'relationship' | 'security' | 'validation' {
    switch (conflictType) {
      case FieldConflictType.NAME_CONFLICT:
      case FieldConflictType.TYPE_MISMATCH:
      case FieldConflictType.CONSTRAINT_CONFLICT:
        return 'field';
      case FieldConflictType.RELATIONSHIP_CONFLICT:
        return 'relationship';
      default:
        return 'field';
    }
  }

  /**
   * 映射解决策略到模板应用结果类型
   */
  private mapResolutionStrategy(strategy: ConflictResolutionStrategy): 'merge' | 'replace' | 'rename' | 'skip' {
    switch (strategy) {
      case ConflictResolutionStrategy.MERGE:
        return 'merge';
      case ConflictResolutionStrategy.REPLACE:
        return 'replace';
      case ConflictResolutionStrategy.RENAME:
        return 'rename';
      case ConflictResolutionStrategy.SKIP:
        return 'skip';
      case ConflictResolutionStrategy.ASK_USER:
        return 'skip'; // 默认跳过需要用户决策的项目
      default:
        return 'skip';
    }
  }
}