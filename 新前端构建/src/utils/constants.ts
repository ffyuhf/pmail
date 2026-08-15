/**
 * 前端全局常量定义
 *
 * 从旧前端 fe/src/utils/constants.ts 迁移。
 * 默认分组 Tag 字符串，统一管理避免魔法字符串。
 *
 * 修改日期: 20260609
 * 修改原因: 新增默认文件夹映射表和匹配函数，解决：
 *   1. 默认文件夹名称国际化（前端替换显示名称，实际文件名不变）
 *   2. 最左栏硬编码文件夹与后端返回的默认文件夹重复显示
 *   后端 tag JSON 可能包含额外字段（如 group_id），需解析 type+status 匹配。
 *
 * 创建日期: 20260609
 */

/**
 * 默认分组 Tag：对应"收件箱"视图
 *
 * 必须包含 group_id:0，与后端 group.go GetUserGroup 返回的收件箱 tag 完全一致。
 * 缺少 group_id 时，Go 的 json.Unmarshal 会将 GroupId 解析为零值 0（而非 -1），
 * 虽然大多数场景结果相同，但保持一致性可避免潜在的边界问题。
 *
 * 修改日期: 20260609
 * 修改原因: 添加 group_id:0 字段，对齐后端 SearchTag{Type:0, Status:-1, GroupId:0}
 */
export const DEFAULT_GROUP_TAG = '{"type":0,"status":-1,"group_id":0}'

/**
 * 默认文件夹定义项
 *
 * 匹配规则：解析 tag JSON 后比较 type + status，忽略 group_id 等额外字段。
 * 后端 group.go GetUserGroup 返回的默认文件夹 tag 结构：
 *   - 收件箱: {type:0, status:-1, group_id:0}
 *   - 发件箱: {type:1, status:-1}
 *   - 草稿箱: {type:0, status:4}
 *   - 垃圾箱: {type:-1, status:5}
 *   - 已删除: {type:-1, status:3}
 */
export interface DefaultFolderDef {
  /** 匹配条件：tag 解析后的 type 值 */
  type: number
  /** 匹配条件：tag 解析后的 status 值 */
  status: number
  /** i18n 国际化 key（对应 i18n/index.ts 中的 key） */
  i18nKey: string
  /** 图标（Emoji，用于 IconNav 左侧导航栏） */
  icon: string
}

/**
 * 默认文件夹列表（5 个系统文件夹）
 *
 * 对齐后端 group.go GetUserGroup 中的硬编码默认文件夹。
 * 前端渲染时用 i18n 替换后端返回的 label，实现国际化。
 */
export const DEFAULT_FOLDERS: DefaultFolderDef[] = [
  { type: 0, status: -1, i18nKey: 'inbox', icon: '📥' },
  { type: 1, status: -1, i18nKey: 'outbox', icon: '📤' },
  { type: 0, status: 4, i18nKey: 'sketch', icon: '📝' },
  { type: -1, status: 5, i18nKey: 'junk', icon: '🗑' },
  { type: -1, status: 3, i18nKey: 'deleted', icon: '🗑' },
]

/**
 * 解析 tag JSON 字符串为 {type, status} 对象
 *
 * 安全解析：JSON 格式错误时返回 null。
 * 忽略 group_id 等额外字段，仅提取 type 和 status。
 *
 * @param tag - 后端返回的 tag JSON 字符串
 * @returns 解析后的对象，或 null
 */
export function parseTagFields(tag: string): { type: number; status: number } | null {
  try {
    const obj = JSON.parse(tag)
    if (typeof obj.type === 'number' && typeof obj.status === 'number') {
      return { type: obj.type, status: obj.status }
    }
  } catch {
    // JSON 解析失败，非标准 tag
  }
  return null
}

/**
 * 判断 tag 是否属于默认文件夹
 *
 * 通过解析 tag JSON 的 type + status 字段与 DEFAULT_FOLDERS 匹配，
 * 避免字符串精确比较（后端可能附加 group_id 等字段导致不匹配）。
 *
 * @param tag - 后端返回的 tag JSON 字符串
 * @returns true 表示该 tag 是系统默认文件夹
 */
export function isDefaultGroupTag(tag: string): boolean {
  const fields = parseTagFields(tag)
  if (!fields) return false
  return DEFAULT_FOLDERS.some((f) => f.type === fields.type && f.status === fields.status)
}

/**
 * 获取默认文件夹的 i18n key
 *
 * @param tag - 后端返回的 tag JSON 字符串
 * @returns 对应的 i18n key，非默认文件夹返回 null
 */
export function getDefaultFolderI18nKey(tag: string): string | null {
  const fields = parseTagFields(tag)
  if (!fields) return null
  const match = DEFAULT_FOLDERS.find((f) => f.type === fields.type && f.status === fields.status)
  return match ? match.i18nKey : null
}