import React from 'react'
import { Table } from '@workday/canvas-kit-react/table'
import { Text } from '@workday/canvas-kit-react/text'
import { LoadingDots } from '@workday/canvas-kit-react/loading-dots'
import { colors } from '@workday/canvas-kit-react/tokens'

type ColumnKey<T extends object> = Extract<keyof T, string>

export interface CatalogTableColumn<T extends object> {
  key: ColumnKey<T>
  label: string
  width?: string
  render?: (item: T) => React.ReactNode
}

export interface CatalogTableProps<T extends object> {
  data: T[]
  columns: CatalogTableColumn<T>[]
  isLoading?: boolean
  onRowClick?: (item: T) => void
  emptyMessage?: string
}

export const CatalogTable = <T extends object>({
  data,
  columns,
  isLoading = false,
  onRowClick,
  emptyMessage = '暂无数据',
}: CatalogTableProps<T>) => {
  const renderCell = (item: T, column: CatalogTableColumn<T>): React.ReactNode => {
    if (column.render) {
      return column.render(item)
    }

    const value = column.key in item ? (item[column.key] as unknown) : undefined
    if (value === null || value === undefined) {
      return '—'
    }
    if (typeof value === 'string' || typeof value === 'number') {
      return value
    }
    return JSON.stringify(value)
  }

  return (
    <Table>
      <Table.Head>
        <Table.Row>
          {columns.map(column => (
            <Table.Header key={column.key} width={column.width}>
              {column.label}
            </Table.Header>
          ))}
        </Table.Row>
      </Table.Head>
      <Table.Body>
        {isLoading ? (
          <Table.Row>
            <Table.Cell colSpan={columns.length}>
              <LoadingDots />
            </Table.Cell>
          </Table.Row>
        ) : data.length === 0 ? (
          <Table.Row>
            <Table.Cell colSpan={columns.length}>
              <Text textAlign="center" color={colors.licorice300}>
                {emptyMessage}
              </Text>
            </Table.Cell>
          </Table.Row>
        ) : (
          data.map((item, index) => {
            const codeCandidate = (item as Record<string, unknown>).code
            const key = typeof codeCandidate === 'string' || typeof codeCandidate === 'number' ? String(codeCandidate) : String(index)
            const clickable = typeof onRowClick === 'function'
            return (
              <Table.Row
                key={key}
                onClick={clickable ? () => onRowClick(item) : undefined}
                style={clickable ? { cursor: 'pointer' } : undefined}
              >
                {columns.map(column => (
                  <Table.Cell key={`${key}-${column.key}`}>{renderCell(item, column)}</Table.Cell>
                ))}
              </Table.Row>
            )
          })
        )}
      </Table.Body>
    </Table>
  )
}
