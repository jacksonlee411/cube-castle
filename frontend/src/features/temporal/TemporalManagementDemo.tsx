/**
 * 时态管理集成示例组件
 * 展示如何在主应用中使用带时间轴的组织详情面板
 */
import React, { useState, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { Card } from '@workday/canvas-kit-react/card';
import { Table } from '@workday/canvas-kit-react/table';
import { Badge } from '../../shared/components/Badge';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import { OrganizationDetailPanel } from './components/OrganizationDetailPanel';
import { useTemporalHealth, type TemporalOrganizationRecord } from '../../shared/hooks/useTemporalAPI';

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

/**
 * 时态管理集成演示页面
 */
export const TemporalManagementDemo: React.FC = () => {
  // 状态管理
  const [selectedOrgCode, setSelectedOrgCode] = useState<string | null>(null);
  const [isDetailPanelOpen, setIsDetailPanelOpen] = useState(false);
  const [searchTerm, setSearchTerm] = useState('');

  // 时态服务健康检查
  const { data: healthData, isLoading: isHealthLoading } = useTemporalHealth();

  // 过滤组织列表
  const filteredOrganizations = mockOrganizations.filter(org =>
    org.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    org.code.includes(searchTerm)
  );

  // 处理查看详情
  const handleViewDetails = useCallback((orgCode: string) => {
    setSelectedOrgCode(orgCode);
    setIsDetailPanelOpen(true);
  }, []);

  // 处理关闭详情面板
  const handleCloseDetailPanel = useCallback(() => {
    setIsDetailPanelOpen(false);
    setSelectedOrgCode(null);
  }, []);

  // 处理保存组织记录
  const handleSaveOrganization = useCallback(async (record: TemporalOrganizationRecord) => {
    console.log('保存组织记录:', record);
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // 这里应该调用实际的API来保存记录
    // 例如: await organizationAPI.update(record);
    
    alert(`组织 "${record.name}" 保存成功！`);
  }, []);

  // 处理删除组织
  const handleDeleteOrganization = useCallback(async (orgCode: string) => {
    console.log('删除组织:', orgCode);
    
    // 模拟API调用
    await new Promise(resolve => setTimeout(resolve, 500));
    
    // 这里应该调用实际的API来删除记录
    // 例如: await organizationAPI.delete(orgCode);
    
    alert(`组织 "${orgCode}" 删除成功！`);
  }, []);

  // 获取状态徽章样式
  const getStatusBadgeVariant = (status: string) => {
    switch (status) {
      case 'ACTIVE': return 'positive';
      case 'PLANNED': return 'caution';
      case 'INACTIVE': return 'neutral';
      default: return 'neutral';
    }
  };

  // 获取状态标签
  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'ACTIVE': return '启用';
      case 'PLANNED': return '计划中';
      case 'INACTIVE': return '停用';
      default: return status;
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
            时态管理集成演示
          </Text>
          
          <Flex alignItems="center" gap={space.s}>
            {isHealthLoading ? (
              <Text fontSize="small">检查时态服务...</Text>
            ) : healthData ? (
              <Badge variant={healthData.status === 'healthy' ? 'positive' : 'negative'}>
                时态服务: {healthData.status === 'healthy' ? '正常' : '异常'}
              </Badge>
            ) : (
              <Badge variant="caution">
                时态服务: 未连接
              </Badge>
            )}
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
          <Table>
            <thead>
              <tr>
                <th>组织代码</th>
                <th>组织名称</th>
                <th>类型</th>
                <th>状态</th>
                <th>层级</th>
                <th>生效日期</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              {filteredOrganizations.map((org) => (
                <tr key={org.code}>
                  <td>
                    <Text fontSize="small" fontFamily="monospace">
                      {org.code}
                    </Text>
                  </td>
                  <td>
                    <Text fontWeight="medium">
                      {org.name}
                    </Text>
                  </td>
                  <td>
                    <Text fontSize="small">
                      {getTypeLabel(org.unit_type)}
                    </Text>
                  </td>
                  <td>
                    <Badge variant={getStatusBadgeVariant(org.status)} size="small">
                      {getStatusLabel(org.status)}
                    </Badge>
                  </td>
                  <td>
                    <Text fontSize="small">
                      L{org.level}
                    </Text>
                  </td>
                  <td>
                    <Text fontSize="small">
                      {new Date(org.effective_date).toLocaleDateString('zh-CN')}
                    </Text>
                  </td>
                  <td>
                    <Flex gap={space.xs}>
                      <SecondaryButton
                        size="small"
                        onClick={() => handleViewDetails(org.code)}
                      >
                        查看详情
                      </SecondaryButton>
                    </Flex>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        )}
      </Card>

      {/* 时态管理详情面板 */}
      {isDetailPanelOpen && selectedOrgCode && (
        <OrganizationDetailPanel
          organizationCode={selectedOrgCode}
          isOpen={isDetailPanelOpen}
          onClose={handleCloseDetailPanel}
          onSave={handleSaveOrganization}
          onDelete={handleDeleteOrganization}
        />
      )}

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
              <strong>实时数据加载</strong>: 连接到端口9091的时态管理服务，获取真实的时态数据
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
      </Card>
    </Box>
  );
};

export default TemporalManagementDemo;