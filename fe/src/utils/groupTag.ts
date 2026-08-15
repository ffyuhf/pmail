/**
 * 分组 Tag 解析工具函数
 *
 * 提取自 ListView.vue 和 EmailDetailView.vue 中重复的 group tag JSON 解析逻辑。
 * 统一处理默认值填充和 tag 解析，避免魔法字符串和重复的 try-catch。
 *
 * 创建日期: 20260429
 */

import { DEFAULT_GROUP_TAG } from './constants'

/** 解析后的分组 Tag 结构 */
interface GroupTag {
  type: number
  status: number
}

/**
 * 规范化 tag 值：空字符串时回退到默认分组 Tag
 * @param tag - 原始 tag 字符串（可能为空）
 * @returns 有效的 tag 字符串
 */
export function normalizeTag(tag: string): string {
  return tag || DEFAULT_GROUP_TAG
}

/**
 * 解析分组 Tag JSON 字符串为结构化对象
 * @param tag - JSON 格式的 tag 字符串
 * @returns 解析后的 GroupTag 对象，解析失败时返回默认值
 */
export function parseGroupTag(tag: string): GroupTag {
  try {
    return JSON.parse(normalizeTag(tag))
  } catch {
    return { type: 0, status: -1 }
  }
}

/**
 * 判断分组 Tag 是否对应"已删除/垃圾箱"分组（status === 3）
 * @param tag - JSON 格式的 tag 字符串
 * @returns true 表示该分组下的邮件应强制删除
 */
export function isTrashGroup(tag: string): boolean {
  return parseGroupTag(tag).status === 3
}
