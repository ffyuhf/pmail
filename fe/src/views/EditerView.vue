<template>
  <!-- 写信视图（Docusaurus BEM 风格） -->
  <div class="composer" :class="{ 'composer--fullscreen': isFullscreen }">
    <div class="composer__header">
      <div class="composer__title-group">
        <!-- 返回按钮：隐藏 Sidebar 后需提供导航回列表的方式 -->
        <el-icon class="composer__back-btn" @click="router.back()" :size="20"><ArrowLeft /></el-icon>
        <h2 class="composer__title">{{ lang.compose }}</h2>
      </div>
      <div class="composer__actions">
        <el-button class="composer__action-btn" @click="openCsvDialog">
          <el-icon><Upload /></el-icon> {{ lang.csv_import }}
        </el-button>
        <el-button class="composer__action-btn" @click="upload">
          <el-icon><Paperclip /></el-icon> {{ lang.add_att }}
        </el-button>
        <el-button type="primary" class="composer__send-btn" @click="send(ruleFormRef)">
          <el-icon><Position /></el-icon> {{ lang.send }}
        </el-button>
        <input v-show="false" ref="fileRef" type="file" @change="fileChange" multiple>
      </div>
    </div>

    <div class="composer__body">
      <el-form :rules="rules" ref="ruleFormRef" :model="ruleForm" class="composer__form">

        <div class="composer__row">
          <div class="composer__field-label">{{ lang.sender }}</div>
          <el-form-item prop="sender" class="composer__flex-grow composer__m0">
            <el-popover trigger="click" :width="400" placement="bottom-start">
              <template #reference>
                <div class="composer__sender-selector">
                  <span class="composer__sender-name">{{ ruleForm.nickName }}</span>
                  <span class="composer__sender-email">{{ '<' + ruleForm.sender + '@' + ruleForm.pickDomain + '>' }}</span>
                  <el-icon class="composer__arrow-down"><ArrowDown /></el-icon>
                </div>
              </template>
              <template #default>
                <div class="composer__sender-edit">
                  <div class="composer__edit-row">
                    <span class="composer__edit-label">Prefix</span>
                    <el-input
                      :disabled="!(globalStatus.userInfos.is_admin)"
                      v-model="ruleForm.sender"
                      :placeholder="lang.sender_desc"
                    />
                  </div>
                  <div class="composer__edit-row">
                    <span class="composer__edit-label">Domain</span>
                    <el-select v-model="ruleForm.pickDomain" class="composer__w-full">
                      <el-option :value="item" v-for="item in ruleForm.domains" :key="item">{{ item }}</el-option>
                    </el-select>
                  </div>
                  <div class="composer__edit-row">
                    <span class="composer__edit-label">{{ lang.nick_name }}</span>
                    <el-input v-model="ruleForm.nickName"/>
                  </div>
                </div>
              </template>
            </el-popover>
          </el-form-item>
        </div>

        <div class="composer__row">
          <div class="composer__field-label">{{ lang.to }}</div>
          <el-form-item prop="receivers" class="composer__flex-grow composer__m0 composer__borderless-select">
            <el-select
              v-model="ruleForm.receivers"
              multiple filterable allow-create :reserve-keyword="false"
              placeholder="Recipients..."
            ></el-select>
          </el-form-item>
        </div>

        <div class="composer__row" v-if="ruleForm.cc.length > 0 || showCcBcc">
          <div class="composer__field-label">{{ lang.cc }}</div>
          <el-form-item prop="cc" class="composer__flex-grow composer__m0 composer__borderless-select">
            <el-select
              v-model="ruleForm.cc"
              multiple filterable allow-create :reserve-keyword="false"
              placeholder="Cc..."
            ></el-select>
          </el-form-item>
        </div>

        <div class="composer__row" v-if="ruleForm.bcc.length > 0 || showCcBcc">
          <div class="composer__field-label">{{ lang.bcc }}</div>
          <el-form-item prop="bcc" class="composer__flex-grow composer__m0 composer__borderless-select">
            <el-select
              v-model="ruleForm.bcc"
              multiple filterable allow-create :reserve-keyword="false"
              placeholder="Bcc..."
            ></el-select>
          </el-form-item>
        </div>

        <div class="composer__row composer__row--subject">
          <el-form-item prop="subject" class="composer__w-full composer__m0 composer__borderless-input">
            <el-input v-model="ruleForm.subject" placeholder="Subject"></el-input>
            <div class="composer__cc-bcc-toggle" @click="showCcBcc = !showCcBcc" v-if="!showCcBcc">Cc/Bcc</div>
          </el-form-item>
        </div>

        <!-- 编辑器区域：v-if="editorReady" 延迟渲染，避免在 transition opacity:0 阶段初始化导致空白 -->
        <div class="composer__editor-wrapper" v-if="editorReady">
          <Toolbar class="composer__editor-toolbar" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode"/>
          <Editor class="composer__editor-content" v-model="valueHtml" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated"/>
        </div>
        <!-- 编辑器加载占位：transition 完成前显示加载提示 -->
        <div class="composer__editor-wrapper composer__editor-loading" v-else>
          <div class="composer__loading-placeholder">Loading editor...</div>
        </div>

        <div class="composer__attachments" v-if="fileList.length > 0">
          <div class="composer__att-chip" v-for="(item, index) in fileList" :key="index">
            <el-icon class="composer__att-icon"><Document /></el-icon>
            <span class="composer__att-name">{{ item.name }}</span>
            <el-icon class="composer__att-remove" @click="delFile(index)"><Close/></el-icon>
          </div>
        </div>
      </el-form>
    </div>

    <!-- CSV 导入对话框 -->
    <el-dialog
      v-model="csvDialogVisible"
      :title="lang.csv_import"
      width="680px"
      destroy-on-close
      class="csv-dialog"
    >
      <!-- 文件上传区域 -->
      <div
        class="csv-upload-area"
        :class="{ 'csv-upload-area--dragover': csvDragOver }"
        @dragover.prevent="csvDragOver = true"
        @dragleave.prevent="csvDragOver = false"
        @drop.prevent="handleCsvDrop"
        @click="triggerCsvFileInput"
      >
        <el-icon class="csv-upload-area__icon" :size="40"><Upload /></el-icon>
        <p class="csv-upload-area__text">{{ lang.csv_upload_area }}</p>
        <p class="csv-upload-area__hint">CSV (*.csv)</p>
        <input v-show="false" ref="csvFileRef" type="file" accept=".csv" @change="handleCsvFileChange">
      </div>

      <!-- 目标字段选择 + 统计信息 -->
      <div class="csv-toolbar" v-if="csvRecipients.length > 0">
        <div class="csv-toolbar__left">
          <span class="csv-toolbar__label">{{ lang.csv_target_field }}</span>
          <el-select v-model="csvTargetField" style="width: 120px">
            <el-option value="to" :label="lang.to" />
            <el-option value="cc" :label="lang.cc" />
            <el-option value="bcc" :label="lang.bcc" />
          </el-select>
        </div>
        <div class="csv-toolbar__right">
          <el-button size="small" @click="csvToggleSelectAll">
            {{ csvAllSelected ? lang.csv_deselect_all : lang.csv_select_all }}
          </el-button>
          <span class="csv-toolbar__count">
            {{ lang.csv_selected_count.replace('{count}', String(csvSelectedCount)) }}
          </span>
        </div>
      </div>

      <!-- 预览表格 -->
      <el-table
        v-if="csvRecipients.length > 0"
        :data="csvRecipients"
        class="csv-preview-table"
        max-height="320px"
        @selection-change="handleCsvSelectionChange"
        ref="csvTableRef"
      >
        <el-table-column type="selection" width="50" />
        <el-table-column property="name" :label="lang.csv_name_col" min-width="120">
          <template #default="{ row }">
            {{ row.name || '-' }}
          </template>
        </el-table-column>
        <el-table-column property="email" :label="lang.csv_email_col" min-width="220" />
      </el-table>

      <!-- 解析结果为空时的提示 -->
      <div class="csv-empty" v-if="csvParseFailed">
        <el-icon :size="32" color="#f56c6c"><CircleCloseFilled /></el-icon>
        <p>{{ lang.csv_no_email_found }}</p>
      </div>

      <template #footer>
        <el-button @click="csvDialogVisible = false">{{ lang.fail === 'Fail!' ? 'Cancel' : '取消' }}</el-button>
        <el-button
          type="primary"
          :disabled="csvSelectedCount === 0"
          @click="confirmCsvImport"
        >
          {{ lang.csv_confirm_import }} ({{ csvSelectedCount }})
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import '@wangeditor/editor/dist/css/style.css'
import {ElMessage} from 'element-plus'
import {computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, shallowRef, watch} from 'vue'
import {Close, Paperclip, Position, ArrowDown, Document, ArrowLeft, Upload, CircleCloseFilled} from '@element-plus/icons-vue';
import lang from '../i18n/i18n';
import {Editor, Toolbar} from '@wangeditor/editor-for-vue'
import {i18nChangeLanguage} from '@wangeditor/editor'
import {useRoute, useRouter} from 'vue-router';
import {emailService} from "@/services/emailService";
import useGroupStore from '../stores/group'
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";
import {createStringEmailValidator} from "@/utils/validators";
import {parseCsv, readCsvFile} from "@/utils/csvParser";
import type {CsvRecipient} from "@/utils/csvParser";

