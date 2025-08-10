import React, { useState, useCallback } from 'react';
import { Modal } from '@workday/canvas-kit-react/modal';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { Flex } from '@workday/canvas-kit-react/layout';
import { TemporalDatePicker, validateTemporalDate } from './TemporalDatePicker';
import { TemporalStatusSelector } from './TemporalStatusSelector';
import type { TemporalStatus } from './TemporalStatusSelector';
import { UnitTypeSelector, UnitType } from '../../organizations/components/OrganizationForm/FormFields';

export interface PlannedOrganizationData {
  name: string;
  unit_type: UnitType;
  description?: string;
  effective_date: string;
  end_date?: string;
  change_reason: string;
  parent_code?: string;
}

export interface PlannedOrganizationFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: PlannedOrganizationData) => Promise<void>;
  loading?: boolean;
  parentOrganization?: {
    code: string;
    name: string;
  };
}

interface FormErrors {
  name?: string;
  unit_type?: string;
  effective_date?: string;
  end_date?: string;
  change_reason?: string;
  general?: string;
}

export const PlannedOrganizationForm: React.FC<PlannedOrganizationFormProps> = ({
  isOpen,
  onClose,
  onSubmit,
  loading = false,
  parentOrganization,
}) => {
  const [formData, setFormData] = useState<PlannedOrganizationData>({
    name: '',
    unit_type: 'DEPARTMENT',
    description: '',
    effective_date: '',
    end_date: '',
    change_reason: '',
    parent_code: parentOrganization?.code,
  });

  const [errors, setErrors] = useState<FormErrors>({});

  // 重置表单
  const resetForm = useCallback(() => {
    setFormData({
      name: '',
      unit_type: 'DEPARTMENT',
      description: '',
      effective_date: '',
      end_date: '',
      change_reason: '',
      parent_code: parentOrganization?.code,
    });
    setErrors({});
  }, [parentOrganization?.code]);

  // 表单验证
  const validateForm = (): boolean => {
    const newErrors: FormErrors = {};

    // 名称验证
    if (!formData.name.trim()) {
      newErrors.name = '组织名称不能为空';
    } else if (formData.name.length < 2) {
      newErrors.name = '组织名称至少需要2个字符';
    } else if (formData.name.length > 100) {
      newErrors.name = '组织名称不能超过100个字符';
    }

    // 组织类型验证
    if (!formData.unit_type) {
      newErrors.unit_type = '请选择组织类型';
    }

    // 生效日期验证
    if (!formData.effective_date) {
      newErrors.effective_date = '生效日期不能为空';
    } else if (!validateTemporalDate.isValidDate(formData.effective_date)) {
      newErrors.effective_date = '请输入有效的日期格式';
    } else if (!validateTemporalDate.isFutureDate(formData.effective_date)) {
      newErrors.effective_date = '计划组织的生效日期必须在当前日期之后';
    }

    // 结束日期验证
    if (formData.end_date) {
      if (!validateTemporalDate.isValidDate(formData.end_date)) {
        newErrors.end_date = '请输入有效的日期格式';
      } else if (formData.effective_date && !validateTemporalDate.isEndDateAfterStartDate(formData.effective_date, formData.end_date)) {
        newErrors.end_date = '结束日期必须在生效日期之后';
      }
    }

    // 变更原因验证
    if (!formData.change_reason.trim()) {
      newErrors.change_reason = '变更原因不能为空';
    } else if (formData.change_reason.length < 5) {
      newErrors.change_reason = '变更原因至少需要5个字符';
    } else if (formData.change_reason.length > 500) {
      newErrors.change_reason = '变更原因不能超过500个字符';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 处理字段变更
  const handleFieldChange = useCallback((field: keyof PlannedOrganizationData, value: string) => {
    setFormData(prev => ({
      ...prev,
      [field]: value,
    }));

    // 清除相关错误
    if (errors[field as keyof FormErrors]) {
      setErrors(prev => ({
        ...prev,
        [field]: undefined,
      }));
    }
  }, [errors]);

  // 提交处理
  const handleSubmit = async () => {
    if (!validateForm()) return;

    try {
      await onSubmit(formData);
      resetForm();
      onClose();
    } catch (error) {
      setErrors({
        general: error instanceof Error ? error.message : '创建计划组织失败',
      });
    }
  };

  // 取消处理
  const handleCancel = () => {
    resetForm();
    onClose();
  };

  const minDate = validateTemporalDate.getTodayString();

  return (
    <Modal isOpen={isOpen} onClose={handleCancel}>
      <Modal.Card style={{ width: '600px', maxWidth: '90vw' }}>
        <Modal.CloseIcon onClick={handleCancel} />
        
        <Modal.Heading>创建计划组织</Modal.Heading>
        
        <Modal.Body>
          <Flex flexDirection="column" gap="m">
            {parentOrganization && (
              <FormField label="上级组织" hintText={`将在 ${parentOrganization.name} 下创建计划组织`}>
                <TextInput value={parentOrganization.name} disabled />
              </FormField>
            )}

            <FormField
              label="组织名称"
              required
              error={errors.name ? FormField.ErrorType.Error : undefined}
              hintText={errors.name}
            >
              <TextInput
                value={formData.name}
                onChange={(e) => handleFieldChange('name', e.target.value)}
                placeholder="请输入组织名称"
              />
            </FormField>

            <UnitTypeSelector
              value={formData.unit_type}
              onChange={(value) => handleFieldChange('unit_type', value)}
              error={errors.unit_type}
              required
            />

            <FormField
              label="组织描述"
              hintText="可选，描述组织的职能和目的"
            >
              <TextArea
                value={formData.description}
                onChange={(e) => handleFieldChange('description', e.target.value)}
                placeholder="请输入组织描述"
                rows={3}
              />
            </FormField>

            <TemporalDatePicker
              label="生效日期"
              value={formData.effective_date}
              onChange={(value) => handleFieldChange('effective_date', value)}
              error={errors.effective_date}
              required
              minDate={minDate}
              helperText="计划组织必须设置未来的生效日期"
            />

            <TemporalDatePicker
              label="结束日期"
              value={formData.end_date}
              onChange={(value) => handleFieldChange('end_date', value)}
              error={errors.end_date}
              minDate={formData.effective_date || minDate}
              helperText="可选，设置组织的计划结束时间"
            />

            <FormField
              label="变更原因"
              required
              error={errors.change_reason ? FormField.ErrorType.Error : undefined}
              hintText={errors.change_reason || '请说明创建此计划组织的原因'}
            >
              <TextArea
                value={formData.change_reason}
                onChange={(e) => handleFieldChange('change_reason', e.target.value)}
                placeholder="例如：业务扩展需要、组织架构调整、新项目启动等"
                rows={3}
              />
            </FormField>

            {errors.general && (
              <div style={{ color: '#D73502', fontSize: '14px', marginTop: '8px' }}>
                {errors.general}
              </div>
            )}
          </Flex>
        </Modal.Body>
        
        <Modal.Footer>
          <Flex gap="s" justifyContent="flex-end">
            <SecondaryButton onClick={handleCancel} disabled={loading}>
              取消
            </SecondaryButton>
            <PrimaryButton onClick={handleSubmit} disabled={loading}>
              {loading ? '创建中...' : '创建计划组织'}
            </PrimaryButton>
          </Flex>
        </Modal.Footer>
      </Modal.Card>
    </Modal>
  );
};