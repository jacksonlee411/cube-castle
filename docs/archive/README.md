# 📦 文档归档区（Archive）

用途：集中存放历史/完成/废弃类文档，保持主文档区精简与权威。

## 📂 目录结构

```
docs/archive/
├── development-plans/              # 开发计划归档（已完成/历史）
├── deprecated-neo4j-era/          # 旧架构（Neo4j/CDC 等）废弃资料
├── deprecated-api-specs/          # 废弃的 API 契约/规范
├── deprecated-api-design/         # 废弃的 API 设计草案/方案
├── deprecated-guides/             # 废弃的开发指南/说明
├── deprecated-notes/              # 废弃的笔记/零散说明
├── deprecated-setup/              # 废弃的安装/环境脚本与文档
├── project-reports/               # 项目报告归档（审计/总结/专项报告）
└── frontend-ux-optimization-deprecated/  # 废弃的前端 UX 优化方案
```

## 📥 归档标准（摘要）
- 开发计划/路线/进展文档：完成后移动至 `archive/development-plans/`。
- 废弃/过时资料：移动至对应 `deprecated-*/` 子目录。
- 报告类文档：阶段性或最终报告归档到 `project-reports/`。
- 临时/测试文件：任务结束后移出主目录并归档或删除。

详细规则见：`docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`（目录结构、命名、审计清单）。

## 🔗 相关入口
- 主导航：`docs/README.md`
- 活跃计划：`docs/development-plans/`
- 参考文档：`docs/reference/`
- 文档治理规范：`docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`

## 🛠️ 维护约定
- 归档内容默认只读：仅允许修复链接/明显错字；内容更新应在活跃目录创建新文档。
- 月度审计：确认需要归档的计划/报告是否已移动；清理滞留的临时文件。

—— 最后更新：2025-09-13 | 维护：Cube Castle 文档维护团队

