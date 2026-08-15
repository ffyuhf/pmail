// Package dsn 实现 RFC 3462 multipart/report 格式的投递状态通知（DSN）生成。
//
// DSN 用于在邮件投递失败（或成功/延迟，本实现仅覆盖 FAILURE）时，
// 向原始发件人发送结构化的退信报告。报告包含三部分：
//  1. Human-readable part（text/plain）：人类可读的退信说明
//  2. Machine-readable part（message/delivery-status）：结构化状态码
//  3. Original message part（text/rfc822-headers 或 message/rfc822）：原始邮件头或全文
//
// 防护措施：
//   - 所有 DSN 报告自动添加 Auto-Submitted: auto-replied 头，防止退信循环
//   - 若原始邮件已包含 Auto-Submitted 头，则不生成 DSN
//
// 参考资料：
//   - RFC 3461: SMTP Service Extension for Delivery Status Notifications
//   - RFC 3462: The Multipart/Report Content Type
//   - RFC 3464: An Extensible Message Format for Delivery Status Notifications
//   - RFC 6533: Extension for Internationalized Email
package dsn

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"mime"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/emersion/go-smtp"
)

// FailedRecipient 描述一个投递失败的收件人及其失败原因。
type FailedRecipient struct {
	// Address 失败的收件人地址（RFC 5321 信封中的 RCPT TO）。
	Address string
	// DSNStatus RFC 3463 增强状态码，例如 "5.1.1"（邮箱不存在）、"5.4.4"（无法路由）。
	DSNStatus string
	// DiagnosticCode 诊断信息，通常为 SMTP 响应码和文本，例如 "550 5.1.1 User unknown".
	DiagnosticCode string
	// RcptOpts 该收件人在 RCPT TO 阶段声明的 DSN 选项（NOTIFY/ORCPT），可为 nil。
	RcptOpts *smtp.RcptOptions
}

// GenerateFailureDSN 生成投递失败的 DSN 报告（RFC 3462 multipart/report 格式）。
//
// 参数：
//   - originalFrom:     原始邮件的发件人地址（退信报告的收件人）
//   - originalHeaders:  原始邮件的完整头部（用于检查 Auto-Submitted 和构造 text/rfc822-headers 部分）
//   - failedRcpts:      失败的收件人列表及其状态码和诊断信息
//   - mailOpts:         MAIL FROM 阶段的 DSN 选项（RET=FULL/HDRS, ENVID），可为 nil
//   - domain:           本地域名（用于构造 MAILER-DAEMON 发件人地址）
//   - originalMsgBody:  原始邮件的完整正文（仅当 RET=FULL 时用于构造 message/rfc822 部分）
//
// 返回：
//   - 符合 RFC 3462 multipart/report 格式的完整邮件字节数据
//   - 若原始邮件已包含 Auto-Submitted 头（即不应生成 DSN），返回 nil, nil
func GenerateFailureDSN(
	originalFrom string,
	originalHeaders textproto.MIMEHeader,
	failedRcpts []FailedRecipient,
	mailOpts *smtp.MailOptions,
	domain string,
	originalMsgBody []byte,
) ([]byte, error) {
	// 防护：若原始邮件已包含 Auto-Submitted 头，不生成 DSN，防止退信循环（RFC 3834）。
	if isAutoSubmitted(originalHeaders) {
		return nil, nil
	}

	// 生成唯一的 MIME boundary
	boundary := generateBoundary()

	var buf bytes.Buffer

	// === 1. 邮件头 ===
	writeDSNHeaders(&buf, originalFrom, domain, boundary, mailOpts)

	// === 2. Human-readable part（text/plain）===
	writeHumanReadablePart(&buf, boundary, originalFrom, failedRcpts)

	// === 3. Machine-readable part（message/delivery-status）===
	writeDeliveryStatusPart(&buf, boundary, originalFrom, failedRcpts, mailOpts, domain)

	// === 4. Original message part ===
	writeOriginalMessagePart(&buf, boundary, mailOpts, originalHeaders, originalMsgBody)

	// 结束 boundary
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)

	return buf.Bytes(), nil
}

// isAutoSubmitted 检查原始邮件是否包含 Auto-Submitted 头。
// 若存在且值不为 "no"，则认为该邮件是自动生成的，不应发送 DSN（RFC 3834）。
func isAutoSubmitted(headers textproto.MIMEHeader) bool {
	val := headers.Get("Auto-Submitted")
	if val == "" {
		return false
	}
	val = strings.ToLower(strings.TrimSpace(val))
	return val != "no"
}

