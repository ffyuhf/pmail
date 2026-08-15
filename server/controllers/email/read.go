package email

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/services/detail"
	"github.com/ffyuhf/pmail/utils/context"
)

type markReadRequest struct {
	IDs []int `json:"ids"`
}

// 修改日期: 20260610 — #6 修复 io.ReadAll/json.Unmarshal 错误未返回响应，清理无效 err 检查
func MarkRead(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Read body error", err.Error()).FPrint(w)
		return
	}
	var reqData markReadRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Params error", err.Error()).FPrint(w)
		return
	}

	if len(reqData.IDs) <= 0 {
		response.NewErrorResponse(response.ParamsError, "ID错误", "").FPrint(w)
		return
	}

	for _, id := range reqData.IDs {
		detail.GetEmailDetail(ctx, id, true)
	}

	response.NewSuccessResponse("success").FPrint(w)
}
