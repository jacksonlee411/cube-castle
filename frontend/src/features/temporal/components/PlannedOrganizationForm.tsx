import React, { useState, useCallback } from 'react';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { Flex } from '@workday/canvas-kit-react/layout';
import { TemporalDatePicker, validateTemporalDate } from './TemporalDatePicker';

// 定义组织类型 - 从FormFields中复制过来
export type UnitType = 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM';

export interface PlannedOrganizationData {
  name: string;
  unitType: UnitType;
  description?: string;
  effectiveDate: string;
  endDate?: string;
  changeReason: string;
  parentCode?: string;
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
  unitType?: string;
  effectiveDate?: string;
  endDate?: string;
  changeReason?: string;
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
    unitType: 'DEPARTMENT',
    description: '',
    effectiveDate: '',
    endDate: '',
    changeReason: '',
    parentCode: parentOrganization?.code,
  });

  const [errors, setErrors] = useState<FormErrors>({});

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

  // 重置表单
  const resetForm = useCallback(() => {
    setFormData({
      name: '',
      unitType: 'DEPARTMENT',
      description: '',
      effectiveDate: '',
      endDate: '',
      changeReason: '',
      parentCode: parentOrganization?.code,
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
    if (!formData.unitType) {
      newErrors.unitType = '请选择组织类型';
    }

    // 生效日期验证
    if (!formData.effectiveDate) {
      newErrors.effectiveDate = '生效日期不能为空';
    } else if (!validateTemporalDate.isValidDate(formData.effectiveDate)) {
      newErrors.effectiveDate = '请输入有效的日期格式';
    } else if (!validateTemporalDate.isFutureDate(formData.effectiveDate)) {
      newErrors.effectiveDate = '计划组织的生效日期必须在当前日期之后';
    }

    // 结束日期验证
    if (formData.endDate) {
      if (!validateTemporalDate.isValidDate(formData.endDate)) {
        newErrors.endDate = '请输入有效的日期格式';
      } else if (formData.effectiveDate && !validateTemporalDate.isEndDateAfterStartDate(formData.effectiveDate, formData.endDate)) {
        newErrors.endDate = '结束日期必须在生效日期之后';
      }
    }

    // 变更原因验证
    if (!formData.changeReason.trim()) {
      newErrors.changeReason = '变更原因不能为空';
    } else if (formData.changeReason.length < 5) {
      newErrors.changeReason = '变更原因至少需要5个字符';
    } else if (formData.changeReason.length > 500) {
      newErrors.changeReason = '变更原因不能超过500个字符';
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
    model.events.hide();
    onClose();
  };

  const minDate = validateTemporalDate.getTodayString();

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card width="600px" maxWidth="90vw">
          <Modal.CloseIcon onClick={handleCancel} />
          
          <Modal.Heading>创建计划组织</Modal.Heading>
        
        <Modal.Body>
          <Flex flexDirection="column" gap="m">
            {parentOrganization && (
            <FormField>
              <FormField.Label>上级组织</FormField.Label>
              <FormField.Field>
                <FormField.Input as={TextInput} value={parentOrganization.name} disabled />
                <FormField.Hint>将在 {parentOrganization.name} 下创建计划组织</FormField.Hint>
              </FormField.Field>
            </FormField>
            )}

            <FormField
              isRequired
              error={errors.name ? "error" : undefined}
            >
              <FormField.Label>组织名称</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextInput}
                  value={formData.name}
                  onChange={(e) => handleFieldChange('name', e.target.value)}
                  placeholder="请输入组织名称"
                />
                {errors.name && (
                  <FormField.Hint>{errors.name}</FormField.Hint>
                )}
              </FormField.Field>
            </FormField>

            <FormField>
              <FormField.Label>组织类型 *</FormField.Label>
              <FormField.Field>
                <select
                  value={formData.unitType}
                  onChange={(e) => handleFieldChange('unitType', e.target.value as UnitType)}
                  style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
                >
                  <option value="DEPARTMENT">部门</option>
                  <option value="ORGANIZATION_UNIT">组织单位</option>
                  <option value="PROJECT_TEAM">项目团队</option>
                </select>
              </FormField.Field>
            </FormField>

            <FormField>
              <FormField.Label>组织描述</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextArea}
                  value={formData.description}
                  onChange={(e) => handleFieldChange('description', e.target.value)}
                  placeholder="请输入组织描述"
                  rows={3}
                />
                <FormField.Hint>可选，描述组织的职能和目的</FormField.Hint>
              </FormField.Field>
            </FormField>

            <TemporalDatePicker
              label="生效日期"
              value={formData.effectiveDate}
              onChange={(value) => handleFieldChange('effectiveDate', value)}
              error={errors.effectiveDate}
              required
              minDate={minDate}
              helperText="计划组织必须设置未来的生效日期"
            />

            <TemporalDatePicker
              label="结束日期"
              value={formData.endDate}
              onChange={(value) => handleFieldChange('endDate', value)}
              error={errors.endDate}
              minDate={formData.effectiveDate || minDate}
              helperText="可选，设置组织的计划结束时间"
            />

            <FormField
              isRequired
              error={errors.changeReason ? "error" : undefined}
            >
              <FormField.Label>变更原因</FormField.Label>
              <FormField.Field>
                <FormField.Input
                  as={TextArea}
                  value={formData.changeReason}
                  onChange={(e) => handleFieldChange('changeReason', e.target.value)}
                  placeholder="例如：业务扩展需要、组织架构调整、新项目启动等"
                  rows={3}
                />
                <FormField.Hint>
                  {errors.changeReason || '请说明创建此计划组织的原因'}
                </FormField.Hint>
              </FormField.Field>
            </FormField>

            {errors.general && (
              <div style={{ color: '#D73502', fontSize: '14px', marginTop: '8px' }}>
                {errors.general}
              </div>
            )}

            {/* 按钮区域 */}
            <Flex gap="s" justifyContent="flex-end" marginTop="l">
              <SecondaryButton onClick={handleCancel} disabled={loading}>
                取消
              </SecondaryButton>
              <PrimaryButton onClick={handleSubmit} disabled={loading}>
                {loading ? '创建中...' : '创建计划组织'}
              </PrimaryButton>
            </Flex>
          </Flex>
        </Modal.Body>
      </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};
