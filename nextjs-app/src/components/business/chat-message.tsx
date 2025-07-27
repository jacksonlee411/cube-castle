'use client'

import { useState } from 'react'
import { format } from 'date-fns'
import { zhCN } from 'date-fns/locale'
import { Bot, User, Copy, Check, Volume2, VolumeX } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card } from '@/components/ui/card'
import { cn } from '@/lib/utils'
import toast from 'react-hot-toast'

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: Date
  intent?: string
  confidence?: number
  loading?: boolean
}

interface ChatMessageProps {
  message: ChatMessage
  isLatest?: boolean
}

export function ChatMessage({ message, isLatest }: ChatMessageProps) {
  const [copied, setCopied] = useState(false)
  const [speaking, setSpeaking] = useState(false)

  const isUser = message.role === 'user'
  const isAssistant = message.role === 'assistant'

  // 复制消息内容
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(message.content)
      setCopied(true)
      toast.success('消息已复制到剪贴板')
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      toast.error('复制失败')
    }
  }

  // 语音播报 (Web Speech API)
  const handleSpeak = () => {
    if (speaking) {
      // 停止播报
      window.speechSynthesis.cancel()
      setSpeaking(false)
      return
    }

    if ('speechSynthesis' in window) {
      const utterance = new SpeechSynthesisUtterance(message.content)
      utterance.lang = 'zh-CN'
      utterance.rate = 0.9
      utterance.pitch = 1

      utterance.onstart = () => setSpeaking(true)
      utterance.onend = () => setSpeaking(false)
      utterance.onerror = () => {
        setSpeaking(false)
        toast.error('语音播报失败')
      }

      window.speechSynthesis.speak(utterance)
    } else {
      toast.error('您的浏览器不支持语音播报')
    }
  }

  // 获取置信度颜色
  const getConfidenceColor = (confidence?: number) => {
    if (!confidence) return 'secondary'
    if (confidence >= 0.8) return 'default'
    if (confidence >= 0.6) return 'secondary'
    return 'destructive'
  }

  // 获取置信度文本
  const getConfidenceText = (confidence?: number) => {
    if (!confidence) return '未知'
    if (confidence >= 0.8) return '高'
    if (confidence >= 0.6) return '中'
    return '低'
  }

  return (
    <div className={cn(
      "flex gap-4",
      isUser ? "flex-row-reverse" : "flex-row"
    )}>
      {/* 头像 */}
      <div className={cn(
        "flex h-8 w-8 shrink-0 items-center justify-center rounded-lg",
        isUser 
          ? "bg-primary text-primary-foreground" 
          : "bg-muted text-muted-foreground"
      )}>
        {isUser ? <User className="h-4 w-4" /> : <Bot className="h-4 w-4" />}
      </div>

      {/* 消息内容 */}
      <div className={cn(
        "flex-1 space-y-2",
        isUser ? "text-right" : "text-left"
      )}>
        {/* 消息气泡 */}
        <Card className={cn(
          "inline-block max-w-[80%] px-4 py-3",
          isUser 
            ? "bg-primary text-primary-foreground ml-auto" 
            : "bg-muted",
          message.loading && "opacity-75"
        )}>
          <div className="space-y-2">
            {/* 消息文本 */}
            <div className="whitespace-pre-wrap break-words">
              {message.loading ? (
                <div className="flex items-center space-x-2">
                  <div className="flex space-x-1">
                    <div className="h-2 w-2 animate-bounce rounded-full bg-current [animation-delay:-0.3s]" />
                    <div className="h-2 w-2 animate-bounce rounded-full bg-current [animation-delay:-0.15s]" />
                    <div className="h-2 w-2 animate-bounce rounded-full bg-current" />
                  </div>
                  <span className="text-sm">AI正在思考...</span>
                </div>
              ) : (
                message.content
              )}
            </div>

            {/* AI消息的元信息 */}
            {isAssistant && !message.loading && message.intent && (
              <div className="flex flex-wrap gap-2 pt-2">
                <Badge variant="outline" className="text-xs">
                  意图: {message.intent}
                </Badge>
                {message.confidence !== undefined && (
                  <Badge 
                    variant={getConfidenceColor(message.confidence)}
                    className="text-xs"
                  >
                    置信度: {getConfidenceText(message.confidence)} ({Math.round(message.confidence * 100)}%)
                  </Badge>
                )}
              </div>
            )}
          </div>
        </Card>

        {/* 消息操作 */}
        {!message.loading && (
          <div className={cn(
            "flex items-center gap-2 text-xs text-muted-foreground",
            isUser ? "justify-end" : "justify-start"
          )}>
            {/* 时间戳 */}
            <span>
              {format(message.timestamp, 'HH:mm', { locale: zhCN })}
            </span>

            {/* 操作按钮 */}
            <div className="flex items-center">
              {/* 复制按钮 */}
              <Button
                variant="ghost"
                size="sm"
                className="h-6 w-6 p-0"
                onClick={handleCopy}
              >
                {copied ? (
                  <Check className="h-3 w-3 text-green-500" />
                ) : (
                  <Copy className="h-3 w-3" />
                )}
              </Button>

              {/* 语音播报按钮 (仅AI消息) */}
              {isAssistant && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-6 w-6 p-0"
                  onClick={handleSpeak}
                >
                  {speaking ? (
                    <VolumeX className="h-3 w-3" />
                  ) : (
                    <Volume2 className="h-3 w-3" />
                  )}
                </Button>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}