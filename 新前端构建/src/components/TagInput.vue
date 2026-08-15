<!--
  TagInput 标签式多选输入组件

  对齐旧前端 el-select multiple filterable allow-create 的体验：
  - 每个邮箱地址显示为独立标签（chip/tag）
  - 输入后按回车或逗号添加标签
  - 输入框失焦（blur）时自动识别输入内容为标签
  - 点击标签文本可进入编辑模式（回填到输入框）
  - 每个标签右侧有 × 可单独删除
  - 退格键可删除最后一个标签
  - 支持粘贴多个邮箱（逗号/分号/空格分隔）
  - 标签过多时出现滚动条（max-height 限制）

  创建日期: 20260610
  修改日期: 20260610
  修改原因: 
    1. 新增 blur 自动识别邮箱地址（无需手动回车）
    2. 新增标签点击编辑功能
    3. 新增 max-height 限制防止标签过多时撑大容器
-->
<template>
  <div class="tag-input" @click="focusInput">
    <!-- 已输入的标签列表（点击标签文本进入编辑模式） -->
    <span
      v-for="(tag, i) in modelValue"
      :key="i"
      class="tag-chip"
    >
      <span class="tag-text" @click.stop="editTag(i)" :title="tag">{{ tag }}</span>
      <button class="tag-remove" @click.stop="removeTag(i)">×</button>
    </span>

    <!-- 输入框（blur 时自动识别输入内容为标签） -->
    <input
      ref="inputRef"
      v-model="inputVal"
      type="text"
      class="tag-field"
      :placeholder="modelValue.length === 0 ? placeholder : ''"
      @keydown="onKeydown"
      @paste="onPaste"
      @blur="onBlur"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * TagInput 组件属性与事件
 *
 * modelValue: 标签数组（v-model 双向绑定）
 * placeholder: 无标签时的占位提示文字
 */
const props = defineProps<{
  modelValue: string[]
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

import { ref, nextTick } from 'vue'

const inputRef = ref<HTMLInputElement | null>(null)
const inputVal = ref('')

/** 聚焦输入框 */
function focusInput() {
  inputRef.value?.focus()
}

/**
 * 添加标签
 *
 * 支持多种分隔符：逗号、分号、空格。
 * 自动去除首尾空格，忽略空值和重复值。
 * @param value - 输入的文本
 */
function addTag(value: string) {
  /* 按逗号/分号/空格分割，去除首尾空格，过滤空值 */
  const parts = value.split(/[,;\s]+/).map((s: string) => s.trim()).filter((s: string) => s.length > 0)
  if (parts.length === 0) return

  const current = [...props.modelValue]
  const existingSet = new Set(current.map((e: string) => e.toLowerCase()))

  for (const part of parts) {
    if (!existingSet.has(part.toLowerCase())) {
      current.push(part)
      existingSet.add(part.toLowerCase())
    }
  }

  emit('update:modelValue', current)
  inputVal.value = ''
}

/** 删除指定索引的标签 */
function removeTag(index: number) {
  const updated = [...props.modelValue]
  updated.splice(index, 1)
  emit('update:modelValue', updated)
}

/**
 * 键盘事件处理
 *
 * - Enter / 逗号 / 分号：添加当前输入为标签
 * - Backspace：输入框为空时删除最后一个标签
 */
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' || e.key === ',' || e.key === ';') {
    e.preventDefault()
    if (inputVal.value.trim()) {
      addTag(inputVal.value)
    }
  } else if (e.key === 'Backspace' && !inputVal.value && props.modelValue.length > 0) {
    /* 退格键删除最后一个标签 */
    removeTag(props.modelValue.length - 1)
  }
}

/**
 * 粘贴事件处理
 *
 * 粘贴的文本可能包含多个邮箱（逗号/分号/换行分隔），
 * 自动分割并批量添加。
 */
function onPaste(e: ClipboardEvent) {
  const text = e.clipboardData?.getData('text') || ''
  if (text.includes(',') || text.includes(';') || text.includes('\n')) {
    e.preventDefault()
    addTag(text)
  }
  /* 单个邮箱粘贴不拦截，让用户按回车或失焦确认 */
}

/**
 * 失焦事件处理
 *
 * 输入框失焦时，如果输入框中有内容，自动识别为标签。
 * 修改日期: 20260610
 * 修改原因: 修复 BUG-1，邮箱地址输入后需回车确认的问题，改为失焦自动识别
 */
function onBlur() {
  /* 延迟执行，避免点击删除按钮时误触发 */
  setTimeout(() => {
    if (inputVal.value.trim()) {
      addTag(inputVal.value)
    }
  }, 100)
}

/**
 * 标签点击编辑
 *
 * 点击标签文本时，将该标签回填到输入框并删除该标签，
 * 用户可直接修改后回车或失焦确认。
 * 
 * @param index - 要编辑的标签索引
 * 修改日期: 20260610
 * 修改原因: 修复 BUG-1，标签需支持单击编辑
 */
function editTag(index: number) {
  /* 将标签值回填到输入框 */
  inputVal.value = props.modelValue[index]
  /* 从标签数组中移除该标签 */
  removeTag(index)
  /* 聚焦输入框以便用户编辑 */
  nextTick(() => {
    inputRef.value?.focus()
  })
}
</script>

<style scoped>
.tag-input {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 6px 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  min-height: 36px;
  /* max-height 限制：标签过多时出现滚动条，修复 BUG-2 按钮变大问题 */
  max-height: 120px;
  overflow-y: auto;
  align-items: center;
  cursor: text;
  transition: border-color var(--transition);
}

.tag-input:focus-within {
  border-color: var(--accent-color);
}

.tag-chip {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  padding: 2px 6px;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 12px;
  line-height: 1.4;
  max-width: 240px;
}

.tag-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  /* 点击标签文本可编辑，显示手型光标 */
  cursor: pointer;
}

.tag-text:hover {
  text-decoration: underline;
}

.tag-remove {
  border: none;
  background: transparent;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  padding: 0 1px;
  line-height: 1;
  flex-shrink: 0;
}

.tag-remove:hover {
  color: var(--danger-color);
}

.tag-field {
  flex: 1;
  min-width: 80px;
  border: none;
  outline: none;
  background: transparent;
  font-size: 13px;
  padding: 2px 0;
  color: var(--text-primary);
}

.tag-field::placeholder {
  color: var(--text-placeholder);
}
</style>