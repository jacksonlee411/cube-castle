import React, { useState } from 'react';
import {
  Box,
  Flex,
  Text
} from '@workday/canvas-kit-react';
import { SystemIcon } from '@workday/canvas-kit-react/icon';
import {
  chevronDownIcon,
  chevronUpIcon
} from '@workday/canvas-system-icons-web';
import { colors, space } from '@workday/canvas-kit-react/tokens';
import type { JsonObject, JsonValue } from '@/shared/types/json';
import type { OrganizationOperationType } from '@/shared/types/contract_gen';

// 字段变更接口定义
export interface FieldChange {
  field: string;
  oldValue: JsonValue | null;
  newValue: JsonValue | null;
  dataType: string;
}

// 组件Props接口
interface FieldChangeTableProps {
  /** 操作类型 */
  operationType: OrganizationOperationType;
  /** 字段变更列表 */
  changes?: FieldChange[];
  /** 创建后数据 (CREATE操作使用) */
  afterData?: JsonObject | null;
  /** 删除前数据 (DELETE操作使用) */
  beforeData?: JsonObject | null;
  /** 是否允许折叠 */
  collapsible?: boolean;
  /** 默认是否展开 */
  defaultExpanded?: boolean;
}

// 字段名本地化映射
const fieldDisplayNames: Record<string, string> = {
  'code': '组织代码',
  'name': '组织名称',
  'unitType': '组织类型',
  'status': '状态',
  'description': '描述',
  'sortOrder': '排序号',
  'parentCode': '上级组织',
  'effectiveDate': '生效日期',
  'endDate': '结束日期',
  'changeReason': '变更原因',
  'level': '层级',
  'createdAt': '创建时间',
  'updatedAt': '更新时间'
};

// 数据类型标签映射
const dataTypeLabels: Record<string, string> = {
  'string': '[文本]',
  'int': '[数字]',
  'date': '[日期]',
  'boolean': '[是/否]',
  'datetime': '[时间]'
};

/**
 * 字段变更表格组件
 * 根据操作类型展示不同格式的数据变更表格
 */
