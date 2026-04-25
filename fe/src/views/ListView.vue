<template>
  <div class="list-view-container">
    <div class="list-header">
      <div class="header-title">
        <h2>{{ groupStore.name }}</h2>
      </div>
      <div class="header-actions">
        <el-button @click="del" class="action-btn" plain>
          <el-icon><Delete /></el-icon> {{ lang.del_btn }}
        </el-button>
        <el-button @click="markRead" class="action-btn" plain>
          <el-icon><View /></el-icon> {{ lang.read_btn }}
        </el-button>
        <el-dropdown class="move-dropdown">
          <el-button class="action-btn" plain>
            <el-icon><Folder /></el-icon> {{ lang.move_btn }}
            <el-icon class="el-icon--right"><EpArrowDownBold/></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="move(group.id,group.name)" v-for="group in groupList" :key="group.id">
                {{ group.name }}
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-button type="primary" class="compose-btn" @click="router.push('/editer')">
          <el-icon><EditPen /></el-icon> {{ lang.compose }}
        </el-button>
      </div>
    </div>

    <div class="list-content">
      <el-table 
        ref="taskTableDataRef" 
        :data="data" 
        :show-header="false" 
        class="modern-mail-table"
        @row-click="rowClick"
        :row-style="rowStyle"
      >
        <el-table-column type="selection" width="40"/>
        
        <!-- 状态指示列：未读圆点 + 危险/错误图标 -->
        <el-table-column width="44" class-name="status-col">
          <template #default="scope">
            <div class="status-indicator">
              <span class="unread-dot" v-if="!scope.row.is_read"></span>
              <el-tooltip effect="dark" :content="lang.dangerous" placement="top-start" v-if="scope.row.dangerous">
                <el-icon color="#ef4444"><Warning /></el-icon>
              </el-tooltip>
              <el-tooltip effect="dark" :content="scope.row.error" placement="top-start" v-if="scope.row.error !== ''">
                <el-icon color="#ef4444"><Warning /></el-icon>
              </el-tooltip>
            </div>
          </template>
        </el-table-column>

        <el-table-column min-width="250">
          <template #default="scope">
            <div class="mail-row-content" :class="{'is-unread': !scope.row.is_read}">
              <div class="mail-main-info">
                <div class="mail-sender">
                  {{ scope.row.sender.Name !== '' ? scope.row.sender.Name : scope.row.sender.EmailAddress }}
                </div>
                <div class="mail-subject">{{ scope.row.title }}</div>
                <div class="mail-snippet">{{ scope.row.desc }}</div>
              </div>
              <div class="mail-meta">
                <div class="mail-date">{{ formatShortDate(scope.row.datetime) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="pagination-wrapper" v-if="totalPage > 0">
      <el-pagination 
        background 
        layout="prev, pager, next" 
        :page-count="totalPage" 
        @current-change="pageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import {EpArrowDownBold} from "vue-icons-plus/ep";
import {Delete, View, Folder, EditPen, Warning} from "@element-plus/icons-vue";
import {useRouter} from 'vue-router'
import {ref, watch} from 'vue'
import useGroupStore from '../stores/group'
import lang from '../i18n/i18n';
import {http} from "@/utils/axios";
import {ElMessage, ElMessageBox} from "element-plus";
import type {EmailListItem, GroupListItem, ApiResponse, EmailListResponse} from "@/types/api";

const router = useRouter();
const groupStore = useGroupStore()
const groupList = ref<GroupListItem[]>([])
const taskTableDataRef = ref<any>(null)
let tag = groupStore.tag;

if (tag === "") {
  tag = '{"type":0,"status":-1}'
}

watch(groupStore, async (newV) => {
  tag = newV.tag;
  if (tag === "") {
    tag = '{"type":0,"status":-1}'
  }
  data.value = []
  updateList()
})

const data = ref<EmailListItem[]>([])
const totalPage = ref(0)

const updateList = function () {
  http.post("/api/email/list", {tag: tag, page_size: 15}).then((res: any) => {
    data.value = res.data.list || []
    totalPage.value = res.data.total_page || 0
  })
}

const updateGroupList = function () {
  http.post("/api/group/list").then((res: any) => {
    groupList.value = res.data || []
  })
}

updateList()
updateGroupList()

const rowClick = function (row: EmailListItem) {
  router.push("/detail/" + row.id)
}

const formatShortDate = (dateStr: string) => {
  if (!dateStr) return "";
  const d = new Date(dateStr);
  const now = new Date();
  if (d.toDateString() === now.toDateString()) {
    return d.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
  }
  return d.toLocaleDateString([], {month: 'short', day: 'numeric'});
}

const markRead = function () {
  let rows: EmailListItem[] = taskTableDataRef.value?.getSelectionRows()
  if (!rows || rows.length === 0) {
    ElMessage.warning('Select emails first');
    return;
  }
  let ids = rows.map((e: EmailListItem) => e.id);
  http.post("/api/email/read", {"ids": ids}).then((res: any) => {
    if (res.errorNo === 0) {
      updateList()
      ElMessage.success('Marked as read');
    } else {
      ElMessage.error(res.errorMsg)
    }
  })
}

const move = function (group_id: string, group_name: string) {
  let rows: EmailListItem[] = taskTableDataRef.value?.getSelectionRows()
  if (!rows || rows.length === 0) {
    ElMessage.warning('Select emails first');
    return;
  }
  let ids = rows.map((e: EmailListItem) => e.id);
  
  ElMessageBox.confirm(lang.move_email_confirm, 'Warning', {
    confirmButtonText: 'OK', cancelButtonText: 'Cancel', type: 'warning'
  }).then(() => {
    http.post("/api/email/move", {"group_id": group_id, "group_name": group_name, "ids": ids}).then((res: any) => {
      if (res.errorNo === 0) {
        updateList()
        ElMessage.success('Move completed')
      } else {
        ElMessage.error(res.errorMsg)
      }
    })
  }).catch(()=>{})
}

const del = function () {
  let rows: EmailListItem[] = taskTableDataRef.value?.getSelectionRows()
  if (!rows || rows.length === 0) {
    ElMessage.warning('Select emails first');
    return;
  }
  let ids = rows.map((e: EmailListItem) => e.id);
  let groupTag = JSON.parse(tag)

  ElMessageBox.confirm(lang.del_email_confirm, 'Warning', {
    confirmButtonText: 'OK', cancelButtonText: 'Cancel', type: 'warning'
  }).then(() => {
    http.post("/api/email/del", {"ids": ids, "forcedDel": groupTag.status === 3}).then((res: any) => {
      if (res.errorNo === 0) {
        updateList()
        ElMessage.success('Deleted successfully')
      } else {
        ElMessage.error(res.errorMsg)
      }
    })
  }).catch(()=>{})
}

const rowStyle = function () {
  return {'cursor': 'pointer'}
}

const pageChange = function (p: number) {
  http.post("/api/email/list", {tag: tag, page_size: 15, current_page: p}).then((res: any) => {
    data.value = res.data.list || []
  })
}
</script>

<!-- 样式改造: Docusaurus 风格, 移除 glassmorphism, 改为纯色卡片 | 日期: 20250425 -->
<style scoped>
/* 列表视图容器：移除额外 padding，由 App.vue #body 统一控制间距 */
.list-view-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--ifm-background-color);
  border: 1px solid var(--ifm-border-color);
  border-radius: var(--ifm-card-border-radius);
  box-shadow: var(--ifm-global-shadow-lw);
  padding: var(--ifm-spacing-md);
  overflow: hidden;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--ifm-spacing-md);
  flex-wrap: wrap;
  gap: var(--ifm-spacing-sm);
}

