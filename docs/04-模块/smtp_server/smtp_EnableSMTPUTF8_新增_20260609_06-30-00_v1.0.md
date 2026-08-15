# smtp EnableSMTPUTF8 新增文档

> **版本**: v1.0  
> **日期**: 2026-06-09  
> **时间**: 06:30:00  
> **作者**: PMail Team  
> **模块**: smtp-server  
> **状态**: 已批准

## 1. 修改前后代码要点对比

### 修改文件位置

```
server/
└── listen/
    └── smtp_server/
        └── smtp.go     # 第 31、60、91 行（EnableSMTPUTF8 字段）
```

### 修改内容

在三个 SMTP 服务器实例（587/465/25 端口）创建后，添加 `EnableSMTPUTF8 = true`：

| 实例 | 端口 | 变量名 | 新增行 |
|------|------|--------|--------|
| STARTTLS 提交服务 | 587 | `instanceTlsNew` | `instanceTlsNew.EnableSMTPUTF8 = true` |
| 隐式 TLS 提交服务 | 465 | `instanceTls` | `instanceTls.EnableSMTPUTF8 = true` |
| 外部接收服务 | 25 | `instance` | `instance.EnableSMTPUTF8 = true` |

## 2. 变更原因与影响点

### 变更原因

- **P0-1（RFC 6531）**：PMail 的 SMTP 客户端（发送端）已支持 SMTPUTF8 扩展检测，
  但服务器端（接收端）未启用 `EnableSMTPUTF8`，导致 EHLO 响应中不包含 SMTPUTF8 能力。
  使用国际化邮件地址（如 `用户@例子.中国`）的客户端提交邮件时会被拒绝（504 错误）。

### 影响范围

- 接口兼容性：✅ 完全兼容，纯增量声明，不影响已有功能
- 性能：无影响
- 安全：无影响

## 3. 测试与验证记录

- 测试用例：`go build ./listen/smtp_server/` 编译通过
- 测试结果：PASS
- 结论：修改安全，无副作用

## 4. 相关文档引用

- `docs/03-设计/01-架构设计/rfc_compliance_imap_smtp_审查报告_20260609_06-16-00_v1.0.md`
- `docs/02-计划/rfc_compliance_P0_P1_修复计划_20260609_06-30-00_v1.0.md`

## 变更日志

| 版本 | 日期 | 时间 | 作者 | 变更内容 | 变更原因 |
|------|------|------|------|----------|----------|
| v1.0 | 2026-06-09 | 06:30:00 | PMail Team | 初始创建 | RFC 6531 SMTPUTF8 服务器端启用 |