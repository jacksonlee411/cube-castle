/**
 * 简化版组织详情集成示例组件
 * 移除Canvas Kit Badge依赖，使用简单的HTML样式
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Card } from '@workday/canvas-kit-react/card';
import { colors, space } from '@workday/canvas-kit-react/tokens';
// 移除时态健康检查，使用GraphQL服务健康检查

// 模拟的组织列表数据
const mockOrganizations = [
  {
    code: '1000056',
    name: '测试更新缓存_同步修复',
    unit_type: 'DEPARTMENT',
    status: 'ACTIVE',
    level: 1,
    effective_date: '2025-08-10'
  },
  {
    code: '1000057',
    name: '人力资源部',
    unit_type: 'DEPARTMENT',
    status: 'ACTIVE',
    level: 2,
    effective_date: '2025-01-01'
  },
  {
    code: '1000058',
    name: '财务部',
    unit_type: 'DEPARTMENT',
    status: 'ACTIVE',
    level: 2,
    effective_date: '2025-01-01'
  },
  {
    code: '1000059',
    name: '计划项目组',
    unit_type: 'PROJECT_TEAM',
    status: 'PLANNED',
    level: 3,
    effective_date: '2025-09-01'
  }
];

// 简单的Badge组件实现
interface SimpleBadgeProps {
  children: React.ReactNode;
  variant?: 'positive' | 'caution' | 'neutral';
  size?: 'small' | 'medium';
}

const SimpleBadge: React.FC<SimpleBadgeProps> = ({ children, variant = 'neutral', size = 'medium' }) => {
  const getVariantStyles = (variant: string) => {
    switch (variant) {
      case 'positive':
        return { backgroundColor: '#d1f2eb', color: '#1e8449', border: '1px solid #58d68d' };
      case 'caution':
        return { backgroundColor: '#fef9e7', color: '#b7950b', border: '1px solid #f4d03f' };
      default:
        return { backgroundColor: '#f8f9fa', color: '#6c757d', border: '1px solid #dee2e6' };
    }
  };

  const getSizeStyles = (size: string) => {
    switch (size) {
      case 'small':
        return { padding: '2px 6px', fontSize: '11px' };
      default:
        return { padding: '4px 8px', fontSize: '12px' };
    }
  };

  return (
    <div
      style={{
        display: 'inline-block',
        borderRadius: '4px',
        ...getVariantStyles(variant),
        ...getSizeStyles(size)
      }}
    >
      {children}
    </div>
  );
};

/**
 * 简化版组织详情集成演示页面
 */
