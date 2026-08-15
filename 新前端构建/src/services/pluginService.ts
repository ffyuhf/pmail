/**
 * 插件 API 服务
 *
 * 从旧前端 fe/src/services/pluginService.ts 迁移。
 * 封装插件的查询接口。
 *
 * 后端协议：GET /api/plugin/list
 *
 * 修改日期: 20260609
 * 修改原因: 旧版端点为 GET /api/plugins（复数），后端实际为 GET /api/plugin/list，
 *           且旧版新增了不存在的 updatePlugin/togglePlugin 接口（后端仅支持列表查询）。
 * @module pluginService
 */
import { http } from '@/utils/axios'

/**
 * 获取插件列表
 * @returns 所有插件及其状态
 */
export function getPlugins() {
  return http.get('/api/plugin/list')
}
