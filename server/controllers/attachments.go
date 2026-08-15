package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/services/attachments"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/spf13/cast"
)

func GetAttachments(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	urlInfos := strings.Split(req.RequestURI, "/")
	if len(urlInfos) != 4 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	emailId := cast.ToInt(urlInfos[2])
	cid := urlInfos[3]

	contentType, content := attachments.GetAttachments(ctx, emailId, cid)

	if len(content) == 0 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}

func Download(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	urlInfos := strings.Split(req.RequestURI, "/")
	if len(urlInfos) != 5 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	emailId := cast.ToInt(urlInfos[3])
	index := cast.ToInt(urlInfos[4])

	fileName, content := attachments.GetAttachmentsByIndex(ctx, emailId, index)

	if len(content) == 0 {
		response.NewErrorResponse(response.ParamsError, "", "").FPrint(w)
		return
	}
	// 修改日期: 20260610 — #8 修复 Content-Disposition 文件名注入
	// 过滤 CRLF 字符防止 HTTP 头注入，使用 RFC 5987 编码防止特殊字符攻击
	w.Header().Set("Content-Type", "application/octet-stream")
	safeName := strings.NewReplacer("\r", "", "\n", "", `"`, "'").Replace(fileName)
	w.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s",
			safeName, url.PathEscape(fileName)))
	w.Write(content)
}
