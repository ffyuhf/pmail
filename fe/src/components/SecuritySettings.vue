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
import {reactive} from 'vue'
import {ElNotification} from 'element-plus'
import lang from '../i18n/i18n';
import {userService} from "@/services/userService";
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
      title: 'Error',
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
</script>

<style scoped>
/* 引入设置组件公共样式（settings__input、settings__form-actions 等） */
@import '@/assets/settings-common.css';

.settings-body {
  width: 100%;
}

.submit-btn {
  font-weight: 500;
  padding: 10px 24px;
}
</style>
