/**
 * 插件 API 服务层
 *
 * 封装所有插件相关的 HTTP 请求。
 * 提供插件列表获取操作。
 *
 * 创建日期: 20260429
 * @module pluginService
 */

import { http } from '@/utils/axios'

/**
 * 获取已安装插件列表
 * @returns 插件名称数组
 */
function getPluginList() {
  return http.get<string[]>('/api/plugin/list')
}

export const pluginService = {
  getPluginList,
}