const router = useRouter();
const route = useRoute();
const groupStore = useGroupStore()
const globalStatus = useGlobalStatusStore();
const showCcBcc = ref(false);
const isFullscreen = ref(false);

/**
 * 编辑器延迟渲染控制：等待 App.vue 的 page-fade transition（250ms）完成后再挂载 wangeditor。
 * wangeditor 在 opacity:0 的容器中初始化时 getBoundingClientRect() 返回尺寸为 0，导致编辑器空白。
 * 延迟 300ms 确保组件完全可见后再创建编辑器实例。新增日期: 20260516 v1.2
 */
const editorReady = ref(false)
onMounted(() => {
  setTimeout(() => {
    editorReady.value = true
  }, 300)
})

if (lang.lang === "zhCn") {
  i18nChangeLanguage('zh-CN')
} else {
  i18nChangeLanguage('en')
}

// 内容 HTML
const valueHtml = ref('<p><br></p>')

const toolbarConfig: Record<string, any> = {}
const editorConfig: Record<string, any> = {
  MENU_CONF: {},
  placeholder: 'Write your message...'
}

editorConfig.MENU_CONF['uploadImage'] = {
  base64LimitSize: 100 * 1024 * 1024 * 1024,
}
const mode = ref('default')
const fileRef = ref();
const ruleFormRef = ref()
const ruleForm = reactive<{
  nickName: string;
  sender: string;
  receivers: string[];
  cc: string[];
  bcc: string[];
  subject: string;
  domains: string[];
  pickDomain: string;
}>({
  nickName: '',
  sender: '',
  receivers: [],
  cc: [],
  bcc: [],
  subject: '',
  domains: [],
  pickDomain: ""
})
const fileList = reactive<{name: string; data: string | ArrayBuffer | null}[]>([]);

