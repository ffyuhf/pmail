/**
 * 日期格式化工具函数
 *
 * 从旧前端 fe/src/utils/dateFormat.ts 迁移。
 * 提供短日期（列表用）和详细日期（详情页用）两种格式化。
 *
 * 创建日期: 20260609
 */

/** 短日期格式化（邮件列表）：今天显示时间，非今天显示月日 */
export function formatShortDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const isToday =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate()
  if (isToday) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  }
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

/** 详细日期格式化（邮件详情页）：完整日期和时间 */
export function formatDetailDate(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleString([], {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}