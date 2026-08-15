import { ref } from 'vue'
import { defineStore } from 'pinia'
import lang from '../i18n/i18n';

const useGroupStore = defineStore('group', () => {
  const tag = ref("")
  const name = ref(lang.inbox)
  /** 搜索关键词：由侧边栏搜索框设置，列表页读取并发送到后端 API */
  const keyword = ref("")
  return { tag, name, keyword }
})

export default useGroupStore
