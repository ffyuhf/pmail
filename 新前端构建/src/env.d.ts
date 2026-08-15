/**
 * 环境类型声明
 *
 * 声明 Vue SFC 模块类型和 Vite 环境变量类型。
 *
 * 创建日期: 20260609
 */

/** Vue 单文件组件类型声明 */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

/** Vite 环境变量类型 */
interface ImportMetaEnv {
  readonly VITE_APP_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}