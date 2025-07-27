'use client'

import { useState, useRef, useEffect } from 'react'
import { Send, Bot, User, RotateCcw, Trash2, Copy, Volume2 } from 'lucide-react'
import { AppNav } from '@/components/business/app-nav'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { ChatMessage } from '@/components/business/chat-message'
import { ChatSuggestions } from '@/components/business/chat-suggestions'
import { useChatStore } from '@/store'
import { apiClient } from '@/lib/api-client'
import toast from 'react-hot-toast'

export default function ChatPage() {
  const [input, setInput] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const scrollAreaRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // 聊天状态
  const {
    messages,
    sessionId,
    loading,
    connected,
    addMessage,
    updateLastMessage,
    setLoading,
    setConnected,
    clearMessages,
    newSession
  } = useChatStore()

  // 自动滚动到底部
  const scrollToBottom = () => {
    if (scrollAreaRef.current) {
      const scrollElement = scrollAreaRef.current.querySelector('[data-radix-scroll-area-viewport]')
      if (scrollElement) {
        scrollElement.scrollTop = scrollElement.scrollHeight
      }
    }
  }

  // 监听消息变化，自动滚动
  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // 聚焦输入框
  useEffect(() => {
    if (inputRef.current && !loading) {
      inputRef.current.focus()
    }
  }, [loading])

  // 发送消息
  const handleSendMessage = async (messageText?: string) => {
    const textToSend = messageText || input
    if (!textToSend.trim() || loading) return

    // 清空输入框
    setInput('')
    setLoading(true)

    // 添加用户消息
    addMessage({
      role: 'user',
      content: textToSend
    })

    // 添加加载中的AI消息
    addMessage({
      role: 'assistant',
      content: '正在思考...',
      loading: true
    })

    try {
      // 调用AI服务
      const response = await apiClient.intelligence.interpretText({
        text: textToSend,
        sessionId: sessionId
      })

      // 更新AI回复
      updateLastMessage({
        content: response.response || '抱歉，我暂时无法理解您的意图，请尝试重新表达。',
        intent: response.intent,
        confidence: response.confidence,
        loading: false
      })

      // 连接状态正常
      setConnected(true)

    } catch (error: any) {
      console.error('AI request failed:', error)
      
      // 更新错误消息
      updateLastMessage({
        content: '抱歉，AI服务暂时不可用，请稍后再试。',
        loading: false
      })

      // 更新连接状态
      setConnected(false)
      
      toast.error('AI服务连接失败')
    } finally {
      setLoading(false)
    }
  }

  // 处理键盘事件
  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  // 建议消息列表
  const suggestions = [
    {
      title: '员工管理',
      message: '帮我查询张三的员工信息',
      icon: User
    },
    {
      title: '组织架构',
      message: '显示技术部的组织架构',
      icon: Bot
    },
    {
      title: '工作流状态',
      message: '查看我的待办工作流',
      icon: RotateCcw
    },
    {
      title: '系统状态',
      message: '系统运行状态如何',
      icon: Volume2
    }
  ]

  // 清空对话
  const handleClearChat = () => {
    if (confirm('确定要清空所有对话记录吗？')) {
      clearMessages()
      toast.success('对话记录已清空')
    }
  }

  // 新建会话
  const handleNewSession = () => {
    if (confirm('确定要开始新的会话吗？当前对话记录将被清空。')) {
      newSession()
      toast.success('已开始新的会话')
    }
  }

  return (
    <div className="min-h-screen bg-background">
      <AppNav />
      <div className="flex h-[calc(100vh-4rem)] flex-col">
      {/* 聊天头部 */}
      <div className="border-b bg-background p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <Bot className="h-6 w-6" />
              </div>
              <div>
                <h1 className="text-lg font-semibold">AI智能助手</h1>
                <p className="text-sm text-muted-foreground">
                  {connected ? '已连接' : '连接异常'} • 会话ID: {sessionId.slice(-8)}
                </p>
              </div>
            </div>
            <Badge variant={connected ? 'default' : 'destructive'}>
              {connected ? '在线' : '离线'}
            </Badge>
          </div>

          <div className="flex items-center space-x-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleNewSession}
              disabled={loading}
            >
              <RotateCcw className="mr-2 h-4 w-4" />
              新会话
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={handleClearChat}
              disabled={messages.length === 0 || loading}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              清空
            </Button>
          </div>
        </div>
      </div>

      {/* 聊天内容区域 */}
      <div className="flex-1 overflow-hidden">
        <ScrollArea className="h-full" ref={scrollAreaRef}>
          <div className="p-4">
            {messages.length === 0 ? (
              // 欢迎界面
              <div className="flex h-full flex-col items-center justify-center space-y-6">
                <div className="text-center">
                  <div className="flex h-20 w-20 items-center justify-center rounded-full bg-primary/10 mx-auto mb-4">
                    <Bot className="h-10 w-10 text-primary" />
                  </div>
                  <h2 className="text-2xl font-bold mb-2">欢迎使用AI智能助手</h2>
                  <p className="text-muted-foreground mb-8 max-w-md">
                    我可以帮助您管理员工信息、查询组织架构、监控工作流状态等。
                    请选择下面的建议或直接输入您的问题。
                  </p>
                </div>

                {/* 建议消息 */}
                <ChatSuggestions
                  suggestions={suggestions}
                  onSuggestionClick={handleSendMessage}
                  disabled={loading}
                />
              </div>
            ) : (
              // 消息列表
              <div className="space-y-6">
                {messages.map((message, index) => (
                  <ChatMessage
                    key={`${message.id}-${index}`}
                    message={message}
                    isLatest={index === messages.length - 1}
                  />
                ))}
              </div>
            )}
          </div>
        </ScrollArea>
      </div>

      {/* 输入区域 */}
      <div className="border-t bg-background p-4">
        <div className="mx-auto max-w-4xl">
          <div className="flex space-x-2">
            <div className="flex-1">
              <Input
                ref={inputRef}
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyPress={handleKeyPress}
                placeholder={
                  connected 
                    ? "输入您的问题..."
                    : "AI服务连接异常，请稍后再试..."
                }
                disabled={loading || !connected}
                className="text-base"
              />
            </div>
            <Button
              onClick={() => handleSendMessage()}
              disabled={!input.trim() || loading || !connected}
              size="sm"
              className="px-4"
            >
              {loading ? (
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
              ) : (
                <Send className="h-4 w-4" />
              )}
            </Button>
          </div>

          {/* 输入提示 */}
          <div className="mt-2 text-center text-xs text-muted-foreground">
            按 Enter 发送消息，Shift + Enter 换行
          </div>
        </div>
      </div>
      </div>
    </div>
  )
}