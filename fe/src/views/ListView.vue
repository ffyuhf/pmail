<template>
  <!-- 邮件列表视图（Docusaurus BEM 风格） -->
  <div class="mail-list">
    <div class="mail-list__header">
      <div class="mail-list__title">
        <h2>{{ groupStore.name }}</h2>
      </div>
      <div class="mail-list__actions">
        <!-- 全选复选框：一键选中/取消当前页所有邮件 -->
        <el-checkbox
          v-model="isAllSelected"
          :indeterminate="isIndeterminate"
          @change="toggleSelectAll"
          class="mail-list__select-all"
        >
          {{ lang.select_all }}
        </el-checkbox>
        <!-- 已选计数：实时显示当前选中邮件数量 -->
        <span v-if="selectedCount > 0" class="mail-list__selected-count">
          {{ selectedCount }} {{ lang.selected_count }}
        </span>
        <el-button @click="del" class="mail-list__action-btn" plain>
          <el-icon><Delete /></el-icon> {{ lang.del_btn }}
        </el-button>
        <el-button @click="markRead" class="mail-list__action-btn" plain>
          <el-icon><View /></el-icon> {{ lang.read_btn }}
        </el-button>
        <el-dropdown class="mail-list__move-dropdown">
          <el-button class="mail-list__action-btn" plain>
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
        <el-button type="primary" class="mail-list__compose-btn" @click="router.push('/editer')">
          <el-icon><EditPen /></el-icon> {{ lang.compose }}
        </el-button>
      </div>
    </div>

    <div class="mail-list__content">
      <el-table
        ref="taskTableDataRef"
        :data="data"
        :show-header="false"
        class="mail-list__table"
        @row-click="rowClick"
        :row-style="rowStyle"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="40"/>

        <!-- 状态指示列：未读圆点 + 危险/错误图标 -->
        <el-table-column width="44" class-name="mail-list__status-col">
          <template #default="scope">
            <div class="mail-list__status">
              <span class="mail-list__unread-dot" v-if="!scope.row.is_read"></span>
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
            <div class="mail-list__row" :class="{'mail-list__row--unread': !scope.row.is_read}">
              <div class="mail-list__row-main">
                <div class="mail-list__sender">
                  {{ scope.row.sender.Name !== '' ? scope.row.sender.Name : scope.row.sender.EmailAddress }}
                </div>
                <div class="mail-list__subject">{{ scope.row.title }}</div>
                <div class="mail-list__snippet">{{ scope.row.desc }}</div>
              </div>
              <div class="mail-list__meta">
                <div class="mail-list__date">{{ formatShortDate(scope.row.datetime) }}</div>
              </div>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="mail-list__pagination" v-if="totalPage > 0">
      <!-- 每页条数选择器：用户可自定义每页显示邮件数量 -->
      <div class="mail-list__page-size">
        <span class="mail-list__page-size-label">{{ lang.per_page }}</span>
        <el-select
          v-model="pageSize"
          class="mail-list__page-size-select"
          @change="handlePageSizeChange"
        >
          <el-option :value="15" label="15"/>
          <el-option :value="25" label="25"/>
          <el-option :value="50" label="50"/>
          <el-option :value="100" label="100"/>
        </el-select>
      </div>
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
import {ref, watch, computed} from 'vue'
import useGroupStore from '../stores/group'
import lang from '../i18n/i18n';
import {emailService} from "@/services/emailService";
import {groupService} from "@/services/groupService";
import {ElMessage, ElMessageBox} from "element-plus";
import type {EmailListItem, GroupListItem} from "@/types/api";
import {formatShortDate} from "@/utils/dateFormat";
import {normalizeTag, isTrashGroup} from "@/utils/groupTag";

const router = useRouter();
const groupStore = useGroupStore()
const groupList = ref<GroupListItem[]>([])
const taskTableDataRef = ref<any>(null)
let tag = normalizeTag(groupStore.tag);

watch(groupStore, async (newV) => {
  tag = normalizeTag(newV.tag);
  data.value = []
  updateList()
})

/** 监听搜索关键词变化，自动刷新列表 */
watch(() => groupStore.keyword, () => {
  data.value = []
  updateList()
})

const data = ref<EmailListItem[]>([])
const totalPage = ref(0)

/** 每页显示条数：用户可通过下拉选择器自定义（默认 15） */
const pageSize = ref(15)

