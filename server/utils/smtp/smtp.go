// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package smtp 实现了 RFC 5321 定义的简单邮件传输协议（SMTP）。
// 同时实现了以下扩展：
//
//	8BITMIME  RFC 1652
//	AUTH      RFC 2554
//	STARTTLS  RFC 3207
//
// 其他扩展可由客户端自行处理。
//
// smtp 包已冻结，不再接受新功能。
// 一些外部包提供了更多功能，参见：
//
//	https://godoc.org/?q=smtp
//
// 在go原始SMTP协议的基础上修复了TLS验证错误、支持了SMTPS协议、 支持自定义HELLO命令的域名信息
package smtp

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"io"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"github.com/ffyuhf/pmail/config"
	log "github.com/sirupsen/logrus"
)

var NoSupportSTARTTLSError = errors.New("smtp: server doesn't support STARTTLS")
var EOFError = errors.New("EOF")

// Client 表示一个到 SMTP 服务器的客户端连接。
type Client struct {
	// Text 是 Client 使用的 textproto.Conn。导出以允许客户端添加扩展。
	Text *textproto.Conn
	// 保持连接引用，以便稍后创建 TLS 连接
	conn net.Conn
	// 客户端是否使用 TLS
	tls        bool
	serverName string
	// 支持的扩展映射
	ext map[string]string
	// 支持的认证机制
	auth       []string
	localName  string // 在 HELO/EHLO 中使用的名称
	didHello   bool   // 是否已发送 HELO/EHLO
	helloError error  // hello 过程中的错误
}

// Dial 返回一个连接到指定地址 SMTP 服务器的新 Client。
// addr 必须包含端口，例如 "mail.example.com:smtp"。
func Dial(addr, fromDomain string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host, fromDomain)
}

// 使用 TLS 加密连接
func DialTls(addr, domain, fromDomain string) (*Client, error) {
	if domain == "" {
		domain = fromDomain
	}

	// TLS 配置
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         domain,
	}

	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 2 * time.Second}, "tcp", addr, tlsconfig)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(addr)
	return NewClient(conn, host, fromDomain)
}

// NewClient 使用现有连接创建新 Client，host 作为认证时使用的服务器名称。
func NewClient(conn net.Conn, host, fromDomain string) (*Client, error) {
	text := textproto.NewConn(conn)
	_, _, err := text.ReadResponse(220)
	if err != nil {
		text.Close()
		return nil, err
	}

	localName := "domain.com"

	if fromDomain != "" {
		localName = fromDomain
	} else if config.Instance != nil && config.Instance.Domain != "" {
		localName = config.Instance.Domain
	}

	c := &Client{Text: text, conn: conn, serverName: host, localName: localName}
	_, c.tls = conn.(*tls.Conn)
	return c, nil
}

// Close 关闭连接。
func (c *Client) Close() error {
	return c.Text.Close()
}

// hello 在需要时执行 hello 交换。
func (c *Client) hello() error {
	if !c.didHello {
		c.didHello = true
		err := c.ehlo()
		if err != nil {
			c.helloError = c.helo()
		}
	}
	return c.helloError
}

// Hello 以给定的主机名向服务器发送 HELO 或 EHLO。
// 仅在客户端需要控制所使用的主机名时才需要调用此方法。
// 否则客户端会自动以 "localhost" 自我介绍。
// 如果调用 Hello，必须在调用其他任何方法之前调用。
func (c *Client) Hello(localName string) error {
	if err := validateLine(localName); err != nil {
		return err
	}
	if c.didHello {
		return errors.New("smtp: Hello called after other methods")
	}
	c.localName = localName
	return c.hello()
}

// cmd 是一个便捷函数，发送命令并返回响应
func (c *Client) cmd(expectCode int, format string, args ...any) (int, string, error) {
	id, err := c.Text.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	code, msg, err := c.Text.ReadResponse(expectCode)
	return code, msg, err
}

// helo 向服务器发送 HELO 问候。仅当服务器不支持 EHLO 时使用。
func (c *Client) helo() error {
	c.ext = nil
	_, _, err := c.cmd(250, "HELO %s", c.localName)
	return err
}

