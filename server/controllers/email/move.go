package email

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/group"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
)

type moveRequest struct {
	GroupId   int    `json:"group_id"`
	GroupName string `json:"group_name"`
	IDs       []int  `json:"ids"`    // email.id（向后兼容，IMAP 使用）
	UeIds     []int  `json:"ue_ids"` // user_email.id（Web 前端使用，精确匹配）。修改日期: 20260510
}

// Move 移动邮件到指定文件夹
// 修改日期: 20260510 — Web 前端传递 ue_ids 精确匹配 user_email 记录，避免跨文件夹影响
func Move(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}
	var reqData moveRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		log.WithContext(ctx).Errorf("%+v", err)
	}

	// 优先使用 ue_ids（Web 前端），否则使用 ids（IMAP 向后兼容）
	ids := reqData.UeIds
	if len(ids) <= 0 {
		ids = reqData.IDs
	}
	if len(ids) <= 0 {
		response.NewErrorResponse(response.ParamsError, "ID错误", "").FPrint(w)
		return
	}

	if name, ok := models.GroupCodeToName[reqData.GroupId]; ok {
		err := group.Move2DefaultBox(ctx, ids, name)
		if err != nil {
			response.NewErrorResponse(response.ServerError, "Error", err.Error()).FPrint(w)
			return
		}
	} else if ok, errMsg := group.MoveMailToGroup(ctx, ids, reqData.GroupId); !ok {
		response.NewErrorResponse(response.ServerError, "Error", errMsg).FPrint(w)
		return
	}
	response.NewSuccessResponse("success").FPrint(w)
}
