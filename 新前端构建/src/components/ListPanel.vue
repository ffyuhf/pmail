<!--
  ListPanel 邮件列表面板组件

  中间 340px 宽的邮件列表面板（对齐「新前端实现标准.html」），提供：
  - 文件夹标题（24px 粗体，对齐标准 .list-title）
  - 搜索框（focus 阴影效果）
  - 分组导航标签（功能增强，保留）
  - 邮件列表（带圆形头像、未读增强、选中左侧竖条、滚动加载）
  - 批量操作工具栏（功能增强，保留）

  创建日期: 20260609
  修改日期: 20260610
  修改原因: 对齐标准 HTML 视觉 —— 添加标题、邮件头像、未读增强、选中竖条、搜索框 focus 阴影
-->
<template>
  <div class="list-panel">
    <!-- 标题区（对齐标准 .list-header + .list-title） -->
    <div class="list-header">
      <h1 class="list-title">{{ currentFolderLabel }}</h1>
      <input
        v-model="searchText"
        class="search-bar"
        :placeholder="lang.search"
        type="text"
        @keyup.enter="emit('search', searchText)"
      />
    </div>

    <!-- 分组标签（横向滚动，功能增强保留） -->
    <div class="group-tabs">
      <button
        v-for="group in groups"
        :key="group.tag"
        class="group-tab"
        :class="{ active: currentTag === group.tag }"
        @click="selectGroup(group.tag)"
      >
        {{ getFolderLabel(group) }}
      </button>
    </div>

    <!-- 批量操作工具栏（选中时显示，功能增强保留） -->
    <div v-if="selectedIds.length > 0" class="batch-bar">
      <label class="batch-check">
        <input type="checkbox" :checked="allSelected" @change="toggleSelectAll" />
        <span>{{ lang.select_all }}</span>
      </label>
      <span class="batch-count">{{ selectedIds.length }} {{ lang.selected_count }}</span>
      <button class="batch-btn" @click="$emit('batch-read')">{{ lang.read_btn }}</button>
      <button class="batch-btn danger" @click="$emit('batch-delete')">{{ lang.del_btn }}</button>
      <button class="batch-btn" @click="$emit('batch-move')">{{ lang.move_btn }}</button>
    </div>

    <!-- 邮件列表（对齐标准 .mail-list-container + .mail-item） -->
    <div class="mail-list-container" @scroll="onScroll">
      <!-- 列表项交错进入动画（对齐标准 animationDelay: idx * 0.05s） -->
      <div
        v-for="(email, idx) in emails"
        :key="email.id"
        class="mail-item"
        :style="{ animationDelay: idx * 0.05 + 's' }"
        :class="{
          active: activeEmailId === email.id,
          unread: !email.is_read,
          selected: selectedIds.includes(email.id),
        }"
        @click="$emit('select-email', email)"
      >
        <!-- 多选框（功能增强保留） -->
        <label class="mail-check" @click.stop>
          <input
            type="checkbox"
            :checked="selectedIds.includes(email.id)"
            @change="toggleSelect(email.id)"
          />
        </label>
        <!-- 邮件头像（对齐标准 .mail-avatar：圆形 44×44，显示发件人首字母） -->
        <div class="mail-avatar">{{ getSenderInitial(email) }}</div>
        <!-- 邮件预览（对齐标准 .mail-preview） -->
        <div class="mail-preview">
          <div class="mail-header-row">
            <span class="mail-sender">
              {{ email.type === 1 ? lang.sender_desc + ' → ' + email.sender.Name : email.sender.Name }}
            </span>
            <span class="mail-time">{{ formatShortDate(email.datetime) }}</span>
          </div>
          <div class="mail-subject">{{ email.title }}</div>
          <div class="mail-snippet">{{ email.desc }}</div>
        </div>
      </div>

      <!-- 加载更多 -->
      <div v-if="loading" class="loading-more">{{ lang.wait_desc }}</div>
      <div v-if="!loading && emails.length === 0" class="empty-list">—</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import lang from '@/i18n'
