/**
 * 组织架构树状图组件
 * 基于Canvas Kit v13企业级设计系统
 * 集成真实GraphQL层级查询API
 */
import { logger } from '@/shared/utils/logger';
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
import { StatusBadge } from '../../../shared/components/StatusBadge';
import type { OrganizationStatus } from '@/shared/types';
import { OrganizationStatusEnum } from '@/shared/types/contract_gen';
import { unifiedGraphQLClient } from '../../../shared/api/unified-client';
import { getOrganizationByCode } from '@/shared/api/facade/organization';
import { OrganizationBreadcrumb } from '../../../shared/components/OrganizationBreadcrumb';
import { useNavigate } from 'react-router-dom';
// SecondaryButton 已在上方导入
import { toParentChainFromCodePath } from '../../../shared/utils/organizationPath';
import { coerceOrganizationLevel, getDisplayLevel } from '../../../shared/utils/organization-helpers';

// 层级节点数据接口
export interface OrganizationTreeNode {
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  parentCode?: string | null;
  parentChain: string[];
  codePath?: string;
  namePath?: string;
  childrenCount: number;
  children: OrganizationTreeNode[];
  // 展示增强字段
  isExpanded?: boolean;
  isSelected?: boolean;
}

interface OrganizationSubtreeNode {
  code: string;
  name: string;
  unitType: string;
  status: string;
  level: number;
  parentCode?: string | null;
  codePath?: string | null;
  namePath?: string | null;
  parentChain?: string[] | null;
  childrenCount?: number | null;
  hierarchyDepth?: number | null;
  children?: OrganizationSubtreeNode[] | null;
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
    case OrganizationStatusEnum.Inactive:
      return OrganizationStatusEnum.Inactive;
    case OrganizationStatusEnum.Planned:
      return OrganizationStatusEnum.Planned;
    case OrganizationStatusEnum.Deleted:
      return OrganizationStatusEnum.Deleted;
    case OrganizationStatusEnum.Active:
    default:
      return OrganizationStatusEnum.Active;
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
                  {node.code} • 第{getDisplayLevel(node.level)}级 • {node.unitType}
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
  const navigate = useNavigate();
  // 状态管理
  const [treeData, setTreeData] = useState<OrganizationTreeNode[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<OrganizationTreeNode | null>(null);

  const mapOrganizationNode = useCallback(
    (org: OrganizationSubtreeNode, includeChildren: boolean): OrganizationTreeNode => ({
      code: org.code,
      name: org.name,
      unitType: org.unitType,
      status: org.status,
      level: coerceOrganizationLevel(org.level, org.hierarchyDepth ?? undefined),
      parentCode: org.parentCode ?? undefined,
      parentChain: Array.isArray(org.parentChain)
        ? org.parentChain
        : toParentChainFromCodePath(org.codePath ?? null),
      codePath: org.codePath ?? undefined,
      namePath: org.namePath ?? undefined,
      childrenCount: org.childrenCount ?? (org.children?.length ?? 0),
      children: includeChildren
        ? (org.children ?? []).map(child => mapOrganizationNode(child, includeChildren))
        : [],
      isExpanded: false,
      isSelected: false,
    }),
    []
  );

  // 拉取指定节点的一层子节点
  const fetchChildren = useCallback(async (code: string): Promise<OrganizationTreeNode[]> => {
    const graphqlQuery = `
      query TemporalEntityTreeChildren($code: String!, $maxDepth: Int) {
        organizationSubtree(code: $code, maxDepth: $maxDepth) {
          children {
            code
            name
            unitType
            status
            level
            parentCode
            codePath
            namePath
            parentChain
            childrenCount
          }
        }
      }
    `;
    const variables = { code, maxDepth };
    const data = await unifiedGraphQLClient.request<{ organizationSubtree?: { children?: OrganizationSubtreeNode[] } }>(
      graphqlQuery,
      variables
    );
    const children = data?.organizationSubtree?.children ?? [];
    return children.map((org) => mapOrganizationNode(org, false));
  }, [mapOrganizationNode, maxDepth]);
  
  // 加载树形数据
  const loadTreeData = useCallback(async (code?: string) => {
    try {
      setIsLoading(true);
      setError(null);

      if (code) {
        const graphqlQuery = `
          query TemporalEntitySubtree($code: String!, $maxDepth: Int) {
            organizationSubtree(code: $code, maxDepth: $maxDepth) {
              code
              name
              unitType
              status
              level
              parentCode
              codePath
              namePath
              parentChain
              childrenCount
              hierarchyDepth
              children {
                code
                name
                unitType
                status
                level
                parentCode
                codePath
                namePath
                parentChain
                childrenCount
                hierarchyDepth
                children {
                  code
                  name
                  unitType
                  status
                  level
                  parentCode
                  codePath
                  namePath
                  parentChain
                  childrenCount
                  hierarchyDepth
                }
              }
            }
          }
        `;

        const data = await unifiedGraphQLClient.request<{ organizationSubtree?: OrganizationSubtreeNode }>(
          graphqlQuery,
          { code, maxDepth }
        );

        const subtree = data.organizationSubtree;
        if (subtree) {
          const mapped = mapOrganizationNode(subtree, true);
          setTreeData(showRoot ? [mapped] : mapped.children);
        } else {
          // Facade 回退：当子树为空时，尝试获取当前组织快照以至少渲染根节点
          try {
            const root = code ? await getOrganizationByCode(code) : null;
            if (root) {
              const mappedRoot: OrganizationTreeNode = {
                code: root.code,
                name: root.name,
                unitType: String(root.unitType),
                status: String(root.status),
                level: root.level ?? coerceOrganizationLevel(undefined),
                parentCode: root.parentCode ?? null ?? undefined,
                codePath: root.codePath ?? undefined ?? undefined,
                namePath: root.namePath ?? undefined ?? undefined,
                parentChain: toParentChainFromCodePath(root.codePath ?? undefined) ?? [],
                childrenCount: 0,
                children: [],
                isExpanded: false,
              };
              setTreeData([mappedRoot]);
            } else {
              setTreeData([]);
            }
          } catch (_fallback) {
            setTreeData([]);
          }
        }
      } else {
        const graphqlQuery = `
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
                hierarchyDepth
              }
            }
          }
        `;

        const data = await unifiedGraphQLClient.request<{ organizations?: { data: OrganizationSubtreeNode[] } }>(
          graphqlQuery,
          {
            filter: {
              parentCode: null
            }
          }
        );

        let treeNodes: OrganizationTreeNode[] = (data.organizations?.data ?? []).map((org) => mapOrganizationNode(org, false));

        // 为根节点并发查询 childrenCount（使用 organizationSubtree maxDepth: 1）
        try {
          const limit = 6; // 限制并发
          const queue = [...treeNodes];
          const updateMap: Record<string, number> = {};

          const runBatch = async (batch: OrganizationTreeNode[]) => {
            await Promise.all(batch.map(async (n) => {
              try {
                const resp = await unifiedGraphQLClient.request<{ organizationSubtree?: { childrenCount: number } }>(
                  `query GetRootChildrenCount($code: String!, $maxDepth: Int) {
                    organizationSubtree(code: $code, maxDepth: $maxDepth) {
                      childrenCount
                    }
                  }`,
                  { code: n.code, maxDepth }
                );
                updateMap[n.code] = resp?.organizationSubtree?.childrenCount ?? 0;
              } catch (_e) {
                // 静默失败，保持0
              }
            }));
          };

          while (queue.length > 0) {
            const batch = queue.splice(0, limit);
            await runBatch(batch);
          }

          treeNodes = treeNodes.map(n => ({ ...n, childrenCount: updateMap[n.code] ?? n.childrenCount }));
        } catch (_e) {
          // 静默：无法获取childrenCount不影响展示
        }

        setTreeData(treeNodes);
      }
    } catch (error) {
      logger.error('Error loading tree data:', error);
      const errorMessage = error instanceof Error
        ? error.message
        : '加载组织架构数据失败';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  }, [mapOrganizationNode, maxDepth, showRoot]);
  
  // 处理节点选择
  const handleNodeSelect = useCallback((node: OrganizationTreeNode) => {
    setSelectedNode(node);
    if (onNodeSelect) {
      onNodeSelect(node);
    } else {
      navigate(`/organizations/${node.code}/temporal`);
    }
  }, [onNodeSelect, navigate]);
  
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
    
    // 如果节点还没有加载子节点，则先从服务拉取一层子节点
    if (!node.isExpanded && node.children.length === 0 && node.childrenCount > 0) {
      try {
        setIsLoading(true);
        const children = await fetchChildren(node.code);
        setTreeData(prev => {
          const attachChildren = (nodes: OrganizationTreeNode[]): OrganizationTreeNode[] =>
            nodes.map(n => {
              if (n.code === node.code) {
                return { ...n, children };
              }
              if (n.children.length > 0) {
                return { ...n, children: attachChildren(n.children) };
              }
              return n;
            });
          return attachChildren(Array.isArray(prev) ? prev : []);
        });
      } catch (e) {
        logger.error('加载子节点失败:', e);
        setError(e instanceof Error ? e.message : '加载子节点失败');
      } finally {
        setIsLoading(false);
      }
    }
    
    setTreeData(updateNodeExpansion);
    onNodeExpand?.(node);
  }, [onNodeExpand, fetchChildren]);
  
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
            <Text typeLevel="subtext.small" fontWeight="medium" marginBottom="xs">
              已选中：{selectedNode.name} ({selectedNode.code})
            </Text>
            <OrganizationBreadcrumb
              codePath={selectedNode.codePath}
              namePath={selectedNode.namePath}
              onNavigate={(code) => {
                if (code) navigate(`/organizations/${code}/temporal`);
              }}
            />
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
              data-testid={temporalEntitySelectors.organization.treeRetryButton}
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
