<!--
  ContentPanel 内容面板组件

  右侧自适应宽度的邮件详情面板，提供：
  - 邮件头信息（卡片式布局：头像 + 发件人/收件人/时间，对齐旧前端 EmailDetailView）
  - 安全警告提示
  - HTML 邮件正文（iframe sandbox 隔离渲染）
  - 附件列表
  - 回复按钮

  创建日期: 20260609
  修改日期: 20260610
  修改原因: 
    1. 邮件头部改为卡片式布局（圆形头像 + 发件人名称粗体 + 邮箱灰色 + 收件人 chip + 时间右侧）
    2. CC 为 null 时隐藏 CC 行
    3. 对齐标准 HTML 视觉 —— 独立工具栏、主题 28px、内容区 max-width 居中
-->
<template>
  <div class="content-panel">
    <!-- 空状态 -->
    <div v-if="!email" class="empty-state">
      <div class="empty-icon">📭</div>
      <p>{{ lang.select_email_hint }}</p>
    </div>

    <!-- 邮件详情 -->
    <template v-else>
      <!-- 独立工具栏（对齐标准 .content-toolbar：左日期 + 右操作按钮） -->
      <div class="content-toolbar">
        <div class="toolbar-date">{{ formatDetailDate(email.send_date) }}</div>
        <div class="toolbar-actions">
          <button v-if="email.type !== undefined && email.type === 0" class="tool-btn" @click="$emit('reply', email)">↩ {{ lang.reply_email }}</button>
          <button class="tool-btn danger" @click="$emit('delete', email)">🗑 {{ lang.del_btn }}</button>
        </div>
      </div>

      <!-- 邮件内容滚动区（对齐标准 .content-scroll） -->
      <div class="content-scroll">
        <div class="email-container">
          <!-- 邮件头部 -->
          <div class="email-header">
            <h1 class="email-subject">{{ email.subject }}</h1>

            <!-- 安全警告 -->
            <div v-if="email.dangerous" class="danger-warning">
          ⚠ {{ lang.dangerous }}
        </div>

        <!-- 
          邮件元信息卡片（对齐旧前端 EmailDetailView meta-card 布局）
          左侧：圆形头像 + 发件人/收件人
          右侧：时间
          修改日期: 20260610
        -->
        <div class="meta-card">
          <div class="meta-card-left">
            <!-- 圆形头像（发件人名称首字母） -->
            <div class="meta-avatar">{{ getInitial(email.from_name || email.from_address) }}</div>
            <div class="meta-card-info">
              <!-- 发件人行：名称（粗体） + 邮箱地址（灰色） -->
              <div class="meta-sender-line">
                <span class="meta-sender-name">{{ email.from_name || email.from_address }}</span>
                <span class="meta-sender-email" v-if="email.from_name"><{{ email.from_address }}></span>
              </div>
              <!-- 收件人行 -->
              <div class="meta-receivers-line">
                <span class="meta-label">{{ lang.to }}:</span>
                <span v-for="(c, i) in parsedTo" :key="'to'+i" class="meta-receiver-chip">
                  {{ c.Name || c.EmailAddress }}<span v-if="i < parsedTo.length - 1">, </span>
                </span>
              </div>
              <!-- 抄送行（仅 CC 存在且非空时显示） -->
              <div v-if="parsedCc.length > 0" class="meta-receivers-line">
                <span class="meta-label">{{ lang.cc }}:</span>
                <span v-for="(c, i) in parsedCc" :key="'cc'+i" class="meta-receiver-chip">
                  {{ c.Name || c.EmailAddress }}<span v-if="i < parsedCc.length - 1">, </span>
                </span>
              </div>
            </div>
          </div>
          <div class="meta-card-right">
            <span class="meta-date">{{ formatDetailDate(email.send_date) }}</span>
          </div>
        </div>

          </div>

          <!-- 邮件正文（iframe 沙箱隔离） -->
          <div class="email-body">
        <iframe
          ref="bodyFrame"
          class="body-iframe"
          sandbox="allow-same-origin"
          :srcdoc="email.html || email.text"
        />
      </div>

          <!-- 附件列表 -->
          <div v-if="email.attachments && email.attachments.length > 0" class="attachments">
            <h3 class="attachments-title">{{ lang.attachment }}</h3>
            <div class="attachment-list">
              <a
                v-for="att in email.attachments"
                :key="att.Index"
                class="attachment-item"
                :href="getAttachmentUrl(email.id, att.Index)"
                target="_blank"
              >
                📎 {{ att.Filename }}
              </a>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import lang from '@/i18n'
