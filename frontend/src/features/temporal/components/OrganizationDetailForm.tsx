/**
 * 组织详情表单组件 (纯日期生效模型)
 * 用于查看和编辑组织的详细信息
 */
import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { Select } from '@workday/canvas-kit-react/select';
import { Checkbox } from '@workday/canvas-kit-react/checkbox';
import { Badge } from '../../../shared/components/Badge';
import { Card } from '@workday/canvas-kit-react/card';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import type { TemporalOrganizationRecord } from '../../../shared/hooks/useTemporalAPI';

export interface OrganizationDetailFormProps {
  /** 组织记录 */
  record: TemporalOrganizationRecord;
  /** 是否处于编辑模式 */
  isEditing: boolean;
  /** 字段变更回调 */
  onFieldChange: (field: keyof TemporalOrganizationRecord, value: string | number | boolean) => void;
}

/**
 * 组织详情表单组件
 */
export const OrganizationDetailForm: React.FC<OrganizationDetailFormProps> = ({
  record,
  isEditing,
  onFieldChange
}) => {
  // 组织类型选项
  const unitTypeOptions = [
    { value: 'COMPANY', label: '公司' },
    { value: 'DEPARTMENT', label: '部门' },
    { value: 'COST_CENTER', label: '成本中心' },
    { value: 'PROJECT_TEAM', label: '项目团队' },
  ];

  // 状态选项
  const statusOptions = [
    { value: 'ACTIVE', label: '启用' },
    { value: 'INACTIVE', label: '停用' },
    { value: 'PLANNED', label: '计划中' },
  ];

  // 获取状态对应的颜色
  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'positive';
      case 'PLANNED': return 'caution';
      case 'INACTIVE': return 'neutral';
      default: return 'neutral';
    }
  };

  return (
    <Box>
      {/* 基础信息卡片 */}
      <Card marginBottom={space.l} padding={space.m}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          详情 基础信息
        </Text>

        <Flex gap={space.m} marginBottom={space.m} flexDirection="row">
          {/* 组织代码 */}
          <Box flex={1}>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              组织代码
            </Text>
            <TextInput
              value={record.code}
              disabled={true}
            />
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              系统自动生成，不可修改
            </Text>
          </Box>

          {/* 租户ID */}
          <Box flex={1}>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              租户ID
            </Text>
            <TextInput
              value={record.tenant_id || ''}
              disabled={true}
            />
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              系统分配的租户标识
            </Text>
          </Box>
        </Flex>

        <Flex gap={space.m} marginBottom={space.m} flexDirection="row">
          {/* 组织名称 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              组织名称 *
            </Text>
            <TextInput
              value={record.name}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('name', e.target.value)}
              placeholder="请输入组织名称"
            />
            {isEditing && (
              <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
                必填字段，建议使用简洁明确的名称
              </Text>
            )}
          </Box>

          {/* 组织类型 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              组织类型 *
            </Text>
            {isEditing ? (
              <Select items={unitTypeOptions}>
                <Select.Input 
                  value={record.unit_type}
                  onChange={(e) => onFieldChange('unit_type', e.target.value)}
                />
                <Select.Popper>
                  <Select.Card>
                    <Select.List>
                      {(option: any) => (
                        <Select.Item key={option.value}>
                          {option.label}
                        </Select.Item>
                      )}
                    </Select.List>
                  </Select.Card>
                </Select.Popper>
              </Select>
            ) : (
              <Text>{unitTypeOptions.find(opt => opt.value === record.unit_type)?.label || record.unit_type}</Text>
            )}
          </Box>

          {/* 状态 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              状态 *
            </Text>
            {isEditing ? (
              <Select items={statusOptions}>
                <Select.Input 
                  value={record.status}
                  onChange={(e) => onFieldChange('status', e.target.value)}
                />
                <Select.Popper>
                  <Select.Card>
                    <Select.List>
                      {(option: any) => (
                        <Select.Item key={option.value}>
                          {option.label}
                        </Select.Item>
                      )}
                    </Select.List>
                  </Select.Card>
                </Select.Popper>
              </Select>
            ) : (
              <Box paddingTop={space.xs}>
                <Badge variant={getStatusBadgeVariant(record.status)}>
                  {statusOptions.find(opt => opt.value === record.status)?.label || record.status}
                </Badge>
              </Box>
            )}
          </Box>
        </Flex>

        <Flex gap={space.m} marginBottom={space.m} flexDirection="row">
          {/* 层级 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              组织层级
            </Text>
            <TextInput
              type="number"
              value={record.level.toString()}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('level', parseInt(e.target.value) || 0)}
              min="0"
              max="10"
            />
          </Box>

          {/* 排序顺序 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              排序顺序
            </Text>
            <TextInput
              type="number"
              value={record.sort_order.toString()}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('sort_order', parseInt(e.target.value) || 0)}
              min="0"
            />
          </Box>

          {/* 组织路径 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              组织路径
            </Text>
            <TextInput
              value={record.path}
              disabled={true}
            />
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              系统自动维护的层级路径
            </Text>
          </Box>
        </Flex>

        {/* 描述 */}
        <Box>
          <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
            描述
          </Text>
          <TextArea
            value={record.description || ''}
            disabled={!isEditing}
            onChange={(e) => isEditing && onFieldChange('description', e.target.value)}
            rows={3}
            placeholder="请输入组织描述信息..."
          />
          {isEditing && (
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              可选字段，用于说明该组织的职责和功能
            </Text>
          )}
        </Box>
      </Card>

      {/* 时态管理信息卡片 */}
      <Card marginBottom={space.l} padding={space.m}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          ⏰ 时态管理信息
        </Text>

        <Flex gap={space.m} marginBottom={space.m} flexDirection="row">
          {/* 生效日期 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              生效日期 *
            </Text>
            <TextInput
              type="date"
              value={record.effective_date?.slice(0, 10) || ''}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('effective_date', e.target.value + 'T00:00:00Z')}
            />
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              该记录开始生效的日期
            </Text>
          </Box>

          {/* 结束日期 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              结束日期
            </Text>
            <TextInput
              type="date"
              value={record.end_date?.slice(0, 10) || ''}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('end_date', e.target.value ? e.target.value + 'T00:00:00Z' : '')}
            />
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              可选，留空表示持续有效
            </Text>
          </Box>
        </Flex>

        {/* 当前有效状态 */}
        <Box marginBottom={space.m}>
          <Flex alignItems="center" gap={space.s}>
            <Checkbox
              checked={record.is_current}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('is_current', e.target.checked)}
              label="当前有效记录"
            />
            
            {record.is_current && (
              <Badge variant="positive" size="small">
                当前生效
              </Badge>
            )}
          </Flex>
          <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
            标识该记录是否为当前有效的版本
          </Text>
        </Box>

        {/* 变更原因 */}
        <Box>
          <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
            变更原因
          </Text>
          <TextArea
            value={record.change_reason || ''}
            disabled={!isEditing}
            onChange={(e) => isEditing && onFieldChange('change_reason', e.target.value)}
            rows={2}
            placeholder="请输入变更原因..."
          />
          <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
            记录本次变更的具体原因，便于后续审计追踪
          </Text>
        </Box>
      </Card>

      {/* 系统信息卡片 */}
      <Card marginBottom={space.l} padding={space.m}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          🔍 系统信息
        </Text>

        <Flex gap={space.m} marginBottom={space.m} flexDirection="row">
          {/* 创建时间 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              创建时间
            </Text>
            <TextInput
              value={record.created_at ? new Date(record.created_at).toLocaleString('zh-CN') : ''}
              disabled={true}
            />
          </Box>

          {/* 更新时间 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              最后更新时间
            </Text>
            <TextInput
              value={record.updated_at ? new Date(record.updated_at).toLocaleString('zh-CN') : ''}
              disabled={true}
            />
          </Box>
        </Flex>

        <Flex gap={space.m} flexDirection="row">
          {/* 批准人 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              批准人
            </Text>
            <TextInput
              value={record.approved_by || '暂无'}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('approved_by', e.target.value)}
              placeholder="请输入批准人"
            />
          </Box>

          {/* 批准时间 */}
          <Box>
            <Text fontSize="small" marginBottom={space.xs} fontWeight="medium">
              批准时间
            </Text>
            <TextInput
              type="datetime-local"
              value={record.approved_at ? new Date(record.approved_at).toISOString().slice(0, 16) : ''}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('approved_at', e.target.value ? new Date(e.target.value).toISOString() : '')}
            />
          </Box>
        </Flex>
      </Card>

      {/* 数据有效性验证提示 */}
      {isEditing && (
        <Card padding={space.m} backgroundColor={colors.blueberry50}>
          <Text fontSize="small" fontWeight="bold" marginBottom={space.xs}>
            提示 编辑提示
          </Text>
          <Box as="ul" marginLeft={space.m}>
            <Box as="li" marginBottom={space.xs}>
              <Text fontSize="small">
                带 * 的字段为必填项，请确保填写完整
              </Text>
            </Box>
            <Box as="li" marginBottom={space.xs}>
              <Text fontSize="small">
                生效日期不能晚于结束日期
              </Text>
            </Box>
            <Box as="li" marginBottom={space.xs}>
              <Text fontSize="small">
                同一组织在同一时间点只能有一个有效记录
              </Text>
            </Box>
            <Box as="li">
              <Text fontSize="small">
                变更原因将显示在时间轴中，建议填写清晰的说明
              </Text>
            </Box>
          </Box>
        </Card>
      )}
    </Box>
  );
};

export default OrganizationDetailForm;