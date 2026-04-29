/**
 * 初始化向导 API 服务层
 *
 * 封装系统初始化设置（Setup）相关的 HTTP 请求。
 * 涵盖数据库、密码、域名、DNS、SSL 等配置步骤的读取与保存。
 *
 * 创建日期: 20260429
 * @module setupService
 */

import { http } from '@/utils/axios'

/** Setup 通用请求参数 */
interface SetupParams {
  action: 'get' | 'set' | 'getParams'
  step: string
  token: string
  [key: string]: unknown
}

/**
 * 通用 Setup API 调用
 * 所有初始化步骤共享相同的端点，通过 action + step 区分操作
 * @param params - 请求参数（action、step、token 及步骤特有字段）
 * @returns 配置数据或操作结果
 */
function setupRequest(params: SetupParams) {
  return http.post('/api/setup', params)
}

/**
 * 获取数据库配置
 * @param token - Setup 鉴权 Token
 * @returns 数据库类型和 DSN
 */
function getDatabaseConfig(token: string) {
  return setupRequest({ action: 'get', step: 'database', token })
}

/**
 * 保存数据库配置
 * @param token - Setup 鉴权 Token
 * @param dbType - 数据库类型（mysql/sqlite/postgres）
 * @param dbDsn - 数据库连接字符串
 * @returns 保存结果
 */
function setDatabaseConfig(token: string, dbType: string, dbDsn: string) {
  return setupRequest({ action: 'set', step: 'database', db_type: dbType, db_dsn: dbDsn, token })
}

/**
 * 获取密码配置状态
 * @param token - Setup 鉴权 Token
 * @returns 管理员账号信息（已设置时返回账号名）
 */
function getPasswordConfig(token: string) {
  return setupRequest({ action: 'get', step: 'password', token })
}

/**
 * 保存管理员密码
 * @param token - Setup 鉴权 Token
 * @param account - 管理员账号
 * @param password - 管理员密码
 * @returns 保存结果
 */
function setPassword(token: string, account: string, password: string) {
  return setupRequest({ action: 'set', step: 'password', account, password, token })
}

/**
 * 获取域名配置
 * @param token - Setup 鉴权 Token
 * @returns 域名信息（web_domain、smtp_domain、domains）
 */
function getDomainConfig(token: string) {
  return setupRequest({ action: 'get', step: 'domain', token })
}

/**
 * 保存域名配置
 * @param token - Setup 鉴权 Token
 * @param webDomain - Web 域名
 * @param smtpDomain - SMTP 域名
 * @param multiDomain - 附加域名（逗号分隔）
 * @returns 保存结果
 */
function setDomainConfig(token: string, webDomain: string, smtpDomain: string, multiDomain: string) {
  return setupRequest({
    action: 'set', step: 'domain',
    web_domain: webDomain, smtp_domain: smtpDomain, multi_domain: multiDomain, token,
  })
}

/**
 * 获取 DNS 配置
 * @param token - Setup 鉴权 Token
 * @returns DNS 记录信息（按域名分组）
 */
function getDnsConfig(token: string) {
  return setupRequest({ action: 'get', step: 'dns', token })
}

/**
 * 获取 SSL 配置
 * @param token - Setup 鉴权 Token
 * @returns SSL 类型、端口等信息
 */
function getSslConfig(token: string) {
  return setupRequest({ action: 'get', step: 'ssl', token })
}

/**
 * 保存 SSL 配置
 * @param token - Setup 鉴权 Token
 * @param sslType - SSL 类型（0=自动、1=手动、2=自动DNS）
 * @param keyPath - SSL key 文件路径（手动模式）
 * @param crtPath - SSL crt 文件路径（手动模式）
 * @returns 保存结果
 */
function setSslConfig(token: string, sslType: string, keyPath: string, crtPath: string) {
  return setupRequest({
    action: 'set', step: 'ssl',
    ssl_type: sslType, key_path: keyPath, crt_path: crtPath, token,
  })
}

/**
 * 获取 SSL DNS 验证参数（DNS 挑战模式）
 * @param token - Setup 鉴权 Token
 * @returns DNS 验证记录列表
 */
function getSslDnsParams(token: string) {
  return setupRequest({ action: 'getParams', step: 'ssl', token })
}

export const setupService = {
  setupRequest,
  getDatabaseConfig,
  setDatabaseConfig,
  getPasswordConfig,
  setPassword,
  getDomainConfig,
  setDomainConfig,
  getDnsConfig,
  getSslConfig,
  setSslConfig,
  getSslDnsParams,
}
