package email

import (
	"encoding/json"
	"io"
	"math"
	"net/http"

	"github.com/ffyuhf/pmail/dto"
	"github.com/ffyuhf/pmail/dto/parsemail"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/services/list"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

type emailListResponse struct {
	CurrentPage int         `json:"current_page"`
	TotalPage   int         `json:"total_page"`
	List        []*emilItem `json:"list"`
}

type emilItem struct {
	ID        int    `json:"id"`
	UeId      int    `json:"ue_id"` // user_email 表的主键 ID，用于精确定位记录。修改日期: 20260510
	Type      int8   `json:"type"`  // 邮件类型：0=接收, 1=发送。修改日期: 20260510
	Title     string `json:"title"`
	Desc      string `json:"desc"`
	Datetime  string `json:"datetime"`
	IsRead    bool   `json:"is_read"`
	Sender    User   `json:"sender"`
	To        []User `json:"to"`
	Dangerous bool   `json:"dangerous"`
	Error     string `json:"error"`
}

type User struct {
	Name         string `json:"Name"`
	EmailAddress string `json:"EmailAddress"`
}

type emailRequest struct {
	Keyword     string `json:"keyword"`
	Tag         string `json:"tag"`
	CurrentPage int    `json:"current_page"`
	PageSize    int    `json:"page_size"`
}

func EmailList(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	var lst []*emilItem
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var retData emailRequest
	err = json.Unmarshal(reqBytes, &retData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	offset := 0
	if retData.CurrentPage >= 1 {
		offset = (retData.CurrentPage - 1) * retData.PageSize
	}

	if retData.PageSize == 0 {
		retData.PageSize = 15
	}

	var tagInfo dto.SearchTag = dto.SearchTag{
		Type:    -1,
		Status:  -1,
		GroupId: -1,
	}
	_ = json.Unmarshal([]byte(retData.Tag), &tagInfo)

	emailList, total := list.GetEmailList(ctx, tagInfo, retData.Keyword, false, offset, retData.PageSize)

	for _, email := range emailList {
		var sender User
		_ = json.Unmarshal([]byte(email.Sender), &sender)

		// Sender 头为空时回退使用 From 头信息，保证发件人展示不缺失
		if sender.EmailAddress == "" {
			sender.EmailAddress = email.FromAddress
			sender.Name = email.FromName
		}

		var tos []User
		_ = json.Unmarshal([]byte(email.To), &tos)

		// 认证结论统一由 parsemail.NewEmailAuthentication 生成
		authentication := parsemail.NewEmailAuthentication(email.SPFCheck == 1, email.DKIMCheck == 1)

		lst = append(lst, &emilItem{
			ID:        email.Id,
			UeId:      email.UeId, // user_email 记录 ID。修改日期: 20260510
			Type:      email.Type, // 0=接收, 1=发送。修改日期: 20260510
			Title:     email.Subject,
			Desc:      email.Text.String,
			Datetime:  email.SendDate.Format("2006-01-02 15:04:05"),
			IsRead:    email.IsRead == 1,
			Sender:    sender,
			To:        tos,
			Dangerous: authentication.Dangerous,
			Error:     email.Error.String,
		})
	}

	ret := emailListResponse{
		CurrentPage: retData.CurrentPage,
		TotalPage:   cast.ToInt(math.Ceil(cast.ToFloat64(total) / cast.ToFloat64(retData.PageSize))),
		List:        lst,
	}
	response.NewSuccessResponse(ret).FPrint(w)
}
