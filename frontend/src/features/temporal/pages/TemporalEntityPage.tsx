import React, { useCallback, useEffect } from 'react'
import type { NavigateOptions, Params } from 'react-router-dom'
import { useNavigate, useParams } from 'react-router-dom'
import { Box } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { SecondaryButton } from '@workday/canvas-kit-react/button'
import { queryClient } from '@/shared/api/queryClient'
import { prefetchPositionDetail, POSITION_DETAIL_QUERY_ROOT_KEY } from '@/shared/hooks/useEnterprisePositions'
import { createTemporalDetailLoader } from '@/shared/loaders/temporalDetailLoader'

export type TemporalEntityKind = 'organization' | 'position'

export type TemporalEntityInvalidKind = 'missing' | 'invalid'

export interface TemporalEntityInvalidMessage {
  title: string
  description: string
  actionLabel?: string
}

export interface TemporalEntityParseResult {
  code?: string
  isCreateMode: boolean
  rawCode?: string
  error?: TemporalEntityInvalidKind
}

export interface TemporalEntityRenderContext {
  entity: TemporalEntityKind
  code?: string
  rawCode?: string
  isCreateMode: boolean
  navigateToList: () => void
  navigateToDetail: (targetCode: string, options?: NavigateOptions) => void
  params: Readonly<Params<string>>
}

export interface TemporalEntityRouteConfig {
  entity: TemporalEntityKind
  listPath: string
  buildDetailPath: (code: string) => string
  parseCode: (rawCode?: string) => TemporalEntityParseResult
  renderContent: (ctx: TemporalEntityRenderContext) => React.ReactNode
  invalidMessages: Record<TemporalEntityInvalidKind, TemporalEntityInvalidMessage>
}

export interface TemporalEntityPageProps {
  config: TemporalEntityRouteConfig
}

const TemporalEntityPage: React.FC<TemporalEntityPageProps> = ({ config }) => {
  const params = useParams<{ code?: string }>()
  const navigate = useNavigate()
  const rawCode = params.code
  const parseResult = config.parseCode(rawCode)

  const { listPath, buildDetailPath } = config

  const navigateToList = useCallback(() => {
    navigate(listPath)
  }, [navigate, listPath])

  const navigateToDetail = useCallback(
    (targetCode: string, options?: NavigateOptions) => {
      navigate(buildDetailPath(targetCode), options)
    },
    [navigate, buildDetailPath],
  )

  // 240B – 路由级 Loader 预热（特性开关）
  useEffect(() => {
    const enabled =
      import.meta.env?.VITE_TEMPORAL_DETAIL_LOADER !== 'false' &&
      !!parseResult.code &&
      config.entity === 'position' &&
      !parseResult.isCreateMode;
    if (!enabled) return;

    const code = parseResult.code!;
    // 通过统一 Loader 工厂进行预热与取消，符合 240B/241 的“统一入口”目标
    const loader = createTemporalDetailLoader({
      entity: 'position',
      prefetch: (client, c) => prefetchPositionDetail(client, c, false),
      cancelQueryRootKey: POSITION_DETAIL_QUERY_ROOT_KEY,
    });
    loader.preheat(queryClient, code).catch(() => {
      // 预热失败不阻塞渲染；实际错误由消费端 Hook 处理
    })
    return () => {
      // 取消在途查询（按职位详情根键范围）
      loader.cancel(queryClient).catch(() => { /* ignore cancel errors */ })
    };
  }, [config.entity, parseResult.code, parseResult.isCreateMode]);

  if (parseResult.error) {
    const message = config.invalidMessages[parseResult.error]
    return (
      <Box padding="xl" textAlign="center">
        <Heading size="medium" marginBottom="m">
          {message.title}
        </Heading>
        <Text typeLevel="body.medium" color="hint" marginBottom="l">
          {message.description}
        </Text>
        <SecondaryButton onClick={navigateToList}>
          {message.actionLabel ?? '返回列表'}
        </SecondaryButton>
      </Box>
    )
  }

  return (
    <>
      {config.renderContent({
        entity: config.entity,
        code: parseResult.code,
        rawCode: parseResult.rawCode ?? rawCode,
        isCreateMode: parseResult.isCreateMode,
        navigateToList,
        navigateToDetail,
        params,
      })}
    </>
  )
}

export default TemporalEntityPage
