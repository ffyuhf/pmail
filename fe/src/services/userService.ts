/**
 * 用户 API 服务层
 *
 * 封装所有用户相关的 HTTP 请求，包括登录、注销、用户信息获取、
 * 密码修改、用户列表查询、用户创建/编辑等操作。
 *
 * 创建日期: 20260429
 * @module userService
 */

import { http } from '@/utils/axios'
import type { UserInfo } from '@/types/api'

/**
 * 用户登录
 * @param account - 账号
 * @param password - 密码
 * @returns 登录结果（含用户信息）
 */
function login(account: string, password: string) {
  return http.post<UserInfo>('/api/login', { account, password })
}

/**
 * 用户注销
 * @returns 注销结果
 */
function logout() {
  return http.post('/api/logout', {})
}

/**
 * 获取当前用户信息
 * @returns 用户信息（account、name、domains、is_admin 等）
 */
function getUserInfo() {
  return http.post<UserInfo>('/api/user/info', {})
}

/**
 * 修改当前用户密码
 * @param password - 新密码
 * @returns 修改结果
 */
function modifyPassword(password: string) {
  return http.post('/api/settings/modify_password', { password })
}

/**
 * 获取用户列表（分页）
 * @param currentPage - 当前页码
 * @param pageSize - 每页条数
 * @returns 用户列表及分页信息
 */
function getUserList(currentPage: number, pageSize: number) {
  return http.post('/api/user/list', { current_page: currentPage, page_size: pageSize })
}

/**
 * 创建新用户
 * @param data - 用户数据（account、username、password、disabled）
 * @returns 创建结果
 */
function createUser(data: { account: string; username: string; password: string; disabled: number }) {
  return http.post('/api/user/create', data)
}

/**
 * 编辑已有用户
 * @param data - 用户数据（account、username、disabled、password 可选）
 * @returns 编辑结果
 */
function editUser(data: { account: string; username: string; disabled: number; password?: string }) {
  return http.post('/api/user/edit', data)
}

export const userService = {
  login,
  logout,
  getUserInfo,
  modifyPassword,
  getUserList,
  createUser,
  editUser,
}
