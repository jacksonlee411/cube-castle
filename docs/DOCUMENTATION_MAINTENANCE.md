# 文档维护指南 (Documentation Maintenance Guidelines)

## 📋 目的 (Purpose)
本文档建立了Cube Castle项目中维护文档质量和防止内容冗余的指南规范。
*This document establishes guidelines for maintaining documentation quality and preventing redundancy in the Cube Castle project.*

## 🗂️ 目录结构 (Directory Structure)

### 主要文档结构 (Primary Documentation Structure)
```
docs/
├── api/             # API规范和生成文档 (API specifications and generated docs)
├── architecture/    # 系统设计和技术架构 (System design and technical architecture)
├── deployment/      # 部署指南和配置 (Deployment guides and configuration)
├── reports/         # 进展报告和测试结果 (Progress reports and test results)
└── troubleshooting/ # 问题解决指南和最佳实践 (Problem-solving guides and best practices)
```

### 附加文档位置 (Additional Documentation Locations)
```
tests/              # 测试文档与测试文件并存 (Test documentation alongside test files)
scripts/            # 脚本文档与自动化并存 (Script documentation alongside automation)
README.md           # 项目概述和快速开始 (Project overview and quick start)
CHANGELOG.md        # 版本历史和变更 (Version history and changes)
```

## 📝 命名约定 (Naming Conventions)

### 文件命名规则 (File Naming Rules)
1. **仅使用英文名称** (*Use English names only*) - 文件名中不包含中文字符 (No Chinese characters in filenames)
2. **使用snake_case** (*Use snake_case*) - `employee_model_design.md` ✅
3. **描述性命名** (*Be descriptive*) - `user_guide.md` ✅ vs `guide.md` ❌
4. **报告包含日期** (*Include date for reports*) - `test_report_20250729_143500.md` ✅
5. **类型前缀清晰** (*Prefix with type for clarity*) - `api_employee_endpoints.md` ✅

### 目录组织 (Directory Organization)
- **按功能分组，不按时间** (*By function, not by time*) - 将相关内容分组在一起 (Group related content together)
- **避免深层嵌套** (*Avoid deep nesting*) - 最多2-3层深度 (Maximum 2-3 levels deep)
- **使用一致命名** (*Use consistent naming*) - 所有目录名小写 (All directory names in lowercase)

## 🔄 维护工作流程 (Maintenance Workflow)

### 创建新文档前 (Before Creating New Documentation)
1. **检查现有文档** (*Check existing docs*) - 首先搜索类似内容 (Search for similar content first)
2. **使用适当位置** (*Use appropriate location*) - 遵循目录结构指南 (Follow directory structure guidelines)
3. **遵循命名约定** (*Follow naming conventions*) - 使用标准化命名模式 (Use standardized naming patterns)
4. **链接相关文档** (*Link related documents*) - 在有用的地方创建交叉引用 (Create cross-references where useful)

