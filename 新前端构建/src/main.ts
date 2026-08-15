/**
 * 应用入口文件
 *
 * 初始化 Vue 3 应用，注册 Pinia 状态管理和 Vue Router 路由。
 * 引入全局 CSS 变量和基础样式。
 *
 * 创建日期: 20260609
 */
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'

/* 全局样式：变量定义必须在 base.css 之前加载 */
import './assets/variables.css'
import './assets/base.css'

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')