// ehlo 向服务器发送 EHLO（扩展问候）。对于支持 EHLO 的服务器应优先使用。
func (c *Client) ehlo() error {
	_, msg, err := c.cmd(250, "EHLO %s", c.localName)
	if err != nil {
		return err
	}
	ext := make(map[string]string)
	extList := strings.Split(msg, "\n")
	if len(extList) > 1 {
		extList = extList[1:]
		for _, line := range extList {
			k, v, _ := strings.Cut(line, " ")
			ext[k] = v
		}
	}
	if mechs, ok := ext["AUTH"]; ok {
		c.auth = strings.Split(mechs, " ")
	}
	c.ext = ext
	return err
}

// StartTLS 发送 STARTTLS 命令并加密后续所有通信。
// 仅支持 STARTTLS 扩展的服务器才能使用此功能。
func (c *Client) StartTLS(config *tls.Config) error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(220, "STARTTLS")
	if err != nil {
		return err
	}
	if config == nil {
		config = &tls.Config{}
	}
	if config.ServerName == "" {
		// 复制一份以避免污染参数
		config = config.Clone()
		config.ServerName = c.serverName
	}
	c.conn = tls.Client(c.conn, config)
	c.Text = textproto.NewConn(c.conn)
	c.tls = true
	return c.ehlo()
}

// TLSConnectionState 返回客户端的 TLS 连接状态。
// 如果 StartTLS 未成功，返回值为零值。
func (c *Client) TLSConnectionState() (state tls.ConnectionState, ok bool) {
	tc, ok := c.conn.(*tls.Conn)
	if !ok {
		return
	}
	return tc.ConnectionState(), true
}

// Verify 检查邮件地址在服务器上的有效性。
// 如果返回 nil，表示地址有效。非 nil 返回不一定表示地址无效。
// 许多服务器出于安全原因不会验证地址。
func (c *Client) Verify(addr string) error {
	if err := validateLine(addr); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "VRFY %s", addr)
	return err
}

// Auth 使用提供的认证机制对客户端进行认证。
// 认证失败会关闭连接。
// 仅支持 AUTH 扩展的服务器才能使用此功能。
func (c *Client) Auth(a smtp.Auth) error {
	if err := c.hello(); err != nil {
		return err
	}
	encoding := base64.StdEncoding
	mech, resp, err := a.Start(&smtp.ServerInfo{Name: c.serverName, TLS: c.tls, Auth: c.auth})
	if err != nil {
		c.Quit()
		return err
	}
	resp64 := make([]byte, encoding.EncodedLen(len(resp)))
	encoding.Encode(resp64, resp)
	code, msg64, err := c.cmd(0, "AUTH %s %s", mech, resp64)
	for err == nil {
		var msg []byte
		switch code {
		case 334:
			msg, err = encoding.DecodeString(msg64)
		case 235:
			// 最后一条消息不是 base64 编码，因为它不是挑战
			msg = []byte(msg64)
		default:
			err = &textproto.Error{Code: code, Msg: msg64}
		}
		if err == nil {
			resp, err = a.Next(msg, code == 334)
		}
		if err != nil {
			// 中止 AUTH
			c.cmd(501, "*")
			c.Quit()
			break
		}
		if resp == nil {
			break
		}
		resp64 = make([]byte, encoding.EncodedLen(len(resp)))
		encoding.Encode(resp64, resp)
		code, msg64, err = c.cmd(0, "%s", string(resp64))
	}
	return err
}

// Mail 使用提供的邮件地址向服务器发送 MAIL 命令。
// 如果服务器支持 8BITMIME 扩展，Mail 会添加 BODY=8BITMIME 参数。
// 如果服务器支持 SMTPUTF8 扩展，Mail 会添加 SMTPUTF8 参数。
// 此操作启动邮件事务，之后跟随一个或多个 Rcpt 调用。
func (c *Client) Mail(from string) error {
	if err := validateLine(from); err != nil {
		return err
	}
	if err := c.hello(); err != nil {
		return err
	}
	cmdStr := "MAIL FROM:<%s>"
	if c.ext != nil {
		if _, ok := c.ext["8BITMIME"]; ok {
			cmdStr += " BODY=8BITMIME"
		}
		if _, ok := c.ext["SMTPUTF8"]; ok {
			cmdStr += " SMTPUTF8"
		}
	}
	_, _, err := c.cmd(250, cmdStr, from)
	return err
}

// Rcpt 使用提供的邮件地址向服务器发送 RCPT 命令。
// 调用 Rcpt 前必须先调用 Mail，之后可调用 Data 或再次调用 Rcpt。
func (c *Client) Rcpt(to string) error {
	if err := validateLine(to); err != nil {
		return err
	}
	_, _, err := c.cmd(25, "RCPT TO:<%s>", to)
	return err
}

