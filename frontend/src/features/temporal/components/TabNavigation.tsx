import React from 'react';
import { Box, Flex, Text } from '@workday/canvas-kit-react';
import { colors, space } from '@workday/canvas-kit-react/tokens';

// 选项卡类型
export type TabType = 'edit-history' | 'new-version' | 'audit-history';

// 选项卡配置
interface TabConfig {
  key: TabType;
  label: string;
  disabled?: boolean;
}

interface TabNavigationProps {
  activeTab: TabType;
  onTabChange: (tab: TabType) => void;
  tabs: TabConfig[];
  disabled?: boolean;
  className?: string;
}

/**
 * 选项卡导航组件
 * 用于在组织详情视图中切换不同功能区域
 */
export const TabNavigation: React.FC<TabNavigationProps> = ({
  activeTab,
  onTabChange,
  tabs,
  disabled = false,
  className
}) => {
  const handleTabClick = (tabKey: TabType) => {
    if (!disabled && !tabs.find(t => t.key === tabKey)?.disabled) {
      onTabChange(tabKey);
    }
  };

  return (
    <Box className={className}>
      <Flex
        borderBottom={`2px solid ${colors.soap300}`}
        marginBottom={space.l}
      >
        {tabs.map((tab) => {
          const isActive = activeTab === tab.key;
          const isDisabled = disabled || tab.disabled;
          
          return (
            <Box
              key={tab.key}
              onClick={() => handleTabClick(tab.key)}
              style={{
                cursor: isDisabled ? 'not-allowed' : 'pointer',
                borderBottom: isActive ? `3px solid ${colors.blueberry600}` : '3px solid transparent',
                paddingBottom: space.s,
                paddingTop: space.s,
                paddingLeft: space.l,
                paddingRight: space.l,
                marginBottom: '-2px',
                transition: 'all 0.2s ease-in-out'
              }}
              onMouseEnter={(e) => {
                if (!isDisabled && !isActive) {
                  e.currentTarget.style.backgroundColor = colors.soap100;
                }
              }}
              onMouseLeave={(e) => {
                if (!isActive) {
                  e.currentTarget.style.backgroundColor = 'transparent';
                }
              }}
            >
              <Text
                typeLevel="body.medium"
                fontWeight={isActive ? 'medium' : 'regular'}
                color={
                  isDisabled 
                    ? colors.licorice300
                    : isActive 
                    ? colors.blueberry600 
                    : colors.licorice600
                }
              >
                {tab.label}
              </Text>
            </Box>
          );
        })}
      </Flex>
    </Box>
  );
};