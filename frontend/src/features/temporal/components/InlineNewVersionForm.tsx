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
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { colors, borderRadius } from '@workday/canvas-kit-react/tokens';
import { type TemporalEditFormData } from './TemporalEditForm';
import { FiveStateStatusSelector, LIFECYCLE_STATES, type LifecycleStatus } from './FiveStateStatusSelector';

export interface InlineNewVersionFormProps {
  organizationCode: string;
  onSubmit: (data: TemporalEditFormData) => Promise<void>;
  onCancel: () => void;
  isSubmitting?: boolean;
  mode?: 'create' | 'edit' | 'edit-history'; // 添加历史记录编辑模式
  initialData?: {
    name: string;
    unit_type: string;
    status: string;
    lifecycle_status?: string;
    description?: string;
    parent_code?: string;
    effective_date?: string;
  };
  // 新增历史记录编辑相关props
  selectedVersion?: {
    record_id: string; // UUID唯一标识符
    created_at: string;
    updated_at: string;
    code: string;
    name: string;
    unit_type: string;
    status: string;
    effective_date: string;
    description?: string;
    parent_code?: string;
  } | null;
  onEditHistory?: (versionData: any) => Promise<void>;
  onDeactivate?: (version: any) => Promise<void>; // 新增作废功能
}

const unitTypeOptions = [
  { label: '公司', value: 'COMPANY' },
  { label: '部门', value: 'DEPARTMENT' },
  { label: '成本中心', value: 'COST_CENTER' },
  { label: '项目团队', value: 'PROJECT_TEAM' },
];


