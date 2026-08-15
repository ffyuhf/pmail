/**
 * 分组（文件夹）API 服务
 *
 * 修改日期: 20260609
 * 修改原因: 对齐旧前端 fe/src/services/groupService.ts 和后端 Go 控制器协议：
 *   - 删除不存在的 updateGroup（后端 http_server.go 无 /api/group/update 路由）
 *   - createGroup parentId 改为 parentId?: number 保持接口一致
 * @module groupService
 */
import { http } from '@/utils/axios'
import type { GroupItem, GroupListItem } from '@/types/api'

/**
 * 获取分组树形结构
 * 用于左侧边栏文件夹导航（后端 GET /api/group）
 * @returns 树形分组列表（含 children）
 */
export function getGroups() {
  return http.get('/api/group')
}

/**
 * 获取分组平铺列表
 * 用于规则设置中的"移动到文件夹"下拉框（后端 POST /api/group/list）
 * @returns 平铺分组列表（id + name）
 */
export function getGroupList() {
  return http.post('/api/group/list')
}

/**
 * 创建新分组
 * @param name - 分组名称
 * @param parentId - 父分组 ID（可选，顶级分组不传）
 */
export function createGroup(name: string, parentId?: number) {
  return http.post('/api/group/add', { name, parent_id: parentId })
}

/**
 * 删除分组
 * @param id - 分组 ID
 */
export function deleteGroup(id: number) {
  return http.post('/api/group/del', { id })
}