const init = function () {
    if ( Object.keys(globalStatus.userInfos).length===0 || globalStatus.userInfos === null || globalStatus.userInfos == undefined ){
      globalStatus.init(()=>{
        ruleForm.sender = globalStatus.userInfos.account
        ruleForm.domains = globalStatus.userInfos.domains
        ruleForm.pickDomain = globalStatus.userInfos.domains[0]
        ruleForm.nickName = globalStatus.userInfos.name
        /** 回信预填：init 异步回调后填充回信参数。新增日期: 20260516 */
        fillReplyParams()
      })
    }else{
      ruleForm.sender = globalStatus.userInfos.account
        ruleForm.domains = globalStatus.userInfos.domains
        ruleForm.pickDomain = globalStatus.userInfos.domains[0]
        ruleForm.nickName = globalStatus.userInfos.name
      /** 回信预填：同步初始化后填充回信参数。新增日期: 20260516 */
      fillReplyParams()
    }
}
init()

/**
 * 回信参数预填：从 route.query 读取回信参数，自动填充收件人、主题、发件人。
 * 仅在存在 reply_to 参数时执行（即从邮件详情页点击回信跳转而来）。
 * 新增日期: 20260516
 */
const fillReplyParams = function () {
  const replyTo = route.query.reply_to as string
  const replySubject = route.query.reply_subject as string
  const replySender = route.query.reply_sender as string
  const replyDomain = route.query.reply_domain as string

  if (!replyTo) return

  /** 收件人 = 原邮件的发件人地址 */
  ruleForm.receivers = [replyTo]

  /** 主题 = Re: + 原主题 */
  if (replySubject) {
    ruleForm.subject = replySubject
  }

  /** 发件人 = 原邮件的收件人地址（即收到该邮件的账号），需验证域名在用户允许列表中 */
  if (replySender && replyDomain) {
    if (ruleForm.domains.includes(replyDomain)) {
      ruleForm.sender = replySender
      ruleForm.pickDomain = replyDomain
    }
  }
}

