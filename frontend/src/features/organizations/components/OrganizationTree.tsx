/**
 * 组织架构树状图组件
 * 基于Canvas Kit v13企业级设计系统
 * 集成真实GraphQL层级查询API
 */
import React, { useState, useCallback, useEffect } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text, Heading } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots';
import { SecondaryButton } from '@workday/canvas-kit-react/button';
import { 
  colors, 
  borderRadius 
} from '@workday/canvas-kit-react/tokens';
import { StatusBadge, type OrganizationStatus } from '../../../shared/components/StatusBadge';
import { unifiedGraphQLClient } from '../../../shared/api/unified-client';

// 层级节点数据接口
export interface OrganizationTreeNode {
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  parentCode?: string | null;
  parentChain: string[];
  childrenCount: number;
  children: OrganizationTreeNode[];
  // 展示增强字段
  isExpanded?: boolean;
  isSelected?: boolean;
}

// 组件属性接口
export interface OrganizationTreeProps {
  /** 根节点代码 */
  rootCode?: string;
  /** 最大展示深度 */
  maxDepth?: number;
  /** 只读模式 */
  readonly?: boolean;
  /** 节点选择回调 */
  onNodeSelect?: (node: OrganizationTreeNode) => void;
  /** 节点展开回调 */
  onNodeExpand?: (node: OrganizationTreeNode) => void;
  /** 自定义宽度 */
  width?: string;
  /** 自定义高度 */
  height?: string;
  /** 显示根节点 */
  showRoot?: boolean;
}

// 状态映射函数 - 符合API契约的3个业务状态
const mapStatusToOrganizationStatus = (status: string): OrganizationStatus => {
  switch (status) {
    case 'ACTIVE': return 'ACTIVE';
    case 'INACTIVE': return 'INACTIVE';
    case 'PLANNED': return 'PLANNED';
    default: return 'ACTIVE';
  }
};

/**
 * 树节点组件
 */
interface TreeNodeProps {
  node: OrganizationTreeNode;
  level: number;
  onSelect?: (node: OrganizationTreeNode) => void;
  onToggle?: (node: OrganizationTreeNode) => void;
  isSelected?: boolean;
  readonly?: boolean;
}

const TreeNode: React.FC<TreeNodeProps> = ({
  node,
  level,
  onSelect,
  onToggle,
  isSelected = false,
  readonly = false
}) => {
  const hasChildren = node.childrenCount > 0;
  const isExpanded = node.isExpanded || false;
  
  // 计算缩进
  const indentSize = level * 24;
  
  return (
    <Box>
      {/* 节点内容 */}
      <Card
        padding="s"
        marginBottom="xs"
        style={{
          marginLeft: `${indentSize}px`,
          backgroundColor: isSelected ? '#E3F2FD' : 'white',
          border: isSelected ? '2px solid #2196F3' : '1px solid #E9ECEF',
          cursor: 'pointer',
          transition: 'all 0.2s ease',
          boxShadow: isSelected 
            ? '0 4px 12px rgba(33, 150, 243, 0.2)' 
            : '0 1px 3px rgba(0,0,0,0.1)',
        }}
        onClick={() => onSelect?.(node)}
        onMouseEnter={(e) => {
          if (!isSelected) {
            e.currentTarget.style.boxShadow = '0 2px 8px rgba(0,0,0,0.15)';
          }
        }}
        onMouseLeave={(e) => {
          if (!isSelected) {
            e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.1)';
          }
        }}
      >
        <Flex alignItems="center" gap="s">
          {/* 展开/折叠按钮 */}
          {hasChildren && (
            <SecondaryButton
              size="small"
              onClick={(e) => {
                e.stopPropagation();
                onToggle?.(node);
              }}
              style={{
                minWidth: '24px',
                width: '24px',
                height: '24px',
                padding: 0
              }}
            >
              <Text fontSize="small">
                {isExpanded ? '−' : '+'}
              </Text>
            </SecondaryButton>
          )}
          
          {/* 节点信息 */}
          <Box flex="1">
            <Flex alignItems="center" justifyContent="space-between" marginBottom="xs">
              {/* 组织名称和层级信息 */}
              <Box>
                <Text typeLevel="body.medium" fontWeight="medium">
                  {node.name}
                </Text>
                <Text typeLevel="subtext.small" color="hint">
                  {node.code} • 第{node.level}级 • {node.unitType}
                </Text>
              </Box>
              
              {/* 状态标识 */}
              <StatusBadge 
                status={mapStatusToOrganizationStatus(node.status)} 
                size="small"
              />
            </Flex>
            
            {/* 子节点统计 */}
            {hasChildren && (
              <Text typeLevel="subtext.small" color="hint">
                {node.childrenCount}个下级单位
              </Text>
            )}
          </Box>
        </Flex>
      </Card>
      
      {/* 子节点 */}
      {isExpanded && node.children.map((child) => (
        <TreeNode
          key={child.code}
          node={child}
          level={level + 1}
          onSelect={onSelect}
          onToggle={onToggle}
          isSelected={child.isSelected}
          readonly={readonly}
        />
      ))}
    </Box>
  );
};

