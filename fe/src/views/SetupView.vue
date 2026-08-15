<template>
  <!-- 安装向导（Docusaurus BEM 风格） -->
  <div class="setup">
    <el-steps :active="active" align-center finish-status="success" class="setup__steps">
      <el-step :title="lang.welcome"/>
      <el-step :title="lang.setDatabase"/>
      <el-step :title="lang.setAdminPassword"/>
      <el-step :title="lang.SetDomail"/>
      <el-step :title="lang.setDNS"/>
      <el-step :title="lang.setSSL"/>
    </el-steps>


    <div v-if="active === 0" class="setup__card">
      <div class="setup__desc">
        <h2>{{ lang.tks_pmail }}</h2>
        <div style="margin-top: 10px;">{{ lang.guid_desc }}</div>
      </div>
    </div>

    <div v-if="active === 1" class="setup__card">
      <div class="setup__desc">
        <h2>{{ lang.select_db }}</h2>
        <div style="margin-top: 10px;">{{ lang.db_desc }}</div>
      </div>
      <div class="setup__form" style="width: 400px;">
        <el-form label-width="120px">
          <el-form-item :label="lang.type">
            <el-select :placeholder="lang.db_select_ph" v-model="dbSettings.type"
                       @change="dbSettings.dsn = ''">
              <el-option label="MySQL" value="mysql"/>
              <el-option label="SQLite3" value="sqlite"/>
              <el-option label="PostgreSQL" value="postgres"/>
            </el-select>
          </el-form-item>

          <el-form-item :label="lang.mysql_dsn" v-if="dbSettings.type === 'mysql'">
            <el-input :rows="2" type="textarea" v-model="dbSettings.dsn"
                      placeholder="root:12345@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local"></el-input>
          </el-form-item>

          <el-form-item :label="lang.pg_dsn" v-if="dbSettings.type === 'postgres'">
            <el-input :rows="2" type="textarea" v-model="dbSettings.dsn"
                      placeholder="postgres://postgres:12345@127.0.0.1:5432/pmail?sslmode=disable"></el-input>
          </el-form-item>

          <el-form-item :label="lang.sqlite_db_path" v-if="dbSettings.type === 'sqlite'">
            <el-input v-model="dbSettings.dsn" placeholder="./config/pmail.db"></el-input>
          </el-form-item>
        </el-form>
      </div>
    </div>


    <div v-if="active === 2" class="setup__card">
      <div class="setup__desc">
        <h2>{{ lang.setAdminPassword }}</h2>
      </div>
      <div class="setup__form" style="width: 400px;">
        <el-form label-width="120px">

          <el-form-item :label="lang.admin_account">
            <el-input v-bind:disabled="adminSettings.hadSeted" placeholder="admin"
                      v-model="adminSettings.account"></el-input>
          </el-form-item>

          <el-form-item :label="lang.password">
            <el-input type="password" v-bind:disabled="adminSettings.hadSeted" placeholder=""
                      v-model="adminSettings.password"></el-input>
          </el-form-item>

          <el-form-item :label="lang.enter_again">
            <el-input type="password" v-bind:disabled="adminSettings.hadSeted" placeholder=""
                      v-model="adminSettings.password2"></el-input>
          </el-form-item>
        </el-form>
      </div>
    </div>


    <div v-if="active === 3" class="setup__card">
      <div class="setup__desc">
        <h2>{{ lang.SetDomail }}</h2>
      </div>
      <div class="setup__form" style="width: 400px;">
        <el-form label-width="120px">

          <el-form-item :label="lang.smtp_domain">
            <el-input placeholder="domaim.com" v-model="domainSettings.smtp_domain">
              <template #prepend>smtp.</template>
            </el-input>
          </el-form-item>

          <el-form-item :label="lang.web_domain">
            <el-input placeholder="pmail.domain.com" v-model="domainSettings.web_domain"></el-input>
          </el-form-item>

          <el-form-item :label="lang.multi_domain_setting">
                        <span>{{ lang.multi_domain_setting_desc }}
                          <el-button @click="addDomain" size="small"
                                     type="success" :icon="Plus"
                                     circle>
                          </el-button>
                        </span>
            <el-input :placeholder="'domain' + i + '.com'" v-for="(item, i) in domainSettings.multi_domain "
                      v-model="domainSettings.multi_domain[i]" :key="item"></el-input>
          </el-form-item>


        </el-form>
      </div>
    </div>

    <div v-if="active === 4" class="setup__card--stack">

      <div class="setup__desc">
        <h2>{{ lang.setDNS }}</h2>
        <div style="margin-top: 10px;">{{ lang.dns_desc }}</div>
      </div>
      <div class="setup__form" width="600px" v-for="(info,domain) in dnsInfos" :key="info">
        <h3>{{ domain }}</h3>
        <el-table :data="info" border style="width: 100%">
          <el-table-column prop="host" :label="lang.hostname" width="110px">
            <template #default="scope">
              <div style="display: flex; align-items: center">
                <el-tooltip :content="lang.dns_root_desc" placement="top"
                            v-if="scope.row.host === '' || scope.row.host === '@' ">
                  {{ scope.row.host }}
                </el-tooltip>
                <span v-else>{{ scope.row.host }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="type" :label="lang.record_type" width="110px"/>
          <el-table-column prop="value" :label="lang.record_value">
            <template #default="scope">
              <div style="display: flex; align-items: center">
                <el-tooltip :content="scope.row.tips" placement="top" v-if="scope.row.tips !== ''">
                  {{ scope.row.value }}
                </el-tooltip>
                <span v-else>{{ scope.row.value }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="ttl" label="TTL" width="110px"/>
        </el-table>
      </div>
    </div>

    <el-alert :closable="false" :title="lang.warning_title" type="error" center
              v-if="active === 5 && sslSettings.type === '0' && port !== 80" :description="lang.autoSSLWarn"/>

    <div v-if="active === 5" class="setup__card">
      <div class="setup__desc">
        <h2>{{ lang.setSSL }}</h2>
        <div style="margin-top: 10px;">{{ lang.setSSL }}</div>
      </div>
      <!-- 修复: width HTML 属性无效，改为 style -->
      <!-- 修改日期: 20260504 -->
      <div class="setup__form" style="width: 600px;">
        <el-form label-width="120px">
          <el-form-item :label="lang.type">
            <el-select :placeholder="lang.ssl_auto" v-model="sslSettings.type" :disabled="dnsChecking">
              <el-option :label="lang.ssl_auto" value="0"/>
              <el-option :label="lang.ssl_manuallyf" value="1"/>
            </el-select>
          </el-form-item>

          <!-- 验证方式与帮助图标用 flex 容器包裹，防止 el-select 撑满导致 "?" 换行 -->
          <!-- 修改日期: 20260504，修复自动设置选项显示异常 -->
          <el-form-item :label="lang.ssl_challenge_type" v-if="sslSettings.type === '0'">
            <div style="display: flex; align-items: center;">
              <el-select :placeholder="lang.ssl_auto_http" v-model="sslSettings.challenge"
                         :disabled="dnsChecking" style="flex: 1;">
                <el-option :label="lang.ssl_auto_http" value="http"/>
                <el-option :label="lang.ssl_auto_dns" value="dns"/>
              </el-select>

              <el-tooltip class="box-item" effect="dark" :content="lang.challenge_typ_desc"
                          placement="top-start">
                <span style="margin-left: 6px; font-size:18px; font-weight: bolder;">?</span>
              </el-tooltip>
            </div>
          </el-form-item>


          <el-form-item :label="lang.ssl_key_path" v-if="sslSettings.type === '1'">
            <el-input placeholder="./config/ssl/private.key" v-model="sslSettings.key_path"></el-input>
          </el-form-item>

          <el-form-item :label="lang.ssl_crt_path" v-if="sslSettings.type === '1'">
            <el-input placeholder="./config/ssl/public.crt" v-model="sslSettings.crt_path"></el-input>
          </el-form-item>
        </el-form>


      </div>

    </div>

    <div v-if="dnsChecking">
      <label>{{ lang.dns_desc }}</label>
      <el-table :data="sslSettings.paramsList" border v-loading="sslSettings.paramsList.length === 0">
        <el-table-column prop="host" :label="lang.hostname" width="110px"/>
        <el-table-column prop="type" :label="lang.record_type" width="110px"/>
        <el-table-column prop="value" :label="lang.record_value">
          <template #default="scope">
            <div style="display: flex; align-items: center">
              <el-tooltip :content="scope.row.tips" placement="top" v-if="scope.row.tips !== ''">
                {{ scope.row.value }}
              </el-tooltip>
              <span v-else>{{ scope.row.value }}</span>
            </div>
          </template>

        </el-table-column>
        <el-table-column prop="ttl" label="TTL" width="110px"/>
      </el-table>

    </div>


    <el-button :element-loading-text="waitDesc" v-loading.fullscreen.lock="fullscreenLoading" class="setup__next-btn"
               style="margin-top: 12px" @click="next">{{
        lang.next
      }}
    </el-button>

  </div>
</template>

<script setup lang="ts">
import {reactive, ref} from 'vue'
import {ElMessage} from 'element-plus'
import lang from '../i18n/i18n';
import axios from 'axios'
import {Plus} from '@element-plus/icons-vue'
import {setupService} from "@/services/setupService";
import {useRoute} from 'vue-router'

// 从 URL 读取 Setup Token，用于接口鉴权。
// 服务端启动时生成随机 Token，通过 URL query param 传递给前端。
// 修改日期: 20260425
// 修改日期: 20260504，增加 fallback：从 window.location.search 解析 token。
// 原因：前端使用 Vue Router hash 模式，route.query 仅读取 # 之后的参数。
// 若用户访问旧格式 URL（http://host/?token=xxx），token 在 # 之前，
// route.query.token 为空，需从 window.location.search 兜底读取。
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

const setupToken = ref<string>(
  (route.query.token as string) || getQueryParamFromSearch('token') || ''
)

const waitDesc = ref(lang.wait_desc);

const adminSettings = reactive({
  "account": "admin",
  "password": "",
  "password2": "",
  "hadSeted": false
})

const dbSettings = reactive({
  "type": "sqlite",
  "dsn": "./config/pmail.db",
  "lable": ""
})

const domainSettings = reactive({
  "web_domain": "",
  "smtp_domain": "",
  "multi_domain": [] as string[]
})

const sslSettings = reactive({
  "type": "0",
  "challenge": "http",
  "key_path": "./config/ssl/private.key",
  "crt_path": "./config/ssl/public.crt",
  "paramsList": [] as any[],
})


const active = ref(0)
const fullscreenLoading = ref(false)
const dnsChecking = ref(false)

const dnsInfos = ref<Record<string, any>>({})

const port = ref(80)


const addDomain = () => {
  domainSettings.multi_domain.push("")
}

const setPassword = () => {
  if (adminSettings.hadSeted) {
    active.value++;
    getDomainConfig();
    return;
  }

  if (adminSettings.password !== adminSettings.password2) {
    ElMessage.error(lang.err_pwd_diff)
  } else {
    /** 通过 setupService 保存管理员密码 */
    setupService.setPassword(setupToken.value, adminSettings.account, adminSettings.password).then((res: any) => {
      if (res.errorNo !== 0) {
        ElMessage.error(res.errorMsg)
      } else {
        active.value++;
        getDomainConfig();
      }
    })
  }
}

const getPassword = () => {
  /** 通过 setupService 获取密码配置 */
  setupService.getPasswordConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      adminSettings.hadSeted = res.data !== ""
      if (adminSettings.hadSeted) {
        adminSettings.account = res.data
        adminSettings.password = "*******"
        adminSettings.password2 = "*******"
      }

    }
  })
}


