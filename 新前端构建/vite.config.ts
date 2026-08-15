/**
 * Vite 构建配置
 *
 * 新前端项目构建配置，包含 Vue 插件、路径别名和 API 代理。
 * API 代理将 /api 请求转发到后端服务，解决开发环境跨域问题。
 *
 * 创建日期: 20260609
 */
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})