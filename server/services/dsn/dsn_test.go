package dsn

import (
	"net/textproto"
	"strings"
	"testing"

	"github.com/emersion/go-smtp"
)

// TestGenerateFailureDSN_BasicFormat 验证 DSN 报告的基本 MIME 结构：
// 1. 包含 multipart/report Content-Type
// 2. 包含三个部分（text/plain, message/delivery-status, text/rfc822-headers）
// 3. 包含 Auto-Submitted 头（防止退信循环）
func TestGenerateFailureDSN_BasicFormat(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")
	headers.Set("Subject", "Test")

	failedRcpts := []FailedRecipient{
		{
			Address:        "nosuch@example.org",
			DSNStatus:      "5.1.1",
			DiagnosticCode: "550 5.1.1 User unknown",
		},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}
	if report == nil {
		t.Fatal("GenerateFailureDSN 返回 nil（不应跳过 DSN 生成）")
	}

	reportStr := string(report)

	// 验证关键头字段
	if !strings.Contains(reportStr, "Content-Type: multipart/report") {
		t.Error("DSN 报告缺少 multipart/report Content-Type")
	}
	if !strings.Contains(reportStr, "report-type=delivery-status") {
		t.Error("DSN 报告缺少 report-type=delivery-status")
	}
	if !strings.Contains(reportStr, "Auto-Submitted: auto-replied") {
		t.Error("DSN 报告缺少 Auto-Submitted 头（防止退信循环）")
	}
	if !strings.Contains(reportStr, "From: Mail Delivery Subsystem <MAILER-DAEMON@example.com>") {
		t.Error("DSN 报告缺少正确的 From 头")
	}
	if !strings.Contains(reportStr, "To: <sender@example.com>") {
		t.Error("DSN 报告缺少正确的 To 头")
	}
}

