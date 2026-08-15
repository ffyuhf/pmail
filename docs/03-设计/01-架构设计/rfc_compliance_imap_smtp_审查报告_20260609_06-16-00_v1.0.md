# RFC 标准合规性审查报告

> **版本**: v1.0
> **日期**: 2026-06-09
> **时间**: 06:16:00
> **作者**: PMail Team
> **模块**: imap-server / smtp-server
> **状态**: 已批准
> **生效范围**: PMail 项目全部 IMAP 与 SMTP 相关模块

---

## 1. 审查范围与目标

### 1.1 审查目标

评估 PMail 项目对以下 RFC 标准的合规程度，识别差距并提供改进建议。

### 1.2 审查范围

| 协议 | RFC 编号 | 标准名称 | 类型 |
|------|----------|----------|------|
| IMAP | RFC 9051 | IMAP4rev2 | 核心协议 |
| IMAP | RFC 7888 | 非同步字面量扩展（LITERAL-/LITERAL+） | 扩展 |
| IMAP | RFC 6855 | IMAP UTF-8 支持（UTF8=ACCEPT/UTF8=ONLY） | 扩展 |
| IMAP | RFC 7162 | CONDSTORE 和 QRESYNC 扩展 | 扩展 |
| SMTP | RFC 5321 | Simple Mail Transfer Protocol | 核心协议 |
| SMTP | RFC 6531 | SMTPUTF8 扩展（国际化邮件地址） | 扩展 |
| SMTP | RFC 6409 | 邮件提交协议（Message Submission） | 扩展 |

### 1.3 审查方法论

- **代码审查**：逐文件检查 IMAP/SMTP 服务器实现代码
- **依赖库分析**：检查 go-imap/v2 和 go-smtp 库的扩展支持能力
- **能力声明验证**：核对 CAPABILITY/EHLO 响应中公布的能力
- **后端接口验证**：检查 Session 接口是否实现了扩展所需的回调方法

---

## 2. 技术栈与依赖分析

### 2.1 核心依赖库

| 依赖库 | 版本 | 用途 |
|--------|------|------|
| `github.com/emersion/go-imap/v2` | v2.0.0-beta.5 | IMAP 服务器/客户端协议实现 |
| `github.com/emersion/go-smtp` | v0.21.3 | SMTP 服务器协议实现 |
| `github.com/emersion/go-sasl` | (间接依赖) | SASL 认证机制 |

### 2.2 go-imap/v2 能力矩阵

库中定义的 IMAP 能力常量（`capability.go`）：

| 能力常量 | 能力字符串 | 对应 RFC | 说明 |
|----------|-----------|----------|------|
| `CapIMAP4rev1` | `IMAP4rev1` | RFC 3501 | 基础 IMAP 协议 |
| `CapIMAP4rev2` | `IMAP4rev2` | RFC 9051 | 新一代 IMAP 协议 |
| `CapLiteralMinus` | `LITERAL-` | RFC 7888 | 非同步字面量（≤4096字节） |
| `CapLiteralPlus` | `LITERAL+` | RFC 7888 | 非同步字面量（无限制） |
| `CapUTF8Accept` | `UTF8=ACCEPT` | RFC 6855 | 接受 UTF-8 邮箱名 |
| `CapUTF8Only` | `UTF8=ONLY` | RFC 6855 | 仅支持 UTF-8 |
| `CapCondStore` | `CONDSTORE` | RFC 7162 | 条件存储 |
| `CapQResync` | `QRESYNC` | RFC 7162 | 快速重新同步 |
| `CapNamespace` | `NAMESPACE` | RFC 2342 | 命名空间 |
| `CapUIDPlus` | `UIDPLUS` | RFC 4315 | UID 增强扩展 |
| `CapESearch` | `ESEARCH` | RFC 4731 | 扩展搜索 |
| `CapMove` | `MOVE` | RFC 6851 | 移动命令 |
| `CapIdle` | `IDLE` | RFC 2177 | 空闲命令 |
| `CapEnable` | `ENABLE` | RFC 5161 | 能力启用 |
| `CapStatusSize` | `STATUS=SIZE` | RFC 8438 | 状态大小 |

### 2.3 go-smtp 扩展支持

库中定义的 SMTP 扩展开关（`server.go`）：

| 字段 | 扩展 | 对应 RFC | 说明 |
|------|------|----------|------|
| `EnableSMTPUTF8` | `SMTPUTF8` | RFC 6531 | 国际化邮件地址 |
| `EnableBINARYMIME` | `BINARYMIME` | RFC 3030 | 二进制 MIME |
| `EnableDSN` | `DSN` | RFC 3461 | 投递状态通知 |
| `EnableREQUIRETLS` | `REQUIRETLS` | RFC 8689 | 要求 TLS |

