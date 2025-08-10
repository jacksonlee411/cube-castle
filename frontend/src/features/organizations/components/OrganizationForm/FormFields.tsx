import React from 'react';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import type { FormFieldsProps } from './FormTypes';

export const FormFields: React.FC<FormFieldsProps> = ({
  formData,
  setFormData,
  isEditing
}) => {
  const updateField = (field: string, value: any) => {
    setFormData({ ...formData, [field]: value });
  };

  return (
    <>
      <FormField marginBottom="m">
        <FormField.Label>组织编码</FormField.Label>
        <FormField.Field>
          <FormField.Input
            as={TextInput}
            value={formData.code}
            onChange={(e) => updateField('code', e.target.value)}
            disabled={true}
            placeholder="系统自动生成编码"
            style={{ backgroundColor: '#f5f5f5', cursor: 'not-allowed' }}
            data-testid="form-field-code"
          />
        </FormField.Field>
        <FormField.Hint>
          {isEditing ? "编码不可修改" : "系统将自动生成唯一编码"}
        </FormField.Hint>
      </FormField>

      <FormField marginBottom="m">
        <FormField.Label>组织名称 *</FormField.Label>
        <FormField.Field>
          <FormField.Input
            as={TextInput}
            value={formData.name}
            onChange={(e) => updateField('name', e.target.value)}
            placeholder="请输入组织名称"
            required
            data-testid="form-field-name"
          />
        </FormField.Field>
      </FormField>

      <FormField marginBottom="m">
        <FormField.Label>组织类型 *</FormField.Label>
        <FormField.Field>
          <select
            value={formData.unit_type}
            onChange={(e) => updateField('unit_type', e.target.value)}
            disabled={false}
            style={{ 
              width: '100%', 
              padding: '8px', 
              borderRadius: '4px', 
              border: '1px solid #ddd',
              backgroundColor: 'white',
              cursor: 'pointer'
            }}
            data-testid="form-field-unit-type"
          >
            <option value="DEPARTMENT">部门</option>
            <option value="COST_CENTER">成本中心</option>
            <option value="COMPANY">公司</option>
            <option value="PROJECT_TEAM">项目团队</option>
          </select>
        </FormField.Field>
      </FormField>

      <FormField marginBottom="m">
        <FormField.Label>上级组织编码</FormField.Label>
        <FormField.Field>
          <FormField.Input
            as={TextInput}
            value={formData.parent_code}
            onChange={(e) => updateField('parent_code', e.target.value)}
            disabled={false}
            placeholder="请输入上级组织编码"
            data-testid="form-field-parent-code"
          />
        </FormField.Field>
      </FormField>

      <FormField marginBottom="m">
        <FormField.Label>组织层级</FormField.Label>
        <FormField.Field>
          <FormField.Input
            as={TextInput}
            type="number"
            value={formData.level}
            disabled={true}
            style={{ backgroundColor: '#f5f5f5', cursor: 'not-allowed' }}
            data-testid="form-field-level"
          />
        </FormField.Field>
        <FormField.Hint>
          层级由上级组织关系自动计算，不可手动修改
        </FormField.Hint>
      </FormField>

      <FormField marginBottom="m">
        <FormField.Label>状态 *</FormField.Label>
        <FormField.Field>
          <select
            value={formData.status}
            onChange={(e) => updateField('status', e.target.value)}
            style={{ 
              width: '100%', 
              padding: '8px', 
              borderRadius: '4px', 
              border: '1px solid #ddd' 
            }}
            data-testid="form-field-status"
          >
            <option value="ACTIVE">激活</option>
            <option value="INACTIVE">停用</option>
            <option value="PLANNED">计划中</option>
          </select>
        </FormField.Field>
      </FormField>

      <FormField marginBottom="m">
        <FormField.Label>排序</FormField.Label>
        <FormField.Field>
          <FormField.Input
            as={TextInput}
            type="number"
            value={formData.sort_order}
            onChange={(e) => updateField('sort_order', parseInt(e.target.value) || 0)}
            min="0"
            data-testid="form-field-sort-order"
          />
        </FormField.Field>
      </FormField>

      <FormField marginBottom="l">
        <FormField.Label>描述</FormField.Label>
        <FormField.Field>
          <FormField.Input
            as={TextArea}
            value={formData.description}
            onChange={(e) => updateField('description', e.target.value)}
            placeholder="请输入组织描述"
            rows={3}
            data-testid="form-field-description"
          />
        </FormField.Field>
      </FormField>
    </>
  );
};