import { ref, computed } from 'vue'
import type { EmailListItem } from '@/types/api'
import { formatShortDate } from '@/utils/dateFormat'
import { getDefaultFolderI18nKey } from '@/utils/constants'

const props = defineProps<{
  emails: EmailListItem[]
  activeEmailId: number | null
  selectedIds: number[]
  loading: boolean
  currentTag: string
  groups: { tag: string; label: string }[]
  allSelected: boolean
}>()

const emit = defineEmits<{
  'select-email': [email: EmailListItem]
  'select-group': [tag: string]
  'toggle-select': [id: number]
  'toggle-select-all': []
  'batch-delete': []
  'batch-read': []
  'batch-move': []
  'load-more': []
  'search': [query: string]
}>()

const searchText = ref('')

/**
 * 当前文件夹显示名称（用于列表标题）
 *
 * 从 groups 中查找当前 tag 对应的 label，
 * 默认文件夹使用 i18n 翻译。
 */
const currentFolderLabel = computed(() => {
  const current = props.groups.find((g) => g.tag === props.currentTag)
  if (current) return getFolderLabel(current)
  /* 默认显示收件箱 */
  return lang.inbox || '收件箱'
})

/**
 * 获取文件夹显示名称
 *
 * 默认文件夹使用前端 i18n 翻译替换后端返回的 label，
 * 用户自定义文件夹保持原始名称不变。
 */
function getFolderLabel(group: { tag: string; label: string }): string {
  const i18nKey = getDefaultFolderI18nKey(group.tag)
  if (i18nKey) {
    return (lang as Record<string, string>)[i18nKey] || group.label
  }
  return group.label
}

/**
 * 获取发件人头像首字母
 *
 * 对齐标准 HTML 中 .mail-avatar 显示逻辑：
 * 取 sender.Name 的第一个字符（大写），无名称时显示 '?'
 *
 * @param email - 邮件列表项
 * @returns 头像显示的首字母
 */
function getSenderInitial(email: EmailListItem): string {
  const name = email.sender?.Name || ''
  if (!name) return '?'
  return name.charAt(0).toUpperCase()
}

function selectGroup(tag: string) {
  emit('select-group', tag)
}

function toggleSelect(id: number) {
  emit('toggle-select', id)
}

function toggleSelectAll() {
  emit('toggle-select-all')
}

function onScroll(e: Event) {
  const el = e.target as HTMLElement
  if (el.scrollTop + el.clientHeight >= el.scrollHeight - 50) {
    emit('load-more')
  }
}
</script>

<style scoped>
/* 列表面板（对齐标准 .list-panel） */
.list-panel {
  width: var(--list-panel-width);
  height: 100%;
  background: var(--bg-color);
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

/* 标题区（对齐标准 .list-head: padding 16px） */
.list-header {
  padding: 16px 16px 12px;
}

/* 文件夹标题（对齐标准 .list-title: 16px 粗体, margin-bottom:10px） */
.list-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 10px;
}

/* 搜索框（对齐标准 .search-input: padding 8px 12px, font-size:11px, radius:6px） */
.search-bar {
  width: 100%;
  padding: 8px 12px;
  background: #f5f5f7;
  border: 1px solid transparent;
  border-radius: 6px;
  font-size: 11px;
  outline: none;
  color: #333;
  transition: all 0.3s ease;
}

.search-bar::placeholder {
  color: var(--text-placeholder);
}

/* focus 状态（对齐标准: 白底 + 边框 + box-shadow 3px） */
.search-bar:focus {
  background: #fff;
  border-color: #d1d1d6;
  box-shadow: 0 0 0 3px rgba(0, 0, 0, 0.05);
}

/* 分组标签（功能增强，保留；padding 对齐标准 16px） */
.group-tabs {
  display: flex;
  gap: 4px;
  padding: 8px 16px;
  overflow-x: auto;
  border-bottom: 1px solid var(--border-color);
}

.group-tab {
  padding: 4px 10px;
  border: none;
  background: transparent;
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  cursor: pointer;
  transition: all var(--transition);
}

.group-tab:hover {
  background: var(--bg-hover);
}

.group-tab.active {
  background: var(--accent-color);
  color: #fff;
}

