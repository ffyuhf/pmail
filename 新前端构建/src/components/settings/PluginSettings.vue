<!--
  PluginSettings 插件管理设置组件

  对齐旧前端 fe/src/components/PluginSettings.vue 完整功能：
  - 通过 tab 切换每个插件的 iframe 配置页面
  - 无插件时显示空状态

  修改日期: 20260609
  修改原因: 旧实现仅有插件列表，缺少 iframe 嵌入。
-->
<template>
  <div class="plugin-settings">
    <div class="section-header">
      <h3>{{ lang.extensions || lang.plugin_settings }}</h3>
      <p class="section-desc">{{ lang.extensions_desc || lang.plugin_settings_desc }}</p>
    </div>

    <div class="plugin-container">
      <!-- Tab 导航 -->
      <div class="plugin-tabs" v-if="pluginNames.length > 0">
        <button
          v-for="name in pluginNames"
          :key="name"
          class="tab-btn"
          :class="{ active: activePlugin === name }"
          @click="activePlugin = name"
        >{{ name }}</button>
      </div>

      <!-- iframe 面板 -->
      <div class="iframe-wrap" v-if="activePlugin && pluginSrcMap[activePlugin]">
        <iframe :src="pluginSrcMap[activePlugin]" frameborder="0"></iframe>
      </div>

      <!-- 空状态 -->
      <div v-if="pluginNames.length === 0" class="empty-state">
        <p>—</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import lang from '@/i18n'
import { getPlugins } from '@/services/pluginService'

/* eslint-disable @typescript-eslint/no-explicit-any */

/** 插件名 → iframe URL 映射 */
const pluginSrcMap = ref<Record<string, string>>({})
const activePlugin = ref('')

const pluginNames = ref<string[]>([])

async function fetchPlugins() {
  const res: any = await getPlugins()
  if (res.errorNo === 0 && res.data && res.data.length > 0) {
    const map: Record<string, string> = {}
    const names: string[] = []
    for (const name of res.data) {
      map[name] = `/api/plugin/settings/${name}/index.html`
      names.push(name)
    }
    pluginSrcMap.value = map
    pluginNames.value = names
    if (names.length > 0) {
      activePlugin.value = names[0]
    }
  }
}

onMounted(fetchPlugins)
</script>

<style scoped>
.plugin-settings { width: 100%; }
.section-header { margin-bottom: 16px; }
.section-header h3 { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.section-desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; }

.plugin-container {
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}

/* Tab 导航 */
.plugin-tabs {
  display: flex; gap: 0;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-secondary);
}
.tab-btn {
  padding: 10px 16px; border: none; background: transparent;
  font-size: 13px; cursor: pointer; color: var(--text-secondary);
  border-bottom: 2px solid transparent; transition: var(--transition);
}
.tab-btn:hover { background: var(--bg-hover); }
.tab-btn.active {
  background: var(--bg-color); color: var(--accent-color);
  border-bottom-color: var(--accent-color);
}

/* iframe */
.iframe-wrap {
  height: calc(100vh - 280px);
  min-height: 400px;
}
.iframe-wrap iframe {
  width: 100%; height: 100%; border: 0;
}

/* 空状态 */
.empty-state {
  display: flex; justify-content: center; align-items: center;
  height: 200px; color: var(--text-placeholder);
}
</style>
