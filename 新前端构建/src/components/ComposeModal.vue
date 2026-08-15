<!--
  ComposeModal 写邮件弹窗组件

  全屏模态窗口，包含：
  - 发件人（Popover 选择器：前缀 + 域名 + 昵称）
  - 收件人、抄送、密送（标签式多选输入，对齐旧前端 el-select multiple）
  - 主题输入
  - contenteditable 富文本编辑区域
  - 附件上传
  - CSV 导入收件人（支持选择目标字段 To/Cc/Bcc）
  - 发送按钮

  修改日期: 20260610
  修改原因: 
    1. 功能对齐旧前端 EditerView.vue —— 发件人验证、回信参数、标签式收件人、CSV 导入
    2. 对齐标准 HTML 视觉 —— 遮罩模糊、弹窗宽度 600px、头部灰色背景、入场动画

  创建日期: 20260609
-->
<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="compose-modal">
      <!-- 头部 -->
      <div class="modal-header">
        <h2>{{ lang.compose }}</h2>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>

      <!-- 表单 -->
      <div class="modal-body">
      <!-- 发件人（Popover 选择器：前缀 + 域名 + 昵称，对齐旧前端 EditerView） -->
      <div class="form-row">
        <label class="row-label">{{ lang.sender }}</label>
        <div class="row-input-wrap">
          <div class="sender-selector" @click="showSenderPopover = !showSenderPopover">
            <span class="sender-name">{{ form.nickName || form.sender }}</span>
            <span class="sender-email"><{{ form.sender }}@{{ form.pickDomain }}></span>
            <span class="sender-arrow">▾</span>
          </div>
          <!-- 发件人编辑弹出框 -->
          <div v-if="showSenderPopover" class="sender-popover" @click.stop>
            <div class="popover-field">
              <label>{{ lang.sender_prefix }}</label>
              <input v-model="form.sender" class="popover-input" :placeholder="lang.sender_desc" :disabled="!isAdmin" />
            </div>
            <div class="popover-field">
              <label>{{ lang.sender_domain }}</label>
              <select v-model="form.pickDomain" class="popover-select">
                <option v-for="d in domains" :key="d" :value="d">{{ d }}</option>
              </select>
            </div>
            <div class="popover-field">
              <label>{{ lang.nick_name }}</label>
              <input v-model="form.nickName" class="popover-input" />
            </div>
          </div>
        </div>
      </div>

        <!-- 收件人（标签式多选，对齐旧前端 el-select multiple filterable allow-create） -->
        <div class="form-row">
          <label class="row-label">{{ lang.to_desc }}</label>
          <div class="row-input-wrap with-actions">
            <TagInput v-model="form.toList" :placeholder="lang.to_desc" />
            <button class="inline-btn" @click="showCc = !showCc">{{ lang.cc }}</button>
            <button class="inline-btn" @click="showBcc = !showBcc">{{ lang.bcc }}</button>
            <button class="inline-btn csv-btn" @click="showCsvImport = true">{{ lang.csv_import }}</button>
          </div>
        </div>

        <!-- 抄送（标签式多选） -->
        <div v-if="showCc" class="form-row">
          <label class="row-label">{{ lang.cc_desc }}</label>
          <div class="row-input-wrap">
            <TagInput v-model="form.ccList" :placeholder="lang.cc_desc" />
          </div>
        </div>

        <!-- 密送（标签式多选） -->
        <div v-if="showBcc" class="form-row">
          <label class="row-label">{{ lang.bcc_desc }}</label>
          <div class="row-input-wrap">
            <TagInput v-model="form.bccList" :placeholder="lang.bcc_desc" />
          </div>
        </div>

        <!-- 主题 -->
        <div class="form-row">
          <label class="row-label">{{ lang.title }}</label>
          <div class="row-input-wrap">
            <input v-model="form.subject" type="text" class="row-input" />
          </div>
        </div>

        <!-- 富文本编辑器 -->
        <div class="editor-wrap">
          <div ref="editorContainer" class="rich-editor" contenteditable="true" @input="onEditorInput"></div>
        </div>

        <!-- 附件 -->
        <div class="attachments-section">
          <div class="attachment-list">
            <div v-for="(file, i) in attachments" :key="i" class="attachment-chip">
              <span>{{ file.name }}</span>
              <button class="remove-btn" @click="removeAttachment(i)">×</button>
            </div>
          </div>
          <button class="add-att-btn" @click="pickAttachment">{{ lang.add_att }}</button>
          <input ref="fileInput" type="file" multiple class="hidden-input" @change="onFileSelected" />
        </div>
      </div>

      <!-- 底部 -->
      <div class="modal-footer">
        <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>
        <button class="send-btn" :disabled="sending" @click="onSend">
          {{ sending ? lang.wait_desc : lang.send }}
        </button>
      </div>

      <!-- CSV 导入弹窗（支持选择目标字段 To/Cc/Bcc，对齐旧前端 csvTargetField） -->
      <CsvImportModal
        v-if="showCsvImport"
        @close="showCsvImport = false"
        @import="onCsvImport"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import lang from '@/i18n'
