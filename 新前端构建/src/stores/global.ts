/**
 * 全局状态 Store
 *
 * 从旧前端 fe/src/stores/useGlobalStatusStore.ts 迁移。
 * 管理用户信息、登录状态等全局共享状态。
 *
 * 创建日期: 20260609
 */
import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { UserInfo } from '@/types/api'
import { getUserInfo } from '@/services/userService'

export const useGlobalStore = defineStore('global', () => {
  /** 当前登录用户信息 */
  const userInfo = ref<UserInfo | null>(null)
  /** 是否已登录 */
  const isLoggedIn = ref(false)
  /** 是否为管理员 */
  const isAdmin = ref(false)

  /** 从后端获取用户信息 */
  async function fetchUserInfo() {
    try {
      const res: any = await getUserInfo()
      /* axios 拦截器已解包，直接读 errorNo */
      if (res.errorNo === 0) {
        userInfo.value = res.data
        isLoggedIn.value = true
        isAdmin.value = res.data?.is_admin || false
      }
    } catch {
      isLoggedIn.value = false
      userInfo.value = null
    }
  }

  /** 清除登录状态（注销时调用） */
  function clearUser() {
    userInfo.value = null
    isLoggedIn.value = false
    isAdmin.value = false
  }

  return { userInfo, isLoggedIn, isAdmin, fetchUserInfo, clearUser }
})