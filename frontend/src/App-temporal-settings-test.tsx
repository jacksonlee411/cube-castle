/**
 * 时态设置面板功能测试应用
 * 验证TemporalSettings组件的配置功能和用户交互
 */
import React, { useState, useCallback, useEffect } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Card } from '@workday/canvas-kit-react/card';
import { Checkbox } from '@workday/canvas-kit-react/checkbox';
import { Badge } from '@workday/canvas-kit-react/badge';
import { Modal, useModalModel } from '@workday/canvas-kit-react/modal';

import type { TemporalQueryParams, EventType } from './shared/types/temporal';

// 创建React Query客户端
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false,
      staleTime: 5 * 60 * 1000,
    },
  },
});

/**
 * 简化的时态设置组件
 */
interface SimpleTemporalSettingsProps {
  isOpen: boolean;
  onClose: () => void;
  queryParams: TemporalQueryParams;
  onSettingsChange: (params: TemporalQueryParams) => void;
}

const SimpleTemporalSettings: React.FC<SimpleTemporalSettingsProps> = ({
  isOpen,
  onClose,
  queryParams,
  onSettingsChange
}) => {
  const model = useModalModel();
  const [localParams, setLocalParams] = useState<TemporalQueryParams>(queryParams);
  const [hasChanges, setHasChanges] = useState(false);

  // Modal state management
  useEffect(() => {
    if (isOpen && model.state.visibility !== 'visible') {
      model.events.show();
    } else if (!isOpen && model.state.visibility === 'visible') {
      model.events.hide();
    }
  }, [isOpen, model]);

  // 重置参数当props变化时
  useEffect(() => {
    setLocalParams(queryParams);
    setHasChanges(false);
  }, [queryParams]);

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
  const handleApply = useCallback(() => {
    onSettingsChange(localParams);
    setHasChanges(false);
    model.events.hide();
    onClose();
  }, [localParams, onSettingsChange, model, onClose]);

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

  // 取消设置
  const handleCancel = useCallback(() => {
    setLocalParams(queryParams);
    setHasChanges(false);
    model.events.hide();
    onClose();
  }, [queryParams, model, onClose]);

  const formatDateTimeLocal = (dateStr?: string) => {
    if (!dateStr) return '';
    try {
      return new Date(dateStr).toISOString().slice(0, 16);
    } catch {
      return '';
    }
  };

  return (
    <Modal model={model}>
      <Modal.Overlay>
        <Modal.Card width={800} data-testid="temporal-settings">
          <Modal.CloseIcon aria-label="关闭" onClick={handleCancel} />
          <Modal.Heading>
            <Flex alignItems="center" gap="s">
              <Text>⚙️ 时态查询设置</Text>
              {hasChanges && (
                <Badge color="peach600">
                  有未保存的更改
                </Badge>
              )}
            </Flex>
          </Modal.Heading>
          <Modal.Body>
            <Box padding="m">
              {/* 基础设置 */}
              <Box marginBottom="l">
                <Text typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                  基础设置
                </Text>

                {/* 查询时间点 */}
                <FormField marginBottom="m">
                  <FormField.Label>查询时间点 (As Of Date)</FormField.Label>
                  <FormField.Field>
                    <FormField.Input
                      as={TextInput}
                      type="datetime-local"
                      value={formatDateTimeLocal(localParams.asOfDate)}
                      onChange={(e) => updateLocalParams({ 
                        asOfDate: e.target.value ? new Date(e.target.value).toISOString() : undefined 
                      })}
                    />
                  </FormField.Field>
                  <FormField.Hint>
                    在历史模式下，显示此时间点有效的数据
                  </FormField.Hint>
                </FormField>

                {/* 查询限制 */}
                <FormField marginBottom="m">
                  <FormField.Label>查询结果限制</FormField.Label>
                  <FormField.Field>
                    <select
                      value={String(localParams.limit || 50)}
                      onChange={(e) => updateLocalParams({ limit: parseInt(e.target.value) })}
                      style={{ 
                        width: '100%', 
                        padding: '8px', 
                        borderRadius: '4px', 
                        border: '1px solid #ddd' 
                      }}
                    >
                      <option value="10">10 条</option>
                      <option value="20">20 条</option>
                      <option value="50">50 条</option>
                      <option value="100">100 条</option>
                      <option value="200">200 条</option>
                    </select>
                  </FormField.Field>
                </FormField>

                {/* 包含停用数据 */}
                <FormField marginBottom="m">
                  <FormField.Field>
                    <Checkbox
                      checked={localParams.includeInactive || false}
                      onChange={(e) => updateLocalParams({ includeInactive: e.target.checked })}
                    >
                      包含停用/失效的组织数据
                    </Checkbox>
                  </FormField.Field>
                  <FormField.Hint>
                    勾选后将显示已停用或失效的组织单元
                  </FormField.Hint>
                </FormField>
              </Box>

              {/* 分隔线 */}
              <Box height="1px" backgroundColor="#e9ecef" marginY="l" />

              {/* 时间范围设置 */}
              <Box marginBottom="l">
                <Text typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                  时间范围筛选
                </Text>

                <Flex gap="m" marginBottom="s">
                  <FormField flex="1">
                    <FormField.Label>开始时间</FormField.Label>
                    <FormField.Field>
                      <FormField.Input
                        as={TextInput}
                        type="datetime-local"
                        value={formatDateTimeLocal(localParams.dateRange?.start)}
                        onChange={(e) => handleDateRangeChange('start', 
                          e.target.value ? new Date(e.target.value).toISOString() : ''
                        )}
                      />
                    </FormField.Field>
                  </FormField>

                  <FormField flex="1">
                    <FormField.Label>结束时间</FormField.Label>
                    <FormField.Field>
                      <FormField.Input
                        as={TextInput}
                        type="datetime-local"
                        value={formatDateTimeLocal(localParams.dateRange?.end)}
                        onChange={(e) => handleDateRangeChange('end', 
                          e.target.value ? new Date(e.target.value).toISOString() : ''
                        )}
                      />
                    </FormField.Field>
                  </FormField>
                </Flex>

                <Text typeLevel="subtext.small" color="hint">
                  用于筛选指定时间范围内的历史记录和时间线事件
                </Text>
              </Box>

              {/* 分隔线 */}
              <Box height="1px" backgroundColor="#e9ecef" marginY="l" />

              {/* 事件类型筛选 */}
              <Box marginBottom="l">
                <Text typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
                  事件类型筛选
                </Text>

                <Text typeLevel="subtext.small" color="hint" marginBottom="s">
                  选择要显示的时间线事件类型:
                </Text>

                <Box
                  display="grid"
                  gridTemplateColumns="repeat(auto-fit, minmax(150px, 1fr))"
                  gap="s"
                  marginBottom="s"
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

                <Text typeLevel="subtext.small" color="hint">
                  未选择任何类型时，将显示所有事件类型
                </Text>
              </Box>

              {/* 操作按钮 */}
              <Flex justifyContent="space-between" alignItems="center" paddingTop="m">
                <SecondaryButton onClick={handleReset}>
                  🔄 重置为默认
                </SecondaryButton>

                <Flex gap="s">
                  <SecondaryButton onClick={handleCancel}>
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
            </Box>
          </Modal.Body>
        </Modal.Card>
      </Modal.Overlay>
    </Modal>
  );
};

