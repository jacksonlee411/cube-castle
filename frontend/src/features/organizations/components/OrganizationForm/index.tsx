import React, { useState, useEffect, useCallback } from 'react';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { useCreateOrganization, useUpdateOrganization } from '../../../../shared/hooks/useOrganizationMutations';
import { FormFields } from './FormFields';
import { validateForm } from './ValidationRules';
import type { OrganizationFormProps, FormData } from './FormTypes';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../../shared/hooks/useOrganizationMutations';

export const OrganizationForm: React.FC<OrganizationFormProps> = ({
  organization,
  isOpen,
  onClose
}) => {
  const createMutation = useCreateOrganization();
  const updateMutation = useUpdateOrganization();
  const isEditing = !!organization;
  
  const [isSubmitting, setIsSubmitting] = useState(false);
  const model = useModalModel();

  const [formData, setFormData] = useState<FormData>({
    code: organization?.code || '',
    name: organization?.name || '',
    unit_type: organization?.unit_type || 'DEPARTMENT',
    status: organization?.status || 'ACTIVE',
    description: organization?.description || '',
    parent_code: organization?.parent_code || '',
    level: organization?.level || 1,
    sort_order: organization?.sort_order || 0,
  });

  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  // Modal state management
  useEffect(() => {
    if (isOpen && model.state.visibility !== 'visible') {
      model.events.show();
    } else if (!isOpen && model.state.visibility === 'visible') {
      model.events.hide();
    }
  }, [isOpen, model]);

  // Reset form data when organization changes
  useEffect(() => {
    setFormData({
      code: organization?.code || '',
      name: organization?.name || '',
      unit_type: organization?.unit_type || 'DEPARTMENT',
      status: organization?.status || 'ACTIVE',
      description: organization?.description || '',
      parent_code: organization?.parent_code || '',
      level: organization?.level || 1,
      sort_order: organization?.sort_order || 0,
    });
    setFormErrors({});
  }, [organization]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    // Prevent double submission
    if (isSubmitting || createMutation.isPending || updateMutation.isPending) {
      return;
    }
    
    // Validate form
    const errors = validateForm(formData, isEditing);
    if (Object.keys(errors).length > 0) {
      setFormErrors(errors);
      return;
    }
    
    setIsSubmitting(true);
    
    try {
      if (isEditing) {
        const updateData: UpdateOrganizationInput = {
          code: organization!.code,
          name: formData.name,
          unit_type: formData.unit_type as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          description: formData.description,
          sort_order: formData.sort_order,
          level: formData.level,
          parent_code: formData.parent_code || undefined,
        };
        
        await updateMutation.mutateAsync(updateData);
      } else {
        const createData: CreateOrganizationInput = {
          code: formData.code && formData.code.trim() ? formData.code.trim() : undefined,
          name: formData.name,
          unit_type: formData.unit_type as 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          level: formData.level,
          sort_order: formData.sort_order,
          description: formData.description,
          parent_code: formData.parent_code || undefined,
        };
        
        await createMutation.mutateAsync(createData);
      }
      
      // Reset form if creating new
      if (!isEditing) {
        setFormData({
          code: '',
          name: '',
          unit_type: 'DEPARTMENT',
          status: 'ACTIVE',
          description: '',
          parent_code: '',
          level: 1,
          sort_order: 0,
        });
      }
      
      model.events.hide();
      onClose();
    } catch (error) {
      console.error(`[Form] ${isEditing ? '更新' : '创建'}失败:`, error);
      
      let errorMessage = '操作失败';
      
      if (error && typeof error === 'object' && 'message' in error) {
        const apiError = error as any;
        
        if (apiError.message.includes('duplicate key value violates unique constraint')) {
          if (apiError.message.includes('uk_tenant_name')) {
            errorMessage = '组织名称已存在，请使用不同的名称';
          } else {
            errorMessage = '数据重复，请检查输入信息';
          }
        } else if (apiError.message.includes('Network connection failed')) {
          errorMessage = '网络连接失败，请检查服务器状态';
        } else {
          errorMessage = apiError.message.split(' - ')[0] || errorMessage;
        }
      } else if (error instanceof Error) {
        errorMessage = error.message;
      }
      
      alert(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  }, [isEditing, formData, createMutation, updateMutation, isSubmitting, model, onClose]);

  const handleClose = () => {
    setIsSubmitting(false);
    setFormErrors({});
    model.events.hide();
    onClose();
  };

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card width={600} data-testid="organization-form">
          <Modal.CloseIcon aria-label="关闭" onClick={handleClose} />
          <Modal.Heading>
            {isEditing ? '编辑组织单元' : '新增组织单元'}
          </Modal.Heading>
          <Modal.Body>
            <form onSubmit={handleSubmit} data-testid="organization-form-content">
              <FormFields
                formData={formData}
                setFormData={setFormData}
                isEditing={isEditing}
              />

              {Object.keys(formErrors).length > 0 && (
                <div style={{ 
                  marginBottom: '16px', 
                  padding: '8px', 
                  backgroundColor: '#fff2f0', 
                  border: '1px solid #ffccc7',
                  borderRadius: '4px' 
                }}>
                  {Object.entries(formErrors).map(([field, error]) => (
                    <div key={field} style={{ color: '#ff4d4f', fontSize: '14px' }}>
                      {error}
                    </div>
                  ))}
                </div>
              )}

              <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '8px' }}>
                <SecondaryButton 
                  type="button" 
                  onClick={handleClose}
                  data-testid="form-cancel-button"
                >
                  取消
                </SecondaryButton>
                <PrimaryButton 
                  type="submit" 
                  disabled={isSubmitting || createMutation.isPending || updateMutation.isPending}
                  data-testid="form-submit-button"
                >
                  {isEditing ? '更新' : '创建'}
                </PrimaryButton>
              </div>
            </form>
          </Modal.Body>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};