/**
 * 日期格式化工具函数
 *
 * 提取自 ListView.vue（formatShortDate）和 EmailDetailView.vue（formatDetailDate），
 * 消除两个视图中重复的日期格式化逻辑。
 *
 * 创建日期: 20260429
 */

/**
 * 短日期格式化（用于邮件列表）
 * - 今天的邮件仅显示时间 (HH:MM)
 * - 非今天的邮件显示月日 (Mon DD)
 *
 * @param dateStr ISO 日期字符串
 * @returns 格式化后的日期字符串
 */
export function formatShortDate(dateStr: string): string {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  const now = new Date();

  // 判断是否为今天（同年同月同日）
  const isToday =
    d.getFullYear() === now.getFullYear() &&
    d.getMonth() === now.getMonth() &&
    d.getDate() === now.getDate();

  if (isToday) {
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  }

  return d.toLocaleDateString([], { month: "short", day: "numeric" });
}

/**
 * 详细日期格式化（用于邮件详情页）
 * - 显示完整日期和时间
 *
 * @param dateStr ISO 日期字符串
 * @returns 格式化后的日期字符串，包含年月日和时分
 */
export function formatDetailDate(dateStr: string): string {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  return d.toLocaleString([], {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
