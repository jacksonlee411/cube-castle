import React from 'react'
import { Card } from '@workday/canvas-kit-react/card'
import { Table } from '@workday/canvas-kit-react/table'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { colors, space } from '@workday/canvas-kit-react/tokens'
import type { PositionRecord } from '@/shared/types/positions'
import { getPositionStatusMeta } from '../statusMeta'
import { SimpleStack } from './SimpleStack'
import { POSITION_VERSION_FIELDS, type PositionVersionFieldKey } from './positionVersionFields'

export interface PositionVersionDiffProps {
  baseVersion: PositionRecord | null
  compareVersion: PositionRecord | null
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
        minWidth: 48,
        textAlign: 'center',
      }}
    >
      {meta.label}
    </span>
  )
}

const renderValue = (version: PositionRecord | null, key: PositionVersionFieldKey): React.ReactNode => {
  if (!version) {
    return '—'
  }

  const rawValue = (version as Record<string, unknown>)[key]

  switch (key) {
    case 'status':
      return <StatusPill status={String(rawValue ?? '')} />
    case 'headcountCapacity':
    case 'headcountInUse':
    case 'availableHeadcount':
      return typeof rawValue === 'number' ? rawValue.toFixed(rawValue % 1 === 0 ? 0 : 1) : '0'
    case 'isCurrent':
    case 'isFuture':
      return rawValue ? '是' : '否'
    default:
      return rawValue === null || rawValue === undefined || rawValue === '' ? '—' : String(rawValue)
  }
}

const hasDifference = (
  baseVersion: PositionRecord | null,
  compareVersion: PositionRecord | null,
  key: PositionVersionFieldKey,
) => {
  const baseRaw = baseVersion ? (baseVersion as Record<string, unknown>)[key] : undefined
  const compareRaw = compareVersion ? (compareVersion as Record<string, unknown>)[key] : undefined

  if (baseRaw === compareRaw) {
    return false
  }

  if (baseRaw === undefined || baseRaw === null) {
    return compareRaw !== undefined && compareRaw !== null
  }
  if (compareRaw === undefined || compareRaw === null) {
    return true
  }

  if (typeof baseRaw === 'number' && typeof compareRaw === 'number') {
    return Math.abs(baseRaw - compareRaw) > Number.EPSILON
  }

  return String(baseRaw) !== String(compareRaw)
}

export const PositionVersionDiff: React.FC<PositionVersionDiffProps> = ({
  baseVersion,
  compareVersion,
  isLoading = false,
}) => {
  return (
    <Card padding={space.l} backgroundColor={colors.frenchVanilla100} data-testid="position-version-diff">
      <SimpleStack gap={space.m}>
        <Heading size="small">版本差异对比</Heading>

        {isLoading ? (
          <Text color={colors.licorice400}>正在加载职位版本，请稍候...</Text>
        ) : !baseVersion ? (
          <Text color={colors.licorice400}>请选择基准版本以查看差异信息。</Text>
        ) : !compareVersion ? (
          <Text color={colors.licorice400}>请选择对比版本，或保持空白查看单个版本详情。</Text>
        ) : (
          <Table data-testid="position-version-diff-table">
            <Table.Head>
              <Table.Row>
                <Table.Header width="160px">字段</Table.Header>
                <Table.Header>基准版本</Table.Header>
                <Table.Header>对比版本</Table.Header>
              </Table.Row>
            </Table.Head>
            <Table.Body>
              {POSITION_VERSION_FIELDS.map(field => {
                const difference = hasDifference(baseVersion, compareVersion, field.key)
                const highlightStyle = difference
                  ? {
                      backgroundColor: colors.cantaloupe100,
                    }
                  : undefined

                return (
                  <Table.Row key={field.key as string}>
                    <Table.Cell>{field.label}</Table.Cell>
                    <Table.Cell style={difference ? { fontWeight: 600 } : undefined}>
                      {renderValue(baseVersion, field.key)}
                    </Table.Cell>
                    <Table.Cell style={highlightStyle}>{renderValue(compareVersion, field.key)}</Table.Cell>
                  </Table.Row>
                )
              })}
            </Table.Body>
          </Table>
        )}
      </SimpleStack>
    </Card>
  )
}

export default PositionVersionDiff
