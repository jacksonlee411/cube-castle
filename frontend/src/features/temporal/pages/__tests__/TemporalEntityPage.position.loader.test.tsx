import { describe, it, expect, vi } from 'vitest'
import { render, cleanup } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import TemporalEntityPage from '../TemporalEntityPage'
import type { TemporalEntityRouteConfig } from '../TemporalEntityPage'
import * as positions from '@/shared/hooks/useEnterprisePositions'

// 提前 mock queryClient 模块，注入可监控的 cancelQueries
const hoisted = vi.hoisted(() => {
  return {
    cancelQueriesMock: vi.fn().mockResolvedValue(undefined),
  }
})
vi.mock('@/shared/api/queryClient', async (orig) => {
  const actual = await (orig() as Promise<Record<string, unknown>>)
  return {
    ...(actual as object),
    queryClient: { cancelQueries: hoisted.cancelQueriesMock },
  }
})

// 确保测试结束后清理
afterEach(() => {
  cleanup()
  vi.restoreAllMocks()
})

const positionConfig: TemporalEntityRouteConfig = {
  entity: 'position',
  listPath: '/positions',
  buildDetailPath: (code) => `/positions/${code}`,
  parseCode: (raw) => ({ isCreateMode: false, code: raw?.toUpperCase(), rawCode: raw }),
  invalidMessages: {
    missing: { title: 'x', description: 'x' },
    invalid: { title: 'x', description: 'x' },
  },
  renderContent: () => <div data-testid="dummy">OK</div>,
}

describe('TemporalEntityPage – position loader integration (240B)', () => {
  it('calls prefetch on mount and cancels on unmount', async () => {
    const prefetchSpy = vi.spyOn(positions, 'prefetchPositionDetail').mockResolvedValue(undefined)

    // 设置特性开关
    // @ts-expect-error define env
    import.meta.env = { VITE_TEMPORAL_DETAIL_LOADER: 'true' }

    const ui = (
      <MemoryRouter initialEntries={['/positions/P1234567']}>
        <Routes>
          <Route path="/positions/:code" element={<TemporalEntityPage config={positionConfig} />} />
        </Routes>
      </MemoryRouter>
    )

    const view = render(ui)
    expect(view.getByTestId('dummy')).toBeInTheDocument()

    // 由于 prefetch 是异步 fire-and-forget，这里仅断言被调用
    expect(prefetchSpy).toHaveBeenCalledOnce()
    expect(prefetchSpy.mock.calls[0]?.[1]).toBe('P1234567')

    // 触发卸载
    view.unmount()

    // 取消查询通过 queryClient.cancelQueries 调用
    expect(hoisted.cancelQueriesMock).toHaveBeenCalled()
  })
})
