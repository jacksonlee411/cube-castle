/**
 * 五状态生命周期管理状态选择器
 * 支持新的五状态系统：CURRENT, HISTORICAL, PLANNED, SUSPENDED, DELETED
 * 版本：v2.1 - 适配五状态生命周期管理系统
 * 创建时间：2025-08-18
 */
import React from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Select } from '@workday/canvas-kit-react/select';
import { FormField } from '@workday/canvas-kit-react/form-field';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { 
  checkCircleIcon, 
  clockIcon, 
  timelineAllIcon,  // 用 timelineAllIcon 替代 historyIcon
  clockPauseIcon,   // 用 clockPauseIcon 替代 pauseIcon
  trashIcon
} from '@workday/canvas-system-icons-web';
import { colors } from '@workday/canvas-kit-react/tokens';

// 五状态定义
export interface LifecycleStatus {
  key: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'SUSPENDED' | 'DELETED';
  label: string;
  description: string;
  color: string;
  icon: any;
  businessStatus?: 'ACTIVE' | 'SUSPENDED';
  dataStatus?: 'NORMAL' | 'DELETED';
}

// 五状态配置
export const LIFECYCLE_STATES: LifecycleStatus[] = [
  {
    key: 'CURRENT',
    label: '当前记录',
    description: '当前生效的组织状态，用户可正常访问',
    color: colors.greenApple600,
    icon: checkCircleIcon,
    businessStatus: 'ACTIVE',
    dataStatus: 'NORMAL'
  },
  {
    key: 'HISTORICAL', 
    label: '历史记录',
    description: '已过期的组织状态，保留用于审计和历史查询',
    color: colors.licorice400,
    icon: timelineAllIcon,  // 更新图标引用
    businessStatus: 'ACTIVE',
    dataStatus: 'NORMAL'
  },
  {
    key: 'PLANNED',
    label: '计划中',
    description: '未来生效的组织状态，尚未开始执行',
    color: colors.blueberry600,
    icon: clockIcon,
    businessStatus: 'ACTIVE',
    dataStatus: 'NORMAL'
  },
  {
    key: 'SUSPENDED',
    label: '已停用',
    description: '暂时停用但可能恢复的组织状态（仍可查询）',
    color: colors.cantaloupe600,
    icon: clockPauseIcon,  // 更新图标引用
    businessStatus: 'SUSPENDED',
    dataStatus: 'NORMAL'
  },
  {
    key: 'DELETED',
    label: '已删除',
    description: '软删除的组织状态，不参与业务流程但保留数据',
    color: colors.cinnamon600,
    icon: trashIcon,
    businessStatus: 'ACTIVE', // 删除前的业务状态
    dataStatus: 'DELETED'
  }
];

export interface FiveStateStatusSelectorProps {
  value?: string;
  onChange: (status: LifecycleStatus) => void;
  disabled?: boolean;
  includeDeleted?: boolean; // 是否显示删除状态选项
  label?: string;
  error?: string;
  required?: boolean;
}

/**
 * 五状态生命周期状态选择器组件
 */
export const FiveStateStatusSelector: React.FC<FiveStateStatusSelectorProps> = ({
  value,
  onChange,
  disabled = false,
  includeDeleted = false,
  label = '组织状态',
  error,
  required = false
}) => {
  // 过滤可选状态（通常不允许直接选择删除状态）
  const availableStates = LIFECYCLE_STATES.filter(state => 
    includeDeleted || state.key !== 'DELETED'
  );

  const selectedState = LIFECYCLE_STATES.find(state => state.key === value);

  const handleSelectionChange = (newValue: string) => {
    const selectedStatus = LIFECYCLE_STATES.find(state => state.key === newValue);
    if (selectedStatus) {
      onChange(selectedStatus);
    }
  };

  return (
    <FormField>
      <FormField.Label required={required}>
        {label}
      </FormField.Label>
      <FormField.Field>
        <Select 
          items={availableStates.map(state => ({ id: state.key, textValue: state.label, ...state }))}
          onSelectionChange={(keys) => {
            const selectedKey = Array.from(keys)[0] as string;
            handleSelectionChange(selectedKey);
          }}
        >
          <Select.Input 
            placeholder="选择状态..." 
            value={selectedState?.label || ''}
          />
          <Select.Popper>
            <Select.Card>
              <Select.List>
                {(state) => (
                  <Select.Item key={state.id}>
                    <Flex alignItems="center" gap="s">
                      <SystemIcon 
                        icon={state.icon} 
                        size={16} 
                        color={state.color}
                      />
                      <Box>
                        <Text 
                          typeLevel="body.medium" 
                          fontWeight="medium"
                          color={state.color}
                        >
                          {state.label}
                        </Text>
                        <Text 
                          typeLevel="subtext.small" 
                          color="hint"
                        >
                          {state.description}
                        </Text>
                      </Box>
                    </Flex>
                  </Select.Item>
                )}
              </Select.List>
            </Select.Card>
          </Select.Popper>
        </Select>
      </FormField.Field>
      {error && (
        <FormField.Hint error={true}>
          {error}
        </FormField.Hint>
      )}
      {selectedState && !error && (
        <FormField.Hint>
          {selectedState.description}
        </FormField.Hint>
      )}
    </FormField>
  );
};

