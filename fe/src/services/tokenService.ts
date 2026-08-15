/**
 * API Token 服务层
 *
 * 封装 API Token 的生成与撤销操作。
 * 生成的 Token 供第三方客户端通过 HTTP Header "Token" 调用 PMail API 时认证使用。
 *
 * 创建日期: 20260815
 * @module tokenService
 */

import { http } from '@/utils/axios'

/**
 * 生成 API Token
 * @param expiresIn - 有效期（秒），0 表示永不过期
 * @returns 生成的 token 及过期时间
 */
function generateToken(expiresIn: number) {
  return http.post('/api/token/generate', { expires_in: expiresIn })
}

/**
 * 撤销 API Token（仅可撤销当前用户自己的 Token）
 * @param token - 要撤销的 Token 字符串
 * @returns 撤销结果
 */
function revokeToken(token: string) {
  return http.post('/api/token/revoke', { token })
}

export const tokenService = {
  generateToken,
  revokeToken,
}
