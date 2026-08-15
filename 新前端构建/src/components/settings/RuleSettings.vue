<!--
  RuleSettings 规则管理设置组件

  对齐旧前端 fe/src/components/RuleSettings.vue 完整功能：
  - el-table 展示规则列表（name/action/params/sort）
  - el-dialog 编辑/创建规则对话框
  - 条件编辑器（field + type + rule，可添加/删除多行）
  - 动作配置（mark_read=1/forward=2/delete=3/move=4）
  - move 动作需要选择目标文件夹
  - forward 动作需要输入转发邮箱

  修改日期: 20260609
  修改原因: 旧实现仅有简化版条件输入，缺少动作配置和编辑对话框。
-->
<template>
  <div class="rule-settings">
    <div class="section-header">
      <h3>{{ lang.auto_rules || lang.rule_setting }}</h3>
      <p class="section-desc">{{ lang.auto_rules_desc || lang.rule_setting_desc }}</p>
    </div>

    <!-- 规则列表表格 -->
    <div class="rule-table-wrap">
      <table class="rule-table">
        <thead>
          <tr>
            <th>{{ lang.rule_name }}</th>
            <th>{{ lang.rule_do }}</th>
            <th>{{ lang.rule_params }}</th>
            <th>{{ lang.rule_priority }}</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="rule in rules" :key="rule.id">
            <td class="cell-name">{{ rule.name }}</td>
            <td>
              <span class="action-tag">{{ getActionName(rule.action) }}</span>
            </td>
            <td class="cell-params">{{ rule.params || '—' }}</td>
            <td>{{ rule.sort }}</td>
            <td class="cell-actions">
              <button class="icon-btn" @click="editRule(rule)" title="Edit">✏</button>
              <button class="icon-btn danger" @click="onDelete(rule)" title="Delete">🗑</button>
            </td>
          </tr>
          <tr v-if="rules.length === 0">
            <td colspan="5" class="empty-cell">—</td>
          </tr>
        </tbody>
      </table>
    </div>

    <div class="section-actions">
      <button class="btn-primary" @click="openNewDialog">{{ lang.new_rule }}</button>
    </div>

    <!-- 编辑/创建规则对话框 -->
    <div v-if="showDialog" class="dialog-overlay" @click.self="closeDialog">
      <div class="dialog-panel">
        <div class="dialog-header">
          <h3>{{ form.id > 0 ? lang.editUser || 'Edit Rule' : lang.new_rule }}</h3>
          <button class="close-btn" @click="closeDialog">×</button>
        </div>
        <div class="dialog-body">
          <div class="form-row">
            <div class="form-field flex-grow">
              <label>{{ lang.rule_name }}</label>
              <input v-model="form.name" class="field-input" placeholder="Rule Name" />
            </div>
            <div class="form-field w-32">
              <label>{{ lang.rule_priority }}</label>
              <input v-model.number="form.sort" type="number" class="field-input" />
            </div>
          </div>

          <div class="section-divider"></div>
          <h4 class="section-title">{{ lang.rule_desc }} (Conditions)</h4>

          <div class="conditions-list">
            <div v-for="(cond, idx) in form.rules" :key="idx" class="condition-row">
              <select v-model="cond.field" class="field-select w-36">
                <option value="From">{{ lang.from }}</option>
                <option value="Subject">{{ lang.subject }}</option>
                <option value="To">{{ lang.to }}</option>
                <option value="Cc">{{ lang.cc }}</option>
                <option value="Content">{{ lang.content }}</option>
              </select>
              <select v-model="cond.type" class="field-select w-32">
                <option value="equal">{{ lang.equal }}</option>
                <option value="contains">{{ lang.contains }}</option>
                <option value="regex">{{ lang.regex }}</option>
              </select>
              <input v-model="cond.rule" class="field-input flex-grow" placeholder="Value" />
              <button class="icon-btn danger" @click="removeCondition(idx)">✕</button>
            </div>
          </div>
          <button class="btn-sm" @click="addCondition">+ {{ lang.rule_desc }}</button>

          <div class="section-divider"></div>
          <h4 class="section-title">{{ lang.rule_do }} (Action)</h4>

          <div class="action-row">
            <select v-model="form.action" class="field-select w-48" @change="onActionChange">
              <option :value="ACTION_MARK_READ">{{ lang.mark_read }}</option>
              <option :value="ACTION_FORWARD">{{ lang.forward }}</option>
              <option :value="ACTION_DELETE">{{ lang.delete }}</option>
              <option :value="ACTION_MOVE">{{ lang.move }}</option>
            </select>

            <!-- move 动作：选择目标文件夹 -->
            <select v-if="form.action === ACTION_MOVE" v-model="form.params" class="field-select flex-grow">
              <option v-for="g in groupList" :key="g.id" :value="String(g.id)">{{ g.name }}</option>
            </select>

            <!-- forward 动作：输入转发邮箱 -->
            <input v-if="form.action === ACTION_FORWARD" v-model="form.params" class="field-input flex-grow" placeholder="Forward Email Address" />
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn-secondary" @click="closeDialog">Cancel</button>
          <button class="btn-primary" @click="submitRule">{{ lang.submit }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import lang from '@/i18n'
import { getRules, createRule, updateRule, deleteRule } from '@/services/ruleService'
import { getGroupList } from '@/services/groupService'
import type { RuleItem, RuleCondition, GroupListItem } from '@/types/api'

/**
 * 动作枚举（对齐后端 rule.go）
 * 1=mark_read, 2=forward, 3=delete, 4=move
 */
const ACTION_MARK_READ = 1
const ACTION_FORWARD = 2
const ACTION_DELETE = 3
const ACTION_MOVE = 4

const ActionName: Record<number, string> = {
  [ACTION_MARK_READ]: lang.mark_read,
  [ACTION_FORWARD]: lang.forward,
  [ACTION_DELETE]: lang.delete,
  [ACTION_MOVE]: lang.move,
}

function getActionName(action: number): string {
  return ActionName[action] || `Action ${action}`
}

/* ===== 状态 ===== */
const rules = ref<any[]>([])
const showDialog = ref(false)
const groupList = ref<GroupListItem[]>([])

/** 表单数据（对齐后端 RuleItem 结构体） */
const form = ref<{
  id: number
  name: string
  sort: number
  rules: { field: string; type: string; rule: string }[]
  action: number
  params: string
}>({
  id: 0,
  name: '',
  sort: 0,
  rules: [{ field: 'From', type: 'contains', rule: '' }],
  action: ACTION_MARK_READ,
  params: '',
})

/* ===== 数据加载 ===== */
async function fetchRules() {
  const res: any = await getRules()
  if (res.errorNo === 0) rules.value = res.data || []
}

async function fetchGroups() {
  const res: any = await getGroupList()
  if (res.errorNo === 0) groupList.value = res.data || []
}

/* ===== 对话框操作 ===== */
function openNewDialog() {
  form.value = {
    id: 0,
    name: '',
    sort: 0,
    rules: [{ field: 'From', type: 'contains', rule: '' }],
    action: ACTION_MARK_READ,
    params: '',
  }
  showDialog.value = true
}

function editRule(rule: any) {
  form.value = {
    id: rule.id,
    name: rule.name || '',
    sort: rule.sort || 0,
    rules: (rule.rules && rule.rules.length > 0) ? rule.rules.map((c: RuleCondition) => ({ field: c.field, type: c.type, rule: c.rule })) : [{ field: 'From', type: 'contains', rule: '' }],
    action: rule.action || ACTION_MARK_READ,
    params: rule.params || '',
  }
  showDialog.value = true
}

function closeDialog() {
  showDialog.value = false
}

/* ===== 条件操作 ===== */
function addCondition() {
  form.value.rules.push({ field: 'From', type: 'contains', rule: '' })
}

function removeCondition(index: number) {
  form.value.rules.splice(index, 1)
}

/* ===== 动作切换 ===== */
function onActionChange() {
  form.value.params = ''
}

/* ===== 提交规则 ===== */
async function submitRule() {
  if (!form.value.name) return
  form.value.sort = Number(form.value.sort) || 0

  const payload = {
    name: form.value.name,
    sort: form.value.sort,
    rules: form.value.rules.filter(c => c.rule !== ''),
    action: form.value.action,
    params: form.value.params,
  }

  try {
    const res: any = form.value.id > 0
      ? await updateRule({ id: form.value.id, ...payload } as any)
      : await createRule(payload as any)

    if (res.errorNo === 0) {
      closeDialog()
      await fetchRules()
    } else {
      alert(res.errorMsg || lang.fail)
    }
  } catch {
    alert(lang.fail)
  }
}

/* ===== 删除规则 ===== */
async function onDelete(rule: any) {
  if (!confirm(lang.del_rule_confirm)) return
  try {
    const res: any = await deleteRule(rule.id)
    if (res.errorNo === 0) await fetchRules()
  } catch { /* ignore */ }
}

onMounted(() => {
  fetchRules()
  fetchGroups()
})
</script>

<style scoped>
.rule-settings { width: 100%; }
.section-header { margin-bottom: 16px; }
.section-header h3 { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.section-desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; }
.section-actions { margin-top: 16px; }
.section-divider {
  height: 1px;
  background: var(--border-color);
  margin: 20px 0;
}
.section-title { font-size: 14px; font-weight: 600; margin-bottom: 12px; }

/* 规则列表表格 */
.rule-table-wrap {
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}
.rule-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.rule-table th {
  text-align: left; padding: 10px 12px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
  font-weight: 600; color: var(--text-secondary);
}
.rule-table td { padding: 10px 12px; border-bottom: 1px solid var(--bg-secondary); }
.rule-table tr:last-child td { border-bottom: none; }
.rule-table tr:hover td { background: var(--bg-hover); }
.cell-name { font-weight: 500; }
.cell-params { color: var(--text-secondary); font-size: 12px; }
.cell-actions { text-align: right; white-space: nowrap; }
.empty-cell { text-align: center; color: var(--text-placeholder); padding: 24px !important; }

/* 动作标签 */
.action-tag {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: var(--bg-secondary);
  color: var(--text-primary);
}

/* 对话框覆盖层 */
.dialog-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.4);
  z-index: 1000;
  display: flex; align-items: center; justify-content: center;
}
.dialog-panel {
  width: 640px;
  max-height: 80vh;
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
.dialog-body { padding: 20px; overflow-y: auto; flex: 1; }
.dialog-footer {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 12px 20px; border-top: 1px solid var(--border-color);
}

/* 表单 */
.form-row { display: flex; gap: 16px; margin-bottom: 16px; }
.form-field { display: flex; flex-direction: column; gap: 4px; }
.form-field label { font-size: 12px; font-weight: 500; color: var(--text-secondary); }
.field-input, .field-select {
  padding: 8px 12px; border: 1px solid var(--border-color);
  border-radius: var(--radius); font-size: 13px; outline: none;
}
.field-input:focus, .field-select:focus { border-color: var(--accent-color); }

/* 条件编辑区 */
.conditions-list { display: flex; flex-direction: column; gap: 8px; margin-bottom: 12px; }
.condition-row { display: flex; gap: 8px; align-items: center; }

/* 动作配置区 */
.action-row { display: flex; gap: 12px; align-items: center; }

/* 按钮 */
.btn-primary {
  padding: 8px 20px; border: none; border-radius: var(--radius);
  background: var(--accent-color); color: #fff; font-size: 13px; cursor: pointer;
}
.btn-primary:hover { background: var(--accent-hover); }
.btn-secondary {
  padding: 8px 20px; border: 1px solid var(--border-color); border-radius: var(--radius);
  background: transparent; color: var(--text-primary); font-size: 13px; cursor: pointer;
}
.btn-secondary:hover { background: var(--bg-hover); }
.btn-sm {
  padding: 6px 12px; border: 1px solid var(--border-color); background: transparent;
  border-radius: var(--radius-sm); font-size: 12px; cursor: pointer;
}
.btn-sm:hover { background: var(--bg-hover); }
.icon-btn {
  width: 28px; height: 28px; border: none; background: transparent;
  border-radius: 4px; cursor: pointer; font-size: 14px;
}
.icon-btn:hover { background: var(--bg-hover); }
.icon-btn.danger { color: var(--danger-color); }
.close-btn {
  width: 32px; height: 32px; border: none; background: transparent;
  font-size: 20px; cursor: pointer; border-radius: var(--radius-sm);
  display: flex; align-items: center; justify-content: center;
}
.close-btn:hover { background: var(--bg-hover); }

/* 工具类 */
.flex-grow { flex: 1; }
.w-32 { width: 120px; }
.w-36 { width: 140px; }
.w-48 { width: 180px; }
</style>
