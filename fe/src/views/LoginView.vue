<template>
  <!-- 登录页面（Docusaurus BEM 风格） -->
  <div class="auth-page">
    <div class="auth-page__container">
      <!-- 左侧品牌区域 -->
      <div class="auth-page__brand">
        <div class="auth-page__brand-info">
          <h1>PMail</h1>
        </div>
      </div>
      <!-- 右侧表单区域 -->
      <div class="auth-page__form-area">
        <div class="auth-page__form-box">
          <h2>{{ lang.login }}</h2>
          <p class="auth-page__subtitle">{{ lang.login_subtitle }}</p>

          <el-form :model="form" class="auth-page__form" @keyup.enter="onSubmit" label-position="top">
            <el-form-item :label="lang.account">
              <el-input
                v-model="form.account"
                :placeholder="lang.account"
                size="large"
                class="auth-page__input"
              />
            </el-form-item>
            <el-form-item :label="lang.password">
              <el-input
                v-model="form.password"
                :placeholder="lang.password"
                type="password"
                size="large"
                class="auth-page__input"
                show-password
              />
            </el-form-item>
            <div class="auth-page__submit">
              <el-button type="primary" @click="onSubmit" size="large" class="auth-page__submit-btn" :loading="loading">
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
import {userService} from "@/services/userService";
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
  /** 通过 userService 登录 */
  userService.login(form.account, form.password).then((res: any) => {
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

<!-- 样式: Docusaurus BEM 风格 | 重构日期: 20260429 -->
<style scoped>
/* 登录页面：全屏居中布局 */
.auth-page {
  width: 100vw;
  height: 100vh;
  background: var(--ifm-background-color);
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--ifm-spacing-lg);
}

/* 登录卡片：左右分栏 */
.auth-page__container {
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

/* 左侧品牌区域：主色渐变背景 */
.auth-page__brand {
  flex: 1;
  background: linear-gradient(145deg, var(--ifm-color-primary-dark) 0%, var(--ifm-color-primary) 55%, var(--ifm-color-primary-light) 100%);
  padding: 64px;
  color: #fff;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.auth-page__brand-info h1 {
  font-size: 48px;
  font-weight: 700;
  margin-bottom: var(--ifm-spacing-sm);
  letter-spacing: -0.02em;
}

/* 右侧表单区域 */
.auth-page__form-area {
  flex: 1;
  padding: 64px;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.auth-page__form-box {
  width: 100%;
  max-width: 380px;
  margin: 0 auto;
}

.auth-page__form-box h2 {
  font-size: 28px;
  font-weight: 700;
  color: var(--ifm-color-content);
  margin-bottom: var(--ifm-spacing-sm);
}

.auth-page__subtitle {
  color: var(--ifm-color-content-secondary);
  margin-bottom: var(--ifm-spacing-xl);
  font-size: 14px;
}

/* 表单标签样式 */
.auth-page__form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--ifm-color-content);
  padding-bottom: 4px;
}

/* 输入框：统一圆角 + 边框风格 */
.auth-page__input :deep(.el-input__wrapper) {
  background-color: var(--ifm-background-color);
  border-radius: var(--ifm-global-radius);
  box-shadow: 0 0 0 1px var(--ifm-border-color) inset !important;
}

.auth-page__input :deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--ifm-color-content-secondary) inset !important;
}

.auth-page__input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--ifm-color-primary) inset !important;
}

/* 提交按钮区 */
.auth-page__submit {
  margin-top: var(--ifm-spacing-xl);
}

.auth-page__submit-btn {
  width: 100%;
  font-weight: 600;
  border-radius: var(--ifm-global-radius);
  height: 48px;
  font-size: 15px;
}

/* 移动端：左右分栏改为上下堆叠 */
@media (max-width: 768px) {
  .auth-page__container {
    flex-direction: column;
    height: auto;
    min-height: 0;
  }
  .auth-page__brand {
    padding: var(--ifm-spacing-xl) var(--ifm-spacing-lg);
    align-items: center;
    text-align: center;
  }
  .auth-page__brand-info h1 {
    font-size: 36px;
  }
  .auth-page__form-area {
    padding: var(--ifm-spacing-xl) var(--ifm-spacing-lg);
  }
}
</style>
