package smtp_server

import "github.com/emersion/go-sasl"

// 使用用户名和密码进行用户认证。
type LoginAuthenticator func(username, password string) error

type loginState int

const (
	loginNotStarted loginState = iota
	loginWaitingUsername
	loginWaitingPassword
)

type loginServer struct {
	state              loginState
	username, password string
	authenticate       LoginAuthenticator
}

// LOGIN 认证机制的服务端实现，详见
// https://tools.ietf.org/html/draft-murchison-sasl-login-00。
//
// LOGIN 已过时，仅应为无法升级使用 PLAIN 的旧客户端启用。
func NewLoginServer(authenticator LoginAuthenticator) sasl.Server {
	return &loginServer{authenticate: authenticator}
}

func (a *loginServer) Next(response []byte) (challenge []byte, done bool, err error) {
	switch a.state {
	case loginNotStarted:
		// 检查初始响应字段，依据 RFC4422 第 3 节
		if response == nil {
			challenge = []byte("Username:")
			break
		}
		a.state++
		fallthrough
	case loginWaitingUsername:
		a.username = string(response)
		challenge = []byte("Password:")
	case loginWaitingPassword:
		a.password = string(response)
		err = a.authenticate(a.username, a.password)
		done = true
	default:
		err = sasl.ErrUnexpectedClientResponse
	}

	a.state++
	return
}
