<!--
  MoveModal 移动邮件分组弹窗

  选择目标分组后确认移动。
  对齐旧前端 ListView.vue 的叶子分组过滤逻辑：
  - 排除含有子文件夹的父级分组
  - 递归平铺 + 递归收集 parentIds

  修改日期: 20260609
  修改原因: 旧实现直接使用树形数据，未过滤父级分组。
-->
<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="move-modal">
      <div class="modal-header">
        <h2>{{ lang.move_btn }}</h2>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div
          v-for="group in leafGroups"
          :key="group.id"
          class="group-item"
          :class="{ selected: selectedId === group.id }"
          @click="selectedId = group.id"
        >
          📁 {{ group.name }}
        </div>
        <div v-if="leafGroups.length === 0" class="empty-list" style="padding: 24px; text-align: center; color: var(--text-placeholder);">—</div>
      </div>
      <div class="modal-footer">
        <button class="btn-secondary" @click="$emit('close')">{{ lang.cancel || 'Cancel' }}</button>
        <button class="btn-primary" :disabled="!selectedId" @click="onConfirm">
          {{ lang.move_btn }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import lang from '@/i18n'
import type { GroupItem } from '@/types/api'

const props = defineProps<{
  groups: GroupItem[]
}>()

const emit = defineEmits<{
  'close': []
  /** 确认移动：传递 groupId 和 groupName（对齐后端 moveRequest） */
  'confirm': [groupId: number, groupName: string]
}>()

const selectedId = ref<number | null>(null)

/**
 * 收集所有含有子文件夹的分组 ID 集合
 * 对齐旧前端 ListView.vue collectParentIds 逻辑
 */
function collectParentIds(items: GroupItem[]): Set<number> {
  const ids = new Set<number>()
  for (const item of items) {
    if (item.children && item.children.length > 0) {
      ids.add(item.id)
      for (const childId of collectParentIds(item.children)) {
        ids.add(childId)
      }
    }
  }
  return ids
}

/**
 * 获取平铺的叶子分组列表（排除含子文件夹的父级分组）
 * 对齐旧前端 ListView.vue updateGroupList 逻辑
 */
const leafGroups = computed(() => {
  const parentIds = collectParentIds(props.groups)
  const result: { id: number; name: string }[] = []

  function flatten(items: GroupItem[]) {
    for (const item of items) {
      if (!parentIds.has(item.id)) {
        result.push({ id: item.id, name: item.label })
      }
      if (item.children && item.children.length > 0) {
        flatten(item.children)
      }
    }
  }

  flatten(props.groups)
  return result
})

/** 确认移动：查找选中的分组名称，一并传递给父组件 */
function onConfirm() {
  if (selectedId.value === null) return
  const found = leafGroups.value.find((g: { id: number; name: string }) => g.id === selectedId.value)
  const groupName = found?.name || ''
  emit('confirm', selectedId.value, groupName)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.move-modal {
  width: 360px;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 { font-size: 16px; font-weight: 600; }

.close-btn {
  width: 32px; height: 32px; border: none; background: transparent;
  font-size: 20px; cursor: pointer; border-radius: var(--radius-sm);
  display: flex; align-items: center; justify-content: center;
}
.close-btn:hover { background: var(--bg-hover); }

.modal-body {
  padding: 12px;
  max-height: 300px;
  overflow-y: auto;
}

.group-item {
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 14px;
  transition: background var(--transition);
}

.group-item:hover { background: var(--bg-hover); }
.group-item.selected { background: var(--bg-active); }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid var(--border-color);
}

.btn-primary, .btn-secondary {
  padding: 8px 16px;
  border: none;
  border-radius: var(--radius);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition);
}

.btn-primary { background: var(--accent-color); color: #fff; }
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary { background: var(--bg-secondary); border: 1px solid var(--border-color); }
.btn-secondary:hover { background: var(--bg-hover); }
</style>