/**
 * 监听路由 query 变化：当从详情页回信跳转至编辑器时，
 * 确保 fillReplyParams 被调用（双重保障，防止 transition 异常导致 init 回调未触发）。
 * 新增日期: 20260516 v1.1
 */
watch(() => route.query, (newQuery: Record<string, string | string[] | undefined>) => {
  if (newQuery.reply_to) {
    fillReplyParams()
  }
})

/** 发件人验证：前缀不能为空且不能包含 @ */
const validateSender = function (rule: any, value: any, callback: any) {
  if (typeof ruleForm.sender === "undefined" || ruleForm.sender === null || ruleForm.sender.trim() === "") {
    callback(new Error(lang.err_sender_must))
  } else if (ruleForm.sender.includes("@")) {
    callback(new Error(lang.only_prefix))
  } else {
    callback()
  }
}

/** 使用 createStringEmailValidator 工厂函数替代重复的 validateReceivers/validateCc/validateBcc */
const rules = reactive({
  sender: [{validator: validateSender, trigger: 'change'}],
  receivers: [{validator: createStringEmailValidator(() => ruleForm.receivers, lang.err_email_format), trigger: 'change'}],
  cc: [{validator: createStringEmailValidator(() => ruleForm.cc, lang.err_email_format), trigger: 'change'}],
  bcc: [{validator: createStringEmailValidator(() => ruleForm.bcc, lang.err_email_format), trigger: 'change'}],
  subject: [{required: true, message: lang.err_title_must, trigger: 'change'}],
})

const editorRef = shallowRef()
onBeforeUnmount(() => {
  const editor = editorRef.value
  if (editor == null) return
  editor.destroy()
})

const handleCreated = (editor: any) => {
  editorRef.value = editor

  // 覆写 wangeditor 内置全屏方法：让全屏作用于整个 composer 而非仅 editor-wrapper
  // 原始实现只给 toolbar 和 editor 的父容器加全屏样式，会丢失收件人、主题、附件等区域
  editor.fullScreen = () => {
    editor.isFullScreen = true
    isFullscreen.value = true
  }
  editor.unFullScreen = () => {
    editor.isFullScreen = false
    isFullscreen.value = false
  }
}

