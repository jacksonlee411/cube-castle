/**
 * 集成时态功能的组织详情面板 (纯日期生效模型)
 * 左侧时间轴 + 右侧详情编辑的完整部门详情面板
 */
import React, { useState, useEffect, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton, TertiaryButton } from '@workday/canvas-kit-react/button';
import { Badge } from '../../../shared/components/Badge';
import { Tooltip } from '@workday/canvas-kit-react/tooltip';
import { colors, space, borderRadius } from '@workday/canvas-kit-react/tokens';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import { xIcon, plusIcon } from '@workday/canvas-system-icons-web';
import { 
  useOrganizationHistory,
  useOrganizationAsOfDate,
  type TemporalOrganizationUnit
} from '../../../shared/hooks/useTemporalGraphQL';
import { TemporalConverter } from '../../../shared/utils/temporal-converter';
import { OrganizationDetailForm } from './OrganizationDetailForm';

export interface OrganizationDetailPanelProps {
  /** 组织代码 */
  organizationCode: string;
  /** 是否显示面板 */
  isOpen: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 保存回调 */
  onSave: (record: TemporalOrganizationUnit) => Promise<void>;
  /** 删除回调 */
  onDelete?: (organizationCode: string) => Promise<void>;
}

// 时间轴节点类型
interface TimelineNode {
  date: string;
  displayDate: string;
  type: 'current' | 'historical' | 'planned';
  label: string;
  record?: TemporalOrganizationUnit;
  changeType?: 'created' | 'modified' | 'planned' | 'activated' | 'deactivated';
}

// 视图模式
type ViewMode = 'view' | 'edit' | 'timeline';

/**
 * 组织详情面板 - 集成时态查询功能
 */
