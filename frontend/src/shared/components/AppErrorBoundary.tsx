import React from 'react'
import { Box, Flex } from '@workday/canvas-kit-react/layout'
import { Text, Heading } from '@workday/canvas-kit-react/text'
import { PrimaryButton, SecondaryButton } from '@workday/canvas-kit-react/button'
import { colors, borderRadius } from '@workday/canvas-kit-react/tokens'

type AppErrorBoundaryState = {
  hasError: boolean
  error?: Error
}

/**
 * 轻量错误边界（用于捕获 React.lazy 加载/评估错误等渲染期异常）
 * - 不改变既有数据装载与错误呈现逻辑；仅在渲染异常时提供兜底提示，避免白屏
 */
export class AppErrorBoundary extends React.Component<React.PropsWithChildren, AppErrorBoundaryState> {
  constructor(props: React.PropsWithChildren) {
    super(props)
    this.state = { hasError: false, error: undefined }
  }

  static getDerivedStateFromError(error: Error): AppErrorBoundaryState {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, _info: React.ErrorInfo) {
    // 仅记录到控制台；项目已有 logger 体系覆盖数据请求错误
    // 这里捕获的是渲染/模块评估异常（如动态导入失败）
    // eslint-disable-next-line no-console
    console.error('[AppErrorBoundary] Caught render error:', error)
  }

  private reload = () => {
    if (typeof window !== 'undefined') {
      window.location.reload()
    }
  }

  private backToHome = () => {
    if (typeof window !== 'undefined') {
      window.location.assign('/')
    }
  }

  render(): React.ReactNode {
    if (!this.state.hasError) {
      return this.props.children
    }
    const message =
      this.state.error?.message ||
      '页面加载出现异常。请重试，若仍失败可联系管理员。'

    return (
      <Box padding="xl">
        <Box
          padding="l"
          backgroundColor={colors.cinnamon100}
          border={`1px solid ${colors.cinnamon600}`}
          borderRadius={borderRadius.l}
        >
          <Heading size="medium" color={colors.cinnamon600} marginBottom="s">
            页面渲染异常
          </Heading>
          <Text typeLevel="body.medium" color={colors.cinnamon600}>
            {message}
          </Text>
          <Flex gap="s" marginTop="m">
            <PrimaryButton onClick={this.reload}>刷新页面</PrimaryButton>
            <SecondaryButton onClick={this.backToHome}>返回首页</SecondaryButton>
          </Flex>
        </Box>
      </Box>
    )
  }
}

export default AppErrorBoundary