const getDbConfig = () => {
  /** 通过 setupService 获取数据库配置 */
  setupService.getDatabaseConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      dbSettings.type = res.data.db_type;
      dbSettings.dsn = res.data.db_dsn;
    }
  })
}

const getDomainConfig = () => {
  /** 通过 setupService 获取域名配置 */
  setupService.getDomainConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      domainSettings.web_domain = res.data.web_domain;
      domainSettings.smtp_domain = res.data.smtp_domain;
      domainSettings.multi_domain = res.data.domains;
    }
  })
}

const setDbConfig = () => {
  // 切换数据库类型为sqlite时，数据库路径为空，则使用默认路径
  if (dbSettings.type === "sqlite" && !dbSettings.dsn) dbSettings.dsn = "./config/pmail.db";
  else if (!dbSettings.dsn) ElMessage({
    message: lang.err_db_dsn_empty,
    type: "error",
  });
  /** 通过 setupService 保存数据库配置 */
  setupService.setDatabaseConfig(setupToken.value, dbSettings.type, dbSettings.dsn).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      active.value++;
      getPassword();
    }
  })
}

/** 通过 setupService 获取 DNS 配置 */
const getDNSConfig = () => {
  setupService.getDnsConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      dnsInfos.value = res.data
    }
  })
}


