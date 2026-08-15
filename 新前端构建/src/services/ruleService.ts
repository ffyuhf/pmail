/**
 * 规则 API 服务
 *
 * 从旧前端 fe/src/services/ruleService.ts 迁移。
 * 封装收信规则的查询、新增、更新、删除操作。
 *
 * 后端协议：所有接口均为 POST，端点格式 /api/rule/{action}。
 *
 * 修改日期: 20260609
 * 修改原因: 旧版使用 RESTful 风格（GET /api/rules、PUT /api/rules/:id、DELETE /api/rules/:id），
 *           后端实际使用 POST /api/rule/get、POST /api/rule/add 等风格，导致所有请求 404。
 * @module ruleService
 */
import { http } from '@/utils/axios'
import type { RuleItem } from '@/types/api'

/**
 * 获取规则列表
 * @returns 所有规则（含条件和动作）
 */
export function getRules() {
  return http.post('/api/rule/get')
}

/**
 * 创建新规则
 * @param data - 规则数据（不含 id）
 */
export function createRule(data: Omit<RuleItem, 'id'>) {
  return http.post('/api/rule/add', data)
}

/**
 * 更新规则
 * @param data - 规则数据（含 id 和需要更新的字段）
 */
export function updateRule(data: Partial<RuleItem> & { id: number }) {
  return http.post('/api/rule/update', data)
}

/**
 * 删除规则
 * @param id - 规则 ID
 */
export function deleteRule(id: number) {
  return http.post('/api/rule/del', { id })
}