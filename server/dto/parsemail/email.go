package parsemail

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/microcosm-cc/bluemonday"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type User struct {
	EmailAddress string `json:"EmailAddress"`
	Name         string `json:"Name"`
}

func (u User) Build() string {
	if u.Name != "" {
		return fmt.Sprintf("\"%s\" <%s>", mime.QEncoding.Encode("utf-8", u.Name), u.EmailAddress)
	}
	return fmt.Sprintf("<%s>", u.EmailAddress)
}

func (u User) GetDomainAccount() (string, string) {
	infos := strings.Split(u.EmailAddress, "@")
	if len(infos) >= 2 {
		return infos[0], infos[1]
	}

	return "", ""
}

type Attachment struct {
	Filename    string
	ContentType string
	Content     []byte
	ContentID   string
}

// Email 邮件消息类型
type Email struct {
	ReplyTo     []*User
	From        *User
	To          []*User
	Bcc         []*User
	Cc          []*User
	Subject     string
	Text        []byte // 纯文本消息（可选）
	HTML        []byte // HTML 消息（可选）
	Sender      *User  // 覆盖 From 作为 SMTP 信封发件人（可选）
	Headers     textproto.MIMEHeader
	Attachments []*Attachment
	ReadReceipt []string
	Date        string
	Status      int // 0未发送，1已发送，2发送失败，3删除，5广告邮件
	MessageId   int64
	MsgID       string // 符合 RFC 规范的 Message-ID，持久化存储在数据库中
	Size        int
}

// GenerateMsgID 生成符合 RFC 规范的 Message-ID，具有足够唯一性以避免被垃圾邮件过滤器拦截。
// 必须在邮件创建时调用一次，并将结果持久化到数据库。
func GenerateMsgID(domain string) string {
	randBytes := make([]byte, 16)
	_, _ = rand.Read(randBytes)
	return fmt.Sprintf("%x.%d@%s", randBytes, time.Now().UnixNano(), domain)
}

// XSS 过滤策略
var (
	strictPolicy  *bluemonday.Policy
	relaxedPolicy *bluemonday.Policy
)

func init() {
	strictPolicy = bluemonday.StrictPolicy()

	relaxedPolicy = bluemonday.NewPolicy()

	relaxedPolicy.AllowElements("p", "br", "strong", "em", "u", "b", "i", "h1", "h2", "h3", "h4", "h5", "h6")
	relaxedPolicy.AllowElements("div", "span", "center")
	relaxedPolicy.AllowElements("ul", "ol", "li")
	relaxedPolicy.AllowElements("blockquote", "cite")

	relaxedPolicy.AllowElements("table", "tbody", "thead", "tr", "td", "th")
	relaxedPolicy.AllowAttrs("width", "height", "border", "cellpadding", "cellspacing").OnElements("table")
	relaxedPolicy.AllowAttrs("align", "valign", "colspan", "rowspan").OnElements("td", "th")
	relaxedPolicy.AllowAttrs("align").OnElements("tr")

	relaxedPolicy.AllowAttrs("style").Globally()
	relaxedPolicy.AllowAttrs("class", "id").Globally()

	relaxedPolicy.AllowAttrs("bgcolor", "color", "background").Globally()
	relaxedPolicy.AllowAttrs("align").OnElements("p", "div", "h1", "h2", "h3", "h4", "h5", "h6")

	relaxedPolicy.AllowElements("img")
	relaxedPolicy.AllowAttrs("src", "alt", "width", "height", "style", "align").OnElements("img")

	relaxedPolicy.AllowElements("a")
	relaxedPolicy.AllowAttrs("href", "style").OnElements("a")
	relaxedPolicy.RequireNoReferrerOnLinks(true)
	relaxedPolicy.AddTargetBlankToFullyQualifiedLinks(true)
	relaxedPolicy.RequireNoFollowOnLinks(true)

	relaxedPolicy.AllowElements("font")
	relaxedPolicy.AllowAttrs("size", "color", "face").OnElements("font")

	relaxedPolicy.AllowElements("style")
	relaxedPolicy.AllowAttrs("type").OnElements("style")

	// 邮件中常见的语义化和格式化 HTML 元素
	// 修复：扩展允许的元素列表，避免复杂 HTML 邮件结构被过度剥离
	relaxedPolicy.AllowElements("section", "article", "header", "footer", "main", "nav")
	relaxedPolicy.AllowElements("aside", "figure", "figcaption")
	relaxedPolicy.AllowElements("hr", "pre", "code", "mark", "time", "wbr")
	relaxedPolicy.AllowElements("abbr", "address", "cite", "q")
	relaxedPolicy.AllowElements("ins", "del", "s", "sub", "sup")
	relaxedPolicy.AllowElements("dl", "dt", "dd")
	relaxedPolicy.AllowElements("details", "summary")
	relaxedPolicy.AllowAttrs("cite").OnElements("blockquote", "q", "ins", "del")
	relaxedPolicy.AllowAttrs("datetime").OnElements("time", "ins", "del")
	relaxedPolicy.AllowAttrs("title").Globally()

	relaxedPolicy.AllowURLSchemes("http", "https", "mailto")

	relaxedPolicy.SkipElementsContent("script", "object", "embed", "iframe", "frame", "frameset")
}

