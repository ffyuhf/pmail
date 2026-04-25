<template>
  <div class="login-wrapper">
    <div class="login-container">
      <div class="login-left">
        <div class="brand-info">
          <h1>PMail</h1>
        </div>
      </div>
      <div class="login-right">
        <div class="login-form-box">
          <h2>{{ lang.login }}</h2>
          <p class="subtitle">{{ lang.login_subtitle }}</p>
          
          <el-form :model="form" class="auth-form" @keyup.enter="onSubmit" label-position="top">
            <el-form-item :label="lang.account">
              <el-input 
                v-model="form.account" 
                :placeholder="lang.account"
                size="large"
                class="premium-input"
              />
            </el-form-item>
            <el-form-item :label="lang.password">
              <el-input 
                v-model="form.password" 
                :placeholder="lang.password"
                type="password" 
                size="large"
                class="premium-input"
                show-password
              />
            </el-form-item>
            <div class="submit-action">
              <el-button type="primary" @click="onSubmit" size="large" class="login-btn" :loading="loading">
                {{ lang.login }}
              </el-button>
            </div>
          </el-form>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {reactive, ref} from 'vue'
import {ElMessage} from 'element-plus'
import {router} from "@/router";
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";

const globalStatus = useGlobalStatusStore();
const loading = ref(false);

const form = reactive({
  account: '',
  password: '',
})

const onSubmit = () => {
  if (!form.account || !form.password) {
    ElMessage.warning(lang.login_fill_required);
    return;
  }
  loading.value = true;
  http.post("/api/login", form).then((res: any) => {
    loading.value = false;
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      Object.assign(globalStatus.userInfos , res.data) 
      router.replace({
        path: '/',
        query: {
          redirect: router.currentRoute.value.fullPath
        }
      })
    }
  }).catch(() => {
    loading.value = false;
  })
}
</script>

<!-- 样式改造: Docusaurus 风格 | 日期: 20250425 -->
<style scoped>
.login-wrapper {
  width: 100vw;
  height: 100vh;
  background: var(--ifm-background-color);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--ifm-spacing-lg);
}

.login-container {
  width: 100%;
  max-width: 1000px;
  min-height: 560px;
  background: var(--ifm-background-surface-color);
  border-radius: var(--ifm-card-border-radius);
  box-shadow: var(--ifm-global-shadow-md);
  border: 1px solid var(--ifm-border-color);
  display: flex;
  overflow: hidden;
}

/* 左侧品牌区域 */
.login-left {
  flex: 1;
  background: linear-gradient(145deg, var(--ifm-color-primary-dark) 0%, var(--ifm-color-primary) 55%, var(--ifm-color-primary-light) 100%);
  padding: 64px;
  color: #fff;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.brand-info h1 {
  font-size: 48px;
  font-weight: 700;
  margin-bottom: var(--ifm-spacing-sm);
  letter-spacing: -0.02em;
}

/* 右侧表单区域 */
.login-right {
  flex: 1;
  padding: 64px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.login-form-box {
  width: 100%;
  max-width: 380px;
  margin: 0 auto;
}

.login-form-box h2 {
  font-size: 28px;
  font-weight: 700;
  color: var(--ifm-color-content);
  margin-bottom: var(--ifm-spacing-sm);
}

.subtitle {
  color: var(--ifm-color-content-secondary);
  margin-bottom: var(--ifm-spacing-xl);
  font-size: 14px;
}

.auth-form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--ifm-color-content);
  padding-bottom: 4px;
}

.premium-input :deep(.el-input__wrapper) {
  background-color: var(--ifm-background-color);
  border-radius: var(--ifm-global-radius);
  box-shadow: 0 0 0 1px var(--ifm-border-color) inset !important;
}

.premium-input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--ifm-color-content-secondary) inset !important;
}

.premium-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--ifm-color-primary) inset !important;
}

.submit-action {
  margin-top: var(--ifm-spacing-xl);
}

.login-btn {
  width: 100%;
  font-weight: 600;
  border-radius: var(--ifm-global-radius);
  height: 48px;
  font-size: 15px;
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
    height: auto;
    min-height: 0;
  }
  .login-left {
    padding: var(--ifm-spacing-xl) var(--ifm-spacing-lg);
    align-items: center;
    text-align: center;
  }
  .brand-info h1 {
    font-size: 36px;
  }
  .login-right {
    padding: var(--ifm-spacing-xl) var(--ifm-spacing-lg);
  }
}
</style>
