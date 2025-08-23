/**
 * 时态组织编辑器组件 (纯日期生效模型)
 * 左侧时间轴 + 右侧详情编辑的布局设计
 */
import React, { useState, useEffect, useCallback } from 'react';
import { Box, Flex } from '@workday/canvas-kit-react/layout';
import { Text } from '@workday/canvas-kit-react/text';
import { Card } from '@workday/canvas-kit-react/card';
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button';
import { TextInput } from '@workday/canvas-kit-react/text-input';
import { TextArea } from '@workday/canvas-kit-react/text-area';
import { colors, space, borderRadius } from '@workday/canvas-kit-react/tokens';
import { useOrganizationHistory, useOrganizationAsOfDate } from '../../../shared/hooks/useTemporalGraphQL';
import type { TemporalOrganizationUnit } from '../../../shared/types/temporal';
import { TemporalConverter } from '../../../shared/utils/temporal-converter';

export interface TemporalOrganizationEditorProps {
  /** 组织代码 */
  organizationCode: string;
  /** 是否显示弹窗 */
  isOpen: boolean;
  /** 关闭回调 */
  onClose: () => void;
  /** 保存回调 */
  onSave: (record: TemporalOrganizationUnit) => Promise<void>;
}

// 时间轴节点类型
interface TimelineNode {
  date: string;
  displayDate: string;
  type: 'current' | 'historical' | 'planned';
  label: string;
  record?: TemporalOrganizationUnit;
}

/**
 * 时态组织编辑器组件
 */