/** 通过 setupService 获取 SSL 配置 */
const getSSLConfig = () => {
  setupService.getSslConfig(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      sslSettings.type = res.data.type
      if (String(sslSettings.type) === "2") {
        sslSettings.type = "0"
        sslSettings.challenge = "dns"
      }

      port.value = res.data.port
    }
  })
}


const setSSLConfig = () => {
  fullscreenLoading.value = true;

  let sslType = sslSettings.type;
  if (sslType === "0" && sslSettings.challenge === "dns") {
    sslType = "2"
  }


  /** 通过 setupService 保存 SSL 配置 */
  setupService.setSslConfig(setupToken.value, sslType, sslSettings.key_path, sslSettings.crt_path).then((res: any) => {
    if (res.errorNo !== 0) {
      fullscreenLoading.value = false;
      ElMessage.error(res.errorMsg)
    } else {
      if (Number(sslType) === 2) {
        fullscreenLoading.value = false;
        dnsChecking.value = true;
        getSSLDNSParams();
      }
      checkStatus();
    }
  })
}


const checkStatus = () => {
  axios.post("/api/ping", {}).then((res) => {
    if (res.data.errorNo !== 0) {
      setTimeout(function () {
        checkStatus()
      }, 1000);
    } else {
      if (Number(sslSettings.type) === 1) {
        window.location.href = "http://" + domainSettings.web_domain;
      } else {
        window.location.href = "https://" + domainSettings.web_domain;
      }
    }
  }).catch(() => {
    setTimeout(function () {
      checkStatus()
    }, 1000);
  })
}