/** 当前选中行列表（由 el-table 的 selection-change 事件维护） */
const selectedRows = ref<EmailListItem[]>([])

/** 已选邮件数量 */
const selectedCount = computed(() => selectedRows.value.length)

/** 全选复选框绑定值：当全部选中时为 true，全部未选时为 false */
const isAllSelected = computed(() => {
  return data.value.length > 0 && selectedRows.value.length === data.value.length
})

/** 半选状态：部分选中时复选框显示 indeterminate 样式 */
const isIndeterminate = computed(() => {
  return selectedRows.value.length > 0 && selectedRows.value.length < data.value.length
})

/** 刷新邮件列表：通过 emailService 获取数据，使用当前 pageSize */
const updateList = function () {
  emailService.getEmailList({tag: tag, page_size: pageSize.value, keyword: groupStore.keyword}).then((res: any) => {
    data.value = res.data.list || []
    totalPage.value = res.data.total_page || 0
  })
}

/** 刷新分组列表：通过 groupService 获取数据 */
const updateGroupList = function () {
  groupService.getGroupList().then((res: any) => {
    groupList.value = res.data || []
  })
}

updateList()
updateGroupList()

const rowClick = function (row: EmailListItem) {
  router.push("/detail/" + row.id)
}

/** el-table 选中项变化回调：同步维护 selectedRows 状态 */
const handleSelectionChange = (rows: EmailListItem[]) => {
  selectedRows.value = rows
}

/** 全选/取消全选：通过 el-table 的 toggleAllSelection 方法实现 */
const toggleSelectAll = (val: boolean | string) => {
  if (val) {
    data.value.forEach((row) => {
      taskTableDataRef.value?.toggleRowSelection(row, true)
    })
  } else {
    taskTableDataRef.value?.clearSelection()
  }
}

/** 获取当前表格选中行的邮件 ID 列表（公共逻辑提取） */
const getSelectedIds = (): number[] | null => {
  if (selectedRows.value.length === 0) {
    ElMessage.warning('Select emails first');
    return null;
  }
  return selectedRows.value.map((e: EmailListItem) => e.id);
}

/** 标记已读：通过 emailService 调用已读接口 */
const markRead = function () {
  const ids = getSelectedIds();
  if (!ids) return;
  emailService.markEmailsRead(ids).then((res: any) => {
    if (res.errorNo === 0) {
      updateList()
      ElMessage.success('Marked as read');
    } else {
      ElMessage.error(res.errorMsg)
    }
  })
}

