/**
 * 时态导航栏组件
 * 提供时态模式切换、时间点选择等核心功能
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { Text } from '@workday/canvas-kit-react/text';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import { colors, space, borderRadius } from '@workday/canvas-kit-react/tokens';
import { useTemporalMode, useTemporalQueryState } from '../../../shared/hooks/useTemporalQuery';
import { useTemporalActions, temporalSelectors } from '../../../shared/stores/temporalStore';
import type { TemporalMode } from '../../../shared/types/temporal';
import { DateTimePicker } from './DateTimePicker';
import { TemporalSettings } from './TemporalSettings';

export interface TemporalNavbarProps {
  /** 是否显示高级设置 */
  showAdvancedSettings?: boolean;
  /** 是否紧凑模式 */
  compact?: boolean;
  /** 自定义样式类名 */
  className?: string;
  /** 模式切换回调 */
  onModeChange?: (mode: TemporalMode) => void;
}

/**
 * 时态导航栏组件
 */
export const TemporalNavbar: React.FC<TemporalNavbarProps> = ({
  showAdvancedSettings = true,
  compact = false,
  className,
  onModeChange
}) => {
  const [showDatePicker, setShowDatePicker] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const openSettings = () => setSettingsOpen(true);
  const closeSettings = () => setSettingsOpen(false);

  // 时态状态和操作
  const { 
    mode, 
    switchToCurrent, 
    switchToHistorical, 
    switchToPlanning,
    isCurrent
  } = useTemporalMode();
  
  const { loading, error, cacheStats, refreshCache } = useTemporalQueryState();
  const { setError } = useTemporalActions();
  const queryParams = temporalSelectors.useQueryParams();

  // 模式切换处理
  const handleModeChange = useCallback(async (newMode: TemporalMode) => {
    try {
      setError(null);
      
      switch (newMode) {
        case 'current':
          await switchToCurrent();
          break;
        case 'historical':
          setShowDatePicker(true);
          return; // 等待用户选择日期
        case 'planning':
          await switchToPlanning();
          break;
      }
      
      onModeChange?.(newMode);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to switch mode');
    }
  }, [switchToCurrent, switchToPlanning, setError, onModeChange]);

  // 历史模式日期选择
  const handleHistoricalDateSelect = useCallback(async (date: string) => {
    try {
      setError(null);
      await switchToHistorical(date);
      setShowDatePicker(false);
      onModeChange?.('historical');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to set historical date');
    }
  }, [switchToHistorical, setError, onModeChange]);

  // 刷新缓存
  const handleRefreshCache = useCallback(async () => {
    try {
      setError(null);
      await refreshCache();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to refresh cache');
    }
  }, [refreshCache, setError]);

  // 获取模式显示信息
  const getModeDisplay = () => {
    switch (mode) {
      case 'current':
        return {
          label: '当前视图',
          color: colors.greenFresca600,
          icon: "⏰",
          description: '显示当前有效的组织架构'
        };
      case 'historical':
        return {
          label: '历史视图',
          color: colors.blueberry600,
          icon: "📜",
          description: `显示历史时间点的组织架构`
        };
      case 'planning':
        return {
          label: '规划视图',
          color: colors.peach600,
          icon: "📅",
          description: '显示未来规划的组织架构变更'
        };
    }
  };

  const modeDisplay = getModeDisplay();

  return (
    <Box
      className={className}
      backgroundColor={colors.soap200}
      padding={compact ? space.s : space.m}
      borderRadius={borderRadius.m}
      boxShadow="0 2px 4px rgba(0,0,0,0.1)"
    >
      <Flex alignItems="center" gap={space.m}>
        {/* 模式切换按钮组 */}
        <Flex gap={space.xs}>
          <Tooltip title="当前有效的组织架构">
            <SecondaryButton
              style={{
                backgroundColor: isCurrent ? colors.blueberry600 : 'transparent',
                color: isCurrent ? 'white' : colors.blueberry600
              }}
              size={compact ? 'small' : 'medium'}
              onClick={() => handleModeChange('current')}
              disabled={loading.organizations}
            >
              ⏰
              {!compact && '当前'}
            </SecondaryButton>
          </Tooltip>
          
          <Tooltip title="查看历史时点的组织架构">
            <SecondaryButton
              size={compact ? 'small' : 'medium'}
              onClick={() => handleModeChange('historical')}
              disabled={loading.organizations}
            >
              📜
              {!compact && '历史'}
            </SecondaryButton>
          </Tooltip>
          
          <Tooltip title="查看未来规划的组织变更">
            <SecondaryButton
              size={compact ? 'small' : 'medium'}
              onClick={() => handleModeChange('planning')}
              disabled={loading.organizations}
            >
              📅
              {!compact && '规划'}
            </SecondaryButton>
          </Tooltip>
        </Flex>

        {/* 当前模式状态显示 */}
        <Flex alignItems="center" gap={space.s}>
          <Flex alignItems="center" gap={space.s}>
            <Text fontSize="small">{modeDisplay.icon}</Text>
            <Text
              fontSize="small"
              color={modeDisplay.color}
              fontWeight="medium"
            >
              {modeDisplay.label}
            </Text>
          </Flex>
          
          {!compact && (
            <Text fontSize="small" color={colors.licorice500}>
              {modeDisplay.description}
            </Text>
          )}
        </Flex>

        {/* 操作按钮区域 */}
        <Flex marginLeft="auto" alignItems="center" gap={space.s}>
          {/* 缓存状态指示器 */}
          {!compact && cacheStats.totalCacheSize > 0 && (
            <Tooltip title={`缓存: ${cacheStats.organizationsCount} 组织, ${cacheStats.timelinesCount} 时间线`}>
              <Flex alignItems="center" gap={space.xs}>
                <Text fontSize="small">💾</Text>
                <Text fontSize="small" color={colors.licorice400}>
                  {cacheStats.totalCacheSize}
                </Text>
              </Flex>
            </Tooltip>
          )}

          {/* 刷新按钮 */}
          <Tooltip title="刷新数据缓存">
            <SecondaryButton
              size={compact ? 'small' : 'medium'}
              onClick={handleRefreshCache}
              disabled={loading.organizations || loading.timeline}
            >
              🔄
            </SecondaryButton>
          </Tooltip>

          {/* 高级设置按钮 */}
          {showAdvancedSettings && (
            <Tooltip title="时态查询设置">
              <SecondaryButton
                size={compact ? 'small' : 'medium'}
                onClick={openSettings}
              >
                ⚙️
              </SecondaryButton>
            </Tooltip>
          )}
        </Flex>
      </Flex>

      {/* 错误提示 */}
      {error && (
        <Box marginTop={space.s}>
          <Text color={colors.cinnamon600} fontSize="small">
            ⚠️ {error}
          </Text>
        </Box>
      )}

      {/* 加载状态指示器 */}
      {(loading.organizations || loading.timeline) && (
        <Box marginTop={space.s}>
          <Text color={colors.blueberry600} fontSize="small">
            🔄 {loading.organizations ? '加载组织数据...' : '加载时间线数据...'}
          </Text>
        </Box>
      )}

      {/* 日期时间选择器弹窗 */}
      {showDatePicker && (
        <DateTimePicker
          isOpen={showDatePicker}
          onClose={() => setShowDatePicker(false)}
          onSelect={handleHistoricalDateSelect}
          defaultDate={new Date().toISOString().split('T')[0]}
          title="选择历史查看时点"
        />
      )}

      {/* 高级设置弹窗 */}
      {settingsOpen && (
        <TemporalSettings
          isOpen={settingsOpen}
          onClose={closeSettings}
          queryParams={queryParams}
        />
      )}
    </Box>
  );
};

export default TemporalNavbar;