const send = function (formEl: any) {
  if (!formEl) return
  formEl.validate((valid: boolean) => {
    if (valid) {
      if(ruleForm.receivers.length === 0 && ruleForm.cc.length === 0 && ruleForm.bcc.length === 0) {
        ElMessage.warning("Please specify at least one recipient");
        return;
      }
      let objectTos = ruleForm.receivers.map(e => ({name: "", email: e}));
      let objectCcs = ruleForm.cc.map(e => ({name: "", email: e}));
      let objectBccs = ruleForm.bcc.map(e => ({name: "", email: e}));

      let text = editorRef.value.getText()

      /** 通过 emailService 发送邮件 */
      emailService.sendEmail({
        from: {name: ruleForm.nickName, email: ruleForm.sender + "@" + ruleForm.pickDomain},
        to: objectTos,
        cc: objectCcs,
        bcc: objectBccs,
        subject: ruleForm.subject,
        text: text,
        html: valueHtml.value,
        attrs: fileList
      }).then((res: any) => {
        if (res.errorNo === 0) {
          ElMessage.success(lang.succ_send)
          groupStore.name = lang.outbox
          groupStore.tag = '{"type":1,"status":-1}'
          router.replace({name: 'list'})
        } else {
          ElMessage.error(res.data)
        }
      })
    } else {
      return false
    }
  })
}

const upload = function () {
  fileRef.value.dispatchEvent(new MouseEvent('click'))
}

const fileChange = function (e: any) {
  let files = e.target.files || e.dataTransfer.files;
  if (!files.length) return;
  for (let i = 0; i < files.length; i++) {
    const reader = new FileReader();
    reader.onload = function fileReadCompleted() {
      fileList.push({
        name: files[i].name,
        data: this.result
      })
    };
    reader.readAsDataURL(files[i]);
  }
}

const delFile = function (index: number) {
  fileList.splice(index, 1);
}

// ============ CSV 导入功能 ============

/** CSV 对话框可见性 */
const csvDialogVisible = ref(false);

/** CSV 文件输入框引用 */
const csvFileRef = ref();

/** CSV 拖拽悬停状态 */
const csvDragOver = ref(false);

/** CSV 解析出的收件人列表 */
const csvRecipients = ref<CsvRecipient[]>([]);

/** CSV 选中的收件人列表 */
const csvSelectedRecipients = ref<CsvRecipient[]>([]);

/** CSV 选中数量 */
const csvSelectedCount = computed(() => csvSelectedRecipients.value.length);

/** 是否全选 */
const csvAllSelected = computed(() =>
  csvRecipients.value.length > 0 && csvSelectedRecipients.value.length === csvRecipients.value.length
);

/** 解析失败标记 */
const csvParseFailed = ref(false);

/** 导入目标字段（to/cc/bcc） */
const csvTargetField = ref<'to' | 'cc' | 'bcc'>('to');

/** 表格引用 */
const csvTableRef = ref();

/** 打开 CSV 导入对话框 */
const openCsvDialog = function () {
  csvDialogVisible.value = true;
  csvRecipients.value = [];
  csvSelectedRecipients.value = [];
  csvParseFailed.value = false;
  csvTargetField.value = 'to';
}

/** 触发 CSV 文件选择 */
const triggerCsvFileInput = function () {
  csvFileRef.value?.dispatchEvent(new MouseEvent('click'));
}

/** 处理 CSV 文件选择事件 */
const handleCsvFileChange = async function (e: Event) {
  const input = e.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  await processCsvFile(file);
  // 重置 input，允许再次选择同一文件
  input.value = '';
}

/** 处理 CSV 文件拖拽放置 */
const handleCsvDrop = async function (e: DragEvent) {
  csvDragOver.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (!file) return;
  await processCsvFile(file);
}

