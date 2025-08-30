/**
 * 内联新增版本表单组件
 * 集成到右侧详情区域，替代Modal弹窗，提升用户体验
 */
import React, { useState, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  checkCircleIcon,    // 可用于组织单位（成功/重要）
  clockIcon,          // 可用于计划/项目团队
  timelineAllIcon,    // 可用于部门（历史/层级）
  exclamationCircleIcon // 警告图标
} from '@workday/canvas-system-icons-web';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { colors } from '@workday/canvas-kit-react/tokens';
import { type TemporalEditFormData } from './TemporalEditForm';
import { StatusBadge, type OrganizationStatus } from '../../../shared/components/StatusBadge';
// 移除违反原则13的EnhancedTemporalDataTable组件导入

// 添加映射函数
const mapLifecycleStatusToOrganizationStatus = (lifecycleStatus: string): OrganizationStatus => {
  switch (lifecycleStatus) {
    case 'CURRENT':
    case 'ACTIVE':
      return 'ACTIVE';
    case 'SUSPENDED':
      return 'SUSPENDED';
    case 'PLANNED':
      return 'PLANNED';
    default:
      return 'ACTIVE';
  }
};

export interface InlineNewVersionFormProps {
  organizationCode: string | null; // 允许null用于创建模式
  onSubmit: (data: TemporalEditFormData) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
  mode?: 'create' | 'edit' | 'insert'; // create: 新组织编码, edit: 编辑记录, insert: 插入新记录
  initialData?: {
    name: string;
    unitType: string;
    status: string;
    lifecycleStatus?: string;
    description?: string;
    parentCode?: string;
    effectiveDate?: string;
  } | null; // 允许null
  // 新增历史记录编辑相关props
  selectedVersion?: {
    recordId: string; // UUID唯一标识符
    createdAt: string;
    updatedAt: string;
    code: string;
    name: string;
    unitType: string;
    status: string;
    effectiveDate: string;
    description?: string;
    parentCode?: string;
  } | null;
  // 新增：传递所有版本数据用于日期范围验证
  allVersions?: Array<{
    recordId: string;
    effectiveDate: string;
    endDate?: string | null;
    isCurrent: boolean;
  }> | null;
  onEditHistory?: (versionData: Record<string, unknown>) => Promise<void>;
  onDeactivate?: (version: Record<string, unknown>) => Promise<void>; // 新增作废功能
  onInsertRecord?: (data: TemporalEditFormData) => Promise<void>; // 新增插入记录功能
  activeTab?: 'edit-history' | 'new-version' | 'audit-history'; // 当前选项卡状态
  onTabChange?: (tab: 'edit-history' | 'new-version' | 'audit-history') => void; // 选项卡切换
  // 版本数据相关props已移除 - 违反原则13
}

const unitTypeOptions = [
  { 
    label: '组织单位', 
    value: 'ORGANIZATION_UNIT',
    description: '企业的重要组织单位，负责特定职能和管理',
    color: colors.greenApple600,
    icon: checkCircleIcon    // 表示重要/核心地位
  },
  { 
    label: '部门', 
    value: 'DEPARTMENT',
    description: '企业内部的功能性组织单位，执行特定业务职能',
    color: colors.blueberry600,
    icon: timelineAllIcon    // 表示层级/结构
  },
  { 
    label: '项目团队', 
    value: 'PROJECT_TEAM',
    description: '临时性组织单位，专注于特定项目或任务的执行',
    color: colors.plum600,
    icon: clockIcon          // 表示时间性/计划性
  },
];


