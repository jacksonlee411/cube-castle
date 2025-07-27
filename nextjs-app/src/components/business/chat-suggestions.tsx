'use client'

import { LucideIcon } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'

interface Suggestion {
  title: string
  message: string
  icon: LucideIcon
}

interface ChatSuggestionsProps {
  suggestions: Suggestion[]
  onSuggestionClick: (message: string) => void
  disabled?: boolean
}

export function ChatSuggestions({
  suggestions,
  onSuggestionClick,
  disabled = false
}: ChatSuggestionsProps) {
  return (
    <div className="w-full max-w-2xl">
      <div className="mb-4 text-center">
        <h3 className="text-lg font-medium mb-2">常用功能</h3>
        <p className="text-sm text-muted-foreground">
          点击下面的建议快速开始对话
        </p>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
        {suggestions.map((suggestion, index) => {
          const IconComponent = suggestion.icon
          
          return (
            <Card
              key={index}
              className="cursor-pointer transition-all duration-200 hover:shadow-md hover:scale-105"
            >
              <CardContent className="p-4">
                <Button
                  variant="ghost"
                  className="w-full h-full p-0 justify-start"
                  onClick={() => onSuggestionClick(suggestion.message)}
                  disabled={disabled}
                >
                  <div className="flex items-center space-x-3 w-full">
                    <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                      <IconComponent className="h-5 w-5" />
                    </div>
                    <div className="flex-1 text-left">
                      <div className="font-medium text-sm mb-1">
                        {suggestion.title}
                      </div>
                      <div className="text-xs text-muted-foreground line-clamp-2">
                        {suggestion.message}
                      </div>
                    </div>
                  </div>
                </Button>
              </CardContent>
            </Card>
          )
        })}
      </div>
      
      {/* 更多建议提示 */}
      <div className="mt-6 text-center">
        <p className="text-xs text-muted-foreground">
          您也可以直接输入问题，比如：
          <br />
          "查找员工张三的信息" 或 "显示技术部的组织架构"
        </p>
      </div>
    </div>
  )
}