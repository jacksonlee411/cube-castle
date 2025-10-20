import React from 'react'
import { Table } from '@workday/canvas-kit-react/table'
import { Text } from '@workday/canvas-kit-react/text'
import { colors } from '@workday/canvas-kit-react/tokens'
import type { PositionRecord } from '@/shared/types/positions'
import { getPositionStatusMeta } from '../../statusMeta'

export interface PositionListProps {
  positions: PositionRecord[]
  selectedCode?: string
  onSelect: (code: string) => void
}

const PositionStatusPill: React.FC<{ status: string }> = ({ status }) => {
  const meta = getPositionStatusMeta(status)
  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '4px 8px',
        borderRadius: 12,
        fontSize: 12,
        fontWeight: 600,
        minWidth: 60,
        color: meta.color,
        backgroundColor: meta.background,
        border: `1px solid ${meta.border}`,
      }}
    >
      {meta.label}
    </span>
  )
}

export const PositionList: React.FC<PositionListProps> = ({ positions, selectedCode, onSelect }) => {
  return (
    <Table data-testid="position-table">
      <Table.Head>
        <Table.Row>
          <Table.Header width="120px">职位编码</Table.Header>
          <Table.Header>职位名称</Table.Header>
          <Table.Header>职类 / 职种</Table.Header>
          <Table.Header>职务 / 职级</Table.Header>
          <Table.Header width="140px">编制</Table.Header>
          <Table.Header width="120px">状态</Table.Header>
          <Table.Header width="160px">汇报对象</Table.Header>
        </Table.Row>
      </Table.Head>
      <Table.Body>
        {positions.length === 0 ? (
          <Table.Row>
            <Table.Cell colSpan={7}>
              <Text textAlign="center" color="hint">
                暂无职位数据
              </Text>
            </Table.Cell>
          </Table.Row>
        ) : (
          positions.map(item => {
            const isSelected = selectedCode === item.code
            return (
              <Table.Row
                key={item.code}
                data-testid={`position-row-${item.code}`}
                onClick={() => onSelect(item.code)}
                style={{
                  cursor: 'pointer',
                  backgroundColor: isSelected ? colors.blueberry50 : undefined,
                }}
              >
                <Table.Cell>{item.code}</Table.Cell>
                <Table.Cell>
                  <Text fontWeight="bold">{item.title}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.organizationName ?? '未设置归属组织'}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <Text>{item.jobFamilyGroupName ?? item.jobFamilyGroupCode}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.jobFamilyName ?? item.jobFamilyCode}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <Text>{item.jobRoleName ?? item.jobRoleCode}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.jobLevelName ?? item.jobLevelCode}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <Text>
                    {item.headcountInUse} / {item.headcountCapacity}
                  </Text>
                  <Text fontSize="12px" color={colors.celery500}>
                    可用 {item.availableHeadcount}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <PositionStatusPill status={item.status} />
                </Table.Cell>
                <Table.Cell>
                  <Text>{item.reportsToPositionCode ?? '未设置'}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.reportsToPositionCode ? '汇报职位编码' : '暂无上级信息'}
                  </Text>
                </Table.Cell>
              </Table.Row>
            )
          })
        )}
      </Table.Body>
    </Table>
  )
}

export default PositionList
