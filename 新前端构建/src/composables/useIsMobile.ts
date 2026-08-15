/**
 * 移动端检测 Composable
 *
 * 从旧前端 fe/src/composables/useIsMobile.ts 迁移。
 * 响应式检测视口宽度是否在移动端范围内（< 768px）。
 *
 * 创建日期: 20260609
 */
import { ref, onMounted, onUnmounted } from 'vue'

export function useIsMobile(breakpoint = 768) {
  const isMobile = ref(false)

  function check() {
    isMobile.value = window.innerWidth < breakpoint
  }

  onMounted(() => {
    check()
    window.addEventListener('resize', check)
  })

  onUnmounted(() => {
    window.removeEventListener('resize', check)
  })

  return { isMobile }
}