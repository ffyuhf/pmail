/**
 * Axios HTTP 客户端实例
 *
 * 修改日期: 20260609
 * 修改原因: 对齐旧前端 fe/src/utils/axios.ts 拦截器逻辑：
 *   - 补充 POST 请求自动 JSON.stringify body
 *   - 响应成功时解包 response.data（对齐旧前端 return response.data）
 *   - 使用 vue-router 进行路由跳转替代 window.location.hash
 *   - 补充 403/402 HTTP 状态码的响应拦截
 * @module axios
 */
import axios from 'axios'
import { router } from '@/router'
import type { ApiResponse } from '@/types/api'

/* 创建 axios 实例，不设 baseURL，使用相对路径同源请求 */
const http = axios.create({
  baseURL: import.meta.env.VITE_APP_URL || '',
  timeout: 60000,
  headers: {
    'Content-Type': 'application/json;charset=UTF-8;',
  },
})

/**
 * 请求拦截器
 * POST 请求自动将 data 转为 JSON 字符串（对齐旧前端）
 */
http.interceptors.request.use(
  (config) => {
    if (config.method === 'POST') {
      config.data = JSON.stringify(config.data)
    }
    return config
  },
  (error) => Promise.reject(error),
)

/**
 * 响应拦截器
 * - 成功时解包 response.data（对齐旧前端）
 * - errorNo 403 → 跳转登录页
 * - errorNo 402 → 跳转 setup 引导页（排除已在 /setup 的情况）
 * - HTTP 状态码 403 → 跳转登录页
 */
http.interceptors.response.use(
  async (response): Promise<any> => {
    /* 响应成功，检查业务错误码 */
    if (response.data.errorNo === 403) {
      await router.replace({
        path: '/login',
        query: { redirect: router.currentRoute.value.fullPath },
      })
    }
    /* 402: 系统未初始化，跳转 setup（排除已在该页的情况，避免循环重定向） */
    if (response.data.errorNo === 402 && router.currentRoute.value.path !== '/setup') {
      await router.replace({
        path: '/setup',
        query: { redirect: router.currentRoute.value.fullPath },
      })
    }
    return response.data
  },
  async (error) => {
    if (error.response && error.response.status) {
      switch (error.response.status) {
        case 403:
          await router.replace({
            path: '/login',
            query: { redirect: router.currentRoute.value.fullPath },
          })
          break
      }
      return Promise.reject(error)
    }
    return Promise.reject(error)
  },
)

export { http }