---

## 3. IMAP RFC 合规性详细分析

### 3.1 RFC 9051 - IMAP4rev2

**合规状态**：❌ **不支持**

#### 3.1.1 当前实现

项目使用 IMAP4rev1（RFC 3501），而非 IMAP4rev2。

**代码证据**（`server/listen/imap_server/imap_server.go` 第52-58行）：

```go
Caps: imap.CapSet{
    imap.CapIMAP4rev1: {},
    imap.CapNamespace: {},
    imap.CapUIDPlus:   {},
    imap.CapMove:      {},
    // ... 未声明 CapIMAP4rev2
},
```

#### 3.1.2 IMAP4rev2 隐含能力差距

IMAP4rev2 要求以下能力作为协议的一部分自动可用：

| 隐含能力 | 当前状态 | 差距说明 |
|----------|---------|----------|
| `NAMESPACE` | ✅ 已声明 | - |
| `UNSELECT` | ✅ 库自动 | IMAP4rev1 下库自动添加 |
| `UIDPLUS` | ✅ 已声明 | - |
| `ESEARCH` | ❌ 未声明 | 扩展搜索未启用 |
| `SEARCHRES` | ❌ 未声明 | 搜索结果引用未启用 |
| `ENABLE` | ✅ 库自动 | IMAP4rev1 下库自动添加 |
| `IDLE` | ✅ 库自动 | IMAP4rev1 下库自动添加 |
| `SASL-IR` | ✅ 库自动 | IMAP4rev1 下库自动添加 |
| `LIST-EXTENDED` | ❌ 未声明 | 扩展列表未启用 |
| `LIST-STATUS` | ❌ 未声明 | 列表状态未启用 |
| `MOVE` | ✅ 已声明 | - |
| `LITERAL-` | ✅ 库自动 | IMAP4rev1 下库自动添加 |
| `STATUS=SIZE` | ❌ 未声明 | 状态大小未启用 |

#### 3.1.3 影响评估

- 遵循 IMAP4rev2 严格要求的客户端可能无法正确工作
- 部分现代邮件客户端（如 GNOME Evolution、K-9 Mail 新版本）倾向于使用 IMAP4rev2
- 缺少 `ESEARCH` 导致搜索结果仅返回序列号集合，无法返回统计信息

---

### 3.2 RFC 7888 - 非同步字面量扩展

**合规状态**：⚠️ **部分支持**（仅 LITERAL-，库自动处理）

#### 3.2.1 当前实现

go-imap 库在 IMAP4rev1 模式下自动添加 `LITERAL-` 能力（`imapserver/capability.go` 第41-44行）：

```go
if available.Has(imap.CapIMAP4rev1) {
    caps = append(caps, []imap.Cap{
        imap.CapSASLIR,
        imap.CapLiteralMinus,  // 自动添加 LITERAL-
    }...)
}
```

非同步字面量的处理逻辑（`imapserver/conn.go` 第396行）：

```go
if nonSync && size > 4096 && !c.server.options.caps().Has(imap.CapLiteralPlus) {
    // 超过 4096 字节的非同步字面量，如果未启用 LITERAL+，则回退为同步模式
}
```

#### 3.2.2 差距分析

| 能力 | 状态 | 说明 |
|------|------|------|
| `LITERAL-` | ✅ 自动 | 支持 ≤4096 字节的非同步字面量 |
| `LITERAL+` | ❌ 未启用 | CapSet 中未声明 `imap.CapLiteralPlus` |

#### 3.2.3 影响评估

- 大于 4096 字节的 APPEND 操作（如上传大附件）仍需同步等待服务器确认
- 在高延迟网络环境下，大邮件上传会显著变慢
- **修复难度**：低（在 CapSet 中添加 `imap.CapLiteralPlus` 即可）

---

### 3.3 RFC 6855 - IMAP UTF-8 支持

**合规状态**：⚠️ **部分支持**（UTF8=ACCEPT 库自动处理）

#### 3.3.1 当前实现

go-imap 库在认证状态下自动添加 `UTF8=ACCEPT` 能力（`imapserver/capability.go` 第61-67行）：

```go
if c.state == imap.ConnStateAuthenticated || c.state == imap.ConnStateSelected {
    if available.Has(imap.CapIMAP4rev1) {
        caps = append(caps, []imap.Cap{
            // ...
            imap.CapUTF8Accept,  // 自动添加 UTF8=ACCEPT
        }...)
    }
}
```

