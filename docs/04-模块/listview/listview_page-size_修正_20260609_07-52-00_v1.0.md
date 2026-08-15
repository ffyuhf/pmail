# listview_page-size_修正_20260609_07-52-00_v1.0

> **版本**: v1.0  
> **日期**: 2026-06-09  
> **时间**: 07:52:00  
> **作者**: PMail Team  
> **模块**: listview  
> **状态**: 已批准  
> **生效范围**: fe/src/views/ListView.vue 前端邮件列表每页显示功能

---

## 1. 修改前后代码要点对比

### 修改文件位置（目录树）

```
fe/
└── src/
    └── views/
        └── ListView.vue    ← 唯一修改文件
```

### 修改点明细

| # | 位置（行号） | 修改前 | 修改后 | 说明 |
|---|-------------|--------|--------|------|
| 1 | 第125行 | `import {ref, watch, computed} from 'vue'` | `import {ref, watch, computed, nextTick} from 'vue'` | 新增 `nextTick` 导入，用于数据更新后重新布局 |
| 2 | 第157行后（新增） | 无 | `const currentPage = ref(1)` | 新增当前页码响应式状态 |
| 3 | 第176-181行 | `updateList` 不传 `current_page`，无 `nextTick` | `updateList` 传入 `current_page: currentPage.value`，数据加载后调用 `nextTick(() => doLayout())` | 统一分页参数，确保 el-table 重新计算布局 |
| 4 | 第112-116行（template） | `el-pagination` 无 `v-model:current-page` | 新增 `v-model:current-page="currentPage"` | 分页组件与状态双向绑定 |
| 5 | 第317-321行 | `pageChange` 直接调用 API，不更新 `currentPage` | `pageChange` 先同步 `currentPage.value = p` 再调用 `updateList()` | 统一翻页逻辑 |
| 6 | 第324-326行 | `handlePageSizeChange` 直接调用 `updateList()` | `handlePageSizeChange` 先重置 `currentPage.value = 1` 再调用 `updateList()` | 切换每页条数后重置到第 1 页 |
| 7 | 第391行（CSS） | `.mail-list__content { overflow: hidden; }` | `.mail-list__content { overflow-y: auto; }` | 允许内容超出时显示垂直滚动条 |

### 函数签名变更

无接口/函数签名变更，仅修改内部实现。

---

## 2. 变更原因与影响点

### 变更原因

用户选择每页显示数量（15/25/50/100）后，列表不会在视觉上显示对应数量的邮件。原因有两个：

1. **CSS 溢出隐藏**：`.mail-list__content` 使用 `overflow: hidden`，当数据量超过容器高度时，多余内容被裁剪，无滚动条可见。用户只有点击全选后才能发现数据确实已加载。
2. **缺少 currentPage 状态管理**：`el-pagination` 未绑定 `current-page`，改变每页条数后不会重置页码，`updateList` 也不传递 `current_page` 参数。

### 影响范围

- **前端**: 仅 `fe/src/views/ListView.vue`
- **后端**: 无变更
- **API**: 无变更

### 接口兼容性

完全兼容，无接口变更。

### 性能影响

- `nextTick + doLayout` 在每次数据加载后触发一次 el-table 重排，开销极小
- `overflow-y: auto` 浏览器原生滚动，无性能影响

### 安全影响

无安全影响。

---

## 3. 测试与验证记录

### 测试用例

| # | 测试场景 | 操作步骤 | 预期结果 | 实际结果 |
|---|----------|----------|----------|----------|
| 1 | 切换每页条数 | 选择 25 条/页 | 列表显示 25 条邮件，出现滚动条 | 待验证 |
| 2 | 切换每页条数 | 选择 50 条/页 | 列表显示 50 条邮件，出现滚动条 | 待验证 |
| 3 | 切换每页条数后翻页 | 选择 100 条/页 → 点击第 2 页 | 显示第 2 页数据 | 待验证 |
| 4 | 改变每页条数后页码重置 | 在第 3 页 → 改为 50 条/页 | 自动回到第 1 页 | 待验证 |
| 5 | 翻页功能 | 点击分页组件的上一页/下一页 | 正确切换页码和数据 | 待验证 |
| 6 | 切换分组 | 切换到不同邮件分组 | 列表正常加载 | 待验证 |

### 结论

待实际运行验证。

---

## 4. 相关文档引用

| 文档类型 | 文件路径 |
|----------|----------|
| 计划书 | `docs/02-计划/listview_page-size_修正_计划_20260609_07-51-00_v1.0.md` |
| 变更日志 | `docs/04-模块/01-变更日志/变更日志_汇总_20260609_07-27-00_v1.0.md` |
| 文档规范 | `docs/01-规范/文档生成及读取规范.md` |

---

## 变更日志

| 版本 | 日期 | 时间 | 作者 | 变更内容 | 变更原因 |
|------|------|------|------|----------|----------|
| v1.0 | 2026-06-09 | 07:52:00 | PMail Team | 初始创建 | 记录每页显示功能修复的代码变更 |