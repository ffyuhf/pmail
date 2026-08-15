<!--
  SetupView 初始化设置页视图

  从旧前端 fe/src/views/SetupView.vue 迁移，保持 Apple 极简风格。
  6 步向导：欢迎 → 数据库 → 密码 → 域名 → DNS → SSL

  核心修复：
  - Token 提取：兼容 hash 路由和普通路由两种 URL 格式
  - API 协议：统一使用 POST /api/setup（action + step + token）
  - 步骤顺序：与后端 setup.go 控制器一致
  - 功能对齐：多域名、PostgreSQL、SSL http/dns 挑战、完成后轮询跳转

  修改日期: 20260609
  修改原因: 旧版引导步骤与后端 API 不匹配，无法完成初始化。
-->
<template>
  <div class="setup-page">
    <div class="setup-card">
      <!-- 步骤指示器 -->
      <div class="steps">
        <div
          v-for="i in totalSteps"
          :key="i"
          class="step-dot"
          :class="{ active: i - 1 === step, done: i - 1 < step }"
        />
      </div>

      <!-- 步骤 0：欢迎 -->
      <template v-if="step === 0">
        <h2 class="setup-title">{{ lang.tks_pmail }}</h2>
        <p class="setup-desc">{{ lang.guid_desc }}</p>
      </template>

      <!-- 步骤 1：数据库 -->
      <div v-if="step === 1" class="step-content">
        <h3>{{ lang.select_db }}</h3>
        <p>{{ lang.db_desc }}</p>
        <div class="form-field">
          <label class="field-label">{{ lang.type }}</label>
          <select v-model="dbSettings.type" class="field-select" @change="onDbTypeChange">
            <option value="sqlite">SQLite3</option>
            <option value="mysql">MySQL</option>
            <option value="postgres">PostgreSQL</option>
          </select>
        </div>
        <div class="form-field" v-if="dbSettings.type === 'mysql'">
          <label class="field-label">{{ lang.mysql_dsn }}</label>
          <textarea v-model="dbSettings.dsn" class="field-textarea" rows="2"
            placeholder="root:12345@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local" />
        </div>
        <div class="form-field" v-if="dbSettings.type === 'postgres'">
          <label class="field-label">{{ lang.pg_dsn }}</label>
          <textarea v-model="dbSettings.dsn" class="field-textarea" rows="2"
            placeholder="postgres://postgres:12345@127.0.0.1:5432/pmail?sslmode=disable" />
        </div>
        <div class="form-field" v-if="dbSettings.type === 'sqlite'">
          <label class="field-label">{{ lang.sqlite_db_path }}</label>
          <input v-model="dbSettings.dsn" type="text" class="field-input" placeholder="./config/pmail.db" />
        </div>
      </div>

      <!-- 步骤 2：管理员密码 -->
      <div v-if="step === 2" class="step-content">
        <h3>{{ lang.setAdminPassword }}</h3>
        <div class="form-field">
          <label class="field-label">{{ lang.admin_account }}</label>
          <input v-model="adminSettings.account" type="text" class="field-input"
            :disabled="adminSettings.hadSeted" placeholder="admin" />
        </div>
        <div class="form-field">
          <label class="field-label">{{ lang.password }}</label>
          <input v-model="adminSettings.password" type="password" class="field-input"
            :disabled="adminSettings.hadSeted" />
        </div>
        <div class="form-field">
          <label class="field-label">{{ lang.enter_again }}</label>
          <input v-model="adminSettings.password2" type="password" class="field-input"
            :disabled="adminSettings.hadSeted" />
        </div>
      </div>

      <!-- 步骤 3：域名 -->
      <div v-if="step === 3" class="step-content">
        <h3>{{ lang.SetDomail }}</h3>
        <p>{{ lang.domain_desc }}</p>
        <div class="form-field">
          <label class="field-label">{{ lang.smtp_domain }}</label>
          <div class="input-group">
            <span class="input-prepend">smtp.</span>
            <input v-model="domainSettings.smtp_domain" type="text" class="field-input" placeholder="domain.com" />
          </div>
        </div>
        <div class="form-field">
          <label class="field-label">{{ lang.web_domain }}</label>
          <input v-model="domainSettings.web_domain" type="text" class="field-input" placeholder="pmail.domain.com" />
        </div>
        <div class="form-field">
          <label class="field-label">{{ lang.multi_domain_setting }}</label>
          <p class="field-hint">{{ lang.multi_domain_setting_desc }}</p>
          <button class="btn-add" @click="addDomain">+</button>
          <div v-for="(_, i) in domainSettings.multi_domain" :key="i" class="form-field" style="margin-top: 6px;">
            <input v-model="domainSettings.multi_domain[i]" type="text" class="field-input"
              :placeholder="`domain${Number(i) + 1}.com`" />
          </div>
        </div>
      </div>

      <!-- 步骤 4：DNS -->
      <div v-if="step === 4" class="step-content">
        <h3>{{ lang.setDNS }}</h3>
        <p>{{ lang.dns_desc }}</p>
        <div v-for="(records, domain) in dnsInfos" :key="String(domain)" class="dns-group">
          <h4 class="dns-domain">{{ domain }}</h4>
          <div class="dns-table">
            <div class="dns-row dns-header">
              <span class="dns-cell">HOSTNAME</span>
              <span class="dns-cell">TYPE</span>
              <span class="dns-cell">VALUE</span>
              <span class="dns-cell">TTL</span>
            </div>
            <div v-for="(row, idx) in records" :key="idx" class="dns-row">
              <span class="dns-cell" :title="(row.host === '' || row.host === '@') ? lang.dns_root_desc : ''">
                {{ row.host }}
              </span>
              <span class="dns-cell">{{ row.type }}</span>
              <span class="dns-cell" :title="row.tips || ''">{{ row.value }}</span>
              <span class="dns-cell">{{ row.ttl }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 步骤 5：SSL -->
      <div v-if="step === 5" class="step-content">
        <h3>{{ lang.setSSL }}</h3>
        <!-- 端口 80 警告 -->
        <div v-if="sslSettings.type === '0' && port !== 80" class="warn-msg">
          {{ lang.autoSSLWarn }}
        </div>
        <div class="form-field">
          <label class="field-label">{{ lang.type }}</label>
          <select v-model="sslSettings.type" class="field-select" :disabled="dnsChecking">
            <option value="0">{{ lang.ssl_auto }}</option>
            <option value="1">{{ lang.ssl_manuallyf }}</option>
          </select>
        </div>
        <!-- HTTP/DNS 验证方式选择（自动模式） -->
        <div class="form-field" v-if="sslSettings.type === '0'">
          <label class="field-label">{{ lang.ssl_challenge_type }}</label>
          <div class="challenge-row">
            <select v-model="sslSettings.challenge" class="field-select" :disabled="dnsChecking">
              <option value="http">{{ lang.ssl_auto_http }}</option>
              <option value="dns">{{ lang.ssl_auto_dns }}</option>
            </select>
            <span class="help-icon" :title="lang.challenge_typ_desc">?</span>
          </div>
        </div>
        <!-- 手动模式：密钥路径 -->
        <div class="form-field" v-if="sslSettings.type === '1'">
          <label class="field-label">{{ lang.ssl_key_path }}</label>
          <input v-model="sslSettings.key_path" type="text" class="field-input" placeholder="./config/ssl/private.key" />
        </div>
        <div class="form-field" v-if="sslSettings.type === '1'">
          <label class="field-label">{{ lang.ssl_crt_path }}</label>
          <input v-model="sslSettings.crt_path" type="text" class="field-input" placeholder="./config/ssl/public.crt" />
        </div>
      </div>

      <!-- DNS 挑战验证中 -->
      <div v-if="dnsChecking" class="step-content">
        <h3>{{ lang.setDNS }}</h3>
        <p>{{ lang.dns_challenge_wait }}</p>
        <div class="dns-table">
          <div class="dns-row dns-header">
            <span class="dns-cell">HOSTNAME</span>
            <span class="dns-cell">TYPE</span>
            <span class="dns-cell">VALUE</span>
            <span class="dns-cell">TTL</span>
          </div>
          <div v-for="(row, idx) in sslSettings.paramsList" :key="idx" class="dns-row">
            <span class="dns-cell">{{ row.host }}</span>
            <span class="dns-cell">{{ row.type }}</span>
            <span class="dns-cell" :title="row.tips || ''">{{ row.value }}</span>
            <span class="dns-cell">{{ row.ttl }}</span>
          </div>
          <div v-if="sslSettings.paramsList.length === 0" class="dns-loading">Loading...</div>
        </div>
      </div>

      <!-- 全屏加载 -->
      <div v-if="fullscreenLoading" class="loading-overlay">
        <div class="loading-text">{{ waitDesc }}</div>
      </div>

      <!-- 错误提示 -->
      <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>

      <!-- 导航按钮 -->
      <div class="step-actions" v-if="!fullscreenLoading">
        <button class="btn-primary" @click="next">{{ lang.next }}</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import lang from '@/i18n'
import { setupService } from '@/services/setupService'

const route = useRoute()

/**
 * 从 window.location.search 中解析指定 query 参数。
 * 用于读取 hash 路由模式下 # 之前的 query 参数。
 * @param name - 参数名
 * @returns 参数值或空字符串
 */
function getQueryParamFromSearch(name: string): string {
  const searchParams = new URLSearchParams(window.location.search)
  return searchParams.get(name) || ''
}

/** Setup Token：先从 route.query 取，再从 window.location.search 兜底 */
const setupToken = ref<string>(
  (route.query.token as string) || getQueryParamFromSearch('token') || ''
)

const totalSteps = 6
const step = ref(0)
const errorMsg = ref('')
const waitDesc = ref(lang.wait_desc)
const fullscreenLoading = ref(false)
const dnsChecking = ref(false)
const port = ref(80)

/** 管理员密码设置 */
const adminSettings = reactive({
  account: 'admin',
  password: '',
  password2: '',
  hadSeted: false,
})

/** 数据库配置 */
const dbSettings = reactive({
  type: 'sqlite',
  dsn: './config/pmail.db',
})

/** 域名配置 */
const domainSettings = reactive({
  web_domain: '',
  smtp_domain: '',
  multi_domain: [] as string[],
})

/** SSL 配置 */
const sslSettings = reactive({
  type: '0',
  challenge: 'http',
  key_path: './config/ssl/private.key',
  crt_path: './config/ssl/public.crt',
  paramsList: [] as any[],
})

/** DNS 记录信息（按域名分组） */
const dnsInfos = ref<Record<string, any>>({})

/** 数据库类型切换时清空 DSN */
function onDbTypeChange() {
  dbSettings.dsn = ''
}

/** 添加附加域名 */
function addDomain() {
  domainSettings.multi_domain.push('')
}

/** 显示错误信息 */
function showError(msg: string) {
  errorMsg.value = msg
}

// ── 预加载函数 ──

/**
 * 获取数据库配置
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包 response.data，直接使用 res
 */
function getDbConfig() {
  setupService.getDatabaseConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      if (res.data) {
        dbSettings.type = res.data.db_type || 'sqlite'
        dbSettings.dsn = res.data.db_dsn || ''
      }
    }
  }).catch(() => showError(lang.fail))
}