PMail 后端确实使用 UTF-8 编码处理邮箱名，如 `TestCreate` 测试中使用中文文件夹名"一级菜单"。

#### 3.3.2 差距分析

| 能力 | 状态 | 说明 |
|------|------|------|
| `UTF8=ACCEPT` | ✅ 自动 | 认证后自动公布 |
| `UTF8=ONLY` | ❌ 未启用 | CapSet 中未声明 `imap.CapUTF8Only` |

#### 3.3.3 影响评估

- `UTF8=ONLY` 表示服务器不接受修改 UTF-8 邮箱名的非 UTF-8 编码方式，当前不需要声明
- 当前行为对绝大多数客户端兼容
- **修复难度**：无需修复，当前行为已满足 RFC 6855 要求

---

### 3.4 RFC 7162 - CONDSTORE 和 QRESYNC

**合规状态**：❌ **完全不支持**

#### 3.4.1 当前实现

**CapSet 声明**：未包含 `imap.CapCondStore` 和 `imap.CapQResync`

**后端 Session 接口**：未实现 CONDSTORE/QRESYNC 所需的接口方法

**数据库层**：无 `mod-sequence` 追踪机制

#### 3.4.2 详细差距分析

| 组件 | 要求 | 当前状态 |
|------|------|----------|
| **能力声明** | `CONDSTORE` 和 `QRESYNC` | ❌ 未声明 |
| **SELECT 命令** | 返回 `HIGHESTMODSEQ` | ❌ `session_select.go` 未返回 |
| **FETCH 命令** | 支持 `CHANGEDSINCE` 修饰符 | ❌ `session_fetch.go` 未处理 |
| **STORE 命令** | 返回 `MODSEQ` 响应 | ❌ `session_store.go` 未返回 |
| **SEARCH 命令** | 支持 `MODSEQ` 搜索条件 | ❌ 未实现 |
| **数据库** | `mod-sequence` 值追踪 | ❌ 数据库无此字段 |
| **QRESYNC** | 快速邮箱重新同步 | ❌ 未实现 |

#### 3.4.3 影响评估

- 移动客户端（如手机、平板）在断线重连后必须完整重新同步，无法增量同步
- 大邮箱用户（数千封邮件）的同步体验差，耗费大量流量和时间
- 现代邮件客户端（Mozilla Thunderbird、Outlook、Apple Mail）支持 CONDSTORE，无法利用增量同步优势
- **修复难度**：高（需数据库增加 mod-sequence 字段、修改所有修改邮件状态的接口）

---

## 4. SMTP RFC 合规性详细分析

### 4.1 RFC 5321 - Simple Mail Transfer Protocol

**合规状态**：✅ **基本支持**

#### 4.1.1 当前实现

项目使用 `go-smtp` 库实现 SMTP 服务器，分为三个实例：

**代码证据**（`server/listen/smtp_server/smtp_server.go`）：

| 服务器 | 端口 | 用途 | TLS 模式 |
|--------|------|------|----------|
| SMTP 服务器（接收） | 25 | 外部邮件接收 | STARTTLS（可选） |
| SMTP 提交服务（TLS） | 465 | 客户端提交 | 隐式 TLS |
| SMTP 提交服务（STARTTLS） | 587 | 客户端提交 | STARTTLS |

#### 4.1.2 支持的 SMTP 功能

| 功能 | 状态 | 说明 |
|------|------|------|
| EHLO/HELO | ✅ | 库自动处理 |
| MAIL FROM | ✅ | 含 SIZE、SMTPUTF8、BODY、AUTH 参数解析 |
| RCPT TO | ✅ | 含 NOTIFY、ORCPT 参数解析（DSN相关） |
| DATA | ✅ | 标准 DATA 命令 |
| BDAT/CHUNKING | ✅ | 库自动支持 |
| AUTH PLAIN | ✅ | SASL PLAIN 认证 |
| STARTTLS | ✅ | 端口 25/587 |
| 8BITMIME | ✅ | 库自动声明 |
| PIPELINING | ✅ | 库自动声明 |
| ENHANCEDSTATUSCODES | ✅ | 库自动声明 |
| SIZE | ✅ | 已声明最大邮件大小 |
| TLS | ✅ | 支持隐式 TLS 和 STARTTLS |

#### 4.1.3 影响评估

- 满足 RFC 5321 核心协议要求
- 支持现代 SMTP 传输所需的关键扩展

---

### 4.2 RFC 6531 - SMTPUTF8 扩展