export const InlineNewVersionForm: React.FC<InlineNewVersionFormProps> = ({
  organizationCode,
  onSubmit,
  onCancel,
  isSubmitting = false,
  mode = 'create',
  initialData,
  selectedVersion,
  onEditHistory,
  onDeactivate
}) => {
  const [formData, setFormData] = useState<TemporalEditFormData>({
    name: '',
    unit_type: 'DEPARTMENT',
    lifecycle_status: 'PLANNED',
    description: '',
    effective_date: new Date().toISOString().split('T')[0], // 默认今天
    parent_code: ''
  });

  const [errors, setErrors] = useState<Record<string, string>>({});
  
  // 历史记录编辑相关状态
  const [isEditingHistory, setIsEditingHistory] = useState(false);
  const [originalHistoryData, setOriginalHistoryData] = useState<any>(null);
  
  // 作废功能相关状态
  const [showDeactivateConfirm, setShowDeactivateConfirm] = useState(false);
  const [isDeactivating, setIsDeactivating] = useState(false);
  
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

  // 初始化表单数据 - 支持预填充模式和历史记录编辑
  useEffect(() => {
    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);
    
    if ((mode === 'edit' || mode === 'edit-history') && initialData) {
      // 编辑模式 - 使用传入的初始数据预填充表单，包括原始生效日期
      setFormData({
        name: initialData.name,
        unit_type: initialData.unit_type,
        lifecycle_status: initialData.lifecycle_status || 'PLANNED',
        description: initialData.description || '',
        effective_date: initialData.effective_date 
          ? new Date(initialData.effective_date).toISOString().split('T')[0] 
          : tomorrow.toISOString().split('T')[0], // 如果没有提供生效日期，使用明天
        parent_code: initialData.parent_code || ''
      });
      
      // 如果是历史记录编辑模式，保存原始数据
      if (mode === 'edit-history' && selectedVersion) {
        setOriginalHistoryData(selectedVersion);
        setIsEditingHistory(false); // 初始时为只读模式
      }
    } else {
      // 新增模式 - 使用默认值
      setFormData({
        name: '',
        unit_type: 'DEPARTMENT',
        lifecycle_status: 'PLANNED',
        description: '',
        effective_date: tomorrow.toISOString().split('T')[0], // 默认明天生效
        parent_code: ''
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
    
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  // 历史记录编辑相关处理函数
  const handleEditHistoryToggle = () => {
    setIsEditingHistory(!isEditingHistory);
  };

  const handleEditHistorySubmit = async () => {
    if (!validateForm() || !onEditHistory || !originalHistoryData) {
      return;
    }

    try {
      // 构建更新数据，包含ID和更新时间戳
      const updateData = {
        ...originalHistoryData,
        name: formData.name,
        unit_type: formData.unit_type,
        status: formData.lifecycle_status,
        description: formData.description,
        effective_date: formData.effective_date,
        parent_code: formData.parent_code,
        updated_at: new Date().toISOString()
      };
      
      await onEditHistory(updateData);
      setIsEditingHistory(false); // 提交后回到只读模式
    } catch (error) {
      console.error('修改历史记录失败:', error);
    }
  };

  const handleCancelEditHistory = () => {
    // 恢复原始数据
    if (originalHistoryData) {
      setFormData({
        name: originalHistoryData.name,
        unit_type: originalHistoryData.unit_type,
        lifecycle_status: originalHistoryData.status || 'PLANNED',
        description: originalHistoryData.description || '',
        effective_date: new Date(originalHistoryData.effective_date).toISOString().split('T')[0],
        parent_code: originalHistoryData.parent_code || ''
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
      // 作废成功后关闭页面
      onCancel();
    } catch (error) {
      console.error('作废失败:', error);
      alert('作废失败，请重试');
    } finally {
      setIsDeactivating(false);
    }
  };

  const handleDeactivateCancel = () => {
    setShowDeactivateConfirm(false);
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
              {mode === 'edit-history' 
                ? (isEditingHistory ? '编辑历史记录' : '查看历史记录')
                : mode === 'edit' ? '编辑组织信息' : '组织信息详情'}
            </Heading>
            <Text typeLevel="subtext.medium" color="hint">
              {mode === 'edit-history' 
                ? `${isEditingHistory ? '修改' : '查看'}组织 ${organizationCode} 的历史记录信息`
                : mode === 'edit' 
                  ? `基于现有版本编辑组织 ${organizationCode} 的信息` 
                  : `为组织 ${organizationCode} 编辑组织信息`}
            </Text>
          </Box>
          
        </Flex>

        {/* 历史记录元数据显示 - 移到最下方 */}

        <form onSubmit={handleSubmit}>
          {/* 生效日期 - 最重要的信息放在最上方 */}
          <Box marginBottom="l">
            <Heading size="small" marginBottom="s" color={colors.blueberry600}>
              生效日期
            </Heading>
            
            <Box marginLeft="m">
              <FormField error={errors.effective_date ? "error" : undefined}>
                <FormField.Label>生效日期 *</FormField.Label>
                <FormField.Field>
                  <TextInput
                    type="date"
                    value={formData.effective_date}
                    onChange={handleInputChange('effective_date')}
                    disabled={isSubmitting || (mode === 'edit-history' && !isEditingHistory)}
                  />
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
              <FormField error={errors.name ? "error" : undefined}>
                <FormField.Label>组织名称 *</FormField.Label>
                <FormField.Field>
                  <TextInput
                    value={formData.name}
                    onChange={handleInputChange('name')}
                    placeholder="请输入组织名称"
                    disabled={isSubmitting || (mode === 'edit-history' && !isEditingHistory)}
                  />
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>上级组织编码</FormField.Label>
                <FormField.Field>
                  <TextInput
                    value={formData.parent_code || ''}
                    onChange={handleInputChange('parent_code')}
                    placeholder="请输入上级组织编码（可选）"
                    disabled={isSubmitting || (mode === 'edit-history' && !isEditingHistory)}
                  />
                </FormField.Field>
              </FormField>

              <FormField>
                <FormField.Label>组织类型 *</FormField.Label>
                <FormField.Field>
                  <select
                    value={formData.unit_type}
                    onChange={handleInputChange('unit_type')}
                    disabled={isSubmitting || (mode === 'edit-history' && !isEditingHistory)}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      border: `1px solid ${colors.soap400}`,
                      borderRadius: borderRadius.m,
                      fontSize: '14px',
                      backgroundColor: (isSubmitting || (mode === 'edit-history' && !isEditingHistory)) ? '#f5f5f5' : 'white'
                    }}
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
                disabled={isSubmitting || (mode === 'edit-history' && !isEditingHistory)}
                includeDeleted={false}
                label="组织状态"
                required={true}
                error={errors.lifecycle_status}
              />

              <FormField>
                <FormField.Label>描述信息</FormField.Label>
                <FormField.Field>
                  <TextArea
                    value={formData.description}
                    onChange={handleInputChange('description')}
                    placeholder="请输入组织描述信息"
                    disabled={isSubmitting || (mode === 'edit-history' && !isEditingHistory)}
                    rows={3}
                  />
                </FormField.Field>
              </FormField>
            </Box>
          </Box>

          {/* 历史记录元数据显示 - 移到表单底部 */}
          {mode === 'edit-history' && originalHistoryData && (
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
                    {originalHistoryData.record_id}
                  </Text>
                </Box>
                <Box>
                  <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
                    创建时间:
                  </Text>
                  <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700}>
                    {new Date(originalHistoryData.created_at).toLocaleString('zh-CN')}
                  </Text>
                </Box>
                <Box>
                  <Text typeLevel="subtext.small" fontWeight="bold" color={colors.licorice500}>
                    最后更新:
                  </Text>
                  <Text typeLevel="subtext.small" marginTop="xs" color={colors.licorice700}>
                    {new Date(originalHistoryData.updated_at).toLocaleString('zh-CN')}
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
            {mode === 'edit-history' ? (
              // 历史记录编辑模式的按钮
              <Flex gap="s" justifyContent="space-between">
                {/* 左侧作废按钮 */}
                <Box>
                  {selectedVersion && !isEditingHistory && (
                    <TertiaryButton 
                      onClick={handleDeactivateClick}
                      disabled={isSubmitting || isDeactivating}
                    >
                      作废此版本
                    </TertiaryButton>
                  )}
                </Box>
                
                {/* 右侧主要操作按钮 */}
                <Flex gap="s">
                  {!isEditingHistory ? (
                    // 只读模式的按钮
                    <>
                      <SecondaryButton 
                        onClick={onCancel}
                        disabled={isSubmitting}
                      >
                        关闭
                      </SecondaryButton>
                      <PrimaryButton 
                        onClick={handleEditHistoryToggle}
                        disabled={isSubmitting}
                      >
                        修改历史记录
                      </PrimaryButton>
                    </>
                  ) : (
                    // 编辑模式的按钮
                    <>
                      <SecondaryButton 
                        onClick={handleCancelEditHistory}
                        disabled={isSubmitting}
                      >
                        取消编辑
                      </SecondaryButton>
                      <PrimaryButton 
                        onClick={handleEditHistorySubmit}
                        disabled={isSubmitting}
                      >
                        {isSubmitting ? '提交中...' : '提交修改'}
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
                  disabled={isSubmitting}
                >
                  取消
                </SecondaryButton>
                <PrimaryButton 
                  type="submit"
                  disabled={isSubmitting}
                >
                  {isSubmitting 
                    ? (mode === 'edit' ? '创建中...' : '创建中...') 
                    : (mode === 'edit' ? '基于此版本创建' : '创建新版本')}
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
              <Modal.Heading>确认作废版本</Modal.Heading>
              <Modal.Body>
                <Box padding="l">
                  <Flex alignItems="flex-start" gap="m" marginBottom="l">
                    <Box fontSize="24px" color={colors.cinnamon600}>⚠️</Box>
                    <Box>
                      <Text typeLevel="body.medium" marginBottom="s">
                        确定要作废生效日期为 <strong>{new Date(selectedVersion.effective_date).toLocaleDateString('zh-CN')}</strong> 的版本吗？
                      </Text>
                      <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                        版本名称: {selectedVersion.name}
                      </Text>
                      <Text typeLevel="subtext.small" color={colors.cinnamon600}>
                        ⚠️ 作废后将自动填补时间空洞，此操作不可撤销
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
                      {isDeactivating ? '作废中...' : '确认作废'}
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