import { sendEmail } from '@/services/emailService'
import type { EmailDetail } from '@/types/api'
import type { EmailSendParams } from '@/services/emailService'
import { isValidEmail } from '@/utils/validators'
import CsvImportModal from '@/components/CsvImportModal.vue'
import TagInput from '@/components/TagInput.vue'
import { useGlobalStore } from '@/stores/global'

/**
 * 组件属性
 *
 * replyTo: 回复的原始邮件（从 ContentPanel reply 事件传入）
 * replySender: 回信时自动选择的发件人前缀（对齐旧前端 route.query.reply_sender）
 * replyDomain: 回信时自动选择的发件人域名（对齐旧前端 route.query.reply_domain）
 */
const props = defineProps<{
  replyTo?: EmailDetail | null
  /** 回信发件人前缀（即收到该邮件的账号前缀），需验证域名在用户允许列表中才自动选择 */
  replySender?: string
  /** 回信发件人域名（即收到该邮件的账号域名），需验证域名在用户允许列表中才自动选择 */
  replyDomain?: string
}>()

const emit = defineEmits<{
  'close': []
  'sent': []
}>()

const globalStore = useGlobalStore()

/** 管理员权限（对齐旧前端 EditerView.vue :disabled="!(globalStatus.userInfos.is_admin)"） */
const isAdmin = computed(() => globalStore.isAdmin)

/* 发件人域名列表（从 globalStore.userInfo.domains 获取） */
const domains = ref<string[]>(globalStore.userInfo?.domains || [])

/* 表单数据（toList/ccList/bccList 为数组，对齐旧前端 el-select multiple 模式） */
const form = ref({
  sender: '',
  toList: [] as string[],
  ccList: [] as string[],
  bccList: [] as string[],
  subject: '',
  content: '',
  nickName: '',
  pickDomain: '',
})

/* 发件人选择器弹出框 */
const showSenderPopover = ref(false)

