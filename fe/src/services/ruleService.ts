/**
 * 规则 API 服务层
 *
 * 封装所有邮件收信规则相关的 HTTP 请求。
 * 提供规则的查询、新增、更新和删除操作。
 *
 * 创建日期: 20260429
 * @module ruleService
 */

import { http } from '@/utils/axios'
import type { RuleItem } from '@/types/api'

/**
 * 获取所有规则列表
 * @returns 规则数组
 */
function getRules() {
  return http.post<RuleItem[]>('/api/rule/get')
}

/**
 * 新增规则
 * @param rule - 规则数据（name、sort、rules、action、params）
 * @returns 新增结果
 */
function addRule(rule: Omit<RuleItem, 'id'>) {
  return http.post('/api/rule/add', rule)
}

/**
 * 更新已有规则
 * @param rule - 规则数据（含 id）
 * @returns 更新结果
 */
function updateRule(rule: RuleItem) {
  return http.post('/api/rule/update', rule)
}

/**
 * 删除规则
 * @param id - 规则 ID
 * @returns 删除结果
 */
function deleteRule(id: number) {
  return http.post('/api/rule/del', { id })
}

export const ruleService = {
  getRules,
  addRule,
  updateRule,
  deleteRule,
}