import { ref, watch, onBeforeUnmount, computed } from 'vue'
import type { EmailDetail, EmailContact } from '@/types/api'
import { formatDetailDate } from '@/utils/dateFormat'
import { getAttachmentUrl } from '@/services/emailService'

const props = defineProps<{
  email: EmailDetail | null
}>()

defineEmits<{
  'reply': [email: EmailDetail]
  'delete': [email: EmailDetail]
}>()

const bodyFrame = ref<HTMLIFrameElement | null>(null)

/**
 * iframe 高度自适应定时器
 *
 * 每 300ms 检测 iframe 内容高度变化并动态调整 iframe 高度。
 * 使用定时器而非 MutationObserver，因为 sandbox 模式下
 * 无法直接访问 iframe 内部 DOM 来监听变化。
 *
 * 修改日期: 20260610
 * 修改原因: 对齐旧前端 iframe 高度自适应行为，解决内容超出时滚动条异常
 */
let heightTimer: ReturnType<typeof setInterval> | null = null

/** 启动高度自适应检测 */
function startHeightWatch() {
  stopHeightWatch()
  heightTimer = setInterval(() => {
    if (!bodyFrame.value) return
    try {
      /* sandbox="allow-same-origin" 允许读取 contentDocument */
      const doc = bodyFrame.value.contentDocument || bodyFrame.value.contentWindow?.document
      if (doc && doc.body) {
        const contentHeight = doc.body.scrollHeight || doc.documentElement.scrollHeight
        if (contentHeight > 0) {
          bodyFrame.value.style.height = contentHeight + 'px'
        }
      }
    } catch {
      /* 跨域限制时忽略，保持默认高度 */
    }
  }, 300)
}

/** 停止高度自适应检测 */
function stopHeightWatch() {
  if (heightTimer) {
    clearInterval(heightTimer)
    heightTimer = null
  }
}

/* 监听邮件变化，启动/停止高度检测 */
watch(() => props.email, (newEmail) => {
  if (newEmail) {
    startHeightWatch()
  } else {
    stopHeightWatch()
  }
})

onBeforeUnmount(() => {
  stopHeightWatch()
})

/**
 * 解析联系人 JSON 字符串为对象数组（对齐旧前端 EmailDetailView）
 * 后端返回的 to/cc 字段可能是 JSON 字符串数组，也可能是普通文本
 *
 * @param raw - 后端返回的联系人字段原始值
 * @returns 解析后的 EmailContact 数组，解析失败返回空数组
 * 修改日期: 20260610
 * 修改原因: 卡片式布局需要逐个渲染收件人 chip，改为返回数组
 */
function parseContacts(raw: string | undefined | null): EmailContact[] {
  if (!raw) return []
  try {
    const parsed = JSON.parse(raw)
    if (Array.isArray(parsed)) return parsed as EmailContact[]
    /* 非数组 JSON（如纯字符串），包装为单元素数组 */
    return [{ Name: '', EmailAddress: raw }]
  } catch {
    /* JSON 解析失败，按原始文本处理 */
    return [{ Name: '', EmailAddress: raw }]
  }
}

/** 解析后的收件人列表（计算属性，缓存避免重复解析） */
const parsedTo = computed(() => parseContacts(props.email?.to))

/** 解析后的抄送列表（计算属性，CC 为 null 时返回空数组，模板中 v-if 隐藏） */
const parsedCc = computed(() => parseContacts(props.email?.cc))

/**
 * 获取发件人头像首字母（对齐旧前端 EmailDetailView 头像）
 *
 * @param name - 发件人名称或邮箱地址
 * @returns 大写的首字母（取第一个字符的大写）
 * 修改日期: 20260610
 * 修改原因: 卡片式布局需要圆形头像显示首字母
 */
function getInitial(name: string): string {
  if (!name) return '?'
  /* 取第一个字符的大写（支持中文、英文） */
  return name.charAt(0).toUpperCase()
}

/**
 * 格式化联系人列表为字符串（保留兼容，供其他地方调用）
 * 后端返回的 to/cc 字段可能是 JSON 字符串数组，也可能是普通文本
 */
function formatContacts(raw: string | undefined | null): string {
  if (!raw) return ''
  const contacts = parseContacts(raw)
  return contacts
    .map((c: EmailContact) => c.Name ? `${c.Name} <${c.EmailAddress}>` : c.EmailAddress)
    .join(', ')
}
</script>