/* 点击外部关闭弹出框 */
function onClickOutside(e: MouseEvent) {
  const target = e.target as HTMLElement
  if (!target.closest('.sender-selector') && !target.closest('.sender-popover')) {
    showSenderPopover.value = false
  }
}
onMounted(() => {
  document.addEventListener('click', onClickOutside)
  /* 初始化发件人默认值（对齐旧前端 EditerView init 逻辑） */
  if (globalStore.userInfo) {
    form.value.sender = globalStore.userInfo.account || ''
    form.value.nickName = globalStore.userInfo.name || ''
    if (globalStore.userInfo.domains && globalStore.userInfo.domains.length > 0) {
      form.value.pickDomain = globalStore.userInfo.domains[0]
    }
  }

  /**
   * 回信预填：填充收件人、主题、发件人（对齐旧前端 EditerView fillReplyParams）
   * replySender/replyDomain 由 HomeView.onReply 从 EmailDetail 解析后传入。
   */
  if (props.replyTo) {
    /* 收件人 = 原邮件的发件人地址 */
    form.value.toList = [props.replyTo.from_address || '']

    /* 主题 = Re: + 原主题 */
    form.value.subject = 'Re: ' + (props.replyTo.subject || '')

    /* 邮件正文预填引用 */
    form.value.content = '<br><blockquote>' + (props.replyTo.text || '') + '</blockquote>'
    nextTick(() => {
      if (editorContainer.value && form.value.content) {
        editorContainer.value.innerHTML = form.value.content
      }
    })
  }

  /**
   * 自动选择发件人域名（对齐旧前端 fillReplyParams 中 ruleForm.domains.includes(replyDomain) 逻辑）
   * 只有当 replyDomain 在用户允许的域名列表中时，才自动切换发件人。
   */
  if (props.replySender && props.replyDomain && domains.value.includes(props.replyDomain)) {
    form.value.sender = props.replySender
    form.value.pickDomain = props.replyDomain
  }
})

const showCc = ref(false)
const showBcc = ref(false)
const attachments = ref<File[]>([])
const sending = ref(false)
const errorMsg = ref('')
const showCsvImport = ref(false)

const editorContainer = ref<HTMLDivElement | null>(null)
const fileInput = ref<HTMLInputElement | null>(null)

/** 编辑器内容同步 */
function onEditorInput() {
  if (editorContainer.value) {
    form.value.content = editorContainer.value.innerHTML
  }
}

/** 点击附件按钮 */
function pickAttachment() {
  fileInput.value?.click()
}

/** 文件选择回调 */
function onFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files) {
    attachments.value.push(...Array.from(target.files))
  }
  target.value = ''
}

/** 移除附件 */
function removeAttachment(index: number) {
  attachments.value.splice(index, 1)
}

/**
 * 将文件读取为 base64 字符串（对齐旧前端 FileReader.readAsDataURL）
 * @param file - 文件对象
 * @returns base64 编码的数据 URI
 */
function fileToBase64(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(reader.result as string)
    reader.onerror = () => reject(reader.error)
    reader.readAsDataURL(file)
  })
}

/**
 * CSV 导入回调（支持目标字段选择，对齐旧前端 confirmCsvImport 中 csvTargetField 逻辑）
 * @param emails - 导入的邮箱列表
 * @param targetField - 目标字段：'to' | 'cc' | 'bcc'
 */
function onCsvImport(emails: string[], targetField: 'to' | 'cc' | 'bcc' = 'to') {
  const validEmails = emails.filter((e: string) => isValidEmail(e))

  /* 根据目标字段选择追加到对应数组（去重，对齐旧前端 confirmCsvImport） */
  const targetArray = targetField === 'to'
    ? form.value.toList
    : targetField === 'cc'
      ? form.value.ccList
      : form.value.bccList

  const existingSet = new Set(targetArray.map((e: string) => e.toLowerCase()))
  for (const email of validEmails) {
    if (!existingSet.has(email.toLowerCase())) {
      targetArray.push(email)
      existingSet.add(email.toLowerCase())
    }
  }

  showCsvImport.value = false
}

/**
 * 发送邮件
 *
 * 验证逻辑对齐旧前端 EditerView send + validateSender：
 * 1. 发件人前缀不能为空
 * 2. 发件人前缀不能包含 @（对齐旧前端 only_prefix 错误提示）
 * 3. 至少有一个收件人（to/cc/bcc 至少一个非空，对齐旧前端 receivers.length === 0 判断）
 * 4. 所有邮箱格式合法
 * 5. 主题不能为空
 */