/** 解析 CSV 文件并更新预览列表 */
const processCsvFile = async function (file: File) {
  // 文件大小校验（5MB）
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.warning(lang.csv_file_too_large);
    return;
  }

  try {
    const text = await readCsvFile(file);
    const result = parseCsv(text);

    csvRecipients.value = result.recipients;
    csvSelectedRecipients.value = [];
    csvParseFailed.value = result.recipients.length === 0 && result.totalRows > 0;

    // 默认全选：等表格渲染后勾选所有行
    if (result.recipients.length > 0) {
      nextTick(() => {
        csvTableRef.value?.toggleAllSelection?.();
      });
    }
  } catch (err) {
    ElMessage.error(String(err));
    csvRecipients.value = [];
    csvParseFailed.value = true;
  }
}

/** 处理表格选中变化 */
const handleCsvSelectionChange = function (selection: CsvRecipient[]) {
  csvSelectedRecipients.value = selection;
}

/** 切换全选/取消全选 */
const csvToggleSelectAll = function () {
  csvTableRef.value?.toggleAllSelection?.();
}

/** 确认导入：将选中的邮箱追加到目标字段 */
const confirmCsvImport = function () {
  if (csvSelectedRecipients.value.length === 0) return;

  const emailsToAdd = csvSelectedRecipients.value.map(r => r.email);

  // 根据目标字段追加到对应数组（去重）
  const targetArray = csvTargetField.value === 'to'
    ? ruleForm.receivers
    : csvTargetField.value === 'cc'
      ? ruleForm.cc
      : ruleForm.bcc;

  const existingSet = new Set(targetArray.map(e => e.toLowerCase()));
  let addedCount = 0;

  for (const email of emailsToAdd) {
    if (!existingSet.has(email.toLowerCase())) {
      targetArray.push(email);
      existingSet.add(email.toLowerCase());
      addedCount++;
    }
  }

  ElMessage.success(lang.csv_import_success.replace('{count}', String(addedCount)));
  csvDialogVisible.value = false;
}
</script>

<!-- 样式: Docusaurus BEM 风格 | 重构日期: 20260429 -->
<style scoped>
/* 写信容器 */
.composer {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--ifm-background-color);
  border-radius: var(--ifm-card-border-radius);
  box-shadow: var(--ifm-global-shadow-lw);
  border: 1px solid var(--ifm-border-color);
  overflow: hidden;
}

.composer__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--ifm-border-color);
  background: var(--ifm-background-surface-color);
}

/* 标题组：返回按钮 + 标题横向排列 */
.composer__title-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

/* 返回按钮：隐藏 Sidebar 后的导航替代方式 */
.composer__back-btn {
  cursor: pointer;
  color: var(--ifm-color-content-secondary);
  transition: color var(--ifm-transition-fast, 0.1s);
}

.composer__back-btn:hover {
  color: var(--ifm-color-primary);
}

.composer__title {
  font-size: 28px;
  font-weight: 600;
  color: var(--ifm-color-content);
  margin: 0;
  letter-spacing: -0.02em;
}

.composer__actions {
  display: flex;
  gap: 12px;
}

.composer__action-btn {
  background: var(--ifm-background-color);
  border: 1px solid var(--ifm-border-color);
  color: var(--ifm-color-content-secondary);
}

.composer__action-btn:hover {
  background: var(--ifm-background-surface-color);
  color: var(--ifm-color-content);
  border-color: var(--ifm-border-color);
}

.composer__send-btn {
  box-shadow: 0 8px 16px rgba(0, 113, 227, 0.2);
}