/**
 * 状态显示标签组件
 */
export interface LifecycleStatusBadgeProps {
  status: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'SUSPENDED' | 'DELETED';
  showIcon?: boolean;
  size?: 'small' | 'medium' | 'large';
}

export const LifecycleStatusBadge: React.FC<LifecycleStatusBadgeProps> = ({
  status,
  showIcon = true,
  size = 'medium'
}) => {
  const state = LIFECYCLE_STATES.find(s => s.key === status);
  if (!state) return null;

  const getStyles = () => {
    switch (size) {
      case 'small':
        return {
          padding: '2px 6px',
          fontSize: '12px',
          iconSize: 12
        };
      case 'large':
        return {
          padding: '6px 12px',
          fontSize: '16px',
          iconSize: 18
        };
      default: // medium
        return {
          padding: '4px 8px',
          fontSize: '14px',
          iconSize: 16
        };
    }
  };

  const styles = getStyles();

  return (
    <Flex
      alignItems="center"
      gap="xs"
      style={{
        padding: styles.padding,
        backgroundColor: `${state.color}15`, // 15% opacity
        border: `1px solid ${state.color}40`, // 40% opacity
        borderRadius: '4px',
        display: 'inline-flex'
      }}
    >
      {showIcon && (
        <SystemIcon
          icon={state.icon}
          size={styles.iconSize}
          color={state.color}
        />
      )}
      <Text
        style={{
          fontSize: styles.fontSize,
          fontWeight: 'medium',
          color: state.color
        }}
      >
        {state.label}
      </Text>
    </Flex>
  );
};

/**
 * 状态转换提示组件
 */
export interface StateTransitionHintProps {
  currentState: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'SUSPENDED' | 'DELETED';
  targetState: 'CURRENT' | 'HISTORICAL' | 'PLANNED' | 'SUSPENDED' | 'DELETED';
}

export const StateTransitionHint: React.FC<StateTransitionHintProps> = ({
  currentState,
  targetState
}) => {
  if (currentState === targetState) return null;

  const getTransitionMessage = (): string => {
    // 定义允许的状态转换
    const transitions: Record<string, string> = {
      'PLANNED->CURRENT': '计划状态将激活为当前生效状态',
      'CURRENT->HISTORICAL': '当前状态将转为历史记录',
      'CURRENT->SUSPENDED': '当前状态将暂停，可随时恢复',
      'SUSPENDED->CURRENT': '暂停状态将恢复为当前生效状态',
      'CURRENT->DELETED': '当前状态将被软删除，不可恢复',
      'HISTORICAL->DELETED': '历史记录将被软删除',
      'SUSPENDED->DELETED': '暂停状态将被软删除',
      'PLANNED->DELETED': '计划状态将被取消删除'
    };

    const key = `${currentState}->${targetState}`;
    return transitions[key] || '状态转换可能影响组织可见性和访问权限';
  };

  const isWarning = targetState === 'DELETED';
  
  return (
    <Box
      padding="s"
      style={{
        backgroundColor: isWarning ? colors.cinnamon100 : colors.blueberry100,
        border: `1px solid ${isWarning ? colors.cinnamon300 : colors.blueberry300}`,
        borderRadius: '4px'
      }}
    >
      <Text
        typeLevel="subtext.medium"
        color={isWarning ? colors.cinnamon600 : colors.blueberry600}
      >
        {getTransitionMessage()}
      </Text>
    </Box>
  );
};

export default FiveStateStatusSelector;