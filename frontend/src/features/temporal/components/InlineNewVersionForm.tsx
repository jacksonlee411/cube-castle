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
import ParentOrganizationSelector from './ParentOrganizationSelector';
// 移除违反原则13的EnhancedTemporalDataTable组件导入

// 添加映射函数
const mapLifecycleStatusToOrganizationStatus = (lifecycleStatus: string): OrganizationStatus => {
  switch (lifecycleStatus) {
    case 'CURRENT':
    case 'ACTIVE':
      return 'ACTIVE';
    case 'INACTIVE':
      return 'INACTIVE';
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
  mode?: 'create' | 'edit'; // create: 新组织编码, edit: 编辑记录
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
    level?: number;
    path?: string | null;
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
  hierarchyPaths?: { codePath: string; namePath: string } | null;
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
  onInsertRecord: _onInsertRecord,
  activeTab: _activeTab = 'edit-history',
  onTabChange: _onTabChange,
  hierarchyPaths = null,
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
  const [parentError, setParentError] = useState<string>('');
  
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
  const currentMode = mode;

  const levelDisplay = selectedVersion?.level;
  const codePathDisplay = React.useMemo(() => {
    if (hierarchyPaths?.codePath) return hierarchyPaths.codePath;
    if (selectedVersion?.path) return selectedVersion.path;
    return '';
  }, [hierarchyPaths, selectedVersion]);

  const namePathDisplay = React.useMemo(() => {
    if (hierarchyPaths?.namePath) return hierarchyPaths.namePath;
    return '';
  }, [hierarchyPaths]);
  
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
    
    if (mode === 'edit' && initialData) {
      // 编辑模式 - 使用传入的初始数据预填充表单，包括原始生效日期
      setFormData({
        name: initialData.name,
        unitType: initialData.unitType,
        lifecycleStatus: initialData.status as 'INACTIVE' | 'PLANNED' | 'DELETED' | 'CURRENT' | 'HISTORICAL' || 'PLANNED', // 修复：使用 status 字段而不是 lifecycleStatus
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
    setParentError('');
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

  const handleParentOrganizationChange = (parentCode: string | undefined) => {
    setFormData(prev => ({ ...prev, parentCode: parentCode ?? '' }));
    if (parentError) {
      setParentError('');
    }
  };

  const handleParentOrganizationError = (message: string) => {
    setParentError(message);
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
      }
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 历史记录编辑相关处理函数
  const handleEditHistoryToggle = () => {
    if (!isEditingHistory && selectedVersion) {
      // 进入修改模式：设置原始数据用于修改
      setOriginalHistoryData(selectedVersion);
      setFormData({
        name: selectedVersion.name,
        unitType: selectedVersion.unitType,
        lifecycleStatus: (selectedVersion.status as 'INACTIVE' | 'PLANNED' | 'DELETED' | 'CURRENT' | 'HISTORICAL') || 'CURRENT',
        description: selectedVersion.description || '',
        effectiveDate: new Date(selectedVersion.effectiveDate).toISOString().split('T')[0],
        parentCode: selectedVersion.parentCode || ''
      });
      setParentError('');
    }
    setIsEditingHistory(!isEditingHistory);
  };

  const handleEditHistorySubmit = async () => {
    if (parentError) {
      setError('请先修正上级组织选择');
      return;
    }

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
        lifecycleStatus: (originalHistoryData.status as 'INACTIVE' | 'PLANNED' | 'DELETED' | 'CURRENT' | 'HISTORICAL') || 'PLANNED',
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
      setError(null); // 清除之前的错误消息
      setSuccessMessage(null); // 清除之前的成功消息
      
      await onDeactivate(selectedVersion);
      
      // onDeactivate成功完成，关闭确认对话框
      setShowDeactivateConfirm(false);
      // 不在这里显示成功消息，让父组件处理成功状态
    } catch (error) {
      console.error('删除失败:', error);
      setShowDeactivateConfirm(false); // 关闭确认对话框
      
      // 提供更友好的错误消息
      const errorMessage = error instanceof Error ? error.message : '删除失败，请重试';
      setError(errorMessage);
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

    if (parentError) {
      setError('请先修正上级组织选择');
      return;
    }

    if (!validateForm()) {
      setError('请检查表单中的错误项并重新提交');
      return;
    }
    
    // Phase 7.3 - 设置加载状态
    setLoading(true);
    
    try {
      // 统一使用onSubmit处理
      await onSubmit(formData);
      
      // Phase 7.3 - 显示成功消息
      setSuccessMessage(
        currentMode === 'create' ? '组织创建成功！' : '版本记录保存成功！'
      );
      
    } catch (error) {
      console.error('提交表单失败:', error);
      
      // Phase 7.3 - 增强错误处理
      const errorMessage = error instanceof Error ? error.message : '操作失败，请重试';
      setError(`${currentMode === 'create' ? '创建组织失败' : '保存记录失败'}: ${errorMessage}`);
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
                  ? (isEditingHistory ? 
                     (originalHistoryData ? '修改版本记录' : '插入新版本记录') : 
                     '查看版本记录')
                  : '版本记录管理'}
            </Heading>
            <Text typeLevel="subtext.medium" color="hint">
              {currentMode === 'create'
                ? '填写新组织的基本信息，系统将自动分配组织编码'
                : currentMode === 'edit' 
                  ? (isEditingHistory ? 
                     (originalHistoryData ? 
                      `修改组织 ${organizationCode} 的现有版本记录` : 
                      `为组织 ${organizationCode} 插入新的版本记录`) :
                     `查看组织 ${organizationCode} 的版本记录信息`)
                    : `为组织 ${organizationCode} 管理版本记录`}
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

              <Box marginTop="m">
                <ParentOrganizationSelector
                  currentCode={organizationCode ?? ''}
                  effectiveDate={formData.effectiveDate}
                  currentParentCode={formData.parentCode}
                  onChange={handleParentOrganizationChange}
                  onValidationError={handleParentOrganizationError}
                  disabled={isSubmitting || (currentMode === 'edit' && !isEditingHistory)}
                />
                {parentError && (
                  <Text typeLevel="subtext.small" color="error" marginTop="xs">
                    {parentError}
                  </Text>
                )}
                <Text typeLevel="subtext.small" color="hint" marginTop="xs">
                  仅允许选择在生效日期有效且状态为 ACTIVE 的组织
                </Text>
              </Box>

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

          {currentMode === 'edit' && selectedVersion && (
            <Box marginBottom="l">
              <Heading size="small" marginBottom="s" color={colors.blueberry600}>
                层级与路径
              </Heading>
              <Box marginLeft="m">
                <FormField>
                  <FormField.Label>组织层级</FormField.Label>
                  <FormField.Field>
                    <TextInput
                      value={levelDisplay !== undefined ? String(levelDisplay) : '—'}
                      disabled
                    />
                  </FormField.Field>
                  <FormField.Hint>层级由后端计算，不可编辑</FormField.Hint>
                </FormField>

                <Box marginTop="m">
                  <FormField>
                    <FormField.Label>组织路径（编码）</FormField.Label>
                    <FormField.Field>
                      <TextInput
                        value={codePathDisplay.trim() || '路径数据暂不可用'}
                        disabled
                      />
                    </FormField.Field>
                    <FormField.Hint>统一 codePath，已与顶部复制按钮联动</FormField.Hint>
                  </FormField>
                </Box>

                <Box marginTop="m">
                  <FormField>
                    <FormField.Label>组织路径（名称）</FormField.Label>
                    <FormField.Field>
                      <TextInput
                        value={namePathDisplay.trim() || '路径数据暂不可用'}
                        disabled
                      />
                    </FormField.Field>
                    <FormField.Hint>读取 GraphQL namePath，提供可读的路径描述</FormField.Hint>
                  </FormField>
                </Box>
              </Box>
            </Box>
          )}

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
                          // 重置表单为插入新版本模式
                          setFormData({
                            name: selectedVersion?.name || '',
                            unitType: selectedVersion?.unitType || 'ORGANIZATION_UNIT',
                            lifecycleStatus: 'CURRENT',
                            description: selectedVersion?.description || '',
                            parentCode: selectedVersion?.parentCode || '',
                            effectiveDate: getCurrentMonthFirstDay()
                          });
                          setErrors({});
                          setOriginalHistoryData(null); // 清空原始数据，标记为插入模式
                          setIsEditingHistory(true); // 进入编辑模式以显示表单
                        }}
                        disabled={isSubmitting || loading}
                      >
                        插入新版本
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
                        onClick={originalHistoryData ? handleEditHistorySubmit : handleSubmit}
                        disabled={isSubmitting || loading}
                      >
                        {(isSubmitting || loading) ? '提交中...' : 
                         originalHistoryData ? '提交修改' : '插入新版本'}
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
                    ? (currentMode === 'create' ? '创建中...' : '保存中...')
                    : (currentMode === 'create' ? '创建组织' : isEditingHistory ? '保存修改' : '保存新版本')}
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
