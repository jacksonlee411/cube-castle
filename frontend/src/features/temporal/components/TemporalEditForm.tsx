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
import { FiveStateStatusSelector, LIFECYCLE_STATES, type LifecycleStatus } from './FiveStateStatusSelector';

export interface TemporalEditFormData {
  name: string;
  unit_type: string;
  lifecycle_status: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'SUSPENDED' | 'DELETED';
  description?: string;
  effective_date: string;
  parent_code?: string;
  change_reason?: string;
  event_type?: string;
}

export interface TemporalVersion {
  code: string;
  name: string;
  unit_type: string;
  lifecycle_status: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'SUSPENDED' | 'DELETED';
  description?: string;
  effective_date: string;
  change_reason?: string;
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
  { label: '公司', value: 'COMPANY' },
  { label: '部门', value: 'DEPARTMENT' },
  { label: '成本中心', value: 'COST_CENTER' },
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
    unit_type: 'DEPARTMENT',
    lifecycle_status: 'PLANNED',
    description: '',
    effective_date: new Date().toISOString().split('T')[0], // 默认今天
    change_reason: '',
    event_type: 'UPDATE'
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

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
          unit_type: initialData.unit_type,
          lifecycle_status: initialData.lifecycle_status,
          description: initialData.description || '',
          effective_date: new Date(initialData.effective_date).toISOString().split('T')[0],
          change_reason: initialData.change_reason || '',
          event_type: 'UPDATE'
        });
      } else {
        // 创建模式 - 使用默认值
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        setFormData({
          name: '',
          unit_type: 'DEPARTMENT',
          lifecycle_status: 'PLANNED',
          description: '',
          effective_date: tomorrow.toISOString().split('T')[0], // 默认明天生效
          change_reason: '',
          event_type: 'RESTRUCTURE'
        });
      }
      setErrors({});
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
    
    if (!formData.effective_date) {
      newErrors.effective_date = '生效日期是必填项';
    } else {
      const effectiveDate = new Date(formData.effective_date);
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      
      if (effectiveDate < today) {
        newErrors.effective_date = '生效日期不能早于今天';
      }
    }
    
    if (!formData.change_reason.trim()) {
      newErrors.change_reason = '变更原因是必填项';
    }
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    
    if (!validateForm()) {
      return;
    }
    
    try {
      await onSubmit(formData);
    } catch (error) {
      console.error('提交表单失败:', error);
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

            <FormField isRequired>
              <FormField.Label>组织类型</FormField.Label>
              <FormField.Field>
                <select
                  value={formData.unit_type}
                  onChange={handleInputChange('unit_type')}
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

            <FiveStateStatusSelector
              value={formData.lifecycle_status}
              onChange={(status: LifecycleStatus) => {
                setFormData(prev => ({ ...prev, lifecycle_status: status.key }));
                // 清除错误
                if (errors.lifecycle_status) {
                  setErrors(prev => ({ ...prev, lifecycle_status: '' }));
                }
              }}
              disabled={isSubmitting}
              includeDeleted={false}
              label="组织状态"
              required={true}
              error={errors.lifecycle_status}
            />

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
              error={errors.effective_date ? "error" : undefined}
            >
              <FormField.Label>生效日期</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextInput}
                  type="date"
                  value={formData.effective_date}
                  onChange={handleInputChange('effective_date')}
                  disabled={isSubmitting}
                />
                {errors.effective_date && (
                  <FormField.Hint>{errors.effective_date}</FormField.Hint>
                )}
              </FormField.Field>
            </FormField>

            <FormField isRequired>
              <FormField.Label>事件类型</FormField.Label>
              <FormField.Field>
                <select
                  value={formData.event_type}
                  onChange={handleInputChange('event_type')}
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
              error={errors.change_reason ? "error" : undefined}
            >
              <FormField.Label>变更原因</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextArea}
                  value={formData.change_reason}
                  onChange={handleInputChange('change_reason')}
                  placeholder="请说明此次变更的原因"
                  disabled={isSubmitting}
                  rows={2}
                />
                {errors.change_reason && (
                  <FormField.Hint>{errors.change_reason}</FormField.Hint>
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