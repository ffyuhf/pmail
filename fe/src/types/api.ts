/**
 * API 响应通用类型定义
 * 用于规范前后端数据交互接口
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
  /** 邮件类型：0=接收, 1=发送。新增日期: 20260510 — 自定义文件夹需区分发送/接收邮件 */
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
