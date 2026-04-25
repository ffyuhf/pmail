<!-- ============================================================
  PMail — Sidebar 侧边栏（Docusaurus 风格）
  改造日期: 20250425
  改造原因: 移除 glassmorphism，改为纯色简洁风格
  ============================================================ -->
<template>
  <div class="sidebar">
    <!-- 搜索框 -->
    <div class="sidebar__search">
      <el-input
        v-model="searchQuery"
        :placeholder="lang.search"
        prefix-icon="Search"
        clearable
        size="default"
      />
    </div>

    <!-- 菜单列表 -->
    <nav class="sidebar__menu">
      <ul class="menu__list">
        <li
          v-for="item in data"
          :key="item.tag"
          class="menu__list-item"
        >
          <a
            class="menu__link"
            :class="{ 'menu__link--active': activeGroup === item.tag }"
            @click="handleMenuSelect(item.tag)"
          >
            {{ item.label }}
          </a>
        </li>
      </ul>
    </nav>

    <!-- 底部设置按钮 -->
    <div class="sidebar__footer" v-if="isLogin">
      <a class="menu__link" @click="openSettings">
        <el-icon :size="16"><Setting /></el-icon>
        <span>{{ lang.settings }}</span>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from "vue-router";
import { ref, watch, computed } from "vue";
import useGroupStore from "../stores/group";
import lang from "../i18n/i18n";
import { http } from "@/utils/axios";
import { Setting } from "@element-plus/icons-vue";
import { useGlobalStatusStore } from "../stores/useGlobalStatusStore";
import type { GroupItem } from "@/types/api";

const groupStore = useGroupStore();
const globalStatus = useGlobalStatusStore();
const isLogin = computed(() => globalStatus.isLogin);
const router = useRouter();
const data = ref<GroupItem[]>([]);
const searchQuery = ref("");
const activeGroup = ref(groupStore.tag);

// 保持活动菜单与 store 同步
watch(() => groupStore.tag, (newVal) => {
  activeGroup.value = newVal;
});

http.get("/api/group").then((res: any) => {
  if (res.data) {
    const list: GroupItem[] = [];
    const traverse = (items: GroupItem[]) => {
      items.forEach(node => {
        list.push(node);
        if(node.children) traverse(node.children);
      });
    }
    traverse(res.data);
    data.value = list;
  }
});

const handleMenuSelect = function (index: string) {
  const selected = data.value.find(d => d.tag === index);
  if (selected) {
    groupStore.name = selected.label;
    groupStore.tag = selected.tag;
    router.push({ name: "list" });
  }
};

const openSettings = function () {
  if (Object.keys(globalStatus.userInfos).length === 0) {
    globalStatus.init(() => {
      globalStatus.settingsDrawerVisible = true;
    });
  } else {
    globalStatus.settingsDrawerVisible = true;
  }
};
</script>

<style scoped>
/* Docusaurus 风格 Sidebar */
.sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--ifm-sidebar-background-color);
  padding: var(--ifm-spacing-md) 0;
  overflow: hidden;
}

/* 搜索框 */
.sidebar__search {
  padding: 0 var(--ifm-spacing-md);
  margin-bottom: var(--ifm-spacing-md);
}

.sidebar__search :deep(.el-input__wrapper) {
  border-radius: var(--ifm-global-radius);
  background: var(--ifm-background-surface-color);
}

.sidebar__search :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--ifm-border-color) inset;
}

.sidebar__search :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--ifm-color-primary) inset;
}

/* 菜单列表 */
.sidebar__menu {
  flex-grow: 1;
  overflow-y: auto;
  padding: 0 var(--ifm-spacing-sm);
}

.menu__list {
  list-style: none;
  padding: 0;
  margin: 0;
}

.menu__list-item {
  margin-bottom: 2px;
}

/* 菜单链接（Docusaurus 风格） */
.menu__link {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-sm);
  padding: 8px 12px;
  font-size: 14px;
  line-height: 1.4;
  color: var(--ifm-color-content-secondary);
  border-radius: var(--ifm-global-radius);
  cursor: pointer;
  text-decoration: none;
  transition: background-color var(--ifm-transition-fast), color var(--ifm-transition-fast);
}

.menu__link:hover {
  background-color: var(--pm-bg-hover);
  color: var(--ifm-color-content);
  text-decoration: none;
}

/* 激活状态：左侧绿色竖条指示 */
.menu__link--active {
  color: var(--ifm-color-primary);
  font-weight: 600;
  background-color: rgba(37, 194, 160, 0.08);
  border-left: 3px solid var(--ifm-color-primary);
  padding-left: 9px;
}

.menu__link--active:hover {
  background-color: rgba(37, 194, 160, 0.12);
  color: var(--ifm-color-primary);
}

/* 底部 */
.sidebar__footer {
  padding: var(--ifm-spacing-sm) var(--ifm-spacing-md);
  border-top: 1px solid var(--ifm-border-color);
  margin-top: var(--ifm-spacing-sm);
}

.sidebar__footer .menu__link {
  color: var(--ifm-color-content-muted);
}

.sidebar__footer .menu__link:hover {
  color: var(--ifm-color-content);
}
</style>
