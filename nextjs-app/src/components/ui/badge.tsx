import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const badgeVariants = cva(
  "inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2",
  {
    variants: {
      variant: {
        default:
          "border-transparent bg-primary text-primary-foreground hover:bg-primary/80",
        secondary:
          "border-transparent bg-secondary text-secondary-foreground hover:bg-secondary/80",
        destructive:
          "border-transparent bg-destructive text-destructive-foreground hover:bg-destructive/80",
        outline: "text-foreground",
        success:
          "border-transparent bg-green-500 text-white hover:bg-green-600",
        warning:
          "border-transparent bg-yellow-500 text-white hover:bg-yellow-600",
        error:
          "border-transparent bg-red-500 text-white hover:bg-red-600",
        processing:
          "border-transparent bg-blue-500 text-white hover:bg-blue-600",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  }
)

export interface BadgeProps
  extends React.HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {
  status?: 'success' | 'error' | 'processing' | 'default' | 'warning'
  text?: React.ReactNode
}

function Badge({ className, variant, status, text, children, ...props }: BadgeProps) {
  const statusIndicator = status && (
    <span 
      className={cn(
        "inline-block w-2 h-2 rounded-full mr-2",
        {
          'bg-green-400': status === 'success',
          'bg-red-400': status === 'error', 
          'bg-blue-400 animate-pulse': status === 'processing',
          'bg-yellow-400': status === 'warning',
          'bg-gray-400': status === 'default',
        }
      )}
    />
  )

  const badgeVariant = status ? status as VariantProps<typeof badgeVariants>['variant'] : variant

  return (
    <div className={cn(badgeVariants({ variant: badgeVariant }), className)} {...props}>
      {statusIndicator}
      {text || children}
    </div>
  )
}

export { Badge, badgeVariants }