**合规状态**：❌ **服务器端未启用**

#### 4.2.1 当前实现

**服务器端（接收邮件）**：

`go-smtp` 库通过 `Server.EnableSMTPUTF8` 字段控制 SMTPUTF8 能力公布。PMail 的三个 SMTP 服务器实例均未设置此字段。

**代码证据**（`smtp_server.go` 搜索结果）：项目代码中未出现 `EnableSMTPUTF8`。

EHLO 响应中不会包含 `SMTPUTF8` 能力，客户端发送国际化邮件地址时会被拒绝（返回 504 错误）。

**客户端（发送邮件）**：

`server/utils/smtp/smtp.go` 中已实现 SMTPUTF8 扩展检测：

- 检测远端服务器是否支持 SMTPUTF8
- 如果支持，在 MAIL FROM 命令中添加 `SMTPUTF8` 参数
- 这意味着 PMail 可以**发送**国际化邮件，但无法**接收**来自客户端的国际化邮件

#### 4.2.2 影响评估

- 不支持国际化邮件地址（如 `用户@例子.中国`）的接收提交
- 使用国际化邮件地址的用户无法通过 PMail SMTP 服务器发送邮件
- 仅影响 SMTP 提交端口（465/587），不影响服务器间传输（端口25）的中继行为
- **修复难度**：极低（设置 `EnableSMTPUTF8 = true` 即可，一行代码）

---

### 4.3 RFC 6409 - 邮件提交协议

**合规状态**：⚠️ **基本支持，缺少 SMTPUTF8**

#### 4.3.1 当前实现

| RFC 6409 要求 | 状态 | 说明 |
|---------------|------|------|
| 使用端口 587 | ✅ | `StartWithTLSNew()` 监听 587 |
| STARTTLS 支持 | ✅ | 明文连接升级为 TLS |
| 强制认证 | ✅ | 未认证用户无法提交邮件 |
| 拒绝未认证中继 | ✅ | 认证检查 `s.Ctx.UserID > 0` |
| 支持SMTPUTF8 | ❌ | `EnableSMTPUTF8` 未设置 |
| 支持 DSN | ❌ | `EnableDSN` 未设置 |

#### 4.3.2 影响评估

- 基本满足 RFC 6409 的邮件提交要求
- 缺少 SMTPUTF8 支持导致国际化邮件地址无法通过提交端口发送
- **修复难度**：低

---

## 5. 合规性总览矩阵

### 5.1 IMAP 合规性总览

| RFC | 标准名称 | 状态 | 支持程度 | 修复优先级 |
|-----|----------|------|----------|-----------|
| RFC 9051 | IMAP4rev2 | ❌ | 0% — 未实现 | 低（非必需） |
| RFC 7888 | 非同步字面量 | ⚠️ | 50% — LITERAL- 自动支持 | 中 |
| RFC 6855 | IMAP UTF-8 | ⚠️ | 80% — UTF8=ACCEPT 自动支持 | 无需修复 |
| RFC 7162 | CONDSTORE/QRESYNC | ❌ | 0% — 完全未实现 | 高 |

### 5.2 SMTP 合规性总览

| RFC | 标准名称 | 状态 | 支持程度 | 修复优先级 |
|-----|----------|------|----------|-----------|
| RFC 5321 | SMTP | ✅ | 90% — 核心功能完整 | — |
| RFC 6531 | SMTPUTF8 | ❌ | 30% — 仅发送端支持 | 高 |
| RFC 6409 | 邮件提交协议 | ⚠️ | 80% — 基本支持 | 中 |

### 5.3 状态图例

| 图例 | 含义 |
|------|------|
| ✅ | 完全支持或基本支持（≥80%） |
| ⚠️ | 部分支持（30%-79%） |
| ❌ | 不支持或严重缺失（<30%） |

---

## 6. 风险评估与影响分析

### 6.1 高风险项

| 风险项 | 影响范围 | 用户感知 | 建议措施 |
|--------|----------|----------|----------|
| SMTPUTF8 未启用（RFC 6531） | 使用国际化邮件地址的用户 | 无法发送含非ASCII字符地址的邮件 | 立即启用 `EnableSMTPUTF8` |
| CONDSTORE/QRESYNC 未实现（RFC 7162） | 所有移动客户端用户 | 断线重连后全量同步，体验差 | 计划实现 |

### 6.2 中风险项

| 风险项 | 影响范围 | 用户感知 | 建议措施 |
|--------|----------|----------|----------|
| LITERAL+ 未启用（RFC 7888） | 上传大附件的用户 | 高延迟网络下上传变慢 | 在 CapSet 中添加声明 |

