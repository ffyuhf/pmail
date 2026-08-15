<template>
  <SettingsCard :title="lang.extensions" :description="lang.extensions_desc">

    <div class="plugin-container">
      <el-tabs class="custom-tabs">
        <el-tab-pane v-for="(src, name) in pluginList" :key="src" :label="name">
          <div class="iframe-wrapper">
            <iframe :src="src"></iframe>
          </div>
        </el-tab-pane>
        <el-tab-pane v-if="Object.keys(pluginList).length === 0" label="No Plugins">
          <div class="empty-state">
            <el-empty description="No plugins are currently installed" :image-size="120" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </SettingsCard>
</template>

<script setup lang="ts">
import {reactive} from 'vue'
import {pluginService} from "@/services/pluginService";
import lang from '../i18n/i18n';
import SettingsCard from "@/components/settings/SettingsCard.vue";

const pluginList = reactive<Record<string, string>>({})

/** 通过 pluginService 获取已安装插件列表 */
pluginService.getPluginList().then((res: any) => {
  if (res.data != null && res.data.length > 0) {
    for (let i = 0; i < res.data.length; i++) {
      let name = res.data[i];
      pluginList[name] = "/api/plugin/settings/" + name + "/index.html";
    }
  }
})
</script>

<style scoped>

.plugin-container {
  flex-grow: 1;
  border: 1px solid var(--ifm-border-color);
  border-radius: var(--ifm-global-radius);
  background: var(--ifm-background-surface-color);
  overflow: hidden;
}

.custom-tabs :deep(.el-tabs__header) {
  margin-bottom: 0;
  background: var(--ifm-background-color);
  padding: 0 16px;
  border-bottom: 1px solid var(--ifm-border-color);
}

.custom-tabs :deep(.el-tabs__content) {
  height: 100%;
}

.iframe-wrapper {
  height: calc(100vh - 250px);
  min-height: 400px;
}

iframe {
  width: 100%;
  height: 100%;
  border: 0;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 300px;
}
</style>
