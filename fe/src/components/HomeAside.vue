<!-- ============================================================
  PMail — Sidebar 侧边栏（Docusaurus 风格）
  改造日期: 20250425
  改造原因: 移除 glassmorphism，改为纯色简洁风格
  改造日期: 20260509
  改造原因: 支持树形折叠菜单，"全部邮件数据"作为可展开/折叠的分组标题
  改造日期: 20260509
  改造原因: 子文件夹列表添加左侧竖线，指示层级关系
  ============================================================ -->
<template>
  <div class="sidebar">
    <!-- 搜索框：回车触发搜索，清空时重置搜索结果 -->
    <div class="sidebar__search">
      <el-input
        v-model="searchQuery"
        :placeholder="lang.search"
        prefix-icon="Search"
        clearable
        size="default"
        @keyup.enter="handleSearch"
        @clear="handleClearSearch"
      />
    </div>

    <!-- 菜单列表：树形结构渲染 -->
    <nav class="sidebar__menu">
      <ul class="menu__list">
        <li
          v-for="item in treeData"
          :key="item.label"
          class="menu__list-item"
        >
          <!-- 有子节点：渲染为可折叠分组标题 -->
          <template v-if="item.children && item.children.length > 0">
            <a
              class="menu__group-title"
              @click="toggleGroup(item.label)"
            >
              <span class="menu__group-arrow">{{ expandedGroups[item.label] ? '▼' : '▶' }}</span>
              <span>{{ item.label }}</span>
            </a>
            <!-- 子节点列表 -->
            <ul v-show="expandedGroups[item.label]" class="menu__sub-list">
              <li
                v-for="child in item.children"
                :key="child.tag"
                class="menu__list-item"
              >
                <a
                  class="menu__link"
                  :class="{ 'menu__link--active': activeGroup === child.tag }"
                  @click="handleMenuSelect(child)"
                >
                  {{ child.label }}
                </a>
              </li>
            </ul>
          </template>
          <!-- 无子节点：渲染为普通菜单项 -->
          <template v-else>
            <a
              class="menu__link"
              :class="{ 'menu__link--active': activeGroup === item.tag }"
              @click="handleMenuSelect(item)"
            >
              {{ item.label }}
            </a>
          </template>
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
import { ref, watch, computed, reactive } from "vue";
import useGroupStore from "../stores/group";
import lang from "../i18n/i18n";
import { groupService } from "@/services/groupService";
import { Setting } from "@element-plus/icons-vue";
import { useGlobalStatusStore } from "../stores/useGlobalStatusStore";
import { useSettingsDrawer } from "@/composables/useSettingsDrawer";
import type { GroupItem } from "@/types/api";

const groupStore = useGroupStore();
const globalStatus = useGlobalStatusStore();
const { openSettings } = useSettingsDrawer();
const isLogin = computed(() => globalStatus.isLogin);
const router = useRouter();

/** 树形分组数据（保留嵌套结构，不再扁平化） */
const treeData = ref<GroupItem[]>([]);
const searchQuery = ref("");
const activeGroup = ref(groupStore.tag);

/** 各分组的展开/折叠状态，key 为分组 label */
const expandedGroups = reactive<Record<string, boolean>>({});

// 保持活动菜单与 store 同步
watch(() => groupStore.tag, (newVal) => {
  activeGroup.value = newVal;
});

/** 通过 groupService 获取分组树，保留原始嵌套结构 */
groupService.getGroupTree().then((res: any) => {
  if (res.data) {
    treeData.value = res.data;
    // 默认展开第一个分组（"全部邮件数据"），确保用户首次看到常用文件夹
    if (res.data.length > 0) {
      expandedGroups[res.data[0].label] = true;
    }
  }
});

/** 切换分组的展开/折叠状态 */
const toggleGroup = function (label: string) {
  expandedGroups[label] = !expandedGroups[label];
};

/** 点击菜单项：更新 store 并导航到邮件列表页 */
const handleMenuSelect = function (item: GroupItem) {
  if (item.tag) {
    groupStore.name = item.label;
    groupStore.tag = item.tag;
    router.push({ name: "list" });
  }
};

/** 回车搜索：将关键词同步到 store，触发列表页自动刷新 */
const handleSearch = function () {
  groupStore.keyword = searchQuery.value.trim();
};

/** 清空搜索：重置 store 关键词，恢复完整列表 */
const handleClearSearch = function () {
  searchQuery.value = "";
  groupStore.keyword = "";
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
  background-color: var(--ifm-background-hover-color);
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

/* 分组标题（可折叠） */
.menu__group-title {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-sm);
  padding: 8px 12px;
  font-size: 14px;
  line-height: 1.4;
  font-weight: 600;
  color: var(--ifm-color-content);
  border-radius: var(--ifm-global-radius);
  cursor: pointer;
  text-decoration: none;
  user-select: none;
  transition: background-color var(--ifm-transition-fast);
}

.menu__group-title:hover {
  background-color: var(--ifm-background-hover-color);
}

/* 折叠箭头图标 */
.menu__group-arrow {
  font-size: 10px;
  width: 16px;
  text-align: center;
  flex-shrink: 0;
  color: var(--ifm-color-content-secondary);
  transition: transform var(--ifm-transition-fast);
}

/* 子节点列表：带左侧竖线指示层级关系 */
.menu__sub-list {
  list-style: none;
  padding: 0;
  padding-left: 20px;
  margin: 0;
  position: relative;
}

/* 左侧竖线：伪元素绘制，颜色与全局边框色一致 */
.menu__sub-list::before {
  content: '';
  position: absolute;
  left: 12px;
  top: 0;
  bottom: 0;
  width: 1px;
  background: var(--ifm-border-color);
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