// stripHTMLTags 剥离所有 HTML 标签并保留原始文本字符。
// 实现方式：先由 bluemonday.StrictPolicy 完成健壮的 XSS 过滤（剥离标签、属性等），
// 再通过 html.UnescapeString 将 HTML 实体还原为原始字符（如 &#39; → '）。
// 适用于邮件标题、发件人名称等纯文本字段。
// 修复：2026-04-29，原直接使用 strictPolicy.Sanitize() 会导致纯文本中的引号被编码为 HTML 实体，
// 例如 Ollama's cloud 被错误保存为 Ollama&#39;s cloud
func stripHTMLTags(text string) string {
	sanitized := strictPolicy.Sanitize(text)
	return html.UnescapeString(sanitized)
}

func sanitizeHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	sanitized := relaxedPolicy.Sanitize(htmlContent)

	dataUrlRegex := regexp.MustCompile(`href\s*=\s*["']data:[^"']*["']`)
	sanitized = dataUrlRegex.ReplaceAllString(sanitized, `rel="nofollow"`)

	jsUrlRegex := regexp.MustCompile(`href\s*=\s*["']javascript:[^"']*["']`)
	sanitized = jsUrlRegex.ReplaceAllString(sanitized, `rel="nofollow"`)

	expressionRegex := regexp.MustCompile(`(?i)expression\s*\(.*?\)`)
	sanitized = expressionRegex.ReplaceAllString(sanitized, "")

	styleExpressionRegex := regexp.MustCompile(`(?i)style\s*=\s*["'][^"']*expression[^"']*["']`)
	sanitized = styleExpressionRegex.ReplaceAllString(sanitized, "")

	cssJsRegex := regexp.MustCompile(`(?i)javascript\s*:`)
	sanitized = cssJsRegex.ReplaceAllString(sanitized, "")

	return sanitized
}

// SanitizeText 清洗文本
func sanitizeText(text string) string {
	return strictPolicy.Sanitize(text)
}

func users2String(users []*User) string {
	ret := ""
	for _, user := range users {
		if ret != "" {
			ret += ", "
		}
		ret += user.Build()
	}
	return ret
}

func (e *Email) BuildTo2String() string {
	return users2String(e.To)
}

func (e *Email) BuildCc2String() string {
	return users2String(e.Cc)
}

func (e *Email) BuildBcc2String() string {
	return users2String(e.Bcc)
}

func NewEmailFromModel(d models.Email) *Email {

	var To []*User
	json.Unmarshal([]byte(d.To), &To)

	var ReplyTo []*User
	json.Unmarshal([]byte(d.ReplyTo), &ReplyTo)

	var Sender *User
	json.Unmarshal([]byte(d.Sender), &Sender)

	var Bcc []*User
	json.Unmarshal([]byte(d.Bcc), &Bcc)

	var Cc []*User
	json.Unmarshal([]byte(d.Cc), &Cc)

	var Attachments []*Attachment
	json.Unmarshal([]byte(d.Attachments), &Attachments)

	return &Email{
		MessageId: cast.ToInt64(d.Id),
		MsgID:     d.MsgID,
		From: &User{
			Name:         d.FromName,
			EmailAddress: d.FromAddress,
		},
		To:          To,
		Subject:     d.Subject,
		Text:        []byte(d.Text.String),
		HTML:        []byte(d.Html.String),
		Sender:      Sender,
		ReplyTo:     ReplyTo,
		Bcc:         Bcc,
		Cc:          Cc,
		Attachments: Attachments,
		Date:        d.SendDate.Format("2006-01-02 15:04:05"),
	}
}

