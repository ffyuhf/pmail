package email

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/services/del_email"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
)

type emailDeleteRequest struct {
	IDs       []int `json:"ids"`    // email.id（向后兼容，IMAP 使用）
	UeIds     []int `json:"ue_ids"` // user_email.id（Web 前端使用，精确匹配）。修改日期: 20260510
	ForcedDel bool  `json:"forcedDel"`
}

// EmailDelete 删除邮件
// 修改日期: 20260510 — Web 前端传递 ue_ids 精确匹配 user_email 记录，避免跨文件夹影响
func EmailDelete(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var reqData emailDeleteRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	// 优先使用 ue_ids（Web 前端），否则使用 ids（IMAP 向后兼容）
	if len(reqData.UeIds) > 0 {
		err = del_email.DelEmailByUeIds(ctx, reqData.UeIds, reqData.ForcedDel)
	} else if len(reqData.IDs) > 0 {
		err = del_email.DelEmail(ctx, reqData.IDs, reqData.ForcedDel)
	} else {
		response.NewErrorResponse(response.ParamsError, "ID错误", "").FPrint(w)
		return
	}

	if err != nil {
		response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
		return
	}
	response.NewSuccessResponse("success").FPrint(w)
}