/** 通过 setupService 保存域名配置 */
const setDomainConfig = () => {
  setupService.setDomainConfig(setupToken.value, domainSettings.web_domain, domainSettings.smtp_domain, domainSettings.multi_domain.join(",")).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      active.value++;
      getDNSConfig();
    }
  })
}

/** 通过 setupService 获取 SSL DNS 验证参数 */
const getSSLDNSParams = () => {
  setupService.getSslDnsParams(setupToken.value).then((res: any) => {
    if (res.errorNo !== 0) {
      ElMessage.error(res.errorMsg)
    } else {
      sslSettings.paramsList = res.data
    }
  })

  if (sslSettings.paramsList.length === 0) {
    setTimeout(function () {
      getSSLDNSParams()
    }, 1000);
  }


}


const next = () => {
  switch (active.value) {
    case 0:
      active.value++
      getDbConfig();
      break
    case 1:
      setDbConfig();
      break;
    case 2:
      setPassword();
      break;
    case 3:
      setDomainConfig();
      break;
    case 4:
      getSSLConfig();
      active.value++
      break
    case 5:
      if (dnsChecking.value) {
        fullscreenLoading.value = true;
        waitDesc.value = lang.dns_challenge_wait;
      } else {
        setSSLConfig();
      }
      break
  }

}
</script>


<!-- 样式: Docusaurus BEM 风格 | 重构日期: 20260429 -->
<style scoped>
/* 安装向导页面 */
.setup {
  width: 100%;
  height: 100%;
  background: var(--ifm-background-color);
  display: flex;
  flex-direction: column;
  gap: var(--ifm-spacing-md);
  padding: var(--ifm-spacing-xl);
  overflow-y: auto;
}

.setup__desc {
  padding-right: var(--ifm-spacing-lg);
  margin-bottom: var(--ifm-spacing-md);
}

.setup__desc h2 {
  font-size: 24px;
  font-weight: 700;
  color: var(--ifm-color-content);
}

/* 步骤条容器 */
.setup__steps {
  padding: var(--ifm-spacing-md) var(--ifm-spacing-lg);
  border-radius: var(--ifm-card-border-radius);
  background: var(--ifm-background-surface-color);
  border: 1px solid var(--ifm-border-color);
  box-shadow: var(--ifm-global-shadow-lw);
}

/* 内容卡片：横向布局 */
.setup__card {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  gap: var(--ifm-spacing-lg);
  padding: var(--ifm-spacing-lg);
  border-radius: var(--ifm-card-border-radius);
  background: var(--ifm-background-surface-color);
  border: 1px solid var(--ifm-border-color);
  box-shadow: var(--ifm-global-shadow-lw);
}

/* 内容卡片：纵向堆叠 */
.setup__card--stack {
  display: flex;
  flex-direction: column;
  gap: var(--ifm-spacing-md);
  padding: var(--ifm-spacing-lg);
  border-radius: var(--ifm-card-border-radius);
  background: var(--ifm-background-surface-color);
  border: 1px solid var(--ifm-border-color);
  box-shadow: var(--ifm-global-shadow-lw);
}

/* 下一步按钮 */
.setup__next-btn {
  align-self: flex-end;
  border-radius: var(--ifm-global-radius);
  padding: 10px 22px;
  font-weight: 600;
}
</style>
