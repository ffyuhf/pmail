/**
 * 分组 Store
 *
 * 从旧前端 fe/src/stores/group.ts 迁移。
 * 管理邮件分组/文件夹的状态。
 *
 * 修改日期: 20260609
 * 修改原因: 增加树形扁平化逻辑 flatGroups，解决子文件夹丢失问题。
 *   后端 GET /api/group 返回树形嵌套结构（含 children），
 *   旧前端 HomeAside.vue 递归渲染整棵树，新前端需要将树扁平化供导航使用。
 *
 * 创建日期: 20260609
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { GroupItem } from '@/types/api'
import { getGroups } from '@/services/groupService'
import { DEFAULT_GROUP_TAG } from '@/utils/constants'

export const useGroupStore = defineStore('group', () => {
  /** 分组树形列表（后端原始数据，含嵌套 children） */
  const groups = ref<GroupItem[]>([])
  /** 当前选中的分组 Tag */
  const currentTag = ref<string>(DEFAULT_GROUP_TAG)
  /** 加载状态 */
  const loading = ref(false)

  /**
   * 扁平化分组列表（计算属性）
   *
   * 递归展开树形结构的所有节点（含子文件夹），用于导航和分组选择。
   * 跳过没有 tag 的纯容器节点（如"All Email"），这类节点在后端仅作为
   * 分组标题，不可选择也不能放入邮件。旧前端 HomeAside.vue 中有 children
   * 的节点渲染为可折叠标题（不可点击），叶子节点才渲染为可点击菜单项。
   *
   * 修改日期: 20260609
   * 修改原因: 修复"All Email"父容器节点显示在可选文件夹列表中的问题，
   *   以及该节点被选中后导致后端查询不按 type 过滤（发件箱邮件泄漏到收件箱）。
   */
  const flatGroups = computed(() => {
    const result: GroupItem[] = []
    function flatten(items: GroupItem[]) {
      for (const item of items) {
        /* 仅收集有 tag 的叶子节点（可选择的文件夹） */
        if (item.tag) {
          result.push(item)
        }
        /* 无论当前节点是否有 tag，都递归展开其子节点 */
        if (item.children && item.children.length > 0) {
          flatten(item.children)
        }
      }
    }
    flatten(groups.value)
    return result
  })

  /** 从后端加载分组列表 */
  async function fetchGroups() {
    loading.value = true
    try {
      const res: any = await getGroups()
      /* axios 拦截器已解包，直接读 errorNo */
      if (res.errorNo === 0) {
        groups.value = res.data || []
      }
    } finally {
      loading.value = false
    }
  }

  /** 切换当前选中的分组 */
  function setCurrentTag(tag: string) {
    currentTag.value = tag || DEFAULT_GROUP_TAG
  }

  return { groups, flatGroups, currentTag, loading, fetchGroups, setCurrentTag }
})