async function onSend() {
  errorMsg.value = ''

  /* 验证1：发件人前缀不能为空 */
  if (!form.value.sender || !form.value.sender.trim()) {
    errorMsg.value = lang.err_sender_must; return
  }

  /* 验证2：发件人前缀不能包含 @（对齐旧前端 validateSender 中 ruleForm.sender.includes("@")） */
  if (form.value.sender.includes('@')) {
    errorMsg.value = lang.only_prefix; return
  }

  /* 验证3：至少一个收件人（对齐旧前端 send 中 receivers.length === 0 判断） */
  const hasRecipient = form.value.toList.length > 0 || form.value.ccList.length > 0 || form.value.bccList.length > 0
  if (!hasRecipient) {
    errorMsg.value = lang.err_email_format; return
  }

  /* 验证4：所有邮箱格式合法 */
  const allEmails = [...form.value.toList, ...form.value.ccList, ...form.value.bccList]
  if (allEmails.some((e: string) => !isValidEmail(e.trim()))) {
    errorMsg.value = lang.err_email_format; return
  }

  /* 验证5：主题不能为空 */
  if (!form.value.subject) {
    errorMsg.value = lang.err_title_must; return
  }

  /* 同步编辑器内容 */
  onEditorInput()

  sending.value = true
  try {
    /* 对齐后端 sendRequest JSON 结构体 */
    const toAddrs = form.value.toList.map(e => ({ name: '', email: e.trim() })).filter(a => a.email)
    const ccAddrs = form.value.ccList.map(e => ({ name: '', email: e.trim() })).filter(a => a.email)
    const bccAddrs = form.value.bccList.map(e => ({ name: '', email: e.trim() })).filter(a => a.email)

    /* 编码附件为 base64（对齐旧前端 EditerView.vue 方式） */
    const encodedAttrs = await Promise.all(
      attachments.value.map(async (f) => {
        const data = await fileToBase64(f)
        /* 去除 data:mime;base64, 前缀（后端 send.go 自行解析） */
        return { name: f.name, data }
      })
    )

    /* 拼接完整发件人邮箱：前缀 + @ + 域名（对齐旧前端 compose send 逻辑） */
    const fromEmail = `${form.value.sender}@${form.value.pickDomain}`
    const params: EmailSendParams = {
      from: { name: form.value.nickName || form.value.sender, email: fromEmail },
      to: toAddrs,
      cc: ccAddrs,
      bcc: bccAddrs,
      subject: form.value.subject,
      text: form.value.content.replace(/<[^>]+>/g, ''),
      html: form.value.content,
      attrs: encodedAttrs,
    }

    const res: any = await sendEmail(params)
    /* axios 拦截器已解包，直接读 errorNo */
    if (res.errorNo === 0) {
      emit('sent')
    } else {
      errorMsg.value = res.errorMsg || lang.fail
    }
  } catch {
    errorMsg.value = lang.fail
  } finally {
    sending.value = false
  }
}
</script>

<style scoped>
/* 遮罩层（对齐标准 .modal-overlay: rgba(0,0,0,0.2), blur(2px), 平滑过渡） */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: rgba(0, 0, 0, 0);
  backdrop-filter: blur(0px);
  z-index: 100;
  display: flex;
  align-items: center;
  justify-content: center;
  visibility: hidden;
  opacity: 0;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 遮罩层显示状态（对齐标准 .modal-overlay.show） */
.modal-overlay.show,
.modal-overlay {
  visibility: visible;
  opacity: 1;
  background: rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(2px);
}

/* 弹窗（对齐标准 .modal-box: 480px, border-radius:12px, scale+translateY 入场动画） */
.compose-modal {
  width: 480px;
  max-height: 90vh;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
  transform: scale(1) translateY(0);
  opacity: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 头部（对齐标准 .m-head: padding 12px 16px） */
.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #e5e5e5;
}

.modal-header h2 {
  font-size: 12px;
  font-weight: 600;
}

.close-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: none;
  font-size: 14px;
  cursor: pointer;
  color: #8e8e93;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.2s;
}

.close-btn:hover {
  color: #1d1d1f;
}

/* 弹窗内容区（对齐标准 .m-body: padding 16px） */
.modal-body {
  padding: 16px;
  flex: 1;
  overflow-y: auto;
}

