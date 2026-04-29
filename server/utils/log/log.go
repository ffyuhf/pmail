// Package log 提供统一的日志辅助函数，为各协议服务器输出结构化日志。
// 日志格式规范：[级别][时间][LogID][协议] 消息
// 每个函数都绑定协议标识，确保日志可按服务类型过滤。
package log

import (
	"fmt"

	pmailContext "github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
)

// EventType 定义日志事件类型常量，用于标识具体的业务操作。
// 格式为 [协议][事件类型]，便于日志过滤和问题诊断。
const (
	// ---- SMTP 事件类型 ----
	// EventSMTPSessionNew 标识 SMTP 新连接建立。
	EventSMTPSessionNew = "SESSION_NEW"
	// EventSMTPSessionClose 标识 SMTP 连接关闭。
	EventSMTPSessionClose = "SESSION_CLOSE"
	// EventSMTPAuth 标识 SMTP 认证请求（含机制信息）。
	EventSMTPAuth = "AUTH"
	// EventSMTPAuthSuccess 标识 SMTP 认证成功。
	EventSMTPAuthSuccess = "AUTH_SUCCESS"
	// EventSMTPAuthFail 标识 SMTP 认证失败。
	EventSMTPAuthFail = "AUTH_FAIL"
	// EventSMTPMailFrom 标识 SMTP MAIL FROM 命令。
	EventSMTPMailFrom = "MAIL_FROM"
	// EventSMTPRcptTo 标识 SMTP RCPT TO 命令。
	EventSMTPRcptTo = "RCPT_TO"
	// EventSMTPMailRecv 标识 SMTP 接收到完整邮件数据。
	EventSMTPMailRecv = "MAIL_RECV"
	// EventSMTPMailSend 标识 SMTP 发送邮件。
	EventSMTPMailSend = "MAIL_SEND"
	// EventSMTPMailDeliver 标识 SMTP 邮件投递完成。
	EventSMTPMailDeliver = "MAIL_DELIVER"
	// EventSMTPMailReject 标识 SMTP 邮件被拒绝（垃圾邮件/验证失败）。
	EventSMTPMailReject = "MAIL_REJECT"
	// EventSMTPMailForwardAdmin 标识 SMTP 邮件转交管理员（收件人不存在）。
	EventSMTPMailForwardAdmin = "MAIL_FORWARD_ADMIN"
	// EventSMTPDKIM 标识 SMTP DKIM 验证结果。
	EventSMTPDKIM = "DKIM"
	// EventSMTPSPF 标识 SMTP SPF 验证结果。
	EventSMTPSPF = "SPF"
	// EventSMTPRateLimit 标识 SMTP 速率限制触发。
	EventSMTPRateLimit = "RATE_LIMIT"
	// EventSMTPPlugin 标识 SMTP 插件执行。
	EventSMTPPlugin = "PLUGIN"

	// ---- IMAP 事件类型 ----
	// EventIMAPSessionNew 标识 IMAP 新连接建立。
	EventIMAPSessionNew = "SESSION_NEW"
	// EventIMAPSessionClose 标识 IMAP 连接关闭。
	EventIMAPSessionClose = "SESSION_CLOSE"
	// EventIMAPAuthSuccess 标识 IMAP 认证成功。
	EventIMAPAuthSuccess = "AUTH_SUCCESS"
	// EventIMAPAuthFail 标识 IMAP 认证失败。
	EventIMAPAuthFail = "AUTH_FAIL"
	// EventIMAPRateLimit 标识 IMAP 速率限制触发。
	EventIMAPRateLimit = "RATE_LIMIT"
	// EventIMAPExpunge 标识 IMAP EXPUNGE 命令执行。
	EventIMAPExpunge = "EXPUNGE"
	// EventIMAPIdleEnter 标识 IMAP IDLE 命令进入。
	EventIMAPIdleEnter = "IDLE_ENTER"
	// EventIMAPIdleTimeout 标识 IMAP IDLE 超时断开。
	EventIMAPIdleTimeout = "IDLE_TIMEOUT"

	// ---- POP3 事件类型 ----
	// EventPOP3SessionNew 标识 POP3 新连接建立。
	EventPOP3SessionNew = "SESSION_NEW"
	// EventPOP3AuthSuccess 标识 POP3 认证成功。
	EventPOP3AuthSuccess = "AUTH_SUCCESS"
	// EventPOP3AuthFail 标识 POP3 认证失败。
	EventPOP3AuthFail = "AUTH_FAIL"
	// EventPOP3RateLimit 标识 POP3 速率限制触发。
	EventPOP3RateLimit = "RATE_LIMIT"
	// EventPOP3AuthRejected 标识 POP3 非TLS连接认证被拒。
	EventPOP3AuthRejected = "AUTH_REJECTED"
	// EventPOP3Cmd 标识 POP3 命令执行。
	EventPOP3Cmd = "CMD"

	// ---- 协议标识常量 ----
	// ProtocolSMTP 标识 SMTP 协议。
	ProtocolSMTP = "SMTP"
	// ProtocolIMAP 标识 IMAP 协议。
	ProtocolIMAP = "IMAP"
	// ProtocolPOP3 标识 POP3 协议。
	ProtocolPOP3 = "POP3"
)

// formatEvent 生成结构化日志消息，格式为 [协议][事件类型] 消息内容。
// 这是所有日志辅助函数的核心格式化入口。
func formatEvent(protocol, eventType, message string) string {
	return fmt.Sprintf("[%s][%s] %s", protocol, eventType, message)
}