export const TemporalOrganizationEditor: React.FC<TemporalOrganizationEditorProps> = ({
  organizationCode,
  isOpen,
  onClose,
  onSave
}) => {
  // 状态管理
  const [selectedDate, setSelectedDate] = useState<string>(TemporalConverter.getCurrentDateString());
  const [selectedRecord, setSelectedRecord] = useState<TemporalOrganizationUnit | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editingRecord, setEditingRecord] = useState<TemporalOrganizationUnit | null>(null);

  // 查询过去6个月到未来3个月的时间范围
  const sixMonthsAgo = new Date();
  sixMonthsAgo.setMonth(sixMonthsAgo.getMonth() - 6);
  const threeMonthsLater = new Date();
  threeMonthsLater.setMonth(threeMonthsLater.getMonth() + 3);

  // 获取时间范围内的所有记录 - 使用GraphQL历史查询
  const { data: historyData, isLoading: isRangeLoading } = useOrganizationHistory(
    organizationCode,
    {
      fromDate: TemporalConverter.dateToDateString(sixMonthsAgo),
      toDate: TemporalConverter.dateToDateString(threeMonthsLater),
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
        if (effectiveDate > today) {
          type = 'planned';
        } else if (record.is_current) {
          type = 'current';
        } else {
          type = 'historical';
        }

        nodes.push({
          date: dateStr,
          displayDate: effectiveDate.toLocaleDateString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit'
          }),
          type,
          label: record.change_reason || getDefaultLabel(type),
          record
        });
      });
    }

    // 添加当前日期节点（如果没有记录）
    const todayStr = TemporalConverter.getCurrentDateString();
    if (!nodes.find(n => n.date === todayStr)) {
      nodes.push({
        date: todayStr,
        displayDate: '今天',
        type: 'current',
        label: '当前状态',
      });
    }

    // 按日期排序
    return nodes.sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime());
  }, [historyData]);

  const timelineNodes = generateTimelineNodes();

  // 当选中日期变化时，更新选中的记录
  useEffect(() => {
    if (selectedData) {
      setSelectedRecord(selectedData);
      setEditingRecord({ ...selectedData });
    }
  }, [selectedData]);

  // 处理时间轴节点点击
  const handleTimelineNodeClick = useCallback((node: TimelineNode) => {
    setSelectedDate(node.date);
    setIsEditing(false);
    if (node.record) {
      setSelectedRecord(node.record);
      setEditingRecord({ ...node.record });
    }
  }, []);

  // 处理开始编辑
  const handleStartEdit = useCallback(() => {
    setIsEditing(true);
    if (selectedRecord) {
      setEditingRecord({ ...selectedRecord });
    }
  }, [selectedRecord]);

  // 处理取消编辑
  const handleCancelEdit = useCallback(() => {
    setIsEditing(false);
    if (selectedRecord) {
      setEditingRecord({ ...selectedRecord });
    }
  }, [selectedRecord]);

  // 处理保存编辑
  const handleSaveEdit = useCallback(async () => {
    if (!editingRecord) return;

    try {
      await onSave(editingRecord);
      setSelectedRecord(editingRecord);
      setIsEditing(false);
    } catch (error) {
      console.error('保存失败:', error);
    }
  }, [editingRecord, onSave]);

  // 处理字段变更
  const handleFieldChange = useCallback((field: keyof TemporalOrganizationUnit, value: string | number | boolean) => {
    if (!editingRecord) return;
    setEditingRecord(prev => prev ? { ...prev, [field]: value } : null);
  }, [editingRecord]);

  if (!isOpen) return null;

  return (
    <Box
      cs={{
        position: "fixed",
        top: "0",
        left: "0",
        right: "0",
        bottom: "0",
        backgroundColor: "rgba(0, 0, 0, 0.5)",
        zIndex: 1000,
        display: "flex",
        alignItems: "center",
        justifyContent: "center"
      }}
    >
      <Card
        cs={{
          width: "1200px",
          height: "80vh",
          maxHeight: "800px",
          overflow: "hidden",
          padding: "0"
        }}
      >
        {/* 标题栏 */}
        <Box
          padding={space.m}
          backgroundColor={colors.soap200}
          borderBottom={`1px solid ${colors.soap400}`}
        >
          <Flex alignItems="center" justifyContent="space-between">
            <Text fontSize="large" fontWeight="bold">
              时态组织编辑器 - {organizationCode}
            </Text>
            <SecondaryButton onClick={onClose}>
              关闭
            </SecondaryButton>
          </Flex>
        </Box>

        {/* 主内容区域 */}
        <Flex height="calc(100% - 80px)">
          {/* 左侧时间轴 */}
          <Box
            width="300px"
            backgroundColor={colors.frenchVanilla100}
            borderRight={`1px solid ${colors.soap400}`}
            overflow="auto"
            padding={space.m}
          >
            <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
              时间轴
            </Text>

            {isRangeLoading ? (
              <Text>加载中...</Text>
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
              </Box>
            )}
          </Box>

          {/* 右侧详情编辑区 */}
          <Box 
            flex={1} 
            padding={space.l} 
            cs={{ overflow: "auto" }}
          >
            {isSelectedLoading ? (
              <Flex alignItems="center" justifyContent="center" height="100%">
                <Text>加载详情中...</Text>
              </Flex>
            ) : selectedRecord ? (
              <Box>
                {/* 记录信息头部 */}
                <Flex alignItems="center" justifyContent="space-between" marginBottom={space.l}>
                  <Box>
                    <Text fontSize="large" fontWeight="bold">
                      {isEditing ? '编辑模式' : '查看模式'}
                    </Text>
                    <Text fontSize="small" color={colors.licorice600}>
                      生效日期: {new Date(selectedRecord.effective_date).toLocaleDateString('zh-CN')}
                    </Text>
                  </Box>
                  <Box>
                    {!isEditing ? (
                      <PrimaryButton onClick={handleStartEdit}>
                        编辑
                      </PrimaryButton>
                    ) : (
                      <Flex gap={space.s}>
                        <SecondaryButton onClick={handleCancelEdit}>
                          取消
                        </SecondaryButton>
                        <PrimaryButton onClick={handleSaveEdit}>
                          保存
                        </PrimaryButton>
                      </Flex>
                    )}
                  </Box>
                </Flex>

                {/* 详情表单 */}
                <OrganizationDetailForm
                  record={editingRecord || selectedRecord}
                  isEditing={isEditing}
                  onFieldChange={handleFieldChange}
                />
              </Box>
            ) : (
              <Flex alignItems="center" justifyContent="center" height="100%">
                <Text>选择时间轴上的节点查看详情</Text>
              </Flex>
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

  return (
    <Box position="relative" marginBottom={space.m}>
      {/* 连接线 */}
      {!isLast && (
        <Box
          cs={{
            position: "absolute",
            left: "11px",
            top: "24px",
            width: "2px",
            height: "32px",
            backgroundColor: colors.soap400
          }}
        />
      )}

      {/* 节点内容 */}
      <Flex
        alignItems="flex-start"
        cs={{
          cursor: "pointer",
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
            width: "24px",
            height: "24px",
            borderRadius: "50%",
            backgroundColor: getNodeColor(node.type),
            marginRight: space.s,
            marginTop: "2px",
            flexShrink: 0,
            display: "flex",
            alignItems: "center",
            justifyContent: "center"
          }}
        >
          <Box
            cs={{
              width: "8px",
              height: "8px",
              borderRadius: "50%",
              backgroundColor: "white"
            }}
          />
        </Box>

        {/* 节点信息 */}
        <Box flex="1" minWidth="0">
          <Text fontSize="small" fontWeight="medium">
            {node.displayDate}
          </Text>
          <Text fontSize="small" color={colors.licorice600}>
            {node.label}
          </Text>
          {node.record && (
            <Text fontSize="small" color={colors.licorice500}>
              {node.record.name}
            </Text>
          )}
        </Box>
      </Flex>
    </Box>
  );
};

// 组织详情表单组件
interface OrganizationDetailFormProps {
  record: TemporalOrganizationUnit;
  isEditing: boolean;
  onFieldChange: (field: keyof TemporalOrganizationUnit, value: string | number | boolean) => void;
}

const OrganizationDetailForm: React.FC<OrganizationDetailFormProps> = ({
  record,
  isEditing,
  onFieldChange
}) => {
  return (
    <Box>
      {/* 基础信息 */}
      <Box marginBottom={space.l}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          基础信息
        </Text>

        <Box 
          cs={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: space.m
          }}
          marginBottom={space.m}
        >
          <Box>
            <Text fontSize="small" marginBottom={space.xs}>组织代码</Text>
            <TextInput
              value={record.code}
              disabled={true}
              cs={{
                backgroundColor: colors.soap200
              }}
            />
          </Box>

          <Box>
            <Text fontSize="small" marginBottom={space.xs}>组织名称</Text>
            <TextInput
              value={record.name}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('name', e.target.value)}
            />
          </Box>
        </Box>

        <Box 
          cs={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: space.m
          }}
          marginBottom={space.m}
        >
          <Box>
            <Text fontSize="small" marginBottom={space.xs}>组织类型</Text>
            <select
              value={record.unit_type}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('unit_type', e.target.value)}
              style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
            >
              <option value="ORGANIZATION_UNIT">组织单位</option>
              <option value="DEPARTMENT">部门</option>
              <option value="PROJECT_TEAM">项目团队</option>
            </select>
          </Box>

          <Box>
            <Text fontSize="small" marginBottom={space.xs}>状态</Text>
            <select
              value={record.status}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('status', e.target.value)}
              style={{ width: '100%', padding: '8px', border: '1px solid #ddd', borderRadius: '4px' }}
            >
              <option value="ACTIVE">启用</option>
              <option value="INACTIVE">停用</option>
              <option value="PLANNED">计划中</option>
            </select>
          </Box>
        </Box>

        <Box>
          <Text fontSize="small" marginBottom={space.xs}>描述</Text>
          <TextArea
            value={record.description || ''}
            disabled={!isEditing}
            onChange={(e) => isEditing && onFieldChange('description', e.target.value)}
            rows={3}
          />
        </Box>
      </Box>

      {/* 时态信息 */}
      <Box marginBottom={space.l}>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          时态信息
        </Text>

        <Box 
          cs={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: space.m
          }}
          marginBottom={space.m}
        >
          <Box>
            <Text fontSize="small" marginBottom={space.xs}>生效日期</Text>
            <TextInput
              type="date"
              value={record.effective_date?.slice(0, 10) || ''}
              disabled={!isEditing}
              onChange={(e) => {
                if (isEditing && e.target.value) {
                  onFieldChange('effective_date', e.target.value + 'T00:00:00Z');
                }
              }}
            />
          </Box>

          <Box>
            <Text fontSize="small" marginBottom={space.xs}>结束日期</Text>
            <TextInput
              type="date"
              value={record.end_date?.slice(0, 10) || ''}
              disabled={!isEditing}
              onChange={(e) => isEditing && onFieldChange('end_date', e.target.value ? e.target.value + 'T00:00:00Z' : '')}
            />
          </Box>
        </Box>

        <Box>
          <Text fontSize="small" marginBottom={space.xs}>变更原因</Text>
          <TextArea
            value={record.change_reason || ''}
            disabled={!isEditing}
            onChange={(e) => isEditing && onFieldChange('change_reason', e.target.value)}
            rows={2}
          />
        </Box>
      </Box>

      {/* 审批信息 */}
      <Box>
        <Text fontSize="medium" fontWeight="bold" marginBottom={space.m}>
          审批信息
        </Text>

        <Box 
          cs={{
            display: "grid",
            gridTemplateColumns: "1fr 1fr",
            gap: space.m
          }}
        >
          <Box>
            <Text fontSize="small" marginBottom={space.xs}>批准人</Text>
            <TextInput
              value={''}
              disabled={true}
              placeholder="暂无批准人信息"
              cs={{
                backgroundColor: colors.soap200
              }}
            />
          </Box>

          <Box>
            <Text fontSize="small" marginBottom={space.xs}>批准时间</Text>
            <TextInput
              value={''}
              disabled={true}
              placeholder="暂无批准时间信息"
              cs={{
                backgroundColor: colors.soap200
              }}
            />
          </Box>
        </Box>
      </Box>
    </Box>
  );
};

// 工具函数
function getDefaultLabel(type: TimelineNode['type']): string {
  switch (type) {
    case 'current': return '当前状态';
    case 'planned': return '计划变更';
    case 'historical': return '历史记录';
    default: return '未知状态';
  }
}

export default TemporalOrganizationEditor;