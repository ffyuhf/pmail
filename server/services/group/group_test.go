package group

import (
	"fmt"
	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/utils/context"
	"testing"
)

func TestGetGroupStatus(t *testing.T) {
	config.Init()
	db.Init("")
	db.Instance.ShowSQL(true)
	ctx := &context.Context{
		UserID:      1,
		UserName:    "admin",
		UserAccount: "admin",
	}
	ret, _ := GetGroupStatus(ctx, "INBOX", []string{"MESSAGES", "UIDNEXT", "UIDVALIDITY", "UNSEEN"})
	fmt.Println(ret)
}
