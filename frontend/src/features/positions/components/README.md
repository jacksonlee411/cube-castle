# positions/components 目录结构

```
components/
├── dashboard/           # 仪表盘与统计组件
├── details/             # 职位详情页卡片与面板
├── form/                # PositionForm 及其子模块
├── layout/              # 通用布局辅助（SimpleStack）
├── list/                # 列表视图组件
├── transfer/            # 职位调动对话框
├── versioning/          # 版本列表、工具栏等
└── index.ts             # 聚合导出
```

## 快速引用

```ts
import { PositionList, PositionHeadcountDashboard, SimpleStack } from '@/features/positions/components';
```

各子目录均提供 `index.ts` 导出，避免直接引用具体文件路径。

## 注意事项

- 新增组件时请按功能选择子目录，并在对应 `index.ts` 中导出。
- 共用工具（如 `SimpleStack`）放在 `layout/`，避免跨目录重复。
- 废弃旧版 `PositionVersionList.tsx`，统一使用 `components/versioning/VersionList.tsx`。
