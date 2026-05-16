<template>
  <!-- 邮件详情视图（Docusaurus BEM 风格） -->
  <div class="mail-detail">
    <div class="mail-detail__header">
      <el-button @click="$router.back()" plain class="mail-detail__back-btn">
        <el-icon><Back /></el-icon>
      </el-button>
      <div class="mail-detail__actions">
        <!-- 回信按钮：仅收到的邮件（type === 0）才显示。新增日期: 20260516 -->
        <el-button v-if="detailData.type === 0" plain @click="handleReply" class="mail-detail__reply-btn">
          <el-icon><ChatRound /></el-icon>
          <span class="mail-detail__reply-text">{{ lang.reply_email }}</span>
        </el-button>
        <el-button plain @click="handleDelete">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="mail-detail__content">
      <h1 class="mail-detail__subject">{{ detailData.subject }}</h1>

      <div class="mail-detail__meta-card">
        <div class="mail-detail__meta-left">
          <div class="mail-detail__avatar">
            {{ getInitial(detailData.from_name || detailData.from_address) }}
          </div>
          <div class="mail-detail__meta-info">
            <div class="mail-detail__sender-line">
              <span class="mail-detail__sender-name">{{ detailData.from_name !== '' ? detailData.from_name : detailData.from_address }}</span>
              <span class="mail-detail__sender-email" v-if="detailData.from_name !== ''"><{{ detailData.from_address }}></span>
            </div>
            <div class="mail-detail__receivers-line">
              <span class="mail-detail__meta-label">{{ lang.to }}:</span>
              <span v-for="(to, index) in tos" :key="index" class="mail-detail__receiver-chip">
                {{ to.Name !== '' ? to.Name : to.EmailAddress }}<span v-if="index < tos.length - 1">, </span>
              </span>
              <span v-if="showCC" class="mail-detail__cc-section">
                <span class="mail-detail__meta-label">{{ lang.cc }}:</span>
                <span v-for="(item, index) in ccs" :key="'cc'+index" class="mail-detail__receiver-chip">
                  {{ item.Name !== '' ? item.Name : item.EmailAddress }}<span v-if="index < ccs.length - 1">, </span>
                </span>
              </span>
            </div>
          </div>
        </div>
        <div class="mail-detail__meta-right">
          <div class="mail-detail__date">{{ formatDetailDate(detailData.send_date) }}</div>
        </div>
      </div>

      <el-divider class="mail-detail__divider"/>

      <div class="mail-detail__body">
        <div class="mail-detail__body-text" v-if="detailData.html === ''">
          {{ detailData.text }}
        </div>
        <div class="mail-detail__body-html" v-else v-html="detailData.html"></div>
      </div>

      <div v-if="detailData.attachments && detailData.attachments.length > 0" class="mail-detail__attachments">
        <el-divider class="mail-detail__divider"/>
        <div class="mail-detail__attachments-title">{{ lang.attachment }} ({{ detailData.attachments.length }})</div>
        <div class="mail-detail__attachments-list">
          <a class="mail-detail__attachment-card" v-for="item in detailData.attachments" :key="item.Index" :href="'/attachments/download/' + detailData.id + '/' + item.Index">
            <div class="mail-detail__att-icon"><el-icon><Document/></el-icon></div>
            <div class="mail-detail__att-name">{{ item.Filename }}</div>
            <div class="mail-detail__att-download"><el-icon><Download/></el-icon></div>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {Document, Back, Delete, Download, ChatRound} from '@element-plus/icons-vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import lang from '../i18n/i18n';
import {emailService} from "@/services/emailService";
import useGroupStore from '../stores/group';
import {formatDetailDate} from "@/utils/dateFormat";
import {isTrashGroup} from "@/utils/groupTag";
import type {EmailDetail, EmailContact, EmailAttachment} from "@/types/api";

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()

const detailData = ref<Partial<EmailDetail>>({
  attachments: [] as EmailAttachment[]
})

const tos = ref<EmailContact[]>([])
const ccs = ref<EmailContact[]>([])
const showCC = ref(false)