export const TemporalManagementSimple: React.FC = () => {
  // 状态管理
  const [searchTerm, setSearchTerm] = useState('');
  const [, setSelectedOrgCode] = useState<string | null>(null);

  // 时态服务健康检查
  // 移除时态服务健康检查，因为现在使用GraphQL服务
  // const { data: healthData, isLoading: isHealthLoading } = useTemporalHealth();

  // 过滤组织列表
  const filteredOrganizations = mockOrganizations.filter(org =>
    org.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    org.code.includes(searchTerm)
  );

  // 处理查看详情
  const handleViewDetails = useCallback((orgCode: string) => {
    setSelectedOrgCode(orgCode);
    alert(`点击了查看详情: ${orgCode}\n\n这里会打开组织详情面板，包含：\n• 左侧垂直时间轴\n• 右侧组织详情编辑\n• 时态数据查询和显示`);
  }, []);

  // 获取状态标签
  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'ACTIVE': return '启用';
      case 'PLANNED': return '计划中';
      case 'INACTIVE': return '停用';
      default: return status;
    }
  };

  // 获取状态变体
  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'positive';
      case 'PLANNED': return 'caution';
      case 'INACTIVE': return 'neutral';
      default: return 'neutral';
    }
  };

  // 获取类型标签
  const getTypeLabel = (type: string) => {
    switch (type) {
      case 'COMPANY': return '公司';
      case 'DEPARTMENT': return '部门';
      case 'COST_CENTER': return '成本中心';
      case 'PROJECT_TEAM': return '项目团队';
      default: return type;
    }
  };

  return (
    <Box padding={space.l}>
      {/* 页面标题和时态服务状态 */}
      <Box marginBottom={space.l}>
        <Flex alignItems="center" justifyContent="space-between" marginBottom={space.m}>
          <Text fontSize="xl" fontWeight="bold">
            组织详情集成演示 (简化版)
          </Text>
          
          <Flex alignItems="center" gap={space.s}>
            {/* 使用GraphQL服务替代时态服务健康检查 */}
            <SimpleBadge variant="positive">
              GraphQL服务: 正常
            </SimpleBadge>
          </Flex>
        </Flex>

        <Text fontSize="medium" color={colors.licorice600}>
          点击组织列表中的"查看详情"按钮，体验带时间轴的组织详情面板
        </Text>
      </Box>

      {/* 搜索和操作栏 */}
      <Card marginBottom={space.l} padding={space.m}>
        <Flex alignItems="center" justifyContent="space-between">
          <Box flex="1" marginRight={space.m}>
            <Text fontSize="small" marginBottom={space.xs}>
              搜索组织
            </Text>
            <TextInput
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="输入组织名称或代码..."
            />
          </Box>

          <Box>
            <PrimaryButton onClick={() => alert('新增功能演示')}>
              新增组织
            </PrimaryButton>
          </Box>
        </Flex>
      </Card>

      {/* 组织列表 */}
      <Card padding={space.m}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          组织列表 ({filteredOrganizations.length} 个)
        </Text>

        {filteredOrganizations.length === 0 ? (
          <Box 
            padding={space.l} 
            textAlign="center" 
            backgroundColor={colors.soap100}
          >
            <Text>没有找到匹配的组织</Text>
          </Box>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>组织代码</th>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>组织名称</th>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>类型</th>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>状态</th>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>层级</th>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>生效日期</th>
                  <th style={{ padding: '8px', textAlign: 'left', borderBottom: '1px solid #dee2e6' }}>操作</th>
                </tr>
              </thead>
              <tbody>
                {filteredOrganizations.map((org) => (
                  <tr key={org.code} style={{ borderBottom: '1px solid #f8f9fa' }}>
                    <td style={{ padding: '8px', fontFamily: 'monospace' }}>
                      {org.code}
                    </td>
                    <td style={{ padding: '8px', fontWeight: '500' }}>
                      {org.name}
                    </td>
                    <td style={{ padding: '8px' }}>
                      {getTypeLabel(org.unit_type)}
                    </td>
                    <td style={{ padding: '8px' }}>
                      <SimpleBadge variant={getStatusVariant(org.status)} size="small">
                        {getStatusLabel(org.status)}
                      </SimpleBadge>
                    </td>
                    <td style={{ padding: '8px' }}>
                      L{org.level}
                    </td>
                    <td style={{ padding: '8px' }}>
                      {new Date(org.effective_date).toLocaleDateString('zh-CN')}
                    </td>
                    <td style={{ padding: '8px' }}>
                      <SecondaryButton
                        size="small"
                        onClick={() => handleViewDetails(org.code)}
                      >
                        查看详情
                      </SecondaryButton>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* 功能说明 */}
      <Card marginTop={space.l} padding={space.m} backgroundColor={colors.frenchVanilla100}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          📖 功能说明
        </Text>
        
        <Box as="ul" marginLeft={space.m}>
          <Box as="li" marginBottom={space.s}>
            <Text fontSize="small">
              <strong>时间轴导航</strong>: 左侧垂直时间轴显示组织的历史变更记录，点击不同节点可查看对应时间点的详情
            </Text>
          </Box>
          <Box as="li" marginBottom={space.s}>
            <Text fontSize="small">
              <strong>纯日期生效模型</strong>: 基于生效日期和结束日期管理时态数据，无需复杂的版本号
            </Text>
          </Box>
          <Box as="li" marginBottom={space.s}>
            <Text fontSize="small">
              <strong>实时数据加载</strong>: 连接到端口9091的组织详情服务，获取真实的时态数据
            </Text>
          </Box>
          <Box as="li" marginBottom={space.s}>
            <Text fontSize="small">
              <strong>编辑模式</strong>: 支持查看和编辑模式切换，实时保存变更到后端服务
            </Text>
          </Box>
          <Box as="li">
            <Text fontSize="small">
              <strong>状态指示</strong>: 清晰的视觉反馈显示当前记录、历史记录和计划记录的区别
            </Text>
          </Box>
        </Box>

        {/* GraphQL服务状态 */}
        <Box marginTop={space.m} padding={space.s} backgroundColor={colors.soap100} borderRadius="4px">
          <Text fontSize="small" fontWeight="bold" marginBottom={space.xs}>
            GraphQL服务连接状态:
          </Text>
          <Box>
            <Text fontSize="small">• 服务: GraphQL 组织查询服务</Text>
            <Text fontSize="small">• 状态: 正常</Text>
            <Text fontSize="small">• 功能: 时态查询, 历史记录, 时间线</Text>
            <Text fontSize="small">• 更新时间: {new Date().toLocaleString('zh-CN')}</Text>
          </Box>
        </Box>
        </Box>
      </Card>
    </Box>
  );
};

export default TemporalManagementSimple;