/**
 * 保存数据库配置
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo
 */
function setDbConfig() {
  if (dbSettings.type === 'sqlite' && !dbSettings.dsn) dbSettings.dsn = './config/pmail.db'
  else if (!dbSettings.dsn) {
    showError(lang.err_db_dsn_empty)
    return
  }
  setupService.setDatabaseConfig(setupToken.value, dbSettings.type, dbSettings.dsn).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      errorMsg.value = ''
      step.value++
      getPasswordConfig()
    }
  }).catch(() => showError(lang.fail))
}

/**
 * 获取密码配置状态
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo / res.data
 */
function getPasswordConfig() {
  setupService.getPasswordConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      adminSettings.hadSeted = res.data !== '' && res.data !== undefined && res.data !== null
      if (adminSettings.hadSeted) {
        adminSettings.account = String(res.data)
        adminSettings.password = '*******'
        adminSettings.password2 = '*******'
      }
    }
  }).catch(() => showError(lang.fail))
}

/** 保存管理员密码 */
function setPassword() {
  if (adminSettings.hadSeted) {
    errorMsg.value = ''
    step.value++
    getDomainConfig()
    return
  }
  if (adminSettings.password !== adminSettings.password2) {
    showError(lang.err_pwd_diff)
    return
  }
  if (!adminSettings.password) {
    showError(lang.err_required_pwd)
    return
  }
  setupService.setPassword(setupToken.value, adminSettings.account, adminSettings.password).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      errorMsg.value = ''
      step.value++
      getDomainConfig()
    }
  }).catch(() => showError(lang.fail))
}

