/**
 * 组织详情编辑表单组件
 * 支持创建和编辑时态版本
 */
import React, { useState, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { StatusBadge, type OrganizationStatus } from '../../../shared/components/StatusBadge';
import ParentOrganizationSelector from './ParentOrganizationSelector';

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

export interface TemporalEditFormData {
  name: string;
  unitType: string;
  lifecycleStatus: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'INACTIVE' | 'DELETED';
  description?: string;
  effectiveDate: string;
  parentCode?: string;
  changeReason?: string;
  eventType?: string;
}

export interface TemporalVersion {
  code: string;
  name: string;
  unitType: string;
  lifecycleStatus: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'INACTIVE' | 'DELETED';
  description?: string;
  effectiveDate: string;
  parentCode?: string;  // 修复：添加上级组织编码字段
  changeReason?: string;
}

export interface TemporalEditFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: TemporalEditFormData) => Promise<void>;
  organizationCode: string;
  initialData?: TemporalVersion | null;
  mode: 'create' | 'edit';
  isSubmitting?: boolean;
}

const unitTypeOptions = [
  { label: '组织单位', value: 'ORGANIZATION_UNIT' },
  { label: '部门', value: 'DEPARTMENT' },
  { label: '项目团队', value: 'PROJECT_TEAM' },
];

const eventTypeOptions = [
  { label: '更新', value: 'UPDATE' },
  { label: '重组', value: 'RESTRUCTURE' },
  { label: '撤销', value: 'DISSOLVE' },
];

