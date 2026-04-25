<template>
  <div class="mail-detail-container">
    <div class="mail-detail-header">
      <el-button @click="$router.back()" plain class="back-btn">
        <el-icon><Back /></el-icon>
      </el-button>
      <div class="action-buttons">
        <el-button plain @click="handleDelete">
          <el-icon><Delete /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="mail-detail-content">
      <h1 class="mail-subject">{{ detailData.subject }}</h1>
      
      <div class="mail-meta-card">
        <div class="meta-left">
          <div class="avatar-placeholder">
            {{ getInitial(detailData.from_name || detailData.from_address) }}
          </div>
          <div class="meta-info">
            <div class="sender-line">
              <span class="sender-name">{{ detailData.from_name !== '' ? detailData.from_name : detailData.from_address }}</span>
              <span class="sender-email" v-if="detailData.from_name !== ''">&lt;{{ detailData.from_address }}&gt;</span>
            </div>
            <div class="receivers-line">
              <span class="meta-label">{{ lang.to }}:</span>
              <span v-for="(to, index) in tos" :key="index" class="receiver-chip">
                {{ to.Name !== '' ? to.Name : to.EmailAddress }}<span v-if="index < tos.length - 1">, </span>
              </span>
              <span v-if="showCC" class="cc-section">
                <span class="meta-label">{{ lang.cc }}:</span>
                <span v-for="(item, index) in ccs" :key="'cc'+index" class="receiver-chip">
                  {{ item.Name !== '' ? item.Name : item.EmailAddress }}<span v-if="index < ccs.length - 1">, </span>
                </span>
              </span>
            </div>
          </div>
        </div>
        <div class="meta-right">
          <div class="mail-date">{{ formatDetailDate(detailData.send_date) }}</div>
        </div>
      </div>

      <el-divider class="custom-divider"/>

      <div class="mail-body">
        <div class="body-text" v-if="detailData.html === ''">
          {{ detailData.text }}
        </div>
        <div class="body-html" v-else v-html="detailData.html"></div>
      </div>

      <div v-if="detailData.attachments && detailData.attachments.length > 0" class="attachments-section">
        <el-divider class="custom-divider"/>
        <div class="attachments-title">{{ lang.attachment }} ({{ detailData.attachments.length }})</div>
        <div class="attachments-list">
          <a class="attachment-card" v-for="item in detailData.attachments" :key="item.Index" :href="'/attachments/download/' + detailData.id + '/' + item.Index">
            <div class="att-icon"><el-icon><Document/></el-icon></div>
            <div class="att-name">{{ item.Filename }}</div>
            <div class="att-download"><el-icon><Download/></el-icon></div>
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import {ref} from 'vue'
import {useRoute, useRouter} from 'vue-router'
import {Document, Back, Delete, Download} from '@element-plus/icons-vue';
import {ElMessage, ElMessageBox} from 'element-plus';
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import useGroupStore from '../stores/group';

const route = useRoute()
const router = useRouter()
const groupStore = useGroupStore()

interface EmailContact { Name: string; EmailAddress: string }
interface Attachment { Index: number; Filename: string }

const detailData = ref<Record<string, any>>({
  attachments: []
})

const tos = ref<EmailContact[]>([])
const ccs = ref<EmailContact[]>([])
const showCC = ref(false)

