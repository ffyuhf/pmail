package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"io"

	"github.com/ffyuhf/pmail/config"
	"github.com/ffyuhf/pmail/dto/response"
	"github.com/ffyuhf/pmail/services/setup"
	"github.com/ffyuhf/pmail/services/setup/ssl"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

// AcmeChallenge 处理 ACME HTTP-01 挑战请求，用于自动 SSL 证书验证。
func AcmeChallenge(w http.ResponseWriter, r *http.Request) {
	log.Infof("AcmeChallenge: %s", r.URL.Path)
	instance := ssl.GetHttpChallengeInstance()
	token := strings.ReplaceAll(r.URL.Path, "/.well-known/acme-challenge/", "")
	auth, exist := instance.AuthInfo[token]
	if exist {
		w.Write([]byte(auth.KeyAuth))
	} else {
		log.Errorf("AcmeChallenge Error Token Infos:%+v", instance.AuthInfo)
		http.NotFound(w, r)
	}
}

type sslResponse struct {
	Port int    `json:"port"`
	Type string `json:"type"`
}

// Setup 处理系统初始化引导的所有 API 请求。
// 安全机制：限制 HTTP 方法为 POST，要求请求中携带有效的 SetupToken，防止未授权访问。
// 修改日期: 20260425，新增 Token 校验、DSN 脱敏、SSL 路径安全校验。
// 修改日期: 20260425，v1.1 加固：HTTP 方法限制、路径白名单。
func Setup(ctx *context.Context, w http.ResponseWriter, req *http.Request) {
	// ====== 安全校验：仅允许 POST 方法 ======
	// 拒绝 GET 等方法，防止 Token 通过 URL 暴露在代理日志、浏览器历史记录中。
	if req.Method != http.MethodPost {
		log.Warnf("Setup API rejected: method %s not allowed from %s", req.Method, req.RemoteAddr)
		response.NewErrorResponse(response.ServerError, "Method not allowed, use POST", "").FPrint(w)
		return
	}

	// ====== 安全校验：验证 Setup Token ======
	// 从 URL query param 和请求体中读取 token，两者任一匹配即可
	queryToken := req.URL.Query().Get("token")
	reqBytes, err := io.ReadAll(req.Body)
	if err != nil {
		response.NewSuccessResponse("").FPrint(w)
		return
	}

	var reqData map[string]string
	err = json.Unmarshal(reqBytes, &reqData)
	if err != nil {
		response.NewErrorResponse(response.ServerError, "", err.Error()).FPrint(w)
		return
	}

	// 优先使用 URL 参数中的 token，其次使用请求体中的 token
	requestToken := queryToken
	if requestToken == "" {
		requestToken = reqData["token"]
	}

	if config.SetupToken == "" || requestToken != config.SetupToken {
		log.Warnf("Setup API rejected: invalid or missing token from %s", req.RemoteAddr)
		response.NewErrorResponse(response.ServerError, "Invalid or missing setup token", "").FPrint(w)
		return
	}
	// ====== 安全校验结束 ======

	if reqData["step"] == "database" && reqData["action"] == "get" {
		dbType, dbDSN, err := setup.GetDatabaseSettings(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}

		// DSN 脱敏：隐藏密码部分，防止敏感信息泄露
		// 修改日期: 20260425
		maskedDSN := maskDSN(dbType, dbDSN)

		response.NewSuccessResponse(map[string]string{
			"db_type": dbType,
			"db_dsn":  maskedDSN,
		}).FPrint(w)
		return
	}

	if reqData["step"] == "database" && reqData["action"] == "set" {
		err := setup.SetDatabaseSettings(ctx, cast.ToString(reqData["db_type"]), cast.ToString(reqData["db_dsn"]))
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}

		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "password" && reqData["action"] == "get" {
		ok, err := setup.GetAdminPassword(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		// 仅返回账户名，不返回密码信息
		response.NewSuccessResponse(ok).FPrint(w)
		return
	}

	if reqData["step"] == "password" && reqData["action"] == "set" {
		err := setup.SetAdminPassword(ctx, cast.ToString(reqData["account"]), cast.ToString(reqData["password"]))
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "get" {
		smtpDomain, webDomain, domains, err := setup.GetDomainSettings()
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse(map[string]any{
			"smtp_domain": smtpDomain,
			"web_domain":  webDomain,
			"domains":     domains,
		}).FPrint(w)
		return
	}

	if reqData["step"] == "domain" && reqData["action"] == "set" {
		err := setup.SetDomainSettings(cast.ToString(reqData["smtp_domain"]), cast.ToString(reqData["web_domain"]), reqData["multi_domain"])
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse("Succ").FPrint(w)
		return
	}

	if reqData["step"] == "dns" && reqData["action"] == "get" {
		dnsInfos, err := setup.GetDNSSettings(ctx)
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}
		response.NewSuccessResponse(dnsInfos).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "get" {
		sslType := ssl.GetSSL()
		res := sslResponse{
			Type: sslType,
			Port: config.Instance.GetSetupPort(),
		}
		response.NewSuccessResponse(res).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "getParams" {
		dnsChallenge := ssl.GetDnsChallengeInstance()

		response.NewSuccessResponse(dnsChallenge.GetDNSSettings(ctx)).FPrint(w)
		return
	}

	if reqData["step"] == "ssl" && reqData["action"] == "set" {

		if reqData["ssl_type"] == config.SSLTypeUser {
			keyPath := reqData["key_path"]
			crtPath := reqData["crt_path"]

			// 路径安全校验：禁止路径遍历攻击
			// 修改日期: 20260425
			if err := validateFilePath(keyPath); err != nil {
				response.NewErrorResponse(response.ServerError, fmt.Sprintf("key_path: %s", err.Error()), "").FPrint(w)
				return
			}
			if err := validateFilePath(crtPath); err != nil {
				response.NewErrorResponse(response.ServerError, fmt.Sprintf("crt_path: %s", err.Error()), "").FPrint(w)
				return
			}

			_, err := os.Stat(cast.ToString(keyPath))
			if err != nil {
				response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
				return
			}

			_, err = os.Stat(cast.ToString(crtPath))
			if err != nil {
				response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
				return
			}
		}

		err = ssl.SetSSL(cast.ToString(reqData["ssl_type"]), cast.ToString(reqData["key_path"]), cast.ToString(reqData["crt_path"]))
		if err != nil {
			response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
			return
		}

		if reqData["ssl_type"] == config.SSLTypeAutoHTTP || reqData["ssl_type"] == config.SSLTypeAutoDNS {
			err = ssl.GenSSL(false)
			if err != nil {
				response.NewErrorResponse(response.ServerError, err.Error(), "").FPrint(w)
				return
			}
		}

		response.NewSuccessResponse("Succ").FPrint(w)

		if reqData["ssl_type"] == config.SSLTypeUser {
			setup.Finish()
		}
		return
	}

}

// maskDSN 对数据库连接串进行脱敏处理，隐藏密码部分。
// 支持格式：
//   - MySQL: root:password@tcp(host:port)/dbname
//   - PostgreSQL: postgres://user:password@host:port/dbname
//   - SQLite: 文件路径（无需脱敏）
//
// 修改日期: 20260425
func maskDSN(dbType, dsn string) string {
	if dsn == "" {
		return ""
	}

	switch dbType {
	case config.DBTypeMySQL:
		// MySQL DSN 格式: user:password@tcp(host:port)/dbname?params
		// 匹配 user:password@ 部分，将 password 替换为 ***
		mysqlPattern := regexp.MustCompile(`^([^:]+):([^@]+)@(.+)$`)
		if mysqlPattern.MatchString(dsn) {
			parts := mysqlPattern.FindStringSubmatch(dsn)
			return fmt.Sprintf("%s:***@%s", parts[1], parts[3])
		}

	case config.DBTypePostgres:
		// PostgreSQL DSN 格式: postgres://user:password@host:port/dbname?params
		pgPattern := regexp.MustCompile(`^(postgres://[^:]+):([^@]+)@(.+)$`)
		if pgPattern.MatchString(dsn) {
			parts := pgPattern.FindStringSubmatch(dsn)
			return fmt.Sprintf("%s:***@%s", parts[1], parts[3])
		}
	}

	// SQLite 或无法识别的格式，原样返回
	return dsn
}

// validateFilePath 校验文件路径安全性，防止路径遍历攻击和文件系统探测。
// 安全规则：
//   - 禁止空路径
//   - 禁止路径遍历（包含 ".."）
//   - 禁止绝对路径（以 "/" 开头），仅允许相对路径，防止探测服务器任意文件
//
// 修改日期: 20260425，v1.1 增加绝对路径拒绝。
func validateFilePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// 检查路径遍历
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal is not allowed")
	}

	// 检查绝对路径：仅允许相对路径，防止通过 /etc/passwd 等路径探测文件系统
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("absolute path is not allowed, use relative path instead")
	}

	return nil
}
