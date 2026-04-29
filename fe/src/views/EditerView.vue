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

        <div class="composer__editor-wrapper">
          <Toolbar class="composer__editor-toolbar" :editor="editorRef" :defaultConfig="toolbarConfig" :mode="mode"/>
          <Editor class="composer__editor-content" v-model="valueHtml" :defaultConfig="editorConfig" :mode="mode" @onCreated="handleCreated"/>
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
  </div>
</template>

<script setup lang="ts">
import '@wangeditor/editor/dist/css/style.css'
import {ElMessage} from 'element-plus'
import {onBeforeUnmount, reactive, ref, shallowRef} from 'vue'
import {Close, Paperclip, Position, ArrowDown, Document, ArrowLeft} from '@element-plus/icons-vue';
import lang from '../i18n/i18n';
import {Editor, Toolbar} from '@wangeditor/editor-for-vue'
import {i18nChangeLanguage} from '@wangeditor/editor'
import {useRouter} from 'vue-router';
import {emailService} from "@/services/emailService";
import useGroupStore from '../stores/group'
import {useGlobalStatusStore} from "@/stores/useGlobalStatusStore";
import {createStringEmailValidator} from "@/utils/validators";

const router = useRouter();
const groupStore = useGroupStore()
const globalStatus = useGlobalStatusStore();
const showCcBcc = ref(false);
const isFullscreen = ref(false);

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
      })
    }else{
      ruleForm.sender = globalStatus.userInfos.account
      ruleForm.domains = globalStatus.userInfos.domains
      ruleForm.pickDomain = globalStatus.userInfos.domains[0]
      ruleForm.nickName = globalStatus.userInfos.name
    }
}
init()

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
</style>