const move = function (group_id: string, group_name: string) {
  const ids = getSelectedIds();
  if (!ids) return;
  ElMessageBox.confirm(lang.move_email_confirm, 'Warning', {
    confirmButtonText: 'OK', cancelButtonText: 'Cancel', type: 'warning'
  }).then(() => {
    emailService.moveEmails({group_id, group_name, ids}).then((res: any) => {
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
  const ids = getSelectedIds();
  if (!ids) return;

  /** 通过 isTrashGroup 判断当前分组是否为垃圾箱（status === 3） */
  const forcedDel = isTrashGroup(tag);

  ElMessageBox.confirm(lang.del_email_confirm, 'Warning', {
    confirmButtonText: 'OK', cancelButtonText: 'Cancel', type: 'warning'
  }).then(() => {
    emailService.deleteEmails({ids, forcedDel}).then((res: any) => {
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

/** 翻页：通过 emailService 获取指定页数据，使用当前 pageSize */
const pageChange = function (p: number) {
  emailService.getEmailList({tag: tag, page_size: pageSize.value, current_page: p, keyword: groupStore.keyword}).then((res: any) => {
    data.value = res.data.list || []
  })
}

/** 每页条数变更：重置到第 1 页并刷新列表 */
const handlePageSizeChange = function () {
  updateList()
}
</script>

<!-- 样式: Docusaurus BEM 风格 | 重构日期: 20260429 -->
<style scoped>
/* 列表视图容器 */
.mail-list {
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

.mail-list__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--ifm-spacing-md);
  flex-wrap: wrap;
  gap: var(--ifm-spacing-sm);
}

.mail-list__title h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--ifm-color-content);
  margin: 0;
}

.mail-list__actions {
  display: flex;
  gap: var(--ifm-spacing-sm);
  align-items: center;
  flex-wrap: wrap;
}

.mail-list__action-btn {
  border-radius: var(--ifm-global-radius);
  border-color: var(--ifm-border-color);
  color: var(--ifm-color-content-secondary);
  background: var(--ifm-background-color);
  font-weight: 500;
}

.mail-list__action-btn:hover {
  background-color: var(--ifm-background-hover-color);
  color: var(--ifm-color-content);
  border-color: var(--ifm-border-color);
}

.mail-list__compose-btn {
  border-radius: var(--ifm-global-radius);
  font-weight: 600;
  margin-left: var(--ifm-spacing-sm);
  padding-inline: 16px;
}

.mail-list__content {
  flex-grow: 1;
  border-radius: var(--ifm-global-radius);
  border: 1px solid var(--ifm-border-color);
  overflow: hidden;
}

.mail-list__table {
  width: 100%;
}

.mail-list__table :deep(tr) {
  transition: background-color var(--ifm-transition-fast);
}

.mail-list__table :deep(tr:hover > td) {
  background-color: var(--ifm-row-hover-background) !important;
}

.mail-list__table :deep(td) {
  padding: 10px 0;
  border-bottom: 1px solid var(--ifm-color-emphasis-300);
}

/* 未读圆点：自定义 CSS 圆点替代 el-badge，避免无子元素时定位异常 */
.mail-list__unread-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background-color: var(--ifm-color-primary);
  flex-shrink: 0;
}

.mail-list__status {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 4px;
}

.mail-list__row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.mail-list__row-main {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-md);
  flex-grow: 1;
  overflow: hidden;
}

/* 发件人：自适应宽度，窄屏可收缩 */
.mail-list__sender {
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

/* 主题：自适应宽度 */
.mail-list__subject {
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
.mail-list__snippet {
  color: var(--ifm-color-content-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 14px;
  flex-grow: 1;
  min-width: 60px;
}

.mail-list__meta {
  min-width: 80px;
  text-align: right;
  padding-right: var(--ifm-spacing-md);
}

.mail-list__date {
  font-size: 12px;
  color: var(--ifm-color-content-secondary);
}

/* 未读行加粗 */
.mail-list__row--unread .mail-list__sender,
.mail-list__row--unread .mail-list__subject {
  font-weight: 700;
  color: var(--ifm-color-content);
}

.mail-list__row--unread .mail-list__date {
  color: var(--ifm-color-primary);
  font-weight: 600;
}

/* 全选复选框样式：添加边框以增强视觉辨识度 */
.mail-list__select-all {
  margin-right: var(--ifm-spacing-xs);
  font-size: 14px;
  color: var(--ifm-color-content-secondary);
  border: 1px solid var(--ifm-border-color);
  border-radius: var(--ifm-global-radius);
  padding: 4px 8px;
}

/* 已选计数：高亮显示选中数量 */
.mail-list__selected-count {
  font-size: 13px;
  color: var(--ifm-color-primary);
  font-weight: 500;
  white-space: nowrap;
}

/* 分页区域：支持每页条数选择器 */
.mail-list__pagination {
  margin-top: var(--ifm-spacing-md);
  display: flex;
  justify-content: center;
  align-items: center;
  gap: var(--ifm-spacing-md);
  padding-bottom: var(--ifm-spacing-xs);
}

/* 每页条数选择器容器 */
.mail-list__page-size {
  display: flex;
  align-items: center;
  gap: var(--ifm-spacing-xs);
}

.mail-list__page-size-label {
  font-size: 13px;
  color: var(--ifm-color-content-secondary);
  white-space: nowrap;
}

.mail-list__page-size-select {
  width: 80px;
}

@media (max-width: 768px) {
  .mail-list__header {
    flex-direction: column;
    align-items: flex-start;
  }
  .mail-list__actions {
    width: 100%;
    overflow-x: auto;
    padding-bottom: var(--ifm-spacing-xs);
  }
  .mail-list__row-main {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--ifm-spacing-xs);
  }
  .mail-list__sender {
    width: 100%;
  }
  .mail-list__subject {
    max-width: 100%;
  }
  .mail-list__snippet {
    display: none;
  }
  .mail-list__row {
    align-items: flex-start;
  }
  .mail-list__meta {
    padding-top: 2px;
  }
}
</style>
