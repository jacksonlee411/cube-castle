// 临时禁用Tooltip组件，等待安装依赖
import * as React from "react"

interface TooltipProviderProps {
  children: React.ReactNode;
}

export const TooltipProvider = ({ children }: TooltipProviderProps) => {
  return <>{children}</>;
};

export const Tooltip = ({ children }: { children: React.ReactNode }) => {
  return <>{children}</>;
};

export const TooltipTrigger = ({ children, asChild, ...props }: { children: React.ReactNode, asChild?: boolean }) => {
  return <>{children}</>;
};

export const TooltipContent = ({ children }: { children: React.ReactNode }) => {
  return <div className="hidden">{children}</div>;
};