.composer__body {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.composer__form {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* 表单行：标签 + 输入框横向排列 */
.composer__row {
  display: flex;
  align-items: center;
  border-bottom: 1px solid var(--ifm-color-emphasis-300);
  padding: 6px 20px;
}

/* 主题行：无标签，输入框占满 */
.composer__row--subject {
  position: relative;
}

.composer__field-label {
  width: 70px;
  color: var(--ifm-color-content-secondary);
  font-weight: 600;
  font-size: 13px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

/* 工具类 */
.composer__flex-grow { flex-grow: 1; }
.composer__m0 { margin: 0; }
.composer__w-full { width: 100%; }

/* 发件人选择器 */
.composer__sender-selector {
  display: inline-flex;
  align-items: center;
  padding: 8px 12px;
  border-radius: 12px;
  cursor: pointer;
  background-color: var(--ifm-background-surface-color);
  border: 1px solid var(--ifm-border-color);
  transition: background 0.2s;
}

.composer__sender-selector:hover {
  background-color: var(--ifm-background-color);
  transform: translateY(-1px);
}

.composer__sender-name {
  font-weight: 600;
  margin-right: 6px;
  color: var(--ifm-color-content);
}

.composer__sender-email {
  color: var(--ifm-color-content-secondary);
  margin-right: 8px;
}

.composer__arrow-down {
  font-size: 12px;
  color: var(--ifm-color-content-muted);
}

/* 发件人编辑卡片（Popover 内容） */
.composer__sender-edit {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.composer__edit-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.composer__edit-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--ifm-color-content-secondary);
}

/* 无边框输入框：使表单行看起来一体化 */
.composer__borderless-select :deep(.el-input__wrapper),
.composer__borderless-input :deep(.el-input__wrapper) {
  box-shadow: none !important;
  background: transparent;
  padding-left: 0;
}

.composer__borderless-select :deep(.el-select .el-input__wrapper) {
  background: var(--ifm-background-surface-color) !important;
  box-shadow: 0 0 0 1px var(--ifm-border-color) inset !important;
  border-radius: 10px;
  padding: 2px 10px;
}

.composer__borderless-select :deep(.el-select .el-input__wrapper:hover),
.composer__borderless-select :deep(.el-select .el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px var(--ifm-color-primary) inset !important;
  background: var(--ifm-background-color) !important;
}

.composer__borderless-select :deep(.el-select__tags .el-tag) {
  background: var(--ifm-background-surface-color);
  border-color: var(--ifm-border-color);
  color: var(--ifm-color-content);
}

.composer__borderless-select :deep(.el-input__inner),
.composer__borderless-input :deep(.el-input__inner) {
  font-size: 15px;
}

/* Cc/Bcc 切换按钮 */
.composer__cc-bcc-toggle {
  position: absolute;
  right: 0;
  top: 50%;
  transform: translateY(-50%);
  font-size: 13px;
  color: var(--ifm-color-primary);
  cursor: pointer;
  font-weight: 500;
}

.composer__cc-bcc-toggle:hover {
  text-decoration: underline;
}

/* 编辑器区域 */
.composer__editor-wrapper {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  min-height: 300px;
  --w-e-toolbar-bg-color: var(--ifm-background-surface-color);
  --w-e-toolbar-color: var(--ifm-color-content);
  --w-e-toolbar-active-color: var(--ifm-color-primary);
  --w-e-toolbar-active-bg-color: var(--ifm-row-hover-background);
  --w-e-textarea-bg-color: transparent;
  --w-e-textarea-color: var(--ifm-color-content);
  --w-e-textarea-slight-bg-color: transparent;
  --w-e-textarea-slight-color: var(--ifm-color-content-muted);
  --w-e-border-color: var(--ifm-color-emphasis-300);
}

.composer__editor-toolbar {
  border-bottom: 1px solid var(--ifm-color-emphasis-300);
  background-color: var(--ifm-background-surface-color) !important;
}

.composer__editor-content {
  flex-grow: 1;
  padding: 14px 20px;
  overflow-y: hidden;
}

.composer__editor-content :deep(.w-e-text-container),
.composer__editor-content :deep(.w-e-scroll),
.composer__editor-content :deep(.w-e-text) {
  background: transparent !important;
  color: var(--ifm-color-content) !important;
}

.composer__editor-content :deep(.w-e-text-placeholder) {
  color: var(--ifm-color-content-muted) !important;
}

/* 编辑器加载占位符：transition 完成前的占位显示 */
.composer__editor-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 300px;
  background: var(--ifm-background-surface-color);
}

.composer__loading-placeholder {
  color: var(--ifm-color-content-muted);
  font-size: 14px;
}

/* 附件预览区 */
.composer__attachments {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 14px 20px;
  border-top: 1px solid var(--ifm-border-color);
  background: var(--ifm-background-surface-color);
}

.composer__att-chip {
  display: flex;
  align-items: center;
  background: var(--ifm-background-color);
  border: 1px solid var(--ifm-border-color);
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 13px;
  box-shadow: var(--ifm-global-shadow-lw);
  transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), box-shadow 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}