// generateBoundary 生成一个随机的 MIME boundary 字符串。
func generateBoundary() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("==%x==", b)
}

// writeDSNHeaders 写入 DSN 报告的顶层邮件头。
func writeDSNHeaders(buf *bytes.Buffer, originalFrom, domain, boundary string, mailOpts *smtp.MailOptions) {
	// Date: DSN 生成时间
	fmt.Fprintf(buf, "Date: %s\r\n", time.Now().Format(time.RFC1123Z))

	// From: MAILER-DAEMON@domain（RFC 3464 Section 4.1）
	fmt.Fprintf(buf, "From: Mail Delivery Subsystem <MAILER-DAEMON@%s>\r\n", domain)

	// To: 原始发件人（退信报告的收件人）
	fmt.Fprintf(buf, "To: <%s>\r\n", originalFrom)

	// Subject: 退信通知主题，包含原始收件人信息（便于识别）
	fmt.Fprintf(buf, "Subject: Delivery Status Notification (Failure)\r\n")

	// MIME-Version
	fmt.Fprintf(buf, "MIME-Version: 1.0\r\n")

	// Auto-Submitted: 标记为自动生成，防止退信循环（RFC 3834）
	fmt.Fprintf(buf, "Auto-Submitted: auto-replied (dsn)\r\n")

	// Content-Type: multipart/report
	fmt.Fprintf(buf, "Content-Type: multipart/report; report-type=delivery-status;\r\n")
	fmt.Fprintf(buf, "\tboundary=\"%s\"\r\n", boundary)

	// 若存在 Envelope-ID，添加 Original-Envelope-Id 头（RFC 3461 Section 5.3）
	if mailOpts != nil && mailOpts.EnvelopeID != "" {
		fmt.Fprintf(buf, "Original-Envelope-Id: %s\r\n", mailOpts.EnvelopeID)
	}

	buf.WriteString("\r\n")
}

// writeHumanReadablePart 写入 DSN 的第一部分：人类可读的退信说明（text/plain）。
func writeHumanReadablePart(buf *bytes.Buffer, boundary string, originalFrom string, failedRcpts []FailedRecipient) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	buf.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	buf.WriteString("\r\n")

	buf.WriteString("This is a delivery status notification (failure) message.\r\n\r\n")
	buf.WriteString(fmt.Sprintf("Your message to the following recipients could not be delivered:\r\n\r\n"))

	for _, rcpt := range failedRcpts {
		buf.WriteString(fmt.Sprintf("  - %s\r\n", rcpt.Address))
		if rcpt.DiagnosticCode != "" {
			buf.WriteString(fmt.Sprintf("    Reason: %s\r\n", rcpt.DiagnosticCode))
		}
	}

	buf.WriteString("\r\n")
}

// writeDeliveryStatusPart 写入 DSN 的第二部分：结构化的投递状态信息（message/delivery-status）。
//
// 格式遵循 RFC 3464 Section 2.1：
//   - Per-message fields（Reporting-MTA, Arrival-Date 等）
//   - Per-recipient fields（Final-Recipient, Action, Status, Diagnostic-Code 等）
func writeDeliveryStatusPart(
	buf *bytes.Buffer,
	boundary string,
	originalFrom string,
	failedRcpts []FailedRecipient,
	mailOpts *smtp.MailOptions,
	domain string,
) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)
	buf.WriteString("Content-Type: message/delivery-status\r\n")
	buf.WriteString("\r\n")

	// --- Per-message fields ---
	fmt.Fprintf(buf, "Reporting-MTA: dns; %s\r\n", domain)
	fmt.Fprintf(buf, "Arrival-Date: %s\r\n", time.Now().Format(time.RFC1123Z))

	// 若存在 Envelope-ID
	if mailOpts != nil && mailOpts.EnvelopeID != "" {
		fmt.Fprintf(buf, "Original-Envelope-Id: %s\r\n", mailOpts.EnvelopeID)
	}

	// --- Per-recipient fields（每个失败收件人一组）---
	for _, rcpt := range failedRcpts {
		buf.WriteString("\r\n") // 组间空行分隔

		// Final-Recipient: 最终收件人地址类型和地址
		addrType := "rfc822"
		if rcpt.RcptOpts != nil && rcpt.RcptOpts.OriginalRecipientType != "" {
			addrType = strings.ToLower(string(rcpt.RcptOpts.OriginalRecipientType))
		}
		fmt.Fprintf(buf, "Final-Recipient: %s; %s\r\n", addrType, rcpt.Address)

		// Action: 投递动作，失败场景固定为 "failed"
		buf.WriteString("Action: failed\r\n")

		// Status: RFC 3463 增强状态码
		status := rcpt.DSNStatus
		if status == "" {
			status = "5.0.0" // 默认通用永久失败状态码
		}
		fmt.Fprintf(buf, "Status: %s\r\n", status)

		// Diagnostic-Code: SMTP 诊断信息
		if rcpt.DiagnosticCode != "" {
			fmt.Fprintf(buf, "Diagnostic-Code: smtp; %s\r\n", rcpt.DiagnosticCode)
		}

		// Original-Recipient: 原始收件人（若与 Final-Recipient 不同）
		if rcpt.RcptOpts != nil && rcpt.RcptOpts.OriginalRecipient != "" {
			orcptType := "rfc822"
			if rcpt.RcptOpts.OriginalRecipientType != "" {
				orcptType = strings.ToLower(string(rcpt.RcptOpts.OriginalRecipientType))
			}
			fmt.Fprintf(buf, "Original-Recipient: %s; %s\r\n", orcptType, rcpt.RcptOpts.OriginalRecipient)
		}
	}

	buf.WriteString("\r\n")
}