/* 批量操作工具栏（功能增强，保留；padding 对齐标准 16px） */
.batch-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
}

.batch-check {
  display: flex;
  align-items: center;
  gap: 4px;
  cursor: pointer;
}

.batch-count {
  color: var(--text-secondary);
}

.batch-btn {
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  background: var(--bg-color);
  border-radius: var(--radius-sm);
  font-size: 11px;
  cursor: pointer;
  transition: all var(--transition);
}

.batch-btn:hover {
  background: var(--bg-hover);
}

.batch-btn.danger {
  color: var(--danger-color);
  border-color: var(--danger-color);
}

/* 邮件列表容器（对齐标准 .mail-list-container） */
.mail-list-container {
  flex: 1;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: transparent transparent;
}

.mail-list-container:hover {
  scrollbar-color: var(--border-color) transparent;
}

/* 邮件项（对齐标准 .mail-item: padding 12px 16px, gap 10px, fadeSlideIn 动画） */
.mail-item {
  padding: 12px 16px;
  border-bottom: 1px solid #f5f5f5;
  cursor: pointer;
  transition: background 0.2s ease, transform 0.15s ease;
  display: flex;
  gap: 10px;
  position: relative;
  animation: fadeSlideIn 0.4s ease forwards;
  opacity: 0;
}

/* 列表项交错进入动画（对齐标准 @keyframes fadeSlideIn） */
@keyframes fadeSlideIn {
  from { opacity: 0; transform: translateX(-10px); }
  to { opacity: 1; transform: translateX(0); }
}

/* 选中左侧竖条指示器（对齐标准 .mail-item::before: width 0→3px） */
.mail-item::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 0;
  background: #1d1d1f;
  transition: width 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

/* hover 平移效果（对齐标准 .mail-item:hover: translateX(2px)） */
.mail-item:hover {
  background: #fafafa;
  transform: translateX(2px);
}

.mail-item.active {
  background: #f5f5f7;
}

/* 选中时显示左侧 3px 黑色竖条（对齐标准 .mail-item.active::before: width 3px） */
.mail-item.active::before {
  width: 3px;
}

/* 多选框（功能增强保留，缩小不干扰头像） */
.mail-check {
  flex-shrink: 0;
  display: flex;
  align-items: flex-start;
  padding-top: 12px;
  cursor: pointer;
}

/* 邮件头像（对齐标准 .m-avatar: 32×32, border-radius:6px, font-size:10px） */
.mail-avatar {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: #e5e5e5;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-weight: 700;
  color: #666;
  font-size: 10px;
  transition: background 0.3s ease, color 0.3s ease;
}

/* 邮件预览区（对齐标准 .mail-preview） */
.mail-preview {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

/* 发件人/时间行（对齐标准 .m-row: margin-bottom 2px） */
.mail-header-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2px;
  align-items: center;
}

/* 发件人名称（对齐标准 .m-sender: 12px, font-weight:500） */
.mail-sender {
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 时间（对齐标准 .m-time: 10px 灰色） */
.mail-time {
  font-size: 10px;
  color: #8e8e93;
  flex-shrink: 0;
  margin-left: 8px;
}

/* 主题（对齐标准 .m-subject: 11px） */
.mail-subject {
  font-size: 11px;
  margin-bottom: 2px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 摘要（对齐标准 .m-snippet: 10px 灰色） */
.mail-snippet {
  font-size: 10px;
  color: #8e8e93;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 未读增强（对齐标准 .unread: 发件人 font-weight:700, 主题 font-weight:600, 头像黑底白字） */
.mail-item.unread .mail-sender {
  font-weight: 700;
}

.mail-item.unread .mail-subject {
  font-weight: 600;
}

.mail-item.unread .mail-avatar {
  background: #1d1d1f;
  color: #fff;
}

/* 选中高亮（功能增强保留） */
.mail-item.selected {
  background: rgba(0, 0, 0, 0.04);
}

/* 加载/空状态 */
.loading-more,
.empty-list {
  padding: 24px;
  text-align: center;
  color: var(--text-placeholder);
  font-size: 13px;
}
</style>