const idParam = Array.isArray(route.params.id) ? route.params.id[0] : route.params.id;
http.post("/api/email/detail", {id: parseInt(idParam)}).then((res: any) => {
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

const getInitial = (name: string) => {
  if (!name) return '?';
  return name.charAt(0).toUpperCase();
}

const formatDetailDate = (dateStr: string) => {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  return d.toLocaleString([], {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute:'2-digit'
  });
}

const handleDelete = () => {
  const id = detailData.value.id || parseInt(Array.isArray(route.params.id) ? route.params.id[0] : route.params.id);
  if (!id) return;

  let tag: string = groupStore.tag;
  if (!tag) {
    tag = '{"type":0,"status":-1}';
  }

  let forcedDel = false;
  try {
    const groupTag = JSON.parse(tag);
    forcedDel = groupTag.status === 3;
  } catch (e) {
    forcedDel = false;
  }

  ElMessageBox.confirm(lang.del_email_confirm, 'Warning', {
    confirmButtonText: 'OK',
    cancelButtonText: 'Cancel',
    type: 'warning',
  }).then(() => {
    http.post("/api/email/del", {ids: [id], forcedDel}).then((res: any) => {
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

<!-- 样式改造: Docusaurus 风格 | 日期: 20250425 -->
<style scoped>
.mail-detail-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--ifm-background-color);
  border-radius: var(--ifm-card-border-radius);
  box-shadow: var(--ifm-global-shadow-lw);
  border: 1px solid var(--ifm-border-color);
  overflow: hidden;
}

.mail-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--ifm-spacing-md) var(--ifm-spacing-lg);
  border-bottom: 1px solid var(--ifm-border-color);
}

.back-btn {
  border-radius: var(--ifm-global-radius);
  font-size: 16px;
  padding: 8px 10px;
}

.action-buttons .el-button {
  border-radius: var(--ifm-global-radius);
  font-size: 16px;
}

.mail-detail-content {
  flex-grow: 1;
  overflow-y: auto;
  padding: var(--ifm-spacing-lg) var(--ifm-spacing-xl);
}

.mail-subject {
  font-size: 28px;
  font-weight: 700;
  color: var(--ifm-color-content);
  margin: 0 0 var(--ifm-spacing-lg) 0;
  line-height: 1.3;
}

.mail-meta-card {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: var(--ifm-spacing-md);
  padding: var(--ifm-spacing-md);
  border-radius: var(--ifm-global-radius);
  border: 1px solid var(--ifm-border-color);
  background: var(--ifm-background-surface-color);
}

.meta-left {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-md);
}

.avatar-placeholder {
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

.sender-line {
  margin-bottom: 4px;
}

.sender-name {
  font-weight: 600;
  font-size: 15px;
  color: var(--ifm-color-content);
  margin-right: var(--ifm-spacing-sm);
}

.sender-email {
  color: var(--ifm-color-content-secondary);
  font-size: 14px;
}

.receivers-line {
  font-size: 13px;
  color: var(--ifm-color-content-secondary);
}

.meta-label {
  color: var(--ifm-color-content-muted);
  margin-right: 4px;
}

.cc-section {
  margin-left: var(--ifm-spacing-sm);
}

.meta-right {
  color: var(--ifm-color-content-secondary);
  font-size: 14px;
}

.custom-divider {
  margin: var(--ifm-spacing-sm) 0 var(--ifm-spacing-lg) 0;
}

.mail-body {
  font-size: 16px;
  line-height: 1.7;
  color: var(--ifm-color-content);
  min-height: 200px;
}

.body-html :deep(img) {
  max-width: 100%;
  height: auto;
}

.body-text {
  white-space: pre-wrap;
}

.attachments-section {
  margin-top: var(--ifm-spacing-xl);
}

.attachments-title {
  font-size: 15px;
  font-weight: 600;
  margin-bottom: var(--ifm-spacing-md);
  color: var(--ifm-color-content);
}

.attachments-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--ifm-spacing-md);
}

.attachment-card {
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

.attachment-card:hover {
  border-color: var(--ifm-color-primary);
  background-color: var(--pm-bg-hover);
  text-decoration: none;
}

.att-icon {
  font-size: 22px;
  color: var(--ifm-color-content-secondary);
  margin-right: var(--ifm-spacing-sm);
}

.att-name {
  flex-grow: 1;
  font-size: 14px;
  color: var(--ifm-color-content);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-right: var(--ifm-spacing-sm);
}

.att-download {
  color: var(--ifm-color-primary);
  font-size: 18px;
}

@media (max-width: 768px) {
  .mail-detail-header {
    padding: var(--ifm-spacing-sm) var(--ifm-spacing-md);
  }
  .mail-detail-content {
    padding: var(--ifm-spacing-md);
  }
  .mail-subject {
    font-size: 22px;
  }
  .mail-meta-card {
    flex-direction: column;
    gap: var(--ifm-spacing-sm);
  }
  .meta-right {
    padding-left: 56px;
  }
  .avatar-placeholder {
    width: 36px;
    height: 36px;
    font-size: 15px;
  }
}
</style>
