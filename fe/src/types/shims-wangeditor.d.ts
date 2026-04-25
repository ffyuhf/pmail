/* @wangeditor/editor-for-vue 类型声明 */
declare module '@wangeditor/editor-for-vue' {
  import { DefineComponent } from 'vue'
  export const Editor: DefineComponent<any, any, any>
  export const Toolbar: DefineComponent<any, any, any>
}

declare module '@wangeditor/editor' {
  export function i18nChangeLanguage(lang: string): void
}
