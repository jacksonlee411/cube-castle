import React from 'react'
import { Table } from '@workday/canvas-kit-react/table'
import { Text } from '@workday/canvas-kit-react/text'
import { colors } from '@workday/canvas-kit-react/tokens'
import { StatusBadge } from '../../../shared/components/StatusBadge'
import type { PositionMock } from '../mockData'

export interface PositionListProps {
  positions: PositionMock[]
  selectedCode?: string
  onSelect: (code: string) => void
}

const statusToStatusBadge = (status: PositionMock['status']) => {
  switch (status) {
    case 'FILLED':
      return 'ACTIVE'
    case 'VACANT':
      return 'INACTIVE'
    case 'PLANNED':
      return 'PLANNED'
    case 'INACTIVE':
      return 'INACTIVE'
    default:
      return 'ACTIVE'
  }
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
          <Table.Header width="160px">主管</Table.Header>
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
            const available = Math.max(item.headcountCapacity - item.headcountInUse, 0)
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
                    {item.organization.name}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <Text>{item.jobFamilyGroup}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.jobFamily}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <Text>{item.jobRole}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.jobLevel}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <Text>
                    {item.headcountInUse} / {item.headcountCapacity}
                  </Text>
                  <Text fontSize="12px" color={colors.celery500}>
                    可用 {available}
                  </Text>
                </Table.Cell>
                <Table.Cell>
                  <StatusBadge status={statusToStatusBadge(item.status)} size="small" />
                </Table.Cell>
                <Table.Cell>
                  <Text>{item.supervisor.name}</Text>
                  <Text fontSize="12px" color={colors.licorice400}>
                    {item.supervisor.code}
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