/**
 * 获取域名配置
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo / res.data
 */
function getDomainConfig() {
  setupService.getDomainConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      if (res.data) {
        domainSettings.web_domain = res.data.web_domain || ''
        domainSettings.smtp_domain = res.data.smtp_domain || ''
        domainSettings.multi_domain = res.data.domains || []
      }
    }
  }).catch(() => showError(lang.fail))
}

/**
 * 保存域名配置
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo
 */
function setDomainConfig() {
  setupService.setDomainConfig(
    setupToken.value,
    domainSettings.web_domain,
    domainSettings.smtp_domain,
    domainSettings.multi_domain.join(','),
  ).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      errorMsg.value = ''
      step.value++
      getDnsConfig()
    }
  }).catch(() => showError(lang.fail))
}

/**
 * 获取 DNS 配置
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo / res.data
 */
function getDnsConfig() {
  setupService.getDnsConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      dnsInfos.value = res.data
    }
  }).catch(() => showError(lang.fail))
}

/**
 * 获取 SSL 配置
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo / res.data
 */
function getSSLConfig() {
  setupService.getSslConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      if (res.data) {
        sslSettings.type = String(res.data.type || '0')
        if (sslSettings.type === '2') {
          sslSettings.type = '0'
          sslSettings.challenge = 'dns'
        }
        port.value = res.data.port || 80
      }
    }
  }).catch(() => showError(lang.fail))
}

