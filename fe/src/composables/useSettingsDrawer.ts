/**
 * 设置抽屉 Composable
 *
 * 提取 HomeHeader 和 HomeAside 中重复的"打开设置抽屉"逻辑。
 * 自动检测用户信息是否已加载：若未加载则先初始化再打开，否则直接打开。
 *
 * 创建日期: 20260429
 * @module useSettingsDrawer
 */

import { useGlobalStatusStore } from '@/stores/useGlobalStatusStore'

/**
 * 设置抽屉开关逻辑
 * @returns openSettings - 打开设置抽屉的方法
 */
export function useSettingsDrawer() {
  const globalStatus = useGlobalStatusStore()

  /**
   * 打开设置抽屉
   * 若用户信息未加载，先调用 init 获取用户信息后再打开
   */
  const openSettings = () => {
    if (Object.keys(globalStatus.userInfos).length === 0) {
      globalStatus.init(() => {
        globalStatus.settingsDrawerVisible = true
      })
    } else {
      globalStatus.settingsDrawerVisible = true
    }
  }

  return { openSettings }
}
