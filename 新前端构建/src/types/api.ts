/**
 * API 响应通用类型定义
 *
 * 从旧前端 fe/src/types/api.ts 迁移，用于规范前后端数据交互接口。
 *
 * 创建日期: 20260609
 */

/** 通用 API 响应结构 */
export interface ApiResponse<T = unknown> {
  errorNo: number;
  errorMsg: string;
  data: T;
}

/** 用户信息 */
export interface UserInfo {
  account: string;
  name: string;
  domains: string[];
  is_admin: boolean;
  [key: string]: unknown;
}

/** 邮件联系人 */
export interface EmailContact {
  Name: string;
  EmailAddress: string;
}

/** 邮件列表项 */
export interface EmailListItem {
  id: number;
  /** user_email 表主键 ID，用于精确操作（删除/移动） */
  ue_id: number;
  /** 邮件类型：0=接收, 1=发送 */
  type: number;
  sender: EmailContact;
  title: string;
  desc: string;
  datetime: string;
  is_read: boolean;
  dangerous: boolean;
  error: string;
}

/** 邮件列表响应 */
export interface EmailListResponse {
  list: EmailListItem[];
  total_page: number;
}

/** 邮件附件 */
export interface EmailAttachment {
  Index: number;
  Filename: string;
}

/** 邮件详情 */
export interface EmailDetail {
  id: number;
  /** user_email 表主键 ID */
  ue_id: number;
  /** 邮件类型：0=接收, 1=发送 */
  type: number;
  subject: string;
  from_name: string;
  from_address: string;
  to: string;
  cc: string;
  send_date: string;
  html: string;
  text: string;
  attachments: EmailAttachment[];
  [key: string]: unknown;
}

/** 分组/文件夹项（树结构） */
export interface GroupItem {
  id: number;
  label: string;
  tag: string;
  children?: GroupItem[];
  parent_id?: number;
}

/** 分组列表项（平铺结构） */
export interface GroupListItem {
  id: string;
  name: string;
}

/** 规则条件 */
export interface RuleCondition {
  field: string;
  type: string;
  rule: string;
}

/** 规则项 */
export interface RuleItem {
  id: number;
  name: string;
  action: number;
  params: string;
  sort: number;
  rules: RuleCondition[];
}

/** 国际化语言键类型 */
export interface LangMessages {
  [key: string]: string;
}