<!--
  HomeView 主页视图

  三栏式邮件主界面：
  - 左侧 IconNav（68px）：分组切换、写邮件、设置
  - 中间 ListPanel（340px）：邮件列表
  - 右侧 ContentPanel（自适应）：邮件详情

  创建日期: 20260609
-->
<template>
  <div class="home-layout">
    <!-- 左侧图标导航 -->
    <IconNav
      @select-group="onSelectGroup"
      @compose="showCompose = true"
      @open-settings="onOpenSettings"
      @logout="onLogout"
    />

    <!-- 中间邮件列表 -->
    <ListPanel
      :emails="emails"
      :active-email-id="activeEmailId"
      :selected-ids="selectedIds"
      :loading="loading"
      :current-tag="groupStore.currentTag"
      :groups="groupList"
      :all-selected="allSelected"
      @select-email="onSelectEmail"
      @select-group="onSelectGroup"
      @toggle-select="toggleSelect"
      @toggle-select-all="toggleSelectAll"
      @batch-delete="onBatchDelete"
      @batch-read="onBatchRead"
      @batch-move="onBatchMove"
      @load-more="loadMore"
      @search="onSearch"
    />

    <!-- 右侧邮件详情 -->
    <ContentPanel
      :email="currentDetail"
      @reply="onReply"
      @delete="onDeleteSingle"
    />

    <!-- 写邮件弹窗（传递 replySender/replyDomain 用于回信自动选择发件人） -->
    <ComposeModal
      v-if="showCompose"
      :reply-to="replyTo"
      :reply-sender="replySender"
      :reply-domain="replyDomain"
      @close="showCompose = false"
      @sent="onEmailSent"
    />

    <!-- 设置面板 -->
    <SettingsPanel
      v-if="settingsVisible"
      :active-tab="settingsTab"
      @close="settingsVisible = false"
    />

    <!-- 移动分组弹窗 -->
    <MoveModal
      v-if="showMoveModal"
      :groups="groupStore.flatGroups"
      @close="showMoveModal = false"
      @confirm="onMoveConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import lang from '@/i18n'
import { useGroupStore } from '@/stores/group'
import { useGlobalStore } from '@/stores/global'
import { getEmailList, getEmailDetail, deleteEmails, markEmailsRead, moveEmails } from '@/services/emailService'
import { logout as logoutApi } from '@/services/userService'
import { getGroupList } from '@/services/groupService'
import type { EmailListItem, EmailDetail } from '@/types/api'
import { DEFAULT_GROUP_TAG } from '@/utils/constants'
import { parseGroupTag } from '@/utils/groupTag'
import IconNav from '@/components/IconNav.vue'
import ListPanel from '@/components/ListPanel.vue'
import ContentPanel from '@/components/ContentPanel.vue'
import ComposeModal from '@/components/ComposeModal.vue'
import SettingsPanel from '@/components/SettingsPanel.vue'
import MoveModal from '@/components/MoveModal.vue'

const router = useRouter()
const groupStore = useGroupStore()
const globalStore = useGlobalStore()

/* ---- 邮件列表状态 ---- */
const emails = ref<EmailListItem[]>([])
const loading = ref(false)
const currentPage = ref(1)
const totalPages = ref(1)
const pageSize = 20
const searchKeyword = ref('') // 搜索关键词（从 ListPanel 传入）

/* ---- 邮件详情状态 ---- */
const currentDetail = ref<EmailDetail | null>(null)
const activeEmailId = ref<number | null>(null)

/* ---- 批量选择状态 ---- */
const selectedIds = ref<number[]>([])

/* ---- 弹窗状态 ---- */
const showCompose = ref(false)
const replyTo = ref<EmailDetail | null>(null)
/** 回信发件人前缀（对齐旧前端 route.query.reply_sender） */
const replySender = ref<string | undefined>(undefined)
/** 回信发件人域名（对齐旧前端 route.query.reply_domain） */
const replyDomain = ref<string | undefined>(undefined)
const settingsVisible = ref(false)
const settingsTab = ref<'security' | 'groups' | 'rules' | 'users' | 'plugins'>('security')
const showMoveModal = ref(false)