/** 保存 SSL 配置 */
function setSSLConfig() {
  fullscreenLoading.value = true
  let sslType = sslSettings.type
  if (sslType === '0' && sslSettings.challenge === 'dns') {
    sslType = '2'
  }
  setupService.setSslConfig(
    setupToken.value, sslType, sslSettings.key_path, sslSettings.crt_path,
  ).then((res: any) => {
    if (res.errorNo !== 0) {
      fullscreenLoading.value = false
      showError(res.errorMsg || lang.fail)
    } else {
      if (Number(sslType) === 2) {
        fullscreenLoading.value = false
        dnsChecking.value = true
        getSSLDNSParams()
      }
      checkStatus()
    }
  }).catch(() => {
    fullscreenLoading.value = false
    showError(lang.fail)
  })
}

/** 轮询 /api/ping，服务重启完成后跳转 */
function checkStatus() {
  axios.post('/api/ping', {}).then((res: any) => {
    if (res.data?.errorNo !== 0) {
      setTimeout(checkStatus, 1000)
    } else {
      if (Number(sslSettings.type) === 1) {
        window.location.href = 'http://' + domainSettings.web_domain
      } else {
        window.location.href = 'https://' + domainSettings.web_domain
      }
    }
  }).catch(() => {
    setTimeout(checkStatus, 1000)
  })
}

/**
 * 获取 SSL DNS 验证参数
 * 修改日期: 20260609
 * 修改原因: 修正响应格式 — axios 拦截器已解包，使用 res.errorNo / res.data
 */
function getSSLDNSParams() {
  setupService.getSslDnsParams(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      showError(res.errorMsg || lang.fail)
    } else {
      if (res.data && res.data.length > 0) {
        sslSettings.paramsList = res.data
      }
    }
  }).catch(() => { /* 忽略，稍后重试 */ })

  if (sslSettings.paramsList.length === 0) {
    setTimeout(getSSLDNSParams, 1000)
  }
}

