import React from 'react'
import { Combobox } from '@workday/canvas-kit-react/combobox'
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
  const [items, setItems] = React.useState<OrgItem[]>([])
  const { canRead } = useOrgPBAC()

  const pageSize = 500
  const cacheKey = React.useMemo(() => `${effectiveDate}::${pageSize}`, [effectiveDate])

  const orgMap = React.useMemo(() => new Map(items.map(o => [o.code, o])), [items])

  React.useEffect(() => {
    let mounted = true
    if (!canRead) {
      setError('您没有权限查看组织列表')
      setItems([])
      setLoading(false)
      return () => { mounted = false }
    }
    const cached = memoryCache.get(cacheKey)
    const now = Date.now()
    if (cached && cached.expiresAt > now) {
      setItems(cached.data)
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
  }, [cacheKey, currentCode, effectiveDate, onValidationError, canRead])

  const filtered = React.useMemo(() => {
    if (!search) return items
    const s = search.trim().toLowerCase()
    return items.filter(it => it.code.toLowerCase().includes(s) || it.name.toLowerCase().includes(s))
  }, [items, search])

  const handleSelect = (value: string) => {
    const [code] = (value || '').split('#')
    const selected = code || undefined
    if (!selected) {
      onChange(undefined)
      return
    }
    const { hasCycle, cyclePath } = detectCycle(currentCode, selected, orgMap)
    if (hasCycle) {
      const pathStr = cyclePath?.join(' -> ') || ''
      const msg = `选择该组织将导致循环依赖：${pathStr}`
      setError(msg)
      onValidationError?.(msg)
      return
    }
    setError(undefined)
    onChange(selected)
  }

  return (
    <FormField error={error} data-testid="form-field" data-error={error}>
      <FormField.Label required={required} data-testid="form-field-label">上级组织</FormField.Label>

      <Combobox
        data-testid="combobox"
        items={filtered.map(f => `${f.code}#${f.name}`)}
        onChange={handleSelect}
        disabled={disabled || loading || !canRead}
      >
        <Combobox.Input
          data-testid="combobox-input"
          placeholder={loading ? '加载中…' : '搜索并选择上级组织...'}
          value={search}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearch(e.target.value)}
          disabled={disabled || loading || !canRead}
        />
        <Combobox.Menu data-testid="combobox-menu">
          <div data-testid="combobox-items">
          <Combobox.MenuList>
            {(item: string) => {
              const [code, name] = item.split('#')
              const org = orgMap.get(code)
              return (
                <Combobox.Item
                  key={code}
                  data-testid={`combobox-item-${code}#${name}`}
                  onClick={() => handleSelect(`${code}#${name}`)}
                >
                  <Flex
                    direction="column"
                    gap="xxs"
                    onClick={() => onChange(code)}
                    data-testid={`combobox-select-${code}`}
                  >
                    <Text weight="medium">{code} - {name}</Text>
                    <Text typeLevel="subtext.small" variant="hint">
                      层级: {org?.level ?? '-'} | 上级: {org?.parentCode || '-'}
                    </Text>
                  </Flex>
                </Combobox.Item>
              )
            }}
          </Combobox.MenuList>
          </div>
        </Combobox.Menu>
      </Combobox>

      <FormField.Hint>显示在所选生效日期有效且状态为 ACTIVE 的组织</FormField.Hint>
      {error && <FormField.Error>{error}</FormField.Error>}
    </FormField>
  )
}

export default ParentOrganizationSelector