/** 分组下拉列表（用于 ListPanel 的标签栏，使用 flatGroups 包含所有子文件夹） */
const groupList = computed(() => {
  return groupStore.flatGroups.map((g) => ({
    tag: g.tag,
    label: g.label,
  }))
})

/** 是否全选 */
const allSelected = computed(() => {
  return emails.value.length > 0 && selectedIds.value.length === emails.value.length
})

/* ---- 数据加载 ---- */

/** 加载邮件列表（含搜索关键词） */
async function fetchEmails() {
  loading.value = true
  try {
    /* 参数名对齐后端 emailRequest 结构体：tag, current_page, page_size */
    const res: any = await getEmailList({
      tag: groupStore.currentTag,
      current_page: currentPage.value,
      page_size: pageSize,
      keyword: searchKeyword.value || undefined,
    })
    /* axios 拦截器已解包 response.data，直接读 errorNo */
    if (res.errorNo === 0) {
      const data = res.data
      if (currentPage.value === 1) {
        emails.value = data.list || []
      } else {
        emails.value.push(...(data.list || []))
      }
      totalPages.value = data.total_page || 1
    }
  } finally {
    loading.value = false
  }
}

/** 加载邮件详情 */
async function fetchDetail(id: number) {
  try {
    const res: any = await getEmailDetail(id)
    /* axios 拦截器已解包，直接读 errorNo */
    if (res.errorNo === 0) {
      currentDetail.value = res.data
    }
  } catch {
    currentDetail.value = null
  }
}

/** 加载更多（滚动到底部） */
function loadMore() {
  if (currentPage.value < totalPages.value && !loading.value) {
    currentPage.value++
    fetchEmails()
  }
}

/* ---- 事件处理 ---- */

/** 切换分组 */
function onSelectGroup(tag: string) {
  groupStore.setCurrentTag(tag)
  currentPage.value = 1
  selectedIds.value = []
  currentDetail.value = null
  activeEmailId.value = null
  fetchEmails()
}

/** 搜索邮件（从 ListPanel 的 search 事件触发） */
function onSearch(query: string) {
  searchKeyword.value = query
  currentPage.value = 1
  selectedIds.value = []
  currentDetail.value = null
  activeEmailId.value = null
  fetchEmails()
}

/** 选中邮件 */
function onSelectEmail(email: EmailListItem) {
  activeEmailId.value = email.id
  fetchDetail(email.id)
  /* 自动标记已读 */
  if (!email.is_read) {
    markEmailsRead([email.id])
    email.is_read = true
  }
}

/** 切换单个选中 */
function toggleSelect(id: number) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) selectedIds.value.splice(idx, 1)
  else selectedIds.value.push(id)
}

/** 切换全选 */
function toggleSelectAll() {
  if (allSelected.value) {
    selectedIds.value = []
  } else {
    selectedIds.value = emails.value.map((e) => e.id)
  }
}

/** 检测当前分组是否为垃圾箱，决定是否强制删除（status === 3） */
function isTrashGroup(): boolean {
  return parseGroupTag(groupStore.currentTag).status === 3
}

/** 批量删除（传递 ue_ids 用于精确定位，垃圾箱中强制删除） */
async function onBatchDelete() {
  if (!confirm(lang.del_email_confirm)) return
  const ueIds = emails.value
    .filter((e) => selectedIds.value.includes(e.id))
    .map((e) => e.ue_id)
  await deleteEmails(selectedIds.value, ueIds, isTrashGroup())
  selectedIds.value = []
  currentPage.value = 1
  fetchEmails()
}

