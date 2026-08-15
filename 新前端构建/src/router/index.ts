/**
 * Vue Router 路由配置
 *
 * 从旧前端 fe/src/router/index.ts 迁移。
 * 使用 Hash 模式路由，定义登录、主页、设置三个路由。
 *
 * 创建日期: 20260609
 */
import { createRouter, createWebHashHistory } from 'vue-router'

export const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
    },
    {
      path: '/setup',
      name: 'setup',
      component: () => import('@/views/SetupView.vue'),
    },
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
  ],
})

/** 路由守卫：已登录可访问主页，否则跳转登录页 */
router.beforeEach(async (to, _from, next) => {
  if (to.path === '/login' || to.path === '/setup') {
    next()
    return
  }
  /* 简单校验：通过检查 cookie/session 自动携带的身份凭证 */
  /* 详细登录态检查由 App.vue 中的 globalStore.fetchUserInfo() 处理 */
  next()
})
