/**
 * 时态设置组件
 * 提供时态查询的高级设置和配置选项
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Select } from '@workday/canvas-kit-react/select';
import { Checkbox } from '@workday/canvas-kit-react/checkbox';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { colors, space, borderRadius } from '@workday/canvas-kit-react/tokens';
import { useTemporalActions } from '../../../shared/stores/temporalStore';
import type { TemporalQueryParams, EventType } from '../../shared/types/temporal';

export interface TemporalSettingsProps {
  /** 是否显示弹窗 */
  isOpen: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 当前查询参数 */
  queryParams: TemporalQueryParams;
}

/**
 * 时态设置组件
 */
export const TemporalSettings: React.FC<TemporalSettingsProps> = ({
  isOpen,
  onClose,
  queryParams
}) => {
  // 本地状态
  const [localParams, setLocalParams] = useState<TemporalQueryParams>(queryParams);
  const [hasChanges, setHasChanges] = useState(false);

  // 时态操作
  const { setQueryParams, clearCache } = useTemporalActions();

  // 事件类型选项
  const eventTypeOptions: { value: EventType; label: string }[] = [
    { value: 'create', label: '创建' },
    { value: 'update', label: '更新' },
    { value: 'delete', label: '删除' },
    { value: 'activate', label: '激活' },
    { value: 'deactivate', label: '停用' },
    { value: 'restructure', label: '重组' },
    { value: 'merge', label: '合并' },
    { value: 'split', label: '拆分' },
    { value: 'transfer', label: '转移' },
    { value: 'rename', label: '重命名' }
  ];

  // 更新本地参数
  const updateLocalParams = useCallback((updates: Partial<TemporalQueryParams>) => {
    setLocalParams(prev => ({ ...prev, ...updates }));
    setHasChanges(true);
  }, []);

  // 处理日期范围变更
  const handleDateRangeChange = useCallback((field: 'start' | 'end', value: string) => {
    const currentRange = localParams.dateRange || { start: '', end: '' };
    updateLocalParams({
      dateRange: {
        ...currentRange,
        [field]: value
      }
    });
  }, [localParams.dateRange, updateLocalParams]);

  // 处理事件类型选择
  const handleEventTypeToggle = useCallback((eventType: EventType) => {
    const currentTypes = localParams.eventTypes || [];
    const newTypes = currentTypes.includes(eventType)
      ? currentTypes.filter(t => t !== eventType)
      : [...currentTypes, eventType];
    
    updateLocalParams({ eventTypes: newTypes });
  }, [localParams.eventTypes, updateLocalParams]);

  // 应用设置
  const handleApply = useCallback(async () => {
    try {
      setQueryParams(localParams);
      setHasChanges(false);
      onClose();
    } catch (error) {
      console.error('Failed to apply settings:', error);
    }
  }, [localParams, setQueryParams, onClose]);

  // 重置设置
  const handleReset = useCallback(() => {
    const defaultParams: TemporalQueryParams = {
      mode: 'current',
      asOfDate: new Date().toISOString(),
      dateRange: {
        start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString()
      },
      limit: 50,
      includeInactive: false,
      eventTypes: []
    };
    
    setLocalParams(defaultParams);
    setHasChanges(true);
  }, []);

  // 清除缓存
  const handleClearCache = useCallback(async () => {
    try {
      await clearCache();
      alert('缓存已清除');
    } catch (error) {
      console.error('Failed to clear cache:', error);
      alert('清除缓存失败');
    }
  }, [clearCache]);

  if (!isOpen) {
    return null;
  }

  return (
    <Modal onClose={onClose}>
      <Card 
        padding={space.l}
        minWidth="600px"
        maxWidth="800px"
        maxHeight="80vh"
        overflow="auto"
      >
        {/* 标题 */}
        <Flex alignItems="center" gap={space.s} marginBottom={space.l}>
          ⚙️
          <Text fontSize="large" fontWeight="bold">
            时态查询设置
          </Text>
          {hasChanges && (
            <Text color="peach600">
              有未保存的更改
            </Text>
          )}
        </Flex>

        {/* 基础设置 */}
        <Box marginBottom={space.l}>
          <Text fontSize="medium" fontWeight="medium" marginBottom={space.m}>
            基础设置
          </Text>

          {/* 查询时间点 */}
          <Box marginBottom={space.m}>
            <Text fontSize="small" marginBottom={space.s}>
              查询时间点 (As Of Date)
            </Text>
            <TextInput
              type="date"
              value={localParams.asOfDate ? 
                localParams.asOfDate.slice(0, 10) : 
                ''
              }
              onChange={(e) => updateLocalParams({ 
                asOfDate: e.target.value ? e.target.value : undefined 
              })}
            />
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              在历史模式下，显示此时间点有效的数据
            </Text>
          </Box>

          {/* 查询限制 */}
          <Box marginBottom={space.m}>
            <Text fontSize="small" marginBottom={space.s}>
              查询结果限制
            </Text>
            <Select
              value={String(localParams.limit || 50)}
              onChange={(value) => updateLocalParams({ limit: parseInt(value) })}
            >
              <MenuItem value="10">10 条</MenuItem>
              <MenuItem value="20">20 条</MenuItem>
              <MenuItem value="50">50 条</MenuItem>
              <MenuItem value="100">100 条</MenuItem>
              <MenuItem value="200">200 条</MenuItem>
            </Select>
          </Box>

          {/* 包含停用数据 */}
          <Box marginBottom={space.m}>
            <Checkbox
              checked={localParams.includeInactive || false}
              onChange={(e) => updateLocalParams({ includeInactive: e.target.checked })}
            >
              包含停用/失效的组织数据
            </Checkbox>
            <Text fontSize="small" color={colors.licorice500} marginTop={space.xs}>
              勾选后将显示已停用或失效的组织单元
            </Text>
          </Box>
        </Box>

        <Divider />

        {/* 时间范围设置 */}
        <Box marginBottom={space.l}>
          <Text fontSize="medium" fontWeight="medium" marginBottom={space.m}>
            时间范围筛选
          </Text>

          <Flex gap={space.m}>
            <Box flex="1">
              <Text fontSize="small" marginBottom={space.s}>
                开始时间
              </Text>
              <TextInput
                type="date"
                value={localParams.dateRange?.start ? 
                  localParams.dateRange.start.slice(0, 10) : 
                  ''
                }
                onChange={(e) => handleDateRangeChange('start', 
                  e.target.value ? e.target.value : ''
                )}
              />
            </Box>

            <Box flex="1">
              <Text fontSize="small" marginBottom={space.s}>
                结束时间
              </Text>
              <TextInput
                type="date"
                value={localParams.dateRange?.end ? 
                  localParams.dateRange.end.slice(0, 10) : 
                  ''
                }
                onChange={(e) => handleDateRangeChange('end', 
                  e.target.value ? e.target.value : ''
                )}
              />
            </Box>
          </Flex>

          <Text fontSize="small" color={colors.licorice500} marginTop={space.s}>
            用于筛选指定时间范围内的历史记录和时间线事件
          </Text>
        </Box>

        <Divider />

        {/* 事件类型筛选 */}
        <Box marginBottom={space.l}>
          <Text fontSize="medium" fontWeight="medium" marginBottom={space.m}>
            事件类型筛选
          </Text>

          <Text fontSize="small" color={colors.licorice600} marginBottom={space.s}>
            选择要显示的时间线事件类型:
          </Text>

          <Box
            display="grid"
            gridTemplateColumns="repeat(auto-fit, minmax(150px, 1fr))"
            gap={space.s}
          >
            {eventTypeOptions.map(option => (
              <Checkbox
                key={option.value}
                checked={(localParams.eventTypes || []).includes(option.value)}
                onChange={() => handleEventTypeToggle(option.value)}
              >
                {option.label}
              </Checkbox>
            ))}
          </Box>

          <Text fontSize="small" color={colors.licorice500} marginTop={space.s}>
            未选择任何类型时，将显示所有事件类型
          </Text>
        </Box>

        <Divider />

        {/* 缓存管理 */}
        <Box marginBottom={space.l}>
          <Text fontSize="medium" fontWeight="medium" marginBottom={space.m}>
            缓存管理
          </Text>

          <Flex alignItems="center" gap={space.s}>
            <SecondaryButton
              variant="secondary"
              size="small"
              onClick={handleClearCache}
            >
              清除所有缓存
            </SecondaryButton>
            <Text fontSize="small" color={colors.licorice500}>
              清除缓存会强制重新加载所有数据
            </Text>
          </Flex>
        </Box>

        {/* 操作按钮 */}
        <Flex justifyContent="space-between" alignItems="center">
          <SecondaryButton
            onClick={handleReset}
          >
            🔄 重置为默认
          </SecondaryButton>

          <Flex gap={space.s}>
            <SecondaryButton onClick={onClose}>
              取消
            </SecondaryButton>
            <PrimaryButton 
              onClick={handleApply}
              disabled={!hasChanges}
            >
              应用设置
            </PrimaryButton>
          </Flex>
        </Flex>
      </Card>
    </Modal>
  );
};

export default TemporalSettings;