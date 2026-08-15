<!--
  SecuritySettings 安全设置组件

  提供修改密码和注销功能。

  创建日期: 20260609
-->
<template>
  <div class="security-settings">
    <!-- 修改密码 -->
    <section class="setting-section">
      <h3>{{ lang.modify_pwd }}</h3>
      <p class="section-desc">{{ lang.modify_pwd_desc }}</p>
      <div class="form-field">
        <label class="field-label">{{ lang.enter_new_pwd }}</label>
        <input v-model="newPwd" type="password" class="field-input" />
      </div>
      <div class="form-field">
        <label class="field-label">{{ lang.confirm_new_pwd }}</label>
        <input v-model="confirmPwd" type="password" class="field-input" />
      </div>
      <div v-if="msg" class="msg" :class="{ error: isError }">{{ msg }}</div>
      <button class="btn-primary" @click="onChangePwd">{{ lang.submit }}</button>
    </section>

    <!-- 注销 -->
    <section class="setting-section">
      <h3>{{ lang.logout }}</h3>
      <p class="section-desc">{{ lang.logout_desc }}</p>
      <button class="btn-danger" @click="onLogout">{{ lang.logout }}</button>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import lang from '@/i18n'
import { changePassword, logout } from '@/services/userService'
import { useGlobalStore } from '@/stores/global'

const router = useRouter()
const globalStore = useGlobalStore()

const newPwd = ref('')
const confirmPwd = ref('')
const msg = ref('')
const isError = ref(false)

async function onChangePwd() {
  msg.value = ''
  if (!newPwd.value) { msg.value = lang.err_required_pwd; isError.value = true; return }
  if (newPwd.value !== confirmPwd.value) { msg.value = lang.err_pwd_diff; isError.value = true; return }
  try {
    const res: any = await changePassword(newPwd.value)
    /* axios 拦截器已解包 response.data，直接读 errorNo */
    if (res.errorNo === 0) { msg.value = lang.succ; isError.value = false }
    else { msg.value = res.errorMsg || lang.fail; isError.value = true }
  } catch { msg.value = lang.fail; isError.value = true }
}

async function onLogout() {
  try { await logout() } finally {
    globalStore.clearUser()
    router.push('/login')
  }
}
</script>

<style scoped>
.setting-section { margin-bottom: 32px; }
.setting-section h3 { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.section-desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 16px; }
.form-field { display: flex; flex-direction: column; gap: 4px; margin-bottom: 12px; }
.field-label { font-size: 13px; color: var(--text-secondary); }
.field-input {
  padding: 8px 12px; border: 1px solid var(--border-color); border-radius: var(--radius);
  font-size: 13px; outline: none; max-width: 320px;
}
.field-input:focus { border-color: var(--accent-color); }
.msg { font-size: 13px; margin-bottom: 8px; }
.msg:not(.error) { color: var(--success-color); }
.msg.error { color: var(--danger-color); }
.btn-primary {
  padding: 8px 20px; border: none; border-radius: var(--radius);
  background: var(--accent-color); color: #fff; font-size: 13px; cursor: pointer;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-danger {
  padding: 8px 20px; border: 1px solid var(--danger-color); border-radius: var(--radius);
  background: transparent; color: var(--danger-color); font-size: 13px; cursor: pointer;
}
.btn-danger:hover { background: var(--danger-bg); }
</style>