/**
 * 时态设置测试组件
 */
const TemporalSettingsTest: React.FC = () => {
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [currentParams, setCurrentParams] = useState<TemporalQueryParams>({
    mode: 'current',
    asOfDate: new Date().toISOString(),
    dateRange: {
      start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
      end: new Date().toISOString()
    },
    limit: 50,
    includeInactive: false,
    eventTypes: []
  });

  const [settingsHistory, setSettingsHistory] = useState<TemporalQueryParams[]>([]);

  const handleOpenSettings = useCallback(() => {
    setIsSettingsOpen(true);
  }, []);

  const handleCloseSettings = useCallback(() => {
    setIsSettingsOpen(false);
  }, []);

  const handleSettingsChange = useCallback((newParams: TemporalQueryParams) => {
    setCurrentParams(newParams);
    setSettingsHistory(prev => [newParams, ...prev].slice(0, 5)); // 保留最近5次设置
    console.log('时态设置已更新:', newParams);
  }, []);

  const handleClearCache = useCallback(() => {
    if (confirm('确定要清除所有缓存吗？')) {
      alert('缓存已清除（模拟操作）');
    }
  }, []);

  const formatDateTime = (dateStr?: string) => {
    if (!dateStr) return '未设置';
    try {
      return new Date(dateStr).toLocaleString('zh-CN');
    } catch {
      return '无效日期';
    }
  };

  return (
    <Box padding="l">
      <Text as="h1" typeLevel="heading.large" marginBottom="l">
        ⚙️ 时态设置面板功能测试
      </Text>
      
      <Text typeLevel="body.medium" marginBottom="m">
        测试TemporalSettings组件的配置功能，包括查询参数设置、事件类型筛选和用户偏好保存。
      </Text>

      {/* 当前设置显示 */}
      <Card marginBottom="l" padding="m">
        <Text as="h2" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
          📊 当前时态设置
        </Text>
        
        <Box display="grid" gridTemplateColumns="repeat(auto-fit, minmax(300px, 1fr))" gap="m">
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">基础配置</Text>
            <Text typeLevel="body.small">查询模式: {currentParams.mode}</Text>
            <Text typeLevel="body.small">时间点: {formatDateTime(currentParams.asOfDate)}</Text>
            <Text typeLevel="body.small">查询限制: {currentParams.limit} 条</Text>
            <Text typeLevel="body.small">包含停用: {currentParams.includeInactive ? '是' : '否'}</Text>
          </Box>
          
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">时间范围</Text>
            <Text typeLevel="body.small">开始: {formatDateTime(currentParams.dateRange?.start)}</Text>
            <Text typeLevel="body.small">结束: {formatDateTime(currentParams.dateRange?.end)}</Text>
          </Box>
          
          <Box>
            <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">事件类型</Text>
            <Flex gap="s" flexWrap="wrap">
              {(currentParams.eventTypes || []).length === 0 ? (
                <Badge color="licorice400">全部类型</Badge>
              ) : (
                currentParams.eventTypes?.map(type => (
                  <Badge key={type} color="blueberry600" size="small">{type}</Badge>
                ))
              )}
            </Flex>
          </Box>
        </Box>
      </Card>

      {/* 控制面板 */}
      <Card marginBottom="l" padding="m">
        <Text as="h2" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
          🎛️ 测试控制面板
        </Text>
        
        <Flex gap="m" alignItems="center" marginBottom="m">
          <PrimaryButton onClick={handleOpenSettings}>
            ⚙️ 打开时态设置
          </PrimaryButton>
          
          <SecondaryButton onClick={handleClearCache}>
            🗑️ 清除缓存
          </SecondaryButton>
          
          <SecondaryButton onClick={() => setSettingsHistory([])}>
            📝 清空历史记录
          </SecondaryButton>
        </Flex>
      </Card>

      {/* 设置历史 */}
      {settingsHistory.length > 0 && (
        <Card marginBottom="l" padding="m">
          <Text as="h2" typeLevel="subtext.large" fontWeight="bold" marginBottom="m">
            📚 设置变更历史
          </Text>
          
          <Box maxHeight="300px" overflow="auto">
            {settingsHistory.map((params, index) => (
              <Box
                key={index}
                padding="s"
                marginBottom="s"
                style={{
                  backgroundColor: index === 0 ? '#f0f7ff' : '#f8f9fa',
                  borderRadius: '4px',
                  border: '1px solid #dee2e6'
                }}
              >
                <Flex justifyContent="space-between" alignItems="center" marginBottom="xs">
                  <Text typeLevel="subtext.small" fontWeight="bold">
                    设置 #{settingsHistory.length - index}
                    {index === 0 && <Badge color="greenFresca600" size="small" marginLeft="s">当前</Badge>}
                  </Text>
                  <Text typeLevel="subtext.small" color="hint">
                    {formatDateTime(params.asOfDate)}
                  </Text>
                </Flex>
                <Text typeLevel="subtext.small">
                  限制: {params.limit} 条 | 
                  停用: {params.includeInactive ? '包含' : '不包含'} | 
                  事件类型: {(params.eventTypes?.length || 0)} 个
                </Text>
              </Box>
            ))}
          </Box>
        </Card>
      )}

      {/* 功能验证要点 */}
      <Card marginBottom="l" padding="m" style={{ backgroundColor: '#f0f7ff', border: '1px solid #d1ecf1' }}>
        <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
          📋 时态设置功能验证要点
        </Text>
        <ul style={{ marginLeft: '20px', lineHeight: '1.6' }}>
          <li>✅ 设置面板的打开和关闭</li>
          <li>✅ 基础设置配置 (时间点、限制、包含停用)</li>
          <li>✅ 时间范围筛选设置</li>
          <li>✅ 事件类型多选筛选</li>
          <li>✅ 设置变更检测和提示</li>
          <li>✅ 应用设置和取消操作</li>
          <li>✅ 重置为默认值</li>
          <li>✅ 设置历史记录跟踪</li>
          <li>✅ 表单验证和错误处理</li>
          <li>✅ 响应式布局和用户体验</li>
        </ul>
      </Card>

      {/* 测试提示 */}
      <Card padding="m" style={{ backgroundColor: '#fff3cd', border: '1px solid #ffeaa7' }}>
        <Text typeLevel="subtext.medium" fontWeight="bold" marginBottom="s">
          💡 测试提示
        </Text>
        <ul style={{ marginLeft: '20px', lineHeight: '1.6' }}>
          <li><strong>设置面板</strong>: 点击"打开时态设置"按钮测试面板功能</li>
          <li><strong>基础设置</strong>: 修改查询时间点、结果限制、是否包含停用数据</li>
          <li><strong>时间范围</strong>: 设置开始和结束时间来筛选历史数据</li>
          <li><strong>事件筛选</strong>: 选择感兴趣的事件类型进行筛选</li>
          <li><strong>应用/取消</strong>: 测试设置的应用和取消功能</li>
          <li><strong>重置功能</strong>: 测试重置为默认设置</li>
          <li><strong>变更检测</strong>: 观察未保存更改的提示</li>
        </ul>
      </Card>

      {/* 时态设置组件 */}
      <SimpleTemporalSettings
        isOpen={isSettingsOpen}
        onClose={handleCloseSettings}
        queryParams={currentParams}
        onSettingsChange={handleSettingsChange}
      />
    </Box>
  );
};

/**
 * 时态设置测试应用
 */
export const TemporalSettingsTestApp: React.FC = () => {
  return (
    <QueryClientProvider client={queryClient}>
      <TemporalSettingsTest />
    </QueryClientProvider>
  );
};

export default TemporalSettingsTestApp;