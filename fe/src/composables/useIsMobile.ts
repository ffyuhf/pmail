/**
 * 移动端检测 Composable
 *
 * 提取自 HomeHeader.vue 中重复的 resize 监听 + 断点判断逻辑。
 * 自动管理事件监听器的注册与清理，避免每个组件重复编写 onMounted/onUnmounted。
 *
 * 创建日期: 20260429
 * @module useIsMobile
 */

import { ref, onMounted, onUnmounted } from 'vue'

/**
 * 响应式移动端检测
 * @param breakpoint - 移动端断点宽度（默认 768px）
 * @returns isMobile - 响应式的布尔值，true 表示当前视口宽度 <= breakpoint
 */
export function useIsMobile(breakpoint = 768) {
  const isMobile = ref(window.innerWidth <= breakpoint)

  const handleResize = () => {
    isMobile.value = window.innerWidth <= breakpoint
  }

  onMounted(() => {
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', handleResize)
  })

  return { isMobile }
}
