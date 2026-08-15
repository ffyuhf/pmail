/**
 * 前端全局常量定义
 *
 * 提取自 ListView.vue 和 EmailDetailView.vue 中重复的默认分组 Tag 字符串。
 * 统一定义，避免魔法字符串散落在各处。
 *
 * 创建日期: 20260429
 */

/**
 * 默认分组 Tag：对应"全部邮件/收件箱"视图
 *
 * type=0 表示按状态筛选，status=-1 表示不限制状态
 * 原始值: '{"type":0,"status":-1}'
 */
export const DEFAULT_GROUP_TAG = '{"type":0,"status":-1}';