const idParam = Array.isArray(route.params.id) ? route.params.id[0] : route.params.id;
/** 通过 emailService 获取邮件详情 */
emailService.getEmailDetail(parseInt(idParam)).then((res: any) => {
  detailData.value = res.data || {}
  detailData.value.attachments = res.data.attachments || [];
  
  if (res.data.to && res.data.to !== "") {
    try { tos.value = JSON.parse(res.data.to) } catch(e) { /* ignore parse error */ }
  }
  if (res.data.cc && res.data.cc !== "") {
    try { ccs.value = JSON.parse(res.data.cc) } catch(e) { /* ignore parse error */ }
  }
  showCC.value = ccs.value && ccs.value.length > 0
})

/** 获取名字首字母作为头像占位 */
const getInitial = (name: string) => {
  if (!name) return '?';
  return name.charAt(0).toUpperCase();
}

/**
 * 回信：通过 SPA 路由跳转到发信页，自动填写发件人（原收件人）与收件人（原发件人）。
 * 修改日期: 20260516 v1.2 — 使用 window.location.href 整页加载规避 WangEditor 空白
 * 修改日期: 20260516 v1.4 — 恢复为 router.push SPA 路由跳转，根因已修复（EditerView 中 fillReplyParams 声明顺序问题）
 */
const handleReply = () => {
  /** 收件人 = 原邮件的发件人地址 */
  const replyTo = detailData.value.from_address || ''
  /** 主题 = Re: + 原主题 */
  const replySubject = detailData.value.subject
    ? (detailData.value.subject.startsWith('Re: ') ? detailData.value.subject : 'Re: ' + detailData.value.subject)
    : ''

  /** 发件人 = 原邮件的第一个收件人地址（即收到该邮件的账号） */
  let replySender = ''
  let replyDomain = ''
  if (tos.value && tos.value.length > 0) {
    const firstToEmail = tos.value[0].EmailAddress || ''
    const parts = firstToEmail.split('@')
    if (parts.length === 2) {
      replySender = parts[0]
      replyDomain = parts[1]
    }
  }

  router.push({
    name: 'editer',
    query: {
      reply_to: replyTo,
      reply_subject: replySubject,
      reply_sender: replySender,
      reply_domain: replyDomain,
    }
  })
}

const handleDelete = () => {
  // 修改日期: 20260516 — 使用 ue_id（user_email 表主键）精确匹配记录，与 ListView 删除逻辑保持一致
  const ueId = detailData.value.ue_id;
  if (!ueId) return;

  /** 通过 isTrashGroup 判断当前分组是否为垃圾箱（status === 3） */
  const forcedDel = isTrashGroup(groupStore.tag);

  ElMessageBox.confirm(lang.del_email_confirm, 'Warning', {
    confirmButtonText: 'OK',
    cancelButtonText: 'Cancel',
    type: 'warning',
  }).then(() => {
    /** 通过 emailService 删除邮件，使用 ue_ids 精确匹配 user_email 记录 */
    emailService.deleteEmails({ids: [], forcedDel, ue_ids: [ueId]}).then((res: any) => {
      if (res.errorNo === 0) {
        ElMessage.success('Deleted successfully');
        router.push({name: 'list'});
      } else {
        ElMessage.error(res.errorMsg);
      }
    });
  }).catch(() => {});
}
</script>

<!-- 样式: Docusaurus BEM 风格 | 重构日期: 20260429 -->
<style scoped>
/* 邮件详情容器 */
.mail-detail {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--ifm-background-color);
  border-radius: var(--ifm-card-border-radius);
  box-shadow: var(--ifm-global-shadow-lw);
  border: 1px solid var(--ifm-border-color);
  overflow: hidden;
}

.mail-detail__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--ifm-spacing-md) var(--ifm-spacing-lg);
  border-bottom: 1px solid var(--ifm-border-color);
}

.mail-detail__back-btn {
  border-radius: var(--ifm-global-radius);
  font-size: 16px;
  padding: 8px 10px;
}