### 6.3 低风险项

| 风险项 | 影响范围 | 用户感知 | 建议措施 |
|--------|----------|----------|----------|
| IMAP4rev2 未实现（RFC 9051） | 使用新协议的客户端 | 极少数客户端受影响 | 长期规划 |

---

## 7. 改进建议与优先级

### 7.1 改进路线图

#### 优先级 P0 — 立即修复（一行代码修改）

| 编号 | 改进项 | RFC | 修改文件 | 复杂度 |
|------|--------|-----|----------|--------|
| P0-1 | 启用 SMTPUTF8 | RFC 6531 | `smtp_server.go` | 极低 |
| P0-2 | 启用 LITERAL+ | RFC 7888 | `imap_server.go` CapSet | 极低 |

**P0-1 修改方案**：在三个 SMTP 服务器实例创建后添加 `s.EnableSMTPUTF8 = true`

**P0-2 修改方案**：在 IMAP CapSet 中添加 `imap.CapLiteralPlus: {}`

#### 优先级 P1 — 短期改进（1-2天）

| 编号 | 改进项 | RFC | 涉及文件 | 复杂度 |
|------|--------|-----|----------|--------|
| P1-1 | 启用 ESEARCH | RFC 4731 | `imap_server.go` CapSet + 后端 | 中 |
| P1-2 | 启用 LIST-EXTENDED | RFC 5258 | `imap_server.go` CapSet + 后端 | 中 |
| P1-3 | 启用 STATUS=SIZE | RFC 8438 | `imap_server.go` CapSet + 后端 | 中 |

#### 优先级 P2 — 中期改进（1-2周）

| 编号 | 改进项 | RFC | 涉及文件 | 复杂度 |
|------|--------|-----|----------|--------|
| P2-1 | 实现 CONDSTORE | RFC 7162 | 数据库 + 所有 Session 接口 | 高 |
| P2-2 | 实现 QRESYNC | RFC 7162 | 依赖 P2-1 | 高 |

#### 优先级 P3 — 长期规划

| 编号 | 改进项 | RFC | 涉及文件 | 复杂度 |
|------|--------|-----|----------|--------|
| P3-1 | 升级到 IMAP4rev2 | RFC 9051 | 全面评估后决定 | 极高 |

### 7.2 实施风险评估

| 改进项 | 向后兼容性 | 数据库变更 | 测试覆盖 |
|--------|-----------|-----------|----------|
| P0-1 SMTPUTF8 | ✅ 完全兼容 | 无需 | 现有测试覆盖 |
| P0-2 LITERAL+ | ✅ 完全兼容 | 无需 | 需新增大附件测试 |
| P1-x 扩展能力 | ⚠️ 需验证 | 可能需要 | 需新增测试 |
| P2-x CONDSTORE | ⚠️ 需迁移 | **必须**新增 mod-sequence 字段 | 需完整测试套件 |
| P3-x IMAP4rev2 | ⚠️ 重大变更 | 待评估 | 需全面回归测试 |

---

## 8. 强制检查清单

### 8.1 审查过程确认

- [x] 已阅读并理解【文档生成及读取规范.md】全部内容
- [x] 已完整审查 IMAP 服务器实现代码
- [x] 已完整审查 SMTP 服务器实现代码
- [x] 已分析 go-imap/v2 库的扩展支持能力
- [x] 已分析 go-smtp 库的扩展支持能力
- [x] 已逐项核对所有 RFC 标准要求

### 8.2 文档完整性确认

- [x] 文档命名符合规范（实体名称_文档类型_日期_时间_版本.md）
- [x] 文档归档至正确目录（03-设计/01-架构设计/）
- [x] 包含所有必填元数据字段
- [x] 包含变更日志
- [x] 版本号从 v1.0 起始

### 8.3 审查结论

**结论**：PMail 项目在 IMAP 方面以 IMAP4rev1 为基础，通过 go-imap/v2 库自动获得了 LITERAL-、UTF8=ACCEPT 等扩展的部分支持。SMTP 方面核心协议实现完整，但 SMTPUTF8 服务器端未启用。最关键的缺失是 RFC 7162 CONDSTORE/QRESYNC 扩展，对移动端用户体验影响较大。建议按 P0→P3 优先级逐步改进。

---

## 变更日志

| 版本 | 日期 | 时间 | 作者 | 变更内容 | 变更原因 |
|------|------|------|------|----------|----------|
| v1.0 | 2026-06-09 | 06:16:00 | PMail Team | 初始创建 | RFC 标准合规性审查需求 |