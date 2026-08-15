<template>
  <SettingsCard :title="lang.account_management" :description="lang.account_management_desc">

    <div class="settings__table-container">
      <el-table :data="userList" class="settings__table" style="width: 100%">
        <el-table-column label="ID" prop="ID" width="80"/>
        <el-table-column :label="lang.account" prop="Account" min-width="150" show-overflow-tooltip/>
        <el-table-column :label="lang.user_name" prop="Name" min-width="120" show-overflow-tooltip/>
        <el-table-column :label="lang.disabled" prop="Disabled" width="120">
          <template #default="scope">
            <el-tag :type="scope.row.Disabled === 1 ? 'info' : 'success'" size="small" effect="plain" class="settings__tag">
              {{ scope.row.Disabled === 1 ? lang.disabled : lang.enabled }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="right" width="100">
          <template #header>
            <el-button type="primary" size="small" @click="createUser" class="settings__btn" plain>
              <el-icon><Plus/></el-icon> {{ lang.new_btn }}
            </el-button>
          </template>
          <template #default="scope">
            <el-button size="small" type="primary" text bg @click="handleEdit(scope.$index, scope.row)" class="settings__btn">
              {{ lang.edit }}
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="pagination-wrapper">
      <el-pagination 
        v-model:current-page="currentPage" 
        small 
        background 
        layout="prev, pager, next"
        :page-count="totalPage" 
        @current-change="refreshList"
      />
    </div>

    <el-dialog v-model="userInfoDialog" :title="title" width="450px" class="settings__dialog">
      <div class="settings__dialog-content">
        <el-form label-position="top">
          <el-form-item :label="lang.account">
            <el-input :disabled="editModel === 'edit'" v-model="editUserInfo.account"/>
          </el-form-item>

          <el-form-item :label="lang.user_name">
            <el-input v-model="editUserInfo.name"/>
          </el-form-item>

          <el-form-item :label="lang.password">
            <el-input :placeholder="lang.resetPwd" v-model="editUserInfo.password" type="password" show-password/>
          </el-form-item>

          <el-form-item>
            <div class="status-switch">
              <span class="switch-label">{{ lang.status }}</span>
              <el-switch 
                v-model="editUserInfo.disabled" 
                class="ml-2" 
                :active-text="lang.disabled"
                :inactive-text="lang.enabled"
                active-color="#ef4444"
                inactive-color="#10b981"
              />
            </div>
          </el-form-item>
        </el-form>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="userInfoDialog = false">{{ lang.cancel }}</el-button>
          <el-button type="primary" @click="submit">{{ lang.confirm }}</el-button>
        </span>
      </template>
    </el-dialog>
  </SettingsCard>
</template>

<script setup lang="ts">
import {reactive, ref} from 'vue'
import lang from '../i18n/i18n';
import {userService} from "@/services/userService";
import {ElNotification} from "element-plus";
import {Plus} from "@element-plus/icons-vue";
import SettingsCard from "@/components/settings/SettingsCard.vue";

const userList = reactive<any[]>([])
const currentPage = ref(1)
const totalPage = ref(1)
const userInfoDialog = ref(false)
const editModel = ref("edit")
const editUserInfo = reactive({
  "account": "",
  "name": "",
  "password": "",
  "disabled": false
})
const title = ref(lang.editUser)

/** 通过 userService 获取用户列表 */
const refreshList = function () {
  userService.getUserList(currentPage.value, 10).then((res: any) => {
    userList.length = 0
    totalPage.value = res.data.total_page
    if (res.data["list"]) {
      userList.push(...res.data["list"])
    }
  })
}

const handleEdit = function (idx: number, row: any) {
  editUserInfo.account = row.Account
  editUserInfo.name = row.Name
  editUserInfo.disabled = row.Disabled === 1
  editUserInfo.password = ""
  editModel.value = "edit"
  title.value = lang.editUser
  userInfoDialog.value = true
}

const createUser = function () {
  editUserInfo.account = ""
  editUserInfo.name = ""
  editUserInfo.disabled = false
  editUserInfo.password = ""
  editModel.value = "create"
  title.value = lang.newUser
  userInfoDialog.value = true
}

/**
 * 统一处理用户提交结果通知
 * 消除 edit/create 分支中重复的 ElNotification + refreshList + 关闭对话框逻辑
 * @param res - API 响应数据
 */
const handleSubmitResult = function (res: any) {
  ElNotification({
    title: res.errorNo === 0 ? lang.succ : lang.fail,
    message: res.errorNo === 0 ? "" : res.data,
    type: res.errorNo === 0 ? 'success' : 'error',
  })
  if (res.errorNo === 0) {
    refreshList()
    userInfoDialog.value = false
  }
}

/** 提交用户信息：通过 userService 新增或编辑用户 */
const submit = function () {
  const isEdit = editModel.value === 'edit'

  /** 构建请求数据：编辑模式下密码为空则不传 */
  const newData: { account: string; username: string; disabled: number; password?: string } = {
    account: editUserInfo.account,
    username: editUserInfo.name,
    disabled: editUserInfo.disabled ? 1 : 0,
  }
  if (isEdit) {
    if (editUserInfo.password !== "") {
      newData.password = editUserInfo.password
    }
    userService.editUser(newData).then(handleSubmitResult)
  } else {
    newData.password = editUserInfo.password
    userService.createUser(newData as any).then(handleSubmitResult)
  }
}

refreshList()
</script>

<style scoped>
/* 引入设置组件公共样式（table-container、dialog、pagination 等） */
@import '@/assets/settings-common.css';

/* 用户管理独有样式 */
.status-tag {
  border-radius: var(--ifm-global-radius);
}

.status-switch {
  display: flex;
  align-items: center;
  gap: 16px;
  width: 100%;
  padding: 8px 12px;
  background: var(--ifm-background-surface-color);
  border-radius: var(--ifm-global-radius);
  border: 1px solid var(--ifm-border-color);
}

.switch-label {
  font-size: 14px;
  font-weight: 500;
  color: var(--ifm-color-content);
}
</style>