.mail-detail__actions .el-button {
  border-radius: var(--ifm-global-radius);
  font-size: 16px;
}

.mail-detail__content {
  flex-grow: 1;
  overflow-y: auto;
  padding: var(--ifm-spacing-lg) var(--ifm-spacing-xl);
}

.mail-detail__subject {
  font-size: 28px;
  font-weight: 700;
  color: var(--ifm-color-content);
  margin: 0 0 var(--ifm-spacing-lg) 0;
  line-height: 1.3;
}

.mail-detail__meta-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--ifm-spacing-md);
  padding: var(--ifm-spacing-md);
  border-radius: var(--ifm-global-radius);
  border: 1px solid var(--ifm-border-color);
  background: var(--ifm-background-surface-color);
}

.mail-detail__meta-left {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-md);
}

.mail-detail__avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--ifm-color-primary);
  color: white;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 18px;
  font-weight: 600;
  flex-shrink: 0;
}

.mail-detail__sender-line {
  margin-bottom: 4px;
}

.mail-detail__sender-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--ifm-color-content);
  margin-right: var(--ifm-spacing-sm);
}

.mail-detail__sender-email {
  color: var(--ifm-color-content-secondary);
  font-size: 14px;
}

.mail-detail__receivers-line {
  font-size: 13px;
  color: var(--ifm-color-content-secondary);
}

.mail-detail__meta-label {
  color: var(--ifm-color-content-muted);
  margin-right: 4px;
}

.mail-detail__cc-section {
  margin-left: var(--ifm-spacing-sm);
}

.mail-detail__meta-right {
  color: var(--ifm-color-content-secondary);
  font-size: 14px;
}

.mail-detail__divider {
  margin: var(--ifm-spacing-sm) 0 var(--ifm-spacing-lg) 0;
}

.mail-detail__body {
  font-size: 16px;
  line-height: 1.7;
  color: var(--ifm-color-content);
  min-height: 200px;
}

.mail-detail__body-html :deep(img) {
  max-width: 100%;
  height: auto;
}

.mail-detail__body-text {
  white-space: pre-wrap;
}

.mail-detail__attachments {
  margin-top: var(--ifm-spacing-xl);
}

.mail-detail__attachments-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: var(--ifm-spacing-md);
  color: var(--ifm-color-content);
}

.mail-detail__attachments-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ifm-spacing-md);
}

.mail-detail__attachment-card {
  display: flex;
  align-items: center;
  padding: var(--ifm-spacing-sm) var(--ifm-spacing-md);
  border: 1px solid var(--ifm-border-color);
  border-radius: var(--ifm-global-radius);
  background-color: var(--ifm-background-surface-color);
  transition: border-color var(--ifm-transition-fast), background-color var(--ifm-transition-fast);
  min-width: 200px;
  max-width: 300px;
}

.mail-detail__attachment-card:hover {
  border-color: var(--ifm-color-primary);
  background-color: var(--ifm-background-hover-color);
  text-decoration: none;
}

.mail-detail__att-icon {
  font-size: 22px;
  color: var(--ifm-color-content-secondary);
  margin-right: var(--ifm-spacing-sm);
}

.mail-detail__att-name {
  flex-grow: 1;
  font-size: 14px;
  color: var(--ifm-color-content);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: var(--ifm-spacing-sm);
}

.mail-detail__att-download {
  color: var(--ifm-color-primary);
  font-size: 18px;
}

@media (max-width: 768px) {
  .mail-detail__header {
    padding: var(--ifm-spacing-sm) var(--ifm-spacing-md);
  }
  .mail-detail__content {
    padding: var(--ifm-spacing-md);
  }
  .mail-detail__subject {
    font-size: 22px;
  }
  .mail-detail__meta-card {
    flex-direction: column;
    gap: var(--ifm-spacing-sm);
  }
  .mail-detail__meta-right {
    padding-left: 56px;
  }
  .mail-detail__avatar {
    width: 36px;
    height: 36px;
    font-size: 15px;
  }
}
</style>
