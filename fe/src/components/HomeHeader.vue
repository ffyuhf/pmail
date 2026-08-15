<!-- ============================================================
  PMail — Navbar 顶部导航栏（Docusaurus 风格）
  改造日期: 20250425
  改造原因: 原 Header 桌面端 display:none 导致不可见
  ============================================================ -->
<template>
  <!-- Docusaurus 风格 Navbar -->
  <nav class="navbar">
    <div class="navbar__inner">
      <div class="navbar__items">
        <!-- 移动端汉堡菜单 -->
        <div class="navbar__toggle" @click="globalStatus.mobileDrawerVisible = true">
          <el-icon :size="20"><EpMenu /></el-icon>
        </div>
        <!-- Logo -->
        <router-link to="/" class="navbar__brand">
          <span class="navbar__title">PMail</span>
        </router-link>
      </div>

      <div class="navbar__items navbar__items--right">
        <!-- 设置按钮 -->
        <div class="navbar__item" @click="openSettings" :title="lang.settings">
          <el-icon :size="18"><Setting /></el-icon>
        </div>
      </div>
    </div>
  </nav>

  <!-- Settings Drawer -->
  <el-drawer v-model="globalStatus.settingsDrawerVisible" :size="isMobile ? '100%' : '600px'" :title="lang.settings" class="settings-drawer" :with-header="true" direction="rtl">
    <el-tabs :tab-position="isMobile ? 'top' : 'left'" class="settings-tabs">
      <el-tab-pane :label="lang.security">
        <SecuritySettings/>
      </el-tab-pane>
      <el-tab-pane :label="lang.group_settings">
        <GroupSettings/>
      </el-tab-pane>
      <el-tab-pane :label="lang.rule_setting">
        <RuleSettings/>
      </el-tab-pane>
      <el-tab-pane v-if="userInfos.is_admin" :label="lang.user_management">
        <UserManagement/>
      </el-tab-pane>
      <el-tab-pane :label="lang.plugin_settings">
        <PluginSettings/>
      </el-tab-pane>
    </el-tabs>
  </el-drawer>
</template>

<script setup lang="ts">
import {EpMenu} from "vue-icons-plus/ep";
import SecuritySettings from '@/components/SecuritySettings.vue'
import lang from '../i18n/i18n';
import GroupSettings from './GroupSettings.vue';
import RuleSettings from './RuleSettings.vue';
import UserManagement from './UserManagement.vue';
import PluginSettings from './PluginSettings.vue';
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";
import {useSettingsDrawer} from "@/composables/useSettingsDrawer";
import {useIsMobile} from "@/composables/useIsMobile";
import {Setting} from "@element-plus/icons-vue";

const globalStatus = useGlobalStatusStore();
const {openSettings} = useSettingsDrawer();
const userInfos = globalStatus.userInfos;

/** 通过 useIsMobile composable 获取响应式移动端状态（自动管理 resize 监听器） */
const {isMobile} = useIsMobile();
</script>

<style scoped>
/* Docusaurus 风格 Navbar */
.navbar {
  height: var(--ifm-navbar-height);
  background: var(--ifm-navbar-background-color);
  box-shadow: var(--ifm-navbar-shadow);
  border-bottom: 1px solid var(--ifm-border-color);
  display: flex;
  align-items: center;
  padding: 0 var(--ifm-spacing-lg);
  position: relative;
  z-index: 100;
  flex-shrink: 0;
}

.navbar__inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.navbar__items {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-sm);
}

.navbar__items--right {
  margin-left: auto;
}

/* 汉堡菜单：仅移动端显示 */
.navbar__toggle {
  display: none;
  width: 32px;
  height: 32px;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  border-radius: var(--ifm-global-radius);
  color: var(--ifm-color-content-secondary);
  transition: background var(--ifm-transition-fast);
}

.navbar__toggle:hover {
  background: var(--ifm-background-hover-color);
  color: var(--ifm-color-content);
}

/* Logo & 标题 */
.navbar__brand {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: var(--ifm-color-content);
  gap: var(--ifm-spacing-sm);
}

.navbar__brand:hover {
  text-decoration: none;
}

.navbar__title {
  font-size: 18px;
  font-weight: 700;
  color: var(--ifm-color-content);
  letter-spacing: -0.01em;
}

/* 右侧导航项 */
.navbar__item {
  width: 36px;
  height: 36px;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
  border-radius: var(--ifm-global-radius);
  color: var(--ifm-color-content-secondary);
  transition: background var(--ifm-transition-fast), color var(--ifm-transition-fast);
}

.navbar__item:hover {
  background: var(--ifm-background-hover-color);
  color: var(--ifm-color-primary);
}

/* 设置抽屉 */
.settings-tabs {
  height: 100%;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .navbar {
    padding: 0 var(--ifm-spacing-md);
  }
  .navbar__toggle {
    display: flex;
  }
}
</style>