func NewEmailFromReader(to []string, r io.Reader, size int) *Email {
	ret := &Email{}
	m, err := message.Read(r)
	if err != nil {
		log.Errorf("email解析错误！ Error %+v", err)
	}

	ret.Size = size
	// 保留发件人的原始 Message-ID，确保存储和复用的一致性。
	if mid := m.Header.Get("Message-Id"); mid != "" {
		ret.MsgID = strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(mid), ">"), "<")
	}
	ret.From = buildUser(m.Header.Get("From"))

	smtpTo := buildUsers(to)

	ret.To = buildUsers(m.Header.Values("To"))

	ret.Bcc = []*User{}

	for _, user := range smtpTo {
		in := false
		for _, u := range ret.To {
			if u.EmailAddress == user.EmailAddress {
				in = true
				break
			}
		}
		if !in {
			ret.Bcc = append(ret.Bcc, user)
		}

	}

	ret.Cc = buildUsers(m.Header.Values("Cc"))
	ret.ReplyTo = buildUsers(m.Header.Values("ReplyTo"))
	ret.Sender = buildUser(m.Header.Get("Sender"))
	if ret.Sender == nil {
		ret.Sender = ret.From
	}

	// 修复：2026-04-29 使用 stripHTMLTags 替代 strictPolicy.Sanitize
	// 原因：StrictPolicy 会将 ' 编码为 &#39; 等实体，导致邮件标题在所有客户端中显示异常
	subject, _ := m.Header.Text("Subject")
	ret.Subject = stripHTMLTags(subject)

	sendTime, err := time.Parse(time.RFC1123Z, m.Header.Get("Date"))
	if err != nil {
		sendTime = time.Now()
	}
	ret.Date = sendTime.Format(time.DateTime)
	m.Walk(func(path []int, entity *message.Entity, err error) error {
		return formatContent(entity, ret)
	})

	// 修复：2026-04-29 使用 stripHTMLTags 替代 strictPolicy.Sanitize，避免发件人信息过度编码
	if ret.From != nil {
		ret.From.Name = stripHTMLTags(ret.From.Name)
		ret.From.EmailAddress = stripHTMLTags(ret.From.EmailAddress)
	}

	return ret
}

func formatContent(entity *message.Entity, ret *Email) error {
	contentType, p, err := entity.Header.ContentType()

	if err != nil {
		log.Errorf("email read error! %+v", err)
		return err
	}

	switch contentType {
	case "multipart/alternative":
	case "multipart/mixed":
	case "text/plain":
		// 修复：2026-04-29 使用 stripHTMLTags 替代 strictPolicy.Sanitize，避免纯文本过度编码
		testContent, _ := io.ReadAll(entity.Body)
		ret.Text = []byte(stripHTMLTags(string(testContent)))
	case "text/html":
		htmlContent, _ := io.ReadAll(entity.Body)
		ret.HTML = []byte(relaxedPolicy.Sanitize(string(htmlContent)))
	// multipart/related 的子节点由外层 Walk 自然遍历处理
	// 修复：移除内层 entity.Walk() 调用，避免双重遍历导致 Body 被消耗后覆盖为空内容
	// 原因：外层 m.Walk() 会遍历到 multipart/related 的每个子节点（text/html, image/png 等），
	// 如果在这里再调用 entity.Walk()，同一个实体的 Body 会被读取两次，
	// 第二次读取时 io.ReadAll(entity.Body) 返回空内容，覆盖了第一次正确读取的 HTML
	case "multipart/related":
	default:
		c, _ := io.ReadAll(entity.Body)
		fileName := p["name"]
		if fileName == "" {
			contentDisposition := entity.Header.Get("Content-Disposition")
			filenameRegex := regexp.MustCompile(`filename\s*=\s*"?([^";]+)"?`)
			matches := filenameRegex.FindStringSubmatch(contentDisposition)
			if len(matches) >= 2 {
				fileName = strings.TrimSpace(matches[1])
				fileName = strings.Trim(fileName, `"`)
			} else {
				fileName = "no_name_file"
			}
		}

		ret.Attachments = append(ret.Attachments, &Attachment{
			Filename:    sanitizeText(fileName),
			ContentType: sanitizeText(strings.TrimSpace(contentType)),
			Content:     c,
			ContentID:   strings.TrimPrefix(strings.TrimSuffix(entity.Header.Get("Content-Id"), ">"), "<"),
		})
	}

	return nil
}

func BuilderUser(str string) *User {
	return buildUser(str)
}

var emailAddressRe = regexp.MustCompile(`<(.*@.*)>`)

