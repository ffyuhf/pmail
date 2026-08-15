/**
 * 邮件 API 服务
 *
 * 修改日期: 20260609
 * 修改原因: 对齐旧前端 fe/src/services/emailService.ts 和后端 Go 控制器协议：
 *   - getEmailList 参数名修正：groupTag→tag, page→current_page, pageSize→page_size
 *   - sendEmail 从 FormData 改为 JSON body（后端控制器 send.go 解析的是 JSON）
 *   - getAttachmentUrl 修正为 /attachments/download/（后端路由在 http_server.go 第53行）
 *   - 移除不存在的 /api/email/attachment 端点
 * @module emailService
 */
import { http } from '@/utils/axios'

/** 邮件列表查询参数 */
export interface EmailListParams {
  /** 分组标签（SearchTag JSON字符串，如 '{"type":0,"status":-1}'） */
  tag: string
  /** 每页显示条数 */
  page_size: number
  /** 当前页码（从 1 开始） */
  current_page?: number
  /** 搜索关键词（可选） */
  keyword?: string
}

/** 邮件发送参数（对齐后端 sendRequest 结构体） */
export interface EmailSendParams {
  from: { name: string; email: string }
  to: { name: string; email: string }[]
  cc: { name: string; email: string }[]
  bcc: { name: string; email: string }[]
  subject: string
  text: string
  html: string
  /** 附件：name + base64 data */
  attrs: { name: string; data: string }[]
}

/**
 * 获取邮件列表
 * @param params - 查询参数（tag、page_size、current_page、keyword）
 */
export function getEmailList(params: EmailListParams) {
  return http.post('/api/email/list', params)
}

/**
 * 获取邮件详情
 * @param id - 邮件 ID
 */
export function getEmailDetail(id: number) {
  return http.post('/api/email/detail', { id })
}

/**
 * 发送邮件（JSON body，含 base64 编码附件）
 * 对齐后端 sendRequest 结构体：from/to/cc/bcc/subject/text/html/attrs
 * @param params - 邮件发送参数
 */
export function sendEmail(params: EmailSendParams) {
  return http.post('/api/email/send', params)
}

/**
 * 删除邮件（支持批量）
 * @param ids - email.id 数组（向后兼容，IMAP 使用）
 * @param ueIds - user_email.id 数组（Web 前端精确匹配）
 * @param forcedDel - 是否强制删除（垃圾箱中彻底删除），默认 false
 */
export function deleteEmails(ids: number[], ueIds?: number[], forcedDel?: boolean) {
  return http.post('/api/email/del', { ids, ue_ids: ueIds, forcedDel })
}

/**
 * 标记邮件已读
 * @param ids - email.id 数组
 */
export function markEmailsRead(ids: number[]) {
  return http.post('/api/email/read', { ids })
}

/**
 * 移动邮件到指定分组
 * 对齐后端 moveRequest 结构体：group_id + group_name + ids + ue_ids
 * @param ids - email.id 数组（向后兼容，IMAP 使用）
 * @param ueIds - user_email.id 数组（Web 前端精确匹配）
 * @param groupId - 目标分组 ID
 * @param groupName - 目标分组名称（对齐 moveRequest.GroupName 字段）
 */
export function moveEmails(ids: number[], ueIds: number[], groupId: number, groupName: string) {
  return http.post('/api/email/move', { ids, ue_ids: ueIds, group_id: groupId, group_name: groupName })
}

/**
 * 下载附件 URL
 * 后端 attachments.go 通过 strings.Split(req.RequestURI, "/") 解析路径，
 * 期望格式: /attachments/download/{emailId}/{index}（共 5 段）
 * 旧前端 EmailDetailView.vue 同样使用此路径格式
 * @param emailId - 邮件 ID
 * @param attachmentIndex - 附件索引
 * @returns 附件下载 URL（路径参数格式）
 */
export function getAttachmentUrl(emailId: number, attachmentIndex: number): string {
  return `/attachments/download/${emailId}/${attachmentIndex}`
}
