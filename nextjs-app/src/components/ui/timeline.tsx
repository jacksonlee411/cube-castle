import * as React from "react"
import { cn } from "@/lib/utils"

const Timeline = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement>
>(({ className, ...props }, ref) => (
  <div ref={ref} className={cn("relative", className)} {...props} />
))
Timeline.displayName = "Timeline"

const TimelineItem = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    dot?: React.ReactNode
    color?: 'default' | 'blue' | 'green' | 'red' | 'gray'
    pending?: boolean
  }
>(({ className, dot, color = 'default', pending, children, ...props }, ref) => {
  const colorClasses = {
    default: 'bg-primary border-primary',
    blue: 'bg-blue-500 border-blue-500',
    green: 'bg-green-500 border-green-500',
    red: 'bg-red-500 border-red-500',
    gray: 'bg-gray-400 border-gray-400',
  }

  return (
    <div ref={ref} className={cn("relative flex pb-8 last:pb-0", className)} {...props}>
      <div className="flex flex-col items-center mr-4">
        <div className={cn(
          "w-3 h-3 rounded-full border-2 z-10",
          pending ? 'border-gray-300 bg-white' : colorClasses[color]
        )}>
          {dot}
        </div>
        <div className="w-px bg-border flex-1 mt-1 last:hidden" />
      </div>
      <div className="flex-1 min-w-0">
        {children}
      </div>
    </div>
  )
})
TimelineItem.displayName = "TimelineItem"

export { Timeline, TimelineItem }