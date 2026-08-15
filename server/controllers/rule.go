package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/i18n"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/services/rule"
	"github.com/ffyuhf/pmail/utils/address"
	"github.com/ffyuhf/pmail/utils/array"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/errors"
)

func GetRule(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	res := rule.GetAllRules(ctx, ctx.UserID)
	response.NewSuccessResponse(res).FPrint(w)
}

func UpsertRule(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	// 修改日期: 20260610 — #6 修复 io.ReadAll 错误未返回响应
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Read body error", err.Error()).FPrint(w)
		return
	}

	var data *dto.Rule
	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "params error", err).FPrint(w)
		return
	}

	if data.Action == dto.FORWARD && !address.IsValidEmailAddress(data.Params) {

		response.NewErrorResponse(response.ParamsError, "ParamsError error", i18n.GetText(ctx.Lang, "invalid_email_address")).FPrint(w)
		return
	}

	for _, r := range data.Rules {
		if !array.InArray(r.Field, []string{"From", "Subject", "To", "Cc", "Text", "Html", "Content"}) {
			response.NewErrorResponse(response.ParamsError, "ParamsError error", "params error! Rule Field Error!").FPrint(w)
			return
		}
	}

	err = save(ctx, data.Encode())
	if err != nil {
		response.NewErrorResponse(response.ServerError, "server error", err).FPrint(w)
		return
	}
	response.NewSuccessResponse("succ").FPrint(w)
}

func save(ctx *context.Context, p *models.Rule) error {

	// 修改日期: 20260610 — #2 修复 Rule UPDATE 缺少 user_id 权限校验
	// 管理员可修改任何规则（无 user_id 限制），普通用户仅能修改自己的规则
	if p.Id > 0 {
		if ctx.IsAdmin {
			_, err := db.Instance.Exec(db.WithContext(ctx, "update rule set name=? ,value = ? ,action = ?,params = ?,sort = ? where id = ?"), p.Name, p.Value, p.Action, p.Params, p.Sort, p.Id)
			if err != nil {
				return errors.Wrap(err)
			}
		} else {
			_, err := db.Instance.Exec(db.WithContext(ctx, "update rule set name=? ,value = ? ,action = ?,params = ?,sort = ? where id = ? and user_id = ?"), p.Name, p.Value, p.Action, p.Params, p.Sort, p.Id, ctx.UserID)
			if err != nil {
				return errors.Wrap(err)
			}
		}
		return nil
	} else {
		_, err := db.Instance.Exec(db.WithContext(ctx, "insert into rule (name,value,user_id,action,params,sort) values (?,?,?,?,?,?)"), p.Name, p.Value, ctx.UserID, p.Action, p.Params, p.Sort)
		if err != nil {
			return errors.Wrap(err)
		}
		return nil
	}

}

type delRuleReq struct {
	Id int `json:"id"`
}

func DelRule(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	// 修改日期: 20260610 — #6 修复 io.ReadAll 错误未返回响应
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Read body error", err.Error()).FPrint(w)
		return
	}

	var data delRuleReq
	err = json.Unmarshal(requestBody, &data)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "params error", err).FPrint(w)
		return
	}

	if data.Id <= 0 {
		response.NewErrorResponse(response.ParamsError, "params error", "id is empty").FPrint(w)
		return
	}

	_, err = db.Instance.Exec(db.WithContext(ctx, "delete from rule where id =? and user_id =?"), data.Id, ctx.UserID)
	if err != nil {
		response.NewErrorResponse(response.ServerError, "unknown error", err).FPrint(w)
		return
	}

	response.NewSuccessResponse("succ").FPrint(w)
}
