package pop3_server

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/Jinnrry/gopop"
	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/emersion/go-message/mail"
)

func Test_action_Retr(t *testing.T) {
	config.Init()
	config.Instance.DbType = config.DBTypeSQLite
	config.Instance.DbDSN = config.ROOT_PATH + "./config/pmail_temp.db"
	db.Init("")

	a := action{}
	session := &gopop.Session{
		Ctx: &context.Context{
			UserID: 1,
		},
	}
	got, got1, err := a.Retr(session, 301)

	_, _, _ = got, got1, err
}

func Test_email(t *testing.T) {
	var b bytes.Buffer

	// 创建邮件头
	var h mail.Header

	// 创建邮件写入器
	mw, _ := mail.CreateWriter(&b, h)

	// 创建文本部分
	tw, _ := mw.CreateInline()

	var html mail.InlineHeader

	html.Header.Set("Content-Transfer-Encoding", "base64")
	w, _ := tw.CreatePart(html)

	io.WriteString(w, "=")

	w.Close()

	tw.Close()

	fmt.Printf("%s", b.String())

}