// formatEventf 生成结构化日志消息，支持格式化参数。
func formatEventf(protocol, eventType, format string, args ...any) string {
	return fmt.Sprintf("[%s][%s] %s", protocol, eventType, fmt.Sprintf(format, args...))
}

// ---- SMTP 日志函数 ----

// SmtpInfo 输出 SMTP 协议的 INFO 级别结构化日志。
func SmtpInfo(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Info(formatEvent(ProtocolSMTP, eventType, message))
}

// SmtpInfof 输出 SMTP 协议的 INFO 级别结构化日志（支持格式化参数）。
func SmtpInfof(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Info(formatEventf(ProtocolSMTP, eventType, format, args...))
}

// SmtpDebug 输出 SMTP 协议的 DEBUG 级别结构化日志。
func SmtpDebug(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Debug(formatEvent(ProtocolSMTP, eventType, message))
}

// SmtpDebugf 输出 SMTP 协议的 DEBUG 级别结构化日志（支持格式化参数）。
func SmtpDebugf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Debug(formatEventf(ProtocolSMTP, eventType, format, args...))
}

// SmtpWarn 输出 SMTP 协议的 WARN 级别结构化日志。
func SmtpWarn(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Warn(formatEvent(ProtocolSMTP, eventType, message))
}

// SmtpWarnf 输出 SMTP 协议的 WARN 级别结构化日志（支持格式化参数）。
func SmtpWarnf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Warn(formatEventf(ProtocolSMTP, eventType, format, args...))
}

// SmtpError 输出 SMTP 协议的 ERROR 级别结构化日志。
func SmtpError(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Error(formatEvent(ProtocolSMTP, eventType, message))
}

// SmtpErrorf 输出 SMTP 协议的 ERROR 级别结构化日志（支持格式化参数）。
func SmtpErrorf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Error(formatEventf(ProtocolSMTP, eventType, format, args...))
}

// ---- IMAP 日志函数 ----

// ImapInfo 输出 IMAP 协议的 INFO 级别结构化日志。
func ImapInfo(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Info(formatEvent(ProtocolIMAP, eventType, message))
}

// ImapInfof 输出 IMAP 协议的 INFO 级别结构化日志（支持格式化参数）。
func ImapInfof(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Info(formatEventf(ProtocolIMAP, eventType, format, args...))
}

// ImapDebug 输出 IMAP 协议的 DEBUG 级别结构化日志。
func ImapDebug(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Debug(formatEvent(ProtocolIMAP, eventType, message))
}

// ImapDebugf 输出 IMAP 协议的 DEBUG 级别结构化日志（支持格式化参数）。
func ImapDebugf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Debug(formatEventf(ProtocolIMAP, eventType, format, args...))
}

// ImapWarn 输出 IMAP 协议的 WARN 级别结构化日志。
func ImapWarn(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Warn(formatEvent(ProtocolIMAP, eventType, message))
}

// ImapWarnf 输出 IMAP 协议的 WARN 级别结构化日志（支持格式化参数）。
func ImapWarnf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Warn(formatEventf(ProtocolIMAP, eventType, format, args...))
}

// ImapError 输出 IMAP 协议的 ERROR 级别结构化日志。
func ImapError(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Error(formatEvent(ProtocolIMAP, eventType, message))
}

// ImapErrorf 输出 IMAP 协议的 ERROR 级别结构化日志（支持格式化参数）。
func ImapErrorf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Error(formatEventf(ProtocolIMAP, eventType, format, args...))
}

// ---- POP3 日志函数 ----

// Pop3Info 输出 POP3 协议的 INFO 级别结构化日志。
func Pop3Info(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Info(formatEvent(ProtocolPOP3, eventType, message))
}

// Pop3Infof 输出 POP3 协议的 INFO 级别结构化日志（支持格式化参数）。
func Pop3Infof(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Info(formatEventf(ProtocolPOP3, eventType, format, args...))
}

// Pop3Debug 输出 POP3 协议的 DEBUG 级别结构化日志。
func Pop3Debug(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Debug(formatEvent(ProtocolPOP3, eventType, message))
}

// Pop3Debugf 输出 POP3 协议的 DEBUG 级别结构化日志（支持格式化参数）。
func Pop3Debugf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Debug(formatEventf(ProtocolPOP3, eventType, format, args...))
}

// Pop3Warn 输出 POP3 协议的 WARN 级别结构化日志。
func Pop3Warn(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Warn(formatEvent(ProtocolPOP3, eventType, message))
}

// Pop3Warnf 输出 POP3 协议的 WARN 级别结构化日志（支持格式化参数）。
func Pop3Warnf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Warn(formatEventf(ProtocolPOP3, eventType, format, args...))
}

// Pop3Error 输出 POP3 协议的 ERROR 级别结构化日志。
func Pop3Error(ctx *pmailContext.Context, eventType, message string) {
	log.WithContext(ctx).Error(formatEvent(ProtocolPOP3, eventType, message))
}

// Pop3Errorf 输出 POP3 协议的 ERROR 级别结构化日志（支持格式化参数）。
func Pop3Errorf(ctx *pmailContext.Context, eventType, format string, args ...any) {
	log.WithContext(ctx).Error(formatEventf(ProtocolPOP3, eventType, format, args...))
}
