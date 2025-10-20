import React from 'react'
import { Flex } from '@workday/canvas-kit-react/layout'
import { TextInput } from '@workday/canvas-kit-react/text-input'
import { Checkbox } from '@workday/canvas-kit-react/checkbox'
import { SecondaryButton } from '@workday/canvas-kit-react/button'
import { space } from '@workday/canvas-kit-react/tokens'

export interface CatalogFiltersProps {
  searchPlaceholder?: string
  searchValue: string
  onSearchChange: (value: string) => void
  includeInactive: boolean
  onIncludeInactiveChange: (checked: boolean) => void
  asOfDate?: string
  onAsOfDateChange?: (value: string | undefined) => void
  extraFilters?: React.ReactNode
  onReset?: () => void
}

export const CatalogFilters: React.FC<CatalogFiltersProps> = ({
  searchPlaceholder = '输入关键字搜索',
  searchValue,
  onSearchChange,
  includeInactive,
  onIncludeInactiveChange,
  asOfDate,
  onAsOfDateChange,
  extraFilters,
  onReset,
}) => {
  return (
    <div
      style={{
        marginBottom: space.l,
        display: 'flex',
        flexDirection: 'column',
        gap: space.s,
        padding: 0,
      }}
    >
      <Flex gap={space.s} flexWrap="wrap">
        <TextInput
          placeholder={searchPlaceholder}
          value={searchValue}
          onChange={event => onSearchChange(event.target.value)}
          width={320}
        />
        {typeof onAsOfDateChange === 'function' && (
          <TextInput
            type="date"
            value={asOfDate ?? ''}
            onChange={event => onAsOfDateChange(event.target.value || undefined)}
          />
        )}
        <Checkbox
          checked={includeInactive}
          onChange={event => onIncludeInactiveChange(event.target.checked)}
          label="包含停用记录"
        />
        {extraFilters}
        {typeof onReset === 'function' && (
          <SecondaryButton onClick={onReset}>重置</SecondaryButton>
        )}
      </Flex>
    </div>
  )
}
