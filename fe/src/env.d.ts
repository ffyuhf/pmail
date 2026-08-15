/// <reference types="vite/client" />

/** 声明 .vue 模块类型，使 TS 能识别 Vue 单文件组件 */
declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}