// TestGenerateFailureDSN_DeliveryStatusFormat 验证 message/delivery-status 部分的格式：
// 1. 包含 Per-message 字段（Reporting-MTA, Arrival-Date）
// 2. 包含 Per-recipient 字段（Final-Recipient, Action: failed, Status, Diagnostic-Code）
func TestGenerateFailureDSN_DeliveryStatusFormat(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")

	failedRcpts := []FailedRecipient{
		{
			Address:        "nosuch@example.org",
			DSNStatus:      "5.1.1",
			DiagnosticCode: "550 5.1.1 User unknown",
		},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	// 验证 delivery-status 部分的 Content-Type
	if !strings.Contains(reportStr, "Content-Type: message/delivery-status") {
		t.Error("DSN 报告缺少 message/delivery-status 部分")
	}

	// 验证 Per-message 字段
	if !strings.Contains(reportStr, "Reporting-MTA: dns; example.com") {
		t.Error("DSN 报告缺少 Reporting-MTA 字段")
	}
	if !strings.Contains(reportStr, "Arrival-Date:") {
		t.Error("DSN 报告缺少 Arrival-Date 字段")
	}

	// 验证 Per-recipient 字段
	if !strings.Contains(reportStr, "Final-Recipient: rfc822; nosuch@example.org") {
		t.Error("DSN 报告缺少 Final-Recipient 字段")
	}
	if !strings.Contains(reportStr, "Action: failed") {
		t.Error("DSN 报告缺少 Action: failed 字段")
	}
	if !strings.Contains(reportStr, "Status: 5.1.1") {
		t.Error("DSN 报告缺少 Status 字段")
	}
	if !strings.Contains(reportStr, "Diagnostic-Code: smtp; 550 5.1.1 User unknown") {
		t.Error("DSN 报告缺少 Diagnostic-Code 字段")
	}
}

// TestGenerateFailureDSN_AutoSubmittedSuppression 验证：
// 若原始邮件包含 Auto-Submitted 头（非 "no"），不生成 DSN，防止退信循环。
func TestGenerateFailureDSN_AutoSubmittedSuppression(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("Auto-Submitted", "auto-replied")

	failedRcpts := []FailedRecipient{
		{Address: "nosuch@example.org", DSNStatus: "5.1.1"},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}
	if report != nil {
		t.Error("对 Auto-Submitted 邮件不应生成 DSN，但返回了非 nil 结果")
	}
}

// TestGenerateFailureDSN_MultipleRecipients 验证多收件人部分失败场景：
// DSN 报告应包含所有失败收件人的 Per-recipient 字段。
func TestGenerateFailureDSN_MultipleRecipients(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")

	failedRcpts := []FailedRecipient{
		{
			Address:        "user1@bad.example.org",
			DSNStatus:      "5.1.1",
			DiagnosticCode: "550 5.1.1 User unknown",
		},
		{
			Address:        "user2@down.example.net",
			DSNStatus:      "4.4.4",
			DiagnosticCode: "451 4.4.4 Unable to route",
		},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	if !strings.Contains(reportStr, "Final-Recipient: rfc822; user1@bad.example.org") {
		t.Error("DSN 报告缺少第一个失败收件人")
	}
	if !strings.Contains(reportStr, "Final-Recipient: rfc822; user2@down.example.net") {
		t.Error("DSN 报告缺少第二个失败收件人")
	}
	if !strings.Contains(reportStr, "Status: 5.1.1") {
		t.Error("DSN 报告缺少第一个收件人的状态码")
	}
	if !strings.Contains(reportStr, "Status: 4.4.4") {
		t.Error("DSN 报告缺少第二个收件人的状态码")
	}
}

// TestGenerateFailureDSN_RetHeaders 验证默认 RET=HDRS 行为：
// 第三部分使用 text/rfc822-headers Content-Type。
func TestGenerateFailureDSN_RetHeaders(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")
	headers.Set("Subject", "Test Subject")

	failedRcpts := []FailedRecipient{
		{Address: "nosuch@example.org", DSNStatus: "5.0.0"},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil, // mailOpts 为 nil，默认 RET=HDRS
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	if !strings.Contains(reportStr, "Content-Type: text/rfc822-headers") {
		t.Error("默认 RET=HDRS 时应使用 text/rfc822-headers")
	}
	if strings.Contains(reportStr, "Content-Type: message/rfc822") {
		t.Error("默认 RET=HDRS 时不应使用 message/rfc822")
	}

	// 验证原始邮件头出现在第三部分
	if !strings.Contains(reportStr, "Subject: Test Subject") {
		t.Error("RET=HDRS 时原始邮件头应出现在第三部分")
	}
}

// TestGenerateFailureDSN_RetFull 验证 RET=FULL 行为：
// 第三部分使用 message/rfc822 Content-Type，包含完整邮件内容。
func TestGenerateFailureDSN_RetFull(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")

	fullBody := []byte("From: sender@example.com\r\nSubject: Test\r\n\r\nBody content here")

	mailOpts := &smtp.MailOptions{
		Return: smtp.DSNReturnFull,
	}

	failedRcpts := []FailedRecipient{
		{Address: "nosuch@example.org", DSNStatus: "5.0.0"},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		mailOpts,
		"example.com",
		fullBody,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	if !strings.Contains(reportStr, "Content-Type: message/rfc822") {
		t.Error("RET=FULL 时应使用 message/rfc822")
	}
	if !strings.Contains(reportStr, "Body content here") {
		t.Error("RET=FULL 时应包含完整邮件正文")
	}
}

// TestGenerateFailureDSN_EnvelopeID 验证 ENVID 参数：
// DSN 报告应包含 Original-Envelope-Id 头。
func TestGenerateFailureDSN_EnvelopeID(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")

	mailOpts := &smtp.MailOptions{
		EnvelopeID: "test-envelope-123",
	}

	failedRcpts := []FailedRecipient{
		{Address: "nosuch@example.org", DSNStatus: "5.0.0"},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		mailOpts,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	if !strings.Contains(reportStr, "Original-Envelope-Id: test-envelope-123") {
		t.Error("DSN 报告缺少 Original-Envelope-Id 头")
	}
}

// TestGenerateFailureDSN_OriginalRecipient 验证 ORCPT 参数：
// DSN 报告的 delivery-status 部分应包含 Original-Recipient 字段。
func TestGenerateFailureDSN_OriginalRecipient(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")

	rcptOpts := &smtp.RcptOptions{
		OriginalRecipient:     "original@example.org",
		OriginalRecipientType: "rfc822",
	}

	failedRcpts := []FailedRecipient{
		{
			Address:   "nosuch@example.org",
			DSNStatus: "5.1.1",
			RcptOpts:  rcptOpts,
		},
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	if !strings.Contains(reportStr, "Original-Recipient: rfc822; original@example.org") {
		t.Error("DSN 报告缺少 Original-Recipient 字段")
	}
}

// TestWantsFailureDSN 验证 WantsFailureDSN 对不同 NOTIFY 参数的行为：
// - nil/空 NOTIFY → true（默认发送失败通知）
// - NOTIFY=FAILURE → true
// - NOTIFY=NEVER → false
// - NOTIFY=SUCCESS → false（未包含 FAILURE）
func TestWantsFailureDSN(t *testing.T) {
	tests := []struct {
		name     string
		opts     *smtp.RcptOptions
		expected bool
	}{
		{
			name:     "nil 选项应返回 true（默认发送失败通知）",
			opts:     nil,
			expected: true,
		},
		{
			name:     "空 Notify 应返回 true（默认发送失败通知）",
			opts:     &smtp.RcptOptions{},
			expected: true,
		},
		{
			name: "NOTIFY=FAILURE 应返回 true",
			opts: &smtp.RcptOptions{
				Notify: []smtp.DSNNotify{smtp.DSNNotifyFailure},
			},
			expected: true,
		},
		{
			name: "NOTIFY=NEVER 应返回 false",
			opts: &smtp.RcptOptions{
				Notify: []smtp.DSNNotify{smtp.DSNNotifyNever},
			},
			expected: false,
		},
		{
			name: "NOTIFY=SUCCESS 应返回 false（未包含 FAILURE）",
			opts: &smtp.RcptOptions{
				Notify: []smtp.DSNNotify{smtp.DSNNotifySuccess},
			},
			expected: false,
		},
		{
			name: "NOTIFY=SUCCESS,FAILURE 应返回 true",
			opts: &smtp.RcptOptions{
				Notify: []smtp.DSNNotify{smtp.DSNNotifySuccess, smtp.DSNNotifyFailure},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WantsFailureDSN(tt.opts)
			if result != tt.expected {
				t.Errorf("WantsFailureDSN() = %v, 期望 %v", result, tt.expected)
			}
		})
	}
}

// TestParseEmailAddress 验证邮件地址解析函数。
func TestParseEmailAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user@example.com", "user@example.com"},
		{"<user@example.com>", "user@example.com"},
		{"User Name <user@example.com>", "user@example.com"},
		{`"User Name" <user@example.com>`, "user@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseEmailAddress(tt.input)
			if result != tt.expected {
				t.Errorf("ParseEmailAddress(%q) = %q, 期望 %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestIsASCII 验证 ASCII 检测函数。
func TestIsASCII(t *testing.T) {
	if !isASCII("hello@example.com") {
		t.Error("纯 ASCII 字符串应返回 true")
	}
	if isASCII("hello@例え.jp") {
		t.Error("包含非 ASCII 字符的字符串应返回 false")
	}
}

// TestGenerateFailureDSN_DefaultStatus 验证：
// 当 DSNStatus 为空时，默认使用 "5.0.0" 通用永久失败状态码。
func TestGenerateFailureDSN_DefaultStatus(t *testing.T) {
	headers := make(textproto.MIMEHeader)
	headers.Set("From", "<sender@example.com>")

	failedRcpts := []FailedRecipient{
		{Address: "nosuch@example.org"}, // DSNStatus 为空
	}

	report, err := GenerateFailureDSN(
		"sender@example.com",
		headers,
		failedRcpts,
		nil,
		"example.com",
		nil,
	)
	if err != nil {
		t.Fatalf("GenerateFailureDSN 返回错误: %v", err)
	}

	reportStr := string(report)

	if !strings.Contains(reportStr, "Status: 5.0.0") {
		t.Error("DSNStatus 为空时应使用默认状态码 5.0.0")
	}
}
