package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"strings"

	"github.com/ffyuhf/pmail/db"
	"github.com/ffyuhf/pmail/models"
	"github.com/ffyuhf/pmail/utils/context"
	log "github.com/sirupsen/logrus"
)

// HasAuth 检查当前用户是否有某个邮件的auth
func HasAuth(ctx *context.Context, email *models.Email) bool {
	if ctx.IsAdmin {
		return true
	}
	var ue []models.UserEmail
	err := db.Instance.Table(&models.UserEmail{}).Where("email_id = ? and user_id = ?", email.Id, ctx.UserID).Find(&ue)
	if err != nil {
		log.Errorf("Error while checking user: %v", err)
		return false
	}

	return len(ue) != 0
}

func DkimGen() string {
	privKeyStr, _ := os.ReadFile("./config/dkim/dkim.priv")
	publicKeyStr, _ := os.ReadFile("./config/dkim/dkim.public")
	if len(privKeyStr) > 0 && len(publicKeyStr) > 0 {
		return string(publicKeyStr)
	}

	var (
		privKey crypto.Signer
		err     error
	)

	// 使用 2048 位 RSA 密钥以满足当前安全标准（NIST/ENISA 要求，1024 位已被证明不安全）
	privKey, err = rsa.GenerateKey(rand.Reader, 2048)

	if err != nil {
		log.Fatalf("Failed to generate key: %v", err)
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		log.Fatalf("Failed to marshal private key: %v", err)
	}

	f, err := os.OpenFile("./config/dkim/dkim.priv", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Failed to create key file: %v", err)
	}
	defer f.Close()

	privBlock := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}
	if err := pem.Encode(f, &privBlock); err != nil {
		log.Fatalf("Failed to write key PEM block: %v", err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("Failed to close key file: %v", err)
	}

	var pubBytes []byte

	switch pubKey := privKey.Public().(type) {
	case *rsa.PublicKey:
		// RFC 6376 对于 RSA 公钥应使用 RSAPublicKey 还是 SubjectPublicKeyInfo 格式并不一致。
		// 勘误 3017（https://www.rfc-editor.org/errata/eid3017）
		// 建议两者都允许。我们使用 SubjectPublicKeyInfo 格式，
		// 以与 opendkim、Gmail 和 Fastmail 等其他实现保持一致。
		pubBytes, err = x509.MarshalPKIXPublicKey(pubKey)
		if err != nil {
			log.Fatalf("Failed to marshal public key: %v", err)
		}
	default:
		panic("unreachable")
	}

	params := []string{
		"v=DKIM1",
		"k=rsa",
		"p=" + base64.StdEncoding.EncodeToString(pubBytes),
	}

	publicKey := strings.Join(params, "; ")

	// 权限 0644：所有者读写，其他用户只读。防止非授权用户篡改 DKIM 公钥。
	os.WriteFile("./config/dkim/dkim.public", []byte(publicKey), 0644)

	return publicKey
}
