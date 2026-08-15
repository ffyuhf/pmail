<!--
  GroupSettings 分组管理设置组件

  对齐旧前端 fe/src/components/GroupSettings.vue 完整功能：
  - 递归树形展示分组结构（带文件夹图标）
  - 每个节点可添加子文件夹（+）/删除（🗑）
  - 新建根分组按钮
  - 新建时 inline input 编辑，blur 或 enter 提交

  修改日期: 20260609
  修改原因: 旧实现仅有平铺列表，缺少树形结构和子文件夹创建。
-->
<template>
  <div class="group-settings">
    <div class="section-header">
      <h3>{{ lang.email_folders }}</h3>
      <p class="section-desc">{{ lang.email_folders_desc }}</p>
    </div>

    <!-- 递归树形分组 -->
    <div class="tree-wrap">
      <TreeNode
        v-for="node in treeData"
        :key="node.id"
        :node="node"
        @add="onAddChild"
        @delete="onDelete"
        @create="onCreate"
      />
    </div>

    <div class="section-actions">
      <button class="btn-primary" @click="addRoot">{{ lang.add_group }}</button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import lang from '@/i18n'
import { getGroups, createGroup, deleteGroup } from '@/services/groupService'
import type { GroupItem } from '@/types/api'
import TreeNode from './TreeNode.vue'

/* eslint-disable @typescript-eslint/no-explicit-any */

/** 树节点扩展：含临时节点标记和父 ID */
export interface TreeNodeData extends GroupItem {
  _temp?: boolean
  _parentId?: number
}

const treeData = ref<TreeNodeData[]>([])

/** 加载树形分组 */
async function fetchGroups() {
  const res: any = await getGroups()
  if (res.errorNo === 0) {
    treeData.value = res.data || []
  }
}

/** 添加根级临时节点 */
function addRoot() {
  treeData.value.push({
    id: -1,
    label: '',
    tag: '',
    children: [],
    parent_id: 0,
    _temp: true as any,
    _parentId: 0,
  } as any)
}

/** 添加子级临时节点 */
function onAddChild(parent: TreeNodeData) {
  if (!parent.children) {
    parent.children = []
  }
  parent.children.push({
    id: -1,
    label: '',
    tag: '',
    children: [],
    parent_id: parent.id,
    _temp: true as any,
    _parentId: parent.id,
  } as any)
}

/** 创建分组（调用后端 API） */
async function onCreate(item: TreeNodeData) {
  if (!item.label) return
  try {
    const res: any = await createGroup(item.label, item._parentId || 0)
    if (res.errorNo === 0) {
      fetchGroups()
    } else {
      alert(res.errorMsg || lang.fail)
      removeTempNode(item)
    }
  } catch {
    removeTempNode(item)
  }
}

/** 删除分组 */
async function onDelete(item: TreeNodeData) {
  if (item._temp) {
    removeTempNode(item)
    return
  }
  if (!confirm(lang.del_rule_confirm)) return
  try {
    const res: any = await deleteGroup(item.id)
    if (res.errorNo === 0) {
      fetchGroups()
    } else {
      alert(res.errorMsg || lang.fail)
    }
  } catch { /* ignore */ }
}

/** 从树中移除临时节点 */
function removeTempNode(item: TreeNodeData) {
  const removeFrom = (list: TreeNodeData[]): boolean => {
    for (let i = list.length - 1; i >= 0; i--) {
      if (list[i] === item) {
        list.splice(i, 1)
        return true
      }
      if (list[i].children && removeFrom(list[i].children as TreeNodeData[])) {
        return true
      }
    }
    return false
  }
  removeFrom(treeData.value)
}

onMounted(fetchGroups)
</script>

<style scoped>
.group-settings { width: 100%; }
.section-header { margin-bottom: 16px; }
.section-header h3 { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.section-desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; }
.section-actions { margin-top: 16px; }
.tree-wrap {
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  padding: 12px;
}
.btn-primary {
  padding: 8px 20px; border: none; border-radius: var(--radius);
  background: var(--accent-color); color: #fff; font-size: 13px; cursor: pointer;
}
.btn-primary:hover { background: var(--accent-hover); }
</style>
