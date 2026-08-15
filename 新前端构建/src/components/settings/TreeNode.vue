<!--
  TreeNode 递归树节点组件

  用于 GroupSettings 的递归树形渲染，支持：
  - 文件夹图标 + 名称显示
  - 临时节点（id === -1）显示 input 编辑框
  - 添加子节点按钮
  - 删除节点按钮
  - blur/enter 提交创建

  创建日期: 20260609
-->
<template>
  <div class="tree-node">
    <div class="node-row" @mouseenter="hover = true" @mouseleave="hover = false">
      <span class="folder-icon">📁</span>

      <!-- 临时节点：inline input 编辑 -->
      <input
        v-if="node._temp"
        v-model="node.label"
        class="node-input"
        :placeholder="lang.folder_name"
        @blur="onSubmit"
        @keyup.enter="onSubmit"
        ref="inputRef"
        autofocus
      />

      <!-- 正式节点：显示名称 -->
      <span v-else class="node-label">{{ node.label }}</span>

      <div v-if="hover || node._temp" class="node-actions">
        <button
          v-if="!node._temp"
          class="icon-btn"
          @click="$emit('add', node)"
          title="Add Subfolder"
        >+</button>
        <button class="icon-btn danger" @click="$emit('delete', node)" title="Delete">🗑</button>
      </div>
    </div>

    <!-- 递归渲染子节点 -->
    <div v-if="node.children && node.children.length > 0" class="node-children">
      <TreeNode
        v-for="child in node.children"
        :key="child.id"
        :node="child"
        @add="$emit('add', $event)"
        @delete="$emit('delete', $event)"
        @create="$emit('create', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import lang from '@/i18n'

export interface TreeNodeData {
  id: number
  label: string
  tag: string
  children?: TreeNodeData[]
  parent_id?: number
  _temp?: boolean
  _parentId?: number
}

const props = defineProps<{
  node: TreeNodeData
}>()

const emit = defineEmits<{
  add: [node: TreeNodeData]
  delete: [node: TreeNodeData]
  create: [node: TreeNodeData]
}>()

const hover = ref(false)
const inputRef = ref<HTMLInputElement | null>(null)

/** blur 或 enter 时提交创建 */
function onSubmit() {
  if (props.node._temp) {
    emit('create', props.node)
  }
}

onMounted(() => {
  if (props.node._temp) {
    nextTick(() => {
      inputRef.value?.focus()
    })
  }
})
</script>

<style scoped>
.tree-node { margin-left: 0; }
.node-row {
  display: flex; align-items: center; gap: 8px;
  padding: 6px 8px; border-radius: var(--radius-sm);
  transition: background var(--transition);
}
.node-row:hover { background: var(--bg-hover); }
.folder-icon { font-size: 14px; flex-shrink: 0; }
.node-label { font-size: 14px; }
.node-input {
  flex: 1; padding: 4px 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm); font-size: 13px; outline: none;
}
.node-input:focus { border-color: var(--accent-color); }
.node-actions {
  display: flex; gap: 4px; margin-left: auto;
}
.node-children {
  margin-left: 24px;
  border-left: 1px solid var(--border-color);
  padding-left: 8px;
}
.icon-btn {
  width: 24px; height: 24px; border: none; background: transparent;
  border-radius: 4px; cursor: pointer; font-size: 12px;
  display: flex; align-items: center; justify-content: center;
}
.icon-btn:hover { background: var(--bg-hover); }
.icon-btn.danger { color: var(--danger-color); }
</style>
