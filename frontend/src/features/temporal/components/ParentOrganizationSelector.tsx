import React from 'react'
import { Combobox, useComboboxModel } from '@workday/canvas-kit-react/combobox'
import { FormField } from '@workday/canvas-kit-react/form-field'
import { Flex } from '@workday/canvas-kit-react/layout'
import { Text } from '@workday/canvas-kit-react/text'
import { UnifiedGraphQLClient } from '../../../shared/api/unified-client'
import { useOrgPBAC } from '../../../shared/hooks/useScopes'

export interface ParentOrganizationSelectorProps {
  currentCode: string
  effectiveDate: string
  onChange: (parentCode: string | undefined) => void

  currentParentCode?: string
  tenantId?: string // 多租户通过统一客户端注入，无需作为变量使用
  disabled?: boolean
  required?: boolean

  onValidationError?: (error: string) => void

  // 可选：缓存TTL（毫秒），默认5分钟
  cacheTtlMs?: number
}

type OrgItem = {
  code: string
  name: string
  unitType: string
  parentCode?: string
  level: number
  effectiveDate: string
  endDate?: string
  isFuture: boolean
}

type QueryResult = {
  organizations: {
    data: OrgItem[]
    pagination: {
      total: number
      page: number
      pageSize: number
      hasNext?: boolean
      hasPrevious?: boolean
    }
  }
}

// 默认 5 分钟 TTL 组件级缓存（以 asOfDate+pageSize 为键）
const DEFAULT_TTL_MS = 5 * 60 * 1000
const memoryCache = new Map<string, { expiresAt: number; data: OrgItem[]; total: number }>()

const QUERY = /* GraphQL */ `
  query GetValidParentOrganizations($asOfDate: String!, $pageSize: Int = 500) {
    organizations(
      filter: { status: ACTIVE, asOfDate: $asOfDate }
      pagination: { page: 1, pageSize: $pageSize, sortBy: "code", sortOrder: "asc" }
    ) {
      data {
        code
        name
        unitType
        parentCode
        level
        effectiveDate
        endDate
        isFuture
      }
      pagination { total page pageSize }
    }
  }
`

function detectCycle(currentCode: string, targetParent?: string, map?: Map<string, OrgItem>) {
  if (!targetParent || !map) return { hasCycle: false as const }
  if (currentCode === targetParent) return { hasCycle: true as const, cyclePath: [currentCode, targetParent] }
  const seen = new Set<string>()
  const path: string[] = [currentCode]
  let cur: string | undefined = targetParent
  while (cur && !seen.has(cur)) {
    path.push(cur)
    if (cur === currentCode) return { hasCycle: true as const, cyclePath: [...path] }
    seen.add(cur)
    cur = map.get(cur)?.parentCode
  }
  return { hasCycle: false as const }
}