/** 主流程：下一步按钮 */
function next() {
  errorMsg.value = ''
  switch (step.value) {
    case 0:
      step.value++
      getDbConfig()
      break
    case 1:
      setDbConfig()
      break
    case 2:
      setPassword()
      break
    case 3:
      setDomainConfig()
      break
    case 4:
      getSSLConfig()
      step.value++
      break
    case 5:
      if (dnsChecking.value) {
        fullscreenLoading.value = true
        waitDesc.value = lang.dns_challenge_wait
      } else {
        setSSLConfig()
      }
      break
  }
}
</script>

<style scoped>
.setup-page {
  min-height: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  background: var(--bg-secondary);
  padding: 40px 16px;
}

.setup-card {
  width: 100%;
  max-width: 640px;
  background: var(--bg-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  padding: 40px 32px;
  position: relative;
}

/* 步骤指示器 */
.steps {
  display: flex;
  gap: 8px;
  justify-content: center;
  margin-bottom: 24px;
}

.step-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: var(--border-color);
  transition: background var(--transition);
}

.step-dot.active { background: var(--accent-color); }
.step-dot.done { background: var(--success-color); }

.setup-title {
  text-align: center;
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 4px;
}

.setup-desc {
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 24px;
}

/* 步骤内容 */
.step-content {
  margin-bottom: 24px;
}

.step-content h3 {
  font-size: 16px;
  margin-bottom: 8px;
}

.step-content p {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

/* 表单字段 */
.form-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-bottom: 12px;
}

.field-label {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
}

.field-hint {
  font-size: 12px;
  color: var(--text-secondary);
  margin: 0 0 4px;
}

.field-input,
.field-select,
.field-textarea {
  padding: 10px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  font-size: 14px;
  outline: none;
  background: var(--bg-color);
  color: var(--text-primary);
  font-family: inherit;
}

.field-input:focus,
.field-select:focus,
.field-textarea:focus {
  border-color: var(--accent-color);
}

.field-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.field-textarea {
  resize: vertical;
}

/* 输入组（smtp 前缀） */
.input-group {
  display: flex;
  align-items: stretch;
}

.input-prepend {
  display: flex;
  align-items: center;
  padding: 0 10px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-right: none;
  border-radius: var(--radius) 0 0 var(--radius);
  font-size: 13px;
  color: var(--text-secondary);
}

.input-group .field-input {
  border-radius: 0 var(--radius) var(--radius) 0;
}

/* 添加域名按钮 */
.btn-add {
  width: 28px;
  height: 28px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  background: var(--bg-secondary);
  font-size: 16px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--accent-color);
  margin-bottom: 8px;
}

.btn-add:hover {
  background: var(--bg-hover);
}

/* DNS 表格 */
.dns-group {
  margin-bottom: 20px;
}

.dns-domain {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 8px;
}

.dns-table {
  width: 100%;
  border: 1px solid var(--border-color);
  border-radius: var(--radius);
  overflow: hidden;
}

.dns-row {
  display: grid;
  grid-template-columns: 110px 110px 1fr 110px;
  font-size: 12px;
}

.dns-header {
  background: var(--bg-secondary);
  font-weight: 600;
}

.dns-cell {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border-color);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.dns-loading {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
}

/* SSL 挑战方式行 */
.challenge-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.challenge-row .field-select {
  flex: 1;
}

.help-icon {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  cursor: help;
  flex-shrink: 0;
}

/* 警告 */
.warn-msg {
  padding: 10px 14px;
  border-radius: var(--radius);
  background: #fff3cd;
  color: #856404;
  font-size: 13px;
  margin-bottom: 12px;
}

/* 错误 */
.error-msg {
  font-size: 13px;
  color: var(--danger-color);
  margin-bottom: 12px;
}

/* 加载遮罩 */
.loading-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.85);
  border-radius: var(--radius-lg);
  z-index: 10;
}

.loading-text {
  font-size: 15px;
  color: var(--text-secondary);
}

/* 导航按钮 */
.step-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.btn-primary {
  padding: 8px 20px;
  border: none;
  border-radius: var(--radius);
  font-size: 14px;
  cursor: pointer;
  background: var(--accent-color);
  color: #fff;
  transition: all var(--transition);
}

.btn-primary:hover { background: var(--accent-hover); }
</style>