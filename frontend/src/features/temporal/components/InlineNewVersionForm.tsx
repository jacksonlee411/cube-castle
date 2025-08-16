/**
 * 内联新增版本表单组件
 * 集成到右侧详情区域，替代Modal弹窗，提升用户体验
 */
import React, { useState, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { Select } from '@workday/canvas-kit-react/select';
import { colors, borderRadius } from '@workday/canvas-kit-react/tokens';
import { type TemporalEditFormData } from './TemporalEditForm';

export interface InlineNewVersionFormProps {
  organizationCode: string;
  onSubmit: (data: TemporalEditFormData) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
}

const unitTypeOptions = [
  { label: '公司', value: 'COMPANY' },
  { label: '部门', value: 'DEPARTMENT' },
  { label: '成本中心', value: 'COST_CENTER' },
  { label: '项目团队', value: 'PROJECT_TEAM' },
];

const statusOptions = [
  { label: '启用', value: 'ACTIVE' },
  { label: '计划中', value: 'PLANNED' },
  { label: '停用', value: 'INACTIVE' },
];

const eventTypeOptions = [
  { label: '更新', value: 'UPDATE' },
  { label: '重组', value: 'RESTRUCTURE' },
  { label: '撤销', value: 'DISSOLVE' },
];

export const InlineNewVersionForm: React.FC<InlineNewVersionFormProps> = ({
  organizationCode,
  onSubmit,
  onCancel,
  isSubmitting = false
}) => {
  const [formData, setFormData] = useState<TemporalEditFormData>({
    name: '',
    unit_type: 'DEPARTMENT',
    status: 'PLANNED',
    description: '',
    effective_date: new Date().toISOString().split('T')[0], // 默认今天
    change_reason: '',
    event_type: 'RESTRUCTURE'
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // 初始化表单数据 - 默认明天生效
  useEffect(() => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    setFormData({
      name: '',
      unit_type: 'DEPARTMENT',
      status: 'PLANNED',
      description: '',
      effective_date: tomorrow.toISOString().split('T')[0], // 默认明天生效
      change_reason: '',
      event_type: 'RESTRUCTURE'
    });
    setErrors({});
  }, []);

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

  return (
    <Box flex="1" padding="m">
      <Card padding="l">
        {/* 表单标题 */}
        <Flex justifyContent="space-between" alignItems="center" marginBottom="l">
          <Box>
            <Heading size="medium" marginBottom="s">
              新增时态版本
            </Heading>
            <Text typeLevel="subtext.medium" color="hint">
              为组织 {organizationCode} 创建新的时态版本
            </Text>
          </Box>
          
          {/* 表单状态指示器 */}
          <Box
            padding="s"
            backgroundColor={colors.soap200}
            borderRadius={borderRadius.m}
          >
            <Text typeLevel="subtext.small" color={colors.licorice700}>
              详情 新增模式
            </Text>
          </Box>
        </Flex>

        {/* 注意事项提示 */}
        <Box
          marginBottom="l"
          padding="m"
          backgroundColor={colors.blueberry50}
          borderRadius={borderRadius.m}
          border={`1px solid ${colors.blueberry200}`}
        >
          <Text typeLevel="subtext.medium" color={colors.blueberry700}>
            提示 <strong>提示:</strong> 新增版本将在指定生效日期自动生效，请确保填写准确的组织信息和变更原因。
            左侧时间轴将保持可见，便于参考历史版本信息。
          </Text>
        </Box>

        <form onSubmit={handleSubmit}>
          {/* 基本信息 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s" color={colors.blueberry600}>
              详情 基本信息
            </Heading>
            
            <Box marginLeft="m">
              <FormField error={errors.name ? "error" : undefined}>
                <FormField.Label>组织名称 *</FormField.Label>
                <FormField.Field>
                  <TextInput
                    value={formData.name}
                    onChange={handleInputChange('name')}
                    placeholder="请输入组织名称"
                    disabled={isSubmitting}
                  />
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>组织类型 *</FormField.Label>
                <FormField.Field>
                  <Select
                    value={formData.unit_type}
                    onChange={handleInputChange('unit_type')}
                  >
                    {unitTypeOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </Select>
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>组织状态 *</FormField.Label>
                <FormField.Field>
                  <Select
                    value={formData.status}
                    onChange={handleInputChange('status')}
                  >
                    {statusOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </Select>
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>描述信息</FormField.Label>
                <FormField.Field>
                  <TextArea
                    value={formData.description}
                    onChange={handleInputChange('description')}
                    placeholder="请输入组织描述信息"
                    disabled={isSubmitting}
                    rows={3}
                  />
                </FormField.Field>
              </FormField>
            </Box>
          </Box>

          {/* 时态信息 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s" color={colors.greenFresca600}>
              ⏰ 时态信息
            </Heading>

            <Box marginLeft="m">
              <FormField error={errors.effective_date ? "error" : undefined}>
                <FormField.Label>生效日期 *</FormField.Label>
                <FormField.Field>
                  <TextInput
                    type="date"
                    value={formData.effective_date}
                    onChange={handleInputChange('effective_date')}
                    disabled={isSubmitting}
                  />
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>事件类型 *</FormField.Label>
                <FormField.Field>
                  <Select
                    value={formData.event_type}
                    onChange={handleInputChange('event_type')}
                  >
                    {eventTypeOptions.map(option => (
                      <option key={option.value} value={option.value}>
                        {option.label}
                      </option>
                    ))}
                  </Select>
                </FormField.Field>
              </FormField>

              <FormField error={errors.change_reason ? "error" : undefined}>
                <FormField.Label>变更原因 *</FormField.Label>
                <FormField.Field>
                  <TextArea
                    value={formData.change_reason}
                    onChange={handleInputChange('change_reason')}
                    placeholder="请详细说明此次变更的原因和目的"
                    disabled={isSubmitting}
                    rows={3}
                  />
                </FormField.Field>
              </FormField>
            </Box>
          </Box>

          {/* 操作按钮 */}
          <Box
            marginTop="xl"
            paddingTop="l"
            borderTop={`1px solid ${colors.soap300}`}
          >
            <Flex gap="s" justifyContent="flex-end">
              <SecondaryButton 
                onClick={onCancel}
                disabled={isSubmitting}
              >
                取消
              </SecondaryButton>
              <PrimaryButton 
                type="submit"
                disabled={isSubmitting}
              >
                {isSubmitting ? '创建中...' : '创建新版本'}
              </PrimaryButton>
            </Flex>
          </Box>
        </form>
      </Card>
    </Box>
  );
};

export default InlineNewVersionForm;