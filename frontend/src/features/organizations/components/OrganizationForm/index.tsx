import React, { useState, useEffect, useCallback } from 'react';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { useCreateOrganization, useUpdateOrganization } from '../../../../shared/hooks/useOrganizationMutations';
// import { useTemporalMode } from '../../../../shared/hooks/useTemporalQuery';
import organizationAPI from '../../../../shared/api/organizations';
import { FormFields } from './FormFields';
import { validateForm } from './ValidationRules';
import type { OrganizationFormProps, FormData } from './FormTypes';
import type { CreateOrganizationInput, UpdateOrganizationInput } from '../../../../shared/hooks/useOrganizationMutations';
import { TemporalConverter } from '../../../../shared/utils/temporal-converter';
import { useMessages } from '../../../../shared/hooks/useMessages';
import { normalizeParentCode } from '../../../../shared/utils/organization-helpers';

export const OrganizationForm: React.FC<OrganizationFormProps> = ({
  organization,
  isOpen,
  onClose,
  temporalMode = 'current',
  isHistorical = false,
  enableTemporalFeatures = true
}) => {
  const createMutation = useCreateOrganization();
  const updateMutation = useUpdateOrganization();
  // const { isCurrent, isPlanning } = useTemporalMode();
  const { showError } = useMessages();
  
  const isEditing = !!organization;
  const [isSubmitting, setIsSubmitting] = useState(false);
  const model = useModalModel();

  const [formData, setFormData] = useState<FormData>({
    code: organization?.code || '',
    name: organization?.name || '',
    unitType: organization?.unitType || 'DEPARTMENT',
    status: organization?.status || 'ACTIVE',
    description: organization?.description || '',
    parentCode: normalizeParentCode.forForm(organization?.parentCode),
    level: organization?.level || 1,
    sortOrder: organization?.sortOrder || 0,
    // 时态字段
    isTemporal: false,
    effectiveFrom: '',
    effectiveTo: '',
    changeReason: '',
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
      unitType: organization?.unitType || 'DEPARTMENT',
      status: organization?.status || 'ACTIVE',
      description: organization?.description || '',
      parentCode: normalizeParentCode.forForm(organization?.parentCode),
      level: organization?.level || 1,
      sortOrder: organization?.sortOrder || 0,
      // 时态字段重置
      isTemporal: false,
      effectiveFrom: '',
      effectiveTo: '',
      changeReason: '',
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
          unitType: formData.unitType as 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          description: formData.description,
          sortOrder: formData.sortOrder,
          level: formData.level,
          parentCode: normalizeParentCode.forAPI(formData.parentCode),
        };
        
        // 时态更新 (统一字符串类型处理)
        if (formData.isTemporal) {
          const temporalUpdateData = {
            ...updateData,
            effectiveFrom: formData.effectiveFrom ? TemporalConverter.dateToIso(formData.effectiveFrom) : TemporalConverter.getCurrentISOString(),
            effectiveTo: formData.effectiveTo ? TemporalConverter.dateToIso(formData.effectiveTo) : undefined,
            changeReason: formData.changeReason
          };
          await organizationAPI.updateTemporal(organization!.code, temporalUpdateData);
        } else {
          await updateMutation.mutateAsync(updateData);
        }
      } else {
        const createData: CreateOrganizationInput = {
          code: formData.code && formData.code.trim() ? formData.code.trim() : undefined,
          name: formData.name,
          unitType: formData.unitType as 'DEPARTMENT' | 'ORGANIZATION_UNIT' | 'PROJECT_TEAM',
          status: formData.status as 'ACTIVE' | 'INACTIVE' | 'PLANNED',
          level: formData.level,
          sortOrder: formData.sortOrder,
          description: formData.description,
          parentCode: normalizeParentCode.forAPI(formData.parentCode),
        };
        
        // 时态创建 (统一字符串类型处理)
        if (formData.isTemporal) {
          const temporalCreateData = {
            ...createData,
            effectiveFrom: TemporalConverter.dateToIso(formData.effectiveFrom!),
            effectiveTo: formData.effectiveTo ? TemporalConverter.dateToIso(formData.effectiveTo) : undefined,
            changeReason: formData.changeReason
          };
          await organizationAPI.createTemporal(temporalCreateData);
        } else {
          await createMutation.mutateAsync(createData);
        }
      }
      
      // Reset form if creating new
      if (!isEditing) {
        setFormData({
          code: '',
          name: '',
          unitType: 'DEPARTMENT',
          status: 'ACTIVE',
          description: '',
          parentCode: '',
          level: 1,
          sortOrder: 0,
          // 重置时态字段
          isTemporal: false,
          effectiveFrom: '',
          effectiveTo: '',
          changeReason: '',
        });
      }
      
      model.events.hide();
      onClose();
    } catch (error) {
      console.error(`[Form] ${isEditing ? '更新' : '创建'}失败:`, error);
      
      let errorMessage = '操作失败';
      
      if (error && typeof error === 'object' && 'message' in error) {
        const apiError = error as { message: string };
        
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
      
      showError(errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  }, [isEditing, formData, createMutation, updateMutation, isSubmitting, model, onClose, organization, showError]);

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
            {formData.isTemporal && (
              <span style={{ 
                marginLeft: '8px', 
                fontSize: '14px', 
                color: '#1890ff',
                fontWeight: 'normal'
              }}>
                设置 组织详情
              </span>
            )}
          </Modal.Heading>
          <Modal.Body>
            <form onSubmit={handleSubmit} data-testid="organization-form-content">
              <FormFields
                formData={formData}
                setFormData={setFormData}
                isEditing={isEditing}
                temporalMode={temporalMode}
                enableTemporalFeatures={enableTemporalFeatures && !isHistorical}
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
                  {isSubmitting ? '处理中...' : 
                   formData.isTemporal ? 
                     (isEditing ? '更新时态组织' : '创建计划组织') : 
                     (isEditing ? '更新' : '创建')
                  }
                </PrimaryButton>
              </div>
            </form>
          </Modal.Body>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};