func buildUser(str string) *User {
	str = strings.TrimSpace(str)
	if str == "" {
		return &User{}
	}

	user := &User{}

	addr, err := mail.ParseAddress(str)
	if err == nil {
		user.EmailAddress = strings.TrimSpace(addr.Address)

		name := strings.TrimSpace(addr.Name)
		if name != "" {
			decoder := mime.WordDecoder{}
			if decoded, err := decoder.Decode(name); err == nil {
				name = decoded
			}
			// 修复：2026-04-29 使用 stripHTMLTags 替代 strictPolicy.Sanitize
			user.Name = stripHTMLTags(name)
		}
		return user
	}

	matched := emailAddressRe.FindStringSubmatch(str)
	if len(matched) == 2 {
		user.EmailAddress = strings.TrimSpace(matched[1])
		namePart := strings.ReplaceAll(str, matched[0], "")
		namePart = strings.Trim(strings.TrimSpace(namePart), "\"")

		// 修复：2026-04-29 使用 stripHTMLTags 替代 strictPolicy.Sanitize
		decoder := mime.WordDecoder{}
		if decoded, err := decoder.Decode(strings.ReplaceAll(namePart, "\"", "")); err == nil {
			user.Name = stripHTMLTags(strings.TrimSpace(decoded))
		} else {
			user.Name = stripHTMLTags(strings.TrimSpace(namePart))
		}
	} else {
		user.EmailAddress = stripHTMLTags(str)
	}

	return user
}

func buildUsers(strs []string) []*User {
	var ret []*User
	for _, line := range strs {
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ",")
		for _, part := range parts {
			if u := buildUser(strings.TrimSpace(part)); u != nil {
				ret = append(ret, u)
			}
		}
	}
	return ret
}

func (e *Email) ForwardBuildBytes(ctx *context.Context, sender *models.User) []byte {
	var b bytes.Buffer

	from := []*mail.Address{{e.From.Name, e.From.EmailAddress}}
	to := []*mail.Address{}
	for _, user := range e.To {
		to = append(to, &mail.Address{
			Name:    user.Name,
			Address: user.EmailAddress,
		})
	}

	senderAddress := []*mail.Address{{sender.Name, fmt.Sprintf("%s@%s", sender.Account, config.Instance.Domains[0])}}
	// 创建邮件头
	var h mail.Header
	h.SetDate(time.Now())
	h.SetAddressList("From", from)
	h.SetAddressList("Sender", senderAddress)
	h.SetAddressList("To", to)
	h.SetText("Subject", e.Subject)
	if e.MsgID != "" {
		h.SetMessageID(e.MsgID)
	} else {
		h.SetMessageID(fmt.Sprintf("%d@%s", e.MessageId, config.Instance.Domain))
	}
	if len(e.Cc) != 0 {
		cc := []*mail.Address{}
		for _, user := range e.Cc {
			cc = append(cc, &mail.Address{
				Name:    user.Name,
				Address: user.EmailAddress,
			})
		}
		h.SetAddressList("Cc", cc)
	}

	// 创建邮件写入器
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// 创建文本部分
	tw, err := mw.CreateInline()
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain; charset=UTF-8")
	// 修复：统一设置 Content-Transfer-Encoding，与 BuildBytes 保持一致
	th.Header.Set("Content-Transfer-Encoding", "base64")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, string(e.Text))
	w.Close()

	var html mail.InlineHeader
	html.Set("Content-Type", "text/html; charset=UTF-8")
	// 修复：统一设置 Content-Transfer-Encoding，与 BuildBytes 保持一致
	html.Header.Set("Content-Transfer-Encoding", "base64")
	w, err = tw.CreatePart(html)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w, string(e.HTML))
	w.Close()

	tw.Close()

	// 创建附件
	for _, attachment := range e.Attachments {
		var ah mail.AttachmentHeader
		ah.Set("Content-Type", attachment.ContentType)
		ah.SetFilename(attachment.Filename)
		w, err = mw.CreateAttachment(ah)
		if err != nil {
			log.WithContext(ctx).Fatal(err)
			continue
		}
		w.Write(attachment.Content)
		w.Close()
	}

	mw.Close()

	// dkim 签名后返回
	return instance.Sign(b.String())
}

