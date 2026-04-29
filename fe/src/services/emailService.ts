/**
 * 邮件 API 服务层
 *
 * 封装所有邮件相关的 HTTP 请求，将 API 调用从视图/组件中解耦。
 * 视图层只需调用本模块提供的方法，无需关心具体的请求路径和参数格式。
 *
 * 创建日期: 20260429
 * @module emailService
 */

import { http } from '@/utils/axios'
import type { EmailListResponse, EmailDetail, EmailAttachment } from '@/types/api'

/** 邮件列表查询参数 */
interface EmailListParams {
  tag: string
  page_size: number
  current_page?: number
  keyword?: string
}

/** 邮件发送参数 */
interface EmailSendParams {
  from: { name: string; email: string }
  to: { name: string; email: string }[]
  cc: { name: string; email: string }[]
  bcc: { name: string; email: string }[]
  subject: string
  text: string
  html: string
  attrs: { name: string; data: string | ArrayBuffer | null }[]
}

/** 邮件移动参数 */
interface EmailMoveParams {
  group_id: string
  group_name: string
  ids: number[]
}

/** 邮件删除参数 */
interface EmailDeleteParams {
  ids: number[]
  forcedDel: boolean
}

/**
 * 获取邮件列表
 * @param params - 查询参数（tag、page_size、current_page、keyword）
 * @returns 邮件列表响应（list + total_page）
 */
function getEmailList(params: EmailListParams) {
  return http.post<EmailListResponse>('/api/email/list', params)
}

/**
 * 获取邮件详情
 * @param id - 邮件 ID
 * @returns 邮件详情数据
 */
function getEmailDetail(id: number) {
  return http.post<EmailDetail>('/api/email/detail', { id })
}

/**
 * 发送邮件
 * @param params - 邮件发送参数（from、to、cc、bcc、subject、text、html、attrs）
 * @returns 发送结果
 */
function sendEmail(params: EmailSendParams) {
  return http.post('/api/email/send', params)
}

/**
 * 删除邮件
 * @param params - 删除参数（ids、forcedDel）
 * @returns 删除结果
 */
function deleteEmails(params: EmailDeleteParams) {
  return http.post('/api/email/del', params)
}

/**
 * 标记邮件已读
 * @param ids - 邮件 ID 数组
 * @returns 操作结果
 */
function markEmailsRead(ids: number[]) {
  return http.post('/api/email/read', { ids })
}

/**
 * 移动邮件到指定分组
 * @param params - 移动参数（group_id、group_name、ids）
 * @returns 移动结果
 */
function moveEmails(params: EmailMoveParams) {
  return http.post('/api/email/move', params)
}

export const emailService = {
  getEmailList,
  getEmailDetail,
  sendEmail,
  deleteEmails,
  markEmailsRead,
  moveEmails,
}