.composer__att-chip:hover {
  transform: translateY(-1px);
  box-shadow: var(--ifm-global-shadow-md);
}

.composer__att-icon {
  margin-right: 6px;
  color: var(--ifm-color-content-secondary);
}

.composer__att-name {
  max-width: 150px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: 8px;
}

.composer__att-remove {
  cursor: pointer;
  color: var(--ifm-color-content-muted);
}

.composer__att-remove:hover {
  color: #ef4444;
}

/* Dark 模式下编辑器修正 */
@media (prefers-color-scheme: dark) {
  .composer__editor-wrapper {
    background: var(--ifm-background-surface-color);
  }

  .composer__editor-content :deep(.w-e-text-container) {
    border-left-color: transparent !important;
    border-right-color: transparent !important;
  }

  .composer__editor-content :deep(.w-e-bar) {
    background: var(--ifm-background-surface-color) !important;
    color: var(--ifm-color-content) !important;
  }

  .composer__editor-content :deep(.w-e-bar-item button),
  .composer__editor-content :deep(.w-e-menu-tooltip-v5) {
    color: var(--ifm-color-content) !important;
  }
}

/* 全屏模式：整个写信组件铺满浏览器窗口 */
.composer--fullscreen {
  position: fixed !important;
  top: 0 !important;
  left: 0 !important;
  right: 0 !important;
  bottom: 0 !important;
  z-index: 9999;
  width: 100vw !important;
  height: 100vh !important;
  border-radius: 0;
  border: none;
}

@media (max-width: 768px) {
  .composer__sender-selector {
    padding: 6px;
    max-width: 100%;
    overflow: hidden;
  }
  .composer__sender-email {
    display: none;
  }
  .composer__header {
    padding: 12px 16px;
  }
}

/* ===== CSV 导入对话框样式 ===== */

/* 文件上传拖拽区域 */
.csv-upload-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 140px;
  border: 2px dashed var(--ifm-border-color);
  border-radius: 12px;
  cursor: pointer;
  transition: border-color 0.2s, background-color 0.2s;
  background: var(--ifm-background-surface-color);
}

.csv-upload-area:hover {
  border-color: var(--ifm-color-primary);
  background: var(--ifm-background-color);
}

.csv-upload-area--dragover {
  border-color: var(--ifm-color-primary);
  background: rgba(var(--ifm-color-primary-rgb, 0, 113, 227), 0.06);
}

.csv-upload-area__icon {
  color: var(--ifm-color-content-muted);
  margin-bottom: 8px;
}

.csv-upload-area__text {
  color: var(--ifm-color-content-secondary);
  font-size: 14px;
  margin: 0;
}

.csv-upload-area__hint {
  color: var(--ifm-color-content-muted);
  font-size: 12px;
  margin: 4px 0 0;
}

/* CSV 工具栏：目标字段选择 + 全选/统计 */
.csv-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 8px 0;
}

.csv-toolbar__left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.csv-toolbar__label {
  font-size: 13px;
  font-weight: 600;
  color: var(--ifm-color-content-secondary);
}

.csv-toolbar__right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.csv-toolbar__count {
  font-size: 13px;
  color: var(--ifm-color-content-secondary);
}

/* CSV 预览表格 */
.csv-preview-table {
  margin-top: 12px;
}

/* CSV 空状态 */
.csv-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 0;
  color: var(--ifm-color-content-secondary);
}

.csv-empty p {
  margin-top: 8px;
  font-size: 14px;
}
</style>