export const TemporalEditForm: React.FC<TemporalEditFormProps> = ({
  isOpen,
  onClose,
  onSubmit,
  organizationCode,
  initialData,
  mode,
  isSubmitting = false
}) => {
  const [formData, setFormData] = useState<TemporalEditFormData>({
    name: '',
    unitType: 'DEPARTMENT',
    lifecycleStatus: 'PLANNED',
    description: '',
    effectiveDate: new Date().toISOString().split('T')[0], // 默认今天
    parentCode: '', // 修复：添加上级组织编码字段
    changeReason: '',
    eventType: 'UPDATE'
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  const [parentError, setParentError] = useState<string>('');
  const [suggestedEffectiveDate, setSuggestedEffectiveDate] = useState<string | undefined>(undefined);

  const handleParentOrganizationChange = (parentCode: string | undefined) => {
    setFormData(prev => ({ ...prev, parentCode: parentCode ?? '' }));
    if (parentError) {
      setParentError('');
    }
    setSuggestedEffectiveDate(undefined);
  };

  const handleParentOrganizationError = (message?: string) => {
    setParentError(message ?? '');
    if (!message) {
      setSuggestedEffectiveDate(undefined);
    }
  };

  const handleParentTemporalError = (error: unknown): boolean => {
    const apiError = error as { code?: string; message?: string; details?: unknown } | undefined;
    if (apiError?.code !== 'TEMPORAL_PARENT_UNAVAILABLE') {
      return false;
    }

    let message = typeof apiError.message === 'string' ? apiError.message : '上级组织在指定日期不可用';
    let suggested: string | undefined;

    if (Array.isArray(apiError.details)) {
      const detail = apiError.details.find((item: any) => item?.code === 'TEMPORAL_PARENT_UNAVAILABLE') as
        | { message?: string; context?: { suggestedDate?: string } }
        | undefined;
      if (detail?.message && typeof detail.message === 'string') {
        message = detail.message;
      }
      const candidate = detail?.context?.suggestedDate;
      if (typeof candidate === 'string' && candidate.trim().length > 0) {
        suggested = candidate;
      }
    }

    setParentError(message);
    setSuggestedEffectiveDate(suggested);
    return true;
  };

  // Modal model
  const model = useModalModel();

  // 同步Modal状态
  React.useEffect(() => {
    if (isOpen && model.state.visibility !== 'visible') {
      model.events.show();
    } else if (!isOpen && model.state.visibility === 'visible') {
      model.events.hide();
    }
  }, [isOpen, model]);

  // 初始化表单数据
  useEffect(() => {
    if (isOpen) {
      if (mode === 'edit' && initialData) {
        setFormData({
          name: initialData.name,
          unitType: initialData.unitType,
          lifecycleStatus: initialData.lifecycleStatus,
          description: initialData.description || '',
          effectiveDate: new Date(initialData.effectiveDate).toISOString().split('T')[0],
          parentCode: initialData.parentCode || '', // 修复：添加上级组织编码初始化
          changeReason: initialData.changeReason || '',
          eventType: 'UPDATE'
        });
      } else {
        // 创建模式 - 使用默认值
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        setFormData({
          name: '',
          unitType: 'DEPARTMENT',
          lifecycleStatus: 'PLANNED',
          description: '',
          effectiveDate: tomorrow.toISOString().split('T')[0], // 默认明天生效
          parentCode: '', // 修复：添加上级组织编码默认值
          changeReason: '',
          eventType: 'RESTRUCTURE'
        });
      }
      setErrors({});
      setParentError('');
    }
  }, [isOpen, mode, initialData]);

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

  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};
    
    if (!formData.name.trim()) {
      newErrors.name = '组织名称是必填项';
    }
    
    if (!formData.effectiveDate) {
      newErrors.effectiveDate = '生效日期是必填项';
    } else {
      const effectiveDate = new Date(formData.effectiveDate);
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      
      if (effectiveDate < today) {
        newErrors.effectiveDate = '生效日期不能早于今天';
      }
    }
    
    if (!formData.changeReason?.trim()) {
      newErrors.changeReason = '变更原因是必填项';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    
    if (parentError) {
      return;
    }

    if (!validateForm()) {
      return;
    }
    
    try {
      await onSubmit(formData);
    } catch (error) {
      console.error('提交表单失败:', error);
      if (!handleParentTemporalError(error)) {
        setSuggestedEffectiveDate(undefined);
        if (error instanceof Error) {
          setParentError(error.message);
        } else {
          setParentError('提交失败，请稍后重试');
        }
      }
    }
  };

  const handleClose = () => {
    if (!isSubmitting) {
      model.events.hide();
      onClose();
    }
  };

  // 只在isOpen为true时才渲染Modal
  if (!isOpen) {
    return null;
  }

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card>
          <Modal.CloseIcon onClick={handleClose} />
          <Modal.Heading>{mode === 'create' ? '新增时态版本' : '编辑时态版本'}</Modal.Heading>
          <Modal.Body>
          <Box padding="l" width="600px">
            <form onSubmit={handleSubmit}>
          <Box marginBottom="l">
            <Text typeLevel="body.small" color="hint">
              组织编码: {organizationCode}
            </Text>
          </Box>

          {/* 基本信息 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s">基本信息</Heading>
            
            <FormField
              isRequired
              error={errors.name ? "error" : undefined}
            >
              <FormField.Label>组织名称</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextInput}
                  value={formData.name}
                  onChange={handleInputChange('name')}
                  placeholder="请输入组织名称"
                  disabled={isSubmitting}
                />
                {errors.name && (
                  <FormField.Hint>{errors.name}</FormField.Hint>
                )}
              </FormField.Field>
            </FormField>

            <Box marginTop="m">
              <ParentOrganizationSelector
                currentCode={organizationCode}
                effectiveDate={formData.effectiveDate}
                currentParentCode={formData.parentCode}
                onChange={handleParentOrganizationChange}
                onValidationError={handleParentOrganizationError}
                disabled={isSubmitting}
              />
              {parentError && (
                <Text typeLevel="subtext.small" color="error" marginTop="xs">
                  {parentError}
                </Text>
              )}
              {suggestedEffectiveDate && (
                <Flex gap="s" marginTop="xs">
                  <SecondaryButton
                    type="button"
                    onClick={() => {
                      setFormData(prev => ({ ...prev, effectiveDate: suggestedEffectiveDate }));
                      setSuggestedEffectiveDate(undefined);
                      setParentError('');
                    }}
                    disabled={isSubmitting}
                  >
                    调整生效日期至 {suggestedEffectiveDate}
                  </SecondaryButton>
                  <SecondaryButton
                    type="button"
                    onClick={() => handleParentOrganizationChange(undefined)}
                    disabled={isSubmitting}
                  >
                    重新选择上级组织
                  </SecondaryButton>
                </Flex>
              )}
            </Box>

            <FormField isRequired>
              <FormField.Label>组织类型</FormField.Label>
              <FormField.Field>
                <select
                  value={formData.unitType}
                  onChange={handleInputChange('unitType')}
                  disabled={isSubmitting}
                  style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
                >
                  {unitTypeOptions.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </FormField.Field>
            </FormField>

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
                <FormField.Input
                  as={TextArea}
                  value={formData.description}
                  onChange={handleInputChange('description')}
                  placeholder="请输入组织描述信息"
                  disabled={isSubmitting}
                  rows={3}
                />
              </FormField.Field>
            </FormField>
          </Box>

          {/* 时态信息 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s">时态信息</Heading>

            <FormField
              isRequired
              error={errors.effectiveDate ? "error" : undefined}
            >
              <FormField.Label>生效日期</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextInput}
                  type="date"
                  value={formData.effectiveDate}
                  onChange={handleInputChange('effectiveDate')}
                  disabled={isSubmitting}
                />
                {errors.effectiveDate && (
                  <FormField.Hint>{errors.effectiveDate}</FormField.Hint>
                )}
              </FormField.Field>
            </FormField>

            <FormField isRequired>
              <FormField.Label>事件类型</FormField.Label>
              <FormField.Field>
                <select
                  value={formData.eventType}
                  onChange={handleInputChange('eventType')}
                  disabled={isSubmitting}
                  style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
                >
                  {eventTypeOptions.map(option => (
                    <option key={option.value} value={option.value}>
                      {option.label}
                    </option>
                  ))}
                </select>
              </FormField.Field>
            </FormField>

            <FormField
              isRequired
              error={errors.changeReason ? "error" : undefined}
            >
              <FormField.Label>变更原因</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextArea}
                  value={formData.changeReason}
                  onChange={handleInputChange('changeReason')}
                  placeholder="请说明此次变更的原因"
                  disabled={isSubmitting}
                  rows={2}
                />
                {errors.changeReason && (
                  <FormField.Hint>{errors.changeReason}</FormField.Hint>
                )}
              </FormField.Field>
            </FormField>
          </Box>

          {/* 操作按钮 */}
          <Flex gap="s" justifyContent="flex-end">
            <SecondaryButton 
              onClick={handleClose}
              disabled={isSubmitting}
            >
              取消
            </SecondaryButton>
            <PrimaryButton 
              type="submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? '提交中...' : mode === 'create' ? '创建' : '更新'}
            </PrimaryButton>
          </Flex>
            </form>
          </Box>
        </Modal.Body>
      </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};

export default TemporalEditForm;
