import React from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Table } from '@workday/canvas-kit-react/table'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import type { PositionRecord } from '@/shared/types/positions'
import { getPositionStatusMeta } from '../statusMeta'

export interface PositionVersionListProps {
  versions: PositionRecord[]
  isLoading?: boolean
}

const StatusPill: React.FC<{ status: string }> = ({ status }) => {
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
        color: meta.color,
        backgroundColor: meta.background,
        border: `1px solid ${meta.border}`,
      }}
    >
      {meta.label}
    </span>
  )
}

export const PositionVersionList: React.FC<PositionVersionListProps> = ({ versions, isLoading = false }) => {
  return (
    <Card padding={space.l} backgroundColor={colors.frenchVanilla100} data-testid="position-version-list">
      <Heading size="small" marginBottom={space.m}>
        职位版本记录
      </Heading>

      {isLoading ? (
        <Text color={colors.licorice400}>正在加载职位版本...</Text>
      ) : versions.length === 0 ? (
        <Text color={colors.licorice400}>暂无职位版本记录</Text>
      ) : (
        <Table>
          <Table.Head>
            <Table.Row>
              <Table.Header width="160px">生效日期</Table.Header>
              <Table.Header width="140px">结束日期</Table.Header>
              <Table.Header width="120px">状态</Table.Header>
              <Table.Header width="200px">创建时间</Table.Header>
              <Table.Header>备注</Table.Header>
            </Table.Row>
          </Table.Head>
          <Table.Body>
            {versions.map(item => (
              <Table.Row key={`${item.code}-${item.effectiveDate}-${item.updatedAt}`}>
                <Table.Cell>{item.effectiveDate}</Table.Cell>
                <Table.Cell>{item.endDate ?? '—'}</Table.Cell>
                <Table.Cell>
                  <StatusPill status={item.status} />
                </Table.Cell>
                <Table.Cell>{item.updatedAt}</Table.Cell>
                <Table.Cell>
                  <Text color={colors.licorice400}>
                    {item.isCurrent ? '当前版本' : item.isFuture ? '计划版本' : '历史版本'}
                  </Text>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table>
      )}
    </Card>
  )
}

export default PositionVersionList