// 组织类型选择器组件（使用原生select暂时替代Canvas Kit Select）
const UnitTypeSelector: React.FC<{
  value: string;
  onChange: (value: string) => void;
  disabled?: boolean;
  label?: string;
  required?: boolean;
}> = ({
  value,
  onChange,
  disabled = false,
  label = '组织类型',
  required = false
}) => {
  const selectedOption = unitTypeOptions.find(opt => opt.value === value);

  return (
    <FormField isRequired={required}>
      <FormField.Label>
        {label} *
      </FormField.Label>
      <FormField.Field>
        <select
          value={value}
          onChange={(e) => {
            console.log('[UnitTypeSelector] 原生select变更:', value, '->', e.target.value);
            onChange(e.target.value);
          }}
          disabled={disabled}
          style={{ 
            width: '100%', 
            padding: '8px', 
            border: '1px solid #ddd', 
            borderRadius: '4px',
            fontSize: '14px'
          }}
        >
          {unitTypeOptions.map(option => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </FormField.Field>
      {selectedOption && (
        <FormField.Hint>
          {selectedOption.description}
        </FormField.Hint>
      )}
    </FormField>
  );
};

// 获取当月1日的日期字符串 (避免时区问题)
const getCurrentMonthFirstDay = () => {
  const now = new Date();
  const year = now.getFullYear();
  const month = now.getMonth() + 1; // getMonth() 返回0-11，需要+1
  const paddedMonth = month.toString().padStart(2, '0');
  return `${year}-${paddedMonth}-01`;
};


export const InlineNewVersionForm: React.FC<InlineNewVersionFormProps> = ({
  organizationCode,
  onSubmit,
  onCancel,
  isSubmitting = false,
  mode = 'create',
  initialData,
  selectedVersion,
  allVersions = null,
  onEditHistory,
  onDeactivate,
  onInsertRecord,
  activeTab = 'edit-history',
  onTabChange,
  // versions, onVersionSelect, onVersionEdit参数已移除 - 违反原则13
}) => {
  const [formData, setFormData] = useState<TemporalEditFormData>({
    name: '',
    unitType: 'DEPARTMENT',
    lifecycleStatus: 'PLANNED',
    description: '',
    effectiveDate: getCurrentMonthFirstDay(), // 默认当月1日
    parentCode: ''
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  
  // 历史记录编辑相关状态
  const [isEditingHistory, setIsEditingHistory] = useState(false);
  
  // 定义历史记录数据的接口类型
  interface OriginalHistoryData {
    recordId: string;
    createdAt: string;
    updatedAt: string;
    code: string;
    name: string;
    unitType: string;
    status: string;
    effectiveDate: string;
    description?: string;
    parentCode?: string;
  }
  
  const [originalHistoryData, setOriginalHistoryData] = useState<OriginalHistoryData | null>(null);
  
  // 作废功能相关状态
  const [showDeactivateConfirm, setShowDeactivateConfirm] = useState(false);
  const [isDeactivating, setIsDeactivating] = useState(false);
  
  // 移除表格视图切换功能（违反原则13）
  
  // Phase 7.3 - 用户体验改进状态
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);
  
  // 动态模式判断 - 根据activeTab确定当前模式
  const currentMode = activeTab === 'new-version' ? 'insert' : mode;
  
  // Modal管理
  const deactivateModalModel = useModalModel();

  // 同步Modal状态
  React.useEffect(() => {
    if (showDeactivateConfirm && deactivateModalModel.state.visibility !== 'visible') {
      deactivateModalModel.events.show();
    } else if (!showDeactivateConfirm && deactivateModalModel.state.visibility === 'visible') {
      deactivateModalModel.events.hide();
    }
  }, [showDeactivateConfirm, deactivateModalModel]);

  // Phase 7.3 - 自动清理成功消息和错误消息
  React.useEffect(() => {
    if (successMessage) {
      const timer = setTimeout(() => setSuccessMessage(null), 3000);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [successMessage]);

  React.useEffect(() => {
    if (error) {
      const timer = setTimeout(() => setError(null), 5000);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [error]);

  // 初始化表单数据 - 支持预填充模式和历史记录编辑
  useEffect(() => {
    const firstDayOfMonth = getCurrentMonthFirstDay();
    
    if ((mode === 'edit' || mode === 'insert') && initialData) {
      // 编辑模式 - 使用传入的初始数据预填充表单，包括原始生效日期
      setFormData({
        name: initialData.name,
        unitType: initialData.unitType,
        lifecycleStatus: initialData.status as 'SUSPENDED' | 'PLANNED' | 'DELETED' | 'CURRENT' | 'HISTORICAL' || 'PLANNED', // 修复：使用 status 字段而不是 lifecycleStatus
        description: initialData.description || '',
        effectiveDate: initialData.effectiveDate 
          ? new Date(initialData.effectiveDate).toISOString().split('T')[0] 
          : firstDayOfMonth, // 如果没有提供生效日期，使用当月1日
        parentCode: initialData.parentCode || ''
      });
      
      // 如果是编辑UUID记录模式，保存原始数据
      if (mode === 'edit' && selectedVersion) {
        setOriginalHistoryData(selectedVersion);
        setIsEditingHistory(false); // 初始时为只读模式
      }
    } else {
      // 新增模式 - 使用默认值
      setFormData({
        name: '',
        unitType: 'DEPARTMENT',
        lifecycleStatus: 'PLANNED',
        description: '',
        effectiveDate: firstDayOfMonth, // 默认当月1日生效
        parentCode: ''
      });
    }
    setErrors({});
  }, [mode, initialData, selectedVersion]);

  const handleInputChange = (field: keyof TemporalEditFormData) => (
    event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>
  ) => {
    const value = event.target.value;
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // 清除该字段的错误
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: '' }));
    }
  };

  // 计算编辑记录模式下的日期范围限制
  const getEditDateRange = (): { minDate: string | null; maxDate: string | null } => {
    // 只在编辑记录模式下才计算范围
    if (mode !== 'edit' || !selectedVersion || !allVersions || allVersions.length === 0) {
      return { minDate: null, maxDate: null };
    }

    // 按生效日期排序所有版本
    const sortedVersions = [...allVersions].sort((a, b) => 
      new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime()
    );

    // 找到当前编辑版本的索引
    const currentIndex = sortedVersions.findIndex(v => v.recordId === selectedVersion.recordId);
    if (currentIndex === -1) {
      return { minDate: null, maxDate: null };
    }

    // 获取前一条记录
    const previousVersion = currentIndex > 0 ? sortedVersions[currentIndex - 1] : null;
    
    // 获取后一条记录
    const nextVersion = currentIndex < sortedVersions.length - 1 ? sortedVersions[currentIndex + 1] : null;

    // 计算最小日期：前一条记录的生效日期的次日
    let minDate: string | null = null;
    if (previousVersion) {
      const prevDate = new Date(previousVersion.effectiveDate);
      prevDate.setDate(prevDate.getDate() + 1);
      minDate = prevDate.toISOString().split('T')[0];
    }

    // 计算最大日期：后一条记录的生效日期的前一日
    let maxDate: string | null = null;
    if (nextVersion) {
      const nextDate = new Date(nextVersion.effectiveDate);
      nextDate.setDate(nextDate.getDate() - 1);
      maxDate = nextDate.toISOString().split('T')[0];
    }

    return { minDate, maxDate };
  };

  // 计算插入新记录的日期范围限制
  const getInsertDateRange = (): { minDate: string | null; maxDate: string | null; insertType: 'history' | 'between' | 'future' } => {
    if (!allVersions || allVersions.length === 0) {
      return { minDate: null, maxDate: null, insertType: 'history' };
    }

    // 按生效日期排序所有版本
    const sortedVersions = [...allVersions].sort((a, b) => 
      new Date(a.effectiveDate).getTime() - new Date(b.effectiveDate).getTime()
    );

    const inputDate = new Date(formData.effectiveDate);
    const earliestDate = new Date(sortedVersions[0].effectiveDate);
    const latestDate = new Date(sortedVersions[sortedVersions.length - 1].effectiveDate);

    // 规则1: 不能插入早于最小生效日期的记录
    const minDate = sortedVersions[0].effectiveDate; // 最早的生效日期

    // 规则2: 如果是插入在两条记录之间
    if (inputDate > earliestDate && inputDate < latestDate) {
      // 找到应该插入的位置
      for (let i = 0; i < sortedVersions.length - 1; i++) {
        const currentDate = new Date(sortedVersions[i].effectiveDate);
        const nextDate = new Date(sortedVersions[i + 1].effectiveDate);
        
        if (inputDate > currentDate && inputDate < nextDate) {
          // 在这两个版本之间插入
          const minInsertDate = new Date(currentDate);
          minInsertDate.setDate(minInsertDate.getDate() + 1);
          
          const maxInsertDate = new Date(nextDate);
          maxInsertDate.setDate(maxInsertDate.getDate() - 1);
          
          return {
            minDate: minInsertDate.toISOString().split('T')[0],
            maxDate: maxInsertDate.toISOString().split('T')[0],
            insertType: 'between'
          };
        }
      }
    }

    // 规则3: 如果是插入未来日期的记录
    if (inputDate >= latestDate) {
      const futureMinDate = new Date(latestDate);
      futureMinDate.setDate(futureMinDate.getDate() + 1);
      return {
        minDate: futureMinDate.toISOString().split('T')[0],
        maxDate: null, // 未来记录没有最大日期限制
        insertType: 'future'
      };
    }

    // 如果是插入历史记录（在最早记录之前）
    return {
      minDate: minDate,
      maxDate: null,
      insertType: 'history'
    };
  };

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.name.trim()) {
      newErrors.name = '组织名称是必填项';
    }
    
    if (!formData.effectiveDate) {
      newErrors.effectiveDate = '生效日期是必填项';
    } else {
      if (currentMode === 'create') {
        // 对于完全新建组织单元，取消生效日期限制
        // 无任何限制
      } else if (currentMode === 'edit') {
        // 编辑记录模式：只要在前后两条记录的生效日期之间即可
        const { minDate, maxDate } = getEditDateRange();
        const effectiveDate = new Date(formData.effectiveDate);
        
        if (minDate) {
          const minDateTime = new Date(minDate);
          if (effectiveDate < minDateTime) {
            const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('zh-CN');
            newErrors.effectiveDate = `生效日期不能早于 ${formatDate(minDate)}（前一版本生效日期之后）`;
          }
        }
        
        if (maxDate && !newErrors.effectiveDate) {
          const maxDateTime = new Date(maxDate);
          if (effectiveDate > maxDateTime) {
            const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('zh-CN');
            newErrors.effectiveDate = `生效日期不能晚于 ${formatDate(maxDate)}（下一版本生效日期之前）`;
          }
        }
      } else if (currentMode === 'insert') {
        // 编辑组织信息模式 - 插入新记录模式
        const effectiveDate = new Date(formData.effectiveDate);
        
        if (!allVersions || allVersions.length === 0) {
          // 如果没有现有版本，无限制（相当于create模式）
        } else {
          // 使用新的插入规则
          const { minDate, maxDate, insertType } = getInsertDateRange();
          const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('zh-CN');
          
          // 规则1: 不能插入早于最小生效日期的记录
          if (minDate) {
            const minDateTime = new Date(minDate);
            if (effectiveDate < minDateTime) {
              if (insertType === 'history') {
                newErrors.effectiveDate = `生效日期不能早于 ${formatDate(minDate)}（最早版本生效日期）`;
              } else if (insertType === 'between') {
                newErrors.effectiveDate = `生效日期必须在 ${formatDate(minDate)} 之后（前一版本生效日期之后）`;
              } else if (insertType === 'future') {
                newErrors.effectiveDate = `生效日期必须在 ${formatDate(minDate)} 之后（最新版本生效日期之后）`;
              }
            }
          }
          
          // 规则2: 在两条记录之间插入时的最大日期限制
          if (maxDate && !newErrors.effectiveDate && insertType === 'between') {
            const maxDateTime = new Date(maxDate);
            if (effectiveDate > maxDateTime) {
              newErrors.effectiveDate = `生效日期必须在 ${formatDate(maxDate)} 之前（下一版本生效日期之前）`;
            }
          }
          
          // 规则3: 未来记录无最大日期限制，但提供友好提示
          if (insertType === 'future' && !newErrors.effectiveDate) {
            // 可以添加提示信息但不报错
            console.log(`插入未来记录：${formData.effectiveDate}`);
          }
        }
      }
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 历史记录编辑相关处理函数
  const handleEditHistoryToggle = () => {
    setIsEditingHistory(!isEditingHistory);
  };

  const handleEditHistorySubmit = async () => {
    if (!validateForm() || !onEditHistory || !originalHistoryData) {
      setError('表单验证失败或缺少必要数据，请重试');
      return;
    }

    // Phase 7.3 - 清理之前的消息
    setError(null);
    setSuccessMessage(null);
    setLoading(true);

    try {
      // 构建更新数据，包含ID和更新时间戳
      const updateData = {
        ...originalHistoryData,
        name: formData.name,
        unitType: formData.unitType,
        status: formData.lifecycleStatus,
        description: formData.description,
        effectiveDate: formData.effectiveDate,
        parentCode: formData.parentCode,
        updatedAt: new Date().toISOString()
      };
      
      await onEditHistory(updateData);
      setIsEditingHistory(false); // 提交后回到只读模式
      
      // Phase 7.3 - 显示成功消息
      setSuccessMessage('历史记录修改成功！');
      
    } catch (error) {
      console.error('修改历史记录失败:', error);
      
      // Phase 7.3 - 增强错误处理
      const errorMessage = error instanceof Error ? error.message : '修改失败，请重试';
      setError(`修改历史记录失败: ${errorMessage}`);
    } finally {
      // Phase 7.3 - 清理加载状态
      setLoading(false);
    }
  };

  const handleCancelEditHistory = () => {
    // 恢复原始数据
    if (originalHistoryData) {
      setFormData({
        name: originalHistoryData.name as string,
        unitType: originalHistoryData.unitType as string,
        lifecycleStatus: (originalHistoryData.status as 'SUSPENDED' | 'PLANNED' | 'DELETED' | 'CURRENT' | 'HISTORICAL') || 'PLANNED',
        description: (originalHistoryData.description as string) || '',
        effectiveDate: new Date(originalHistoryData.effectiveDate as string).toISOString().split('T')[0],
        parentCode: (originalHistoryData.parentCode as string) || ''
      });
    }
    setIsEditingHistory(false);
    setErrors({});
  };

  // 作废功能处理函数
  const handleDeactivateClick = () => {
    setShowDeactivateConfirm(true);
  };

  const handleDeactivateConfirm = async () => {
    if (!onDeactivate || !selectedVersion || isDeactivating) return;
    
    try {
      setIsDeactivating(true);
      await onDeactivate(selectedVersion);
      setShowDeactivateConfirm(false);
      // 作废成功后保持在当前页面，用户可以观察操作结果
      // 显示成功提示，让用户知道操作已完成
      setSuccessMessage(`版本删除成功！生效日期：${new Date(selectedVersion.effectiveDate).toLocaleDateString('zh-CN')}`);
      // 移除 onCancel() 调用，让用户自己决定是否离开页面
    } catch (error) {
      console.error('删除失败:', error);
      setError('删除失败，请重试');
    } finally {
      setIsDeactivating(false);
    }
  };

  const handleDeactivateCancel = () => {
    setShowDeactivateConfirm(false);
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    
    // Phase 7.3 - 清理之前的错误和成功消息
    setError(null);
    setSuccessMessage(null);
    
    console.log('[InlineNewVersionForm] 提交表单前的formData:', formData);
    
    if (!validateForm()) {
      setError('请检查表单中的错误项并重新提交');
      return;
    }
    
    // Phase 7.3 - 设置加载状态
    setLoading(true);
    
    try {
      // 根据当前模式调用不同的处理函数
      if (currentMode === 'insert' && onInsertRecord) {
        await onInsertRecord(formData);
      } else {
        await onSubmit(formData);
      }
      
      // Phase 7.3 - 显示成功消息
      setSuccessMessage(
        currentMode === 'create' ? '组织创建成功！' :
        currentMode === 'insert' ? '新版本记录插入成功！' :
        '记录修改成功！'
      );
      
    } catch (error) {
      console.error('提交表单失败:', error);
      
      // Phase 7.3 - 增强错误处理
      const errorMessage = error instanceof Error ? error.message : '操作失败，请重试';
      setError(`${currentMode === 'create' ? '创建组织失败' : 
                currentMode === 'insert' ? '插入记录失败' : 
                '修改记录失败'}: ${errorMessage}`);
    } finally {
      // Phase 7.3 - 清理加载状态
      setLoading(false);
    }
  };

  return (
    <Box flex="1">
      <Card padding="l">
        {/* 表单标题 */}
        <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
          <Box>
            <Heading size="medium" marginBottom="s">
              {currentMode === 'create' 
                ? '新建组织信息'
                : currentMode === 'edit' 
                  ? (isEditingHistory ? '编辑记录' : '查看记录')
                  : '插入新版本记录'}
            </Heading>
            <Text typeLevel="subtext.medium" color="hint">
              {currentMode === 'create'
                ? '填写新组织的基本信息，系统将自动分配组织编码'
                : currentMode === 'edit' 
                  ? (isEditingHistory ? `修改组织 ${organizationCode} 的记录信息` : 
                     `查看组织 ${organizationCode} 的记录信息`)
                  : currentMode === 'insert' 
                    ? `为组织 ${organizationCode} 插入新版本记录，将创建新的记录` 
                    : `为组织 ${organizationCode} 插入新版本记录`}
            </Text>
          </Box>
          
          {/* 移除视图切换器（违反原则13） */}
          
        </Flex>

        {/* Phase 7.3 - 错误和成功消息显示 */}
        {(error || successMessage) && (
          <Box marginBottom="l">
            {error && (
              <Box
                padding="m"
                backgroundColor={colors.cinnamon100}
                border={`1px solid ${colors.cinnamon600}`}
                borderRadius="4px"
                marginBottom="s"
              >
                <Flex alignItems="center" gap="s">
                  <SystemIcon icon={exclamationCircleIcon} color={colors.cinnamon600} size={20} />
                  <Text color={colors.cinnamon600} typeLevel="body.small" fontWeight="medium">
                    {error}
                  </Text>
                </Flex>
              </Box>
            )}
            
            {successMessage && (
              <Box
                padding="m"
                backgroundColor={colors.greenApple100}
                border={`1px solid ${colors.greenApple600}`}
                borderRadius="4px"
                marginBottom="s"
              >
                <Flex alignItems="center" gap="s">
                  <SystemIcon icon={checkCircleIcon} color={colors.greenApple600} size={20} />
                  <Text color={colors.greenApple600} typeLevel="body.small" fontWeight="medium">
                    {successMessage}
                  </Text>
                </Flex>
              </Box>
            )}
          </Box>
        )}

        {/* 历史记录元数据显示 - 移到最下方 */}

        {/* 表单视图 */}
        <form onSubmit={handleSubmit}>
          {/* 生效日期 - 最重要的信息放在最上方 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s" color={colors.blueberry600}>
              生效日期
            </Heading>
            
            <Box marginLeft="m">
              <FormField isRequired error={errors.effectiveDate ? "error" : undefined}>
                <FormField.Label>生效日期 *</FormField.Label>
                <FormField.Field>
                  <TextInput
                    type="date"
                    value={formData.effectiveDate}
                    onChange={handleInputChange('effectiveDate')}
                    disabled={isSubmitting || (currentMode === 'edit' && !isEditingHistory)}
                  />
                  {errors.effectiveDate && (
                    <FormField.Hint>{errors.effectiveDate}</FormField.Hint>
                  )}
                </FormField.Field>
              </FormField>
            </Box>
          </Box>

          {/* 基本信息 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s" color={colors.blueberry600}>
              基本信息
            </Heading>
            
            <Box marginLeft="m">
              <FormField isRequired error={errors.name ? "error" : undefined}>
                <FormField.Label>组织名称 *</FormField.Label>
                <FormField.Field>
                  <TextInput
                    value={formData.name}
                    onChange={handleInputChange('name')}
                    placeholder="请输入组织名称"
                    disabled={isSubmitting || (currentMode === 'edit' && !isEditingHistory)}
                  />
                  {errors.name && (
                    <FormField.Hint>{errors.name}</FormField.Hint>
                  )}
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>上级组织编码</FormField.Label>
                <FormField.Field>
                  <TextInput
                    value={formData.parentCode || ''}
                    onChange={handleInputChange('parentCode')}
                    placeholder="请输入上级组织编码（可选）"
                    disabled={isSubmitting || (currentMode === 'edit' && !isEditingHistory)}
                  />
                </FormField.Field>
              </FormField>

              <UnitTypeSelector
                value={formData.unitType}
                onChange={(newValue) => {
                  console.log('[InlineNewVersionForm] 组织类型选择变更:', formData.unitType, '->', newValue);
                  setFormData(prev => ({ ...prev, unitType: newValue }));
                  // 清除相关错误
                  if (errors.unitType) {
                    setErrors(prev => ({ ...prev, unitType: '' }));
                  }
                }}
                disabled={isSubmitting || (currentMode === 'edit' && !isEditingHistory)}
                label="组织类型"
                required={true}
              />

              <FormField>
                <FormField.Label>组织状态 *</FormField.Label>
                <FormField.Field>
                  <StatusBadge 
                    status={mapLifecycleStatusToOrganizationStatus(formData.lifecycleStatus)} 
                    size="medium"
                  />
                  <Text typeLevel="subtext.small" color="hint" marginTop="xs">
                    状态由系统根据操作自动管理
                  </Text>
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>描述信息</FormField.Label>
                <FormField.Field>
                  <TextArea
                    value={formData.description}
                    onChange={handleInputChange('description')}
                    placeholder="请输入组织描述信息"
                    disabled={isSubmitting || (currentMode === 'edit' && !isEditingHistory)}
                    rows={3}
                  />
                </FormField.Field>
              </FormField>
            </Box>
          </Box>

          {/* 记录信息 - 总是显示（当有数据时） */}
          {originalHistoryData && (
            <Box marginBottom="l" marginTop="l">
              <Heading size="small" marginBottom="s" color={colors.licorice600}>
                记录信息
              </Heading>
              <Box
                cs={{
                  display: "grid",
                  gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))",
                  gap: "12px"
                }}
              >
                <Box>
                  <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
                    记录UUID:
                  </Text>
                  <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700} style={{fontFamily: 'monospace'}}>
                    {originalHistoryData.recordId as string}
                  </Text>
                </Box>
                <Box>
                  <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
                    创建时间:
                  </Text>
                  <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700}>
                    {new Date(originalHistoryData.createdAt as string).toLocaleString('zh-CN')}
                  </Text>
                </Box>
                <Box>
                  <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
                    最后更新:
                  </Text>
                  <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700}>
                    {new Date(originalHistoryData.updatedAt as string).toLocaleString('zh-CN')}
                  </Text>
                </Box>
              </Box>
            </Box>
          )}

          {/* 操作按钮 */}
          <Box
            marginTop="xl"
            paddingTop="l"
            borderTop={`1px solid ${colors.soap300}`}
          >
            {currentMode === 'edit' ? (
              // 记录编辑模式的按钮
              <Flex gap="s" justifyContent="space-between">
                {/* 左侧作废按钮 */}
                <Box>
                  {selectedVersion && !isEditingHistory && (
                    <TertiaryButton 
                      onClick={handleDeactivateClick}
                      disabled={isSubmitting || isDeactivating}
                    >
                      删除此记录
                    </TertiaryButton>
                  )}
                </Box>
                
                {/* 右侧主要操作按钮 */}
                <Flex gap="s">
                  {!isEditingHistory ? (
                    // 只读模式的按钮 - 调整顺序和样式：插入新记录、修改记录、关闭
                    <>
                      <SecondaryButton 
                        onClick={() => {
                          if (onTabChange) {
                            onTabChange('new-version');
                          }
                        }}
                        disabled={isSubmitting || loading}
                      >
                        插入新记录
                      </SecondaryButton>
                      <SecondaryButton 
                        onClick={handleEditHistoryToggle}
                        disabled={isSubmitting || loading}
                      >
                        修改记录
                      </SecondaryButton>
                      <PrimaryButton 
                        onClick={onCancel}
                        disabled={isSubmitting}
                      >
                        关闭
                      </PrimaryButton>
                    </>
                  ) : (
                    // 编辑模式的按钮
                    <>
                      <SecondaryButton 
                        onClick={handleCancelEditHistory}
                        disabled={isSubmitting || loading}
                      >
                        取消编辑
                      </SecondaryButton>
                      <PrimaryButton 
                        onClick={handleEditHistorySubmit}
                        disabled={isSubmitting || loading}
                      >
                        {/* Phase 7.3 - 增强加载状态显示 */}
                        {(isSubmitting || loading) ? '提交中...' : '提交修改'}
                      </PrimaryButton>
                    </>
                  )}
                </Flex>
              </Flex>
            ) : (
              // 原有的新增/编辑版本模式的按钮
              <Flex gap="s" justifyContent="flex-end">
                <SecondaryButton 
                  onClick={onCancel}
                  disabled={isSubmitting || loading}
                >
                  取消
                </SecondaryButton>
                <PrimaryButton 
                  type="submit"
                  disabled={isSubmitting || loading}
                >
                  {/* Phase 7.3 - 增强加载状态显示 */}
                  {(isSubmitting || loading) 
                    ? (currentMode === 'create' ? '创建中...' : currentMode === 'insert' ? '插入中...' : '修改中...')
                    : (currentMode === 'create' ? '创建组织' : currentMode === 'insert' ? '插入新记录' : '提交修改')}
                </PrimaryButton>
              </Flex>
            )}
          </Box>
        </form>
      </Card>

      {/* 作废确认对话框 */}
      {showDeactivateConfirm && selectedVersion && (
        <Modal model={deactivateModalModel}>
          <Modal.Overlay>
            <Modal.Card>
              <Modal.CloseIcon onClick={handleDeactivateCancel} />
              <Modal.Heading>确认删除版本</Modal.Heading>
              <Modal.Body>
                <Box padding="l">
                  <Flex alignItems="flex-start" gap="m" marginBottom="l">
                    <SystemIcon icon={exclamationCircleIcon} size={24} color={colors.cinnamon600} />
                    <Box>
                      <Text typeLevel="body.medium" marginBottom="s">
                        确定要删除生效日期为 <strong>{new Date(selectedVersion.effectiveDate).toLocaleDateString('zh-CN')}</strong> 的版本吗？
                      </Text>
                      <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                        版本名称: {selectedVersion.name}
                      </Text>
                      <Text typeLevel="subtext.small" color={colors.cinnamon600}>
                        删除后记录将标记为已删除状态，此操作不可撤销
                      </Text>
                    </Box>
                  </Flex>
                  
                  <Flex gap="s" justifyContent="flex-end">
                    <SecondaryButton 
                      onClick={handleDeactivateCancel}
                      disabled={isDeactivating}
                    >
                      取消
                    </SecondaryButton>
                    <PrimaryButton 
                      onClick={handleDeactivateConfirm}
                      disabled={isDeactivating}
                    >
                      {isDeactivating ? '删除中...' : '确认删除'}
                    </PrimaryButton>
                  </Flex>
                </Box>
              </Modal.Body>
            </Modal.Card>
          </Modal.Overlay>
        </Modal>
      )}
    </Box>
  );
};

export default InlineNewVersionForm;