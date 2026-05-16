<!-- ============================================================
  PMail — 根布局组件（Docusaurus 风格）
  布局结构: Navbar (顶部) + Sidebar (左侧) + Content (右侧)
  改造日期: 20250425
  ============================================================ -->
<script setup>
import {RouterView, useRoute} from 'vue-router'
import HomeHeader from '@/components/HomeHeader.vue'
import HomeAside from '@/components/HomeAside.vue';
import {ref, watch, onMounted, computed} from 'vue'
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const route = useRoute()
const pageName = ref(route.name)
const globalStatus = useGlobalStatusStore();

onMounted(() => {
  globalStatus.init(() => {});
});

watch(
    () => route.fullPath,
    () => {
      pageName.value = route.name
    }
)

/** 布局控制计算属性：消除模板中重复的 pageName 条件判断 */
const showHeader = computed(() => pageName.value !== 'login' && pageName.value !== 'setup')
const showSidebar = computed(() => showHeader.value && pageName.value !== 'editer')
const isFullBleed = computed(() => !showHeader.value || pageName.value === 'editer')

</script>

<template>
  <div id="main">
    <!-- Navbar: login/setup 页不显示 -->
    <HomeHeader v-if="showHeader"/>
    <div id="content">
      <!-- Sidebar: login/setup/editer 页隐藏 -->
      <div id="aside" v-if="showSidebar">
        <HomeAside/>
      </div>
      <!-- 移动端侧边栏抽屉：login/setup/editer 页隐藏 -->
      <el-drawer
          v-if="showSidebar"
          v-model="globalStatus.mobileDrawerVisible"
          direction="ltr"
          size="260px"
          :with-header="false"
          class="mobile-aside-drawer"
      >
        <HomeAside/>
      </el-drawer>
      <!-- 主内容区：login/setup/editer 页去除内边距 -->
      <div id="body" :class="{ 'full-bleed': isFullBleed }">
        <RouterView v-slot="{ Component }">
          <!-- 修改日期: 20260516 — 临时移除 transition 排查空屏问题 -->
          <component :is="Component" :key="route.fullPath" />
        </RouterView>
      </div>
    </div>
  </div>
</template>


<style scoped>
/* 主容器: 纵向 flex，Navbar 在上，Content 在下 */
#main {
  height: 100%;
  display: flex;
  flex-direction: column;
}

/* Sidebar: 固定宽度、白色背景、右侧边框 */
#aside {
  width: var(--ifm-sidebar-width);
  min-width: var(--ifm-sidebar-width);
  max-width: var(--ifm-sidebar-width);
  border-right: 1px solid var(--ifm-border-color);
  background: var(--ifm-sidebar-background-color);
  overflow: hidden;
}

/* 主内容区：减小 padding 避免与视图容器双重内边距 */
#body {
  width: 100%;
  height: 100%;
  padding: var(--ifm-spacing-sm);
  box-sizing: border-box;
  overflow: hidden;
}

/* Login/Setup 全屏页面无内边距 */
#body.full-bleed {
  padding: 0;
}

/* 内容区: 横向 flex */
#content {
  display: flex;
  flex-grow: 1;
  overflow: hidden;
  min-height: 0;
}

/* 移动端抽屉内边距清零 */
.mobile-aside-drawer :deep(.el-drawer__body) {
  padding: 0;
}

/* 移动端适配 */
@media (max-width: 768px) {
  #aside {
    display: none !important;
  }
  #body {
    padding: var(--ifm-spacing-md);
  }
  #body.full-bleed {
    padding: 0;
  }
}
</style>
