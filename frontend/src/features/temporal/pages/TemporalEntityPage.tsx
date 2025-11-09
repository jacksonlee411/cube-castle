import React, { useCallback } from 'react'
import type { NavigateOptions, Params } from 'react-router-dom'
import { useNavigate, useParams } from 'react-router-dom'
import { Box } from '@workday/canvas-kit-react/layout'
import { Heading, Text } from '@workday/canvas-kit-react/text'
import { SecondaryButton } from '@workday/canvas-kit-react/button'

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
