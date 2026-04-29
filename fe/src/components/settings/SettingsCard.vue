<!--
  SettingsCard — 设置页公共布局组件
  
  提取自 SecuritySettings、GroupSettings、RuleSettings、PluginSettings、UserManagement
  五个组件中重复的 settings-card / settings-header / settings-desc 结构和样式。
  
  用法:
    <SettingsCard title="标题" description="描述文本">
      <template #default>卡片主体内容</template>
      <template #footer>底部操作按钮（可选）</template>
    </SettingsCard>
  
  创建日期: 20260429
-->
<template>
  <div class="settings-card">
    <div class="settings-header">
      <h3 :class="dangerTitle ? 'danger-text' : ''">{{ title }}</h3>
      <p class="settings-desc" v-if="description">{{ description }}</p>
    </div>
    <slot></slot>
    <slot name="footer"></slot>
  </div>
</template>

<script setup lang="ts">
/**
 * @prop title     卡片标题
 * @prop description 卡片描述文本（可选）
 * @prop dangerTitle 是否将标题显示为危险色（如注销场景）
 */
defineProps<{
  title: string;
  description?: string;
  dangerTitle?: boolean;
}>();
</script>

<style scoped>
.settings-card {
  padding: 0;
}

.settings-header {
  margin-bottom: 24px;
}

.settings-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: var(--ifm-color-content);
  margin: 0 0 8px 0;
}

.danger-text {
  color: #ef4444 !important;
}

.settings-desc {
  font-size: 14px;
  color: var(--ifm-color-content-secondary);
  margin: 0;
}
</style>
