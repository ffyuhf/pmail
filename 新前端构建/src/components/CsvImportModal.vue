<!--
  CsvImportModal CSV 导入收件人弹窗

  支持拖拽上传 CSV 文件，解析并预览后导入到收件人字段。

  创建日期: 20260609
-->
<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="csv-modal">
      <div class="modal-header">
        <h2>{{ lang.csv_import }}</h2>
        <button class="close-btn" @click="$emit('close')">×</button>
      </div>

      <div class="modal-body">
        <p class="desc">{{ lang.csv_import_desc }}</p>

        <!-- 上传区域 -->
        <div
          class="upload-area"
          :class="{ dragover: isDragging }"
          @dragover.prevent="isDragging = true"
          @dragleave="isDragging = false"
          @drop.prevent="onDrop"
          @click="pickFile"
        >
          <p>{{ lang.csv_upload_area }}</p>
        </div>
        <input ref="fileInput" type="file" accept=".csv,.txt" class="hidden" @change="onFileSelected" />

        <!-- 目标字段选择（对齐旧前端 csvTargetField: to/cc/bcc 选择器） -->
        <div v-if="recipients.length > 0" class="target-field-row">
          <span class="target-label">{{ lang.csv_target_field }}</span>
          <select v-model="targetField" class="target-select">
            <option value="to">{{ lang.to }}</option>
            <option value="cc">{{ lang.cc }}</option>
            <option value="bcc">{{ lang.bcc }}</option>
          </select>
        </div>

        <!-- 预览表格 -->
        <div v-if="recipients.length > 0" class="preview">
          <div class="preview-header">
            <span>{{ lang.csv_preview }} ({{ recipients.length }})</span>
            <div>
              <button class="preview-btn" @click="selectAll">{{ lang.csv_select_all }}</button>
              <button class="preview-btn" @click="deselectAll">{{ lang.csv_deselect_all }}</button>
            </div>
          </div>
          <div class="preview-table">
            <div v-for="(r, i) in recipients" :key="i" class="preview-row">
              <input type="checkbox" v-model="r.selected" />
              <span class="col-name">{{ r.name || '—' }}</span>
              <span class="col-email">{{ r.email }}</span>
            </div>
          </div>
        </div>

        <!-- 错误提示 -->
        <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>
      </div>

      <div class="modal-footer">
        <button class="btn-secondary" @click="$emit('close')">{{ lang.cancel || 'Cancel' }}</button>
        <button class="btn-primary" :disabled="selectedCount === 0" @click="doImport">
          {{ lang.csv_confirm_import }} ({{ selectedCount }})
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import lang from '@/i18n'
import { parseCsv, readCsvFile, type CsvRecipient } from '@/utils/csvParser'

const emit = defineEmits<{
  'close': []
  /** 导入事件，传递邮箱列表和目标字段（to/cc/bcc，对齐旧前端 csvTargetField） */
  'import': [emails: string[], targetField: 'to' | 'cc' | 'bcc']
}>()

const fileInput = ref<HTMLInputElement | null>(null)
const isDragging = ref(false)
const errorMsg = ref('')

/** 导入目标字段（对齐旧前端 csvTargetField），默认 to */
const targetField = ref<'to' | 'cc' | 'bcc'>('to')

interface SelectableRecipient extends CsvRecipient { selected: boolean }
const recipients = ref<SelectableRecipient[]>([])

const selectedCount = computed(() => recipients.value.filter((r) => r.selected).length)

function pickFile() {
  fileInput.value?.click()
}

async function onFileSelected(e: Event) {
  const target = e.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) await processFile(file)
  target.value = ''
}

async function onDrop(e: DragEvent) {
  isDragging.value = false
  const file = e.dataTransfer?.files?.[0]
  if (file) await processFile(file)
}

async function processFile(file: File) {
  errorMsg.value = ''
  if (file.size > 5 * 1024 * 1024) {
    errorMsg.value = lang.csv_file_too_large
    return
  }
  try {
    const text = await readCsvFile(file)
    const result = parseCsv(text)
    if (result.recipients.length === 0) {
      errorMsg.value = lang.csv_no_email_found
      return
    }
    recipients.value = result.recipients.map((r) => ({ ...r, selected: true }))
  } catch {
    errorMsg.value = lang.fail
  }
}

function selectAll() {
  recipients.value.forEach((r) => { r.selected = true })
}

function deselectAll() {
  recipients.value.forEach((r) => { r.selected = false })
}

function doImport() {
  const emails = recipients.value.filter((r) => r.selected).map((r) => r.email)
  emit('import', emails, targetField.value)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.csv-modal {
  width: 520px;
  max-height: 80vh;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h2 { font-size: 16px; font-weight: 600; }

.close-btn {
  width: 32px; height: 32px; border: none; background: transparent;
  font-size: 20px; cursor: pointer; border-radius: var(--radius-sm);
  display: flex; align-items: center; justify-content: center;
}
.close-btn:hover { background: var(--bg-hover); }

.modal-body { padding: 16px 20px; flex: 1; overflow-y: auto; }

.desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; }

.upload-area {
  border: 2px dashed var(--border-color);
  border-radius: var(--radius);
  padding: 32px;
  text-align: center;
  color: var(--text-placeholder);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition);
  margin-bottom: 16px;
}
.upload-area:hover,
.upload-area.dragover {
  border-color: var(--accent-color);
  background: rgba(0, 0, 0, 0.02);
}

.hidden { display: none; }

/* 目标字段选择行（对齐旧前端 csv-toolbar） */
.target-field-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.target-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}
.target-select {
  padding: 4px 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 13px;
  background: var(--bg-secondary);
  color: var(--text-primary);
  outline: none;
}

.preview { margin-bottom: 12px; }

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 500;
}

.preview-btn {
  padding: 2px 8px;
  border: 1px solid var(--border-color);
  background: transparent;
  border-radius: var(--radius-sm);
  font-size: 11px;
  cursor: pointer;
  margin-left: 4px;
}
.preview-btn:hover { background: var(--bg-hover); }

.preview-table {
  max-height: 240px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
}

.preview-row {
  display: flex;
  gap: 8px;
  padding: 6px 8px;
  align-items: center;
  font-size: 12px;
  border-bottom: 1px solid var(--border-color);
}
.preview-row:last-child { border-bottom: none; }

.col-name { min-width: 80px; color: var(--text-secondary); }
.col-email { flex: 1; }

.error-msg { font-size: 13px; color: var(--danger-color); }

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 20px;
  border-top: 1px solid var(--border-color);
}

.btn-primary, .btn-secondary {
  padding: 8px 16px;
  border: none;
  border-radius: var(--radius);
  font-size: 13px;
  cursor: pointer;
  transition: all var(--transition);
}

.btn-primary {
  background: var(--accent-color);
  color: #fff;
}
.btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-secondary {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
}
.btn-secondary:hover { background: var(--bg-hover); }
</style>