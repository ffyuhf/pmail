<template>
  <!-- 密码修改区块 -->
  <SettingsCard :title="lang.modify_pwd" :description="lang.modify_pwd_desc">
    <div class="settings-body">
      <el-form :model="ruleForm" :rules="rules" status-icon label-position="top" class="settings-form">
        <el-form-item :label="lang.modify_pwd" prop="new_pwd">
          <el-input type="password" v-model="ruleForm.new_pwd" :placeholder="lang.enter_new_pwd" show-password class="settings__input"/>
        </el-form-item>

        <el-form-item :label="lang.enter_again" prop="new_pwd2">
          <el-input type="password" v-model="ruleForm.new_pwd2" :placeholder="lang.confirm_new_pwd" show-password class="settings__input"/>
        </el-form-item>

        <div class="settings__form-actions">
          <el-button type="primary" @click="submit" class="settings__btn submit-btn">
            {{ lang.submit }}
          </el-button>
        </div>
      </el-form>
    </div>
  </SettingsCard>

  <!-- API Token 管理区块 -->
  <SettingsCard :title="lang.api_token_management" :description="lang.api_token_management_desc">
    <div class="settings-body">
      <!-- 生成区 -->
      <el-form label-position="top" class="settings-form">
        <el-form-item :label="lang.token_expire_label">
          <el-select v-model="tokenExpire" class="settings__input">
            <el-option :label="lang.token_never" :value="0"/>
            <el-option :label="lang.token_7d" :value="604800"/>
            <el-option :label="lang.token_30d" :value="2592000"/>
          </el-select>
        </el-form-item>

        <div class="settings__form-actions">
          <el-button type="primary" :loading="generating" @click="generateToken" class="settings__btn submit-btn">
            {{ lang.token_generate }}
          </el-button>
        </div>
      </el-form>

      <!-- 生成结果展示（一次性，仅当前会话可见） -->
      <div v-if="generatedToken" class="token-result">
        <el-input :model-value="generatedToken" readonly>
          <template #append>
            <el-button @click="copyToken">{{ lang.token_copy }}</el-button>
          </template>
        </el-input>
      </div>

      <!-- 撤销区 -->
      <el-form label-position="top" class="settings-form token-revoke-form">
        <el-form-item>
          <el-input v-model="revokeTokenValue" :placeholder="lang.token_revoke_ph" class="settings__input"/>
        </el-form-item>

        <div class="settings__form-actions">
          <el-button type="danger" plain :loading="revoking" @click="revokeToken" class="settings__btn">
            {{ lang.token_revoke }}
          </el-button>
        </div>
      </el-form>
    </div>
  </SettingsCard>

  <!-- 注销区块 -->
  <SettingsCard :title="lang.logout" :description="lang.logout_desc" :danger-title="true">
    <div class="settings-body">
      <el-button type="danger" plain @click="logout" class="settings__btn">
        {{ lang.logout }}
      </el-button>
    </div>
  </SettingsCard>
</template>

<script setup lang="ts">
import {reactive, ref} from 'vue'
import {ElMessageBox, ElNotification} from 'element-plus'
import lang from '../i18n/i18n';
import {userService} from "@/services/userService";
import {tokenService} from "@/services/tokenService";
import SettingsCard from "@/components/settings/SettingsCard.vue";

const ruleForm = reactive({
  new_pwd: "",
  new_pwd2: ""
})

const rules = reactive({
  new_pwd: [{required: true, message: lang.err_required_pwd, trigger: 'blur'},],
  new_pwd2: [{required: true, message: lang.err_required_pwd, trigger: 'blur'},],
})

/** 注销：通过 userService 调用注销接口 */
const logout = function () {
  userService.logout().then(() => {
    location.reload();
  })
}

/** 修改密码：通过 userService 调用密码修改接口 */
const submit = function () {
  if (ruleForm.new_pwd === "") return;
  if (ruleForm.new_pwd !== ruleForm.new_pwd2) {
    ElNotification({
      title: lang.error,
      message: lang.err_pwd_diff,
      type: 'error',
    })
    return
  }
  userService.modifyPassword(ruleForm.new_pwd).then((res: any) => {
    ElNotification({
      title: res.errorNo === 0 ? lang.succ : lang.fail,
      message: res.data,
      type: res.errorNo === 0 ? 'success' : 'error',
    })
  })
}

/** Token 有效期（秒）：0=永不过期、7天、30天 */
const tokenExpire = ref(0)
const generating = ref(false)
const revoking = ref(false)
const generatedToken = ref("")
const revokeTokenValue = ref("")

/** 生成 API Token */
const generateToken = function () {
  generating.value = true
  tokenService.generateToken(tokenExpire.value).then((res: any) => {
    if (res.errorNo === 0) {
      generatedToken.value = res.data.token || ""
      ElNotification({
        title: lang.succ,
        message: lang.token_generated,
        type: 'success',
      })
    } else {
      ElNotification({
        title: lang.fail,
        message: res.errorMsg || res.data,
        type: 'error',
      })
    }
  }).finally(() => {
    generating.value = false
  })
}

/** 复制生成的 Token 到剪贴板 */
const copyToken = function () {
  if (!generatedToken.value) return;
  navigator.clipboard?.writeText(generatedToken.value).then(() => {
    ElNotification({
      title: lang.succ,
      message: lang.token_copied,
      type: 'success',
    })
  })
}

/** 撤销指定 Token（二次确认后执行） */
const revokeToken = function () {
  const token = revokeTokenValue.value.trim()
  if (!token) return;

  ElMessageBox.confirm(lang.token_revoke_confirm, lang.token_revoke, {
    confirmButtonText: lang.submit,
    cancelButtonText: lang.cancel,
    type: 'warning',
  }).then(() => {
    revoking.value = true
    tokenService.revokeToken(token).then((res: any) => {
      if (res.errorNo === 0) {
        revokeTokenValue.value = ""
        if (generatedToken.value === token) {
          generatedToken.value = ""
        }
        ElNotification({
          title: lang.succ,
          message: lang.token_revoked,
          type: 'success',
        })
      } else {
        ElNotification({
          title: lang.fail,
          message: res.errorMsg || res.data,
          type: 'error',
        })
      }
    }).finally(() => {
      revoking.value = false
    })
  }).catch(() => {})
}
</script>

<style scoped>
/* 引入设置组件公共样式（settings__input、settings__form-actions 等） */
@import '@/assets/settings-common.css';

.settings-body {
  width: 100%;
}

/* Token 生成结果与撤销区样式 */
.token-result {
  margin-bottom: 20px;
}

.token-revoke-form {
  margin-top: 24px;
}

.submit-btn {
  font-weight: 500;
  padding: 10px 24px;
}
</style>
