import React, { useState, useCallback } from 'react';
import type { FormFieldsProps } from './FormTypes';
import { TemporalConverter } from '../../../../shared/utils/temporal-converter';

// 日期格式化工具 (使用TemporalConverter统一处理)
const formatDateForInput = (dateStr?: string) => {
  if (!dateStr) return '';
  try {
    return TemporalConverter.dateToDateString(dateStr);
  } catch {
    return '';
  }
};

export const FormFields: React.FC<FormFieldsProps> = ({
  formData,
  setFormData,
  isEditing,
  temporalMode = 'current',
  enableTemporalFeatures = true
}) => {
  const [showAdvancedTemporal, setShowAdvancedTemporal] = useState(false);
  
  const updateField = useCallback((field: string, value: string | number | boolean) => {
    setFormData({ ...formData, [field]: value });
  }, [formData, setFormData]);

  const isTemporal = formData.is_temporal as boolean;
  const isPlannedStatus = formData.status === 'PLANNED';

  // 如果状态是计划中，自动启用时态功能
  React.useEffect(() => {
    if (isPlannedStatus && !isTemporal) {
      updateField('is_temporal', true);
      // 使用TemporalConverter设置默认生效时间为明天上午9点
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      tomorrow.setHours(9, 0, 0, 0);
      updateField('effective_from', TemporalConverter.dateToIso(tomorrow));
    }
  }, [isPlannedStatus, isTemporal, updateField]);

  const inputStyle = {
    width: '100%',
    padding: '8px',
    borderRadius: '4px',
    border: '1px solid #ddd',
    fontSize: '14px'
  };

  const labelStyle = {
    display: 'block' as const,
    marginBottom: '4px',
    fontSize: '14px',
    fontWeight: '500' as const
  };

  const fieldStyle = {
    marginBottom: '16px'
  };

  const hintStyle = {
    fontSize: '12px',
    color: '#666',
    marginTop: '4px'
  };

  const cardStyle = {
    marginTop: '24px',
    padding: '16px',
    backgroundColor: '#f8f9fa',
    border: '1px solid #e9ecef',
    borderRadius: '4px'
  };

  return (
    <>
      <div style={fieldStyle}>
        <label style={labelStyle}>
          组织编码
        </label>
        <input
          type="text"
          value={formData.code}
          onChange={(e) => updateField('code', e.target.value)}
          disabled={true}
          placeholder="系统自动生成编码"
          style={{ ...inputStyle, backgroundColor: '#f5f5f5', cursor: 'not-allowed' }}
          data-testid="form-field-code"
        />
        <div style={hintStyle}>
          {isEditing ? "编码不可修改" : "系统将自动生成唯一编码"}
        </div>
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>
          组织名称 *
        </label>
        <input
          type="text"
          name="name"
          value={formData.name}
          onChange={(e) => updateField('name', e.target.value)}
          placeholder="请输入组织名称"
          required
          style={inputStyle}
          data-testid="form-field-name"
        />
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>
          组织类型 *
        </label>
        <select
          name="unit_type"
          value={formData.unit_type}
          onChange={(e) => updateField('unit_type', e.target.value)}
          style={inputStyle}
          data-testid="form-field-unit-type"
        >
          <option value="DEPARTMENT">部门</option>
          <option value="COST_CENTER">成本中心</option>
          <option value="COMPANY">公司</option>
          <option value="PROJECT_TEAM">项目团队</option>
        </select>
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>
          上级组织编码
        </label>
        <input
          type="text"
          value={formData.parent_code}
          onChange={(e) => updateField('parent_code', e.target.value)}
          placeholder="请输入上级组织编码"
          style={inputStyle}
          data-testid="form-field-parent-code"
        />
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>
          组织层级
        </label>
        <input
          type="number"
          value={formData.level}
          disabled={true}
          style={{ ...inputStyle, backgroundColor: '#f5f5f5', cursor: 'not-allowed' }}
          data-testid="form-field-level"
        />
        <div style={hintStyle}>
          层级由上级组织关系自动计算，不可手动修改
        </div>
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>
          状态 *
        </label>
        <select
          value={formData.status}
          onChange={(e) => updateField('status', e.target.value)}
          style={inputStyle}
          data-testid="form-field-status"
        >
          <option value="ACTIVE">激活</option>
          <option value="INACTIVE">停用</option>
          <option value="PLANNED">计划中</option>
        </select>
        {isPlannedStatus && (
          <div style={hintStyle}>
            计划 计划中的组织将自动启用组织详情功能
          </div>
        )}
      </div>

      <div style={fieldStyle}>
        <label style={labelStyle}>
          排序
        </label>
        <input
          type="number"
          value={formData.sort_order}
          onChange={(e) => updateField('sort_order', parseInt(e.target.value) || 0)}
          min="0"
          style={inputStyle}
          data-testid="form-field-sort-order"
        />
      </div>

      <div style={{ ...fieldStyle, marginBottom: '24px' }}>
        <label style={labelStyle}>
          描述
        </label>
        <textarea
          name="description"
          value={formData.description}
          onChange={(e) => updateField('description', e.target.value)}
          placeholder="请输入组织描述"
          rows={3}
          style={inputStyle}
          data-testid="form-field-description"
        />
      </div>

      {/* 组织详情功能区域 - 完全移除Canvas Kit组件 */}
      {enableTemporalFeatures && (
        <div style={cardStyle}>
          <div style={{ marginBottom: '16px' }}>
            <h3 style={{ fontSize: '16px', fontWeight: 'bold', margin: 0 }}>
              设置 组织详情设置
            </h3>
            <p style={{ fontSize: '12px', color: '#666', margin: '4px 0 0 0' }}>
              配置组织的生效和失效时间，实现精确的组织详情
            </p>
          </div>

          <div style={fieldStyle}>
            <label style={labelStyle}>
              <input
                type="checkbox"
                checked={isTemporal}
                onChange={(e) => updateField('is_temporal', e.target.checked)}
                disabled={isPlannedStatus}
                data-testid="form-field-is-temporal"
                style={{ marginRight: '8px' }}
              />
              启用组织详情
            </label>
            {isPlannedStatus && (
              <div style={hintStyle}>
                计划中的组织必须启用组织详情
              </div>
            )}
          </div>

          {isTemporal && (
            <>
              <div style={fieldStyle}>
                <label style={labelStyle}>
                  生效时间 *
                </label>
                <input
                  type="date"
                  value={formatDateForInput(formData.effective_from as string)}
                  onChange={(e) => updateField('effective_from', e.target.value)}
                  style={inputStyle}
                  data-testid="form-field-effective-from"
                />
                <div style={hintStyle}>
                  组织开始生效的日期和时间
                </div>
              </div>

              <div style={fieldStyle}>
                <label style={labelStyle}>
                  失效时间
                </label>
                <input
                  type="date"
                  value={formatDateForInput(formData.effective_to as string)}
                  onChange={(e) => updateField('effective_to', e.target.value)}
                  style={inputStyle}
                  data-testid="form-field-effective-to"
                />
                <div style={hintStyle}>
                  组织停止生效的日期和时间（可选，留空表示永久生效）
                </div>
              </div>

              <div style={fieldStyle}>
                <label style={labelStyle}>
                  变更原因 *
                </label>
                <textarea
                  value={formData.change_reason as string || ''}
                  onChange={(e) => updateField('change_reason', e.target.value)}
                  placeholder="请输入此次变更的原因和背景..."
                  rows={2}
                  style={inputStyle}
                  data-testid="form-field-change-reason"
                />
                <div style={hintStyle}>
                  详细说明此次组织变更的原因，便于历史追溯
                </div>
              </div>

              {/* 高级时态设置 */}
              <div style={{ marginTop: '16px' }}>
                <button
                  type="button"
                  onClick={() => setShowAdvancedTemporal(!showAdvancedTemporal)}
                  style={{ 
                    background: 'none', 
                    border: 'none', 
                    padding: 0, 
                    cursor: 'pointer',
                    textDecoration: 'underline',
                    color: '#1976d2',
                    fontSize: '12px'
                  }}
                  data-testid="toggle-advanced-temporal"
                >
                  {showAdvancedTemporal ? '隐藏高级设置 ▲' : '显示高级设置 ▼'}
                </button>
              </div>

              {showAdvancedTemporal && (
                <div style={{ 
                  marginTop: '16px', 
                  padding: '16px', 
                  backgroundColor: '#fff', 
                  border: '1px solid #dee2e6', 
                  borderRadius: '4px' 
                }}>
                  <p style={{ fontSize: '12px', color: '#666', margin: '0 0 16px 0' }}>
                    高级时态设置（适用于复杂的组织架构变更场景）
                  </p>
                  
                  <div style={{ marginBottom: '8px' }}>
                    <label style={labelStyle}>
                      当前时态模式
                    </label>
                    <div style={{ 
                      padding: '4px 8px', 
                      backgroundColor: temporalMode === 'current' ? '#e8f5e8' : temporalMode === 'historical' ? '#e8f0ff' : '#fff3cd',
                      borderRadius: '4px',
                      display: 'inline-block',
                      fontSize: '12px'
                    }}>
                      {temporalMode === 'current' ? '刷新 当前模式' : 
                       temporalMode === 'historical' ? '📖 历史模式' : 
                       '计划 规划模式'}
                    </div>
                    <div style={hintStyle}>
                      当前的时态查询模式，影响数据的显示和编辑行为
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      )}
    </>
  );
};