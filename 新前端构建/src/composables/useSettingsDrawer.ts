/**
 * 设置面板抽屉状态 Composable
 *
 * 从旧前端 fe/src/composables/useSettingsDrawer.ts 迁移。
 * 管理设置侧滑面板的显示/隐藏和当前激活的设置项。
 *
 * 创建日期: 20260609
 */
import { ref } from 'vue'

/** 设置面板可用的设置项类型 */
export type SettingsTab = 'security' | 'groups' | 'rules' | 'users' | 'plugins'

export function useSettingsDrawer() {
  /** 设置面板是否可见 */
  const visible = ref(false)
  /** 当前激活的设置项 */
  const activeTab = ref<SettingsTab>('security')

  /** 打开设置面板 */
  function openSettings(tab?: SettingsTab) {
    if (tab) activeTab.value = tab
    visible.value = true
  }

  /** 关闭设置面板 */
  function closeSettings() {
    visible.value = false
  }

  return { visible, activeTab, openSettings, closeSettings }
}