// writeOriginalMessagePart 写入 DSN 的第三部分：原始邮件内容。
//
// 根据 RET 参数决定返回内容范围：
//   - RET=HDRS（默认）：仅返回原始邮件头（text/rfc822-headers）
//   - RET=FULL：返回完整邮件（message/rfc822）
func writeOriginalMessagePart(
	buf *bytes.Buffer,
	boundary string,
	mailOpts *smtp.MailOptions,
	originalHeaders textproto.MIMEHeader,
	originalMsgBody []byte,
) {
	fmt.Fprintf(buf, "--%s\r\n", boundary)

	// 判断 RET 参数：默认为 HDRS
	returnFull := mailOpts != nil && mailOpts.Return == smtp.DSNReturnFull

	if returnFull && len(originalMsgBody) > 0 {
		// RET=FULL：返回完整邮件（message/rfc822）
		buf.WriteString("Content-Type: message/rfc822\r\n")
		buf.WriteString("\r\n")
		buf.Write(originalMsgBody)
	} else {
		// RET=HDRS（默认）：仅返回邮件头（text/rfc822-headers）
		buf.WriteString("Content-Type: text/rfc822-headers\r\n")
		buf.WriteString("\r\n")
		// 写入原始邮件的所有头部字段
		for field, values := range originalHeaders {
			for _, val := range values {
				// 对非 ASCII 头字段进行编码
				encodedVal := maybeEncodeHeader(val)
				fmt.Fprintf(buf, "%s: %s\r\n", field, encodedVal)
			}
		}
		buf.WriteString("\r\n")
	}

	buf.WriteString("\r\n")
}

// maybeEncodeHeader 对包含非 ASCII 字符的头字段值进行 RFC 2047 编码。
// 若内容为纯 ASCII 则原样返回，避免不必要的编码。
func maybeEncodeHeader(val string) string {
	if isASCII(val) {
		return val
	}
	return mime.QEncoding.Encode("utf-8", val)
}

// isASCII 检查字符串是否仅包含 ASCII 字符。
func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

// ParseEmailAddress 从邮件地址字符串中提取纯地址部分（去除显示名和尖括号）。
// 例如 "User Name <user@example.com>" → "user@example.com"
func ParseEmailAddress(addr string) string {
	a, err := mail.ParseAddress(addr)
	if err != nil {
		// 解析失败时尝试去除尖括号
		addr = strings.Trim(addr, " <>\"")
		return addr
	}
	return a.Address
}

// WantsFailureDSN 检查指定收件人是否请求了失败通知。
// 根据 RFC 3461 Section 4.1：
//   - 若未指定 NOTIFY 参数，默认行为等效于 NOTIFY=FAILURE
//   - 若 NOTIFY 包含 FAILURE，则请求失败通知
//   - 若 NOTIFY 包含 NEVER，则不发送任何通知
func WantsFailureDSN(opts *smtp.RcptOptions) bool {
	if opts == nil || len(opts.Notify) == 0 {
		// 未指定 NOTIFY，默认发送失败通知
		return true
	}
	for _, n := range opts.Notify {
		if n == smtp.DSNNotifyNever {
			return false
		}
		if n == smtp.DSNNotifyFailure {
			return true
		}
	}
	return false
}