.form-row {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  align-items: flex-start;
}

.row-label {
  min-width: 80px;
  font-size: 13px;
  color: var(--text-secondary);
  padding-top: 8px;
  flex-shrink: 0;
}

.row-input-wrap {
  flex: 1;
  display: flex;
  gap: 4px;
  /* 修复发件人 popover 定位：作为 position:absolute 的定位基准 */
  position: relative;
}

.row-input-wrap.with-actions {
  gap: 4px;
  /* 修复 BUG-2：标签过多时按钮不被纵向拉伸，对齐到顶部 */
  align-items: flex-start;
}

/* 输入框（对齐标准 .m-input: 无 border, 底部细线分隔, font-size:12px） */
.row-input {
  flex: 1;
  padding: 8px 0;
  border: none;
  border-bottom: 1px solid #f5f5f7;
  border-radius: 0;
  font-size: 12px;
  outline: none;
  background: transparent;
  color: var(--text-primary);
  font-family: inherit;
  transition: border-color 0.2s ease;
}

.row-input:focus {
  border-color: #1d1d1f;
}

.inline-btn {
  padding: 6px 10px;
  border: 1px solid var(--border-color);
  background: transparent;
  border-radius: var(--radius-sm);
  font-size: 11px;
  cursor: pointer;
  white-space: nowrap;
  transition: all var(--transition);
}

.inline-btn:hover {
  background: var(--bg-hover);
}

.csv-btn {
  color: var(--accent-color);
  border-color: var(--accent-color);
}

/* 发件人选择器（对齐旧前端 EditerView） */
.sender-selector {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  cursor: pointer;
  background: var(--bg-secondary);
  transition: var(--transition);
}
.sender-selector:hover {
  background: var(--bg-hover);
}
.sender-name {
  font-weight: 600;
  font-size: 13px;
}
.sender-email {
  font-size: 12px;
  color: var(--text-secondary);
}
.sender-arrow {
  font-size: 10px;
  color: var(--text-placeholder);
}
.sender-popover {
  position: absolute;
  top: 100%;
  left: 0;
  margin-top: 4px;
  width: 320px;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  box-shadow: var(--shadow-lg);
  padding: 16px;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.popover-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.popover-field label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}
.popover-input, .popover-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
  outline: none;
}
.popover-input:focus, .popover-select:focus {
  border-color: var(--accent-color);
}

.editor-wrap {
  margin-bottom: 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  min-height: 200px;
}

.rich-editor {
  min-height: 200px;
  padding: 12px;
  font-size: 14px;
  outline: none;
  line-height: 1.6;
  color: var(--text-primary);
}

.attachments-section {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.attachment-list {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.attachment-chip {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
  font-size: 12px;
}

.remove-btn {
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 14px;
  color: var(--text-secondary);
  padding: 0 2px;
}

.add-att-btn {
  padding: 6px 12px;
  border: 1px dashed var(--border-color);
  background: transparent;
  border-radius: var(--radius-sm);
  font-size: 12px;
  cursor: pointer;
  transition: all var(--transition);
}

.add-att-btn:hover {
  background: var(--bg-hover);
}

.hidden-input {
  display: none;
}

/* 底部（对齐标准 .m-footer: padding 12px 16px） */
.modal-footer {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  border-top: 1px solid #e5e5e5;
}

.error-msg {
  font-size: 13px;
  color: var(--danger-color);
  flex: 1;
}

/* 发送按钮（对齐标准 .send-btn: padding 6px 16px, font-size:11px, border-radius:6px） */
.send-btn {
  padding: 6px 16px;
  border: none;
  border-radius: 6px;
  background: #1d1d1f;
  color: #fff;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.send-btn:hover:not(:disabled) {
  background: #333;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

/* 按钮点击缩放（对齐标准 .send-btn:active） */
.send-btn:active:not(:disabled) {
  transform: scale(0.95);
}

.send-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>