'use client'

import { useState } from 'react'
import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { 
  Home, 
  Users, 
  Building2, 
  Brain, 
  Settings, 
  LogOut,
  Menu,
  X,
  Castle
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const navigationItems = [
  {
    title: '概览',
    href: '/dashboard',
    icon: Home,
  },
  {
    title: '员工管理',
    href: '/employees',
    icon: Users,
  },
  {
    title: '组织架构',
    href: '/organizations',
    icon: Building2,
  },
  {
    title: 'AI 助手',
    href: '/chat',
    icon: Brain,
  },
  {
    title: '系统设置',
    href: '/settings',
    icon: Settings,
  },
]

interface AppNavProps {
  className?: string
}

export function AppNav({ className }: AppNavProps) {
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false)
  const pathname = usePathname()

  return (
    <nav className={cn('border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60', className)}>
      <div className="container-responsive flex h-16 items-center justify-between">
        {/* Logo */}
        <Link href="/dashboard" className="flex items-center space-x-2">
          <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
            <Castle className="h-5 w-5" />
          </div>
          <span className="text-xl font-bold">Cube Castle</span>
        </Link>

        {/* Desktop Navigation */}
        <div className="hidden md:flex items-center space-x-6">
          {navigationItems.map((item) => {
            const isActive = pathname === item.href
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  'flex items-center space-x-2 text-sm font-medium transition-colors hover:text-primary',
                  isActive 
                    ? 'text-primary' 
                    : 'text-muted-foreground'
                )}
              >
                <item.icon className="h-4 w-4" />
                <span>{item.title}</span>
              </Link>
            )
          })}
        </div>

        {/* User Menu */}
        <div className="hidden md:flex items-center space-x-4">
          <Button variant="ghost" size="sm">
            <LogOut className="mr-2 h-4 w-4" />
            退出
          </Button>
        </div>

        {/* Mobile Menu Button */}
        <Button
          variant="ghost"
          size="sm"
          className="md:hidden"
          onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
        >
          {isMobileMenuOpen ? (
            <X className="h-5 w-5" />
          ) : (
            <Menu className="h-5 w-5" />
          )}
        </Button>
      </div>

      {/* Mobile Navigation */}
      {isMobileMenuOpen && (
        <div className="border-t bg-background md:hidden">
          <div className="container-responsive py-4">
            <div className="space-y-3">
              {navigationItems.map((item) => {
                const isActive = pathname === item.href
                return (
                  <Link
                    key={item.href}
                    href={item.href}
                    className={cn(
                      'flex items-center space-x-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                      isActive 
                        ? 'bg-primary text-primary-foreground' 
                        : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                    )}
                    onClick={() => setIsMobileMenuOpen(false)}
                  >
                    <item.icon className="h-4 w-4" />
                    <span>{item.title}</span>
                  </Link>
                )
              })}
              <div className="border-t pt-3">
                <Button variant="ghost" size="sm" className="w-full justify-start">
                  <LogOut className="mr-2 h-4 w-4" />
                  退出
                </Button>
              </div>
            </div>
          </div>
        </div>
      )}
    </nav>
  )
}