type dataCloser struct {
	c *Client
	io.WriteCloser
}

func (d *dataCloser) Close() error {
	d.WriteCloser.Close()
	_, _, err := d.c.Text.ReadResponse(250)
	return err
}

// Data 向服务器发送 DATA 命令并返回一个可用于写入邮件头和正文的 writer。
// 调用者应在调用 c 的其他方法之前关闭 writer。
// 调用 Data 前必须先调用一次或多次 Rcpt。
func (c *Client) Data() (io.WriteCloser, error) {
	_, _, err := c.cmd(354, "DATA")
	if err != nil {
		return nil, err
	}
	return &dataCloser{c, c.Text.DotWriter()}, nil
}

func SendMailWithTls(domain string, addr string, a smtp.Auth, from string, fromDomain string, to []string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := DialTls(addr, domain, fromDomain)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// SendMail 连接到指定地址的服务器，尽可能切换到 TLS，
// 尽可能使用可选机制 a 进行认证，然后发送邮件。
// addr 必须包含端口，例如 "mail.example.com:smtp"。
//
// to 参数中的地址是 SMTP RCPT 地址。
//
// msg 参数应为 RFC 822 格式的邮件，先是头部，然后是空行，最后是邮件正文。
// msg 的行应以 CRLF 结尾。msg 头部通常应包含 "From"、"To"、"Subject"、"Cc" 等字段。
// 发送 "Bcc" 邮件的方式是将地址包含在 to 参数中但不包含在 msg 头部中。
//
// SendMail 函数和 net/smtp 包是低级机制，不支持 DKIM 签名、
// MIME 附件（参见 mime/multipart 包）或其他邮件功能。
// 标准库之外存在更高级的包。
// 修复TSL验证问题
func SendMail(domain string, addr string, a smtp.Auth, from string, fromDomain string, to []string, msg []byte) error {

	log.Debugf("SendMail,%s ,%s ,%s ,%s ,%v ", domain, addr, from, fromDomain, to)

	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := Dial(addr, fromDomain)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}
	if ok, _ := c.Extension("STARTTLS"); !ok {
		return NoSupportSTARTTLSError
	}

	var config *tls.Config
	if domain != "" {
		config = &tls.Config{
			ServerName: domain,
		}
	}

	if err = c.StartTLS(config); err != nil {
		return err
	}
	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// SendMailUnsafe 无TLS加密的邮件发送方式
func SendMailUnsafe(domain string, addr string, a smtp.Auth, from string, fromDomain string, to []string, msg []byte) error {
	if err := validateLine(from); err != nil {
		return err
	}
	for _, recp := range to {
		if err := validateLine(recp); err != nil {
			return err
		}
	}
	c, err := Dial(addr, fromDomain)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.hello(); err != nil {
		return err
	}

	if a != nil && c.ext != nil {
		if _, ok := c.ext["AUTH"]; !ok {
			return errors.New("smtp: server doesn't support AUTH")
		}
		if err = c.Auth(a); err != nil {
			return err
		}
	}
	if err = c.Mail(from); err != nil {
		return err
	}
	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(msg)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}

// Extension 报告服务器是否支持某个扩展。
// 扩展名称不区分大小写。如果扩展被支持，
// Extension 还会返回一个包含服务器为该扩展指定的参数的字符串。
func (c *Client) Extension(ext string) (bool, string) {
	if err := c.hello(); err != nil {
		return false, ""
	}
	if c.ext == nil {
		return false, ""
	}
	ext = strings.ToUpper(ext)
	param, ok := c.ext[ext]
	return ok, param
}

// Reset 向服务器发送 RSET 命令，中止当前邮件事务。
func (c *Client) Reset() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "RSET")
	return err
}

// Noop 向服务器发送 NOOP 命令。仅检查与服务器的连接是否正常。
func (c *Client) Noop() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(250, "NOOP")
	return err
}

// Quit 发送 QUIT 命令并关闭与服务器的连接。
func (c *Client) Quit() error {
	if err := c.hello(); err != nil {
		return err
	}
	_, _, err := c.cmd(221, "QUIT")
	if err != nil {
		return err
	}
	return c.Text.Close()
}

// validateLine 检查行是否包含 CR 或 LF，依据 RFC 5321。
func validateLine(line string) error {
	if strings.ContainsAny(line, "\n\r") {
		return errors.New("smtp: A line must not contain CR or LF")
	}
	return nil
}
