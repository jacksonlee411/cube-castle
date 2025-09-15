/**
 * 组织路径工具函数
 * - 解析 codePath/namePath（形如 "/1000000/1000001" 或 "/公司/技术部"）
 * - 组合为可用于面包屑的项列表
 */

export interface BreadcrumbItem {
  code?: string;
  name?: string;
}

export function splitPath(path?: string | null): string[] {
  if (!path) return [];
  return path.split('/').filter(Boolean);
}

/**
 * 将 codePath 与 namePath 对齐合并为面包屑项
 */
export function toBreadcrumbItems(codePath?: string | null, namePath?: string | null): BreadcrumbItem[] {
  const codes = splitPath(codePath);
  const names = splitPath(namePath);

  const maxLen = Math.max(codes.length, names.length);
  const items: BreadcrumbItem[] = [];
  for (let i = 0; i < maxLen; i++) {
    items.push({ code: codes[i], name: names[i] });
  }
  return items;
}

/**
 * 从 codePath 获取父级编码链
 */
export function toParentChainFromCodePath(codePath?: string | null): string[] {
  return splitPath(codePath);
}

