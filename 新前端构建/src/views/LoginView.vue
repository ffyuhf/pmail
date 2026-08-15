<!--
  LoginView 登录页视图

  极简 Apple 风格登录页，包含：
  - 品牌 Logo 和标语
  - 账号密码输入框
  - 登录按钮

  创建日期: 20260609
-->
<template>
  <div class="login-page">
    <div class="login-card">
      <!-- 品牌 -->
      <div class="brand">
        <div class="brand-logo">P</div>
        <h1 class="brand-name">PMail</h1>
        <p class="brand-desc">{{ lang.login_brand_desc }}</p>
      </div>

      <!-- 表单 -->
      <form class="login-form" @submit.prevent="onLogin">
        <div class="form-field">
          <label class="field-label">{{ lang.account }}</label>
          <input
            v-model="account"
            type="text"
            class="field-input"
            autocomplete="username"
          />
        </div>
        <div class="form-field">
          <label class="field-label">{{ lang.password }}</label>
          <input
            v-model="password"
            type="password"
            class="field-input"
            autocomplete="current-password"
          />
        </div>

        <!-- 错误提示 -->
        <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>

        <button type="submit" class="login-btn" :disabled="submitting">
          {{ submitting ? lang.wait_desc : lang.login }}
        </button>
      </form>

      <p class="login-subtitle">{{ lang.login_subtitle }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import lang from '@/i18n'
import { login } from '@/services/userService'
import { useGlobalStore } from '@/stores/global'

const router = useRouter()
const globalStore = useGlobalStore()

const account = ref('')
const password = ref('')
const errorMsg = ref('')
const submitting = ref(false)

async function onLogin() {
  errorMsg.value = ''
  if (!account.value || !password.value) {
    errorMsg.value = lang.login_fill_required
    return
  }
  submitting.value = true
  try {
    const res: any = await login(account.value, password.value)
    /* axios 拦截器已解包，直接读 errorNo */
    if (res.errorNo === 0) {
      await globalStore.fetchUserInfo()
      router.push('/')
    } else {
      errorMsg.value = res.errorMsg || lang.fail
    }
  } catch {
    errorMsg.value = lang.fail
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.login-page {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-secondary);
}

.login-card {
  width: 360px;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  padding: 40px 32px;
}

.brand {
  text-align: center;
  margin-bottom: 32px;
}

.brand-logo {
  width: 56px;
  height: 56px;
  border-radius: var(--radius);
  background: var(--accent-color);
  color: #fff;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 24px;
  margin-bottom: 12px;
}

.brand-name {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 4px;
}

.brand-desc {
  font-size: 13px;
  color: var(--text-secondary);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 16px;
}

.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.field-input {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  font-size: 14px;
  outline: none;
  transition: border-color var(--transition);
}

.field-input:focus {
  border-color: var(--accent-color);
}

.error-msg {
  font-size: 13px;
  color: var(--danger-color);
}

.login-btn {
  padding: 10px;
  border: none;
  border-radius: var(--radius);
  background: var(--accent-color);
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: background var(--transition);
}

.login-btn:hover:not(:disabled) {
  background: var(--accent-hover);
}

.login-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.login-subtitle {
  text-align: center;
  font-size: 12px;
  color: var(--text-placeholder);
}
</style>