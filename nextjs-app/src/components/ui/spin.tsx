import * as React from "react"
import { Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"

const Spin = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    spinning?: boolean
    size?: 'small' | 'default' | 'large'
    tip?: string
  }
>(({ className, spinning = true, size = 'default', tip, children, ...props }, ref) => {
  const sizeClasses = {
    small: 'w-4 h-4',
    default: 'w-6 h-6',
    large: 'w-8 h-8',
  }

  if (!spinning && children) {
    return <div ref={ref} {...props}>{children}</div>
  }

  const spinner = (
    <div className={cn("flex items-center justify-center", className)} ref={ref} {...props}>
      <div className="flex flex-col items-center gap-2">
        <Loader2 className={cn("animate-spin", sizeClasses[size])} />
        {tip && <p className="text-sm text-muted-foreground">{tip}</p>}
      </div>
    </div>
  )

  if (children && spinning) {
    return (
      <div className="relative">
        <div className="opacity-50 pointer-events-none">{children}</div>
        <div className="absolute inset-0 flex items-center justify-center bg-white/50">
          {spinner}
        </div>
      </div>
    )
  }

  return spinner
})
Spin.displayName = "Spin"

export { Spin }