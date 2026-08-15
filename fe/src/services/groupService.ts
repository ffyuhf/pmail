/**
 * 分组 API 服务层
 *
 * 封装所有分组（邮件文件夹）相关的 HTTP 请求。
 * 提供树结构和平铺结构两种列表获取方式，以及分组的增删操作。
 *
 * 创建日期: 20260429
 * @module groupService
 */

import { http } from '@/utils/axios'
import type { GroupItem, GroupListItem } from '@/types/api'

/**
 * 获取分组树结构
 * 用于侧边栏菜单展示（嵌套层级结构）
 * @returns 分组树数组
 */
function getGroupTree() {
  return http.get<GroupItem[]>('/api/group')
}

/**
 * 获取分组平铺列表
 * 用于邮件移动下拉菜单等场景（仅 id + name）
 * @returns 分组平铺列表
 */
function getGroupList() {
  return http.post<GroupListItem[]>('/api/group/list')
}

/**
 * 新增分组
 * @param name - 分组名称
 * @param parentId - 父级分组 ID（0 为顶级）
 * @returns 新增结果
 */
function addGroup(name: string, parentId: number) {
  return http.post('/api/group/add', { name, parent_id: parentId })
}

/**
 * 删除分组
 * @param id - 分组 ID
 * @returns 删除结果
 */
function deleteGroup(id: number) {
  return http.post('/api/group/del', { id })
}

export const groupService = {
  getGroupTree,
  getGroupList,
  addGroup,
  deleteGroup,
}
