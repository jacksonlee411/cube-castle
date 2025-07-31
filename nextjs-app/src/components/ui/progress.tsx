import * as React from "react"
import { cn } from "@/lib/utils"

const Progress = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    value?: number
    max?: number
    status?: 'normal' | 'active' | 'exception' | 'success'
    strokeColor?: string
    showInfo?: boolean
  }
>(({ className, value = 0, max = 100, status = 'normal', showInfo = true, ...props }, ref) => {
  const percentage = Math.min(Math.max((value / max) * 100, 0), 100)
  
  return (
    <div ref={ref} className={cn("w-full", className)} {...props}>
      <div className="flex items-center gap-2">
        <div className="flex-1 bg-secondary rounded-full h-2 overflow-hidden">
          <div
            className={cn(
              "h-full transition-all duration-300 ease-in-out",
              {
                'bg-primary': status === 'normal',
                'bg-blue-500': status === 'active',
                'bg-red-500': status === 'exception',
                'bg-green-500': status === 'success',
              }
            )}
            style={{ width: `${percentage}%` }}
          />
        </div>
        {showInfo && (
          <span className="text-sm text-muted-foreground min-w-[3rem] text-right">
            {Math.round(percentage)}%
          </span>
        )}
      </div>
    </div>
  )
})
Progress.displayName = "Progress"

export { Progress }