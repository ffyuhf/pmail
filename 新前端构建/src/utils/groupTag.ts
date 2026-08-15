/**
 * 分组 Tag 解析工具函数
 *
 * 从旧前端 fe/src/utils/groupTag.ts 迁移。
 *
 * 创建日期: 20260609
 */

import { DEFAULT_GROUP_TAG } from './constants'

/** 解析后的分组 Tag 结构 */
interface GroupTag {
  type: number
  status: number
}

/** 规范化 tag 值：空字符串时回退到默认分组 Tag */
export function normalizeTag(tag: string): string {
  return tag || DEFAULT_GROUP_TAG
}

/** 解析分组 Tag JSON 字符串为结构化对象 */
export function parseGroupTag(tag: string): GroupTag {
  try {
    return JSON.parse(normalizeTag(tag))
  } catch {
    return { type: 0, status: -1 }
  }
}

/** 判断分组 Tag 是否对应"已删除/垃圾箱"分组（status === 3） */
export function isTrashGroup(tag: string): boolean {
  return parseGroupTag(tag).status === 3
}