export const ParentOrganizationSelector: React.FC<ParentOrganizationSelectorProps> = ({
  currentCode,
  effectiveDate,
  onChange,
  currentParentCode: _currentParentCode,
  disabled,
  required,
  onValidationError,
  cacheTtlMs,
}) => {
  const [loading, setLoading] = React.useState(false)
  const [error, setError] = React.useState<string | undefined>(undefined)
  const [search, setSearch] = React.useState('')
  const [selectedCode, setSelectedCode] = React.useState<string | undefined>(_currentParentCode)
  const [items, setItems] = React.useState<OrgItem[]>([])
  const { canRead } = useOrgPBAC()

  const pageSize = 500
  const cacheKey = React.useMemo(() => `${effectiveDate}::${pageSize}`, [effectiveDate])

  const orgMap = React.useMemo(() => new Map(items.map(o => [o.code, o])), [items])

  const formatLabel = React.useCallback((item?: OrgItem | null) => {
    if (!item) return ''
    return `${item.code} - ${item.name}`
  }, [])

  const filtered = React.useMemo(() => {
    if (!search) return items
    const s = search.trim().toLowerCase()
    return items.filter(it => it.code.toLowerCase().includes(s) || it.name.toLowerCase().includes(s))
  }, [items, search])

  const comboboxModel = useComboboxModel({
    items: filtered,
    getId: React.useCallback((item: OrgItem) => item.code, []),
    getTextValue: React.useCallback((item: OrgItem) => `${item.code} - ${item.name}`, []),
    shouldVirtualize: false,
    value: search,
  })

  const comboboxEventsRef = React.useRef(comboboxModel.events)
  comboboxEventsRef.current = comboboxModel.events

  React.useEffect(() => {
    let mounted = true
    if (!canRead) {
      setError('您没有权限查看组织列表')
      setItems([])
      setLoading(false)
      setSelectedCode(undefined)
      setSearch('')
      comboboxEventsRef.current?.unselectAll?.()
      comboboxEventsRef.current?.hide?.()
      return () => { mounted = false }
    }
    const cached = memoryCache.get(cacheKey)
    const now = Date.now()
    if (cached && cached.expiresAt > now) {
      setItems(cached.data)
        if (_currentParentCode) {
          const cachedItem = cached.data.find(item => item.code === _currentParentCode)
          if (cachedItem) {
            setSelectedCode(_currentParentCode)
            setSearch(formatLabel(cachedItem))
            comboboxEventsRef.current?.setSelectedIds?.([_currentParentCode])
          }
        }
        return
    }
    setLoading(true)
    const client = new UnifiedGraphQLClient()
    client
      .request<QueryResult>(QUERY, { asOfDate: effectiveDate, pageSize })
      .then((data) => {
        if (!mounted) return
        const list = (data.organizations?.data || []).filter(o => o.code !== currentCode)
        setItems(list)
        setError(undefined)
        if (_currentParentCode) {
          const preselected = list.find(item => item.code === _currentParentCode)
          if (preselected) {
            setSelectedCode(_currentParentCode)
            setSearch(formatLabel(preselected))
            comboboxEventsRef.current?.setSelectedIds?.([_currentParentCode])
          }
        }
        memoryCache.set(cacheKey, { data: list, total: data.organizations?.pagination?.total || list.length, expiresAt: now + (cacheTtlMs ?? DEFAULT_TTL_MS) })
      })
      .catch((e: unknown) => {
        if (!mounted) return
        const msg = e instanceof Error ? e.message : '加载组织列表失败，请稍后重试'
        setError(msg)
        onValidationError?.(msg)
      })
      .finally(() => mounted && setLoading(false))
    return () => {
      mounted = false
    }
  }, [cacheKey, cacheTtlMs, canRead, currentCode, effectiveDate, formatLabel, onValidationError, _currentParentCode])

  React.useEffect(() => {
    if (!_currentParentCode) {
      return
    }
    if (_currentParentCode === selectedCode) {
      return
    }
    const match = orgMap.get(_currentParentCode)
    if (!match) {
      return
    }
    setSelectedCode(_currentParentCode)
    setSearch(formatLabel(match))
    comboboxEventsRef.current?.setSelectedIds?.([_currentParentCode])
  }, [_currentParentCode, formatLabel, orgMap, selectedCode])

  const clearSelection = React.useCallback(() => {
    comboboxEventsRef.current?.unselectAll?.()
    setSelectedCode(undefined)
    setSearch('')
    setError(undefined)
    onValidationError?.(undefined)
    onChange(undefined)
  }, [onChange, onValidationError])

  const applySelection = React.useCallback((nextCode: string | undefined) => {
    if (!nextCode) {
      clearSelection()
      return
    }
    const item = orgMap.get(nextCode)
    if (!item) {
      return
    }
    const { hasCycle, cyclePath } = detectCycle(currentCode, nextCode, orgMap)
    if (hasCycle) {
      const pathStr = cyclePath?.join(' -> ') || ''
      const msg = `选择该组织将导致循环依赖：${pathStr}`
      setError(msg)
      onValidationError?.(msg)
      if (selectedCode) {
        comboboxEventsRef.current?.setSelectedIds?.([selectedCode])
        const existing = orgMap.get(selectedCode)
        setSearch(formatLabel(existing))
      } else {
        comboboxEventsRef.current?.unselectAll?.()
        setSearch('')
      }
      return
    }
    setError(undefined)
    onValidationError?.(undefined)
    setSelectedCode(nextCode)
    setSearch(formatLabel(item))
    onChange(nextCode)
    comboboxEventsRef.current?.hide?.()
  }, [clearSelection, currentCode, formatLabel, onChange, onValidationError, orgMap, selectedCode])

  const selectedIds = comboboxModel.state?.selectedIds

  React.useEffect(() => {
    if (!selectedIds || selectedIds === 'all') {
      return
    }
    const [next] = selectedIds
    if (!next && selectedCode) {
      clearSelection()
      return
    }
    if (next && next !== selectedCode) {
      applySelection(next)
    }
  }, [applySelection, clearSelection, selectedCode, selectedIds])

  const handleInputChange = React.useCallback(
    (event: React.ChangeEvent<HTMLInputElement>) => {
      const value = event.target.value
      setSearch(value)
      comboboxEventsRef.current?.show?.()
      if (error) {
        setError(undefined)
        onValidationError?.(undefined)
      }
      if (!value) {
        clearSelection()
      } else {
        comboboxEventsRef.current?.unselectAll?.()
      }
    },
    [clearSelection, error, onValidationError]
  )

  const handleComboboxChange = React.useCallback(
    (value: string) => {
      const code = value?.split('#')[0] || value
      if (!code) {
        clearSelection()
        return
      }
      comboboxEventsRef.current?.setSelectedIds?.([code])
      applySelection(code)
    },
    [applySelection, clearSelection]
  )

  return (
    <FormField error={error} required={required}>
      <FormField.Label>上级组织</FormField.Label>
      <FormField.Field>
        <Combobox
          model={comboboxModel}
          onChange={handleComboboxChange}
          items={filtered}
        >
          <Combobox.Input
            data-testid="combobox-input"
            placeholder={loading ? '加载中…' : '搜索并选择上级组织...'}
            value={search}
            onFocus={() => canRead && !disabled && comboboxEventsRef.current?.show?.()}
            onChange={handleInputChange}
            disabled={disabled || loading || !canRead}
          />
          <Combobox.Menu>
            <Combobox.Menu.Popper>
              <Combobox.Menu.Card>
                {loading && canRead ? (
                  <Flex padding="s" justifyContent="center">
                    <Text size="small">加载中…</Text>
                  </Flex>
                ) : (
                  <Combobox.Menu.List>
                    {(item: OrgItem) => (
                      <Combobox.Menu.Item key={item.code} data-id={item.code} data-testid={`combobox-item-${item.code}`}>
                        <Flex direction="column" gap="xxs">
                          <Text fontWeight="semibold">{item.code} - {item.name}</Text>
                          <Text size="small" color="frenchVanilla100" as="span">
                            层级: {item.level ?? '-'} ｜ 上级: {item.parentCode || '-'}
                          </Text>
                        </Flex>
                      </Combobox.Menu.Item>
                    )}
                  </Combobox.Menu.List>
                )}
                {!loading && canRead && filtered.length === 0 && (
                  <Flex padding="s">
                    <Text size="small" data-testid="combobox-empty">未找到匹配的组织</Text>
                  </Flex>
                )}
                {!canRead && (
                  <Flex padding="s">
                    <Text size="small">您没有权限查看组织列表</Text>
                  </Flex>
                )}
              </Combobox.Menu.Card>
            </Combobox.Menu.Popper>
          </Combobox.Menu>
        </Combobox>
      </FormField.Field>
      <FormField.Hint>显示在所选生效日期有效且状态为 ACTIVE 的组织</FormField.Hint>
      {error && <FormField.Error>{error}</FormField.Error>}
    </FormField>
  )
}

export default ParentOrganizationSelector
