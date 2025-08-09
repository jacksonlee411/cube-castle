# Page snapshot

```yaml
- dialog "新增组织单元":
  - button "关闭"
  - heading "新增组织单元" [level=2]
  - text: 组织编码
  - textbox "组织编码" [disabled]
  - paragraph: 系统将自动生成唯一编码
  - text: 组织名称 *
  - textbox "组织名称 *"
  - text: 组织类型 *
  - combobox:
    - option "部门" [selected]
    - option "成本中心"
    - option "公司"
    - option "项目团队"
  - text: 上级组织编码
  - textbox "上级组织编码"
  - text: 组织层级 *
  - spinbutton "组织层级 *": "1"
  - text: 状态 *
  - combobox:
    - option "激活" [selected]
    - option "停用"
    - option "计划中"
  - text: 排序
  - spinbutton "排序": "0"
  - text: 描述
  - textbox "描述"
  - button "取消"
  - button "创建"
```