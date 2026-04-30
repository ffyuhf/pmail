# PMail

> 拷贝项目地址：[https://github.com/jinnrry/pmail/](https://github.com/jinnrry/pmail/)

> 一台服务器、一个域名、一行代码、一分钟时间，你就能够搭建出一个自己的域名邮箱。

PMail是一个追求极简部署流程、极致资源占用的个人域名邮箱服务器。单文件运行，包含完整的收发邮件服务和Web端邮件管理功能。只需一台服务器、一个域名、一行代码、一分钟部署时间，你就能够搭建出自己的域名邮箱。

欢迎各类PR，无论你是修复bug、新增功能、修改翻译。

## 为什么拷贝这个项目

jinnrry作者更的太慢了，功能对我来说太少了

## 项目优势

### 1、部署简单

使用Go语言编写，支持跨平台，编译后单文件运行，单文件包含完整的前后端代码。修改配置文件，运行即可。

### 2、资源占用极小

编译后二进制文件仅18MB，运行过程中占用内存28M以内。

### 3、安全方面

支持DKIM、SPF校验。正确配置的情况下，Email Test得分10分。

> ###### 正确配置包括但不限于：一个正常的顶级域名后戳，一个有固定的IP地址（支持rDNS）的服务器

**认证与密码安全：**

- 密码采用 bcrypt （10）安全哈希存储，替代传统弱哈希算法，支持渐进式自动迁移
- 全协议暴力破解防护（HTTP/SMTP/IMAP/POP3），基于IP和账户的指数退避频率限制
- API Token 采用随机不透明令牌（Opaque Token），支持撤销和客户端IP绑定

**传输与通信安全：**
- SMTP 465/587 端口强制 TLS 认证，拒绝不安全连接，符合 RFC 6409 规范
- 自动 SSL 证书：实现 ACME 协议，程序将自动获取并更新 Let's Encrypt 证书
- HTTP 安全响应头（X-Content-Type-Options、X-Frame-Options、Strict-Transport-Security、X-XSS-Protection）

**Web 应用安全：**
- CSRF 防护：SameSite Cookie + Origin/Referer 双层校验
- IDOR 越权访问防护：邮件详情查询强制验证用户归属关系
- Setup 初始化接口安全加固：Token 验证、DSN 脱敏、路径遍历防护、HTTP 方法限制

**数据与文件安全：**
- DKIM 使用 2048 位 RSA 密钥签名，满足现代安全标准
- 配置文件、SSL 私钥、DKIM 密钥等敏感文件采用最小权限（0600/0400/0644）

### 4、HTML 邮件兼容性

修复 `multipart/related` 双重遍历 Bug，HTML 邮件不再显示为纯文本。扩展 XSS 安全策略，支持 HTML5 语义元素（section/article/header 等），同时仍阻止 script/iframe/object 等危险标签。修复大 HTML 邮件因 SMTP 超时被中断的问题（超时从 10s 增加到 60s），转发邮件自动添加正确的编码头。修复邮件标题和发件人名称中 HTML 实体编码（如 `&#39;`）显示异常的问题。

### 5、前端体验

- 邮件搜索：侧边栏搜索框支持关键词搜索，回车触发、一键清空
- 批量操作：列表全选/取消全选（含半选状态）、每页条数选择器（15/25/50/100）
- 编辑器全屏：全屏模式覆盖整个写信界面（收件人、主题、附件等），而非仅编辑区
- 视觉统一：采用 Docusaurus BEM 风格 CSS 变量体系，暗色模式下颜色一致
- 布局自适应：列表行宽度按屏幕自适应，窄屏自动隐藏摘要列

### 6、自动SSL证书

实现了ACME协议，程序将自动获取并更新Let's Encrypt证书。

默认情况下，会为web后台也生成ssl证书，让后台使用https访问，如果你有自己的网关层，不需要https的话，在配置文件中将 `httpsEnabled`
设置为 `2`，这样管理后台就不会使用https协议。（ 注意：即使你不需要https，也请保证ssl证书文件路径正确，http协议虽然不使用证书了，但是smtp协议还需要证书）

### 7、邮件客户端支持

只要支持pop3、smtp、imap协议的邮件客户端均可使用

### 8、多域名、多用户支持

支持多域名、多用户且完整支持收发邮件

# 如何部署

## 0、检查IP、域名

先去[spamhaus](https://check.spamhaus.org/)检查你的域名和服务器IP是否有屏蔽记录

## 1、下载文件

* [点击这里](https://github.com/ffyuhf/pmail/releases)下载一个与你匹配的程序文件（支持 Linux amd64/arm64、Windows、macOS 等多平台）。
* 或者使用Docker运行 `docker pull ghcr.io/ffyuhf/pmail:latest`

## 2、运行

`./pmail -p 80` 

> `-p 指定引导设置界面的http端口，默认为80端口，注意该参数仅影响引导设置阶段，设置完成后如果需要修改端口请修改配置文件`

> [!IMPORTANT]
> 如果引导设置阶段使用非80端口，将无法自动设置SSL证书

或者

`docker run -p 25:25 -p 80:80 -p 443:443 -p 110:110 -p 465:465 -p 587:587 -p 995:995 -p 993:993 -v $(pwd)/config:/work/config ghcr.io/ffyuhf/pmail:latest`

> [!IMPORTANT]
> 如果你服务器开启了防火墙，你需要打开25、80、110、443、465、587、995、993端口

## 3、配置

为安全考虑，程序启动后会在日志中打印 Setup URL（含 Token 参数），例如：

```
Please click http://YOUR_IP/?token=xxxxxxxx to continue.
```

- **直接运行**：查看终端输出的启动日志，点击或复制其中的 URL 到浏览器打开
- **Docker 运行**：通过 `docker logs <容器ID>` 查看启动日志获取 URL

> [!IMPORTANT]
> Setup Token 为一次性鉴权凭据，用于保护初始化接口安全。必须通过启动日志中带 `?token=xxx` 参数的 URL 访问配置页面，直接访问 `http://IP` 将被拒绝。

## 4、邮箱得分测试

建议找一下邮箱测试服务(比如[https://www.mail-tester.com/](https://www.mail-tester.com/))进行邮件得分检测，避免自己某些步骤漏配，导致发件进对方垃圾箱。

# 配置文件说明

配置文件位于 `config/config.json`，首次运行时自动生成。通常通过 Setup 向导自动配置，也可手动编辑。

```jsonc
{
  // ===== 基础设置 =====
  "logLevel": "info",              // 日志级别：debug / info / warn / error
  "domain": "domain.com",          // 主域名
  "domains": ["domain.com"],       // 所有收信域名（多域名时填写全部）
  "webDomain": "mail.domain.com",  // Web 管理后台域名

  // ===== SSL 证书 =====
  "sslType": "0",                  // 证书模式：0=自动(HTTP挑战) / 1=手动上传 / 2=自动(DNS挑战)
  "SSLPrivateKeyPath": "config/ssl/private.key",  // SSL 私钥路径
  "SSLPublicKeyPath": "config/ssl/public.crt",    // SSL 证书路径

  // ===== DKIM 签名 =====
  "dkimPrivateKeyPath": "config/dkim/dkim.priv",  // DKIM 私钥路径（2048位，自动生成）

  // ===== 数据库 =====
  "dbType": "sqlite",              // 数据库类型：sqlite / mysql / postgres
  "dbDSN": "./config/pmail.db",    // SQLite: 文件路径 | MySQL: user:pass@tcp(host:port)/dbname?parseTime=True&loc=Local

  // ===== Web 服务 =====
  "httpsEnabled": 0,               // HTTPS 开关：0=默认启用 / 1=启用 / 2=不启用（仅 HTTP）
  "httpPort": 80,                  // HTTP 端口，默认 80
  "httpsPort": 443,                // HTTPS 端口，默认 443

  // ===== 垃圾邮件过滤「且无有效收件人时过滤」 =====
  "spamFilterLevel": 0,            // 0=不过滤 / 1=SPF+DKIM 均失败时过滤 / 2=SPF 不通过时过滤 / 3=DKIM 不通过时过滤

  // ===== 消息推送（可选/此处不用可删） =====
  "weChatPushAppId": "",           // 微信推送 AppID
  "weChatPushSecret": "",          // 微信推送 Secret
  "weChatPushTemplateId": "",      // 微信推送模板 ID
  "weChatPushUserId": "",          // 微信推送用户 ID
  "tgBotToken": "",                // Telegram Bot Token
  "tgChatId": "",                  // Telegram Chat ID
  "webPushUrl": "",                // 自定义 Web Push URL
  "webPushToken": "",              // 自定义 Web Push Token

  // ===== 系统状态 =====
  "isInit": true                   // false 时进入安装引导流程，Setup 完成后自动设为 true
}
```

### 数据库 DSN 格式参考

| 数据库类型 | dbDSN 示例 |
|-----------|-----------|
| SQLite | `./config/pmail.db` |
| MySQL | `root:password@tcp(127.0.0.1:3306)/pmail?parseTime=True&loc=Local` |
| PostgreSQL | `postgres://user:password@localhost:5432/pmail?sslmode=disable` |

# 第三方邮件客户端配置

POP3地址： pop.[你的域名]

POP3端口： 110/995(SSL)

SMTP地址： smtp.[你的域名]

SMTP端口： 25/465、587(SSL)

IMAP地址： imap.[Your Domain]

IMAP端口： 993(SSL)

# 插件

[微信推送](server/hooks/wechat_push/README.md)

[垃圾邮件屏蔽](server/hooks/spam_block/README.md)

# 其他优秀开源插件

[Telegram推送](https://github.com/ydzydzydz/pmail_telegram_push)

## 插件安装
> [!IMPORTANT]
> 插件以独立进程的方式运行在你的服务器上，请自行审查第三方插件的安全性。PMail目前仅维护上述三款插件

将插件二进制文件放到`plugins`文件夹中即可

# 参与开发



## ⚠️ 不兼容变更「对于原版」

### 1. API Token 格式变更

- **旧格式**：`{account}:md5(hash+timestamp):timestamp` — 已不再支持
- **新格式**：64 字符随机 hex 字符串
- **迁移方式**：通过 `/api/token/generate` 端点重新生成
- **影响范围**：仅影响通过 HTTP `Token` Header 调用 API 的外部集成，Web 端 Session 不受影响

### 2. 数据库 Schema 变更

- `user` 表 `password` 字段：`char(32)` → `varchar(72)`（xorm Sync2 自动迁移）
- 新增 `api_tokens` 表（xorm Sync2 自动创建）

### 3. Setup 访问方式变更

- Setup 页面必须通过启动日志中带 `?token=xxx` 参数的 URL 访问
- Setup 接口仅接受 POST 请求

## 升级注意事项

### 必须操作

| 操作 | 说明 |
|------|------|
| **更新 DNS DKIM 记录** | DKIM 密钥已从 1024 位升级为 2048 位，首次启动自动重新生成，需将新公钥更新到 DNS TXT 记录 |
| **重新生成 API Token** | 旧格式 Token 已失效，需通过 `/api/token/generate` 重新生成 |
| **记录 Setup Token** | 首次部署时 Setup URL 携带一次性 Token，从启动日志获取 |

### 自动处理（无需手动操作）

| 项目 | 说明 |
|------|------|
| 密码迁移 | 用户登录时自动从 MD5 升级为 bcrypt |
| 数据库表结构 | xorm Sync2 自动处理字段类型变更和新表创建 |
| Cookie 属性 | Session Cookie 自动添加 SameSite=Lax |

### 建议操作

| 操作 | 说明 |
|------|------|
| 手动修复文件权限 | 对已有配置文件执行 `chmod 600`、SSL 私钥执行 `chmod 400` |
| 验证邮件客户端 | 确认使用的邮件客户端支持 STARTTLS（现代客户端均支持） |

---

## 项目架构

1、前端： vue3+element-plus

前端代码位于 `fe`目录中，运行参考 `fe`目录中的README文件

2、后端： golang + MySQL/SQLite

后端代码进入 `server`文件夹，运行 `main.go`文件

3、编译项目

`make build`

4、单元测试

`make test`

## 后端接口文档

[go to wiki](https://github.com/Jinnrry/PMail/wiki/%E5%90%8E%E7%AB%AF%E6%8E%A5%E5%8F%A3%E6%96%87%E6%A1%A3)

## 插件开发

[go to wiki](https://github.com/Jinnrry/PMail/wiki/%E6%8F%92%E4%BB%B6%E5%BC%80%E5%8F%91%E8%AF%B4%E6%98%8E)
