import * as React from "react"
import { cn } from "@/lib/utils"

const Typography = React.forwardRef<
  HTMLDivElement,
  React.HTMLAttributes<HTMLDivElement> & {
    variant?: 'h1' | 'h2' | 'h3' | 'h4' | 'p' | 'text'
  }
>(({ className, variant = 'p', ...props }, ref) => {
  if (variant === 'h1') {
    return (
      <h1
        className={cn('scroll-m-20 text-4xl font-extrabold tracking-tight lg:text-5xl', className)}
        ref={ref as React.ForwardedRef<HTMLHeadingElement>}
        {...props}
      />
    )
  }
  
  if (variant === 'h2') {
    return (
      <h2
        className={cn('scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight first:mt-0', className)}
        ref={ref as React.ForwardedRef<HTMLHeadingElement>}
        {...props}
      />
    )
  }
  
  if (variant === 'h3') {
    return (
      <h3
        className={cn('scroll-m-20 text-2xl font-semibold tracking-tight', className)}
        ref={ref as React.ForwardedRef<HTMLHeadingElement>}
        {...props}
      />
    )
  }
  
  if (variant === 'h4') {
    return (
      <h4
        className={cn('scroll-m-20 text-xl font-semibold tracking-tight', className)}
        ref={ref as React.ForwardedRef<HTMLHeadingElement>}
        {...props}
      />
    )
  }
  
  if (variant === 'text') {
    return (
      <span
        className={cn('text-sm font-medium leading-none', className)}
        ref={ref as React.ForwardedRef<HTMLSpanElement>}
        {...props}
      />
    )
  }

  return (
    <p
      className={cn('leading-7 [&:not(:first-child)]:mt-6', className)}
      ref={ref as React.ForwardedRef<HTMLParagraphElement>}
      {...props}
    />
  )
})
Typography.displayName = "Typography"

const Title = React.forwardRef<
  HTMLHeadingElement,
  React.HTMLAttributes<HTMLHeadingElement> & {
    level?: 1 | 2 | 3 | 4 | 5
  }
>(({ className, level = 1, ...props }, ref) => {
  if (level === 1) {
    return (
      <h1
        className={cn('scroll-m-20 text-4xl font-extrabold tracking-tight', className)}
        ref={ref}
        {...props}
      />
    )
  }
  
  if (level === 2) {
    return (
      <h2
        className={cn('scroll-m-20 border-b pb-2 text-3xl font-semibold tracking-tight', className)}
        ref={ref}
        {...props}
      />
    )
  }
  
  if (level === 3) {
    return (
      <h3
        className={cn('scroll-m-20 text-2xl font-semibold tracking-tight', className)}
        ref={ref}
        {...props}
      />
    )
  }
  
  if (level === 4) {
    return (
      <h4
        className={cn('scroll-m-20 text-xl font-semibold tracking-tight', className)}
        ref={ref}
        {...props}
      />
    )
  }
  
  return (
    <h5
      className={cn('text-lg font-semibold', className)}
      ref={ref}
      {...props}
    />
  )
})
Title.displayName = "Title"

const Text = React.forwardRef<
  HTMLSpanElement,
  React.HTMLAttributes<HTMLSpanElement> & {
    type?: 'secondary' | 'success' | 'warning' | 'danger'
    strong?: boolean
    code?: boolean
  }
>(({ className, type, strong, code, children, ...props }, ref) => {
  const content = strong ? <strong>{children}</strong> : children
  const codeContent = code ? <code className="relative rounded bg-muted px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold">{content}</code> : content

  return (
    <span
      className={cn(
        {
          'text-muted-foreground': type === 'secondary',
          'text-green-600 dark:text-green-400': type === 'success',
          'text-yellow-600 dark:text-yellow-400': type === 'warning',
          'text-red-600 dark:text-red-400': type === 'danger',
        },
        className
      )}
      ref={ref}
      {...props}
    >
      {codeContent}
    </span>
  )
})
Text.displayName = "Text"

export { Typography, Title, Text }