<!--
  SettingsPanel 设置面板组件

  全屏模态设置面板，左侧标签栏 + 右侧内容区。
  包含：安全设置、分组管理、规则管理、用户管理、插件管理。

  创建日期: 20260609
-->
<template>
  <div class="settings-overlay" @click.self="$emit('close')">
    <div class="settings-panel">
      <!-- 左侧标签栏 -->
      <nav class="settings-nav">
        <h2 class="nav-title">{{ lang.settings }}</h2>
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="nav-item"
          :class="{ active: activeTab === tab.key }"
          @click="activeTab = tab.key"
        >
          <span class="nav-icon">{{ tab.icon }}</span>
          <span class="nav-label">{{ tab.label }}</span>
        </button>
      </nav>

      <!-- 右侧内容区 -->
      <div class="settings-content">
        <div class="content-header">
          <h2>{{ currentLabel }}</h2>
          <button class="close-btn" @click="$emit('close')">×</button>
        </div>
        <div class="content-body">
          <SecuritySettings v-if="activeTab === 'security'" />
          <GroupSettings v-else-if="activeTab === 'groups'" />
          <RuleSettings v-else-if="activeTab === 'rules'" />
          <UserManagement v-else-if="activeTab === 'users'" />
          <PluginSettings v-else-if="activeTab === 'plugins'" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import lang from '@/i18n'
import { useGlobalStore } from '@/stores/global'
import SecuritySettings from '@/components/settings/SecuritySettings.vue'
import GroupSettings from '@/components/settings/GroupSettings.vue'
import RuleSettings from '@/components/settings/RuleSettings.vue'
import UserManagement from '@/components/settings/UserManagement.vue'
import PluginSettings from '@/components/settings/PluginSettings.vue'

const props = defineProps<{
  activeTab?: string
}>()

defineEmits<{ 'close': [] }>()

const globalStore = useGlobalStore()

type TabKey = 'security' | 'groups' | 'rules' | 'users' | 'plugins'

const activeTab = ref<TabKey>((props.activeTab as TabKey) || 'security')

/**
 * 标签列表（用户管理仅管理员可见）
 * 对齐旧前端 HomeHeader.vue: <el-tab-pane v-if="userInfos.is_admin">
 */
const tabs = computed(() => {
  const all: { key: TabKey; label: string; icon: string }[] = [
    { key: 'security', label: lang.security, icon: '🔒' },
    { key: 'groups', label: lang.group_settings, icon: '📁' },
    { key: 'rules', label: lang.rule_setting, icon: '📋' },
    { key: 'users', label: lang.user_management, icon: '👥' },
    { key: 'plugins', label: lang.plugin_settings, icon: '🔌' },
  ]
  /* 用户管理标签仅管理员可见（对齐旧前端 HomeHeader.vue 权限判断） */
  if (!globalStore.isAdmin) {
    return all.filter(t => t.key !== 'users')
  }
  return all
})

const currentLabel = computed(() => tabs.value.find((t) => t.key === activeTab.value)?.label || '')
</script>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 1200;
  display: flex;
  align-items: center;
  justify-content: center;
}

/**
 * 设置面板：固定高度，不随内容变化
 * 对齐旧前端 el-drawer 行为：抽屉大小固定（600px 宽、全高），切换标签时面板不变
 */
.settings-panel {
  width: 800px;
  height: 85vh;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  display: flex;
  overflow: hidden;
}

.settings-nav {
  width: 200px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border-color);
  padding: 20px 0;
  flex-shrink: 0;
}

.nav-title {
  font-size: 16px;
  font-weight: 700;
  padding: 0 16px;
  margin-bottom: 16px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  width: 100%;
  padding: 10px 16px;
  border: none;
  background: transparent;
  font-size: 14px;
  cursor: pointer;
  text-align: left;
  transition: background var(--transition);
}

.nav-item:hover { background: var(--bg-hover); }
.nav-item.active { background: var(--bg-active); font-weight: 500; }
.nav-icon { font-size: 16px; }

.settings-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.content-header h2 { font-size: 16px; font-weight: 600; }

.close-btn {
  width: 32px; height: 32px; border: none; background: transparent;
  font-size: 20px; cursor: pointer; border-radius: var(--radius-sm);
  display: flex; align-items: center; justify-content: center;
}
.close-btn:hover { background: var(--bg-hover); }

.content-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>