<style scoped>
/* 内容面板（对齐标准 .content-panel：overflow:hidden 防止双重滚动） */
.content-panel {
  flex: 1;
  height: 100%;
  background: var(--bg-color);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: var(--text-placeholder);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

/* 独立工具栏（对齐标准 .content-head: padding 12px 24px） */
.content-toolbar {
  padding: 12px 24px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
  flex-shrink: 0;
}

/* 日期（对齐标准 .c-date: 11px, color:#8e8e93, transition:opacity 0.2s） */
.toolbar-date {
  font-size: 11px;
  color: #8e8e93;
  transition: opacity 0.2s ease;
}

.toolbar-actions {
  display: flex;
  gap: 6px;
}

/* 工具栏按钮（对齐标准 .act-btn: padding 4px 10px, font-size:10px） */
.tool-btn {
  padding: 4px 10px;
  background: transparent;
  border: 1px solid #e5e5e5;
  border-radius: 4px;
  cursor: pointer;
  font-size: 10px;
  color: #333;
  transition: all 0.2s ease;
}

.tool-btn:hover {
  background: #f5f5f7;
  border-color: #d1d1d6;
}

/* 按钮点击缩放（对齐标准 .act-btn:active） */
.tool-btn:active {
  transform: scale(0.95);
  background: #e5e5e5;
}

.tool-btn.danger {
  color: var(--danger-color);
}

/* 内容滚动区（对齐标准 .content-body: padding 24px） */
.content-scroll {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
}

/* 内容居中容器（对齐标准 .email-wrapper: max-width 680px + contentFadeIn 动画） */
.email-container {
  max-width: 680px;
  margin: 0 auto;
  animation: contentFadeIn 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 内容切换动画（对齐标准 @keyframes contentFadeIn） */
@keyframes contentFadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.email-header {
  margin-bottom: 32px;
}

/* 邮件主题（对齐标准 .e-subject: 20px, font-weight:700, letter-spacing:-0.5px） */
.email-subject {
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 20px;
  letter-spacing: -0.5px;
  line-height: 1.2;
}

.danger-warning {
  padding: 10px 14px;
  background: var(--danger-bg);
  color: var(--danger-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
  margin-bottom: 12px;
}

/* 邮件元信息卡片（对齐标准 .sender-box: margin-bottom 24px, padding-bottom 16px） */
.meta-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 24px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f5f5f7;
  gap: 12px;
}

.meta-card-left {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.meta-card-right {
  flex-shrink: 0;
  text-align: right;
}

/* 圆形头像（对齐标准 .s-avatar: 36×36, font-size:12px） */
.meta-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #e5e5e5;
  color: #666;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}

.meta-card-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

/* 发件人行（对齐标准 .s-info h4: 12px, font-weight:600） */
.meta-sender-line {
  display: flex;
  align-items: baseline;
  gap: 6px;
  font-size: 12px;
}

.meta-sender-name {
  font-weight: 600;
  color: var(--text-primary);
}

.meta-sender-email {
  font-size: 10px;
  color: #8e8e93;
}

/* 收件人/抄送行（对齐标准 .s-info span: 10px） */
.meta-receivers-line {
  display: flex;
  align-items: baseline;
  gap: 4px;
  font-size: 10px;
  flex-wrap: wrap;
}

.meta-label {
  color: var(--text-secondary);
  flex-shrink: 0;
}

.meta-receiver-chip {
  color: var(--text-primary);
  white-space: nowrap;
}

/* 时间（右侧，对齐标准 .s-info span: 10px 灰色） */
.meta-date {
  font-size: 10px;
  color: #8e8e93;
  white-space: nowrap;
}

/* 邮件正文（对齐标准 .e-content: font-size 13px, line-height 1.6） */
.email-body {
  padding: 0;
  font-size: 13px;
  line-height: 1.6;
  color: #333;
}

.body-iframe {
  width: 100%;
  height: 100%;
  min-height: 400px;
  border: none;
  display: block;
}

.attachments {
  padding: 16px 24px;
  border-top: 1px solid var(--border-color);
}

.attachments-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-secondary);
}

.attachment-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.attachment-item {
  padding: 6px 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-primary);
  transition: background var(--transition);
}

.attachment-item:hover {
  background: var(--bg-hover);
}
</style>