func (e *Email) BuildPart(ctx *context.Context, loc []int) []byte {
	if len(loc) == 0 {
		return nil
	}

	// 处理顶层 part (part 1 = alternative, part 2+ = attachments)
	if len(loc) == 1 {
		partIdx := loc[0]
		if partIdx == 1 {
			// Part 1 是 alternative，不能直接获取，需要获取子部分
			return nil
		}
		// Part 2, 3, ... 是附件
		attachIdx := partIdx - 2
		if attachIdx >= 0 && attachIdx < len(e.Attachments) {
			encoded := base64.StdEncoding.EncodeToString(e.Attachments[attachIdx].Content)
			encoded += "\r\n"
			return []byte(encoded)
		}
		return nil
	}

	// 处理 alternative 的子部分 (1.1, 1.2)
	if loc[0] == 1 && len(loc) >= 2 {
		subIdx := loc[1]

		// 根据 BODYSTRUCTURE 的构建顺序：先 text，后 html
		// 如果只有一个存在，那个就是 1.1
		hasText := len(e.Text) > 0
		hasHtml := len(e.HTML) > 0

		if hasText && hasHtml {
			// 两者都有: 1.1=text, 1.2=html
			if subIdx == 1 {
				encoded := base64.StdEncoding.EncodeToString(e.Text)
				encoded += "\r\n"
				return []byte(encoded)
			}
			if subIdx == 2 {
				encoded := base64.StdEncoding.EncodeToString(e.HTML)
				encoded += "\r\n"
				return []byte(encoded)
			}
		} else if hasText {
			// 只有 text: 1.1=text
			if subIdx == 1 {
				encoded := base64.StdEncoding.EncodeToString(e.Text)
				encoded += "\r\n"
				return []byte(encoded)
			}
		} else if hasHtml {
			// 只有 html: 1.1=html
			if subIdx == 1 {
				encoded := base64.StdEncoding.EncodeToString(e.HTML)
				encoded += "\r\n"
				return []byte(encoded)
			}
		}
	}

	return nil
}

func (e *Email) BuildBytes(ctx *context.Context, dkim bool) []byte {
	var b bytes.Buffer

	from := []*mail.Address{{e.From.Name, e.From.EmailAddress}}
	to := []*mail.Address{}
	for _, user := range e.To {
		to = append(to, &mail.Address{
			Name:    user.Name,
			Address: user.EmailAddress,
		})
	}

	// 创建邮件头
	var h mail.Header
	if e.Date != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", e.Date, time.Local)
		if err != nil {
			log.WithContext(ctx).Errorf("Time Error ! Err:%+v", err)
			h.SetDate(time.Now())
		} else {
			h.SetDate(t)
		}
	} else {
		h.SetDate(time.Now())
	}
	if e.MsgID != "" {
		h.SetMessageID(e.MsgID)
	} else {
		h.SetMessageID(fmt.Sprintf("%d@%s", e.MessageId, config.Instance.Domain))
	}
	h.SetAddressList("From", from)
	h.SetAddressList("Sender", from)
	h.SetAddressList("To", to)
	h.SetText("Subject", e.Subject)
	if len(e.Cc) != 0 {
		cc := []*mail.Address{}
		for _, user := range e.Cc {
			cc = append(cc, &mail.Address{
				Name:    user.Name,
				Address: user.EmailAddress,
			})
		}
		h.SetAddressList("Cc", cc)
	}

	// 创建邮件写入器
	mw, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	// 创建文本部分
	tw, err := mw.CreateInline()
	if err != nil {
		log.WithContext(ctx).Fatal(err)
	}

	if len(e.Text) > 0 {
		var th mail.InlineHeader
		th.Header.Set("Content-Transfer-Encoding", "base64")
		th.SetContentType("text/plain", map[string]string{
			"charset": "UTF-8",
		})
		w, err := tw.CreatePart(th)
		if err != nil {
			log.Fatal(err)
		}
		io.WriteString(w, string(e.Text))
		w.Close()
	}

	var html mail.InlineHeader
	html.SetContentType("text/html", map[string]string{
		"charset": "UTF-8",
	})
	html.Header.Set("Content-Transfer-Encoding", "base64")
	w, err := tw.CreatePart(html)
	if err != nil {
		log.Fatal(err)
	}
	if len(e.HTML) > 0 {
		io.WriteString(w, string(e.HTML))
	} else {
		io.WriteString(w, string(e.Text))
	}

	w.Close()

	tw.Close()

	// 创建附件
	for _, attachment := range e.Attachments {
		var ah mail.AttachmentHeader
		ah.Set("Content-Type", attachment.ContentType)
		ah.SetFilename(attachment.Filename)
		w, err = mw.CreateAttachment(ah)
		if err != nil {
			log.WithContext(ctx).Fatal(err)
			continue
		}
		w.Write(attachment.Content)
		w.Close()
	}

	mw.Close()

	if dkim {
		// dkim 签名后返回
		return instance.Sign(b.String())
	}
	return b.Bytes()
}