export const OrganizationDetailPanel: React.FC<OrganizationDetailPanelProps> = ({
  organizationCode,
  isOpen,
  onClose,
  onSave,
  onDelete
}) => {
  // 状态管理
  const [selectedDate, setSelectedDate] = useState<string>(TemporalConverter.getCurrentDateString());
  const [selectedRecord, setSelectedRecord] = useState<TemporalOrganizationUnit | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>('view');
  const [editingRecord, setEditingRecord] = useState<TemporalOrganizationUnit | null>(null);
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const [isTimelineCollapsed, setIsTimelineCollapsed] = useState(false);

  // 查询足够长的时间范围以包含所有历史数据
  const startDate = new Date('1900-01-01'); // 从1900年开始查询
  const endDate = new Date();
  endDate.setFullYear(endDate.getFullYear() + 2); // 到未来2年

  // 获取时间范围内的所有记录 - 使用GraphQL历史查询
  const { data: historyData, isLoading: isRangeLoading, error: rangeError } = useOrganizationHistory(
    organizationCode,
    {
      fromDate: TemporalConverter.dateToDateString(startDate),
      toDate: TemporalConverter.dateToDateString(endDate),
      enabled: isOpen
    }
  );

  // 获取当前选中日期的详细记录 - 使用GraphQL时间点查询
  const { data: selectedData, isLoading: isSelectedLoading } = useOrganizationAsOfDate(
    organizationCode,
    selectedDate,
    { enabled: isOpen && !!selectedDate }
  );

  // 生成时间轴节点
  const generateTimelineNodes = useCallback((): TimelineNode[] => {
    const nodes: TimelineNode[] = [];
    const today = new Date();
    
    if (historyData && historyData.length > 0) {
      // 为每个记录创建时间轴节点
      historyData.forEach(record => {
        const effectiveDate = TemporalConverter.isoToDate(record.effective_date);
        const dateStr = TemporalConverter.dateToDateString(record.effective_date);
        
        let type: 'current' | 'historical' | 'planned';
        let changeType: TimelineNode['changeType'] = 'modified';
        
        if (effectiveDate > today) {
          type = 'planned';
          changeType = 'planned';
        } else if (record.is_current) {
          type = 'current';
          changeType = record.status === 'ACTIVE' ? 'activated' : 'modified';
        } else {
          type = 'historical';
          changeType = record.status === 'INACTIVE' ? 'deactivated' : 'modified';
        }

        // 判断是否是创建记录
        if (!record.change_reason && record.created_at === record.effective_date) {
          changeType = 'created';
        }

        nodes.push({
          date: dateStr,
          displayDate: formatTimelineDate(effectiveDate, today),
          type,
          label: record.change_reason || getDefaultLabel(changeType),
          record,
          changeType
        });
      });
    }

    // 添加当前日期节点（如果没有对应记录）
    const todayStr = TemporalConverter.getCurrentDateString();
    if (!nodes.find(n => n.date === todayStr)) {
      nodes.push({
        date: todayStr,
        displayDate: '今天',
        type: 'current',
        label: '当前状态',
        changeType: 'modified'
      });
    }

    // 按日期倒序排序（最新的在上面）
    return nodes.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
  }, [historyData]);

  const timelineNodes = generateTimelineNodes();

  // 当选中日期变化时，更新选中的记录
  useEffect(() => {
    if (selectedData) {
      setSelectedRecord(selectedData);
      setEditingRecord({ ...selectedData });
      setHasUnsavedChanges(false);
    }
  }, [selectedData]);

  // 处理时间轴节点点击
  const handleTimelineNodeClick = useCallback((node: TimelineNode) => {
    if (hasUnsavedChanges) {
      const confirmed = window.confirm('有未保存的更改，确定要切换到其他时间点吗？');
      if (!confirmed) return;
    }

    setSelectedDate(node.date);
    setViewMode('view');
    setHasUnsavedChanges(false);
    
    if (node.record) {
      setSelectedRecord(node.record);
      setEditingRecord({ ...node.record });
    }
  }, [hasUnsavedChanges]);

  // 处理编辑模式切换
  const handleEditModeToggle = useCallback(() => {
    if (viewMode === 'edit') {
      if (hasUnsavedChanges) {
        const confirmed = window.confirm('有未保存的更改，确定要取消编辑吗？');
        if (!confirmed) return;
      }
      setViewMode('view');
      setHasUnsavedChanges(false);
      if (selectedRecord) {
        setEditingRecord({ ...selectedRecord });
      }
    } else {
      setViewMode('edit');
    }
  }, [viewMode, hasUnsavedChanges, selectedRecord]);

  // 处理保存
  const handleSave = useCallback(async () => {
    if (!editingRecord) return;

    try {
      await onSave(editingRecord);
      setSelectedRecord(editingRecord);
      setViewMode('view');
      setHasUnsavedChanges(false);
      
      // 刷新时间轴数据
    } catch (error) {
      console.error('保存失败:', error);
      alert('保存失败，请重试');
    }
  }, [editingRecord, onSave]);

  // 处理字段变更
  const handleFieldChange = useCallback((field: keyof TemporalOrganizationUnit, value: string | number | boolean) => {
    if (!editingRecord) return;
    
    setEditingRecord(prev => prev ? { ...prev, [field]: value } : null);
    setHasUnsavedChanges(true);
  }, [editingRecord]);

  // 处理删除
  const handleDelete = useCallback(async () => {
    if (!onDelete || !selectedRecord) return;
    
    const confirmed = window.confirm(`确定要删除组织 "${selectedRecord.name}" 吗？此操作不可恢复。`);
    if (!confirmed) return;

    try {
      await onDelete(organizationCode);
      onClose();
    } catch (error) {
      console.error('删除失败:', error);
      alert('删除失败，请重试');
    }
  }, [onDelete, selectedRecord, organizationCode, onClose]);

  if (!isOpen) return null;

  return (
    <Box
      cs={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        zIndex: 1000,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center'
      }}
    >
      <Card
        width="1400px"
        height="90vh"
        maxHeight="900px"
        overflow="hidden"
        padding="0"
      >
        {/* 标题栏 */}
        <Box
          padding={space.m}
          backgroundColor={colors.soap200}
          borderBottom={`1px solid ${colors.soap400}`}
        >
          <Flex alignItems="center" justifyContent="space-between">
            <Flex alignItems="center" gap={space.s}>
              <Text fontSize="large" fontWeight="bold">
                组织详情 - {organizationCode}
              </Text>
              
              {/* 视图模式指示器 */}
              <Badge
                variant={viewMode === 'edit' ? 'caution' : 'neutral'}
                size="small"
              >
                {viewMode === 'edit' ? '编辑模式' : '查看模式'}
              </Badge>

              {hasUnsavedChanges && (
                <Badge variant="caution" size="small">
                  有未保存更改
                </Badge>
              )}
            </Flex>

            <Flex alignItems="center" gap={space.s}>
              {/* 时间轴折叠按钮 */}
              <TertiaryButton
                onClick={() => setIsTimelineCollapsed(!isTimelineCollapsed)}
                size="small"
              >
                {isTimelineCollapsed ? '显示时间轴' : '隐藏时间轴'}
              </TertiaryButton>

              <SecondaryButton onClick={onClose}>
                关闭
              </SecondaryButton>
            </Flex>
          </Flex>
        </Box>

        {/* 主内容区域 */}
        <Flex height="calc(100% - 80px)">
          {/* 左侧时间轴 */}
          {!isTimelineCollapsed && (
            <Box
              width="320px"
              backgroundColor={colors.frenchVanilla100}
              borderRight={`1px solid ${colors.soap400}`}
              overflow="auto"
              padding={space.m}
            >
              <Flex alignItems="center" justifyContent="space-between" marginBottom={space.m}>
                <Text fontSize="medium" fontWeight="bold">
                  时间轴
                </Text>
                <Text fontSize="small" color={colors.licorice600}>
                  {timelineNodes.length} 个记录
                </Text>
              </Flex>

              {isRangeLoading ? (
                <Box cs={{ display: 'flex', alignItems: 'center', justifyContent: 'center', padding: space.l }}>
                  <Text>加载时间轴...</Text>
                </Box>
              ) : rangeError ? (
                <Box padding={space.m} backgroundColor={colors.cinnamon100} borderRadius={borderRadius.m}>
                  <Text fontSize="small" color={colors.cinnamon600}>
                    时间轴加载失败：{rangeError.message}
                  </Text>
                </Box>
              ) : (
                <Box>
                  {timelineNodes.map((node, index) => (
                    <TimelineNodeComponent
                      key={node.date}
                      node={node}
                      isSelected={selectedDate === node.date}
                      isLast={index === timelineNodes.length - 1}
                      onClick={() => handleTimelineNodeClick(node)}
                    />
                  ))}
                  
                  {timelineNodes.length === 0 && (
                    <Box 
                      padding={space.l} 
                      textAlign="center"
                      backgroundColor={colors.soap100}
                      borderRadius={borderRadius.m}
                    >
                      <Text fontSize="small" color={colors.licorice600}>
                        暂无时间轴记录
                      </Text>
                    </Box>
                  )}
                </Box>
              )}
            </Box>
          )}

          {/* 右侧详情编辑区 */}
          <Box 
            flex="1" 
            padding={space.l} 
            overflow="auto"
            backgroundColor="white"
          >
            {isSelectedLoading ? (
              <Box cs={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Text>加载详情中...</Text>
              </Box>
            ) : selectedRecord ? (
              <Box>
                {/* 详情头部操作栏 */}
                <Flex alignItems="center" justifyContent="space-between" marginBottom={space.l}>
                  <Box>
                    <Text fontSize="large" fontWeight="bold">
                      {selectedRecord.name}
                    </Text>
                    <Flex alignItems="center" gap={space.s} marginTop={space.xs}>
                      <Text fontSize="small" color={colors.licorice600}>
                        生效日期: {new Date(selectedRecord.effective_date).toLocaleDateString('zh-CN')}
                      </Text>
                      
                      {selectedRecord.end_date && (
                        <Text fontSize="small" color={colors.licorice600}>
                          结束日期: {new Date(selectedRecord.end_date).toLocaleDateString('zh-CN')}
                        </Text>
                      )}

                      <Badge
                        variant={selectedRecord.is_current ? 'positive' : 'neutral'}
                        size="small"
                      >
                        {selectedRecord.is_current ? '当前有效' : '历史记录'}
                      </Badge>
                    </Flex>
                  </Box>

                  <Flex gap={space.s}>
                    {viewMode !== 'edit' ? (
                      <>
                        <SecondaryButton onClick={handleEditModeToggle}>
                          编辑
                        </SecondaryButton>
                        
                        {onDelete && (
                          <SecondaryButton 
                            onClick={handleDelete}
                            variant="inverse"
                          >
                            删除
                          </SecondaryButton>
                        )}
                      </>
                    ) : (
                      <>
                        <SecondaryButton onClick={handleEditModeToggle}>
                          取消
                        </SecondaryButton>
                        <PrimaryButton 
                          onClick={handleSave}
                          disabled={!hasUnsavedChanges}
                        >
                          保存
                        </PrimaryButton>
                      </>
                    )}
                  </Flex>
                </Flex>

                {/* 详情表单 */}
                <OrganizationDetailForm
                  record={editingRecord || selectedRecord}
                  isEditing={viewMode === 'edit'}
                  onFieldChange={handleFieldChange}
                />
              </Box>
            ) : (
              <Box cs={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%' }}>
                <Box textAlign="center">
                  <Text fontSize="medium" marginBottom={space.s}>
                    选择时间轴上的节点查看详情
                  </Text>
                  <Text fontSize="small" color={colors.licorice600}>
                    左侧时间轴显示了该组织的历史变更记录
                  </Text>
                </Box>
              </Box>
            )}
          </Box>
        </Flex>
      </Card>
    </Box>
  );
};

// 时间轴节点组件
interface TimelineNodeComponentProps {
  node: TimelineNode;
  isSelected: boolean;
  isLast: boolean;
  onClick: () => void;
}

const TimelineNodeComponent: React.FC<TimelineNodeComponentProps> = ({
  node,
  isSelected,
  isLast,
  onClick
}) => {
  const getNodeColor = (type: TimelineNode['type']) => {
    switch (type) {
      case 'current': return colors.blueberry400;
      case 'planned': return colors.peach400;
      case 'historical': return colors.licorice400;
      default: return colors.soap600;
    }
  };

  const getNodeBackgroundColor = (type: TimelineNode['type'], selected: boolean) => {
    if (selected) {
      switch (type) {
        case 'current': return colors.blueberry50;
        case 'planned': return colors.peach50;
        case 'historical': return colors.soap100;
      }
    }
    return 'transparent';
  };

  const getChangeTypeIcon = (changeType?: TimelineNode['changeType']) => {
    switch (changeType) {
      case 'created': return <SystemIcon icon={plusIcon} size={12} color={colors.blueberry400} />;
      case 'activated': return '启用';
      case 'deactivated': return <SystemIcon icon={xIcon} size={12} color={colors.cinnamon600} />;
      case 'planned': return '计划';
      default: return '编辑';
    }
  };

  return (
    <Box position="relative" marginBottom={space.m}>
      {/* 连接线 */}
      {!isLast && (
        <Box
          position="absolute"
          left="15px"
          top="32px"
          width="2px"
          height="40px"
          backgroundColor={colors.soap400}
        />
      )}

      {/* 节点内容 */}
      <Tooltip title={`点击查看 ${node.displayDate} 的详情`}>
        <Flex
          alignItems="flex-start"
          cs={{
            cursor: 'pointer',
            padding: space.s,
            borderRadius: borderRadius.m,
            backgroundColor: getNodeBackgroundColor(node.type, isSelected),
            border: isSelected ? `2px solid ${getNodeColor(node.type)}` : '2px solid transparent'
          }}
          onClick={onClick}
        >
          {/* 时间轴圆点 */}
          <Box
            cs={{
              width: '32px',
              height: '32px',
              borderRadius: '50%',
              backgroundColor: getNodeColor(node.type),
              marginRight: space.s,
              marginTop: '2px',
              flexShrink: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              border: '2px solid white'
            }}
          >
            <Text fontSize="small">
              {getChangeTypeIcon(node.changeType)}
            </Text>
          </Box>

          {/* 节点信息 */}
          <Box flex="1" minWidth="0">
            <Text fontSize="small" fontWeight="medium">
              {node.displayDate}
            </Text>
            <Text fontSize="small" color={colors.licorice600} marginBottom={space.xs}>
              {node.label}
            </Text>
            
            {node.record && (
              <>
                <Text fontSize="small" color={colors.licorice700} fontWeight="medium">
                  {node.record.name}
                </Text>
                
                <Flex alignItems="center" gap={space.xs} marginTop={space.xs}>
                  <Badge
                    variant={
                      node.record.status === 'ACTIVE' ? 'positive' :
                      node.record.status === 'PLANNED' ? 'caution' : 'neutral'
                    }
                    size="small"
                  >
                    {node.record.status === 'ACTIVE' ? '启用' :
                     node.record.status === 'PLANNED' ? '计划' : '停用'}
                  </Badge>
                  
                  <Text fontSize="small" color={colors.licorice500}>
                    {node.record.unit_type}
                  </Text>
                </Flex>
              </>
            )}
          </Box>
        </Flex>
      </Tooltip>
    </Box>
  );
};

// 工具函数
function formatTimelineDate(date: Date, today: Date): string {
  const diffTime = date.getTime() - today.getTime();
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
  
  if (diffDays === 0) return '今天';
  if (diffDays === -1) return '昨天';
  if (diffDays === 1) return '明天';
  if (diffDays > 1 && diffDays <= 7) return `${diffDays}天后`;
  if (diffDays < -1 && diffDays >= -7) return `${Math.abs(diffDays)}天前`;
  
  return date.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit'
  });
}

function getDefaultLabel(changeType: TimelineNode['changeType']): string {
  switch (changeType) {
    case 'created': return '组织创建';
    case 'activated': return '状态激活';
    case 'deactivated': return '状态停用';
    case 'planned': return '计划变更';
    case 'modified': return '信息修改';
    default: return '状态变更';
  }
}

export default OrganizationDetailPanel;