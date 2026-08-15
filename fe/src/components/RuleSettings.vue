<template>
  <SettingsCard :title="lang.auto_rules" :description="lang.auto_rules_desc">

    <div class="settings__table-container">
      <el-table :data="data" class="settings__table" style="width: 100%">
        <el-table-column prop="name" :label="lang.rule_name" min-width="150" show-overflow-tooltip/>
        <el-table-column prop="action" :label="lang.rule_do" width="120">
          <template #default="scope">
            <el-tag size="small" effect="plain" class="settings-tag">{{ ActionName[scope.row.action] }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="params" :label="lang.rule_params" min-width="150" show-overflow-tooltip/>
        <el-table-column prop="sort" :label="lang.rule_priority" width="100"/>
        <el-table-column width="100" align="right">
          <template #default="scope">
            <div class="settings__row-actions">
              <el-button size="small" type="primary" :icon="Edit" circle text @click="editRule(scope.row)"/>
              <el-popconfirm confirm-button-text="Yes" cancel-button-text="No" :icon="InfoFilled"
                             @confirm="delRule(scope.row.id)" icon-color="#ef4444" :title="lang.del_rule_confirm">
                <template #reference>
                  <el-button size="small" type="danger" :icon="Delete" circle text/>
                </template>
              </el-popconfirm>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="settings__form-actions">
      <el-button type="primary" @click="dialogVisible = true" class="settings__btn">
        <el-icon><Plus /></el-icon> {{ lang.new_rule }}
      </el-button>
    </div>

    <el-dialog v-model="dialogVisible" :title="addRuleForm.id > 0 ? 'Edit Rule' : lang.new_rule" width="700px" class="settings__dialog">
      <div class="settings__dialog-content">
        <el-form :model="addRuleForm" label-position="top">
          <div class="form-row">
            <el-form-item :label="lang.rule_name" class="flex-grow">
              <el-input v-model="addRuleForm.name" placeholder="Rule Name"/>
            </el-form-item>
            <el-form-item :label="lang.rule_priority" class="w-32">
              <el-input v-model="addRuleForm.sort" type="number" oninput="value=value.replace(/[^\-\d]/g, '')"/>
            </el-form-item>
          </div>
          
          <el-divider class="custom-divider"/>
          <div class="section-title">{{ lang.rule_desc }} (Conditions)</div>
          
          <div class="rules-list">
            <div v-for="(rule, index) in addRuleForm.rules" :key="index" class="rule-line">
              <el-select v-model="rule.field" placeholder="Field" class="w-32">
                <el-option key="From" :label="lang.from" value="From"/>
                <el-option key="Subject" :label="lang.subject" value="Subject"/>
                <el-option key="To" :label="lang.to" value="To"/>
                <el-option key="Cc" :label="lang.cc" value="Cc"/>
                <el-option key="Content" :label="lang.content" value="Content"/>
              </el-select>

              <el-select v-model="rule.type" placeholder="Type" class="w-32">
                <el-option key="equal" :label="lang.equal" value="equal"/>
                <el-option key="contains" :label="lang.contains" value="contains"/>
                <el-option key="regex" :label="lang.regex" value="regex"/>
              </el-select>

              <el-input v-model="rule.rule" placeholder="Value" class="flex-grow"/>
              <el-button type="danger" :icon="Delete" @click="removeRuleLine(index)" text circle/>
            </div>
          </div>
          <div class="add-condition-action">
            <el-button size="small" type="primary" plain @click="addRule()"><el-icon><Plus/></el-icon> Add Condition</el-button>
          </div>

          <el-divider class="custom-divider"/>
          <div class="section-title">{{ lang.rule_do }} (Action)</div>
          
          <div class="action-row">
            <el-select v-model="addRuleForm.action" placeholder="Select Action" @change="ruleTypeChange" class="w-48">
              <el-option key="mark_read" :label="lang.mark_read" :value="READ"/>
              <el-option key="move" :label="lang.move" :value="MOVE"/>
              <el-option key="delete" :label="lang.delete" :value="DELETE"/>
              <el-option key="forward" :label="lang.forward" :value="FORWARD"/>
            </el-select>
            
            <el-select v-if="addRuleForm.action === 4" v-model="addRuleForm.params" @click="refreshGroupInfos" placeholder="Select Folder" class="flex-grow">
              <el-option v-for="gp in groupData.list" :key="gp.id" :label="gp.name" :value="gp.id"/>
            </el-select>

            <el-input v-if="addRuleForm.action === 2" v-model="addRuleForm.params" class="flex-grow" placeholder="Forward Email Address"/>
          </div>
        </el-form>
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">Cancel</el-button>
          <el-button type="primary" @click="submitRule()">
            {{ lang.submit }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </SettingsCard>
</template>

<script setup lang="ts">
import {reactive, ref} from 'vue';
import lang from '../i18n/i18n';
import {Delete, Edit, InfoFilled, Plus} from '@element-plus/icons-vue'
import {ruleService} from "@/services/ruleService";
import {groupService} from "@/services/groupService";
import {ElNotification} from "element-plus";
import SettingsCard from "@/components/settings/SettingsCard.vue";

const data = ref<any[]>([])
const dialogVisible = ref(false)
const READ = 1
const FORWARD = 2
const DELETE = 3
const MOVE = 4

const ActionName: Record<number, string> = {
  1: lang.mark_read,
  2: lang.forward,
  3: lang.delete,
  4: lang.move
}

/** 通过 ruleService 获取规则列表 */
const init = function () {
  ruleService.getRules().then((res: any) => {
    data.value = res.data || []
  })
}

init()

const groupData = reactive<{
  list: any[]
}>({
  list: []
})

/** 通过 groupService 获取分组平铺列表 */
const refreshGroupInfos = function () {
  groupService.getGroupList().then((res: any) => {
    if (res.data != null) {
      groupData.list = res.data
      for (let i = 0; i < groupData.list.length; i++) {
        groupData.list[i].id += ""
      }
    }
  })
}

refreshGroupInfos()

const addRuleForm = reactive<{
  id: number;
  name: string;
  sort: number;
  rules: {field: string; type: string; rule: string}[];
  action: number | string;
  params: string;
}>({
  "id": 0,
  "name": "",
  "sort": 0,
  "rules": [{"field": "", "type": "", "rule": ""}],
  "action": "",
  "params": ""
})

/** 通过 ruleService 删除规则 */
const delRule = function (id: number) {
  ruleService.deleteRule(id).then((res: any) => {
    ElNotification({
      title: res.errorNo === 0 ? lang.succ : lang.fail,
      message: res.data,
      type: res.errorNo === 0 ? 'success' : 'error',
    })
    init()
  })
}

const editRule = function (ruleInfo: any) {
  addRuleForm.id = ruleInfo.id
  addRuleForm.name = ruleInfo.name
  addRuleForm.rules = ruleInfo.rules || []
  addRuleForm.action = ruleInfo.action
  addRuleForm.params = ruleInfo.params
  addRuleForm.sort = ruleInfo.sort
  dialogVisible.value = true
}

const removeRuleLine = function (index: number) {
  addRuleForm.rules.splice(index, 1);
}

const addRule = function () {
  addRuleForm.rules.push({"field": "", "type": "", "rule": ""})
}

/** 通过 ruleService 新增或更新规则 */
const submitRule = function () {
  addRuleForm.sort = parseInt(String(addRuleForm.sort))

  const request = addRuleForm.id > 0
    ? ruleService.updateRule(addRuleForm as any)
    : ruleService.addRule(addRuleForm as any);

  request.then((res: any) => {
    if (res.errorNo !== 0) {
      ElNotification({
        title: lang.fail,
        message: res.data,
        type: 'error',
      })
    } else {
      init()
      dialogVisible.value = false
      addRuleForm.id = 0
      addRuleForm.name = ""
      addRuleForm.sort = 0
      addRuleForm.rules = [{"field": "", "type": "", "rule": ""}]
      addRuleForm.action = ""
      addRuleForm.params = ""
    }
  })
}

const ruleTypeChange = function () {
  addRuleForm.params = ''
}
</script>

<style scoped>
/* 引入设置组件公共样式（table-container、dialog、form-actions 等） */
@import '@/assets/settings-common.css';

/* 规则编辑器独有样式 */
.form-row {
  display: flex;
  gap: 16px;
}

.flex-grow {
  flex-grow: 1;
}

.w-32 {
  width: 120px;
}

.w-48 {
  width: 180px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--ifm-color-content);
  margin-bottom: 16px;
}

.rules-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 12px;
}

.rule-line {
  display: flex;
  gap: 8px;
  align-items: center;
}

.add-condition-action {
  margin-top: 12px;
}

.action-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.custom-divider {
  margin: 24px 0;
  border-color: var(--ifm-border-color);
}
</style>
