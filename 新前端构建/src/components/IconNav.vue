<!--
  IconNav 图标导航栏组件

  左侧 68px 宽的图标导航栏（对齐「新前端实现标准.html」），提供：
  - nav-top：撰写按钮（黑色圆形）+ 分组快捷图标（圆形按钮）
  - nav-bottom：设置按钮 + 用户头像（首字母圆形）

  创建日期: 20260609
  修改日期: 20260610
  修改原因: 对齐标准 HTML 视觉样式 —— 按钮改圆形、compose 黑色阴影、添加用户头像
-->
<template>
  <nav class="icon-nav">
    <!-- 上半区：撰写 + 分组快捷图标 -->
    <div class="nav-top">
      <!-- 撰写按钮（黑色圆形，对齐标准 .compose-btn） -->
      <!-- 对齐标准 HTML: compose 按钮 + 符号 -->
      <button class="nav-btn compose-btn" :title="lang.compose" @click="$emit('compose')">
        <span class="nav-icon">+</span>
      </button>
      <!-- 分组快捷图标（圆形按钮，对齐标准 .nav-btn.active） -->
      <button
        v-for="group in quickGroups"
        :key="group.tag"
        class="nav-btn"
        :class="{ active: currentTag === group.tag }"
        :title="group.label"
        @click="$emit('select-group', group.tag)"
      >
        <span class="nav-icon">{{ group.icon }}</span>
      </button>
    </div>

    <!-- 下半区：设置 + 注销 + 用户头像 -->
    <div class="nav-bottom">
      <button class="nav-btn" :title="lang.settings" @click="$emit('open-settings')">
        <span class="nav-icon">⚙</span>
      </button>
      <button class="nav-btn" :title="lang.logout" @click="$emit('logout')">
        <span class="nav-icon">⏻</span>
      </button>
      <!-- 用户头像（圆形，显示首字母，对齐标准 .user-avatar） -->
      <div class="user-avatar" :title="userName">{{ userInitial }}</div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import lang from '@/i18n'
import { computed } from 'vue'
import { useGroupStore } from '@/stores/group'
import { useGlobalStore } from '@/stores/global'
import { DEFAULT_FOLDERS, isDefaultGroupTag } from '@/utils/constants'

defineEmits<{
  'select-group': [tag: string]
  'compose': []
  'open-settings': []
  'logout': []
}>()

const groupStore = useGroupStore()
const globalStore = useGlobalStore()
const currentTag = computed(() => groupStore.currentTag)

/**
 * 用户名称（优先使用 name，否则 account）
 * 用于用户头像的 title 属性
 */
const userName = computed(() => globalStore.userInfo?.name || globalStore.userInfo?.account || '')

/**
 * 用户头像首字母（取 name 或 account 的第一个字符大写）
 * 对齐标准 HTML 中 .user-avatar 显示逻辑
 */
const userInitial = computed(() => {
  const name = userName.value
  return name ? name.charAt(0).toUpperCase() : '?'
})

/**
 * 快捷分组列表（计算属性）
 *
 * 修复逻辑：
 * 1. 硬编码显示全部 5 个默认文件夹（收件箱/发件箱/草稿箱/垃圾箱/已删除），
 *    使用前端 i18n 替换显示名称，实现国际化
 * 2. 从 flatGroups 中过滤掉所有默认文件夹（通过 isDefaultGroupTag 匹配），
 *    仅显示用户自定义文件夹，消除重复
 *
 * tag 匹配规则：解析 JSON 后比较 type+status，忽略 group_id 等额外字段。
 * 这样无论后端返回的 tag 是否包含 group_id，都能正确匹配。
 */
const quickGroups = computed(() => {
  /* 默认文件夹：使用 i18n 名称和专属图标 */
  const defaults = DEFAULT_FOLDERS.map((f) => ({
    tag: JSON.stringify({ type: f.type, status: f.status }),
    label: (lang as Record<string, string>)[f.i18nKey] || f.i18nKey,
    icon: f.icon,
  }))
  /*
   * 自定义文件夹：从 flatGroups 中排除所有默认文件夹
   *
   * flatGroups 已跳过无 tag 的容器节点（如"All Email"），
   * 此处额外防御性过滤 g.tag 非空，确保不会出现无 tag 项。
   *
   * 修改日期: 20260609
   * 修改原因: 防御性检查，确保无 tag 的容器节点不会出现在导航栏中
   */
  const customs = groupStore.flatGroups
    .filter((g) => g.tag && !isDefaultGroupTag(g.tag))
    .map((g) => ({ tag: g.tag, label: g.label, icon: '📁' }))
  return [...defaults, ...customs]
})
</script>

<style scoped>
/* 图标导航栏（对齐标准 .nav-rail: 56px 宽, padding-top: 16px） */
.icon-nav {
  width: var(--icon-nav-width);
  height: 100%;
  background: #fafafa;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 16px 0;
  justify-content: space-between;
  z-index: 10;
}

/* 上半区（对齐标准 .nav-top: gap 4px） */
.nav-top {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  gap: 4px;
}

/* 下半区（对齐标准 .nav-bottom: gap 4px） */
.nav-bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  gap: 4px;
}

/* 导航按钮（对齐标准 .icon-btn: 36×36, border-radius:8px 圆角方形） */
.nav-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  border: none;
  background: transparent;
  color: #8e8e93;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.nav-btn:hover {
  background: #f0f0f0;
  color: #1d1d1f;
}

/* 点击缩放反馈（对齐标准 .icon-btn:active） */
.nav-btn:active {
  transform: scale(0.92);
}

/* 选中状态（对齐标准 .icon-btn.active: 黑底白字） */
.nav-btn.active {
  background: #1d1d1f;
  color: #fff;
}

/* 撰写按钮（对齐标准 .icon-btn.compose: 黑底白字 + 阴影 + hover 缩放） */
.compose-btn {
  background: #1d1d1f;
  color: #fff;
  border-radius: 8px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.compose-btn:hover {
  background: #333;
  transform: scale(1.05);
}

/* 图标字号（对齐标准: font-size:14px） */
.nav-icon {
  font-size: 14px;
  line-height: 1;
}

/* 用户头像（对齐标准 .avatar-sm: 28×28, font-size:10px, margin-bottom:12px, hover scale(1.1)） */
.user-avatar {
  width: 28px;
  height: 28px;
  margin-bottom: 12px;
  border-radius: 50%;
  background: #e5e5e5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
  cursor: pointer;
  color: #666;
  transition: transform 0.2s ease;
}

.user-avatar:hover {
  transform: scale(1.1);
}
</style>