/**
 * 组织架构树状图主组件
 */
export const OrganizationTree: React.FC<OrganizationTreeProps> = ({
  rootCode,
  maxDepth = 5,
  readonly = false,
  onNodeSelect,
  onNodeExpand,
  width = "100%",
  height = "calc(100vh - 200px)",
  showRoot = true
}) => {
  // 状态管理
  const [treeData, setTreeData] = useState<OrganizationTreeNode[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<OrganizationTreeNode | null>(null);
  
  // 加载树形数据
  const loadTreeData = useCallback(async (code?: string) => {
    try {
      setIsLoading(true);
      setError(null);
      
      // 如果没有指定根节点，则查询顶级节点
      let graphqlQuery: string;
      let variables: Record<string, unknown>;
      
      if (code) {
        // 查询指定节点的子树
        graphqlQuery = `
          query GetOrganizationSubtree($code: String!, $maxDepth: Int) {
            organizationSubtree(code: $code, maxDepth: $maxDepth) {
              code
              name
              unitType
              status
              level
              parentCode
              parentChain
              childrenCount
              children {
                code
                name
                unitType
                status
                level
                parentCode
                parentChain
                childrenCount
                children {
                  code
                  name
                  unitType
                  status
                  level
                  parentCode
                  parentChain
                  childrenCount
                }
              }
            }
          }
        `;
        variables = { code, maxDepth };
      } else {
        // 查询根级节点
        graphqlQuery = `
          query GetRootOrganizations($filter: OrganizationFilter) {
            organizations(filter: $filter) {
              data {
                code
                name
                unitType
                status
                level
                parentCode
                codePath
                namePath
              }
            }
          }
        `;
        variables = {
          filter: {
            parentCode: null
          }
        };
      }
      
      const data = await unifiedGraphQLClient.request(graphqlQuery, variables);
        
        let treeNodes: OrganizationTreeNode[];
        
        if (code) {
          // 处理子树响应
          const subtree = (data as { organizationSubtree?: OrganizationTreeNode }).organizationSubtree;
          if (subtree) {
            treeNodes = showRoot ? [subtree] : subtree.children || [];
          } else {
            treeNodes = [];
          }
        } else {
          // 处理根节点响应
          const organizations = (data as { organizations?: { data: Record<string, unknown>[] } }).organizations?.data || [];
          treeNodes = organizations.map((org: Record<string, unknown>) => ({
            code: org.code as string,
            name: org.name as string,
            unitType: org.unitType as string,
            status: org.status as string,
            level: (org.level as number) || 1,
            parentCode: org.parentCode as string | undefined,
            parentChain: org.codePath ? (org.codePath as string).split('/').filter(Boolean) : [],
            childrenCount: 0, // 需要后续查询获取
            children: [],
            isExpanded: false
          }));
        }
        
        setTreeData(treeNodes);
    } catch (error) {
      console.error('Error loading tree data:', error);
      const errorMessage = error instanceof Error 
        ? error.message 
        : '加载组织架构数据失败';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  }, [maxDepth, showRoot]);
  
  // 处理节点选择
  const handleNodeSelect = useCallback((node: OrganizationTreeNode) => {
    setSelectedNode(node);
    onNodeSelect?.(node);
  }, [onNodeSelect]);
  
  // 处理节点展开/折叠
  const handleNodeToggle = useCallback(async (node: OrganizationTreeNode) => {
    const updateNodeExpansion = (nodes: OrganizationTreeNode[]): OrganizationTreeNode[] => {
      return nodes.map(n => {
        if (n.code === node.code) {
          return { ...n, isExpanded: !n.isExpanded };
        }
        if (n.children.length > 0) {
          return { ...n, children: updateNodeExpansion(n.children) };
        }
        return n;
      });
    };
    
    // 如果节点还没有加载子节点，则先加载
    if (!node.isExpanded && node.children.length === 0 && node.childrenCount > 0) {
      try {
        setIsLoading(true);
        // 这里应该加载子节点数据
        // 暂时使用模拟数据
        await new Promise(resolve => setTimeout(resolve, 500));
      } finally {
        setIsLoading(false);
      }
    }
    
    setTreeData(updateNodeExpansion);
    onNodeExpand?.(node);
  }, [onNodeExpand]);
  
  // 初始化加载
  useEffect(() => {
    loadTreeData(rootCode);
  }, [loadTreeData, rootCode]);
  
  return (
    <Box
      width={width}
      height={height}
      backgroundColor="#F8F9FA"
      borderRadius={borderRadius.m}
      border="1px solid #E9ECEF"
      padding="m"
      overflowY="auto"
    >
      {/* 头部区域 */}
      <Box marginBottom="m">
        <Flex justifyContent="space-between" alignItems="center" marginBottom="s">
          <Heading size="medium">组织架构图</Heading>
          <SecondaryButton
            size="small"
            onClick={() => loadTreeData(rootCode)}
            disabled={isLoading}
          >
            {isLoading ? '刷新中...' : '刷新'}
          </SecondaryButton>
        </Flex>
        
        <Text typeLevel="subtext.small" color="hint">
          点击节点查看详情，点击 + / − 展开或折叠子节点
        </Text>
        
        {selectedNode && (
          <Box 
            marginTop="s"
            padding="s"
            backgroundColor={colors.blueberry100}
            borderRadius={borderRadius.s}
          >
            <Text typeLevel="subtext.small" fontWeight="medium">
              已选中：{selectedNode.name} ({selectedNode.code})
            </Text>
          </Box>
        )}
      </Box>
      
      {/* 错误提示 */}
      {error && (
        <Box
          marginBottom="m"
          padding="m"
          backgroundColor={colors.cinnamon100}
          border={`1px solid ${colors.cinnamon600}`}
          borderRadius={borderRadius.m}
        >
          <Flex alignItems="center" gap="s">
            <Text color={colors.cinnamon600}>⚠️</Text>
            <Box flex="1">
              <Text color={colors.cinnamon600} typeLevel="body.small" fontWeight="medium">
                加载失败
              </Text>
              <Text color={colors.cinnamon600} typeLevel="subtext.small">
                {error}
              </Text>
            </Box>
            <SecondaryButton
              size="small"
              onClick={() => loadTreeData(rootCode)}
              disabled={isLoading}
            >
              重试
            </SecondaryButton>
          </Flex>
        </Box>
      )}
      
      {/* 树状图内容 */}
      {isLoading ? (
        <Box textAlign="center" padding="l">
          <LoadingDots />
          <Text marginTop="s" typeLevel="subtext.small">加载组织架构中...</Text>
        </Box>
      ) : (
        <Box>
          {treeData.length > 0 ? (
            <Box>
              {treeData.map((node) => (
                <TreeNode
                  key={node.code}
                  node={node}
                  level={0}
                  onSelect={handleNodeSelect}
                  onToggle={handleNodeToggle}
                  isSelected={selectedNode?.code === node.code}
                  readonly={readonly}
                />
              ))}
            </Box>
          ) : (
            <Box textAlign="center" padding="l">
              <Text color="hint">暂无组织架构数据</Text>
              {!rootCode && (
                <Text typeLevel="subtext.small" color="hint" marginTop="s">
                  系统中没有根级组织单位
                </Text>
              )}
            </Box>
          )}
        </Box>
      )}
    </Box>
  );
};

export default OrganizationTree;