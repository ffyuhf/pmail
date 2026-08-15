# user_edituser_admin-disable-guard_修正

> **版本**: v1.0  
> **日期**: 2026-06-09  
> **时间**: 23:36:00  
> **作者**: PMail Team  
> **模块**: user  
> **状态**: 已批准

## 1. 修改前后代码要点对比

### 修改文件位置（目录树）

```
server/
└── controllers/
    └── user.go          ← 唯一修改文件
```

### 修改文件名及行号

- **文件**: `server/controllers/user.go`
- **修改位置**: `EditUser` 函数，原第 158 行之后（`if user.ID == 0` 检查之后），新增第 161-165 行

### 函数名/接口签名的变更

无接口签名变更，仅函数内部逻辑变更。

### 修改前代码

```go
if user.ID == 0 {
    response.NewErrorResponse(response.ParamsError, "User not found", "").FPrint(w)
    return
}
if reqData.Username != "" && reqData.Username != user.Name {
```

### 修改后代码

```go
if user.ID == 0 {
    response.NewErrorResponse(response.ParamsError, "User not found", "").FPrint(w)
    return
}

// 禁止禁用管理员账号，防止系统被锁定无法登录
// 修改时间：2026-06-09 23:36，修改原因：管理员可被禁用导致系统无管理员可用
if reqData.Disabled == 1 && user.IsAdmin == 1 {
    response.NewErrorResponse(response.NoAccessPrivileges, "Cannot disable admin account", "").FPrint(w)
    return
}

if reqData.Username != "" && reqData.Username != user.Name {
```

## 2. 变更原因与影响点

### 变更原因
`EditUser` 接口缺少管理员账号保护逻辑，管理员账号可通过设置 `disabled=1` 被禁用。一旦管理员被禁用，由于 `Login` 函数查询条件包含 `disabled=0`，管理员将无法登录系统，导致系统无管理员可用。

### 影响范围
- 仅影响 `EditUser` API 的禁用操作分支
- 对用户名修改、密码修改等操作无影响
- 对 `CreateUser`、`UserList`、`Login` 等其他接口无影响

### 接口兼容性
- 完全兼容：请求格式不变
- 仅新增一种错误响应：当尝试禁用管理员时返回 `NoAccessPrivileges` 错误码

### 性能影响
无性能影响，仅增加一次整数比较判断。

### 安全影响
- 修复了管理员可被禁用的安全漏洞
- 防止系统因管理员被禁用而失去管理能力

## 3. 测试与验证记录

### 测试用例

| 用例 | 操作 | 预期结果 |
|------|------|----------|
| TC1 | 禁用管理员账号（`disabled=1`, 目标 `is_admin=1`） | 返回 `NoAccessPrivileges` 错误，操作被拒绝 |
| TC2 | 禁用普通用户（`disabled=1`, 目标 `is_admin=0`） | 正常禁用，行为不变 |
| TC3 | 修改管理员用户名（不涉及 `disabled` 字段） | 正常修改，行为不变 |
| TC4 | 启用已禁用的管理员（`disabled=0`, 目标 `is_admin=1`） | 正常启用，行为不变 |

### 测试结果
- Go 编译通过，无语法错误
- 逻辑验证：仅当 `reqData.Disabled == 1 && user.IsAdmin == 1` 同时满足时才拦截

### 结论
修复逻辑正确，满足预期目标。

## 4. 相关文档引用

| 文档 | 路径 |
|------|------|
| 计划书 | `docs/02-计划/user_edituser_admin-disable-guard_计划_20260609_23-35-00_v1.0.md` |
| 变更日志 | `docs/04-模块/01-变更日志/变更日志_汇总_20260609_23-36-00_v1.8.md` |

## 变更日志

| 版本 | 日期 | 时间 | 作者 | 变更内容 | 变更原因 |
|------|------|------|------|----------|----------|
| v1.0 | 2026-06-09 | 23:36:00 | PMail Team | 初始创建 | 管理员账号可被禁用的安全漏洞修复 |