/** 批量标记已读 */
async function onBatchRead() {
  await markEmailsRead(selectedIds.value)
  emails.value.forEach((e) => {
    if (selectedIds.value.includes(e.id)) e.is_read = true
  })
  selectedIds.value = []
}

/** 批量移动（弹出分组选择） */
function onBatchMove() {
  showMoveModal.value = true
}

/** 确认移动到指定分组（传递 ue_ids 和 group_name 用于精确定位） */
async function onMoveConfirm(groupId: number, groupName: string) {
  const ueIds = emails.value
    .filter((e) => selectedIds.value.includes(e.id))
    .map((e) => e.ue_id)
  await moveEmails(selectedIds.value, ueIds, groupId, groupName)
  showMoveModal.value = false
  selectedIds.value = []
  currentPage.value = 1
  fetchEmails()
}

/** 删除单封邮件（传递 ue_id 用于精确定位，垃圾箱中强制删除） */
async function onDeleteSingle(email: EmailDetail) {
  if (!confirm(lang.del_email_confirm)) return
  await deleteEmails([email.id], email.ue_id ? [email.ue_id] : undefined, isTrashGroup())
  currentDetail.value = null
  activeEmailId.value = null
  currentPage.value = 1
  fetchEmails()
}

/**
 * 回复邮件
 *
 * 从 EmailDetail 的 to 字段（收件人）中解析当前用户的邮箱地址，
 * 提取前缀（replySender）和域名（replyDomain）传给 ComposeModal，
 * 用于自动选择发件人（对齐旧前端 EditerView fillReplyParams 逻辑）。
 *
 * 修改日期: 20260610
 * 修改原因: 对齐旧前端回信参数传递，自动选择发件人域名
 */
function onReply(email: EmailDetail) {
  replyTo.value = email

  /* 从 to 字段中解析当前用户的邮箱地址 */
  replySender.value = undefined
  replyDomain.value = undefined

  if (email.to && globalStore.userInfo) {
    try {
      const toList = JSON.parse(email.to)
      if (Array.isArray(toList)) {
        /* 查找 to 列表中属于当前用户域名的邮箱 */
        const userDomains = globalStore.userInfo.domains || []
        for (const contact of toList) {
          const addr = contact.EmailAddress || ''
          for (const domain of userDomains) {
            if (addr.endsWith('@' + domain)) {
              replySender.value = addr.split('@')[0]
              replyDomain.value = domain
              break
            }
          }
          if (replySender.value) break
        }
      }
    } catch {
      /* to 字段非 JSON，尝试直接匹配 */
    }
  }

  showCompose.value = true
}

/**
 * 邮件发送成功回调
 *
 * 对齐旧前端 EditerView send 成功后逻辑：
 * 1. 切换分组到发件箱（tag = {"type":1,"status":-1}）
 * 2. 刷新列表
 *
 * 修改日期: 20260610
 * 修改原因: 对齐旧前端发送成功后自动跳转发件箱
 */
function onEmailSent() {
  showCompose.value = false
  replyTo.value = null
  replySender.value = undefined
  replyDomain.value = undefined

  /* 切换到发件箱并刷新列表（对齐旧前端 groupStore.name = lang.outbox） */
  const outboxTag = '{"type":1,"status":-1}'
  onSelectGroup(outboxTag)
}

/** 打开设置 */
function onOpenSettings() {
  settingsTab.value = 'security'
  settingsVisible.value = true
}

/** 注销 */
async function onLogout() {
  try {
    await logoutApi()
  } finally {
    globalStore.clearUser()
    router.push('/login')
  }
}

/* ---- 初始化 ---- */
onMounted(async () => {
  await globalStore.fetchUserInfo()
  await groupStore.fetchGroups()
  await fetchEmails()
})
</script>

<style scoped>
.home-layout {
  display: flex;
  height: 100%;
  width: 100%;
}

/* 移动端适配 */
@media (max-width: 768px) {
  .home-layout {
    flex-direction: column;
  }
}
</style>