export const FieldChangeTable: React.FC<FieldChangeTableProps> = ({
  operationType,
  changes = [],
  afterData,
  beforeData,
  collapsible = true,
  defaultExpanded = false
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  // 格式化显示值
  const formatValue = (value: JsonValue | null | undefined, dataType?: string): string => {
    if (value === null || value === undefined) {
      return '-';
    }

    if (value === '') {
      return '(空)';
    }

    if (typeof value === 'boolean') {
      return value ? '是' : '否';
    }

    if (typeof value === 'number') {
      return Number.isFinite(value) ? value.toString() : '(非数字)';
    }

    if (dataType === 'date' || dataType === 'datetime') {
      try {
        const date = new Date(String(value));
        return new Intl.DateTimeFormat('zh-CN', {
          year: 'numeric',
          month: '2-digit',
          day: '2-digit',
          hour: dataType === 'datetime' ? '2-digit' : undefined,
          minute: dataType === 'datetime' ? '2-digit' : undefined,
          timeZone: 'Asia/Shanghai'
        }).format(date);
      } catch {
        return String(value);
      }
    }

    if (typeof value === 'object') {
      try {
        return JSON.stringify(value);
      } catch {
        return '[复杂数据]';
      }
    }

    const stringValue = String(value);
    // 长文本截断
    if (stringValue.length > 50) {
      return stringValue.substring(0, 47) + '...';
    }
    
    return stringValue;
  };

  // 获取字段显示名称
  const getFieldDisplayName = (fieldName: string): string => {
    return fieldDisplayNames[fieldName] || fieldName;
  };

  // 获取数据类型标签
  const getDataTypeLabel = (dataType: string): string => {
    return dataTypeLabels[dataType] || `[${dataType}]`;
  };

  // 渲染表格头部
  const renderTableHeader = (columns: string[]) => (
    <Flex
      padding={space.s}
      style={{
        backgroundColor: colors.soap100,
        borderBottom: `1px solid ${colors.soap300}`,
        fontWeight: 'bold'
      }}
    >
      {columns.map((column, index) => (
        <Box 
          key={index}
          flex={index === 0 ? '0 0 30%' : '1'}
          paddingX={space.xs}
        >
          <Text typeLevel="subtext.medium" fontWeight="bold" color={colors.licorice600}>
            {column}
          </Text>
        </Box>
      ))}
    </Flex>
  );

  // 渲染UPDATE操作表格
  const renderUpdateTable = () => {
    // 若后端未提供changes，尝试用快照兜底推导变更（只处理一层浅比较）
    let sourceChanges = changes;
    if (!sourceChanges.length && (beforeData || afterData)) {
      const b = (beforeData ?? {}) as Record<string, JsonValue | null | undefined>;
      const a = (afterData ?? {}) as Record<string, JsonValue | null | undefined>;
      const keys = Array.from(new Set([...Object.keys(b), ...Object.keys(a)])).filter(
        (k) => k !== 'id',
      );
      const inferType = (v: unknown): string => {
        if (v === null || v === undefined) return 'string';
        if (typeof v === 'boolean') return 'boolean';
        if (typeof v === 'number') return Number.isInteger(v) ? 'int' : 'number';
        if (typeof v === 'string') {
          // 简单判断日期/时间
          if (/^\d{4}-\d{2}-\d{2}$/.test(v)) return 'date';
          if (/^\d{4}-\d{2}-\d{2}T/.test(v)) return 'datetime';
          return 'string';
        }
        return 'string';
      };
      const derived: FieldChange[] = [];
      for (const key of keys) {
        const oldV = b[key] ?? null;
        const newV = a[key] ?? null;
        const oldStr = JSON.stringify(oldV);
        const newStr = JSON.stringify(newV);
        if (oldStr !== newStr) {
          derived.push({
            field: key,
            oldValue: oldV as JsonValue | null,
            newValue: newV as JsonValue | null,
            dataType: inferType(newV ?? oldV),
          });
        }
      }
      sourceChanges = derived;
    }

    if (!sourceChanges.length) return null;

    return (
      <Box style={{ border: `1px solid ${colors.soap300}`, borderRadius: '4px' }}>
        {renderTableHeader(['字段名称', '变动前', '变动后'])}
        
        {sourceChanges.map((change, index) => (
          <Flex
            key={index}
            padding={space.s}
            style={{
              backgroundColor: index % 2 === 0 ? 'white' : colors.soap50,
              borderBottom: index < sourceChanges.length - 1 ? `1px solid ${colors.soap200}` : 'none'
            }}
          >
            {/* 字段名称列 */}
            <Box flex="0 0 30%" paddingX={space.xs}>
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {getFieldDisplayName(change.field)}
              </Text>
              <Text typeLevel="subtext.small" color={colors.licorice400}>
                {getDataTypeLabel(change.dataType)}
              </Text>
            </Box>
            
            {/* 变动前列 */}
            <Box 
              as="div"
              flex="1" 
              paddingX={space.xs}
              style={{
                backgroundColor: colors.cinnamon50,
                margin: `0 ${space.xs}`,
                padding: space.xs,
                borderRadius: '4px',
                border: `1px solid ${colors.cinnamon200}`
              }}
            >
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {formatValue(change.oldValue, change.dataType)}
              </Text>
            </Box>
            
            {/* 变动后列 */}
            <Box 
              as="div"
              flex="1" 
              paddingX={space.xs}
              style={{
                backgroundColor: colors.greenApple50,
                margin: `0 ${space.xs}`,
                padding: space.xs,
                borderRadius: '4px',
                border: `1px solid ${colors.greenApple200}`
              }}
            >
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {formatValue(change.newValue, change.dataType)}
              </Text>
            </Box>
          </Flex>
        ))}
      </Box>
    );
  };

  // 渲染CREATE操作表格
  const renderCreateTable = () => {
    if (!afterData) return null;

    const entries = Object.entries(afterData).filter(([key, value]) => 
      value !== null && value !== undefined && key !== 'id'
    );

    if (!entries.length) return null;

    return (
      <Box style={{ border: `1px solid ${colors.soap300}`, borderRadius: '4px' }}>
        {renderTableHeader(['字段名称', '初始值'])}
        
        {entries.map(([key, value], index) => (
          <Flex
            key={index}
            padding={space.s}
            style={{
              backgroundColor: index % 2 === 0 ? 'white' : colors.soap50,
              borderBottom: index < entries.length - 1 ? `1px solid ${colors.soap200}` : 'none'
            }}
          >
            {/* 字段名称列 */}
            <Box flex="0 0 30%" paddingX={space.xs}>
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {getFieldDisplayName(key)}
              </Text>
              <Text typeLevel="subtext.small" color={colors.licorice400}>
                [文本]
              </Text>
            </Box>
            
            {/* 初始值列 */}
            <Box flex="1" paddingX={space.xs}>
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {formatValue(value)}
              </Text>
            </Box>
          </Flex>
        ))}
      </Box>
    );
  };

  // 渲染DELETE操作表格
  const renderDeleteTable = () => {
    if (!beforeData) return null;

    const entries = Object.entries(beforeData).filter(([key, value]) => 
      value !== null && value !== undefined && key !== 'id'
    );

    if (!entries.length) return null;

    return (
      <Box style={{ border: `1px solid ${colors.soap300}`, borderRadius: '4px' }}>
        {renderTableHeader(['字段名称', '删除前值'])}
        
        {entries.map(([key, value], index) => (
          <Flex
            key={index}
            padding={space.s}
            style={{
              backgroundColor: index % 2 === 0 ? 'white' : colors.soap50,
              borderBottom: index < entries.length - 1 ? `1px solid ${colors.soap200}` : 'none'
            }}
          >
            {/* 字段名称列 */}
            <Box flex="0 0 30%" paddingX={space.xs}>
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {getFieldDisplayName(key)}
              </Text>
              <Text typeLevel="subtext.small" color={colors.licorice400}>
                [文本]
              </Text>
            </Box>
            
            {/* 删除前值列 */}
            <Box flex="1" paddingX={space.xs}>
              <Text typeLevel="subtext.medium" color={colors.licorice600}>
                {formatValue(value)}
              </Text>
            </Box>
          </Flex>
        ))}
      </Box>
    );
  };

  // 选择渲染函数
  const renderTable = () => {
    switch (operationType) {
      case 'UPDATE':
        return renderUpdateTable();
      case 'CREATE':
        return renderCreateTable();
      case 'DELETE':
        return renderDeleteTable();
      // 状态类操作（停用/重新启用/作废）统一按“更新表格”展示 changes
      // 这些事件通常仅提供 changes（以及可能的 before/after 片段），不依赖删除前快照
      case 'SUSPEND':
      case 'REACTIVATE':
      case 'DEACTIVATE':
        return renderUpdateTable();
      default:
        return null;
    }
  };

  // 获取表格标题
  const getTableTitle = () => {
    switch (operationType) {
      case 'UPDATE':
        return `字段变更详情 (${changes.length} 项)`;
      // 状态类操作走更新语义展示
      case 'SUSPEND':
      case 'REACTIVATE':
      case 'DEACTIVATE':
        return `字段变更详情 (${changes.length} 项)`;
      case 'CREATE':
        return '创建记录 - 初始数据';
      case 'DELETE':
        return '删除记录 - 删除前状态';
      default:
        return '数据变更详情';
    }
  };

  const tableContent = renderTable();
  
  if (!tableContent) {
    return (
      <Box marginTop={space.m}>
        <Text typeLevel="subtext.medium" color={colors.licorice400}>
          暂无变更数据
        </Text>
      </Box>
    );
  }

  return (
    <Box marginTop={space.m}>
      {/* 表格标题和折叠控制 */}
      {collapsible ? (
        <Flex
          alignItems="center"
          justifyContent="space-between"
          marginBottom={space.s}
          onClick={() => setIsExpanded(!isExpanded)}
          style={{ cursor: 'pointer' }}
        >
          <Text typeLevel="subtext.medium" fontWeight="bold" color={colors.licorice600}>
            {getTableTitle()}
          </Text>
          <SystemIcon 
            icon={isExpanded ? chevronUpIcon : chevronDownIcon}
            size={16}
            color={colors.licorice400}
          />
        </Flex>
      ) : (
        <Text 
          typeLevel="subtext.medium" 
          fontWeight="bold" 
          color={colors.licorice600}
          marginBottom={space.s}
        >
          {getTableTitle()}
        </Text>
      )}

      {/* 表格内容 */}
      {(!collapsible || isExpanded) && tableContent}
    </Box>
  );
};

export default FieldChangeTable;
