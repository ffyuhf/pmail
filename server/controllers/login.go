package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/i18n"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/session"
	"github.com/ffyuhf/pmail/utils/array"
	"github.com/ffyuhf/pmail/utils/context"
	"github.com/ffyuhf/pmail/utils/errors"
	"github.com/ffyuhf/pmail/utils/password"
	"github.com/ffyuhf/pmail/utils/ratelimit"
	log "github.com/sirupsen/logrus"
)

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// 修改日期: 20260610 — #6 修复 io.ReadAll/json.Unmarshal 错误未返回响应
func Login(ctx *context.Context, w http.ResponseWriter, req *http.Request) {

	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Read body error", err.Error()).FPrint(w)
		return
	}
	var reqData loginRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Params error", err.Error()).FPrint(w)
		return
	}

	// 暴力破解防护：提取客户端 IP，检查速率限制
	clientIP := ratelimit.ExtractIP(req.RemoteAddr)
	if lockErr := ratelimit.Check(clientIP, reqData.Account); lockErr != nil {
		log.WithField("ip", clientIP).Warnf("Login rate limited: %v", lockErr)
		response.NewErrorResponse(response.ParamsError, i18n.GetText(ctx.Lang, "aperror"), "").FPrint(w)
		return
	}

	// 指数退避延迟：根据历史失败次数增加等待时间
	ratelimit.WaitDelay(clientIP, reqData.Account)

	var user models.User

	// 仅用账号查询，不将密码作为查询条件（支持双算法验证）
	_, err = db.Instance.Where("account =? and disabled=0", reqData.Account).Get(&user)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Errorf("%+v", err)
	}

	// 使用双算法验证：先bcrypt，后旧MD5
	if user.ID != 0 {
		ok, needsUpgrade := password.Verify(reqData.Password, user.Password)
		if !ok {
			// 密码错误，记录失败并返回通用错误
			ratelimit.RecordFailure(clientIP, reqData.Account)
			response.NewErrorResponse(response.ParamsError, i18n.GetText(ctx.Lang, "aperror"), "").FPrint(w)
			return
		}

		// 登录成功，清除速率限制记录
		ratelimit.RecordSuccess(clientIP, reqData.Account)

		// 旧MD5密码自动升级为bcrypt
		if needsUpgrade {
			newHash := password.Encode(reqData.Password)
			_, _ = db.Instance.Table("user").Where("id=?", user.ID).Update(map[string]interface{}{"password": newHash})
		}
		userStr, _ := json.Marshal(user)
		session.Instance.Put(req.Context(), "user", string(userStr))

		domains := config.Instance.Domains
		domains = array.Difference(domains, []string{config.Instance.Domain})
		domains = append([]string{config.Instance.Domain}, domains...)

		response.NewSuccessResponse(map[string]any{
			"account":  user.Account,
			"name":     user.Name,
			"is_admin": user.IsAdmin,
			"domains":  domains,
		}).FPrint(w)
	} else {
		// 用户不存在，记录失败
		ratelimit.RecordFailure(clientIP, reqData.Account)
		response.NewErrorResponse(response.ParamsError, i18n.GetText(ctx.Lang, "aperror"), "").FPrint(w)
	}
}

func Logout(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	session.Instance.Clear(ctx.Context)
	response.NewSuccessResponse("Success").FPrint(w)
}

// generateTokenRequest 生成 API Token 的请求参数
type generateTokenRequest struct {
	ExpiresIn int64 `json:"expires_in"` // 有效期（秒），0 表示永不过期
}

// GenerateAPIToken 为已登录用户生成随机不透明 API Token
// 生成的 Token 可通过 HTTP Header "Token" 用于 API 认证，替代旧的密码派生格式
func GenerateAPIToken(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	if ctx.UserID == 0 {
		response.NewErrorResponse(response.NeedLogin, i18n.GetText(ctx.Lang, "login_exp"), "").FPrint(w)
		return
	}

	// 修改日期: 20260610 — #6 修复 io.ReadAll/json.Unmarshal 错误未返回响应
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Read body error", err.Error()).FPrint(w)
		return
	}
	var reqData generateTokenRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Params error", err.Error()).FPrint(w)
		return
	}

	// 生成 32 字节随机数，编码为 64 字符 hex 字符串
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		log.Errorf("failed to generate random token: %v", err)
		response.NewErrorResponse(response.ServerError, "token generation failed", "").FPrint(w)
		return
	}
	tokenStr := hex.EncodeToString(tokenBytes)

	// 计算过期时间
	var expiresAt time.Time
	if reqData.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(reqData.ExpiresIn) * time.Second)
	}

	// 存入数据库
	apiToken := models.ApiToken{
		UserID:     ctx.UserID,
		Token:      tokenStr,
		ClientIP:   ratelimit.ExtractIP(req.RemoteAddr),
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		LastUsedAt: time.Now(),
	}
	_, err = db.Instance.Insert(&apiToken)
	if err != nil {
		log.Errorf("failed to save api token: %v", err)
		response.NewErrorResponse(response.ServerError, "token save failed", "").FPrint(w)
		return
	}

	response.NewSuccessResponse(map[string]any{
		"token":      tokenStr,
		"expires_at": expiresAt,
	}).FPrint(w)
}

// revokeTokenRequest 撤销 API Token 的请求参数
type revokeTokenRequest struct {
	Token string `json:"token"`
}

// RevokeAPIToken 撤销指定的 API Token，使其立即失效
func RevokeAPIToken(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	if ctx.UserID == 0 {
		response.NewErrorResponse(response.NeedLogin, i18n.GetText(ctx.Lang, "login_exp"), "").FPrint(w)
		return
	}

	// 修改日期: 20260610 — #6 修复 io.ReadAll/json.Unmarshal 错误未返回响应
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Read body error", err.Error()).FPrint(w)
		return
	}
	var reqData revokeTokenRequest
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		response.NewErrorResponse(response.ParamsError, "Params error", err.Error()).FPrint(w)
		return
	}

	if reqData.Token == "" {
		response.NewErrorResponse(response.ParamsError, "token is required", "").FPrint(w)
		return
	}

	// 仅允许撤销属于自己的 Token
	_, err = db.Instance.Table(&models.ApiToken{}).
		Where("token = ? AND user_id = ?", reqData.Token, ctx.UserID).
		Delete(&models.ApiToken{})
	if err != nil {
		log.Errorf("failed to revoke api token: %v", err)
		response.NewErrorResponse(response.ServerError, "revoke failed", "").FPrint(w)
		return
	}

	response.NewSuccessResponse("Success").FPrint(w)
}
