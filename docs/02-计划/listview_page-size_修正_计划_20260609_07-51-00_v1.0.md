# listview_page-size_修正_计划_20260609_07-51-00_v1.0

> **版本**: v1.0  
> **日期**: 2026-06-09  
> **时间**: 07:51:00  
> **作者**: PMail Team  
> **模块**: listview  
> **状态**: 已批准  
> **生效范围**: fe/src/views/ListView.vue 前端邮件列表分页显示功能

---

## 1. 目标与变更范围

### 要改什么

- 修复 `fe/src/views/ListView.vue` 中每页显示功能（page-size select）不生效的问题
- 用户选择每页显示数量（15/25/50/100）后，列表应立即显示对应数量的邮件
- 修复 el-table 容器溢出隐藏导致超出内容不可见的问题
- 统一分页状态管理，确保 currentPage 与 el-pagination 同步

### 不改什么

- 不修改后端 API（`server/controllers/email/list.go` 分页逻辑已正确）
- 不修改 `fe/src/services/emailService.ts` 接口定义
- 不修改其他页面或组件

---

## 2. 影响评估与依赖

### 影响点

| 影响范围 | 说明 | 影响程度 |
|----------|------|----------|
| ListView.vue CSS | `.mail-list__content` 的 overflow 属性变更 | 低 |
| ListView.vue script | 新增 `currentPage` ref，修改 `updateList`、`pageChange`、`handlePageSizeChange` | 低 |
| ListView.vue template | `el-pagination` 绑定 `v-model:current-page` | 低 |

### 风险评估

- **风险等级**: 低
- **影响文件数**: 1（仅 `fe/src/views/ListView.vue`）
- **向后兼容性**: 完全兼容，无接口变更

### 依赖关系

```
ListView.vue → emailService.getEmailList() → POST /api/email/list
```

后端 API 已支持 `page_size` 和 `current_page` 参数，无需后端变更。

---

## 3. 实施步骤与检查点

- [x] 步骤1：创建本计划书文档
- [ ] 步骤2：修复 CSS 溢出 — `.mail-list__content` 的 `overflow: hidden` → `overflow-y: auto`
- [ ] 步骤3：添加 `currentPage` ref 并绑定到 `el-pagination`
- [ ] 步骤4：统一 `updateList` 和 `pageChange` 的分页逻辑，确保都传递 `current_page`
- [ ] 步骤5：数据更新后通过 `nextTick` 调用 `doLayout` 确保 el-table 重新布局
- [ ] 步骤6：创建代码修改文档
- [ ] 步骤7：更新变更日志汇总

### 验收条件

1. 选择每页显示 25/50/100 后，列表立即显示对应数量的邮件
2. 改变每页条数后，分页组件自动重置到第 1 页
3. 翻页功能正常工作
4. 列表内容超出容器时出现滚动条

---

## 4. 回滚方案与失败判定

### 回滚方案

恢复 `fe/src/views/ListView.vue` 的以下变更：
1. CSS: `overflow-y: auto` → `overflow: hidden`
2. 删除 `currentPage` ref 及相关绑定
3. 恢复 `updateList` 和 `pageChange` 的原始逻辑

### 失败判定条件

- 修改后列表完全不显示数据
- 分页组件无法翻页
- 选择每页条数后页面报错
- 滚动条异常导致布局错乱

---

## 5. 所需文档清单与模板引用

| 文档类型 | 文件路径 | 状态 |
|----------|----------|------|
| 计划书 | `docs/02-计划/listview_page-size_修正_计划_20260609_07-51-00_v1.0.md` | 本文档 |
| 代码修改文档 | `docs/04-模块/listview/listview_page-size_修正_20260609_07-51-00_v1.0.md` | 待创建 |
| 变更日志 | `docs/04-模块/01-变更日志/变更日志_汇总_20260609_07-27-00_v1.0.md` | 待更新 |

---

## 6. 审批与共识要求

- **审批人**: PMail Team
- **审批状态**: 已批准（自审自批，小范围前端修复）

---

## 变更日志

| 版本 | 日期 | 时间 | 作者 | 变更内容 | 变更原因 |
|------|------|------|------|----------|----------|
| v1.0 | 2026-06-09 | 07:51:00 | PMail Team | 初始创建 | 修复每页显示功能不生效问题 |