.header-title h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--ifm-color-content);
  margin: 0;
}

.header-actions {
  display: flex;
  gap: var(--ifm-spacing-sm);
  align-items: center;
  flex-wrap: wrap;
}

.action-btn {
  border-radius: var(--ifm-global-radius);
  border-color: var(--ifm-border-color);
  color: var(--ifm-color-content-secondary);
  background: var(--ifm-background-color);
  font-weight: 500;
}

.action-btn:hover {
  background-color: var(--pm-bg-hover);
  color: var(--ifm-color-content);
  border-color: var(--ifm-border-color);
}

.compose-btn {
  border-radius: var(--ifm-global-radius);
  font-weight: 600;
  margin-left: var(--ifm-spacing-sm);
  padding-inline: 16px;
}

.list-content {
  flex-grow: 1;
  border-radius: var(--ifm-global-radius);
  border: 1px solid var(--ifm-border-color);
  overflow: hidden;
}

.modern-mail-table {
  width: 100%;
}

.modern-mail-table :deep(tr) {
  transition: background-color var(--ifm-transition-fast);
}

.modern-mail-table :deep(tr:hover > td) {
  background-color: var(--pm-row-hover) !important;
}

.modern-mail-table :deep(td) {
  padding: 10px 0;
  border-bottom: 1px solid var(--ifm-color-emphasis-300);
}

/* 未读圆点：自定义 CSS 圆点替代 el-badge，避免无子元素时定位异常 */
.unread-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--ifm-color-primary);
  flex-shrink: 0;
}

.status-indicator {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 4px;
}

.mail-row-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.mail-main-info {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-md);
  flex-grow: 1;
  overflow: hidden;
}

/* 发件人：自适应宽度，窄屏可收缩 */
.mail-sender {
  min-width: 100px;
  max-width: 180px;
  font-weight: 500;
  color: var(--ifm-color-content);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  flex-shrink: 1;
}

/* 主题：自适应宽度，取消固定 max-width */
.mail-subject {
  font-weight: 500;
  color: var(--ifm-color-content);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  min-width: 120px;
  max-width: 40%;
  font-size: 14px;
  flex-shrink: 0;
}

/* 摘要：占据剩余空间 */
.mail-snippet {
  color: var(--ifm-color-content-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  flex-grow: 1;
  min-width: 60px;
}

.mail-meta {
  min-width: 80px;
  text-align: right;
  padding-right: var(--ifm-spacing-md);
}

.mail-date {
  font-size: 12px;
  color: var(--ifm-color-content-secondary);
}

.is-unread .mail-sender,
.is-unread .mail-subject {
  font-weight: 700;
  color: var(--ifm-color-content);
}

.is-unread .mail-date {
  color: var(--ifm-color-primary);
  font-weight: 600;
}

.pagination-wrapper {
  margin-top: var(--ifm-spacing-md);
  display: flex;
  justify-content: center;
  padding-bottom: var(--ifm-spacing-xs);
}

@media (max-width: 768px) {
  .list-header {
    flex-direction: column;
    align-items: flex-start;
  }
  .header-actions {
    width: 100%;
    overflow-x: auto;
    padding-bottom: var(--ifm-spacing-xs);
  }
  .mail-main-info {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--ifm-spacing-xs);
  }
  .mail-sender {
    width: 100%;
  }
  .mail-subject {
    max-width: 100%;
  }
  .mail-snippet {
    display: none;
  }
  .mail-row-content {
    align-items: flex-start;
  }
  .mail-meta {
    padding-top: 2px;
  }
}
</style>
