<!--
  UserManagement 用户管理设置组件

  对齐旧前端 fe/src/components/UserManagement.vue 完整功能：
  - 用户列表表格（ID/Account/Name/Disabled 状态标签）
  - 分页
  - 编辑/创建用户对话框（account/name/password/disabled swith）
  - createUser/editUser API 调用

  修改日期: 20260609
  修改原因: 旧实现缺少编辑对话框、禁用开关和分页。
-->
<template>
  <div class="user-management">
    <div class="section-header">
      <h3>{{ lang.user_management }}</h3>
      <p class="section-desc">{{ lang.user_management_desc }}</p>
    </div>

    <!-- 用户列表表格 -->
    <div class="user-table-wrap">
      <table class="user-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>{{ lang.account }}</th>
            <th>{{ lang.user_name }}</th>
            <th>{{ lang.disabled }}</th>
            <th class="cell-actions">
              <button class="btn-primary btn-sm" @click="openNewDialog">+ {{ lang.newUser }}</button>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="user in userList" :key="user.ID">
            <td>{{ user.ID }}</td>
            <td class="cell-text">{{ user.Account }}</td>
            <td>{{ user.Name }}</td>
            <td>
              <span class="status-tag" :class="user.Disabled === 1 ? 'disabled' : 'enabled'">
                {{ user.Disabled === 1 ? lang.disabled : lang.enabled }}
              </span>
            </td>
            <td class="cell-actions">
              <button class="icon-btn" @click="handleEdit(user)">{{ lang.editUser }}</button>
            </td>
          </tr>
          <tr v-if="userList.length === 0">
            <td colspan="5" class="empty-cell">—</td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 分页 -->
    <div class="pagination" v-if="totalPage > 1">
      <button
        v-for="p in totalPage" :key="p"
        class="page-btn"
        :class="{ active: currentPage === p }"
        @click="goPage(p)"
      >{{ p }}</button>
    </div>

    <!-- 编辑/创建用户对话框 -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog-panel">
        <div class="dialog-header">
          <h3>{{ editModel === 'edit' ? lang.editUser : lang.newUser }}</h3>
          <button class="close-btn" @click="closeDialog">×</button>
        </div>
        <div class="dialog-body">
          <div class="form-field">
            <label>{{ lang.account }}</label>
            <input
              v-model="editUser.account"
              class="field-input"
              :disabled="editModel === 'edit'"
            />
          </div>
          <div class="form-field">
            <label>{{ lang.user_name }}</label>
            <input v-model="editUser.name" class="field-input" />
          </div>
          <div class="form-field">
            <label>{{ lang.password }}</label>
            <input
              v-model="editUser.password"
              type="password"
              class="field-input"
              :placeholder="editModel === 'edit' ? lang.resetPwd : ''"
            />
          </div>
          <div class="form-field">
            <div class="switch-row">
              <span>{{ lang.disabled }}</span>
              <label class="switch">
                <input type="checkbox" v-model="editUser.disabled" />
                <span class="slider"></span>
              </label>
              <span>{{ editUser.disabled ? lang.disabled : lang.enabled }}</span>
            </div>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn-secondary" @click="closeDialog">Cancel</button>
          <button class="btn-primary" @click="submit">{{ lang.submit }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import lang from '@/i18n'
import { getUserList, createUser, updateUser } from '@/services/userService'

/* eslint-disable @typescript-eslint/no-explicit-any */
const userList = ref<any[]>([])
const currentPage = ref(1)
const pageSize = 10
const totalPage = ref(1)
const showDialog = ref(false)
const editModel = ref<'create' | 'edit'>('create')

const editUser = ref({
  account: '',
  name: '',
  password: '',
  disabled: false,
})

/** 加载用户列表（分页） */
async function fetchUsers() {
  const res: any = await getUserList(currentPage.value, pageSize)
  if (res.errorNo === 0) {
    userList.value = res.data?.list || []
    totalPage.value = res.data?.total_page || 1
  }
}

function goPage(p: number) {
  currentPage.value = p
  fetchUsers()
}

/** 打开新建用户对话框 */
function openNewDialog() {
  editUser.value = { account: '', name: '', password: '', disabled: false }
  editModel.value = 'create'
  showDialog.value = true
}

/** 打开编辑用户对话框 */
function handleEdit(row: any) {
  editUser.value = {
    account: row.Account || '',
    name: row.Name || '',
    password: '',
    disabled: row.Disabled === 1,
  }
  editModel.value = 'edit'
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
}

/** 提交创建/编辑用户 */
async function submit() {
  if (!editUser.value.account) return
  const isEdit = editModel.value === 'edit'

  const payload: any = {
    account: editUser.value.account,
    username: editUser.value.name,
    disabled: editUser.value.disabled ? 1 : 0,
  }

  if (isEdit) {
    if (editUser.value.password) {
      payload.password = editUser.value.password
    }
  } else {
    if (!editUser.value.password) return
    payload.password = editUser.value.password
  }

  try {
    const res: any = isEdit
      ? await updateUser(payload)
      : await createUser(payload)

    if (res.errorNo === 0) {
      closeDialog()
      fetchUsers()
    } else {
      alert(res.errorMsg || lang.fail)
    }
  } catch {
    alert(lang.fail)
  }
}

onMounted(fetchUsers)
</script>

<style scoped>
.user-management { width: 100%; }
.section-header { margin-bottom: 16px; }
.section-header h3 { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.section-desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; }

/* 用户列表表格 */
.user-table-wrap {
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}
.user-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.user-table th {
  text-align: left; padding: 10px 12px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  font-weight: 600; color: var(--text-secondary);
}
.user-table td { padding: 10px 12px; border-bottom: 1px solid var(--bg-secondary); }
.user-table tr:last-child td { border-bottom: none; }
.user-table tr:hover td { background: var(--bg-hover); }
.cell-text { font-weight: 500; }
.cell-actions { text-align: right; white-space: nowrap; }
.empty-cell { text-align: center; color: var(--text-placeholder); padding: 24px !important; }

/* 状态标签 */
.status-tag {
  display: inline-block; padding: 2px 8px;
  border-radius: 4px; font-size: 12px;
}
.status-tag.enabled { background: #e6f7e6; color: #2e7d32; }
.status-tag.disabled { background: #f5f5f5; color: #999; }

/* 分页 */
.pagination { display: flex; gap: 4px; justify-content: center; margin-top: 12px; }
.page-btn {
  width: 28px; height: 28px; border: 1px solid var(--border-color);
  border-radius: 4px; background: transparent; cursor: pointer; font-size: 12px;
}
.page-btn:hover { background: var(--bg-hover); }
.page-btn.active { background: var(--accent-color); color: #fff; border-color: var(--accent-color); }

/* 对话框覆盖层 */
.dialog-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 1000;
  display: flex; align-items: center; justify-content: center;
}
.dialog-panel {
  width: 450px;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  display: flex; flex-direction: column;
}
.dialog-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 16px 20px; border-bottom: 1px solid var(--border-color);
}
.dialog-header h3 { font-size: 16px; font-weight: 600; }
.dialog-body { padding: 20px; display: flex; flex-direction: column; gap: 16px; }
.dialog-footer {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 12px 20px; border-top: 1px solid var(--border-color);
}

/* 表单 */
.form-field { display: flex; flex-direction: column; gap: 4px; }
.form-field label { font-size: 12px; font-weight: 500; color: var(--text-secondary); }
.field-input {
  padding: 8px 12px; border: 1px solid var(--border-color);
  border-radius: var(--radius); font-size: 13px; outline: none;
}
.field-input:focus { border-color: var(--accent-color); }
.field-input:disabled { background: var(--bg-secondary); }

/* 开关 */
.switch-row { display: flex; align-items: center; gap: 12px; font-size: 13px; }
.switch {
  position: relative; display: inline-block; width: 44px; height: 24px;
}
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute; cursor: pointer; inset: 0;
  background: #ccc; border-radius: 24px; transition: 0.2s;
}
.slider::before {
  content: ''; position: absolute; width: 18px; height: 18px;
  left: 3px; bottom: 3px; background: white; border-radius: 50%; transition: 0.2s;
}
.switch input:checked + .slider { background: #ef4444; }
.switch input:checked + .slider::before { transform: translateX(20px); }

/* 按钮 */
.btn-primary {
  padding: 8px 20px; border: none; border-radius: var(--radius);
  background: var(--accent-color); color: #fff; font-size: 13px; cursor: pointer;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-primary.btn-sm { padding: 4px 12px; font-size: 12px; }
.btn-secondary {
  padding: 8px 20px; border: 1px solid var(--border-color); border-radius: var(--radius);
  background: transparent; color: var(--text-primary); font-size: 13px; cursor: pointer;
}
.btn-secondary:hover { background: var(--bg-hover); }
.icon-btn {
  border: none; background: transparent; cursor: pointer;
  font-size: 13px; color: var(--accent-color); padding: 4px 8px;
  border-radius: 4px;
}
.icon-btn:hover { background: var(--bg-hover); }
.close-btn {
  width: 32px; height: 32px; border: none; background: transparent;
  font-size: 20px; cursor: pointer; border-radius: var(--radius-sm);
  display: flex; align-items: center; justify-content: center;
}
.close-btn:hover { background: var(--bg-hover); }
</style>