### 定期维护任务 (每月) (Regular Maintenance Tasks - Monthly)
1. **删除过时报告** (*Remove outdated reports*) - 归档6个月以上的报告 (Archive reports older than 6 months)
2. **合并相似内容** (*Consolidate similar content*) - 合并重复或重叠的文档 (Merge duplicate or overlapping docs)
3. **更新交叉引用** (*Update cross-references*) - 确保所有链接保持有效 (Ensure all links remain valid)
4. **标准化命名** (*Standardize naming*) - 重命名不遵循约定的文件 (Rename files that don't follow conventions)

### 质量标准 (Quality Standards)
- **每个文档一个主题** (*One topic per document*) - 避免混合不相关的主题 (Avoid mixing unrelated subjects)
- **清晰结构** (*Clear structure*) - 一致使用标题、列表和格式 (Use headers, lists, and formatting consistently)
- **双语描述性内容** (*Bilingual descriptive content*) - 新增文档的描述性内容必须提供中英文双语说明 (New documents must provide bilingual Chinese-English explanations for descriptive content)
- **更新时间戳** (*Update timestamps*) - 在文档头部包含最后修改日期时间 (Include last modified datetime in document headers)
- **版本信息** (*Version information*) - 适用时引用特定版本 (Reference specific versions when applicable)

## 🌐 双语内容指南 (Bilingual Content Guidelines)

### 双语要求范围 (Bilingual Requirements Scope)
**适用内容** (*Content Types That Require Bilingual Treatment*):
- **标题和子标题** (*Titles and subtitles*) - 所有主要标题应提供中英文版本 (All major headings should provide Chinese-English versions)
- **概述和总结** (*Overviews and summaries*) - 文档概述必须双语呈现 (Document overviews must be presented bilingually)
- **业务流程描述** (*Business process descriptions*) - 业务逻辑和流程说明需要双语 (Business logic and process explanations require bilingual treatment)
- **用户指导说明** (*User guidance instructions*) - 操作步骤和指导信息双语呈现 (Operational steps and guidance information presented bilingually)
- **错误信息和警告** (*Error messages and warnings*) - 重要的错误和警告信息需要双语 (Important error and warning messages require bilingual presentation)

**豁免内容** (*Exempt Content Types*):
- **代码示例** (*Code examples*) - 代码本身保持英文，但注释可以双语 (Code remains in English, but comments can be bilingual)
- **技术规格** (*Technical specifications*) - API规格、数据结构等技术细节 (API specs, data structures, and other technical details)
- **外部引用** (*External references*) - 第三方文档和链接 (Third-party documentation and links)

### 双语格式规范 (Bilingual Format Standards)

#### 标题格式 (Title Format)
```markdown
# English Title | 中文标题
## English Subtitle | 中文副标题
```

#### 段落格式 (Paragraph Format)
```markdown
English description of the concept or process.

中文概念或流程描述。
```

#### 列表项格式 (List Item Format)
```markdown
- **English Item** (*English explanation*) - 中文解释 (Chinese explanation)
```

#### 代码注释格式 (Code Comment Format)
```go
// English comment | 中文注释
// Process employee assignment | 处理员工分配
func AssignEmployee() {
    // Implementation | 实现
}
```

### 实施优先级 (Implementation Priority)

#### 高优先级 (High Priority)
1. **新文档创建** (*New document creation*) - 所有新文档必须遵循双语要求 (All new documents must follow bilingual requirements)
2. **面向用户的内容** (*User-facing content*) - API文档、用户指南等 (API documentation, user guides, etc.)
3. **业务流程文档** (*Business process documentation*) - 工作流程、业务规则说明 (Workflows, business rule explanations)

#### 中优先级 (Medium Priority)
1. **现有重要文档更新** (*Updates to existing important documents*) - 架构文档、设计文档 (Architecture docs, design documents)
2. **报告和总结** (*Reports and summaries*) - 项目报告、实施总结 (Project reports, implementation summaries)

#### 低优先级 (Low Priority)
1. **内部技术文档** (*Internal technical documentation*) - 开发者内部文档 (Internal developer documentation)
2. **临时性文档** (*Temporary documentation*) - 会议记录、临时说明 (Meeting notes, temporary instructions)

## 🚫 避免事项 (What to Avoid)

### 文件管理反模式 (File Management Anti-Patterns)
- ❌ **中文文件名** (*Chinese filenames*) - 始终使用英文 (Always use English)
- ❌ **重复内容** (*Duplicate content*) - 每个主题一个真实来源 (One source of truth per topic)
- ❌ **仓库中的临时文件** (*Temporary files in repo*) - 使用适当的临时目录 (Use proper temporary directories)
- ❌ **混合命名风格** (*Mixed naming styles*) - 保持一致 (Be consistent)
- ❌ **深层目录嵌套** (*Deep directory nesting*) - 保持结构扁平和逻辑性 (Keep structure flat and logical)

### 内容反模式 (Content Anti-Patterns)
- ❌ **过时信息** (*Outdated information*) - 删除或更新过时内容 (Remove or update obsolete content)
- ❌ **个人笔记** (*Personal notes*) - 将个人笔记排除在共享文档之外 (Keep individual notes out of shared docs)
- ❌ **不完整文档** (*Incomplete documents*) - 提交前完成文档 (Finish documents before committing)
- ❌ **损坏链接** (*Broken links*) - 测试所有内部和外部引用 (Test all internal and external references)

## 🎯 实施检查清单 (Implementation Checklist)

### 新文档检查 (For New Documentation)
- [ ] 检查现有类似内容 (Check for existing similar content)
- [ ] 选择适当的目录位置 (Choose appropriate directory location)
- [ ] 遵循命名约定 (Follow naming conventions)
- [ ] 包含清晰标题和目的 (Include clear title and purpose)
- [ ] 确保描述性内容提供中英文双语说明 (Ensure descriptive content provides bilingual Chinese-English explanations)
- [ ] 添加最后更新日期时间 (Add last updated datetime)
- [ ] 链接到相关文档 (Link to related documents)
- [ ] 审核完整性 (Review for completeness)

### 维护审核检查 (For Maintenance Reviews)
- [ ] 识别并删除重复内容 (Identify and remove duplicate content)
- [ ] 标准化文件命名 (Standardize file naming)
- [ ] 组织到适当目录 (Organize into appropriate directories)
- [ ] 更新交叉引用 (Update cross-references)
- [ ] 归档过时材料 (Archive outdated materials)
- [ ] 验证所有链接工作 (Verify all links work)

## 📊 成功指标 (Success Metrics)

- **文件命名合规性** (*File Naming Compliance*): 100%英文文件名 (100% English filenames)
- **目录组织** (*Directory Organization*): 所有文档在适当类别中 (All docs in appropriate categories)
- **双语内容合规性** (*Bilingual Content Compliance*): 新文档100%提供双语描述性内容 (100% of new documents provide bilingual descriptive content)
- **内容新鲜度** (*Content Freshness*): 没有超过1年未审核的文档 (No docs older than 1 year without review)
- **交叉引用准确性** (*Cross-Reference Accuracy*): 所有内部链接功能正常 (All internal links functional)
- **重复率** (*Duplication Rate*): 文档间零重复内容 (Zero duplicate content across docs)

## 🔄 审核计划 (Review Schedule)

- **每周** (*Weekly*): 检查新文档合规性 (Check new documentation for compliance)
- **每月** (*Monthly*): 审核和清理文档结构 (Review and clean up documentation structure)
- **每季度** (*Quarterly*): 归档旧报告和更新交叉引用 (Archive old reports and update cross-references)
- **每年** (*Yearly*): 完整文档审计和重组 (Complete documentation audit and reorganization)

---

**最后更新** (*Last Updated*): 2025-07-31 15:45:00  
**下次审核** (*Next Review*): 2025-08-31 15:45:00  
**更新内容** (*Update Summary*): 增加双语内容指南和要求 (Added bilingual content guidelines and requirements)