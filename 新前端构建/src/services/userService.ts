/**
 * 用户 API 服务
 *
 * 修改日期: 20260609
 * 修改原因: 对齐旧前端 fe/src/services/userService.ts 和后端 Go 控制器协议：
 *   - logout 端点修正：/api/login→/api/logout（后端 http_server.go 第36行）
 *   - changePassword 参数修正：new_password→password（后端 settings.go 第16行）
 *   - getUserList 参数修正：page→current_page, page_size（后端 user.go 第64-66行）
 *   - createUser 参数修正：name→username, is_admin→disabled（后端 user.go 第19-25行）
 *   - 删除不存在的 deleteUser（后端无 /api/user/del 路由）
 * @module userService
 */
import { http } from '@/utils/axios'

/**
 * 用户登录
 * @param account - 账号
 * @param password - 密码
 */
export function login(account: string, password: string) {
  return http.post('/api/login', { account, password })
}

/**
 * 用户注销
 * 后端独立端点 POST /api/logout（http_server.go 第36行）
 */
export function logout() {
  return http.post('/api/logout', {})
}

/**
 * 获取当前用户信息
 * 后端端点 POST /api/user/info（http_server.go 第56行）
 */
export function getUserInfo() {
  return http.post('/api/user/info', {})
}

/**
 * 修改当前用户密码
 * @param password - 新密码（后端 settings.go 期望字段名为 password）
 */
export function changePassword(password: string) {
  return http.post('/api/settings/modify_password', { password })
}

/**
 * 获取用户列表（管理员）
 * @param currentPage - 当前页码（从 1 开始）
 * @param pageSize - 每页条数
 */
export function getUserList(currentPage: number, pageSize: number) {
  return http.post('/api/user/list', { current_page: currentPage, page_size: pageSize })
}

/**
 * 创建新用户（管理员）
 * @param data - 用户数据（对齐后端 userCreateRequest 结构体）
 */
export function createUser(data: {
  account: string
  username: string
  password: string
  disabled: number
}) {
  return http.post('/api/user/create', data)
}

/**
 * 编辑已有用户（管理员）
 * @param data - 用户数据（对齐后端 userCreateRequest，通过 account 或 id 定位）
 */
export function updateUser(data: {
  id?: number
  account?: string
  username?: string
  password?: string
  disabled?: number
}) {
  return http.post('/api/user/edit', data)
}
